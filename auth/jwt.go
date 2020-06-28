package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const expiredErr = jwt.ValidationErrorExpired | jwt.ValidationErrorNotValidYet

var (
	ErrJWTExpired   = errors.New("Your JWT token has expired")
	ErrInvalidToken = errors.New("Your token is an invalid JWT token")
)

type jwtClaims struct {
	Data interface{} `json:"claim"`
	jwt.Claims
}

// EncodeJWT creates a JWT token for some given struct using the HMAC algorithm.
func EncodeJWT(secret []byte, t time.Duration, v interface{}) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		Data: v,
		Claims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(t).Unix(),
		},
	})

	return token.SignedString(secret)
}

// DecodeJWT extracts a struct from a JWT token using the HMAC algorithm
func DecodeJWT(secret []byte, token []byte, v *interface{}) error {
	claim := new(jwtClaims)
	t, err := jwt.ParseWithClaims(string(token), claim, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if !t.Valid {
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

	*v = claim.Data

	return nil
}
