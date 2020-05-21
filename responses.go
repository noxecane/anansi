package anansi

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

// SendJSON response writes a JSON encoded version of `v` to the writer, making
// sure what deserves to be logged gets logged
func SendJSON(r *http.Request, w http.ResponseWriter, code int, v interface{}) {
	raw, _ := json.Marshal(v)
	entry := middleware.GetLogEntry(r).(*logs.ZeroLogEntry)

	// log API responses
	if v != nil {
		entry.Log.UpdateContext(func(ctx zerolog.Context) zerolog.Context {
			return ctx.RawJSON("response", ext.CompactJSON(raw))
		})
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_, _ = w.Write(raw)
}

// SendSuccess sends a JSON success message with status code 200
func SendSuccess(r *http.Request, w http.ResponseWriter, v interface{}) {
	SendJSON(r, w, 200, v)
}

// SendError sends a JSON error message
func SendError(r *http.Request, w http.ResponseWriter, err APIError) {
	SendJSON(r, w, err.Code, err)
}

// CompactJSON removes insignificant space from JSON
func CompactJSON(raw []byte) []byte {
	buffer := new(bytes.Buffer)

	if err := json.Compact(buffer, raw); err != nil {
		panic(err)
	}

	return buffer.Bytes()
}
