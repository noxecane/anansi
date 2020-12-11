package jwt

import (
	"fmt"
	"reflect"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

const expiredErr = jwt.ValidationErrorExpired | jwt.ValidationErrorNotValidYet

var (
	ErrJWTExpired   = errors.New("token has expired")
	ErrInvalidToken = errors.New("token is an invalid")
	ErrNoClaims     = errors.New("no claims in token")
)

// Encodes generates and signs a JWT token for the given payload using the HMAC algorithm.
func Encode(secret []byte, t time.Duration, payload map[string]interface{}) (string, error) {
	jwtClaims := jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(t).Unix(),
	}

	for k, v := range payload {
		jwtClaims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)

	return token.SignedString(secret)
}

// EncodeStruct generates a JWT token for the given struct using the HMAC algorithm.
func EncodeStruct(secret []byte, t time.Duration, v interface{}) (string, error) {
	payload := make(map[string]interface{})

	r := reflect.ValueOf(v)
	if r.Kind() == reflect.Ptr {
		r = r.Elem()
	}

	// we only accept structs
	if r.Kind() != reflect.Struct {
		return "", errors.Errorf("can only encode structs; got %T", v)
	}

	typ := r.Type()
	for i := 0; i < r.NumField(); i++ {
		ft := typ.Field(i)

		// use json tag if available
		n := ft.Tag.Get("json")
		if n == "" {
			n = ft.Name
		}

		payload[n] = r.Field(i).Interface()
	}

	return Encode(secret, t, payload)
}

// EncodeEmbedded attaches the payload as an entry to the final claim using the key
// `claim` to prevent clashes with JWT field names.
func EncodeEmbedded(secret []byte, t time.Duration, v interface{}) (string, error) {
	return Encode(secret, t, map[string]interface{}{"claim": v})
}

// Decode validates and parses the given JWT token into a map
func Decode(secret []byte, token []byte) (map[string]interface{}, error) {
	jwtToken, err := jwt.Parse(string(token), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if jwtToken == nil || !jwtToken.Valid {
		if verr, ok := err.(*jwt.ValidationError); ok {
			switch {
			case verr.Errors&jwt.ValidationErrorMalformed != 0:
				return nil, ErrInvalidToken
			case verr.Errors&expiredErr != 0:
				return nil, ErrJWTExpired
			default:
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("could not convert JWT to map claims")
	}

	return claims, nil
}

// DecodeStruct validates and parses a JWT token into a struct.
func DecodeStruct(secret []byte, tokenBytes []byte, v interface{}) error {
	claims, err := Decode(secret, tokenBytes)
	if err != nil {
		return err
	}

	// convert claims data map to struct
	config := &mapstructure.DecoderConfig{Result: v, TagName: `json`}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return errors.Wrap(err, "could not convert claims to struct")
	}

	return decoder.Decode(claims)
}

// DecodeEmbedded validates and parses a JWT token into a struct. It expects the
// struct's payload to be attached to the key `claim` of the actual JWT claim. Note that
// the struct should have json tags
func DecodeEmbedded(secret []byte, tokenBytes []byte, v interface{}) error {
	claims, err := Decode(secret, tokenBytes)
	if err != nil {
		return err
	}

	if claims["claim"] == nil {
		return nil
	}

	// convert claims data map to struct
	config := &mapstructure.DecoderConfig{Result: v, TagName: `json`}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return errors.Wrap(err, "could not convert claims to struct")
	}

	return decoder.Decode(claims["claim"])
}
