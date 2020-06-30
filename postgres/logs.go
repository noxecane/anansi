package postgres

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/rs/zerolog"
)

type queryLogger struct {
	zerolog.Logger
}

func NewQueryLogger(log zerolog.Logger) queryLogger {
	return queryLogger{Logger: log.With().Logger()}
}

func (dbLog queryLogger) BeforeQuery(ctx context.Context, q *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}

func (dbLog queryLogger) AfterQuery(ctx context.Context, q *pg.QueryEvent) error {
	query, err := q.FormattedQuery()
	if err != nil {
		return err
	}

	dbLog.Debug().Str("postgres_query", query).Msg("")
	return nil
}
