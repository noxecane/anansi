package middleware

import (
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

type MiddlwareConfig struct {
	Environment      string        // Application environment(dev, test e.t.c.)
	Timeout          time.Duration // Duration of request before it returns a 504. Defaults to 1 minute
	CompressionLevel int           // Level of compression for responses. Defaults to 5
	CORSOrigins      []string      // list of allowed origins
}

// Sets a reasonable set of middleware in the right order taking into consideration
// those that defer computation(especially)
//
// Middleware set up include:
// - Automatic request IDs
// - Response time middleware
// - Real IP middleware
// - Middleware for hanging slashes
// - Compressing response body
// - CORS handling for dev and production
// - Request Logging
// - Response time header
// - Panic Recovery(with special support for APIError)
// - Timeouts on request conctext
func DefaultMiddleware(router *chi.Mux, log zerolog.Logger, conf MiddlwareConfig) {
	if conf.CORSOrigins != nil && len(conf.CORSOrigins) > 0 {
		if conf.Environment == "dev" {
			router.Use(devCORS().Handler)
		} else {
			router.Use(secureCORS(conf.CORSOrigins).Handler)
		}
	}

	if conf.CompressionLevel == 0 {
		conf.CompressionLevel = 5
	}

	if conf.Timeout == 0 {
		conf.Timeout = time.Minute
	}

	router.Use(middleware.Compress(conf.CompressionLevel))
	router.Use(ResponseTime)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.RedirectSlashes)
	router.Use(AttachLogger(log))
	router.Use(TrackRequest())
	router.Use(Timeout(conf.Timeout))
	router.Use(Recoverer(conf.Environment))
}
