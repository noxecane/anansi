package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/random-guys/go-siber/middleware"
)

// RequestDuration adds a middleware to observe request durations and report to prometheus.
// Note that this handler only works if the response time middleware is used. Also ensure
// it's used before any middleware that might send response(whether directly or during defer)
func RequestDuration(reg prometheus.Registerer) func(http.Handler) http.Handler {
	hist := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "http",
		Name:      "request_duration_seconds",
		Help:      "Duration of HTTP requests in seconds",
		Buckets:   prometheus.DefBuckets,
	}, []string{"statusCode", "method", "path"})

	reg.MustRegister(hist)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)

			defer func() {
				tw, ok := w.(middleware.TimedResponseWriter)
				if ok {
					hist.
						WithLabelValues(strconv.Itoa(tw.Code()), r.Method, r.URL.String()).
						Observe(float64(tw.Duration().Milliseconds()))
				}
			}()
		})
	}
}
