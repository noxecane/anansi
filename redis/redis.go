package redis

import (
	"fmt"
	"sync"

	"github.com/go-redis/redis/v7"
)

// Client is the global redis client in case managing connection objects is
// too much work.
var Client *redis.Client
var redisOnce sync.Once

// RedisEnv is the definition of environment variables needed
// to connect to redis
type RedisEnv struct {
	RedisHost     string `required:"true" split_words:"true"`
	RedisPort     int    `required:"true" split_words:"true"`
	RedisPassword string `default:"" split_words:"true"`
}

// ConnectDB initialises a global connection for `Client`
func ConnectClient(env RedisEnv, opts *redis.Options) {
	redisOnce.Do(func() {
		var err error
		Client, err = NewRedisClient(env, opts)

		if err != nil {
			panic(err)
		}
	})
}

// NewRedisClient creates a client for redis and tests its connection
func NewRedisClient(env RedisEnv, opts *redis.Options) (*redis.Client, error) {
	opts.Addr = fmt.Sprintf("%s:%d", env.RedisHost, env.RedisPort)
	opts.Password = env.RedisPassword

	client := redis.NewClient(opts)

	// test the connection
	_, err := client.Ping().Result()

	return client, err
}
