package webpack

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/noxecane/anansi/api"
	"github.com/noxecane/anansi/requests"
	"github.com/noxecane/anansi/responses"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

// WebpackOpts are configuration values for the Webpack middleware
type WebpackOpts struct {
	Environment      string                // Application environment(dev, test e.t.c.)
	Timeout          time.Duration         // Duration before request context times out. Defaults to 1 minute
	CompressionLevel int                   // Level of compression for responses, ranging from 1-9. Defaults to 5
	CORSOrigins      []string              // list of allowed origins
	Registry         prometheus.Registerer // registry for prometheus. This is where we add response time collector
}

// Webpack sets a reasonable set of middleware in the right order taking into consideration
// those that defer computation(especially)
//
// The middleware set up includes:
//
// - Automatic Request IDs
//
// - Response time middleware and header(and metrics if a registry is passed)
//
// - Real IP middleware
//
// - Middleware for hanging slashes
//
// - Compressing response body
//
// - CORS handling for dev and production
//
// - Request Logging
//
// - Panic Recovery(with special support for api.Error)
//
// - Timeouts on request context
func Webpack(router *chi.Mux, log zerolog.Logger, conf WebpackOpts) {
	if len(conf.CORSOrigins) > 0 {
		requests.CORS(conf.Environment, conf.CORSOrigins...)
	}

	if conf.CompressionLevel == 0 {
		conf.CompressionLevel = 5
	}

	if conf.Timeout == 0 {
		conf.Timeout = time.Minute
	}

	router.Use(middleware.Compress(conf.CompressionLevel))

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.RedirectSlashes)

	router.Use(requests.AttachLogger(log))
	router.Use(requests.Log)
	router.Use(requests.Timeout(conf.Timeout))

	router.Use(responses.ResponseTime)
	if conf.Registry != nil {
		router.Use(responses.RequestDuration(conf.Registry))
	}
	router.Use(api.Recoverer(conf.Environment))
}

// HTMLPack sets a reasonable set of middleware in the right order taking into consideration
// those that defer computation(especially)
//
// The middleware set up includes:
//
// - Automatic Request IDs
//
// - Real IP middleware
//
// - Compressing response body
//
// - Request Logging
//
// - Response time header
//
// - Panic Recovery(with special support for api.Error)
//
// - Timeouts on request context
func HTMLpack(router *chi.Mux, log zerolog.Logger, conf WebpackOpts) {
	if conf.CompressionLevel == 0 {
		conf.CompressionLevel = 5
	}

	if conf.Timeout == 0 {
		conf.Timeout = time.Minute
	}

	router.Use(middleware.Compress(conf.CompressionLevel))

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(requests.AttachLogger(log))
	router.Use(requests.Log)
	router.Use(requests.Timeout(conf.Timeout))

	router.Use(responses.ResponseTime)
	router.Use(api.Recoverer(conf.Environment))
}
