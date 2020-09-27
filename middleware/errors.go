package middleware

import (
	"context"
	"net/http"
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
					if e, ok := rvr.(anansi.APIError); ok {
						anansi.SendError(r, w, e)
					} else {
						ctx := r.Context()

						// always log errors regardless of the type
						log := zerolog.Ctx(ctx)
						err := rvr.(error) // it would be serious if this wasn't an error
						log.Err(err).Msg("")

						// make sure request hasn't timed out
						if ctx.Err() == context.DeadlineExceeded {
							http.Error(w, http.StatusText(http.StatusGatewayTimeout), http.StatusGatewayTimeout)
						} else {
							// give dev a chance to trace unknown errors
							if env == "dev" || env == "test" {
								debug.PrintStack()
							}
							http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
						}
					}
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
