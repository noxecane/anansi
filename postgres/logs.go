package postgres

import (
	"context"
	"fmt"

	"github.com/go-pg/pg/v9"
)

type queryLogger struct {
}

func NewQueryLogger() queryLogger {
	return queryLogger{}
}

func (_ queryLogger) BeforeQuery(ctx context.Context, q *pg.QueryEvent) (context.Context, error) {
	return ctx, nil
}

func (log queryLogger) AfterQuery(ctx context.Context, q *pg.QueryEvent) error {
	if query, err := q.FormattedQuery(); err != nil {
		return err
	} else {
		fmt.Println("postgres_query", query)
	}

	return nil
}
