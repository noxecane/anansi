package sessions

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/noxecane/anansi/jwt"
	"github.com/noxecane/anansi/tokens"
	"github.com/redis/go-redis/v9"
)

var sharedTestStore tokens.Store
var client *redis.Client
var secret = []byte("ot4EvohHaeSeeshoo1eih7oow0FooWee")
var scheme = "Test"

func newRedisClient(ctx context.Context) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Username: "default",
		Password: "iP6tim0naosoohei",
		Addr:     "localhost:6379",
		DB:       0,
	})

	// test the connection
	_, err := client.Ping(ctx).Result()

	return client, err
}

func flushRedis(ctx context.Context, t *testing.T) {
	if _, err := client.FlushDB(ctx).Result(); err != nil {
		t.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	var err error
	client, err = newRedisClient(context.Background())
	if err != nil {
		panic(err)
	}

	sharedTestStore = tokens.NewStore(client, secret)

	defer os.Exit(m.Run())

	if err := client.Close(); err != nil {
		panic(err)
	}
}

func Test_getAuthorization(t *testing.T) {
	t.Run("fails when header format is incorrect", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/entities", nil)
		req.Header.Set("Authorization", "Bearer ")

		_, _, err := getAuthorization(req)
		if err == nil {
			t.Error("Expected LoadBearer to fail with error")
		}

		if err != ErrHeaderFormat {
			t.Errorf("Expected error from LoadBearer to be ErrHeaderFormat, got %s", err)
		}
	})

	t.Run("fails when header is not set", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/entities", nil)

		_, _, err := getAuthorization(req)
		if err == nil {
			t.Error("Expected LoadBearer to fail with error")
		}

		if err != ErrEmptyHeader {
			t.Errorf("Expected error from LoadBearer to be ErrEmptyHeader, got %s", err)
		}
	})
}

func TestFromAuth(t *testing.T) {
	type session struct {
		Name string
	}

	manager := NewManager(sharedTestStore, secret, Config{BearerDuration: time.Minute, HeadlessScheme: scheme})

	t.Run("loads bearer session from auth header", func(t *testing.T) {
		defer flushRedis(context.TODO(), t)

		token, err := sharedTestStore.Commission(context.TODO(), time.Minute, "key", session{"Premium"})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest("GET", "/entities", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		var s session
		if err := manager.FromAuth(req, &s); err != nil {
			t.Fatal(err)
		}

		if s.Name != "Premium" {
			t.Errorf(`Expected name in session to be "%s", got %s`, "Premium", s.Name)
		}
	})

	t.Run("loads headless session from auth header", func(t *testing.T) {
		token, err := jwt.Encode(secret, time.Minute, session{"Headless"})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest("GET", "/entities", nil)
		req.Header.Set("Authorization", scheme+" "+token)

		var s session
		if err := manager.FromAuth(req, &s); err != nil {
			t.Fatal(err)
		}

		if s.Name != "Headless" {
			t.Errorf(`Expected name in session to be "%s", got %s`, "Headless", s.Name)
		}
	})

	t.Run("fails if scheme is unsupported", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/entities", nil)
		req.Header.Set("Authorization", "Unknown engagement")

		err := manager.FromAuth(req, &session{})
		if err == nil {
			t.Error("Expected FromAuth to fail with error")
		}

		if err != ErrUnsupportedScheme {
			t.Errorf("Expected error from FromAuth to be ErrUnsupportedScheme, got %s", err)
		}
	})
}

func TestNewSessionAndFromCookie(t *testing.T) {
	type session struct {
		Name string
	}

	manager := NewManager(sharedTestStore, secret, Config{CookieDuration: time.Minute})

	t.Run("creates session and loads from cookie", func(t *testing.T) {
		defer flushRedis(context.TODO(), t)

		// Create session with unique ID
		token, err := manager.NewSession(context.TODO(), "user-123", session{"Premium"})
		if err != nil {
			t.Fatal(err)
		}

		// Write to cookie
		w := httptest.NewRecorder()
		manager.ToCookie(w, token, "/")

		// Extract cookie from response
		cookies := w.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatal("Expected cookie to be set")
		}

		// Create request with cookie
		req := httptest.NewRequest("GET", "/entities", nil)
		req.AddCookie(cookies[0])

		// Load from cookie
		var s session
		if err := manager.FromCookie(req, &s); err != nil {
			t.Fatal(err)
		}

		if s.Name != "Premium" {
			t.Errorf(`Expected name in session to be "%s", got %s`, "Premium", s.Name)
		}
	})

	t.Run("fails when no cookie is present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/entities", nil)

		err := manager.FromCookie(req, &session{})
		if err == nil {
			t.Error("Expected FromCookie to fail with error")
		}

		if err != ErrEmptyAuthCookie {
			t.Errorf("Expected error to be ErrEmptyAuthCookie, got %s", err)
		}
	})
}

func TestLogoutCookie(t *testing.T) {
	type session struct {
		Name string
	}

	manager := NewManager(sharedTestStore, secret, Config{CookieDuration: time.Minute})

	t.Run("revokes cookie token and clears cookie", func(t *testing.T) {
		defer flushRedis(context.TODO(), t)

		// Create a session first
		req := httptest.NewRequest("GET", "/entities", nil)
		w := httptest.NewRecorder()

		// Create session and write to cookie
		token, err := manager.NewSession(req.Context(), "user-123", session{"Premium"})
		if err != nil {
			t.Fatal(err)
		}
		manager.ToCookie(w, token, "/")

		// Extract the cookie from the response
		cookies := w.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatal("Expected cookie to be set")
		}
		cookie := cookies[0]

		// Create a new request with the cookie
		logoutReq := httptest.NewRequest("POST", "/logout", nil)
		logoutReq.AddCookie(cookie)
		logoutW := httptest.NewRecorder()

		// Logout
		err = manager.LogoutCookie(logoutReq, logoutW)
		if err != nil {
			t.Fatal(err)
		}

		// Verify cookie is cleared
		logoutCookies := logoutW.Result().Cookies()
		if len(logoutCookies) == 0 {
			t.Fatal("Expected logout cookie to be set")
		}

		logoutCookie := logoutCookies[0]
		if logoutCookie.Value != "" {
			t.Error("Expected cookie value to be empty after logout")
		}
		if logoutCookie.MaxAge != -1 {
			t.Error("Expected cookie MaxAge to be -1 after logout")
		}

		// Verify token is revoked by trying to load it
		loadReq := httptest.NewRequest("GET", "/test", nil)
		loadReq.AddCookie(cookie)

		var s session
		err = manager.FromCookie(loadReq, &s)
		if err == nil {
			t.Error("Expected FromCookie to fail after logout")
		}
	})

	t.Run("fails when no cookie is present", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/logout", nil)
		w := httptest.NewRecorder()

		err := manager.LogoutCookie(req, w)
		if err == nil {
			t.Error("Expected LogoutCookie to fail with error")
		}

		if err != ErrEmptyAuthCookie {
			t.Errorf("Expected error to be ErrEmptyAuthCookie, got %s", err)
		}
	})
}

func TestLogoutAuth(t *testing.T) {
	type session struct {
		Name string
	}

	manager := NewManager(sharedTestStore, secret, Config{BearerDuration: time.Minute, HeadlessScheme: scheme})

	t.Run("revokes bearer token", func(t *testing.T) {
		defer flushRedis(context.TODO(), t)

		// Create a bearer token
		token, err := manager.NewSession(context.TODO(), "test-key", session{"Premium"})
		if err != nil {
			t.Fatal(err)
		}

		// Create request with bearer token
		logoutReq := httptest.NewRequest("POST", "/logout", nil)
		logoutReq.Header.Set("Authorization", "Bearer "+token)

		// Logout
		err = manager.LogoutAuth(logoutReq)
		if err != nil {
			t.Fatal(err)
		}

		// Verify token is revoked by trying to load it
		loadReq := httptest.NewRequest("GET", "/test", nil)
		loadReq.Header.Set("Authorization", "Bearer "+token)

		var s session
		err = manager.FromAuth(loadReq, &s)
		if err == nil {
			t.Error("Expected FromAuth to fail after logout")
		}
	})

	t.Run("does not fail for headless tokens (stateless)", func(t *testing.T) {
		token, err := manager.NewHeadlessSession(session{"Headless"})
		if err != nil {
			t.Fatal(err)
		}

		// Create request with headless token
		logoutReq := httptest.NewRequest("POST", "/logout", nil)
		logoutReq.Header.Set("Authorization", scheme+" "+token)

		// Logout should succeed (no-op for headless)
		err = manager.LogoutAuth(logoutReq)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("fails when no authorization header is present", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/logout", nil)

		err := manager.LogoutAuth(req)
		if err == nil {
			t.Error("Expected LogoutAuth to fail with error")
		}

		if err != ErrEmptyHeader {
			t.Errorf("Expected error to be ErrEmptyHeader, got %s", err)
		}
	})

	t.Run("fails when authorization header format is incorrect", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/logout", nil)
		req.Header.Set("Authorization", "Bearer")

		err := manager.LogoutAuth(req)
		if err == nil {
			t.Error("Expected LogoutAuth to fail with error")
		}

		if err != ErrHeaderFormat {
			t.Errorf("Expected error to be ErrHeaderFormat, got %s", err)
		}
	})
}
