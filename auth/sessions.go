package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/tsaron/anansi"
)

const cookieKey = "web_session"

var (
	ErrAuthorisationFormat = errors.New("Your authorization header format is invalid")
	ErrUnsupportedScheme   = errors.New("Your scheme is not supported")
	ErrEmptyToken          = errors.New("There was no token supplied to the authorization header")
)

type SessionStore struct {
	store           *TokenStore
	timeout         time.Duration
	headlessTimeout time.Duration
	devMode         bool
}

type sessionKey struct{}
type errorKey struct{}

func NewSessionStore(store *TokenStore, sCycle, hCycle string, devMode bool) *SessionStore {
	var timeout time.Duration
	var headlessTimeout time.Duration
	var err error

	if timeout, err = time.ParseDuration(sCycle); err != nil {
		panic(err)
	}

	if headlessTimeout, err = time.ParseDuration(hCycle); err != nil {
		panic(err)
	}

	return &SessionStore{store, timeout, headlessTimeout, devMode}
}

// Load retrieves a user's session object based on the session key from the Authorization
// header or the session cookie and fails with an error if it faces any issue parsing any of them.
func (s *SessionStore) load(r *http.Request, w http.ResponseWriter, session *interface{}) error {
	var token string
	var err error
	var cookie *http.Cookie

	if cookie, err = r.Cookie(cookieKey); err != nil {
		authHeader := r.Header.Get("Authorization")
		// if there's no authorisation header, then there's no use going further
		if len(authHeader) == 0 {
			return nil
		}

		splitAuth := strings.Split(authHeader, " ")

		// we are expecting "${Scheme} ${Token}"
		if len(splitAuth) != 2 {
			return ErrAuthorisationFormat
		}

		scheme := strings.ToLower(splitAuth[0])
		if scheme != "bearer" && scheme != "headless" {
			return ErrUnsupportedScheme
		}

		token = splitAuth[1]

		if len(token) == 0 {
			return ErrEmptyToken
		}

		if scheme == "headless" {
			err = DecodeJWT(s.store.secret, []byte(token), session)
			return err
		} else {
			err = s.store.Refresh(token, s.timeout, session)
			return err
		}
	} else {
		token = cookie.Value
		if err = s.store.Refresh(token, s.timeout, session); err != nil {
			return err
		}

		// extend the cookie's lifetime
		cookie.Expires = time.Now().Add(s.timeout)
		http.SetCookie(w, cookie)
	}

	if err = s.store.Refresh(token, s.timeout, session); err != nil {
		return err
	}

	return nil
}

// Secure loads a user session into the request context
func (s *SessionStore) Secure() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var session interface{}
			var ctx context.Context

			if err := s.load(r, w, &session); err != nil {
				ctx = context.WithValue(r.Context(), errorKey{}, err)
			} else {
				ctx = context.WithValue(r.Context(), sessionKey{}, session)
			}

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// Get retrieves a user session stored in the context, panics with an appropriate error if the
// error value has been set on the request context(for failed session loads). Make sure to use this
// for handlers protected by the Secure method.
func Get(r *http.Request) interface{} {
	ctx := r.Context()
	err := ctx.Value(errorKey{}).(error)
	session := ctx.Value(sessionKey{})

	if err != nil {
		switch err {
		case ErrAuthorisationFormat, ErrEmptyToken, ErrUnsupportedScheme, ErrInvalidToken, ErrJWTExpired, ErrTokenNotFound:
			panic(anansi.APIError{
				Code:    http.StatusUnauthorized,
				Message: err.Error(),
			})
		default:
			panic(anansi.APIError{
				Code:    http.StatusUnauthorized,
				Message: "We could not load your session. Please reach out to support to check the problem",
				Err:     err,
			})
		}
	}

	if session == nil {
		panic(anansi.APIError{
			Code:    http.StatusUnauthorized,
			Message: "We could not find a session for your request",
		})
	}

	return session
}
