package postgres

import (
	"fmt"
	"strings"

	"github.com/go-pg/pg/v9"
)

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
