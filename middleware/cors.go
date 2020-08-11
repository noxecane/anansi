package middleware

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/rs/cors"
)

// CORS sets CORS for the handler based on the app environment, making the
// rules lax in dev environment.
func CORS(router *chi.Mux, appEnv string, origins ...string) {
	if appEnv == "dev" {
		router.Use(devCORS().Handler)
	} else {
		router.Use(secureCORS(origins).Handler)
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
func secureCORS(origins []string) *cors.Cors {
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
