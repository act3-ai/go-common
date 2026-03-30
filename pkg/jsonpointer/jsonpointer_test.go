package jsonpointer

import (
	"math"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/act3-ai/go-common/pkg/testutil"
)

func TestEncoding(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		escaped string
	}{
		{
			name:    "no special characters",
			raw:     "hello",
			escaped: "hello",
		},
		{
			name:    "empty string",
			raw:     "",
			escaped: "",
		},
		{
			name:    "tilde only",
			raw:     "~",
			escaped: "~0",
		},
		{
			name:    "slash only",
			raw:     "/",
			escaped: "~1",
		},
		{
			name:    "both tilde and slash",
			raw:     "~/",
			escaped: "~0~1",
		},
		{
			name:    "slash then tilde",
			raw:     "/~",
			escaped: "~1~0",
		},
		{
			name:    "multiple tildes",
			raw:     "~~",
			escaped: "~0~0",
		},
		{
			name:    "multiple slashes",
			raw:     "//",
			escaped: "~1~1",
		},
		{
			name:    "complex string with path-like structure",
			raw:     "a/b~c",
			escaped: "a~1b~0c",
		},
		{
			name:    "string with mixed characters",
			raw:     "hello/world~test",
			escaped: "hello~1world~0test",
		},
		{
			name:    "RFC6901 example - a/b",
			raw:     "a/b",
			escaped: "a~1b",
		},
		{
			name:    "RFC6901 example - m~n",
			raw:     "m~n",
			escaped: "m~0n",
		},
		{
			name:    "edge case - ~01 sequence",
			raw:     "~01",
			escaped: "~001",
		},
		{
			name:    "edge case - ~10 sequence",
			raw:     "~10",
			escaped: "~010",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("Escape", func(t *testing.T) {
				got := Escape(tt.raw)
				assert.Equal(t, tt.escaped, got)
			})
			t.Run("Unescape", func(t *testing.T) {
				got := Unescape(tt.escaped)
				assert.Equal(t, tt.raw, got)
			})
		})
	}
}

func TestTokens(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []string
		pointer string
	}{
		{
			name:    "no tokens",
			tokens:  nil,
			pointer: "",
		},
		{
			name:    "empty slice",
			tokens:  []string{},
			pointer: "",
		},
		{
			name:    "single empty token",
			tokens:  []string{""},
			pointer: "/",
		},
		{
			name:    "no special characters",
			tokens:  []string{"a", "b", "0"},
			pointer: "/a/b/0",
		},
		{
			name:    "all the tricks",
			tokens:  []string{"hello", "~", "/", "~/", "/~", "~~", "//", "a/b~c", "~01", "~10"},
			pointer: "/hello/~0/~1/~0~1/~1~0/~0~0/~1~1/a~1b~0c/~001/~010",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("FromTokens", func(t *testing.T) {
				got := FromTokens(tt.tokens...)
				assert.Equal(t, tt.pointer, got)
			})
			t.Run("ToTokens", func(t *testing.T) {
				got := ToTokens(tt.pointer)
				if len(tt.tokens) == 0 {
					assert.Len(t, got, 0)
				} else {
					assert.Equal(t, tt.tokens, got)
				}
			})
			t.Run("Tokens", func(t *testing.T) {
				got := slices.Collect(Tokens(tt.pointer))
				if len(tt.tokens) == 0 {
					assert.Len(t, got, 0)
				} else {
					assert.Equal(t, tt.tokens, got)
				}
			})
		})
	}

	t.Run("early iterator exit", func(t *testing.T) {
		assert.NotPanics(t, func() {
			count := 0
			for range Tokens("/1/2/3") {
				count++
				if count == 1 {
					break
				}
			}
		})
	})
}

func TestPopToken(t *testing.T) {
	tests := []struct {
		name      string
		pointer   string
		token     string
		remainder string
		ok        bool
	}{
		{
			name:      "empty string",
			pointer:   "",
			token:     "",
			remainder: "",
			ok:        false,
		},
		{
			name:      "just a slash",
			pointer:   "/",
			token:     "",
			remainder: "",
			ok:        true,
		},
		{
			name:      "just a non-slash",
			pointer:   "a",
			token:     "",
			remainder: "",
			ok:        false,
		},
		{
			name:      "final token",
			pointer:   "/a~1b",
			token:     "a/b",
			remainder: "",
			ok:        true,
		},
		{
			name:      "remainder",
			pointer:   "/a~1b/foo/bar/a~1b",
			token:     "a/b",
			remainder: "/foo/bar/a~1b",
			ok:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, remainder, ok := PopToken(tt.pointer)
			assert.Equal(t, tt.token, token)
			assert.Equal(t, tt.remainder, remainder)
			assert.Equal(t, tt.ok, ok)
		})
	}
}

func TestParseArrayIndexToken(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		index      int
		isNewIndex bool
		wantErr    string
	}{
		{
			name:       "empty string",
			token:      "",
			index:      0,
			isNewIndex: false,
			wantErr:    `invalid array index: parsing "": empty value`,
		},
		{
			name:       "non-numeric string",
			token:      "this",
			index:      0,
			isNewIndex: false,
			wantErr:    `invalid array index: parsing "this": invalid syntax`,
		},
		{
			name:       "dash token",
			token:      "-",
			index:      0,
			isNewIndex: true,
			wantErr:    "",
		},
		{
			name:       "dash token with leading space",
			token:      " -",
			index:      0,
			isNewIndex: false,
			wantErr:    `invalid array index: parsing " -": invalid syntax`,
		},
		{
			name:       "dash token with trailing space",
			token:      "- ",
			index:      0,
			isNewIndex: false,
			wantErr:    `invalid array index: parsing "- ": invalid syntax`,
		},
		{
			name:       "zero",
			token:      "0",
			index:      0,
			isNewIndex: false,
			wantErr:    ``,
		},
		{
			name:       "zero with leading space",
			token:      " 0",
			index:      0,
			isNewIndex: false,
			wantErr:    `invalid array index: parsing " 0": invalid syntax`,
		},
		{
			name:       "zero with trailing space",
			token:      "0 ",
			index:      0,
			isNewIndex: false,
			wantErr:    `invalid array index: parsing "0 ": invalid syntax`,
		},
		{
			name:       "number",
			token:      "17",
			index:      17,
			isNewIndex: false,
			wantErr:    ``,
		},
		{
			name:       "number with leading zero",
			token:      "017",
			index:      0,
			isNewIndex: false,
			wantErr:    `invalid array index: parsing "017": invalid syntax`,
		},
		{
			name:       "number with multiple leading zeros",
			token:      "0017",
			index:      0,
			isNewIndex: false,
			wantErr:    `invalid array index: parsing "0017": invalid syntax`,
		},
		{
			name:       "number with leading space",
			token:      " 17",
			index:      0,
			isNewIndex: false,
			wantErr:    `invalid array index: parsing " 17": invalid syntax`,
		},
		{
			name:       "number with trailing space",
			token:      "17 ",
			index:      0,
			isNewIndex: false,
			wantErr:    `invalid array index: parsing "17 ": invalid syntax`,
		},
		{
			name:       "MaxUint64 overflow",
			token:      strconv.FormatUint(math.MaxUint64, 10) + "0",
			index:      0,
			isNewIndex: false,
			wantErr:    `invalid array index: strconv.Atoi: parsing "184467440737095516150": value out of range`,
		},
		{
			name:       "MaxInt overflow",
			token:      strconv.FormatUint(math.MaxInt+1, 10),
			index:      0,
			isNewIndex: false,
			wantErr:    `invalid array index: strconv.Atoi: parsing "9223372036854775808": value out of range`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, isNewIndex, err := ParseArrayIndexToken(tt.token)
			assert.Equal(t, tt.index, index, "ParseArrayIndexToken() index")
			assert.Equal(t, tt.isNewIndex, isNewIndex, "ParseArrayIndexToken() isNewIndex")
			testutil.AssertErrorIf(t, tt.wantErr != "", err, "ParseArrayIndexToken() error")
			if tt.wantErr != "" && err != nil {
				assert.Equal(t, tt.wantErr, err.Error(), "ParseArrayIndexToken() error content")
			}
		})
	}
}
