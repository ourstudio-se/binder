package parsers

import (
	"os"
	"strings"
)

// EnvParser is a configuration parser
// which reads from environment variables.
type EnvParser struct {
	prefixes []string
}

// NewEnvParser returns a new EnvParser.
//
// An EnvParser read environment variables into
// a configuration.
func NewEnvParser() *EnvParser {
	return &EnvParser{}
}

// NewEnvParserWithPrefix returns a new EnvParser
// which reads environment variables with any of
// the specified prefixes.
func NewEnvParserWithPrefix(prefixes ...string) *EnvParser {
	if len(prefixes) == 0 {
		prefixes = append(prefixes, "")
	}

	return &EnvParser{prefixes}
}

// Parse returns environment variables as a map[string]interface{},
// which might be prefixed with a specified prefix.
func (p *EnvParser) Parse() (map[string]interface{}, error) {
	values := make(map[string]interface{})

	for _, v := range os.Environ() {
		for _, prefix := range p.prefixes {
			lc := strings.ToLower(v)
			lcp := strings.ToLower(prefix)
			if prefix != "" && !strings.HasPrefix(lc, lcp) {
				continue
			}

			kvp := strings.SplitN(v, "=", 2)
			if len(kvp) != 2 {
				continue
			}

			key := strings.Replace(strings.TrimSpace(kvp[0]), prefix, "", 1)
			value := strings.TrimSpace(kvp[1])

			values[key] = value
		}
	}

	return values, nil
}
