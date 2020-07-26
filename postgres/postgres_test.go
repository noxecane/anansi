package postgres

import (
	"testing"

	"github.com/go-pg/pg/v9"
)

func TestConnectToPostgres(t *testing.T) {
	env := PostgresEnv{
		PostgresDatabase:   "fakedb",
		PostgresHost:       "fakehost",
		PostgresPassword:   "fakepassword",
		PostgresPort:       3000,
		PostgresSecureMode: false,
		PostgresUser:       "fakeuser",
	}
	_, err := NewDB("", env, &pg.Options{})

	if err == nil {
		t.Error("Passing wrong connection details doesn't result in an error")
	}
}
