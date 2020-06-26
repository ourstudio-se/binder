package parsers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewKeyValueParser(t *testing.T) {
	sep := "__"
	p := NewKeyValueParser(strings.NewReader(""), WithKeyValueSeparator(sep))

	assert.Equal(t, sep, p.separator)
}

func Test_KeyValue_Parse(t *testing.T) {
	key := "thekey"
	value := "thevalue"
	sep := "%"
	conf := fmt.Sprintf("%s%s %s", key, sep, value)
	r := strings.NewReader(conf)

	p := NewKeyValueParser(r, WithKeyValueSeparator(sep))
	values, err := p.Parse()
	assert.NoError(t, err)

	assert.Equal(t, value, values[key])
}
