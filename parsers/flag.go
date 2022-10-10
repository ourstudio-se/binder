package parsers

import "flag"

type FlagParser struct{}

func NewFlagParser() *FlagParser {
	return &FlagParser{}
}

func (p *FlagParser) Parse() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	flag.Visit(func(flag *flag.Flag) {
		result[flag.Name] = flag.Value.String()
	})

	return result, nil
}
