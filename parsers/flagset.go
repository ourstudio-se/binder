package parsers

import (
	"github.com/spf13/pflag"
)

type FlagSetParser struct {
	flagSet *pflag.FlagSet
}

func NewFlagSetParser(flagSet *pflag.FlagSet) *FlagSetParser {
	return &FlagSetParser{flagSet}
}

func (p *FlagSetParser) Parse() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	p.flagSet.Visit(func(flag *pflag.Flag) {
		result[flag.Name] = flag.Value.String()
	})

	return result, nil
}
