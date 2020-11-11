package tokens

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"syreclabs.com/go/faker"
)

var sharedTestStore *Store
var client *redis.Client
var ctx = context.TODO()

type SampleStruct struct {
	Message string `json:"message"`
	User    string `json:"user_id"`
}

func newRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// test the connection
	_, err := client.Ping(ctx).Result()

	return client, err
}

func flushRedis(t *testing.T) {
	if _, err := client.FlushDB(ctx).Result(); err != nil {
		t.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	var err error
	client, err = newRedisClient()
	if err != nil {
		panic(err)
	}

	sharedTestStore = &Store{redis: client, secret: []byte("mykeys")}

	code := m.Run()

	os.Exit(code)
}

func TestCommissionDecomission(t *testing.T) {
	sample := SampleStruct{Message: faker.Lorem().Sentence(5), User: faker.Lorem().Word()}

	defer flushRedis(t)

	var token string
	var err error

	t.Run("the token is associated with the data", func(t *testing.T) {
		if token, err = sharedTestStore.Commission(ctx, time.Second, sample.User, sample); err != nil {
			t.Fatal(err)
		}

		if token == "" {
			t.Error("Expected a token to be defined, got an empty string")
		}

		result := new(SampleStruct)
		if err = sharedTestStore.Decommission(ctx, token, result); err != nil {
			t.Fatal(err)
		}

		if result.Message != sample.Message {
			t.Errorf("Expected message to be \"%s\", got \"%s\"", sample.Message, result.Message)
		}
	})

	t.Run("the token expires after timeout", func(t *testing.T) {
		if token, err = sharedTestStore.Commission(ctx, time.Second, sample.User, sample); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Second)

		if err = sharedTestStore.Decommission(ctx, token, &SampleStruct{}); err == nil && err == ErrTokenNotFound {
			t.Errorf("Expected error to be thrown")
		}
	})

	t.Run("decomission fails when random token is passed", func(t *testing.T) {
		if err = sharedTestStore.Decommission(ctx, "token", &SampleStruct{}); err == nil && err == ErrTokenNotFound {
			t.Errorf("Expected error to be thrown")
		}
	})
}

func TestCommissionExtend(t *testing.T) {
	sample := SampleStruct{Message: "A sample message"}

	defer flushRedis(t)

	var token string
	var err error

	t.Run("the token gets refreshed for an extra 300ms", func(t *testing.T) {
		if token, err = sharedTestStore.Commission(ctx, time.Second, sample.User, sample); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond * 800)

		if err = sharedTestStore.Extend(ctx, token, time.Second, &SampleStruct{}); err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Millisecond * 800)

		result := new(SampleStruct)
		if err = sharedTestStore.Decommission(ctx, token, result); err != nil {
			t.Fatal(err)
		}
		if result.Message != sample.Message {
			t.Errorf("Expected message to be \"%s\", got \"%s\"", sample.Message, result.Message)
		}
	})

	t.Run("the token expires after original time plus refresh timeout", func(t *testing.T) {
		if token, err = sharedTestStore.Commission(ctx, time.Second, sample.User, sample); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond * 100)

		if err = sharedTestStore.Extend(ctx, token, time.Second, &SampleStruct{}); err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Second)

		if err = sharedTestStore.Decommission(ctx, token, &SampleStruct{}); err == nil && err == ErrTokenNotFound {
			t.Errorf("Expected error to be thrown")
		}
	})

	t.Run("it's impossible to refresh an expired token", func(t *testing.T) {
		if token, err = sharedTestStore.Commission(ctx, time.Second, sample.User, sample); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Second)

		if err = sharedTestStore.Extend(ctx, token, time.Second, &SampleStruct{}); err == nil && err == ErrTokenNotFound {
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
		if token, err = sharedTestStore.Commission(ctx, time.Second, sample.User, sample); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond * 300)

		temp := new(SampleStruct)
		if err = sharedTestStore.Peek(ctx, token, temp); err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Millisecond * 300)

		result := new(SampleStruct)
		if err = sharedTestStore.Decommission(ctx, token, result); err != nil {
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

	var token string
	var err error

	t.Run("the token is rendered useless immediately", func(t *testing.T) {
		if token, err = sharedTestStore.Commission(ctx, time.Second, sample.User, sample); err != nil {
			t.Fatal(err)
		}

		if err = sharedTestStore.Revoke(ctx, sample.User); err != nil {
			t.Fatal(err)
		}

		if err = sharedTestStore.Peek(ctx, token, &SampleStruct{}); err == nil && err == ErrTokenNotFound {
			t.Errorf("Expected error to be thrown")
		}
	})
}

func TestCommissionReset(t *testing.T) {
	sample := SampleStruct{Message: "A sample message"}

	defer flushRedis(t)

	var token string
	var err error

	t.Run("the token gets refreshed for an extra 300ms", func(t *testing.T) {
		if token, err = sharedTestStore.Commission(ctx, time.Second, sample.User, sample); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond * 800)

		if err = sharedTestStore.Extend(ctx, token, time.Second, &SampleStruct{}); err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Millisecond * 800)

		result := new(SampleStruct)
		if err = sharedTestStore.Decommission(ctx, token, result); err != nil {
			t.Fatal(err)
		}
		if result.Message != sample.Message {
			t.Errorf("Expected message to be \"%s\", got \"%s\"", sample.Message, result.Message)
		}
	})

	t.Run("the token expires after original time plus refresh timeout", func(t *testing.T) {
		if token, err = sharedTestStore.Commission(ctx, time.Second, sample.User, sample); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond * 100)

		if err = sharedTestStore.Extend(ctx, token, time.Second, &SampleStruct{}); err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Second)

		if err = sharedTestStore.Decommission(ctx, token, &SampleStruct{}); err == nil && err == ErrTokenNotFound {
			t.Errorf("Expected error to be thrown")
		}
	})

	t.Run("it's impossible to refresh an expired token", func(t *testing.T) {
		if token, err = sharedTestStore.Commission(ctx, time.Second, sample.User, sample); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Second)

		if err = sharedTestStore.Extend(ctx, token, time.Second, &SampleStruct{}); err == nil && err == ErrTokenNotFound {
			t.Errorf("Expected error to be thrown")
		}
	})
}
