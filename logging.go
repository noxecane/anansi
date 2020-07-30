package anansi

import (
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

var kubeProbe = regexp.MustCompile("(?i)kube-probe|prometheus")

func NewLogger(service string) zerolog.Logger {
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	return zerolog.
		New(os.Stdout).
		With().
		Timestamp().
		Str("service", service).
		Str("host", host).
		Logger()
}

// stolen from https://github.com/rs/zerolog/blob/master/hlog/hlog.go
func AttachLogger(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a copy of the logger (including internal context slice)
			// to prevent data race when using UpdateContext.
			l := log.With().Logger()
			r = r.WithContext(l.WithContext(r.Context()))
			next.ServeHTTP(w, r)
		})
	}
}

// RequestLogger updates a future log entry with the request parameters such as request ID and headers.
func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := zerolog.Ctx(r.Context())

			if reqID := middleware.GetReqID(r.Context()); reqID != "" {
				log.UpdateContext(func(ctx zerolog.Context) zerolog.Context {
					return ctx.Str("id", reqID)
				})
			}

			formattedHeaders := formatHeaders(r.Header)

			log.UpdateContext(func(ctx zerolog.Context) zerolog.Context {
				return ctx.
					Str("method", r.Method).
					Str("remote_address", r.RemoteAddr).
					Str("url", r.URL.String()).
					Interface("headers", formattedHeaders)
			})

			requestBody := ReadBody(r)

			if len(requestBody) == 0 {
				return
			}

			log.UpdateContext(func(ctx zerolog.Context) zerolog.Context {
				return ctx.RawJSON("request", CompactJSON(requestBody))
			})

			next.ServeHTTP(w, r)
		})
	}
}

func ResponseLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := zerolog.Ctx(r.Context())

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				log.UpdateContext(func(ctx zerolog.Context) zerolog.Context {
					return ctx.
						Int("status", ww.Status()).
						Int("length", ww.BytesWritten()).
						Float64("elapsed", float64(time.Since(t1).Milliseconds())).
						Interface("headers", formatHeaders(ww.Header()))
				})
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

func formatHeaders(headers http.Header) map[string]interface{} {
	lowerCaseHeaders := make(map[string]interface{})

	for k, v := range headers {
		lowerKey := strings.ToLower(k)
		if len(v) == 0 {
			lowerCaseHeaders[lowerKey] = ""
		} else if len(v) == 1 {
			lowerCaseHeaders[lowerKey] = v[0]
		} else {
			lowerCaseHeaders[lowerKey] = v
		}
	}

	return lowerCaseHeaders
}
