package parsers

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func Test_FlagSet_Parse(t *testing.T) {
	flagSet := pflag.FlagSet{}

	key := "thekey"
	value := "thevalue"

	flagSet.String(key, "", "")
	_ = flagSet.Parse([]string{"--" + key, value})

	p := NewFlagSetParser(&flagSet)
	values, err := p.Parse()
	assert.NoError(t, err)

	assert.Equal(t, value, values[key])
}
