package parsers

import "os"

// FileParser is a configuration parser
// which reads a backing configuration file.
type FileParser struct {
	fp  string
	sep string
}

// NewFileParser returns a new FileParser.
//
// A FileParser reads a backing configuration file
// with key/value pairs, separated by the specified separator.
func NewFileParser(fp string, sep string) *FileParser {
	return &FileParser{fp, sep}
}

// Parse returns the key/value pairs as a map[string]interface{}.
func (p *FileParser) Parse() (map[string]interface{}, error) {
	h, err := os.Open(p.fp)
	if err != nil {
		return nil, err
	}

	defer func() { _ = h.Close() }()

	kvp := NewKeyValueParser(h, WithKeyValueSeparator(p.sep))
	return kvp.Parse()
}
