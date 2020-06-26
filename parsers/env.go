package parsers

import (
	"os"
	"strings"
)

// EnvParser is a configuration parser
// which reads from environment variables
type EnvParser struct {
	prefix string
}

// NewEnvParser returns a new EnvParser.
//
// An EnvParser read environment variables into
// a configuration
func NewEnvParser() *EnvParser {
	return &EnvParser{""}
}

// NewEnvParserWithPrefix returns a new EnvParser
// which reads environment variables with a
// specified prefix
func NewEnvParserWithPrefix(prefix string) *EnvParser {
	prefix = strings.ToLower(prefix)
	return &EnvParser{prefix}
}

// Parse returns environment variables as a map[string]interface{},
// which might be prefixed with a specified prefix
func (p *EnvParser) Parse() (map[string]interface{}, error) {
	values := make(map[string]interface{})

	for _, v := range os.Environ() {
		lc := strings.ToLower(v)
		if p.prefix != "" && !strings.HasPrefix(lc, p.prefix) {
			continue
		}

		kvp := strings.SplitN(v, "=", 2)
		if len(kvp) != 2 {
			continue
		}

		key := strings.Replace(strings.ToLower(strings.TrimSpace(kvp[0])), p.prefix, "", 1)
		value := strings.TrimSpace(kvp[1])

		values[key] = value
	}

	return values, nil
}
