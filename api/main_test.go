package api

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/noxecane/anansi/sessions"
	"github.com/noxecane/anansi/tokens"
	"github.com/redis/go-redis/v9"
)

var store *sessions.Manager

const (
	secret = "ot4EvohHaeSeeshoo1eih7oow0FooWee"
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

	store = sessions.NewManager(
		tokens.NewStore(client, []byte(secret)), []byte(secret), sessions.Config{
			BearerDuration: time.Minute,
		},
	)

	defer os.Exit(m.Run())

	if err := client.Close(); err != nil {
		panic(err)
	}
}
