package binder

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/ourstudio-se/binder/parsers"
)

// Option is passed to the constructor of
// Config, and used through functional parameters
type Option func(*Config)

// WithParser is an Option to instantiate a
// custom parser with a Config
func WithParser(p Parser) Option {
	return func(c *Config) {
		c.Use(p)
	}
}

// WithEnv is an Option to instantiate a
// parser which reads environment variables
// when instantiating a Config
func WithEnv(prefixes ...string) Option {
	return WithParser(parsers.NewEnvParserWithPrefix(prefixes...))
}

// WithFile is an Option to instantiate a
// parser which reads a backing file using a
// specific key/value separator
func WithFile(filepath string, sep string) Option {
	return WithParser(parsers.NewFileParser(filepath, sep))
}

// WithKubernetesVolume is an Option to instantiate
// a parser which reads a Kubernetes mounted
// volume when instantiating a Config
func WithKubernetesVolume(path string) Option {
	return WithParser(parsers.NewKubernetesVolumeParser(path))
}

// WithURL is an Option to instantiate a
// parser which reads a remote file when
// instantiating a Config
func WithURL(u *url.URL) Option {
	return WithParser(parsers.NewRemoteFileParser(u))
}

// WithValue is an Option to add custom key/value
// pairs to a configuration
func WithValue(key string, value interface{}) Option {
	kv := fmt.Sprintf("%s=%v", key, value)
	r := strings.NewReader(kv)

	return WithParser(parsers.NewKeyValueParser(r, parsers.WithKeyValueSeparator("=")))
}

// WithWatch adds a file path watch, which can be
// used to reload configuration values that originates
// from a FileParser or a KubernetesVolumeParser when
// the backing files changes
func WithWatch(path string) Option {
	return func(c *Config) {
		c.Watch(path)
	}
}
