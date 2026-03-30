// package jsonpointer is an implementation of JSON Pointers as defined by [RFC6901].
//
// [RFC6901]: https://datatracker.ietf.org/doc/html/rfc6901
package jsonpointer

import (
	"errors"
	"fmt"
	"iter"
	"regexp"
	"strconv"
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

var (
	// ErrInvalidArrayIndex is returned when a token is an invalid index.
	ErrInvalidArrayIndex = errors.New("invalid array index")
)

// ParseArrayIndexToken parses the unescaped token from a JSON Pointer as a
// JSON array index.
//
// Returns the parsed index, a boolean indicating if the token is the
// sentinel value "-" which references a (nonexistent) member after the last
// array element, and any error encountered during parsing.
//
// RFC6901 states that if the currently referenced value is a JSON array,
// the reference token MUST contain either:
//
//   - characters comprised of digits (see ABNF below; note that
//     leading zeros are not allowed) that represent an unsigned
//     base-10 integer value, making the new referenced value the
//     array element with the zero-based index identified by the
//     token, or
//   - exactly the single character "-", making the new referenced
//     value the (nonexistent) member after the last array element.
func ParseArrayIndexToken(token string) (index int, isNewIndex bool, err error) {
	switch token {
	case "":
		// Empty string is an error
		return 0, false, fmt.Errorf("%w: parsing %q: empty value", ErrInvalidArrayIndex, token)
	case "-":
		// Reference the (nonexistent) member after the last array element
		return 0, true, nil
	default:
		// Parse as digits using stricter function than strconv.Atoi
		index, err := parseUnsignedIntegerStrict(token)
		if err != nil {
			return 0, false, fmt.Errorf("%w: %w", ErrInvalidArrayIndex, err)
		}
		return index, false, nil
	}
}

// parseUnsignedIntegerStrict parses an unsigned integer without leading zeros.
func parseUnsignedIntegerStrict(s string) (int, error) {
	if len(regexIntNoLeadingZero.FindStringIndex(s)) == 0 {
		return 0, fmt.Errorf("parsing %q: %w", s, strconv.ErrSyntax)
	}
	// Parse into int
	return strconv.Atoi(s)
}

// regexIntNoLeadingZero validates strings as unsigned integers without leading zeros.
var regexIntNoLeadingZero = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)
