package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tsaron/anansi"
)

type SessionStore struct {
	store   *TokenStore
	timeout time.Duration
}

type sessionKey struct{}

func NewSessionStore(store *TokenStore, strTime string) *SessionStore {
	timeout, err := time.ParseDuration(strTime)
	if err != nil {
		panic(err)
	}

	return &SessionStore{store, timeout}
}

// Load retrieves a user's session object based on the session
// key from the Authorization header and panics with the right error if
// there are any issues with it.
func (s *SessionStore) Load(r *http.Request, session interface{}) {
	authHeader := r.Header.Get("Authorization")
	splitAuth := strings.Split(authHeader, " ")

	// we are expecting "Bearer ${Token}"
	if len(splitAuth) != 2 {
		panic(anansi.APIError{
			Code:    http.StatusUnauthorized,
			Message: "You are not logged in.",
			Err:     errors.New("Wrong authorization header format"),
		})
	}

	if strings.ToLower(splitAuth[0]) != "bearer" {
		panic(anansi.APIError{
			Code:    http.StatusUnauthorized,
			Message: "Cannot support any other scheme other than 'Bearer'",
		})
	}

	if len(splitAuth[1]) == 0 {
		panic(anansi.APIError{
			Code:    http.StatusUnauthorized,
			Message: "You are not logged in.",
			Err:     errors.New("There was no token supplied to the authorization header"),
		})
	}

	var err error

	if err = s.store.Refresh(splitAuth[1], s.timeout, session); err != nil {
		panic(anansi.APIError{
			Code:    http.StatusUnauthorized,
			Message: "You are not logged in.",
			Err:     err,
		})
	}
}

// IsSecure loads a user session into the request context
func (s *SessionStore) IsSecure() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var session interface{}
			s.Load(r, &session)

			ctx := context.WithValue(r.Context(), sessionKey{}, session)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// Attach creates a new session and attaches it a request. This would be mainly
// useful for tests.
func (s *SessionStore) Attach(r *http.Request, session interface{}) error {
	var token string
	var err error
	if token, err = s.store.New(s.timeout, session); err != nil {
		return err
	}

	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	return nil
}

// GetSession retrieves a user session stored in the context, panics
// with a 401 error if it's not there.
func GetSession(r *http.Request) interface{} {
	ctx := r.Context()
	session := ctx.Value(sessionKey{})
	if session == nil {
		panic(anansi.APIError{
			Code:    http.StatusUnauthorized,
			Message: "You are not logged in.",
		})
	}

	return session
}
