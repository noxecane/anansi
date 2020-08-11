package postgres

import (
	"fmt"
	"strings"

	"github.com/go-pg/pg/v9"
)

// PostgresEnv is the definition of environment variables needed
// to setup a postgres connection
type PostgresEnv struct {
	PostgresHost       string `required:"true" split_words:"true"`
	PostgresPort       int    `required:"true" split_words:"true"`
	PostgresSecureMode bool   `required:"true" split_words:"true"`
	PostgresUser       string `required:"true" split_words:"true"`
	PostgresPassword   string `required:"true" split_words:"true"`
	PostgresDatabase   string `required:"true" split_words:"true"`
}

func SetSchema(schema string, opts *pg.Options) {
	opts.ApplicationName = schema
	opts.OnConnect = func(c *pg.Conn) error {
		_, err := c.Exec("set search_path=?", schema)
		return err
	}
}

// CleanUpTables removes all the rows in the passed tables. It is useful
// for cleaning up the DB for tests.
func CleanUpTables(db *pg.DB, tables ...string) error {
	query := fmt.Sprintf("truncate %s cascade", strings.Join(tables, ","))
	if _, err := db.Exec(query); err != nil {
		return err
	}

	return nil
}
