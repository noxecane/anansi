package sessions

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/noxecane/anansi/html"
	"github.com/noxecane/anansi/jwt"
	"github.com/noxecane/anansi/tokens"
)

const (
	DefaultSessionKey      = "anansi_session"
	DefaultHeadlessScheme  = "API"
	DefaultSessionDuration = time.Hour
)

var (
	// ErrHeaderFormat is returned when the authorization header doesn't
	// contain 2 fields separated by whitespace.
	ErrHeaderFormat = errors.New("incorrect authorization header format")
	// ErrEmptyHeader is returned when the authorization header is not set
	ErrEmptyHeader = errors.New("empty authorization header")
	// ErrUnsupportedScheme is returned when the scheme is either not bearer nor the set custom scheme
	ErrUnsupportedScheme = errors.New("unsupported authorization scheme")
	// ErrEmptyAuthCookie is returned when the authentication cookie is not set
	ErrEmptyAuthCookie = errors.New("no cookie set for session")
)

type Manager struct {
	store           tokens.Store
	secret          []byte
	isProd          bool
	scheme          string
	cookieKey       string
	cookieTimeout   time.Duration
	headlessTimeout time.Duration
	bearerTimeout   time.Duration
}

// Config defines particular controls for session management
type Config struct {
	// Signals whether the app is running in a production environment
	IsProduction bool
	// The scheme to recognise for headless requests. defaults to
	// DefaultHeadlessScheme
	HeadlessScheme string
	// How long each headless session should last Config.BearerDuration
	HeadlessDuration time.Duration
	// How long bearer sessions should last. Defaults to DefaultSessionDuration
	BearerDuration time.Duration
	// The name of the server cookie the session token is written to
	CookieKey string
	// How long cookie sessions should last. Defaults to Config.BearerDuration
	CookieDuration time.Duration
}

func NewManager(store tokens.Store, secret []byte, config Config) *Manager {
	if config.HeadlessScheme == "" {
		config.HeadlessScheme = DefaultHeadlessScheme
	}

	if config.BearerDuration == 0 {
		config.BearerDuration = DefaultSessionDuration
	}

	if config.HeadlessDuration == 0 {
		config.HeadlessDuration = config.BearerDuration
	}

	if config.CookieDuration == 0 {
		config.CookieDuration = config.BearerDuration
	}

	if config.CookieKey == "" {
		config.CookieKey = DefaultSessionKey
	}

	return &Manager{
		store:           store,
		secret:          secret,
		isProd:          config.IsProduction,
		scheme:          config.HeadlessScheme,
		cookieKey:       config.CookieKey,
		cookieTimeout:   config.CookieDuration,
		bearerTimeout:   config.BearerDuration,
		headlessTimeout: config.HeadlessDuration,
	}
}

// NewHeadlessToken creates a new token for headless session access
func (m *Manager) NewHeadlessToken(r *http.Request, k string, v interface{}) (string, error) {
	return m.store.Commission(r.Context(), m.headlessTimeout, k, v)
}

// NewBearerToken creates a new token for bearer session access
func (m *Manager) NewBearerToken(r *http.Request, k string, v interface{}) (string, error) {
	return m.store.Commission(r.Context(), m.bearerTimeout, k, v)
}

// NewCookieSession creates and writes a cookie that gives access to the session
func (m *Manager) NewCookieSession(r *http.Request, w http.ResponseWriter, v interface{}) error {
	token, err := m.store.Commission(r.Context(), m.cookieTimeout, m.cookieKey, v)
	if err != nil {
		return err
	}
	ck := &http.Cookie{Name: m.cookieKey, Value: token}
	http.SetCookie(w, html.SecureCookie(m.isProd, ck))

	return nil
}

// LoadCookie loads a stateful session from the request's cookie.
func (m *Manager) LoadCookie(r *http.Request, v interface{}) error {
	ck, _ := r.Cookie(m.cookieKey)
	if ck == nil {
		return ErrEmptyAuthCookie
	}

	return m.store.Extend(r.Context(), ck.Value, m.headlessTimeout, v)
}

// LoadBearer loads a stateful session using the session from key the authorization header.
func (m *Manager) LoadBearer(r *http.Request, v interface{}) error {
	scheme, token, err := getAuthorization(r)
	if err != nil {
		return err
	}

	if scheme != "bearer" {
		return ErrUnsupportedScheme
	}

	return m.store.Extend(r.Context(), token, m.headlessTimeout, v)
}

// LoadHeadless loads a stateless session from the encoded token in the authorization header.
func (m *Manager) LoadHeadless(r *http.Request, v interface{}) error {
	scheme, token, err := getAuthorization(r)
	if err != nil {
		return err
	}

	if scheme != m.scheme {
		return ErrUnsupportedScheme
	}

	return jwt.Decode(m.secret, token, v)
}

// Load to load either bearer, cookie or headless session
func (m *Manager) Load(r *http.Request, v interface{}) error {
	err := m.LoadBearer(r, v)
	if err == ErrEmptyHeader {
		return m.LoadCookie(r, v)
	} else if err == ErrUnsupportedScheme {
		return m.LoadHeadless(r, v)
	}

	return err
}

func getAuthorization(r *http.Request) (string, string, error) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		return "", "", ErrEmptyHeader
	}

	splitAuth := strings.Fields(authHeader)

	if len(splitAuth) != 2 {
		return "", "", ErrHeaderFormat
	}

	return strings.ToLower(splitAuth[0]), splitAuth[1], nil
}
