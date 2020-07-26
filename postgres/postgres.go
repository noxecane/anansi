package postgres

import (
	"crypto/tls"
	"fmt"
	"strings"
	"sync"

	"github.com/go-pg/pg/v9"
)

// DB is the global database in case managing connection objects is
// too much work.
var DB *pg.DB
var pgOnce sync.Once

// PostgresEnv is the definition of environment variables needed
// to setup a postgres connection
type PostgresEnv struct {
	PostgresHost         string `required:"true" split_words:"true"`
	PostgresPort         int    `required:"true" split_words:"true"`
	PostgresSecureMode   bool   `required:"true" split_words:"true"`
	PostgresUser         string `required:"true" split_words:"true"`
	PostgresPassword     string `required:"true" split_words:"true"`
	PostgresDatabase     string `required:"true" split_words:"true"`
	PostgresMigrationDir string `split_words:"true"`
}

// ConnectDB initialises a global connection for `DB`
func ConnectDB(schema string, env PostgresEnv) {
	pgOnce.Do(func() {
		var err error
		DB, err = NewDB(schema, env)

		if err != nil {
			panic(err)
		}
	})
}

// NewDB creates a connection to a postgres DB and ensures the connection is live.
func NewDB(schema string, env PostgresEnv) (*pg.DB, error) {
	opts := &pg.Options{
		Addr:     fmt.Sprintf("%s:%d", env.PostgresHost, env.PostgresPort),
		User:     env.PostgresUser,
		Password: env.PostgresPassword,
		Database: env.PostgresDatabase,
	}

	var err error

	if schema != "" {
		opts.ApplicationName = schema
		opts.OnConnect = func(c *pg.Conn) error {
			if _, err2 := c.Exec("set search_path=?", schema); err2 != nil {
				err = err2
			}
			return nil
		}
	}

	if env.PostgresSecureMode {
		opts.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	db := pg.Connect(opts)
	_, err = db.Exec("select version()")

	return db, err
}

// CleanUpTables removes all the rows in the passed tables. It is
// useful for cleaning up the DB for tests.
func CleanUpTables(db *pg.DB, tables ...string) error {
	query := fmt.Sprintf("truncate %s cascade", strings.Join(tables, ","))
	if _, err := db.Exec(query); err != nil {
		return err
	}

	return nil
}
