package binder

import (
	"errors"
	"reflect"
	"sync"

	"github.com/radovskyb/watcher"
)

const configStructTagName string = "config"

// Parser is an interface which defines
// the minimum requirement to implement
// a custom configuration parser
type Parser interface {
	Parse() (map[string]interface{}, error)
}

// Config is the configuration handler,
// which can be read from or bound to
// a custom type
type Config struct {
	parsers []Parser
	binders []reflect.Value
	cache   *Values
	errch   chan error
	watch   *watcher.Watcher
	m       sync.RWMutex
}

// New is the configuration constructor,
// taking Option(s) as functional parameters
// to support a plethora of backing
// configuration parsers
func New(opts ...Option) *Config {
	c := &Config{}
	c.errch = make(chan error, 1)

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Close cleans up channels and file watches
func (c *Config) Close() {
	if c.watch != nil {
		c.watch.Close()
	}

	close(c.errch)
}

// Errors returns a `chan error` where
// any possible errors will be reported, say from
// a parsers as an example
func (c *Config) Errors() <-chan error {
	return c.errch
}

func (c *Config) errs(err error) {
	select {
	case c.errch <- err:
		return
	default:
		return
	}
}

// Use appends a backing Parser to the
// configuration handler
func (c *Config) Use(p Parser) {
	c.parsers = append(c.parsers, p)
}

// Watch adds a file or directory watch to the
// specified path, which will trigger a re-bind
// for any bound configuration
func (c *Config) Watch(path string) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.watch == nil {
		c.watch = newFileWatcher(c.apply, c.errs)
	}

	err := c.watch.Add(path)
	if err != nil {
		c.errs(err)
	}
}

// Values iterates through all specified
// backing parsers, and retrieves configuration
// values from all of them
func (c *Config) Values() *Values {
	if c.cache == nil {
		c.build()
	}

	return c.cache
}

func (c *Config) build() {
	m := make(map[string]*Value)

	for _, p := range c.parsers {
		raw, err := p.Parse()
		if err != nil {
			c.errs(err)
		}

		for k, v := range raw {
			m[k] = &Value{v}
		}
	}

	c.m.Lock()
	defer c.m.Unlock()

	c.cache = &Values{m}
}

// Bind takes one or more pointers to a custom type,
// which configuration values will be bound to
func (c *Config) Bind(outs ...interface{}) {
	for _, out := range outs {
		v := reflect.ValueOf(out)
		if v.Kind() != reflect.Ptr || v.IsNil() {
			c.errs(errors.New("cannot bind to non-pointer or nil"))
			continue
		}

		c.binders = append(c.binders, v)
		c.bind(v)
	}
}

func (c *Config) bind(v reflect.Value) {
	if c.cache == nil {
		c.build()
	}

	c.m.RLock()
	defer c.m.RUnlock()

	elem := v.Elem()
	t := elem.Type()
	changed := false
	for i := 0; i < t.NumField(); i++ {
		tag, ok := t.Field(i).Tag.Lookup(configStructTagName)
		if !ok || tag == "" || tag == "-" {
			continue
		}

		if elem.Field(i).Kind() == reflect.String {
			value, ok := c.cache.Get(tag)
			if ok {
				cur := elem.Field(i).String()
				changed = cur != "" && cur != value
				elem.Field(i).SetString(value)
			}
		}
		if elem.Field(i).Kind() == reflect.Array || elem.Field(i).Kind() == reflect.Slice {
			value, ok := c.cache.GetStrings(tag)
			if ok {
				cur := elem.Field(i).Interface().([]string)
				changed = cur != nil && isSliceEqual(cur, value)
				elem.Field(i).Set(reflect.ValueOf(value))
			}
		}
		if elem.Field(i).Kind() == reflect.Int {
			value, ok := c.cache.GetInt(tag)
			if ok {
				cur := elem.Field(i).Int()
				changed = cur != 0 && cur != int64(value)
				elem.Field(i).SetInt(int64(value))
			}
		}
		if elem.Field(i).Kind() == reflect.Float32 {
			value, ok := c.cache.GetFloat(tag)
			if ok {
				cur := elem.Field(i).Float()
				changed = cur != 0 && cur != float64(value)
				elem.Field(i).SetFloat(value)
			}
		}
		if elem.Field(i).Kind() == reflect.Float64 {
			value, ok := c.cache.GetFloat(tag)
			if ok {
				cur := elem.Field(i).Float()
				changed = cur != 0 && cur != value
				elem.Field(i).SetFloat(value)
			}
		}
		if elem.Field(i).Kind() == reflect.Bool {
			value, ok := c.cache.GetBool(tag)
			if ok {
				cur := elem.Field(i).Bool()
				changed = !cur && cur != value
				elem.Field(i).SetBool(value)
			}
		}
	}

	method := v.MethodByName("Notify")
	n := reflect.Value{}
	if changed && method != n {
		method.Call([]reflect.Value{})
	}
}

func (c *Config) apply() {
	c.build()

	for _, v := range c.binders {
		c.bind(v)
	}
}

func isSliceEqual(as interface{}, bs interface{}) bool {
	a := as.([]interface{})
	b := bs.([]interface{})

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func newFileWatcher(fn func(), errfn func(error)) *watcher.Watcher {
	w := watcher.New()
	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write)

	go func() {
		for {
			select {
			case <-w.Event:
				fn()
			case <-w.Closed:
				return
			case err := <-w.Error:
				errfn(err)
			}
		}
	}()

	return w
}
