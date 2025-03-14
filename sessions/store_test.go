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
		Addr: "localhost:6379",
		DB:   0,
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

	manager := NewManager(sharedTestStore, secret, Config{HeadlessScheme: scheme})

	t.Run("loads the headless session", func(t *testing.T) {
		token, err := jwt.Encode(secret, time.Minute, session{"Premium"})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest("GET", "/entities", nil)
		req.Header.Set("Authorization", scheme+" "+token)

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
