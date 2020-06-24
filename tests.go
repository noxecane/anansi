package anansi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockRequest records the http response to a generated request
func MockRequest(req *http.Request, handler http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	return rr
}

// GetRequest creates a mock GET request
func GetRequest(t *testing.T, path string) *http.Request {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	return req
}

// SearchRequest creates a mock GET request with query parameters
func SearchRequest(t *testing.T, path string, query map[string]string) *http.Request {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		t.Fatal(err)
	}

	urlQuery := req.URL.Query()
	for key, value := range query {
		urlQuery.Set(key, value)
	}
	req.URL.RawQuery = urlQuery.Encode()

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	return req
}

// PostRequest creates a mock POST request with a JSON body
func PostRequest(t *testing.T, path string, body interface{}) *http.Request {
	raw, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", path, bytes.NewBuffer(raw))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	return req
}

// PutRequest creates a mock PUT request with a JSON body
func PutRequest(t *testing.T, path string, body interface{}) *http.Request {
	raw, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("PUT", path, bytes.NewBuffer(raw))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	return req
}

// DeleteRequest creates a mock DELETE request
func DeleteRequest(t *testing.T, path string) *http.Request {
	req, err := http.NewRequest("DELETE", path, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	return req
}

// GetResponseBody creates a map of the JSON response body of off a mock request
func GetResponseBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}

	return body
}

// IsStatus confirms the status of a test response
func IsStatus(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}
