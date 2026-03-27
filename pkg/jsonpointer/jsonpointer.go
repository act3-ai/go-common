// package jsonpointer is an implementation of JSON Pointers as defined by [RFC6901].
//
// [RFC6901]: https://datatracker.ietf.org/doc/html/rfc6901
package jsonpointer

import (
	"iter"
	"strings"

	"github.com/act3-ai/go-common/pkg/basicenc"
)

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
})

// FromTokens constructs a JSON Pointer from tokens
// by escaping them and adding separators.
func FromTokens(tokens ...string) string {
	w := strings.Builder{}
	for _, token := range tokens {
		w.WriteString("/" + Escape(token))
	}
	return w.String()
}

// Tokens produces an iterator over the unescaped tokens of a JSON Pointer.
// If p is not a valid JSON Pointer, the iterator will not yield any values.
func Tokens(p string) iter.Seq[string] {
	return func(yield func(string) bool) {
		if p == "" || !IsValid(p) {
			return
		}
		for token := range strings.SplitSeq(p[1:], "/") {
			if !yield(Unescape(token)) {
				return
			}
		}
	}
}

// ToTokens produces the unescaped tokens of a JSON Pointer.
// If p is not a valid JSON Pointer, the returned slice will be nil.
func ToTokens(p string) []string {
	if p == "" || !IsValid(p) {
		return nil
	}
	// Split into tokens
	tokens := strings.Split(p[1:], "/")
	// Unescape each token value
	for i, token := range tokens {
		tokens[i] = Unescape(token)
	}
	return tokens
}

// PopToken pops the first token from the front of the JSON Pointer string.
//
// Returns the unescaped value of the token, the remainder of the JSON Pointer
// as a valid JSON Pointer if there are more tokens, and a boolean indicating
// if a token could be produced from the JSON Pointer.
func PopToken(p string) (token, remainder string, ok bool) {
	switch p {
	case "":
		// Empty string terminates
		return "", "", false
	case "/":
		// Special case: "/"
		return "", "", true
	}
	// Validate that the first character is the separator
	if p[0] != '/' {
		// Early return for malformed JSON Pointer
		return "", "", false
	}
	// Split around next instance of the separator
	token, remainder, found := strings.Cut(p[1:], "/")
	if found {
		// Restore the separator so the remainder is a valid JSON Pointer
		remainder = "/" + remainder
	}
	// Unescape the token
	token = Unescape(token)
	return token, remainder, true
}

// IsValid reports if the string is a valid JSON Pointer.
func IsValid(p string) bool {
	// From RFC6901:
	// "" evaluates to the entire JSON document
	// "/" refers to the value of the field named ""
	return p == "" || p[0] == '/'
}
