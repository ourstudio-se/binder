package binder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeOption struct {
	called bool
}

func newFakeOption() *fakeOption {
	return &fakeOption{false}
}

func (fo *fakeOption) fn(_ *Config) {
	fo.called = true
}

type fakeParser struct {
	called bool
	r      map[string]interface{}
	err    error
}

func newFakeParser(r map[string]interface{}) *fakeParser {
	return &fakeParser{false, r, nil}
}

func (p *fakeParser) Parse() (map[string]interface{}, error) {
	return p.r, p.err
}

type fakeBinder struct {
	ValueField string `config:"binder_key"`
	notified   bool
}

func (f *fakeBinder) Notify() {
	f.notified = true
}

type fakeBinder2 struct {
	ValueField2 string `config:"binder_key_two"`
}

type fakeBinder3 struct {
	ValueField string `config:"binder_key"`
}

func Test_New(t *testing.T) {
	fakes := []*fakeOption{
		newFakeOption(),
		newFakeOption(),
		newFakeOption(),
	}
	var opts []Option

	for _, fake := range fakes {
		opts = append(opts, fake.fn)
	}

	_ = New(opts...)

	everyFn := true
	for _, opt := range fakes {
		if !opt.called {
			everyFn = false
		}
	}

	assert.True(t, everyFn)
}

func Test_Use(t *testing.T) {
	fakes := []*fakeParser{
		newFakeParser(nil),
		newFakeParser(nil),
		newFakeParser(nil),
	}

	c := New()

	for _, fake := range fakes {
		c.Use(fake)
	}

	assert.Equal(t, len(fakes), len(c.parsers))
}

func Test_Build(t *testing.T) {
	m1 := make(map[string]interface{})
	m1["k1"] = "v1"

	m2 := make(map[string]interface{})
	m2["k2"] = "v2"

	c := New(
		WithParser(newFakeParser(m1)),
		WithParser(newFakeParser(m2)))
	c.build()
	v := c.Values()

	v1, _ := v.Get("k1")
	assert.Equal(t, m1["k1"], v1)

	v2, _ := v.Get("k2")
	assert.Equal(t, m2["k2"], v2)
}

func Test_Build_Key_Collision(t *testing.T) {
	m1 := make(map[string]interface{})
	m1["key"] = "value1"

	m2 := make(map[string]interface{})
	m2["key"] = "value2"

	c := New(
		WithParser(newFakeParser(m1)),
		WithParser(newFakeParser(m2)))

	c.build()
	v := c.Values()

	value, _ := v.Get("key")
	assert.Equal(t, m2["key"], value)
}

func Test_Bind(t *testing.T) {
	m := make(map[string]interface{})
	m["binder_key"] = "value"
	m["binder_key_two"] = "value_two"

	c := New(
		WithParser(newFakeParser(m)))

	var b1 fakeBinder
	var b2 fakeBinder2
	var b3 fakeBinder3

	c.Bind(&b1)
	c.Bind(&b2, &b3)

	assert.Equal(t, "value", b1.ValueField)
	assert.Equal(t, "value_two", b2.ValueField2)
	assert.Equal(t, "value", b3.ValueField)
}

type fakeRebindParser struct {
	value string
}

func (p *fakeRebindParser) Parse() (map[string]interface{}, error) {
	m := make(map[string]interface{})
	m["binder_key"] = p.value

	return m, nil
}

func Test_Rebind_Value(t *testing.T) {
	p := &fakeRebindParser{"value1"}
	c := New(WithParser(p))

	var b fakeBinder
	c.Bind(&b)

	p.value = "value2"
	c.apply()

	assert.Equal(t, "value2", b.ValueField)
}

func Test_Rebind_Notify(t *testing.T) {
	p := &fakeRebindParser{"value1"}
	c := New(WithParser(p))

	var b fakeBinder
	c.Bind(&b)

	p.value = "value2"
	c.apply()

	assert.True(t, b.notified)
}

func Test_Rebind_Notify_NoOp(t *testing.T) {
	p := &fakeRebindParser{"value1"}
	c := New(WithParser(p))

	var b fakeBinder
	c.Bind(&b)
	c.apply()

	assert.False(t, b.notified)
}
