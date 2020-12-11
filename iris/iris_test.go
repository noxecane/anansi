package iris

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"syreclabs.com/go/faker"
)

func TestNewRequest(t *testing.T) {
	t.Run("sets the right headers", func(t *testing.T) {
		auth := "Bearer " + faker.Lorem().Characters(32)
		requestID := faker.Lorem().Characters(16)
		service := faker.Company().Name()

		req := httptest.NewRequest("GEt", "/", nil)
		req.Header.Set("Authorization", auth)
		req.Header.Set("X-Request-ID", requestID)

		client := NewClient(Config{
			Secret:         []byte("secret"),
			Service:        service,
			HeadlessScheme: "scheme",
		})

		req2, err := client.NewRequest(req, "GET", "/internal", nil)
		if err != nil {
			t.Fatal(err)
		}

		if req2.Header.Get("X-Request-ID") != requestID {
			t.Errorf("Expected request ID to be %s, got %s", requestID, req2.Header.Get("X-Request-ID"))
		}

		if req2.Header.Get("Authorization") != auth {
			t.Errorf("Expected authorization header to be set to %s, got %s", auth, req2.Header.Get("Authorization"))
		}

		if req2.Header.Get("X-Origin-Service") != service {
			t.Errorf("Expected origin service of request to be %s, got %s", service, req2.Header.Get("X-Origin-Service"))
		}
	})

	t.Run("times out when parent request times out", func(t *testing.T) {
		client := NewClient(Config{
			Secret:         []byte("secret"),
			Service:        faker.Company().Name(),
			HeadlessScheme: "scheme",
		})

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(time.Millisecond * 800)
			_, _ = w.Write([]byte("too late"))
		}))
		defer server.Close()

		req := httptest.NewRequest("GEt", "/", nil)
		req.Header.Set("Authorization", "Bearer "+faker.Lorem().Characters(32))
		req.Header.Set("X-Request-ID", faker.Lorem().Characters(16))

		ctx, cancel := context.WithTimeout(req.Context(), time.Millisecond*500)
		defer cancel()
		req = req.WithContext(ctx)

		req2, err := client.NewRequest(req, "GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}

		httpClient := server.Client()
		resp, err := httpClient.Do(req2)
		if err == nil {
			t.Fatal("Expected request to fail with error")
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(string(b))

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected request to fail because deadline was exceeded, got %v", err)
		}
	})
}
