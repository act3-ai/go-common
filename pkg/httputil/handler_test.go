package httputil_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
)

func pathMW(pattern string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "handled by", pattern)
		next.ServeHTTP(w, r)
	})
}

func basicMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "basic")
		next.ServeHTTP(w, r)
	})
}

func Test_WrapRouter(t *testing.T) {
	tests := []struct {
		name        string
		middlewares []httputil.RouteMiddlewareFunc
		req         *http.Request
		wantBody    string
	}{
		{
			name:        "path-middleware",
			middlewares: []httputil.RouteMiddlewareFunc{pathMW},
			req:         httptest.NewRequest(http.MethodGet, "/here", nil),
			wantBody:    "handled by GET /here\nDone\n",
		},
		{
			name: "basic-middleware",
			middlewares: []httputil.RouteMiddlewareFunc{
				httputil.WrapMiddleware(basicMW),
			},
			req:      httptest.NewRequest(http.MethodGet, "/here", nil),
			wantBody: "basic\nDone\n",
		},
		{
			name: "mixed",
			middlewares: []httputil.RouteMiddlewareFunc{
				pathMW,
				httputil.WrapMiddleware(basicMW),
				pathMW,
			},
			req:      httptest.NewRequest(http.MethodGet, "/here", nil),
			wantBody: "handled by GET /here\nbasic\nhandled by GET /here\nDone\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := httputil.WrapRouter(&http.ServeMux{}, tt.middlewares...)

			router.Handle("GET /here", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "Done")
			}))

			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.req)
			// Test that the output is what we expect
			body, err := io.ReadAll(w.Result().Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantBody, string(body))
		})
	}
}
