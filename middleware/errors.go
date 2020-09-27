package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/random-guys/go-siber"
	"github.com/rs/zerolog"
)

type Catch func(w http.ResponseWriter, r *http.Request, v interface{}) bool

// Recoverer creates a middleware that handles panics from chi controllers. It handles
// printing(optionally stack trace in dev env) and responding to the client for all
// errors except JSendErrors. Note that all errors(bar JSendError) respond with a 500
func Recoverer(env string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {
					if e, ok := rvr.(siber.JSendError); ok {
						siber.SendError(r, w, e)
					} else {
						// log errors before printing stack trace
						log := zerolog.Ctx(r.Context())
						err := rvr.(error) // kill yourself
						log.Err(err).Msg("")

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
