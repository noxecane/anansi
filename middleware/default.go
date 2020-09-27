package middleware

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// Sets up basic middleware that come with go-chi
// - RequestID
// - RealIP
// - RedirectSlashes
// - Compress(with compression level of 5)
func DefaultMiddleware(router *chi.Mux) {
	router.Use(middleware.Compress(5))
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.RedirectSlashes)
}
