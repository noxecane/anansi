package postgres

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migrate runs the SQL migration files in dir using go-migrate.
// This means you have to ensure the SQL files follow the go-migrate format
// Also note that the directory must be relative to final executable without
// the preceding "/" or "./".
func Migrate(dir string, env PostgresEnv) error {
	// interprete the secure mode flag
	var sslMode string
	if env.PostgresSecureMode {
		sslMode = "require"
	} else {
		sslMode = "disable"
	}

	postgresURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		env.PostgresUser,
		env.PostgresPassword,
		env.PostgresHost,
		env.PostgresPort,
		env.PostgresDatabase,
		sslMode,
	)

	var mig *migrate.Migrate
	var err error

	dir = fmt.Sprintf("file://%s", dir)
	if mig, err = migrate.New(dir, postgresURL); err != nil {
		return err
	}

	if err = mig.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
