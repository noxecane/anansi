package requests

import (
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"

	ozzo "github.com/go-ozzo/ozzo-validation/v4"
)

type myQuery struct {
	Name     string `json:"name" mod:"trim"`
	Quantity uint   `json:"quantity" mod:"trim"`
	Close    bool   `json:"should_close"`
}

func (m *myQuery) Validate() error {
	return ozzo.ValidateStruct(m,
		ozzo.Field(&m.Name, ozzo.Required),
		ozzo.Field(&m.Quantity, ozzo.Required, ozzo.Min(1)),
	)
}

func TestQueryParams(t *testing.T) {
	type noValidation struct {
		Name string `json:"name"`
	}

	t.Run("extract query params without validation", func(t *testing.T) {
		name := "daniel"
		req := httptest.NewRequest("GET", "http://www.example.com", nil)
		req.URL.RawQuery = url.Values{"name": []string{name}}.Encode()

		var n noValidation
		err := QueryParams(req, &n)
		if err != nil {
			t.Fatal(err)
		}

		if n.Name == "" {
			t.Error(`Expected Name to be set in query`)
		}

		if n.Name != name {
			t.Errorf(`Expected Name to be %s`, name)
		}
	})

	t.Run("fails with validation error", func(t *testing.T) {
		name := "daniel"
		req := httptest.NewRequest("GET", "http://www.example.com", nil)
		req.URL.RawQuery = url.Values{"name": []string{name}}.Encode()

		var n myQuery
		err := QueryParams(req, &n)
		if err == nil {
			t.Error("Expected QueryParam to fail with error")
		}

		var e ozzo.Errors
		if !errors.As(err, &e) {
			t.Errorf(`Expected QueryParam to fail with ErrValidation, got %T`, err)
		}
	})
}
