package binder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Values_Get(t *testing.T) {
	m := make(map[string]*Value)
	m["key"] = &Value{v: "value"}

	v := &Values{
		m,
	}

	value, ok := v.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "value", value)
}

func Test_Values_GetStrings(t *testing.T) {
	m := make(map[string]*Value)
	m["key"] = &Value{v: []string{"val1", "val2"}}

	v := &Values{
		m,
	}

	values, ok := v.GetStrings("key")
	assert.True(t, ok)
	assert.EqualValues(t, []string{"val1", "val2"}, values)
}

func Test_Values_GetInt(t *testing.T) {
	m := make(map[string]*Value)
	m["key"] = &Value{v: 100}

	v := &Values{
		m,
	}

	value, ok := v.GetInt("key")
	assert.True(t, ok)
	assert.Equal(t, 100, value)
}

func Test_Values_GetInt_Fail(t *testing.T) {
	m := make(map[string]*Value)
	m["key"] = &Value{v: "x"}

	v := &Values{
		m,
	}

	_, ok := v.GetInt("key")
	assert.False(t, ok)
}

func Test_Values_GetFloat(t *testing.T) {
	m := make(map[string]*Value)
	m["key"] = &Value{v: 100.01}

	v := &Values{
		m,
	}

	value, ok := v.GetFloat("key")
	assert.True(t, ok)
	assert.Equal(t, 100.01, value)
}

func Test_Values_GetFloat_Fail(t *testing.T) {
	m := make(map[string]*Value)
	m["key"] = &Value{v: "x"}

	v := &Values{
		m,
	}

	_, ok := v.GetFloat("key")
	assert.False(t, ok)
}

func Test_Values_GetBool(t *testing.T) {
	m := make(map[string]*Value)
	m["key"] = &Value{v: true}

	v := &Values{
		m,
	}

	value, ok := v.GetBool("key")
	assert.True(t, ok)
	assert.True(t, value)
}

func Test_Values_GetBool_Fail(t *testing.T) {
	m := make(map[string]*Value)
	m["key"] = &Value{v: "x"}

	v := &Values{
		m,
	}

	_, ok := v.GetBool("key")
	assert.False(t, ok)
}
