package anansi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/tsaron/anansi/jwt"
	"github.com/tsaron/anansi/tokens"
)

var (
	ErrAuthorisationFormat = errors.New("Your authorization header format is invalid")
	ErrUnsupportedScheme   = errors.New("Your scheme is not supported")
	ErrEmptyToken          = errors.New("There was no token supplied to the authorization header")
	ErrHeaderNotSet        = errors.New("Authorization header is not set")
)

type SessionStore struct {
	Store   *tokens.Store
	Timeout time.Duration

	Secret []byte

	Scheme    string
	ClaimsKey string
}

// Load retrieves a user's session object based on the session key from the Authorization
// header or the session cookie and fails with an error if it faces any issue parsing any of them.
func (s *SessionStore) Load(r *http.Request, session interface{}) {
	var err error

	authHeader := r.Header.Get("Authorization")

	// if there's no authorisation header, then there's no use going further
	if len(authHeader) == 0 {
		panic(APIError{
			Code:    http.StatusUnauthorized,
			Message: ErrHeaderNotSet.Error(),
			Err:     ErrHeaderNotSet,
		})
	}

	splitAuth := strings.Split(authHeader, " ")

	// we are expecting "${Scheme} ${Token}"
	if len(splitAuth) != 2 {
		panic(APIError{
			Code:    http.StatusUnauthorized,
			Message: ErrAuthorisationFormat.Error(),
			Err:     ErrAuthorisationFormat,
		})
	}

	scheme := splitAuth[0]
	if scheme != s.Scheme && scheme != "Bearer" {
		panic(APIError{
			Code:    http.StatusUnauthorized,
			Message: ErrUnsupportedScheme.Error(),
			Err:     ErrUnsupportedScheme,
		})
	}

	token := splitAuth[1]

	if len(token) == 0 {
		panic(APIError{
			Code:    http.StatusUnauthorized,
			Message: ErrEmptyToken.Error(),
			Err:     ErrEmptyToken,
		})
	}

	if scheme == "Bearer" {
		err = s.Store.Peek(token, session)
		if err == nil {
			err = s.Store.Extend(token, s.Timeout)
		}
	} else {
		err = jwt.Decode(s.ClaimsKey, s.Secret, []byte(token), session)
	}

	if err != nil {
		panic(APIError{
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
			var session interface{}

			authHeader := r.Header.Get("Authorization")
			// if there's no authorisation header, then there's no use going further
			if len(authHeader) == 0 {
				panic(APIError{
					Code:    http.StatusUnauthorized,
					Message: ErrHeaderNotSet.Error(),
					Err:     ErrHeaderNotSet,
				})
			}

			splitAuth := strings.Split(authHeader, " ")

			// we are expecting "${Scheme} ${Token}"
			if len(splitAuth) != 2 {
				panic(APIError{
					Code:    http.StatusUnauthorized,
					Message: ErrAuthorisationFormat.Error(),
					Err:     ErrAuthorisationFormat,
				})
			}

			scheme := splitAuth[0]
			if scheme != s.Scheme {
				panic(APIError{
					Code:    http.StatusUnauthorized,
					Message: ErrUnsupportedScheme.Error(),
					Err:     ErrUnsupportedScheme,
				})
			}

			token := splitAuth[1]

			if len(token) == 0 {
				panic(APIError{
					Code:    http.StatusUnauthorized,
					Message: ErrEmptyToken.Error(),
					Err:     ErrEmptyToken,
				})
			}

			if err := jwt.Decode(s.ClaimsKey, s.Secret, []byte(token), &session); err != nil {
				panic(APIError{
					Code:    http.StatusUnauthorized,
					Message: err.Error(),
					Err:     err,
				})
			}

			next.ServeHTTP(w, r)
		})
	}
}
