// package jsonpointer is an implementation of JSON Pointers as defined by [RFC6901].
//
// [RFC6901]: https://datatracker.ietf.org/doc/html/rfc6901
package jsonpointer

import (
	"iter"
	"strings"

	"github.com/act3-ai/go-common/pkg/basicenc"
)

// Tokens produces an iterator over the unescaped tokens of a JSON Pointer.
func Tokens(p string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for token := range strings.SplitSeq(p, "/") {
			if !yield(Unescape(token)) {
				return
			}
		}
	}
}

// ToTokens produces the unescaped tokens of a JSON Pointer.
func ToTokens(p string) []string {
	// Split into tokens
	tokens := strings.Split(p, "/")
	// Unescape each token value
	for i, token := range tokens {
		tokens[i] = Unescape(token)
	}
	return tokens
}

// FromTokens constructs a JSON Pointer from tokens
// by escaping them and adding separators.
func FromTokens(tokens ...string) string {
	w := &strings.Builder{}
	for _, token := range tokens {
		w.WriteString("/" + Escape(token))
	}
	return w.String()
}

// Escape escapes a string for use as a reference token in a JSON pointer according to RFC6901.
// It replaces "~" with "~0" and "/" with "~1" as required by the specification.
// This function should be used when constructing JSON pointers from string values that may contain
// these special characters.
func Escape(s string) string {
	return encoder.Encode(s)
}

// Unescape unescapes a string from the escaped form according to RFC6901.
// It replaces "~0" with "~" and "~1" with "/" as required by the specification.
// This function should be used when evaluating parts of JSON pointers
// that may contain these special characters.
func Unescape(s string) string {
	return encoder.Decode(s)
}

// encoder stores the encoding rules for JSON Pointers.
var encoder = basicenc.NewBasicEncoding([][2]string{
	{"~", "~0"},
	{"/", "~1"},
}...)
