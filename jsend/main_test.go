package jsend

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/random-guys/go-siber/sessions"
	"github.com/random-guys/go-siber/tokens"
)

var store *sessions.Store

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

	secret := []byte("monday-is-not-your-mate")
	store = sessions.NewStore(secret, "Test", time.Minute, tokens.NewStore(client, secret))

	defer os.Exit(m.Run())

	if err := client.Close(); err != nil {
		panic(err)
	}
}
