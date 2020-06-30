package ajax

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock records the http response to a generated request
func Mock(req *http.Request, handler http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	return rr
}

// Run runst the request and parses the response into v.
func Run(r *http.Request, v interface{}, client http.Client) error {
	var err error
	var resp *http.Response

	if resp, err = client.Do(r); err != nil {
		return err
	}

	// make sure to clean up the response
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(v)
}

// Get creates a mock GET request
func Get(t *testing.T, path string) *http.Request {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	return req
}

// Search creates a mock GET request with query parameters
func Search(t *testing.T, path string, query map[string]string) *http.Request {
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

// Post creates a mock POST request with a JSON body
func Post(t *testing.T, path string, body interface{}) *http.Request {
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

// Put creates a mock PUT request with a JSON body
func Put(t *testing.T, path string, body interface{}) *http.Request {
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

// Delete creates a mock DELETE request
func Delete(t *testing.T, path string) *http.Request {
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
