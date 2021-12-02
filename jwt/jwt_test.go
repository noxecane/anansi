package jwt

import (
	"testing"
	"time"

	"syreclabs.com/go/faker"
)

type jwtStruct struct {
	Name string `json:"name"`
}

func TestEncodeDecode(t *testing.T) {
	secret := []byte("Die8ohsuyahno5dohL6oofaiShie3fie")

	t.Run("should encode and decode data", func(t *testing.T) {
		payload := jwtStruct{faker.Name().FirstName()}
		token, err := Encode(secret, time.Minute, payload)
		if err != nil {
			t.Fatal(err)
		}

		if token == "" {
			t.Error("Expected a token, got an empty string")
		}

		var parsed jwtStruct
		err = Decode(secret, token, &parsed)
		if err != nil {
			t.Fatal(err)
		}

		if parsed.Name != payload.Name {
			t.Errorf("Expected the parsed name to be %s, got %s", payload.Name, parsed.Name)
		}
	})

	t.Run("should expire after expiry is past", func(t *testing.T) {
		payload := map[string]interface{}{"name": faker.Name().FirstName()}
		token, err := Encode(secret, time.Second, payload)
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Second * 2)

		err = Decode(secret, token, &struct{}{})
		if err == nil {
			t.Fatal("Expected Decode to fail with non-nil error")
		}

		if err != ErrJWTExpired {
			t.Errorf("Expected Decode to fail with ErrJWTExpired, failed with %v", err)
		}
	})
}
