package auth

import (
	"testing"
	"time"
)

type jwtStruct struct {
	Name string `json:"name"`
}

func TestEncodeJWT(t *testing.T) {
	jwt := jwtStruct{Name: "Olakunkle"}
	token, err := EncodeJWT([]byte("mysecret"), time.Minute, jwt)
	if err != nil {
		t.Fatal(err)
	}

	if token == "" {
		t.Error("Expected EncodeJWT to generate a token")
	}
}

func TestDecodeJWT(t *testing.T) {
	jwt := jwtStruct{Name: "Olakunle"}
	token, err := EncodeJWT([]byte("mysecret"), time.Minute, jwt)
	if err != nil {
		t.Fatal(err)
	}

	// token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiIiLCJpc3QiOiIiLCJpYXQiOjE1OTU3OTE0NDAsImV4cCI6MTU5NTc5NTA0MCwianRpIjoiand0X25vbmNlIn0.cBCDA2bE5tc9VMhD8bJuDK4_iYACPgFluFJQUaujqV8"
	var loaded jwtStruct
	if err := DecodeJWT([]byte("mysecret"), []byte(token), &loaded); err != nil {
		t.Fatal(err)
	}

	if loaded.Name != "Olakunle" {
		t.Errorf("Expected Name to be %s, got %s", "Olakunle", loaded.Name)
	}
}
