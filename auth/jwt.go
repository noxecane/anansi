package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const expiredErr = jwt.ValidationErrorExpired | jwt.ValidationErrorNotValidYet

// ExpiresAt returns the time the duration ends
func ExpiresAt(d time.Duration) time.Time {
	return time.Now().Add(d)
}

// EncodeJWT creates a JWT token for the given struct using the HMAC algorithm.
func EncodeJWT(secret []byte, duration time.Duration, claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return str, nil
}

// DecodeJWT extracts a struct from a JWT token using the HMAC algorithm
func DecodeJWT(secret []byte, sessionToken []byte, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(string(sessionToken), claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if !token.Valid {
		if verr, ok := err.(*jwt.ValidationError); ok {
			switch {
			case verr.Errors&jwt.ValidationErrorMalformed != 0:
				return errors.New("This is not a JWT token")
			case verr.Errors&expiredErr != 0:
				return errors.New("This token has expired")
			default:
				return err
			}
		}
	}

	return nil
}
