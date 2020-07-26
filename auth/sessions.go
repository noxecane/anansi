package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/tsaron/anansi"
)

var (
	ErrAuthorisationFormat = errors.New("Your authorization header format is invalid")
	ErrUnsupportedScheme   = errors.New("Your scheme is not supported")
	ErrEmptyToken          = errors.New("There was no token supplied to the authorization header")
	ErrHeaderNotSet        = errors.New("Authorization header is not set")
)

type SessionStore struct {
	store     *TokenStore
	timeout   time.Duration
	cookieKey string
	scheme    string
}

func NewSessionStore(store *TokenStore, sCycle, cookieKey, scheme string) *SessionStore {
	var timeout time.Duration
	var err error

	if timeout, err = time.ParseDuration(sCycle); err != nil {
		panic(err)
	}

	return &SessionStore{store, timeout, cookieKey, scheme}
}

// Load retrieves a user's session object based on the session key from the Authorization
// header or the session cookie and fails with an error if it faces any issue parsing any of them.
func (s *SessionStore) Load(r *http.Request, w http.ResponseWriter, session interface{}) {
	var err error
	var cookie *http.Cookie

	if cookie, err = r.Cookie(s.cookieKey); err != nil {
		authHeader := r.Header.Get("Authorization")

		// if there's no authorisation header, then there's no use going further
		if len(authHeader) == 0 {
			panic(anansi.APIError{
				Code:    http.StatusUnauthorized,
				Message: ErrHeaderNotSet.Error(),
				Err:     ErrHeaderNotSet,
			})
		}

		splitAuth := strings.Split(authHeader, " ")

		// we are expecting "${Scheme} ${Token}"
		if len(splitAuth) != 2 {
			panic(anansi.APIError{
				Code:    http.StatusUnauthorized,
				Message: ErrAuthorisationFormat.Error(),
				Err:     ErrAuthorisationFormat,
			})
		}

		scheme := splitAuth[0]
		if scheme != s.scheme && scheme != "Bearer" {
			panic(anansi.APIError{
				Code:    http.StatusUnauthorized,
				Message: ErrUnsupportedScheme.Error(),
				Err:     ErrUnsupportedScheme,
			})
		}

		token := splitAuth[1]

		if len(token) == 0 {
			panic(anansi.APIError{
				Code:    http.StatusUnauthorized,
				Message: ErrEmptyToken.Error(),
				Err:     ErrEmptyToken,
			})
		}

		if scheme == "Bearer" {
			err = s.store.Refresh(token, s.timeout, session)
		} else {
			err = DecodeJWT(s.store.secret, []byte(token), session)
		}

		if err != nil {
			panic(anansi.APIError{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
				Err:     err,
			})
		}
	} else {
		err = s.store.Refresh(cookie.Value, s.timeout, session)

		if err != nil {
			panic(anansi.APIError{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
				Err:     err,
			})
		}

		// extend the cookie's lifetime
		cookie.Expires = time.Now().Add(s.timeout)
		http.SetCookie(w, cookie)
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
				panic(anansi.APIError{
					Code:    http.StatusUnauthorized,
					Message: ErrHeaderNotSet.Error(),
					Err:     ErrHeaderNotSet,
				})
			}

			splitAuth := strings.Split(authHeader, " ")

			// we are expecting "${Scheme} ${Token}"
			if len(splitAuth) != 2 {
				panic(anansi.APIError{
					Code:    http.StatusUnauthorized,
					Message: ErrAuthorisationFormat.Error(),
					Err:     ErrAuthorisationFormat,
				})
			}

			scheme := splitAuth[0]
			if scheme != s.scheme {
				panic(anansi.APIError{
					Code:    http.StatusUnauthorized,
					Message: ErrUnsupportedScheme.Error(),
					Err:     ErrUnsupportedScheme,
				})
			}

			token := splitAuth[1]

			if len(token) == 0 {
				panic(anansi.APIError{
					Code:    http.StatusUnauthorized,
					Message: ErrEmptyToken.Error(),
					Err:     ErrEmptyToken,
				})
			}

			if err := DecodeJWT(s.store.secret, []byte(token), session); err != nil {
				panic(anansi.APIError{
					Code:    http.StatusUnauthorized,
					Message: err.Error(),
					Err:     err,
				})
			}

			next.ServeHTTP(w, r)
		})
	}
}
