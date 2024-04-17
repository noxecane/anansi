package html

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Server creates route for routing requests for static files
func Server(router *chi.Mux, folder string) {
	fs := http.FileServer(http.Dir("./" + folder))
	handler := http.StripPrefix("/"+folder+"/", fs)

	router.Get("/"+folder+"/*", handler.ServeHTTP)
}
