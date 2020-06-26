package parsers

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewEnvParserWithPrefix(t *testing.T) {
	prefix := "AbCdEfGh"
	p := NewEnvParserWithPrefix(prefix)

	assert.Equal(t, strings.ToLower(prefix), p.prefix)
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
