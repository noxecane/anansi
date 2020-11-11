package tokens

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ErrTokenNotFound = errors.New("The passed token has either expired or never existed")
)

type Store struct {
	redis  *redis.Client
	secret []byte
}

func NewStore(r *redis.Client, secret []byte) *Store {
	return &Store{redis: r, secret: secret}
}

// Commission creates a single use token that expires after the given timeout.
func (ts *Store) Commission(ctx context.Context, t time.Duration, key string, data interface{}) (string, error) {
	var encoded []byte
	var err error
	var token string

	// create token from has of the key
	sig := hmac.New(sha256.New, ts.secret)
	if _, err := sig.Write([]byte(key)); err != nil {
		return "", err
	}
	token = hex.EncodeToString(sig.Sum(nil))

	// TODO: replace this something lighter and faster
	if encoded, err = json.Marshal(data); err != nil {
		return "", err
	}

	if _, err = ts.redis.Set(ctx, token, encoded, t).Result(); err != nil {
		return "", err
	}

	return token, nil
}

// Peek gets the data the token references without changing its lifetime.
func (ts *Store) Peek(ctx context.Context, token string, data interface{}) error {
	return ts.peekToken(ctx, token, data)
}

// Extend sets the new duration before an existing token times out. Note that it doesn't
// take into account how long the old token had to expire, as it uses the new duration
// entirely.
func (ts *Store) Extend(ctx context.Context, token string, timeout time.Duration, data interface{}) error {
	var ok bool
	var err error

	if err := ts.peekToken(ctx, token, data); err != nil {
		return err
	}

	if ok, err = ts.redis.Expire(ctx, token, timeout).Result(); err != nil {
		return err
	}

	if !ok {
		return ErrTokenNotFound
	}

	return nil
}

// Reset changes the contents of the token without changing it's TTL
func (ts *Store) Reset(ctx context.Context, key string, data interface{}) error {
	var err error
	var encoded []byte
	var token string

	// recreate the token from the  key.
	sig := hmac.New(sha256.New, ts.secret)
	if _, err := sig.Write([]byte(key)); err != nil {
		return err
	}
	token = hex.EncodeToString(sig.Sum(nil))

	// make sure the token existed before.
	if _, err = ts.redis.Get(ctx, token).Result(); err != nil {
		if err == redis.Nil {
			return ErrTokenNotFound
		}

		return err
	}

	//	we already know the key exists
	ttl, err := ts.redis.TTL(ctx, token).Result()
	if err != nil {
		return err
	}

	// TODO: replace this something lighter and faster
	if encoded, err = json.Marshal(data); err != nil {
		return err
	}

	if _, err = ts.redis.Set(ctx, token, encoded, ttl).Result(); err != nil {
		return err
	}
	return nil
}

// Decommission loads the value referenced by the token and dispenses of the token,
// making it unvailable for further use.
func (ts *Store) Decommission(ctx context.Context, token string, data interface{}) error {
	var err error

	if err = ts.peekToken(ctx, token, data); err != nil {
		return err
	}

	if _, err = ts.redis.Del(ctx, token).Result(); err != nil {
		return err
	}

	return nil
}

// Revoke renders the token generated for the given key useless.
func (ts *Store) Revoke(ctx context.Context, key string) error {
	var err error
	var token string

	sig := hmac.New(sha256.New, ts.secret)
	if _, err := sig.Write([]byte(key)); err != nil {
		return err
	}

	token = hex.EncodeToString(sig.Sum(nil))

	var del int64
	if del, err = ts.redis.Del(ctx, token).Result(); err != nil {
		return err
	}

	// make sure it deleted a key, else no revocation happened
	if del == 0 {
		return ErrTokenNotFound
	}

	return nil
}

func (ts *Store) peekToken(ctx context.Context, tokenKey string, data interface{}) error {
	var encoded string
	var err error

	if encoded, err = ts.redis.Get(ctx, tokenKey).Result(); err != nil {
		if err == redis.Nil {
			return ErrTokenNotFound
		}

		return err
	}

	if err = json.Unmarshal([]byte(encoded), data); err != nil {
		return err
	}

	return nil
}
