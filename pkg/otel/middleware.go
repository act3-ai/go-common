package otel

import (
	"net/http"

	"gitlab.com/act3-ai/asce/go-common/pkg/httputil"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// MiddlewareFunc is an HTTP middleware that creates a route tag
// for the matched request pattern and starts a span for the request.
//
// The underlying router/servemux must set the http.Request.Pattern
// field (net/http.ServeMux does this).
func MiddlewareFunc(next http.Handler) http.Handler {
	return withRouteTagMiddleware(createSpanMiddleware(next))
}

func withRouteTagMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			next = otelhttp.WithRouteTag(r.Pattern, next)
			next.ServeHTTP(w, r)
		})
}

func createSpanMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			span := trace.SpanFromContext(ctx)
			ctx, span = span.TracerProvider().Tracer("handler").Start(ctx, r.Pattern)
			defer span.End()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
}

var _ httputil.MiddlewareFunc = MiddlewareFunc
