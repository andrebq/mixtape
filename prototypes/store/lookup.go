package store

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

func Match[T any, E ~[]T](ctx context.Context, out E, tx sqlx.ExtContext, pattern map[string]any) (E, error) {
	var zero T
	m, err := mappingFor(zero)
	if err != nil {
		return out, err
	}
	stmt, pattern, err := m.match(pattern)
	if err != nil {
		return out, err
	}
	println("stmt: ", stmt)
	rows, err := sqlx.NamedQueryContext(ctx, tx, stmt, pattern)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var item T
		err = rows.StructScan(&item)
		if err != nil {
			return out, err
		}
		out = append(out, item)
	}
	return out, nil
}

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
