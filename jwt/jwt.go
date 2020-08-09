package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
)

const expiredErr = jwt.ValidationErrorExpired | jwt.ValidationErrorNotValidYet

var (
	ErrJWTExpired   = errors.New("Your JWT token has expired")
	ErrInvalidToken = errors.New("Your token is an invalid JWT token")
	ErrNoClaims     = errors.New("There are no claims in your token")
)

// Encode creates a JWT token for some given struct using the HMAC algorithm. It uses
// the key to separate the data stored from the JWT claims to prevent clashes. This works
// best with structs, please use the jwt-go library directly for primitive types.
func Encode(key string, secret []byte, t time.Duration, v interface{}) (string, error) {
	jwtClaims := jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(t).Unix(),
	}

	// save data using key
	jwtClaims[key] = v

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

	return token.SignedString(secret)
}

// Decode verifies a JWT token and extract the claims stored at key. Note that it expects v
// to be a pointer to a struct(based off what Encode does.)
func Decode(key string, secret []byte, tokenBytes []byte, v interface{}) error {
	token, err := jwt.Parse(string(tokenBytes), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if !token.Valid {
		if verr, ok := err.(*jwt.ValidationError); ok {
			switch {
			case verr.Errors&jwt.ValidationErrorMalformed != 0:
				return ErrInvalidToken
			case verr.Errors&expiredErr != 0:
				return ErrJWTExpired
			default:
				return err
			}
		}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("could not convert JWT to map claims")
	}

	// ignore tokens without claim data
	if claims[key] == nil {
		return nil
	}

	// convert claims data map to struct
	config := &mapstructure.DecoderConfig{Result: v, TagName: `json`}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(claims[key])
}
