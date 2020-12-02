package requests

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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
	b2, err := ioutil.ReadAll(req.Body)
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

		if len(n.Name) == 0 {
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
		rw.Write([]byte(fmt.Sprintf("%d", id)))
	})

	router.Get("/entities", func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				rw.WriteHeader(500)
				rw.Write([]byte(""))
			}
		}()

		// this should panic
		_, _ = IDParam(r, "user_id")
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("returns a valid ID", func(t *testing.T) {
		id := faker.RandomInt(1, 30)
		resp, err := testRequest(ts, "GET", fmt.Sprintf("/entities/%d", id), nil)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != 200 {
			t.Fatalf("Expected response code to be 200, got %d", resp.StatusCode)
		}

		raw, err := testResp(resp)
		if err != nil {
			t.Fatal(err)
		}

		rid, err := strconv.Atoi(string(raw))
		if err != nil {
			t.Fatal(err)
		}

		if rid != id {
			t.Errorf("Expected request to return %d, got %d", id, rid)
		}
	})

	t.Run("returns a 400 when param is negative", func(t *testing.T) {
		id := faker.RandomInt(-10, -1)
		resp, err := testRequest(ts, "GET", fmt.Sprintf("/entities/%d", id), nil)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 400 {
			t.Fatalf("Expected response code to be 400, got %d", resp.StatusCode)
		}
	})

	t.Run("returns a 400 when not an integer", func(t *testing.T) {
		resp, err := testRequest(ts, "GET", "/entities/some-id", nil)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 400 {
			t.Fatalf("Expected response code to be 400, got %d", resp.StatusCode)
		}
	})

	t.Run("panics when param is not defined on route", func(t *testing.T) {
		resp, err := testRequest(ts, "GET", "/entities", nil)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 500 {
			t.Fatalf("Expected response code to be 400, got %d", resp.StatusCode)
		}
	})
}

func TestStringParam(t *testing.T) {
	router := chi.NewRouter()

	router.Get("/entities", func(rw http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				rw.WriteHeader(500)
				rw.Write([]byte(""))
			}
		}()

		// this should panic
		_, _ = IDParam(r, "user_id")
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("panics when param is not defined on route", func(t *testing.T) {
		resp, err := testRequest(ts, "GET", "/entities", nil)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 500 {
			t.Fatalf("Expected response code to be 400, got %d", resp.StatusCode)
		}
	})
}

func testRequest(ts *httptest.Server, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req)
}

func testResp(resp *http.Response) ([]byte, error) {
	respBody, err := ioutil.ReadAll(resp.Body)
	// only close if no errors occured
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return respBody, nil
}
