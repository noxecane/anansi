package middleware

import (
	"context"
	"net/http"
	"runtime/debug"

	"github.com/random-guys/go-siber"
	"github.com/rs/zerolog"
)

// Recoverer creates a middleware that handles panics from chi controllers. It handles
// printing(optionally stack trace in dev env) and responding to the client for all
// errors except JSendErrors. Note that all errors(bar JSendError) respond with a 500
func Recoverer(env string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				// abort early if request timed out
				if r.Context().Err() == context.DeadlineExceeded {
					http.Error(w, http.StatusText(http.StatusGatewayTimeout), http.StatusGatewayTimeout)
				}

				if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {
					if e, ok := rvr.(siber.JSendError); ok {
						siber.SendError(r, w, e)
					} else {
						ctx := r.Context()

						// always log errors regardless of the type
						log := zerolog.Ctx(ctx)
						err := rvr.(error) // it would be serious if this wasn't an error
						log.Err(err).Msg("")

						// give dev a chance to trace unknown errors
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
