package middleware

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/rs/zerolog"
	"github.com/tsaron/anansi"
)

// Recoverer creates a middleware that handles panics from chi controllers. It handles
// printing(optionally stack trace in dev env) and responding to the client for all
// errors except APIErrors. Note that all errors(bar APIError) respond with a 500
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

					if e, ok := rvr.(anansi.APIError); ok {
						anansi.SendError(r, w, e)
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
