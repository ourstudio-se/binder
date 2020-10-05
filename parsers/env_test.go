package parsers

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewEnvParserWithPrefix(t *testing.T) {
	prefixes := []string{"AbCdEfGh", "ijklm"}
	p := NewEnvParserWithPrefix(prefixes...)

	assert.Equal(t, prefixes, p.prefixes)
}

func Test_Env_Parse(t *testing.T) {
	prefix := "CONFIG_TEST_"
	key := "binder"
	envvar := fmt.Sprintf("%s%s", prefix, key)

	expected := "VAR_VALUE"
	_ = os.Setenv(envvar, expected)

	p := NewEnvParserWithPrefix(prefix)
	values, err := p.Parse()
	assert.NoError(t, err)

	assert.Equal(t, expected, values[key])
}

func Test_Env_Without_Prefix(t *testing.T) {
	key := "binder"
	expected := "VAR_VALUE"
	err := os.Setenv(key, expected)
	assert.NoError(t, err)

	p := NewEnvParserWithPrefix("")
	values, err := p.Parse()
	assert.NoError(t, err)

	assert.Equal(t, expected, values[key])
}
