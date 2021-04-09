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
// a custom configuration parser.
type Parser interface {
	Parse() (map[string]interface{}, error)
}

// Config is the configuration handler,
// which can be read from or bound to
// a custom type.
type Config struct {
	parsers []Parser
	binders []reflect.Value
	cache   *Values
	errch   chan error
	watch   *watcher.Watcher
	m       sync.Mutex
}

// New is the configuration constructor,
// taking Option(s) as functional parameters
// to support a plethora of backing
// configuration parsers.
func New(opts ...Option) *Config {
	c := &Config{}
	c.errch = make(chan error, 1)

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Close cleans up channels and file watches.
func (c *Config) Close() {
	if c.watch != nil {
		c.watch.Close()
	}

	close(c.errch)
}

// Errors returns a `chan error` where
// any possible errors will be reported, say from
// a parsers as an example.
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
// configuration handler.
func (c *Config) Use(p Parser) {
	c.parsers = append(c.parsers, p)
}

// Watch adds a file or directory watch to the
// specified path, which will trigger a re-bind
// for any bound configuration.
func (c *Config) Watch(path string) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.watch == nil {
		c.watch = newFileWatcher(c.apply, c.errs)
	}

	if err := c.watch.Add(path); err != nil {
		c.errs(err)
	}
}

// Values iterates through all specified
// backing parsers, and retrieves configuration
// values from all of them.
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
// which configuration values will be bound to.
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

	c.m.Lock()
	defer c.m.Unlock()

	elem := v.Elem()
	t := elem.Type()
	changed := false
	for i := 0; i < t.NumField(); i++ {
		tag, ok := t.Field(i).Tag.Lookup(configStructTagName)
		if !ok || tag == "" || tag == "-" {
			continue
		}

		changed = c.bindValue(elem.Field(i), tag)
	}

	method := v.MethodByName("Notify")
	n := reflect.Value{}
	if changed && method != n {
		method.Call([]reflect.Value{})
	}
}

func (c *Config) bindValue(elem reflect.Value, tag string) bool {
	switch elem.Kind() {
	case reflect.String:
		return c.bindString(elem, tag)
	case reflect.Array, reflect.Slice:
		return c.bindStringArray(elem, tag)
	case reflect.Int:
		return c.bindInt(elem, tag)
	case reflect.Float32:
		return c.bindFloat32(elem, tag)
	case reflect.Float64:
		return c.bindFloat64(elem, tag)
	case reflect.Bool:
		return c.bindBool(elem, tag)
	}

	return false
}

func (c *Config) bindString(elem reflect.Value, tag string) bool {
	value, ok := c.cache.Get(tag)
	if ok {
		cur := elem.String()
		elem.SetString(value)
		return cur != "" && cur != value
	}

	return false
}

func (c *Config) bindStringArray(elem reflect.Value, tag string) bool {
	value, ok := c.cache.GetStrings(tag)
	if ok {
		cur := elem.Interface().([]string)
		elem.Set(reflect.ValueOf(value))
		return cur != nil && isSliceEqual(cur, value)
	}

	return false
}

func (c *Config) bindInt(elem reflect.Value, tag string) bool {
	value, ok := c.cache.GetInt(tag)
	if ok {
		cur := elem.Int()
		elem.SetInt(int64(value))
		return cur != 0 && cur != int64(value)
	}

	return false
}

func (c *Config) bindFloat32(elem reflect.Value, tag string) bool {
	value, ok := c.cache.GetFloat(tag)
	if ok {
		cur := elem.Float()
		elem.SetFloat(value)
		return cur != 0 && cur != float64(value)
	}

	return false
}

func (c *Config) bindFloat64(elem reflect.Value, tag string) bool {
	value, ok := c.cache.GetFloat(tag)
	if ok {
		cur := elem.Float()
		elem.SetFloat(value)
		return cur != 0 && cur != value
	}

	return false
}

func (c *Config) bindBool(elem reflect.Value, tag string) bool {
	value, ok := c.cache.GetBool(tag)
	if ok {
		cur := elem.Bool()
		elem.SetBool(value)
		return !cur && cur != value
	}

	return false
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
