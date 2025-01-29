package httputil

import (
	"net/http"
)

// HandlerInterface represents an HTTP request handler.
type HandlerInterface interface {
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// MiddlewareFunc is an alias for middleware functions.
type MiddlewareFunc = func(http.Handler) http.Handler

// WrapHandler wraps a [HandlerInterface] with middlewares.
func WrapHandler(mux HandlerInterface, middlewares ...MiddlewareFunc) HandlerInterface {
	if len(middlewares) == 0 {
		return mux
	}
	if mwmux, ok := mux.(*mwHandler); ok {
		mwmux.middlewares = append(mwmux.middlewares, middlewares...)
		return mwmux
	}
	return &mwHandler{
		mux:         mux,
		middlewares: middlewares,
	}
}

// mwHandler wraps a Handler with the given middleware functions.
type mwHandler struct {
	mux         HandlerInterface
	middlewares []func(http.Handler) http.Handler
}

// Handle implements [HandlerInterface].
func (h *mwHandler) Handle(pattern string, handler http.Handler) {
	h.mux.Handle(pattern, callMiddlewares(handler, h.middlewares))
}

// HandleFunc implements [HandlerInterface].
func (h *mwHandler) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	// Call the mwHandler's implementation of Handle so middlewares are called.
	h.Handle(pattern, http.HandlerFunc(handler))
}

// ServeHTTP implements [HandlerInterface].
func (h *mwHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func callMiddlewares(handler http.Handler, middlewares []MiddlewareFunc) http.Handler {
	for _, m := range middlewares {
		handler = m(handler)
	}
	return handler
}
