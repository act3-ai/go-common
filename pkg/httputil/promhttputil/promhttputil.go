// Package promhttputil contains a prometheus metrics middleware, relocated from httputil.
package promhttputil

import (
	"net/http"
	"strings"
	"time"

	"github.com/act3-ai/go-common/pkg/httputil"
	"github.com/prometheus/client_golang/prometheus"
)

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

var _ httputil.MiddlewareFunc = PrometheusMiddleware
