package redis

import (
	"testing"
)

func TestConnectToRedis(t *testing.T) {
	_, err := NewRedisClient(RedisEnv{
		RedisHost:     "fakehost",
		RedisPassword: "fakepassword",
		RedisPort:     3000,
	})

	if err == nil {
		t.Error("Passing wrong connection details doesn't result in an error")
	}
}
