package binder

import (
	"net/url"
	"strings"
	"testing"

	"github.com/ourstudio-se/binder/parsers"
	"github.com/stretchr/testify/assert"
)

func Test_WithParser(t *testing.T) {
	p := &fakeParser{}
	c := New(WithParser(p))

	assert.Len(t, c.parsers, 1)
	assert.Equal(t, p, c.parsers[0])
}

func Test_WithEnv(t *testing.T) {
	p := parsers.NewEnvParserWithPrefix("")
	c := New(WithEnv(""))

	assert.Len(t, c.parsers, 1)
	assert.Equal(t, p, c.parsers[0])
}

func Test_WithFile(t *testing.T) {
	fp := "/tmp/path"
	p := parsers.NewFileParser(fp, "=")
	c := New(WithFile(fp, "="))

	assert.Len(t, c.parsers, 1)
	assert.Equal(t, p, c.parsers[0])
}

func Test_WithKubernetesVolume(t *testing.T) {
	vp := "/tmp/path"
	p := parsers.NewKubernetesVolumeParser(vp)
	c := New(WithKubernetesVolume(vp))

	assert.Len(t, c.parsers, 1)
	assert.Equal(t, p, c.parsers[0])
}

func Test_WithURL(t *testing.T) {
	u := &url.URL{}
	p := parsers.NewRemoteFileParser(u)
	c := New(WithURL(u))

	assert.Len(t, c.parsers, 1)
	assert.Equal(t, p, c.parsers[0])
}

func Test_WithValue(t *testing.T) {
	r := strings.NewReader("key=value")
	p := parsers.NewKeyValueParser(r, parsers.WithKeyValueSeparator("="))
	c := New(WithValue("key", "value"))

	assert.Len(t, c.parsers, 1)
	assert.Equal(t, p, c.parsers[0])
}

func Test_WithWatch(t *testing.T) {
	c := New(WithWatch("/tmp/binder"))

	assert.NotNil(t, c.watch)
}
