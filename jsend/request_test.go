package jsend

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	ozzo "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/random-guys/go-siber/json"
	"syreclabs.com/go/faker"
)

type myStruct struct {
	Name string `json:"name"`
}

func (m *myStruct) Validate() error {
	return ozzo.ValidateStruct(m,
		ozzo.Field(&m.Name, ozzo.Required),
	)
}

func TestIDParam(t *testing.T) {
	router := chi.NewRouter()
	errMessage := "user_id must be an integer ID"

	router.Use(Recoverer("production"))
	router.Get("/entities/{user_id}", func(_ http.ResponseWriter, r *http.Request) {
		_ = IDParam(r, "user_id")
	})

	t.Run("returns a 400 when param is negative", func(t *testing.T) {
		id := faker.RandomInt(-10, -1)
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/entities/%d", id), nil)
		router.ServeHTTP(res, req)

		resp := Err{}
		err := json.Unmarshal(res.Body.Bytes(), &resp)
		if err != nil {
			t.Fatal(err)
		}

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected the status code to be %d, got %d", http.StatusBadRequest, resp.Code)
		}

		if resp.Message != errMessage {
			t.Errorf("Expected error message to be %s, got %s", errMessage, resp.Message)
		}
	})

	t.Run("returns a 400 when not an integer", func(t *testing.T) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/entities/some-invalid-id", nil)
		router.ServeHTTP(res, req)

		resp := Err{}
		err := json.Unmarshal(res.Body.Bytes(), &resp)
		if err != nil {
			t.Fatal(err)
		}

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected the status code to be %d, got %d", http.StatusBadRequest, resp.Code)
		}

		if resp.Message != errMessage {
			t.Errorf("Expected error message to be %s, got %s", errMessage, resp.Message)
		}
	})
}

func TestReadJSON(t *testing.T) {
	type noValidation struct {
		Name string `json:"name"`
	}

	t.Run("parses from JSON struct without validation", func(t *testing.T) {
		data := `{ "name": "Yuko Omo", "age": 24, "extra": [1,2,3]}`
		req := httptest.NewRequest("POST", "http://www.example.com", strings.NewReader(data))
		req.Header.Add("Content-type", "application/json; charset=utf-8")

		var n noValidation
		ReadJSON(req, &n)

		if n.Name == "" {
			t.Error(`Expected Name to be set from request body`)
		}
	})

	t.Run("panics with 415 with wrong content type", func(t *testing.T) {
		defer func() {
			err := recover()
			if err == nil {
				t.Fatal("Expected ReadJSON to panic")
			}

			e, ok := err.(Err)
			if !ok {
				t.Fatal("Expected ReadJSON to panic with Err type")
			}

			if e.Code != http.StatusUnsupportedMediaType {
				t.Errorf("Expected the status code to be %d, got %d", http.StatusUnsupportedMediaType, e.Code)
			}
		}()

		data := `{ "extra": [1,2,3]}`
		req := httptest.NewRequest("POST", "http://www.example.com", strings.NewReader(data))
		req.Header.Add("Content-type", "application/x-www-form-urlencoded; charset=utf-8")

		var n noValidation
		ReadJSON(req, &n)
	})

	t.Run("fails with validation error", func(t *testing.T) {
		defer func() {
			err := recover()
			if err == nil {
				t.Fatal("Expected ReadJSON to panic")
			}

			e, ok := err.(Err)
			if !ok {
				t.Fatal("Expected ReadJSON to panic with Err type")
			}

			if e.Code != http.StatusUnprocessableEntity {
				t.Errorf("Expected the status code to be %d, got %d", http.StatusUnprocessableEntity, e.Code)
			}

			if e.Data == nil {
				t.Errorf("Expected metadata for the error to be set")
			}

			errMsg := "We could not validate your request."
			if e.Message != errMsg {
				t.Errorf("Expected error message to be %s, got %s", errMsg, e.Message)
			}
		}()

		data := `{ "extra": [1,2,3]}`
		req := httptest.NewRequest("POST", "http://www.example.com", strings.NewReader(data))
		req.Header.Add("Content-type", "application/json; charset=utf-8")

		var m myStruct
		ReadJSON(req, &m)
	})

	t.Run("fails with decoding error", func(t *testing.T) {
		defer func() {
			err := recover()
			if err == nil {
				t.Fatal("Expected ReadJSON to panic")
			}

			e, ok := err.(Err)
			if !ok {
				t.Fatal("Expected ReadJSON to panic with Err type")
			}

			if e.Code != http.StatusBadRequest {
				t.Errorf("Expected the status code to be %d, got %d", http.StatusBadRequest, e.Code)
			}

			if e.Err == nil {
				t.Errorf("Expected source of the error to be set")
			}

			errMsg := "We cannot parse your request body."
			if e.Message != errMsg {
				t.Errorf("Expected error message to be %s, got %s", errMsg, e.Message)
			}
		}()

		data := `some-string`
		req := httptest.NewRequest("POST", "http://www.example.com", strings.NewReader(data))
		req.Header.Add("Content-type", "application/json; charset=utf-8")

		var m myStruct
		ReadJSON(req, &m)
	})
}
