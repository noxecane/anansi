package anansi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
)

// DefaultMiddleware sets up some middlware for a router as well as
// liveliness(/) and not found handlers:
// - RequestID
// - RealIP
// - Logger(using zerolog)
// - Recoverer
// - RedirectSlashes
// - Compress(with compression level of 5)
// - Timeout(with 1 minute)
func DefaultMiddleware(router *chi.Mux, appEnv string, log zerolog.Logger) {
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(ZeroMiddleware(log, appEnv))
	router.Use(Recoverer(appEnv))
	router.Use(middleware.RedirectSlashes)
	router.Use(middleware.Compress(5))
	router.Use(middleware.Timeout(time.Minute))
}

// DefaultRoutes adds liveness(/) and Not found handlers for the passed router.
func DefaultRoutes(router *chi.Mux) {
	router.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Up and Running!"))
		if err != nil {
			panic(err)
		}
	})

	router.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Whoops!! This route doesn't exist", 404)
	})
}

// CORS sets CORS for the handler. It enables localhost by default when appEnv
// is not "dev".
func CORS(router *chi.Mux, appEnv string, origins ...string) {
	if appEnv == "dev" {
		router.Use(devCORS().Handler)
	} else {
		origins = append(origins, "http://localhost")
		router.Use(secureCORS(origins...).Handler)
	}
}

// devCORS creates a very permissive CORS instance
func devCORS() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
}

// secureCORS is a lot like DevCORS except with a limited set of origins
func secureCORS(origins ...string) *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
}
