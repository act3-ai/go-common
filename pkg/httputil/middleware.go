package httputil

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/ksuid"

	"gitlab.com/act3-ai/asce/go-common/pkg/logger"
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

// LoggingMiddleware injects a logger into the context.
//
// A previous implementation contained a memory leak because the tracing attributes were always appended to the given logger.
func LoggingMiddleware(log *slog.Logger) middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			path := r.URL.Path
			id := InstanceFromContext(ctx)
			ctx = logger.NewContext(ctx, log.With(
				slog.String("path", path),
				slog.Any("qs", r.URL.Query()),
				slog.String("instance", id.String()),
			))
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
	log := logger.FromContext(ctx)
	// TODO or use https://github.com/felixge/httpsnoop directly
	return handlers.CustomLoggingHandler(nil, next,
		func(_ io.Writer, params handlers.LogFormatterParams) {
			log.Info("Completed request", "code", params.StatusCode)
		},
	)
}
*/

// HTTPDuration is prometheus histogram of the time for the server to handle a HTTP request
// Users need to register this with a prometheus.Registerer
var HTTPDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "http_request_duration_seconds",
	Help:    "Duration of HTTP requests in seconds.",
	Buckets: []float64{0.1, .25, .5, 1, 2.5, 5, 10},
}, []string{"method", "route"})

// PrometheusMiddleware records timing metrics
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// call the next handler
		next.ServeHTTP(w, r)

		// This must be done after calling next.ServeHTTP()
		pattern := strings.TrimPrefix(r.Pattern, r.Method+" ")
		HTTPDuration.WithLabelValues(r.Method, pattern).Observe(time.Since(start).Seconds())
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
		ctx := r.Context()
		defer func() {
			if rvr := recover(); rvr != nil {
				log := logger.FromContext(r.Context())
				switch t := rvr.(type) {
				case error:
					if errors.Is(t, http.ErrAbortHandler) {
						log.InfoContext(ctx, "Handler panic-ed", "error", t)
					} else {
						log.ErrorContext(ctx, "Handler panic-ed", "error", t)
					}
				default:
					log.ErrorContext(ctx, "Handler panic-ed with unknown error", "value", rvr)
				}
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

var _ middlewareFunc = RecovererMiddleware
