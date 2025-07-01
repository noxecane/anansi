package sessions

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/noxecane/anansi/html"
	"github.com/noxecane/anansi/jwt"
	"github.com/noxecane/anansi/tokens"
)

var ClearCookie struct{}

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
	// How long each headless session should last. Defaults to one hour
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

// NewSession creates a new stateful session with the given sessionID
func (m *Manager) NewSession(ctx context.Context, sessionID string, v any) (string, error) {
	return m.store.Commission(ctx, m.cookieTimeout, sessionID, v)
}

// NewHeadlessSession creates a new stateless session token
func (m *Manager) NewHeadlessSession(v any) (string, error) {
	return jwt.Encode(m.secret, m.headlessTimeout, v)
}

// ToCookie writes a session token to an HTTP cookie
func (m *Manager) ToCookie(w http.ResponseWriter, token string, path string) {
	ck := &http.Cookie{Name: m.cookieKey, Value: token, Path: path}
	http.SetCookie(w, html.SecureCookie(m.isProd, ck))
}

// ToAuth writes a session token to the Authorization header
func (m *Manager) ToAuth(w http.ResponseWriter, token string, isHeadless bool) {
	scheme := "Bearer"
	if isHeadless {
		scheme = m.scheme
	}
	w.Header().Set("Authorization", scheme+" "+token)
}

// FromCookie loads a session from the request's cookie
func (m *Manager) FromCookie(r *http.Request, v any) error {
	ck, _ := r.Cookie(m.cookieKey)
	if ck == nil {
		return ErrEmptyAuthCookie
	}
	return m.store.Extend(r.Context(), ck.Value, m.cookieTimeout, v)
}

// FromAuth loads a session from the Authorization header (supports both bearer and headless)
func (m *Manager) FromAuth(r *http.Request, v any) error {
	scheme, token, err := getAuthorization(r)
	if err != nil {
		return err
	}

	switch scheme {
	case "bearer":
		return m.store.Extend(r.Context(), token, m.bearerTimeout, v)
	case strings.ToLower(m.scheme):
		return jwt.Decode(m.secret, token, v)
	default:
		return ErrUnsupportedScheme
	}
}

// Load attempts to load a session from either Authorization header or cookie
func (m *Manager) Load(r *http.Request, v any) error {
	err := m.FromAuth(r, v)
	switch err {
	case ErrEmptyHeader:
		return m.FromCookie(r, v)
	case nil:
		return nil
	default:
		// Try cookie fallback for other auth errors
		if cookieErr := m.FromCookie(r, v); cookieErr == nil {
			return nil
		}
		// Return original auth error if cookie also fails
		return err
	}
}

// LogoutCookie clears the authentication cookie
func (m *Manager) LogoutCookie(r *http.Request, w http.ResponseWriter) error {
	ck, _ := r.Cookie(m.cookieKey)
	if ck == nil {
		return ErrEmptyAuthCookie
	}

	err := m.store.Decommission(r.Context(), ck.Value, &ClearCookie)
	if err != nil {
		return err
	}

	ck = &http.Cookie{
		Name:    m.cookieKey,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
	}

	http.SetCookie(w, html.SecureCookie(m.isProd, ck))

	return nil
}

// LogoutAuth revokes a session token from Authorization header
func (m *Manager) LogoutAuth(r *http.Request) error {
	scheme, token, err := getAuthorization(r)
	if err != nil {
		return err
	}
	// Only revoke stateful tokens (bearer), headless tokens are stateless
	if scheme == "bearer" {
		return m.store.Decommission(r.Context(), token, &ClearCookie)
	}
	return nil // Headless tokens don't need revocation
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
