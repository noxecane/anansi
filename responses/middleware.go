package responses

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// ResponseTime adds a "X-Response-Time" header once the handler writes the header
// of the response. Ensure to Use this middleware before any middleware that Write.
func ResponseTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := newWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
	})
}

// RequestDuration adds a middleware to observe request durations and report to prometheus
// with the fully qualified name http_request_duration_seconds and labels for the status codes,
// method and the request path. Note that this handler only works if ResponseTime middleware is
// already Used. Also ensure it's used before any middleware that calls ResponseWriter.Write.
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
				tw, ok := w.(TimedResponseWriter)
				if ok {
					hist.
						WithLabelValues(strconv.Itoa(tw.Code()), r.Method, r.URL.String()).
						Observe(float64(tw.Duration().Milliseconds()))
				}
			}()
		})
	}
}
