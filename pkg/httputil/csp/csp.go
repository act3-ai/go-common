// Package csp contains a representation of Content-Security-Policy header directives.
package csp

import (
	"maps"
	"net/http"
	"slices"
	"strings"
)

const (
	// The canonical header key.
	HeaderKey = "Content-Security-Policy"

	// Defined directive names.
	BaseUri        = "base-uri"
	ConnectSrc     = "connect-src"
	DefaultSrc     = "default-src"
	FormAction     = "form-action"
	FrameAncestors = "frame-ancestors"
	ImgSrc         = "img-src"
	ScriptSrc      = "script-src"
	StyleSrc       = "style-src"
	WorkerSrc      = "worker-src"

	// Keywords available for directives.
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
