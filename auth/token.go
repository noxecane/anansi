package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
)

const tokenMapKey = "general_token_map"

type TokenStore struct {
	redis *redis.Client
}

func NewTokenStore(r *redis.Client) *TokenStore {
	return &TokenStore{redis: r}
}

// New creates a 32 character single use token that expires after the given timeout.
func (ts *TokenStore) New(timeout time.Duration, data interface{}) (string, error) {
	var token string
	var encoded []byte
	var err error

	if token, err = RandomString(32); err != nil {
		return "", err
	}

	// TODO: replace this something lighter and faster
	if encoded, err = json.Marshal(data); err != nil {
		return "", err
	}

	// One would naturally prefer hash maps but they don't support individual subkey expiry.
	tokenKey := fmt.Sprintf("%s::%s", tokenMapKey, token)

	if _, err = ts.redis.Set(tokenKey, encoded, timeout).Result(); err != nil {
		return "", err
	}

	return token, nil
}

// Peek gets the data the token references without changing its lifetime.
func (ts *TokenStore) Peek(token string, data interface{}) error {
	tokenKey := fmt.Sprintf("%s::%s", tokenMapKey, token)
	return ts.peekToken(tokenKey, data)
}

// Refresh loads the data the token references and refreshes it's lifetime to timeout.
func (ts *TokenStore) Refresh(token string, timeout time.Duration, data interface{}) error {
	var err error

	tokenKey := fmt.Sprintf("%s::%s", tokenMapKey, token)
	if err = ts.peekToken(tokenKey, data); err != nil {
		return err
	}

	if _, err = ts.redis.Expire(tokenKey, timeout).Result(); err != nil {
		return err
	}

	return nil
}

// Destroy loads the value referenced by the token and dispenses of the token,
// making it unvailable for any further use.
func (ts *TokenStore) Destroy(token string, data interface{}) error {
	var err error

	tokenKey := fmt.Sprintf("%s::%s", tokenMapKey, token)
	if err = ts.peekToken(tokenKey, data); err != nil {
		return err
	}

	if _, err = ts.redis.Del(tokenKey).Result(); err != nil {
		return err
	}

	return nil
}

func (ts *TokenStore) peekToken(tokenKey string, data interface{}) error {
	var encoded string
	var err error

	if encoded, err = ts.redis.Get(tokenKey).Result(); err != nil {
		return err
	}

	if err = json.Unmarshal([]byte(encoded), data); err != nil {
		return err
	}

	return nil
}
