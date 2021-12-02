package jwt

import (
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

var (
	ErrJWTExpired   = errors.New("token has expired")
	ErrInvalidToken = errors.New("token is an invalid")
)

type CustomClaim struct {
	jwt.Claims
	CustomClaims interface{} `json:"urn:custom:claims"`
}

// Encode encodes and encrypts claims as JWE. Note that the claim passed is wrapped to prevent clash
// Make sure your secret is at least 32 bytes
func Encode(secret []byte, t time.Duration, v interface{}) (string, error) {
	enc, err := jose.NewEncrypter(
		jose.A256GCM,
		jose.Recipient{Algorithm: jose.DIRECT, Key: secret},
		&jose.EncrypterOptions{ExtraHeaders: map[jose.HeaderKey]interface{}{jose.HeaderType: "JWT"}},
	)
	if err != nil {
		return "", err
	}

	if c, ok := v.(CustomClaim); ok {
		c.IssuedAt = jwt.NewNumericDate(time.Now())
		c.Expiry = jwt.NewNumericDate(time.Now().Add(t))

		return jwt.Encrypted(enc).Claims(c).CompactSerialize()
	} else {
		def := jwt.Claims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Expiry:   jwt.NewNumericDate(time.Now().Add(t)),
		}
		c := CustomClaim{Claims: def, CustomClaims: v}

		return jwt.Encrypted(enc).Claims(c).CompactSerialize()
	}
}

// Decodes and decrypts a JWE token. Note that it expects the claim to be wrapped
// using `urn:custom:claims`. Make sure your secret is at least 32 bytes
func Decode(secret []byte, token string, v interface{}) error {
	tok, err := jwt.ParseEncrypted(token)
	if err != nil {
		return err
	}

	var claims CustomClaim
	if err := tok.Claims(secret, &claims); err != nil {
		return ErrInvalidToken
	}

	if err := claims.ValidateWithLeeway(jwt.Expected{Time: time.Now()}, 0); err != nil {
		if err == jwt.ErrExpired {
			return ErrJWTExpired
		}
		return err
	}

	if claims.CustomClaims == nil {
		return nil
	}

	// convert claims data map to struct
	config := &mapstructure.DecoderConfig{Result: v, TagName: `json`}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return errors.Wrap(err, "could not convert claims to struct")
	}

	return decoder.Decode(claims.CustomClaims)
}
