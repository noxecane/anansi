package auth

import (
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
)

var sharedTestStore *TokenStore
var client *redis.Client

type SampleStruct struct {
	Message string `json:"message"`
}

func newRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// test the connection
	_, err := client.Ping().Result()

	return client, err
}

func flushRedis(t *testing.T) {
	if _, err := client.FlushDB().Result(); err != nil {
		t.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	var err error
	client, err = newRedisClient()
	if err != nil {
		panic(err)
	}

	sharedTestStore = &TokenStore{redis: client, secret: []byte("mykeys")}

	code := m.Run()

	os.Exit(code)
}

func TestCommissionDecomission(t *testing.T) {
	sample := SampleStruct{Message: "A sample message"}

	defer flushRedis(t)

	var token string
	var err error

	t.Run("the token is associated with the data", func(t *testing.T) {
		if token, _, err = sharedTestStore.Commission(time.Millisecond*300, sample); err != nil {
			t.Fatal(err)
		}

		if token == "" {
			t.Error("Expected a token to be defined, got an empty string")
		}

		result := new(SampleStruct)
		if err = sharedTestStore.Decommission(token, result); err != nil {
			t.Fatal(err)
		}

		if result.Message != sample.Message {
			t.Errorf("Expected message to be \"%s\", got \"%s\"", sample.Message, result.Message)
		}
	})

	t.Run("the token expires after timeout", func(t *testing.T) {
		if token, _, err = sharedTestStore.Commission(time.Millisecond*100, sample); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond * 100)

		if err = sharedTestStore.Decommission(token, &SampleStruct{}); err == nil && err == ErrTokenNotFound {
			t.Errorf("Expected error to be thrown")
		}
	})

	t.Run("decomission fails when random token is passed", func(t *testing.T) {
		if err = sharedTestStore.Decommission("token", &SampleStruct{}); err == nil && err == ErrTokenNotFound {
			t.Errorf("Expected error to be thrown")
		}
	})
}

func TestCommissionRefresh(t *testing.T) {
	sample := SampleStruct{Message: "A sample message"}

	defer flushRedis(t)

	var token string
	var err error

	t.Run("the token gets refreshed for an extra 300ms", func(t *testing.T) {
		if token, _, err = sharedTestStore.Commission(time.Millisecond*300, sample); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond * 100)

		temp := new(SampleStruct)
		if err = sharedTestStore.Refresh(token, time.Second, temp); err != nil {
			t.Fatal(err)
		}
		if temp.Message != sample.Message {
			t.Errorf("Expected message to be \"%s\", got %v", sample.Message, temp)
		}

		time.Sleep(time.Millisecond * 700)

		result := new(SampleStruct)
		if err = sharedTestStore.Decommission(token, result); err != nil {
			t.Fatal(err)
		}
		if result.Message != sample.Message {
			t.Errorf("Expected message to be \"%s\", got \"%s\"", sample.Message, result.Message)
		}
	})

	t.Run("the token expires after original time plus refresh timeout", func(t *testing.T) {
		if token, _, err = sharedTestStore.Commission(time.Millisecond*300, sample); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond * 100)

		if err = sharedTestStore.Refresh(token, time.Second, &SampleStruct{}); err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Second)

		if err = sharedTestStore.Decommission(token, &SampleStruct{}); err == nil && err == ErrTokenNotFound {
			t.Errorf("Expected error to be thrown")
		}
	})

	t.Run("it's impossible to refresh an expired token", func(t *testing.T) {
		if token, _, err = sharedTestStore.Commission(time.Millisecond*100, sample); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond * 100)

		if err = sharedTestStore.Refresh(token, time.Millisecond*100, &SampleStruct{}); err == nil && err == ErrTokenNotFound {
			t.Errorf("Expected error to be thrown")
		}
	})
}

func TestCommissionPeek(t *testing.T) {
	sample := SampleStruct{Message: "A sample message"}

	defer flushRedis(t)

	var token string
	var err error

	t.Run("the token doesn't expire till decommission is called", func(t *testing.T) {
		if token, _, err = sharedTestStore.Commission(time.Millisecond*300, sample); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond * 100)

		temp := new(SampleStruct)
		if err = sharedTestStore.Peek(token, temp); err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Millisecond * 100)

		result := new(SampleStruct)
		if err = sharedTestStore.Decommission(token, result); err != nil {
			t.Fatal(err)
		}
		if result.Message != temp.Message {
			t.Errorf("Expected  peek \"%s\" to be the same as decomission \"%s\"", temp.Message, result.Message)
		}
	})
}

func TestCommissionRevoke(t *testing.T) {
	sample := SampleStruct{Message: "A sample message"}

	defer flushRedis(t)

	var key string
	var token string
	var err error

	t.Run("the token is rendered useless immediately", func(t *testing.T) {
		if token, key, err = sharedTestStore.Commission(time.Millisecond*300, sample); err != nil {
			t.Fatal(err)
		}

		if err = sharedTestStore.Revoke(key); err != nil {
			t.Fatal(err)
		}

		if err = sharedTestStore.Peek(token, &SampleStruct{}); err == nil && err == ErrTokenNotFound {
			t.Errorf("Expected error to be thrown")
		}
	})
}
