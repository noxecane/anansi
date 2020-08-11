package postgres

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migrate runs the SQL migration files in dir using go-migrate.
// This means you have to ensure the SQL files follow the go-migrate format
// Also note that the directory must be an absolute path
func Migrate(dir, dbName, schema string, postgresURL string) error {

	var mig *migrate.Migrate
	var db *sql.DB
	var driver database.Driver
	var err error

	dir = fmt.Sprintf("file:///%s", dir)
	if db, err = sql.Open("postgres", postgresURL); err != nil {
		return err
	}
	defer db.Close()

	conf := postgres.Config{
		DatabaseName: dbName,
		SchemaName:   schema,
	}
	if driver, err = postgres.WithInstance(db, &conf); err != nil {
		return err
	}

	if mig, err = migrate.NewWithDatabaseInstance(dir, "postgres", driver); err != nil {
		return err
	}
	defer mig.Close()

	if err = mig.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
