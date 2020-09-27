package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/random-guys/go-siber"
	"github.com/rs/zerolog"
)

// AttachLogger attaches a new zerolog.Logger to each new HTTP request.
// Stolen from https://github.com/rs/zerolog/blob/master/hlog/hlog.go
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

// TrackRequest updates a future log entry with the request parameters such as request ID and headers.
func TrackRequest() func(http.Handler) http.Handler {
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
					Interface("request_headers", formattedHeaders)
			})

			requestBody := siber.ReadBody(r)

			if len(requestBody) != 0 {
				log.UpdateContext(func(ctx zerolog.Context) zerolog.Context {
					buffer := new(bytes.Buffer)

					if err := json.Compact(buffer, requestBody); err != nil {
						panic(err)
					}

					return ctx.RawJSON("request", buffer.Bytes())
				})
			}

			next.ServeHTTP(w, r)
		})
	}
}

func TrackResponse() func(http.Handler) http.Handler {
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
						Interface("response_headers", formatHeaders(ww.Header()))
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
