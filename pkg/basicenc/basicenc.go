package basicenc

import (
	"slices"
	"strings"
)

// BasicEncoding represents a basic encoding using replacements.
type BasicEncoding struct {
	encoder *strings.Replacer
	decoder *strings.Replacer
}

// NewBasicEncoding creates a basic encoding using the given encoding
// replacements to encode/decode values.
func NewBasicEncoding(encodings ...[2]string) *BasicEncoding {
	return &BasicEncoding{
		encoder: newBasicEncoder(encodings...),
		decoder: newBasicDecoder(encodings...),
	}
}

// Encode encodes the value.
func (enc *BasicEncoding) Encode(value string) string {
	return enc.encoder.Replace(value)
}

// Decode decodes the value.
func (enc *BasicEncoding) Decode(value string) string {
	return enc.decoder.Replace(value)
}

// newBasicEncoder produces a [strings.Replacer] that
// encodes values by performing the given replacements.
func newBasicEncoder(encodings ...[2]string) *strings.Replacer {
	values := make([]string, 0, len(encodings)*2)
	// Add replacements in forward order
	for _, replace := range encodings {
		values = append(values, replace[0], replace[1])
	}
	return strings.NewReplacer(values...)
}

// newBasicDecoder produces a [strings.Replacer] that
// decodes values by reversing the given replacements.
func newBasicDecoder(encodings ...[2]string) *strings.Replacer {
	values := make([]string, 0, len(encodings)*2)
	// Add replacements in reverse order
	for _, replace := range slices.Backward(encodings) {
		values = append(values, replace[1], replace[0]) // reverse the replacement
	}
	return strings.NewReplacer(values...)
}
