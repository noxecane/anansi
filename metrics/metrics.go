package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/random-guys/go-siber/middleware"
)

// RequestDurationHistogram is the histogram for tracking request durations based
// on their status code, method or the specific path.
var RequestDurationHistogram *prometheus.HistogramVec

func init() {
	RequestDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "http",
		Name:      "request_duration_seconds",
		Help:      "Duration of HTTP requests in seconds",
		Buckets:   prometheus.DefBuckets,
	}, []string{"statusCode", "method", "path"})
}

// RequestDuration adds a middleware to observe request durations and report to prometheus.
// It basically observes values using RequestDurationHistogram. Note that this handler only
// works if the response time middleware is used. Also make sure to add the RequestDurationHistogram
// to a registry
func RequestDuration() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			tw, ok := w.(middleware.TimedResponseWriter)
			if ok {
				RequestDurationHistogram.
					WithLabelValues(strconv.Itoa(tw.Code()), r.Method, r.URL.String()).
					Observe(float64(tw.Duration().Milliseconds()))
			}
		}()
	})
}
