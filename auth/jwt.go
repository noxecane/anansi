package auth

import (
	"encoding/json"
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

// EncodeJWT creates a JWT token for some given struct using the HMAC algorithm.
func EncodeJWT(secret []byte, t time.Duration, v interface{}) (string, error) {
	str, _ := json.Marshal(v)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"claims": string(str),
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(t).Unix(),
	})

	return token.SignedString(secret)
}

// DecodeJWT extracts a struct from a JWT token using the HMAC algorithm
func DecodeJWT(secret []byte, token []byte, v interface{}) error {
	t, err := jwt.Parse(string(token), func(token *jwt.Token) (interface{}, error) {
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

	if claims, ok := t.Claims.(jwt.MapClaims); !ok {
		return errors.New("Could not convert JWT to map claims")
	} else {
		b := claims["claims"].(string)
		return json.Unmarshal([]byte(b), v)
	}
}
