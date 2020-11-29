package siber

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

type jsendSuccess struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func Send(w http.ResponseWriter, code int, data []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	w.WriteHeader(code)
	_, err := w.Write(data)
	if err != nil {
		panic(err)
	}
}

// SendSuccess sends a JSON success message with status code 200
func SendSuccess(r *http.Request, w http.ResponseWriter, v interface{}) {
	log := zerolog.Ctx(r.Context())
	raw := getJSON(log, jsendSuccess{http.StatusOK, v})

	Send(w, http.StatusOK, raw)

	log.Info().
		Int("status", http.StatusOK).
		Int("length", len(raw)).
		Interface("response_headers", NormaliseHeader(w.Header())).
		Msg("")
}

// SendError sends a JSON error message
func SendError(r *http.Request, w http.ResponseWriter, err JSendError) {
	log := zerolog.Ctx(r.Context())
	raw := getJSON(log, err)

	Send(w, err.Code, raw)

	log.Err(err).
		Int("status", err.Code).
		Int("length", len(raw)).
		Interface("response_headers", NormaliseHeader(w.Header())).
		Msg("")
}

func getJSON(log *zerolog.Logger, v interface{}) []byte {
	raw, _ := json.Marshal(v)

	// log API responses
	if v != nil {
		log.UpdateContext(func(ctx zerolog.Context) zerolog.Context {
			buffer := new(bytes.Buffer)

			if err := json.Compact(buffer, raw); err != nil {
				panic(err)
			}

			return ctx.RawJSON("response", buffer.Bytes())
		})
	}

	return raw
}

// NormaliseHeader extracts the headers into a map and converts all single value
// headers to the value directly rather than a slice of single value string
func NormaliseHeader(headers http.Header) map[string]interface{} {
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
