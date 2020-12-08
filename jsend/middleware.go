package jsend

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/random-guys/go-siber/sessions"
	"github.com/rs/zerolog"
)

// STACK_SIZE is the number of bytes to print to stderr when recovering from a panic
var STACK_SIZE = 12 * 1024

// Recoverer creates a middleware that handles panics from chi controllers. It handles
// printing(optionally stack trace in dev env) and responding to the client for all
// errors except Err.
//
// Note that all errors(bar Err and request context timeouts) respond with a 500
func Recoverer(env string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {
					if e, ok := rvr.(Err); ok {
						Error(r, w, e)
					} else {
						ctx := r.Context()
						// always log errors regardless of the type
						log := zerolog.Ctx(ctx)
						err := rvr.(error) // it would be serious if this wasn't an error
						log.Err(err).Msg("")

						// give dev a chance to trace unknown errors
						if env == "dev" || env == "test" {
							stack := make([]byte, STACK_SIZE)
							stack = stack[:runtime.Stack(stack, false)]
							fmt.Fprintf(os.Stderr, "recovering from panic:\n%s", stack)
						}

						// make sure timeouts are reported as 504
						if ctx.Err() == context.DeadlineExceeded {
							http.Error(w, http.StatusText(http.StatusGatewayTimeout), http.StatusGatewayTimeout)
						} else {
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

func Headless(store *sessions.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			type void struct{}
			var empty void

			// force a panic if you have to
			LoadHeadless(store, r, empty)

			// nothing to worry about
			next.ServeHTTP(w, r)
		})
	}
}
