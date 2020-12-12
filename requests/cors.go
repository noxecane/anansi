package requests

import (
	"net/http"

	"github.com/rs/cors"
)

// CORS sets CORS for the handler based on the app environment, making the
// rules lax in dev environment.
func CORS(appEnv string, origins ...string) func(http.Handler) http.Handler {
	if appEnv == "dev" {
		return devCORS().Handler
	} else {
		return secureCORS(origins).Handler
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
