package siber

import (
	"crypto/rand"
	"encoding/hex"
)

// RandomBytes returns securely generated random bytes. It will return an
// error if the system's secure random number generator fails to
// function correctly, in which case the caller should not continue.
func RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// RandomString returns a securely generated random hex string.
// It will return an error if the system's secure random  number generator fails
// to function correctly, in which case the caller should not continue. Keep in mind
// that s is the length of the returned string, not the number of bytes to be produced
func RandomString(s int) (string, error) {
	// use half of the size as hex always returns double the size.
	b, err := RandomBytes(s / 2)
	return hex.EncodeToString(b), err
}
