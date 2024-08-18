package api

import (
	"bytes"
	jsonslow "encoding/json"
	"net/http"

	"github.com/noxecane/anansi"
	"github.com/noxecane/anansi/json"
	"github.com/noxecane/anansi/responses"
	"github.com/rs/zerolog"
)

// Success sends a JSend success message with status code 200. It logs the response
// if a zerolog.Logger is attached to the request.
func Success(r *http.Request, w http.ResponseWriter, v interface{}) {
	log := zerolog.Ctx(r.Context())
	raw := getJSON(log, v)

	responses.Send(w, http.StatusOK, raw)

	log.Info().
		Int("status", http.StatusOK).
		Int("length", len(raw)).
		Interface("response_headers", anansi.SimpleMap(w.Header())).
		Msg("")
}

// Error sends a JSend error message. It logs the response if a zerolog.Logger is attached to the request.
func Error(r *http.Request, w http.ResponseWriter, err Err) {
	log := zerolog.Ctx(r.Context())
	raw := getJSON(log, err)

	responses.Send(w, err.Code, raw)

	log.Err(err).
		Int("status", err.Code).
		Int("length", len(raw)).
		Interface("response_headers", anansi.SimpleMap(w.Header())).
		Msg("")
}

func getJSON(log *zerolog.Logger, v interface{}) []byte {
	raw, _ := json.Marshal(v)

	// log API responses
	if v != nil {
		log.UpdateContext(func(ctx zerolog.Context) zerolog.Context {
			buffer := new(bytes.Buffer)

			if err := jsonslow.Compact(buffer, raw); err != nil {
				panic(err)
			}

			return ctx.RawJSON("response", buffer.Bytes())
		})
	}

	return raw
}
