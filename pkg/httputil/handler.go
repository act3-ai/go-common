package httputil

import (
	"net/http"
)

// Router represents an HTTP request server like [http.ServeMux].
type Router interface {
	Handle(pattern string, handler http.Handler)
	http.Handler
}

// MiddlewareFunc is an alias for middleware functions.
type MiddlewareFunc func(http.Handler) http.Handler

// WrapHandler wraps a [ServeMuxer] with middlewares.
func WrapHandler(mux Router, middlewares ...MiddlewareFunc) Router {
	if len(middlewares) == 0 {
		return mux
	}
	mwars := make([]RouteMiddlewareFunc, len(middlewares))
	for i, mwar := range middlewares {
		mwars[i] = WrapMiddleware(mwar)
	}
	return WrapRouter(mux, mwars...)
}

// WrapMiddleware adapts a [MiddlewareFunc] to a RouteMiddlewareFunc by ignoring the pattern argument
func WrapMiddleware(middleware MiddlewareFunc) RouteMiddlewareFunc {
	return func(pattern string, handler http.Handler) http.Handler {
		return middleware(handler)
	}
}

// RouteMiddlewareFunc modifies a handler as it is registered with a router.
type RouteMiddlewareFunc func(pattern string, handler http.Handler) http.Handler

// WrapRouter wraps a [Router] with RouteMiddleware functions.
func WrapRouter(mux Router, middlewares ...RouteMiddlewareFunc) Router {
	if len(middlewares) == 0 {
		return mux
	}
	return &mwRouter{
		Router:      mux,
		middlewares: middlewares,
	}
}

var _ Router = &mwRouter{}

// mwRouter wraps a ServeMuxer with the given route middleware functions.
type mwRouter struct {
	Router
	middlewares []RouteMiddlewareFunc
}

// Handle implements httputil.ServeMuxer.
func (h *mwRouter) Handle(pattern string, handler http.Handler) {
	for _, mware := range h.middlewares {
		handler = mware(pattern, handler)
	}
	h.Router.Handle(pattern, handler)
}
