package middleware

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/random-guys/siber"
	"github.com/rs/zerolog"
)

type Catch func(w http.ResponseWriter, r *http.Request, v interface{}) bool

// RecovererWithHandler creates a middleware that handles panics from chi controllers. It
// automatically handles APIController errors passing the right error code.
func Recoverer(env string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {

					log := zerolog.Ctx(r.Context())
					if log != nil {
						err := rvr.(error) // kill yourself
						log.Err(err).Msg("")
					} else {
						fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
					}

					if e, ok := rvr.(siber.JSendError); ok {
						siber.SendError(r, w, e)
					} else {
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
