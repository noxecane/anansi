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

func (_ queryLogger) BeforeQuery(ctx context.Context, q *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}

func (log queryLogger) AfterQuery(ctx context.Context, q *pg.QueryEvent) error {
	if query, err := q.FormattedQuery(); err != nil {
		return err
	} else {
		log.Debug().Str("postgres_query", query).Msg("")
	}

	return nil
}
