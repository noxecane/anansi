package jsend

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadBearer(t *testing.T) {
	type void struct{}

	t.Run("panics with empty header message", func(t *testing.T) {
		message := "Your request is not authenticated"
		defer checkErr(t, http.StatusUnauthorized, false, false, message)

		req := httptest.NewRequest("GET", "/", nil)
		LoadBearer(store, req, void{})
	})

	t.Run("panics with bad header format message", func(t *testing.T) {
		message := "Your authorization header is incorrect"
		defer checkErr(t, http.StatusUnauthorized, false, false, message)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer southern cross")
		LoadBearer(store, req, void{})
	})

	t.Run("panics with bad scheme message", func(t *testing.T) {
		message := "We don't support your authorization scheme"
		defer checkErr(t, http.StatusUnauthorized, false, false, message)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Southern cross")
		LoadBearer(store, req, void{})
	})

	t.Run("panics with invalid token message", func(t *testing.T) {
		message := "Your token is either invalid or has expired"
		defer checkErr(t, http.StatusUnauthorized, false, false, message)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer southerncross")
		LoadBearer(store, req, void{})
	})
}

func TestLoadHeadless(t *testing.T) {
	type void struct{}

	t.Run("panics with empty header message", func(t *testing.T) {
		message := "Your request is not authenticated"
		defer checkErr(t, http.StatusUnauthorized, false, false, message)

		req := httptest.NewRequest("GET", "/", nil)
		LoadHeadless(store, req, void{})
	})

	t.Run("panics with bad header format message", func(t *testing.T) {
		message := "Your authorization header is incorrect"
		defer checkErr(t, http.StatusUnauthorized, false, false, message)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Test southern cross")
		LoadHeadless(store, req, void{})
	})

	t.Run("panics with bad scheme message", func(t *testing.T) {
		message := "We don't support your authorization scheme"
		defer checkErr(t, http.StatusUnauthorized, false, false, message)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer cross")
		LoadHeadless(store, req, void{})
	})

	t.Run("panics with invalid token message", func(t *testing.T) {
		message := "Your token is either invalid or has expired"
		defer checkErr(t, http.StatusUnauthorized, false, false, message)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Test southerncross")
		LoadHeadless(store, req, void{})
	})
}

func checkErr(t *testing.T, code int, nilErr, nilData bool, message string) {
	err := recover()
	if err == nil {
		t.Fatal("Expected ReadJSON to panic")
	}

	e, ok := err.(Err)
	if !ok {
		t.Fatal("Expected ReadJSON to panic with Err type")
	}

	if e.Code != code {
		t.Errorf("Expected the status code to be %d, got %d", code, e.Code)
	}

	if nilErr && e.Err == nil {
		t.Errorf("Expected source of the error to be set")
	}

	if nilData && e.Data == nil {
		t.Errorf("Expected metadata for the error to be set")
	}

	if message != "" && e.Message != message {
		t.Errorf("Expected error message to be %s, got %s", message, e.Message)
	}
}
