// Package chiutil defines utility functions for use with the [github.com/go-chi/chi/v5] framework.
package chiutil

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// SetPattern sets the [http.Request.Pattern] field.
//
// As of 10/18/2024, [chi.Router] does not set the [http.Request.Pattern] field itself.
func SetPattern(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Pattern == "" {
			if rctx := chi.RouteContext(r.Context()); rctx != nil {
				routePattern := strings.Join(rctx.RoutePatterns, "")
				routePattern = strings.ReplaceAll(routePattern, "/*/", "/")
				r.Pattern = routePattern
			}
		}
		next.ServeHTTP(w, r)
	})
}
