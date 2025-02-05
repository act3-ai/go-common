package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// noopMiddlewareFunc is no-op [MiddlewareFunc].
var noopMiddlewareFunc = func(next http.Handler) http.Handler { return next }

// interface check
var _ Router = (*noopHandler)(nil)

// noopHandler is a no-op implementation of [Router].
type noopHandler struct{}

// Handle implements [Router].
func (*noopHandler) Handle(pattern string, handler http.Handler) {}

// HandleFunc implements [Router].
func (*noopHandler) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
}

// ServeHTTP implements [Router].
func (*noopHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

func TestWrapHandler(t *testing.T) {
	type args struct {
		mux         ServeMuxer
		middlewares []MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want ServeMuxer
	}{
		{"default",
			args{
				http.NewServeMux(),
				[]MiddlewareFunc{noopMiddlewareFunc},
			},
			&mwHandler{http.NewServeMux(), []MiddlewareFunc{noopMiddlewareFunc}}},
		{"append",
			args{
				&mwHandler{
					http.NewServeMux(),
					[]MiddlewareFunc{noopMiddlewareFunc},
				},
				[]MiddlewareFunc{noopMiddlewareFunc},
			},
			&mwHandler{
				http.NewServeMux(),
				[]MiddlewareFunc{noopMiddlewareFunc, noopMiddlewareFunc},
			},
		},
		{"emptyArgs",
			args{nil, nil},
			nil},
		{"nilSlice",
			args{http.NewServeMux(), nil},
			http.NewServeMux()},
		{"emptySlice",
			args{http.NewServeMux(), []MiddlewareFunc{}},
			http.NewServeMux()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapHandler(tt.args.mux, tt.args.middlewares...)
			assert.IsType(t, tt.want, got)
			gotmw, gotok := got.(*mwHandler)
			wantmw, wantok := tt.want.(*mwHandler)
			if gotok && wantok {
				assert.Equal(t, wantmw.mux, gotmw.mux)
				assert.Len(t, gotmw.middlewares, len(wantmw.middlewares))
			}
		})
	}
}

func Test_mwHandler_Handle(t *testing.T) {
	noopHandlerFunc := func(w http.ResponseWriter, r *http.Request) {}
	type args struct {
		pattern string
		handler http.Handler
	}
	tests := []struct {
		name string
		h    *mwHandler
		args args
	}{
		{"default",
			&mwHandler{http.NewServeMux(), []MiddlewareFunc{noopMiddlewareFunc}},
			args{"GET /here", http.HandlerFunc(noopHandlerFunc)}},
		{"nilMiddlewares",
			&mwHandler{http.NewServeMux(), nil},
			args{"GET /here", http.HandlerFunc(noopHandlerFunc)}},
		{"emptyMiddlewares",
			&mwHandler{http.NewServeMux(), []MiddlewareFunc{}},
			args{"GET /here", http.HandlerFunc(noopHandlerFunc)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.h.Handle(tt.args.pattern, tt.args.handler)
		})
	}
}

func Test_mwHandler_ServeHTTP(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		h    *mwHandler
		args args
	}{
		{"default",
			&mwHandler{
				mux:         &noopHandler{},
				middlewares: []MiddlewareFunc{noopMiddlewareFunc},
			},
			args{
				httptest.NewRecorder(),
				httptest.NewRequest(http.MethodGet, "/here", nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.h.ServeHTTP(tt.args.w, tt.args.r)
		})
	}
}
