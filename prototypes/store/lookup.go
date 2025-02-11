package store

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

func LookupOne[T any](ctx context.Context, tx sqlx.ExtContext, sample T) (T, error) {
	var zero T
	m, err := mappingFor(sample)
	if err != nil {
		return zero, err
	}
	rows, err := sqlx.NamedQueryContext(ctx, tx, m.lookup, sample)
	if err != nil {
		return zero, err
	}
	defer rows.Close()
	if !rows.Next() {
		return zero, sql.ErrNoRows
	}
	err = rows.StructScan(&zero)
	return zero, err
}
