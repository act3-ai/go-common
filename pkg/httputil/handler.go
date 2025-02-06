package httputil

import (
	"net/http"
)

// Router represents an HTTP request handler.
type Router interface {
	Handle(pattern string, handler http.Handler)
}

// ServeMuxer represents an HTTP request server like [http.ServeMuxer].
type ServeMuxer interface {
	Router
	http.Handler
}

// MiddlewareFunc is an alias for middleware functions.
type MiddlewareFunc = func(http.Handler) http.Handler

// WrapHandler wraps a [Router] with middlewares.
func WrapHandler(mux ServeMuxer, middlewares ...MiddlewareFunc) ServeMuxer {
	if len(middlewares) == 0 {
		return mux
	}
	if mwmux, ok := mux.(*mwHandler); ok {
		mwmux.middlewares = append(mwmux.middlewares, middlewares...)
		return mwmux
	}
	return &mwHandler{
		ServeMuxer:  mux,
		middlewares: middlewares,
	}
}

// mwHandler wraps a Handler with the given middleware functions.
type mwHandler struct {
	ServeMuxer
	middlewares []MiddlewareFunc
}

// Handle implements [Router].
func (h *mwHandler) Handle(pattern string, handler http.Handler) {
	for _, mware := range h.middlewares {
		handler = mware(handler)
	}
	h.ServeMuxer.Handle(pattern, handler)
}

// RouteMiddlewareFunc modifies a handler as it is registered with a router.
type RouteMiddlewareFunc = func(pattern string, handler http.Handler) http.Handler

// WrapRouter wraps a ServeMuxer with RouteMiddleware functions.
func WrapRouter(mux ServeMuxer, middlewares ...RouteMiddlewareFunc) ServeMuxer {
	if len(middlewares) == 0 {
		return mux
	}
	if mwmux, ok := mux.(*mwRouter); ok {
		mwmux.middlewares = append(mwmux.middlewares, middlewares...)
		return mwmux
	}
	return &mwRouter{
		ServeMuxer:  mux,
		middlewares: middlewares,
	}
}

var _ Router = &mwRouter{}

// mwRouter wraps a ServeMuxer with the given route middleware functions.
type mwRouter struct {
	ServeMuxer
	middlewares []RouteMiddlewareFunc
}

// Handle implements httputil.ServeMuxer.
func (h *mwRouter) Handle(pattern string, handler http.Handler) {
	for _, mware := range h.middlewares {
		handler = mware(pattern, handler)
	}
	h.ServeMuxer.Handle(pattern, handler)
}
