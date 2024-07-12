package html

import (
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

// Server creates route for routing requests for static files
func Server(router *chi.Mux, folder string) {
	// create path from home directory to folder
	dir := filepath.Join(".", folder)

	fs := http.FileServer(http.Dir(dir))
	handler := http.StripPrefix("/"+folder+"/", fs)

	router.Get("/"+folder+"/*", handler.ServeHTTP)
}
