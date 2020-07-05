package anansi

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type BasicEnv struct {
	AppEnv string `default:"dev" split_words:"true"`
	Name   string `required:"true"`
	Port   int    `required:"true"`
	Scheme string `required:"true"`
	Secret []byte `required:"true"`
}

// LoadEnv loads environment variables into Env
func LoadEnv(env interface{}) error {
	// try to load from .env first
	err := godotenv.Load()
	if err != nil {
		perr, ok := err.(*os.PathError)
		if !ok || !errors.Is(perr.Unwrap(), os.ErrNotExist) {
			return err
		}
	}

	return envconfig.Process("", env)
}
