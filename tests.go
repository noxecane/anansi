package anansi

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
)

type Dict map[string]interface{}

func EnableFaker() {
	rand.Seed(time.Now().Unix())
	faker.SetGenerateUniqueValues(true)
}

// MockRequest records the http response to a generated request
func MockRequest(req *http.Request, handler http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	return rr
}

// IsStatus confirms the status of a test response
func IsStatus(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func GetRequest(t *testing.T, path string) *http.Request {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	return req
}

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

// GetResponseBody creates a map of the JSON response body of off
// a test request
func GetResponseBody(t *testing.T, rr *httptest.ResponseRecorder) Dict {
	var body map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}

	return body
}
