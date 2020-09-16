package siber

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

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
