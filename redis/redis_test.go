package redis

import (
	"testing"

	"github.com/go-redis/redis/v7"
)

func TestConnectToRedis(t *testing.T) {
	env := RedisEnv{
		RedisHost:     "fakehost",
		RedisPassword: "fakepassword",
		RedisPort:     3000,
	}
	opts := &redis.Options{
		DB: 0,
	}
	_, err := NewRedisClient(env, opts)

	if err == nil {
		t.Error("Passing wrong connection details doesn't result in an error")
	}
}
