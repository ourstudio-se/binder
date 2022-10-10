package parsers

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Flag_Parse(t *testing.T) {
	key := "thekey"
	value := "thevalue"

	flag.String(key, "", "")

	os.Args = []string{"program", "--" + key, value}
	flag.Parse()

	p := NewFlagParser()
	values, err := p.Parse()
	assert.NoError(t, err)

	assert.Equal(t, value, values[key])
}
