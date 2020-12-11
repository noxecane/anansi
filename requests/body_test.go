package requests

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	ozzo "github.com/go-ozzo/ozzo-validation/v4"
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

func TestReadBody(t *testing.T) {
	data := "some-strings"
	req := httptest.NewRequest("POST", "http://www.example.com", strings.NewReader(data))

	b, err := ReadBody(req)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != data {
		t.Errorf("Expected request body to be %s, got %s", data, string(b))
	}

	// make sure we can still read the body
	b2, _ := ioutil.ReadAll(req.Body)
	if !bytes.Equal(b, b2) {
		t.Errorf("Expected body to remain unchanged, got %s", string(b2))
	}
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
		err := ReadJSON(req, &n)
		if err != nil {
			t.Fatal(err)
		}

		if n.Name == "" {
			t.Error(`Expected Name to be set from request body`)
		}
	})

	t.Run("fails with wrong content type", func(t *testing.T) {
		data := `{ "extra": [1,2,3]}`
		req := httptest.NewRequest("POST", "http://www.example.com", strings.NewReader(data))
		req.Header.Add("Content-type", "application/x-www-form-urlencoded; charset=utf-8")

		var n noValidation
		err := ReadJSON(req, &n)
		if err == nil {
			t.Error("Expected read JSON to fail with error")
		}

		if err != ErrNotJSON {
			t.Errorf(`Expected ReadJSON to fail with ErrNotJSON, got %T`, err)
		}
	})

	t.Run("fails with validation error", func(t *testing.T) {
		data := `{ "extra": [1,2,3]}`
		req := httptest.NewRequest("POST", "http://www.example.com", strings.NewReader(data))
		req.Header.Add("Content-type", "application/json; charset=utf-8")

		var m myStruct
		err := ReadJSON(req, &m)
		if err == nil {
			t.Error("Expected read JSON to fail with error")
		}

		var e ozzo.Errors
		if !errors.As(err, &e) {
			t.Errorf(`Expected ReadJSON to fail with ErrValidation, got %T`, err)
		}
	})

	t.Run("fails with decoding error", func(t *testing.T) {
		data := `some-string`
		req := httptest.NewRequest("POST", "http://www.example.com", strings.NewReader(data))
		req.Header.Add("Content-type", "application/json; charset=utf-8")

		var m myStruct
		err := ReadJSON(req, &m)
		if err == nil {
			t.Error("Expected read JSON to fail with error")
		}

		var e ozzo.Errors
		if err == ErrNotJSON || errors.As(err, &e) {
			t.Errorf(`Expected ReadJSON to fail with decode error, got "%v"`, err)
		}
	})
}

func TestIDParam(t *testing.T) {
	router := chi.NewRouter()

	router.Get("/entities/{user_id}", func(rw http.ResponseWriter, r *http.Request) {
		id, err := IDParam(r, "user_id")
		if err != nil {
			http.Error(rw, err.Error(), 400)
			return
		}

		rw.WriteHeader(200)
		_, _ = rw.Write([]byte(fmt.Sprintf("%d", id)))
	})

	router.Get("/entities", func(_ http.ResponseWriter, r *http.Request) {
		// this should panic
		_, _ = IDParam(r, "user_id")
	})

	t.Run("returns a valid ID", func(t *testing.T) {
		id := faker.RandomInt(1, 30)

		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/entities/%d", id), nil)
		router.ServeHTTP(res, req)

		if res.Code != 200 {
			t.Fatalf("Expected response code to be 200, got %d", res.Code)
		}

		rid, err := strconv.Atoi(res.Body.String())
		if err != nil {
			t.Fatal(err)
		}

		if rid != id {
			t.Errorf("Expected request to return %d, got %d", id, rid)
		}
	})

	t.Run("returns a 400 when param is negative", func(t *testing.T) {
		id := faker.RandomInt(-10, -1)
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", fmt.Sprintf("/entities/%d", id), nil)
		router.ServeHTTP(res, req)

		if res.Code != 400 {
			t.Fatalf("Expected response code to be 400, got %d", res.Code)
		}
	})

	t.Run("returns a 400 when not an integer", func(t *testing.T) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/entities/some-invalid-id", nil)
		router.ServeHTTP(res, req)

		if res.Code != 400 {
			t.Fatalf("Expected response code to be 400, got %d", res.Code)
		}
	})

	t.Run("panics when param is not defined on route", func(t *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				t.Error("Expected handler to panic")
			}
		}()

		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/entities", nil)
		router.ServeHTTP(res, req)
	})
}

func TestStringParam(t *testing.T) {
	router := chi.NewRouter()

	router.Get("/entities", func(_ http.ResponseWriter, r *http.Request) {
		// this should panic
		_, _ = IDParam(r, "user_id")
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("panics when param is not defined on route", func(t *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				t.Error("Expected handler to panic")
			}
		}()

		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/entities", nil)
		router.ServeHTTP(res, req)
	})
}
