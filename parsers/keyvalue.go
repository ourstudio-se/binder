package parsers

import (
	"bufio"
	"io"
	"strings"
)

const defaultKeyValueSeparator string = ":"

type KeyValueParser struct {
	r         io.Reader
	separator string
}

type KeyValueParserOption func(*KeyValueParser)

func WithKeyValueSeparator(sep string) KeyValueParserOption {
	return func(kvp *KeyValueParser) {
		kvp.separator = sep
	}
}

func NewKeyValueParser(r io.Reader, opts ...KeyValueParserOption) *KeyValueParser {
	kvp := &KeyValueParser{r, defaultKeyValueSeparator}

	for _, opt := range opts {
		opt(kvp)
	}

	return kvp
}

func (p *KeyValueParser) Parse() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	scanner := bufio.NewScanner(p.r)

	for scanner.Scan() {
		kv := strings.SplitN(scanner.Text(), p.separator, 2)

		if len(kv) != 2 || kv[0] == "" {
			continue
		}

		result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}

	return result, nil
}
