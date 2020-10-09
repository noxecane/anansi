package postgres

import (
	"fmt"
	"strings"

	"github.com/go-pg/pg/v10"
)

// CleanUpTables removes all the rows in the passed tables. It is useful
// for cleaning up the DB for tests.
func CleanUpTables(db *pg.DB, tables ...string) error {
	query := fmt.Sprintf("truncate %s cascade", strings.Join(tables, ","))
	if _, err := db.Exec(query); err != nil {
		return err
	}

	return nil
}
