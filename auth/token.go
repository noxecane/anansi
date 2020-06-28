package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
)

const tokenMapKey = "general_token_map"

var (
	ErrTokenNotFound = errors.New("The passed token has either expired or never existed")
)

type TokenStore struct {
	redis  *redis.Client
	secret []byte
}

func NewTokenStore(r *redis.Client, secret []byte) *TokenStore {
	return &TokenStore{redis: r, secret: secret}
}

// Commission creates a single use token that expires after the given timeout. The token is a hash of a
// random 32 character key that can be used to revoke the token later on.
func (ts *TokenStore) Commission(t time.Duration, data interface{}) (token string, key string, err error) {
	var encoded []byte

	if key, err = RandomString(32); err != nil {
		return
	}

	// create a hash for the key
	if token, err = ts.getHash(key); err != nil {
		return
	}

	// TODO: replace this something lighter and faster
	if encoded, err = json.Marshal(data); err != nil {
		return
	}

	// One would naturally prefer hash maps but they don't support individual subkey expiry.
	tokenKey := fmt.Sprintf("%s::%s", tokenMapKey, token)

	if _, err = ts.redis.Set(tokenKey, encoded, t).Result(); err != nil {
		return
	}

	return
}

// CommissionWithKey creates a single use token that expires after the given timeout. Unlike New, it accepts a
// key rather can creating one
func (ts *TokenStore) CommissionWithKey(t time.Duration, key string, data interface{}) (string, error) {
	var encoded []byte
	var err error
	var token string

	// create a hash for the key
	if token, err = ts.getHash(key); err != nil {
		return "", err
	}

	// TODO: replace this something lighter and faster
	if encoded, err = json.Marshal(data); err != nil {
		return "", err
	}

	// One would naturally prefer hash maps but they don't support individual subkey expiry.
	tokenKey := fmt.Sprintf("%s::%s", tokenMapKey, token)

	if _, err = ts.redis.Set(tokenKey, encoded, t).Result(); err != nil {
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

// Decommission loads the value referenced by the token and dispenses of the token,
// making it unvailable for any further use.
func (ts *TokenStore) Decommission(token string, data interface{}) error {
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

// Revoke renders the token generated for the given key useless.
func (ts *TokenStore) Revoke(key string) error {
	var err error
	var token string

	if token, err = ts.getHash(key); err != nil {
		return err
	}

	var del int64
	tokenKey := fmt.Sprintf("%s::%s", tokenMapKey, token)
	if del, err = ts.redis.Del(tokenKey).Result(); err != nil {
		return err
	}

	// make sure it deleted a key, else no revocation happened
	if del == 0 {
		return ErrTokenNotFound
	}

	return nil
}

func (ts *TokenStore) getHash(key string) (string, error) {
	sig := hmac.New(sha256.New, ts.secret)
	if _, err := sig.Write([]byte(key)); err != nil {
		return "", err
	}

	return hex.EncodeToString(sig.Sum(nil)), nil
}

func (ts *TokenStore) peekToken(tokenKey string, data interface{}) error {
	var encoded string
	var err error

	if encoded, err = ts.redis.Get(tokenKey).Result(); err != nil {
		return ErrTokenNotFound
	}

	if err = json.Unmarshal([]byte(encoded), data); err != nil {
		return err
	}

	return nil
}
