package siber

import (
	"encoding/base64"
	"testing"

	"syreclabs.com/go/faker"
)

func TestEncryptDecrypt(t *testing.T) {
	secret := []byte("simple-test-secret")

	t.Run("Encrypt creates an encrypted string", func(t *testing.T) {
		someString := faker.Lorem().Characters(16)
		enc, err := Encrypt(secret, []byte(someString))
		if err != nil {
			t.Fatal(err)
		}

		if enc == "" {
			t.Error("Expected a encrypted string to be defined, got an empty string")
		}

		dec, err := Decrypt(secret, enc)
		if err != nil {
			t.Fatal(err)
		}

		if string(dec) != someString {
			t.Errorf("Expected %s, got %s", someString, string(dec))
		}
	})

	t.Run("Decrypt for normal strings", func(t *testing.T) {
		someString := faker.Lorem().Characters(64)

		_, err := Decrypt(secret, someString)
		if err == nil {
			t.Error("Expected Decrypt to fail for normal strings")
		}

		encoded := base64.URLEncoding.EncodeToString([]byte(someString))
		_, err = Decrypt(secret, encoded)
		if err == nil {
			t.Error("Expected Decrypt to fail for base64 encoded string")
		}
	})
}
