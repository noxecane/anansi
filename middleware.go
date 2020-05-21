package anansi

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
)

// DevCORS creates a very permissive CORS instance
func DevCORS() *cors.Cors {
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

// SecureCORS is a lot like DevCORS except with a limited set of origins
func SecureCORS(origins ...string) *cors.Cors {
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

func Recoverer(next http.Handler) func(http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {

				logEntry := middleware.GetLogEntry(r)
				if logEntry != nil {
					logEntry.Panic(rvr, debug.Stack())
				} else {
					fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
				}
				// debug.PrintStack()

				var err error
				if e, ok := rvr.(error); ok {
					err = e
				} else {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

				if e, ok := err.(APIError); ok {
					SendError(r, w, e)
				} else if e := interpreter(err); e.Code != 0 {
					SendError(r, w, e)
				} else {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
