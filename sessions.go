package siber

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/random-guys/go-siber/jsend"
	"github.com/random-guys/go-siber/jwt"
	"github.com/random-guys/go-siber/tokens"
)

var (
	// ErrHeaderFormat is returned when the authorization header doesn't
	// contain 2 fields separated by whitespace.
	ErrHeaderFormat = errors.New("incorrect authorization header format")
	// ErrEmptyHeader is returned when the authorization header is not set
	ErrEmptyHeader = errors.New("empty authorization header")
	// ErrUnsupportedScheme is returned when the scheme is either not bearer nor the set custom scheme
	ErrUnsupportedScheme = errors.New("unsupported authorization scheme")
)

type SessionStore struct {
	store   *tokens.Store
	timeout time.Duration
	secret  []byte
	scheme  string
}

func NewSessionStore(secret []byte, scheme string, timeout time.Duration, store *tokens.Store) *SessionStore {
	return &SessionStore{store, timeout, secret, strings.ToLower(scheme)}
}

// LoadBearer loads a stateful session using the session from key the authorization header.
func (s *SessionStore) LoadBearer(r *http.Request, v interface{}) error {
	scheme, token, err := getAuthorization(r)
	if err != nil {
		return err
	}

	if scheme != "bearer" {
		return ErrUnsupportedScheme
	}

	return s.store.Extend(r.Context(), token, s.timeout, v)
}

// LoadHeadless loads a stateless session from the encoded token in the authorization header.
func (s *SessionStore) LoadHeadless(r *http.Request, v interface{}) error {
	scheme, token, err := getAuthorization(r)
	if err != nil {
		return err
	}

	if scheme != s.scheme {
		return ErrUnsupportedScheme
	}

	return jwt.DecodeEmbedded(s.secret, []byte(token), v)
}

// Load trys both LoadBearer and LoadHeadless.
func (s *SessionStore) Load(r *http.Request, v interface{}) error {
	err := s.LoadBearer(r, v)
	if err == ErrUnsupportedScheme {
		return s.LoadHeadless(r, v)
	}

	return err
}

// Secure loads a user session into the request context
func (s *SessionStore) Headless() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scheme, token, _ := getAuthorization(r)

			if scheme != s.scheme {
				panic(jsend.Err{
					Code:    http.StatusUnauthorized,
					Message: ErrUnsupportedScheme.Error(),
				})
			}

			if token == "" {
				panic(jsend.Err{
					Code:    http.StatusUnauthorized,
					Message: ErrEmptyToken.Error(),
				})
			}

			// read and discard session data
			if err := jwt.DecodeEmbedded(s.secret, []byte(token), &struct{}{}); err != nil {
				panic(jsend.Err{
					Code:    http.StatusUnauthorized,
					Message: "Your token is invalid",
					Err:     err,
				})
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getAuthorization(r *http.Request) (string, string, error) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		panic(jsend.Err{
			Code:    http.StatusUnauthorized,
			Message: ErrEmptyHeader.Error(),
		})
	}

	splitAuth := strings.Fields(authHeader)

	if len(splitAuth) != 2 {
		return "", "", ErrHeaderFormat
	}

	return strings.ToLower(splitAuth[0]), splitAuth[1], nil
}
