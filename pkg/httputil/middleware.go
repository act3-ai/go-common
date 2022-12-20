package httputil

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/segmentio/ksuid"
)

type middlewareFunc = func(http.Handler) http.Handler

// contextInstanceKey is how we find the unique instance ID in a context.Context.
type contextInstanceKey struct{}

// InstanceFromContext returns the instance for this request to uniquely identify the request
func InstanceFromContext(ctx context.Context) ksuid.KSUID {
	if v := ctx.Value(contextInstanceKey{}); v != nil {
		return v.(ksuid.KSUID)
	}
	// panic("instance missing from context")
	return ksuid.Nil
}

// TracingMiddleware injects a tracing ID into the context
func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := ksuid.New()
		ctx = context.WithValue(ctx, contextInstanceKey{}, id)
		// Call the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

var _ middlewareFunc = TracingMiddleware

// LoggingMiddleware injects a logger into the context
func LoggingMiddleware(log logr.Logger) middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			path := r.URL.Path
			id := InstanceFromContext(ctx)
			log = log.WithValues("path", path, "qs", r.URL.Query(), "instance", id.String())
			ctx = logr.NewContext(r.Context(), log)
			// Call the next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ServerHeaderMiddleware injects the Server into the response headers
func ServerHeaderMiddleware(server string) middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Server", server)

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

/*
// statusMiddleware logs the status response, install after the LoggingMiddleware
func statusMiddleware(next http.Handler) http.Handler {
	ctx := r.Context()
	log := logr.FromContextOrDiscard(ctx)
	// TODO or use https://github.com/felixge/httpsnoop directly
	return handlers.CustomLoggingHandler(nil, next,
		func(_ io.Writer, params handlers.LogFormatterParams) {
			log.Info("Completed request", "code", params.StatusCode)
		},
	)
}
*/

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration of HTTP requests in seconds.",
	}, []string{"method", "route"})
)

// PrometheusMiddleware records timing metrics
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// call the next handler
		next.ServeHTTP(w, r)

		// This must be done after calling next.ServeHTTP()
		rctx := chi.RouteContext(r.Context())
		routePattern := strings.Join(rctx.RoutePatterns, "")
		routePattern = strings.ReplaceAll(routePattern, "/*/", "/")

		httpDuration.WithLabelValues(r.Method, routePattern).Observe(float64(time.Since(start).Microseconds()) / 1000000)
	})
}

var _ middlewareFunc = PrometheusMiddleware

// RecovererMiddleware is a middleware that recovers from panics, logs the panic (and a
// backtrace), and returns a HTTP 500 (Internal Server Error) status if
// possible. Recoverer prints a request ID if one is provided.
//
// KMT - I am not sure we need this middleware since the golang server already recovers from panics.  It just does not use our logger or return a 500.
func RecovererMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				log := logr.FromContextOrDiscard(r.Context())
				switch t := rvr.(type) {
				case error:
					if errors.Is(t, http.ErrAbortHandler) {
						log.Info("Handler panic-ed", "error", t)
					} else {
						log.Error(t, "Handler panic-ed")
					}
				default:
					log.Error(nil, "Handler panic-ed with unknown error", "value", rvr)
				}
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

var _ middlewareFunc = RecovererMiddleware
