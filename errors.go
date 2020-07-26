package anansi

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/go-chi/chi/middleware"
)

// APIError is a struct describing an error
type APIError struct {
	Code    int         `json:"-"`
	Message string      `json:"message"`
	Meta    interface{} `json:"meta"`
	Err     error       `json:"-"`
}

// implements the error interface
func (e APIError) Error() string { return e.Message }

func (e APIError) Unwrap() error { return e.Err }

// Recoverer creates a middleware that handles panics from chi controllers. It uses
// the passed interpreters to try to convert errors to APIErrors where possible
// otherwise it returns a 500 error. When the panic is an APIError or is interpreted
// as one, it sends a response with the right error code.
// TODO: add support for wrapped errors in APIError.
func Recoverer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {

				logEntry := middleware.GetLogEntry(r)
				if logEntry != nil {
					logEntry.Panic(rvr, debug.Stack())
				} else {
					fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
				}

				if e, ok := rvr.(APIError); ok {
					SendError(r, w, e)
				} else {
					debug.PrintStack()
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
