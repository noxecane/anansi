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
func (s *SessionStore) Load(r *http.Request, w http.ResponseWriter, session *interface{}) error {
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

// IsSecure loads a user session into the request context
func (s *SessionStore) IsSecure() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var session interface{}
			var ctx context.Context

			if err := s.Load(r, w, &session); err != nil {
				ctx = context.WithValue(r.Context(), errorKey{}, err)
			} else {
				ctx = context.WithValue(r.Context(), sessionKey{}, session)
			}

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// HeadlessSession creates a session that uses the headless scheme and attaches it to the request
// This would be mainly useful for tests and inter-service requests.
func (s *SessionStore) HeadlessSession(r *http.Request, session interface{}) (string, error) {
	if token, err := EncodeJWT(s.store.secret, s.headlessTimeout, session); err != nil {
		return "", err
	} else {
		r.Header.Set("Authorization", fmt.Sprintf("Headless %s", token))
		return token, nil
	}
}

// BearerSession creates a session that uses the "BearerSession" scheme and attaches it to the request
// This would be mainly useful for tests.
func (s *SessionStore) BearerSession(r *http.Request, session interface{}) (token, key string, err error) {
	if token, key, err = s.store.Commission(s.timeout, session); err != nil {
		return
	} else {
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		return
	}
}

// BearerSessionFromKey is like `Bearer` but with support for using used defined keys to define session keys.
// This would be mainly useful for tests.
func (s *SessionStore) BearerSessionFromKey(r *http.Request, key string, session interface{}) (string, error) {
	if token, err := s.store.CommissionWithKey(s.timeout, key, session); err != nil {
		return "", err
	} else {
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		return token, nil
	}
}

// CookieSession creates a new cookie session and attaches it a response. Use this to create new user sessions.
func (s *SessionStore) CookieSession(w http.ResponseWriter, session interface{}) (token, key string, err error) {
	expires := time.Now().Add(s.timeout)
	if token, key, err = s.store.Commission(s.timeout, session); err != nil {
		return
	} else {
		cookie := http.Cookie{Name: cookieKey, Value: token, Expires: expires}
		http.SetCookie(w, &cookie)
		return
	}
}

// CookieSessionFromKey creates a new cookie session and attaches it a response. Use this to create new
// user sessions where you already have a key for revocation.
func (s *SessionStore) CookieSessionFromKey(w http.ResponseWriter, key string, session interface{}) (string, error) {
	expires := time.Now().Add(s.timeout)
	if token, err := s.store.CommissionWithKey(s.timeout, key, session); err != nil {
		return "", err
	} else {
		cookie := http.Cookie{Name: cookieKey, Value: token, Expires: expires, HttpOnly: true, Secure: !s.devMode}
		http.SetCookie(w, &cookie)
		return token, nil
	}
}

func Revoke() {

}

// GetSession retrieves a user session stored in the context, panics with a 401 error if it's not there.
func GetSession(r *http.Request) interface{} {
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
			Message: "You need to be signed in",
		})
	}

	return session
}
