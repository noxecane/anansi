package siber

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/secretbox"
)

// Encrypt encrypts a string using secretbox
func Encrypt(secret []byte, value []byte) (string, error) {
	// create a random nonce
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return "", err
	}

	// only copy 32 bytes off the secret
	var secretKey [32]byte
	copy(secretKey[:], secret)

	// offset the encrypted message with the nonce for decryption
	encrypted := secretbox.Seal(nonce[:], value, &nonce, &secretKey)

	// the bytes generated are unusable
	return base64.URLEncoding.EncodeToString(encrypted), nil
}

// Decrypt decrypts a string using secretbox
func Decrypt(secret []byte, encrypted string) ([]byte, error) {
	var encBytes []byte
	var err error

	if encBytes, err = base64.URLEncoding.DecodeString(encrypted); err != nil {
		return encBytes, err
	}

	// extract the nonce from message
	var decryptNonce [24]byte
	copy(decryptNonce[:], encBytes[:24])

	// only copy 32 bytes off the secret
	var secretKey [32]byte
	copy(secretKey[:], secret)

	var decrypted []byte
	var ok bool
	if decrypted, ok = secretbox.Open(nil, encBytes[24:], &decryptNonce, &secretKey); !ok {
		return nil, errors.New("Could not decrypt your message")
	}

	return decrypted, nil
}
