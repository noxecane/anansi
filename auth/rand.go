package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mrand "math/rand"
	"strconv"
	"strings"
	"time"
)

func init() {
	mrand.Seed(time.Now().UnixNano())
}

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

// RandomStringChoice returns a random string from an array of choices
func RandomStringChoice(elems []string) string {
	return elems[mrand.Intn(len(elems))]
}

// RandomIntChoice returns a random int from an array of choices
func RandomIntChoice(elems []int) int {
	return elems[mrand.Intn(len(elems))]
}

// RandomDigits returns a string of numbers based on len
func RandomDigits(len int) string {
	ps := mrand.Perm(9)
	var strps []string

	for _, v := range ps {
		strps = append(strps, strconv.Itoa(v))
	}

	return strings.Join(strps, "")
}

// RandomFormat generates a string based on format(fmt.Sprintf) and
// the choices given. It's up to the user to make sure the number of
// choices match the number of format operators.
func RandomFormat(format string, possibleChoices ...[]string) string {
	var choiceVals []interface{}

	for _, choices := range possibleChoices {
		choice := choices[mrand.Intn(len(choices))]
		choiceVals = append(choiceVals, choice)
	}

	return fmt.Sprintf(format, choiceVals...)
}
