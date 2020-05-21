package anansi

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

var kubeProbe = regexp.MustCompile("(?i)kube-probe|prometheus")

// ZeroLogger is a wrapper around zero log for chi
type ZeroLogger struct {
	BaseLog zerolog.Logger
	Env     string
}

// ZeroLogEntry is an info event for request detauis
type ZeroLogEntry struct {
	Log *zerolog.Logger
}

// ZeroMiddleware creates a middleware for logging http Requests
func ZeroMiddleware(log zerolog.Logger, env string) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(ZeroLogger{BaseLog: log, Env: env})
}

// Panic logs final requests that failed with a panic
func (e *ZeroLogEntry) Panic(v interface{}, _ []byte) {
	e.Log.UpdateContext(func(ctx zerolog.Context) zerolog.Context {
		return ctx.
			Str("error", fmt.Sprintf("%+v", v))
	})
}

// Write logs the response metadata for a request
func (e *ZeroLogEntry) Write(status, bytes int, elapsed time.Duration) {
	e.Log.
		Info().
		Int("status", status).
		Int("length", bytes).
		Float64("elapsed", float64(elapsed.Milliseconds())).
		Msg("")
}

// NewLogEntry creates a special log for each request and storing it's request
// info for write logs or panic logs
func (l ZeroLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	newLogger := l.BaseLog.With().Logger()
	entry := &ZeroLogEntry{Log: &newLogger}

	if kubeProbe.MatchString(r.UserAgent()) {
		return entry
	}

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		newLogger.UpdateContext(func(ctx zerolog.Context) zerolog.Context {
			return ctx.Str("id", reqID)
		})
	}

	lowerCaseHeaders := make(map[string]interface{})

	for k, v := range r.Header {
		lowerKey := strings.ToLower(k)
		if len(v) == 0 {
			lowerCaseHeaders[lowerKey] = ""
		} else if len(v) == 1 {
			lowerCaseHeaders[lowerKey] = v[0]
		} else {
			lowerCaseHeaders[lowerKey] = v
		}
	}

	newLogger.UpdateContext(func(ctx zerolog.Context) zerolog.Context {
		return ctx.
			Str("method", r.Method).
			Str("remote_address", r.RemoteAddr).
			Str("url", r.URL.String()).
			Interface("headers", lowerCaseHeaders)
	})

	if l.Env == "dev" {
		requestBody := ext.ReadBody(r)

		if len(requestBody) == 0 {
			return entry
		}

		entry.Log.UpdateContext(func(ctx zerolog.Context) zerolog.Context {
			return ctx.RawJSON("request", ext.CompactJSON(requestBody))
		})
	}

	return entry
}
