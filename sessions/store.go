package sessions

import (
	"errors"
	"net/http"
	"strings"
	"time"

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

type Store struct {
	store   tokens.Store
	timeout time.Duration
	secret  []byte
	scheme  string
}

func NewStore(secret []byte, scheme string, timeout time.Duration, store tokens.Store) *Store {
	return &Store{store, timeout, secret, strings.ToLower(scheme)}
}

// Save stores the session using the given key and creates a token for accesing it.
func (s *Store) Save(r *http.Request, k string, v interface{}) (string, error) {
	return s.store.Commission(r.Context(), s.timeout, k, v)
}

// LoadBearer loads a stateful session using the session from key the authorization header.
func (s *Store) LoadBearer(r *http.Request, v interface{}) error {
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
func (s *Store) LoadHeadless(r *http.Request, v interface{}) error {
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
func (s *Store) Load(r *http.Request, v interface{}) error {
	err := s.LoadBearer(r, v)
	if err == ErrUnsupportedScheme {
		return s.LoadHeadless(r, v)
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
