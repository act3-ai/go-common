package httputil

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/act3-ai/go-common/pkg/logger"
)

// contextInstanceKey is how we find the unique instance ID in a context.Context.
type contextInstanceKey struct{}

// InstanceFromContext returns the instance for this request to uniquely identify the request
func InstanceFromContext(ctx context.Context) uuid.UUID {
	if v := ctx.Value(contextInstanceKey{}); v != nil {
		return v.(uuid.UUID)
	}
	// panic("instance missing from context")
	return uuid.Nil
}

// TracingMiddleware injects a tracing ID into the context
func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id, err := uuid.NewV7()
		if err != nil {
			log := logger.FromContext(r.Context())
			log.ErrorContext(ctx, "Failed to generate UUID", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		ctx = context.WithValue(ctx, contextInstanceKey{}, id)
		// Call the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

var _ MiddlewareFunc = TracingMiddleware

// LoggingMiddleware injects a logger into the context.
//
// A previous implementation contained a memory leak because the tracing attributes were always appended to the given logger.
func LoggingMiddleware(log *slog.Logger) MiddlewareFunc {
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
func ServerHeaderMiddleware(server string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Server", server)

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

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

var _ MiddlewareFunc = RecovererMiddleware

// TimeoutMiddleware cancels the request context after a given timeout duration
func TimeoutMiddleware(next http.Handler, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer func() {
			cancel()
			if ctx.Err() == context.DeadlineExceeded {
				w.WriteHeader(http.StatusGatewayTimeout)
			}
		}()

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// AllowContentTypeMiddleware enforces a allowlist of request Content-Types
func AllowContentTypeMiddleware(next http.Handler, contentTypes ...string) http.Handler {
	// format contentype strings
	allowedContentTypes := make(map[string]struct{}, len(contentTypes))
	for _, ctype := range contentTypes {
		allowedContentTypes[strings.TrimSpace(strings.ToLower(ctype))] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength == 0 {
			next.ServeHTTP(w, r)
			return
		}

		s := strings.ToLower(strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0]))
		if _, ok := allowedContentTypes[s]; ok {
			next.ServeHTTP(w, r)
			return
		}
		w.WriteHeader(http.StatusUnsupportedMediaType)
	})
}
