package binder

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Values is a collection of configuration values.
type Values struct {
	m map[string]*Value
}

func (v *Values) getIgnoreCase(key string) *Value {
	for k, v := range v.m {
		if strings.EqualFold(key, k) {
			return v
		}
	}

	return nil
}

// Get returns the value matching the specified key,
// as a string. It returns true as second return
// value if the specified key exist, or false
// if no such key was found.
func (v *Values) Get(key string) (string, bool) {
	return v.m[key].String()
}

func (v *Values) getString(key string, op BindMode) (string, bool) {
	if op.has(ModeStrict) {
		return v.Get(key)
	}

	return v.getIgnoreCase(key).String()
}

// GetStrings returns the value matching the specified
// key as a collection of strings. It returns true as
// second return value if the specified key exist, or
// false if no such key was found or if it was in the
// wrong format.
func (v *Values) GetStrings(key string) ([]string, bool) {
	return v.m[key].StringArray()
}

func (v *Values) getStrings(key string, op BindMode) ([]string, bool) {
	if op.has(ModeStrict) {
		return v.GetStrings(key)
	}

	return v.getIgnoreCase(key).StringArray()
}

// GetInt returns the value matching the specified key,
// as an integer. It returns true as second return
// value if the specified key exist, or false
// if no such key was found.
func (v *Values) GetInt(key string) (int, bool) {
	return v.m[key].Int()
}

func (v *Values) getInt(key string, op BindMode) (int, bool) {
	if op.has(ModeStrict) {
		return v.GetInt(key)
	}

	return v.getIgnoreCase(key).Int()
}

// GetFloat returns the value matching the specified key,
// as a float. It returns true as second return
// value if the specified key exist, or false
// if no such key was found.
func (v *Values) GetFloat(key string) (float64, bool) {
	return v.m[key].Float()
}

func (v *Values) getFloat(key string, op BindMode) (float64, bool) {
	if op.has(ModeStrict) {
		return v.GetFloat(key)
	}

	return v.getIgnoreCase(key).Float()
}

// GetBool returns the value matching the specified key,
// as a boolean. It returns true as second return
// value if the specified key exist, or false
// if no such key was found.
func (v *Values) GetBool(key string) (bool, bool) {
	return v.m[key].Bool()
}

func (v *Values) getBool(key string, op BindMode) (bool, bool) {
	if op.has(ModeStrict) {
		return v.GetBool(key)
	}

	return v.getIgnoreCase(key).Bool()
}

// Value wraps a configuration value.
type Value struct {
	v interface{}
}

// String returns a configuration value in string format.
func (c *Value) String() (string, bool) {
	if c == nil {
		return "", false
	}

	if s, ok := c.v.(string); ok {
		return s, true
	}

	return fmt.Sprintf("%v", c.v), true
}

// StringArray returns a configuration collection of strings.
func (c *Value) StringArray() ([]string, bool) {
	if c == nil {
		return nil, false
	}

	o := reflect.ValueOf(c.v)
	var s []string
	for i := 0; i < o.Len(); i++ {
		s = append(s, o.Index(i).Interface().(string))
	}

	return s, true
}

// Int returns a configuration value as an integer, and
// return true as second return value if the value could
// be returned as an int - otherwise it returns false.
func (c *Value) Int() (int, bool) {
	if c == nil {
		return 0, false
	}

	if i, ok := c.v.(int); ok {
		return i, true
	}

	if s, ok := c.v.(string); ok {
		i, err := strconv.Atoi(s)
		if err == nil {
			return i, true
		}
	}

	return 0, false
}

// Float returns a configuration value as a float, and
// return true as second return value if the value could
// be returned as a float - otherwise it returns false.
func (c *Value) Float() (float64, bool) {
	if c == nil {
		return 0, false
	}

	if f, ok := c.v.(float32); ok {
		return float64(f), true
	}
	if f, ok := c.v.(float64); ok {
		return f, true
	}
	if f, ok := c.v.(int); ok {
		return float64(f), true
	}

	if s, ok := c.v.(string); ok {
		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return f, true
		}
	}

	return 0, false
}

// Bool returns a configuration value as a boolean, and
// return true as second return value if the value could
// be returned as a boolean - otherwise it returns false.
func (c *Value) Bool() (bool, bool) {
	if c == nil {
		return false, false
	}

	if b, ok := c.v.(bool); ok {
		return b, true
	}

	if s, ok := c.v.(string); ok {
		b, err := strconv.ParseBool(s)
		if err == nil {
			return b, true
		}
	}

	return false, false
}
