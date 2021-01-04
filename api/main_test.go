package api

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/tsaron/anansi/sessions"
	"github.com/tsaron/anansi/tokens"
)

var store *sessions.Store

const (
	secret = "monday-is-not-your-mate"
	scheme = "Test"
)

func TestMain(m *testing.M) {
	var err error

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// test the connection
	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	store = sessions.NewStore([]byte(secret), scheme, time.Minute, tokens.NewStore(client, []byte(secret)))

	defer os.Exit(m.Run())

	if err := client.Close(); err != nil {
		panic(err)
	}
}
