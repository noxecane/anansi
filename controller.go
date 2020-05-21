package anansi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
)

// DefaultMiddleware sets up some middlware for a router as well as
// liveliness(/) and not found handlers:
// - RequestID
// - RealIP
// - Logger(using zerolog)
// - Recoverer
// - CORS
// - RedirectSlashes
// - Compress(with compression level of 5)
// - Timeout(with 1 minute)
func DefaultMiddleware(env BasicEnv, log zerolog.Logger, router *chi.Mux) {
	if env.AppEnv == "dev" {
		router.Use(DevCORS().Handler)
	} else {
		router.Use(
			SecureCORS(
				"https://*godview.netlify.com",
				"https://*tsaron.com",
				"http://localhost").
				Handler)
	}

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(ZeroMiddleware(log, env.AppEnv))
	router.Use(Recoverer)
	router.Use(middleware.RedirectSlashes)
	router.Use(middleware.Compress(5))
	router.Use(middleware.Timeout(time.Minute))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Up and Running!"))
		if err != nil {
			panic(err)
		}
	})

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Whoops!! This route doesn't exist", 404)
	})
}
