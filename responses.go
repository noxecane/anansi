package anansi

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// SendSuccess sends a JSON success message with status code 200
func SendSuccess(r *http.Request, w http.ResponseWriter, v interface{}) {
	raw := getJSON(r, v)

	log.Info().Msg("")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

// SendError sends a JSON error message
func SendError(r *http.Request, w http.ResponseWriter, err APIError) {
	raw := getJSON(r, err)

	log.Err(err).Msg("")

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}

func getJSON(r *http.Request, v interface{}) []byte {
	raw, _ := json.Marshal(v)

	log := zerolog.Ctx(r.Context())

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
