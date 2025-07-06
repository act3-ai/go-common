// Package csp contains a representation of Content-Security-Policy header directives.
//
//nolint:var-naming
package csp

//nolint:var-naming

import (
	"maps"
	"net/http"
	"slices"
	"strings"
)

// HeaderKey is the canonical header key.
const HeaderKey = "Content-Security-Policy"

// Directive names.
const (
	BaseURI        = "base-uri"
	ConnectSource  = "connect-src"
	DefaultSource  = "default-src"
	FormAction     = "form-action"
	FrameAncestors = "frame-ancestors"
	ImageSource    = "img-src"
	ScriptSource   = "script-src"
	StyleSource    = "style-src"
	WorkerSource   = "worker-src"
)

// Keywords available for directives.
const (
	KeywordBlob         = "blob:"
	KeywordData         = "data:"
	KeywordNone         = "'none'"
	KeywordSelf         = "'self'"
	KeywordUnsafeHashes = "'unsafe-hashes'"
)

// ContentSecurityPolicy represents Content-Security-Policy header directives
type ContentSecurityPolicy map[string][]string

// Encoded returns the encoded form of the header.
func (policy ContentSecurityPolicy) Encoded() string {
	directives := make([]string, 0, len(policy))
	for _, key := range slices.Sorted(maps.Keys(policy)) {
		directives = append(directives, key+" "+strings.Join(policy[key], " "))
	}
	return strings.Join(directives, "; ") + ";"
}

// Middleware sets the Content-Security-Policy header in the handler's responses.
func (policy ContentSecurityPolicy) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(HeaderKey, policy.Encoded())
		next.ServeHTTP(w, r)
	})
}
