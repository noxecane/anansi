package middleware

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/rs/zerolog"
	"github.com/tsaron/anansi"
)

type Catch func(w http.ResponseWriter, r *http.Request, v interface{}) bool

// Recoverer creates a middleware that can detect APIError from panic. Internally uses
// RecoverWithHandler.
func Recoverer(env string) func(http.Handler) http.Handler {
	return RecovererWithHandler(env, func(w http.ResponseWriter, r *http.Request, v interface{}) bool {
		if e, ok := v.(anansi.APIError); ok {
			anansi.SendError(r, w, e)
			return true
		}
		return false
	})
}

// RecovererWithHandler creates a middleware that handles panics from chi controllers. It uses
// the passed catch to interprete and handle the error(like send it as JSON) and returns 500
// if it can't be interpreted.
// TODO: add support for wrapped errors in APIError.
func RecovererWithHandler(env string, catch Catch) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {
					// only do this if catch could not handle it.
					if !catch(w, r, rvr) {
						log := zerolog.Ctx(r.Context())
						if log == nil {
							err := rvr.(error) // kill yourself
							log.Err(err).Msg("")
						} else {
							fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
						}

						if env == "dev" || env == "test" {
							debug.PrintStack()
						}
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					}
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
