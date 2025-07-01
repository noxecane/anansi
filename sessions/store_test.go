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

func TestLoadBearer(t *testing.T) {
	type session struct {
		Name string
	}

	manager := NewManager(sharedTestStore, secret, Config{BearerDuration: time.Minute})

	t.Run("loads the bearer session", func(t *testing.T) {
		defer flushRedis(context.TODO(), t)

		token, err := sharedTestStore.Commission(context.TODO(), time.Minute, "key", session{"Premium"})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest("GET", "/entities", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		var s session
		if err := manager.LoadBearer(req, &s); err != nil {
			t.Fatal(err)
		}

		if s.Name != "Premium" {
			t.Errorf(`Expected name in session to be "%s", got %s`, "Premium", s.Name)
		}
	})

	t.Run("fails if scheme is not Bearer", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/entities", nil)
		req.Header.Set("Authorization", scheme+" engagement")

		err := manager.LoadBearer(req, &session{})
		if err == nil {
			t.Error("Expected LoadBearer to fail with error")
		}

		if err != ErrUnsupportedScheme {
			t.Errorf("Expected error from LoadBearer to be ErrUnsupportedScheme, got %s", err)
		}
	})
}

func TestLoadHeadless(t *testing.T) {
	type session struct {
		Name string
	}

	customScheme := "Premium"
	manager := NewManager(sharedTestStore, secret, Config{HeadlessScheme: customScheme})

	t.Run("loads the headless session", func(t *testing.T) {
		token, err := jwt.Encode(secret, time.Minute, session{"Premium"})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest("GET", "/entities", nil)
		req.Header.Set("Authorization", customScheme+" "+token)

		var s session
		if err := manager.LoadHeadless(req, &s); err != nil {
			t.Fatal(err)
		}

		if s.Name != "Premium" {
			t.Errorf(`Expected name in session to be "%s", got %s`, "Premium", s.Name)
		}
	})

	t.Run("fails if scheme is not set headless scheme", func(t *testing.T) {
		token := "engagement"
		req := httptest.NewRequest("GET", "/entities", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		err := manager.LoadHeadless(req, &session{})
		if err == nil {
			t.Error("Expected LoadHeadless to fail with error")
		}

		if err != ErrUnsupportedScheme {
			t.Errorf("Expected error from LoadHeadless to be ErrUnsupportedScheme, got %s", err)
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

		err := manager.NewCookieSession(req, w, session{"Premium"})
		if err != nil {
			t.Fatal(err)
		}

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
		err = manager.LoadCookie(loadReq, &s)
		if err == nil {
			t.Error("Expected LoadCookie to fail after logout")
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

func TestLogoutBearer(t *testing.T) {
	type session struct {
		Name string
	}

	manager := NewManager(sharedTestStore, secret, Config{BearerDuration: time.Minute})

	t.Run("revokes bearer token", func(t *testing.T) {
		defer flushRedis(context.TODO(), t)

		// Create a bearer token
		req := httptest.NewRequest("GET", "/entities", nil)
		token, err := manager.NewBearerToken(req, "test-key", session{"Premium"})
		if err != nil {
			t.Fatal(err)
		}

		// Create request with bearer token
		logoutReq := httptest.NewRequest("POST", "/logout", nil)
		logoutReq.Header.Set("Authorization", "Bearer "+token)

		// Logout
		err = manager.LogoutBearer(logoutReq)
		if err != nil {
			t.Fatal(err)
		}

		// Verify token is revoked by trying to load it
		loadReq := httptest.NewRequest("GET", "/test", nil)
		loadReq.Header.Set("Authorization", "Bearer "+token)

		var s session
		err = manager.LoadBearer(loadReq, &s)
		if err == nil {
			t.Error("Expected LoadBearer to fail after logout")
		}
	})

	t.Run("fails when no authorization header is present", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/logout", nil)

		err := manager.LogoutBearer(req)
		if err == nil {
			t.Error("Expected LogoutBearer to fail with error")
		}

		if err != ErrEmptyHeader {
			t.Errorf("Expected error to be ErrEmptyHeader, got %s", err)
		}
	})

	t.Run("fails when authorization header format is incorrect", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/logout", nil)
		req.Header.Set("Authorization", "Bearer")

		err := manager.LogoutBearer(req)
		if err == nil {
			t.Error("Expected LogoutBearer to fail with error")
		}

		if err != ErrHeaderFormat {
			t.Errorf("Expected error to be ErrHeaderFormat, got %s", err)
		}
	})
}
