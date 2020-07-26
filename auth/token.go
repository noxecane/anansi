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

var (
	ErrTokenNotFound = errors.New("The passed token has either expired or never existed")
)

type TokenStore struct {
	redis     *redis.Client
	namespace string
	secret    []byte
}

func NewTokenStore(r *redis.Client, secret []byte) *TokenStore {
	return &TokenStore{redis: r, secret: secret}
}

// Commission creates a single use token that expires after the given timeout.
func (ts *TokenStore) Commission(t time.Duration, key string, data interface{}) (string, error) {
	var encoded []byte
	var err error
	var token string

	// create a hash for the key
	sig := hmac.New(sha256.New, ts.secret)
	if _, err := sig.Write([]byte(key)); err != nil {
		return "", err
	}

	token = hex.EncodeToString(sig.Sum(nil))

	// TODO: replace this something lighter and faster
	if encoded, err = json.Marshal(data); err != nil {
		return "", err
	}

	if _, err = ts.redis.Set(token, encoded, t).Result(); err != nil {
		return "", err
	}

	return token, nil
}

// Peek gets the data the token references without changing its lifetime.
func (ts *TokenStore) Peek(token string, data interface{}) error {
	return ts.peekToken(token, data)
}

// Refresh loads the data the token references and refreshes it's lifetime so it can last
// for as long as the given timeout. Note that it doesn't add timeout to the existing
// lifetime but rather gives it a new one afresh.
func (ts *TokenStore) Refresh(token string, timeout time.Duration, data interface{}) error {
	var err error

	if err = ts.peekToken(token, data); err != nil {
		return err
	}

	if _, err = ts.redis.Expire(token, timeout).Result(); err != nil {
		return err
	}

	return nil
}

// Reset changes the contents of the token without changing it's TTL
func (ts *TokenStore) Reset(key string, timeout time.Duration, data interface{}) error {
	var err error
	var encoded []byte
	var token string

	sig := hmac.New(sha256.New, ts.secret)
	if _, err := sig.Write([]byte(key)); err != nil {
		return err
	}

	token = hex.EncodeToString(sig.Sum(nil))

	if _, err = ts.redis.Get(token).Result(); err != nil {
		return ErrTokenNotFound
	}

	//	we already know the key exists
	ttl, _ := ts.redis.TTL(token).Result()

	// TODO: replace this something lighter and faster
	if encoded, err = json.Marshal(data); err != nil {
		return err
	}

	if _, err = ts.redis.Set(token, encoded, ttl).Result(); err != nil {
		return err
	}
	return nil
}

// Decommission loads the value referenced by the token and dispenses of the token,
// making it unvailable for further use.
func (ts *TokenStore) Decommission(token string, data interface{}) error {
	var err error

	if err = ts.peekToken(token, data); err != nil {
		return err
	}

	if _, err = ts.redis.Del(token).Result(); err != nil {
		return err
	}

	return nil
}

// Revoke renders the token generated for the given key useless.
func (ts *TokenStore) Revoke(key string) error {
	var err error
	var token string

	sig := hmac.New(sha256.New, ts.secret)
	if _, err := sig.Write([]byte(key)); err != nil {
		return err
	}

	token = hex.EncodeToString(sig.Sum(nil))

	var del int64
	tokenKey := fmt.Sprintf("%s::%s", ts.namespace, token)
	if del, err = ts.redis.Del(tokenKey).Result(); err != nil {
		return err
	}

	// make sure it deleted a key, else no revocation happened
	if del == 0 {
		return ErrTokenNotFound
	}

	return nil
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
