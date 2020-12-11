package iris

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/random-guys/go-siber/jwt"
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
		_, err = httpClient.Do(req2)
		if err == nil {
			t.Fatal("Expected request to fail with error")
		}

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected request to fail because deadline was exceeded, got %v", err)
		}
	})
}

func TestNewHeadlessRequest(t *testing.T) {
	t.Run("sets the right headers", func(t *testing.T) {
		type session struct{ User string }

		requestID := faker.Lorem().Characters(16)
		service := "user-service"
		scheme := "Test"

		req := httptest.NewRequest("GEt", "/", nil)
		req.Header.Set("Authorization", "Bearer "+faker.Lorem().Characters(32))
		req.Header.Set("X-Request-ID", requestID)

		client := NewClient(Config{
			Secret:         []byte("secret"),
			Service:        service,
			HeadlessScheme: scheme,
		})

		sput := session{uuid.New().String()}
		req2, err := client.NewHeadlessRequest(req, "GET", "/internal", sput, nil)
		if err != nil {
			t.Fatal(err)
		}

		if req2.Header.Get("X-Request-ID") != requestID {
			t.Errorf("Expected request ID to be %s, got %s", requestID, req2.Header.Get("X-Request-ID"))
		}

		if req2.Header.Get("X-Origin-Service") != service {
			t.Errorf("Expected origin service of request to be %s, got %s", service, req2.Header.Get("X-Origin-Service"))
		}

		header := strings.Fields(req2.Header.Get("Authorization"))
		if header[0] != scheme {
			t.Errorf("Expected authorisation scheme to be %s, got %s", scheme, header[0])
		}

		var sget session
		if err := jwt.DecodeEmbedded(client.serviceSecret, []byte(header[1]), &sget); err != nil {
			t.Fatal(err)
		}
		if sput.User != sget.User {
			t.Errorf("Expected user ID in session to be %s, got %s", sput.User, sget.User)
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

		req2, err := client.NewHeadlessRequest(req, "GET", server.URL, struct{}{}, nil)
		if err != nil {
			t.Fatal(err)
		}

		httpClient := server.Client()
		_, err = httpClient.Do(req2)
		if err == nil {
			t.Fatal("Expected request to fail with error")
		}

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected request to fail because deadline was exceeded, got %v", err)
		}
	})
}

func TestNewBaseRequest(t *testing.T) {
	t.Run("sets the right headers", func(t *testing.T) {
		type session struct{ User string }
		client := NewClient(Config{
			Secret:         []byte("secret"),
			Service:        "user-service",
			HeadlessScheme: "Test",
		})

		sput := session{uuid.New().String()}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		req, err := client.NewBaseRequest(ctx, "GET", "/internal", sput, nil)
		if err != nil {
			t.Fatal(err)
		}

		if req.Header.Get("X-Request-ID") == "" {
			t.Error("Expected request ID to be set")
		}

		if req.Header.Get("X-Origin-Service") != client.serviceName {
			t.Errorf("Expected origin service of request to be %s, got %s", client.serviceName, req.Header.Get("X-Origin-Service"))
		}

		header := strings.Fields(req.Header.Get("Authorization"))
		if header[0] != client.headlessScheme {
			t.Errorf("Expected authorisation scheme to be %s, got %s", client.headlessScheme, header[0])
		}

		var sget session
		if err := jwt.DecodeEmbedded(client.serviceSecret, []byte(header[1]), &sget); err != nil {
			t.Fatal(err)
		}
		if sput.User != sget.User {
			t.Errorf("Expected user ID in session to be %s, got %s", sput.User, sget.User)
		}
	})

	t.Run("times out when parent request times out", func(t *testing.T) {
		client := NewClient(Config{
			Secret:         []byte("secret"),
			Service:        faker.Company().Name(),
			HeadlessScheme: "scheme",
		})

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(time.Millisecond * 400)
			_, _ = w.Write([]byte("too late"))
		}))
		defer server.Close()

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
		defer cancel()

		req, err := client.NewBaseRequest(ctx, "GET", server.URL, struct{}{}, nil)
		if err != nil {
			t.Fatal(err)
		}

		httpClient := server.Client()
		_, err = httpClient.Do(req)
		if err == nil {
			t.Fatal("Expected request to fail with error")
		}

		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected request to fail because deadline was exceeded, got %v", err)
		}
	})
}
