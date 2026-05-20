package translator

import (
	"github.com/shepard-labs/go-us-uk-english-translator/internal/dict"
	"github.com/shepard-labs/go-us-uk-english-translator/internal/replace"
)

type Direction string

const (
	DirectionAmerican Direction = Direction(dict.DirectionAmerican)
	DirectionBritish  Direction = Direction(dict.DirectionBritish)
)

// Converter is the main entry point for translating English spellings.
type Converter struct {
	dictionary         map[string]string
	skipExclusionZones bool
}

// ConverterOptions allows customizing the converter behavior.
type ConverterOptions struct {
	Direction          Direction
	UserDictionaryPath string
	SkipExclusionZones bool
}

// NewConverter initializes a new converter with the specified direction.
// It loads the built-in dictionary and optionally merges a user dictionary.
// By default, exclusion zones (code blocks, URLs, etc.) are skipped.
func NewConverter(direction Direction) (*Converter, error) {
	return NewConverterWithOptions(ConverterOptions{
		Direction:          direction,
		SkipExclusionZones: true, // Default to true per spec
	})
}

// NewConverterWithOptions initializes a new converter with custom options.
func NewConverterWithOptions(opts ConverterOptions) (*Converter, error) {
	d, err := dict.LoadDictionary(dict.Direction(opts.Direction), opts.UserDictionaryPath)
	if err != nil {
		return nil, err
	}

	return &Converter{
		dictionary:         d,
		skipExclusionZones: opts.SkipExclusionZones,
	}, nil
}

// Convert processes the input text, applying the configured translations,
// and returns the converted text along with the number of replacements made.
func (c *Converter) Convert(input string, filename string) (string, int) {
	return replace.Convert(input, c.dictionary, c.skipExclusionZones, filename)
}
