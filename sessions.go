package siber

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/random-guys/go-siber/jwt"
	"github.com/random-guys/go-siber/tokens"
)

var (
	ErrAuthorisationFormat = errors.New("Your authorization header format is invalid")
	ErrUnsupportedScheme   = errors.New("Your scheme is not supported")
	ErrEmptyToken          = errors.New("There was no token supplied to the authorization header")
	ErrHeaderNotSet        = errors.New("Authorization header is not set")
)

type SessionStore struct {
	store   *tokens.Store
	timeout time.Duration
	secret  []byte
	scheme  string
}

func NewSessionStore(secret []byte, scheme string, timeout time.Duration, store *tokens.Store) *SessionStore {
	return &SessionStore{store, timeout, secret, scheme}
}

// Load retrieves a user's session object based on the session key from the Authorization
// header or the session cookie and fails with an error if it faces any issue parsing any of them.
func (s *SessionStore) Load(r *http.Request, session interface{}) {
	var err error

	scheme, token := getAuthorization(r)

	if scheme != s.scheme && scheme != "bearer" {
		panic(JSendError{
			Code:    http.StatusUnauthorized,
			Message: ErrUnsupportedScheme.Error(),
			Err:     ErrUnsupportedScheme,
		})
	}

	if token == "" {
		panic(JSendError{
			Code:    http.StatusUnauthorized,
			Message: ErrEmptyToken.Error(),
			Err:     ErrEmptyToken,
		})
	}

	if scheme == "bearer" {
		err = s.store.Extend(token, s.timeout, session)
	} else {
		err = jwt.DecodeEmbedded(s.secret, []byte(token), session)
	}

	if err != nil {
		panic(JSendError{
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
			Err:     err,
		})
	}
}

// Secure loads a user session into the request context
func (s *SessionStore) Headless() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scheme, token := getAuthorization(r)
			if scheme != s.scheme {
				panic(JSendError{
					Code:    http.StatusUnauthorized,
					Message: ErrUnsupportedScheme.Error(),
					Err:     ErrUnsupportedScheme,
				})
			}

			if token == "" {
				panic(JSendError{
					Code:    http.StatusUnauthorized,
					Message: ErrEmptyToken.Error(),
					Err:     ErrEmptyToken,
				})
			}

			// read and discard session data
			if err := jwt.DecodeEmbedded(s.secret, []byte(token), &struct{}{}); err != nil {
				panic(JSendError{
					Code:    http.StatusUnauthorized,
					Message: err.Error(),
					Err:     err,
				})
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getAuthorization(r *http.Request) (scheme, token string) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		panic(JSendError{
			Code:    http.StatusUnauthorized,
			Message: ErrHeaderNotSet.Error(),
			Err:     ErrHeaderNotSet,
		})
	}

	splitAuth := strings.Split(strings.TrimSpace(authHeader), " ")

	if len(splitAuth) != 2 {
		panic(JSendError{
			Code:    http.StatusUnauthorized,
			Message: ErrAuthorisationFormat.Error(),
			Err:     ErrAuthorisationFormat,
		})
	}

	return strings.ToLower(splitAuth[0]), splitAuth[1]
}
