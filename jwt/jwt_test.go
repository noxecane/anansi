package jwt

import (
	"fmt"
	"testing"
	"time"

	"syreclabs.com/go/faker"
)

type jwtStruct struct {
	Name string `json:"name"`
}

func TestEncode(t *testing.T) {
	secret := []byte("test-secret")

	t.Run("should encode a map", func(t *testing.T) {
		payload := map[string]interface{}{"name": faker.Name().FirstName()}
		token, err := Encode(secret, time.Minute, payload)
		if err != nil {
			t.Fatal(err)
		}

		if token == "" {
			t.Error("Expected a token, got an empty string")
		}

		parsed, err := Decode(secret, []byte(token))
		if err != nil {
			t.Fatal(err)
		}

		if parsed["name"] != payload["name"] {
			t.Errorf("Expected the parsed name to be %s, got %s", payload["name"], parsed["name"])
		}
	})

	t.Run("should fail with an error", func(t *testing.T) {
		payload := map[string]interface{}{"name": faker.Name().FirstName()}
		token, err := Encode(secret, time.Second, payload)
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Second * 2)

		_, err = Decode(secret, []byte(token))
		if err == nil {
			t.Fatal("Expected Decode to fail with non-nil error")
		}

		if err != ErrJWTExpired {
			t.Errorf("Expected Decode to fail with ErrJWTExpired, failed with %v", err)
		}
	})
}

func TestEncodeStruct(t *testing.T) {
	secret := []byte("test-secret")
	payload := jwtStruct{faker.Name().FirstName()}

	token, err := EncodeStruct(secret, time.Second, payload)
	if err != nil {
		t.Fatal(err)
	}

	if token == "" {
		t.Error("Expected a token, got an empty string")
	}

	var parsed jwtStruct
	err = DecodeStruct(secret, []byte(token), &parsed)
	if err != nil {
		t.Fatal(err)
	}

	if parsed.Name != payload.Name {
		t.Errorf("Expected the parsed name to be %s, got %s", payload.Name, parsed.Name)
	}
}

func TestEncodeEmbedded(t *testing.T) {
	secret := []byte("test-secret")
	payload := jwtStruct{faker.Name().FirstName()}

	token, err := EncodeEmbedded(secret, time.Second, payload)
	if err != nil {
		t.Fatal(err)
	}

	if token == "" {
		t.Error("Expected a token, got an empty string")
	}

	parsed, err := Decode(secret, []byte(token))
	if err != nil {
		t.Fatal(err)
	}

	if parsed["claim"] == nil {
		t.Fatal("Expected the claim to be non-nil")
	}

	data, ok := parsed["claim"].(map[string]interface{})
	fmt.Println(data)
	if !ok {
		t.Fatalf("Expected claim to be a map of string to string, got %T", parsed["claim"])
	}

	if data["name"] != payload.Name {
		t.Errorf("Expected the parsed name to be %s, got %s", payload.Name, data["name"])
	}
}

func TestDecodeEmbedded(t *testing.T) {
	secret := []byte("test-secret")
	payload := jwtStruct{faker.Name().FirstName()}

	token, err := EncodeEmbedded(secret, time.Second, payload)
	if err != nil {
		t.Fatal(err)
	}

	if token == "" {
		t.Error("Expected a token, got an empty string")
	}

	var parsed jwtStruct
	err = DecodeEmbedded(secret, []byte(token), &parsed)
	if err != nil {
		t.Fatal(err)
	}

	if parsed.Name != payload.Name {
		t.Errorf("Expected the parsed name to be %s, got %s", payload.Name, parsed.Name)
	}
}
