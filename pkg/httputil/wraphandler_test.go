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
var _ HandlerInterface = (*noopHandlerInterface)(nil)

// noopHandlerInterface is a no-op implementation of [HandlerInterface].
type noopHandlerInterface struct{}

// Handle implements [HandlerInterface].
func (*noopHandlerInterface) Handle(pattern string, handler http.Handler) {}

// HandleFunc implements [HandlerInterface].
func (*noopHandlerInterface) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
}

// ServeHTTP implements [HandlerInterface].
func (*noopHandlerInterface) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

func TestWrapHandler(t *testing.T) {
	type args struct {
		mux         HandlerInterface
		middlewares []MiddlewareFunc
	}
	tests := []struct {
		name string
		args args
		want HandlerInterface
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

func TestHandler_Handle(t *testing.T) {
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

func TestHandler_HandleFunc(t *testing.T) {
	// mux := http.NewServeMux()
	noopHandlerFunc := func(w http.ResponseWriter, r *http.Request) {}
	type args struct {
		pattern string
		handler func(http.ResponseWriter, *http.Request)
	}
	tests := []struct {
		name string
		h    *mwHandler
		args args
	}{
		{"default", &mwHandler{http.NewServeMux(), []MiddlewareFunc{noopMiddlewareFunc}}, args{"GET /here", noopHandlerFunc}},
		{"nilMiddlewares", &mwHandler{http.NewServeMux(), nil}, args{"GET /here", noopHandlerFunc}},
		{"emptyMiddlewares", &mwHandler{http.NewServeMux(), []MiddlewareFunc{}}, args{"GET /here", noopHandlerFunc}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.h.HandleFunc(tt.args.pattern, tt.args.handler)
		})
	}
}

func TestHandler_ServeHTTP(t *testing.T) {
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
				mux:         &noopHandlerInterface{},
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
