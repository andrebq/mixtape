package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type (
	ErrNotMapped struct{ reflect.Type }
)

func (e ErrNotMapped) Error() string {
	return fmt.Sprintf("type %v is not known to store registry, did you call MustRegister", e.Type)
}

func tableExists(ctx context.Context, db sqlx.ExtContext, tableName string) (bool, error) {
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?;"
	row := db.QueryRowxContext(ctx, query, tableName)
	var name string
	err := row.Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	return name != "", nil
}

func getExistingColumns(ctx context.Context, db sqlx.ExtContext, tableName string) (map[string]string, error) {
	columns := make(map[string]string)
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s);", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cid int
	var name, ctype string
	var notnull, pk int
	var dfltValue sql.NullString
	for rows.Next() {
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return nil, err
		}
		columns[name] = ctype
	}
	return columns, nil
}

func Migrate(ctx context.Context, db sqlx.ExtContext, types ...reflect.Type) error {
	for _, typeInfo := range types {
		mapping, _ := registry.Get(typeInfo)
		if mapping == nil {
			return ErrNotMapped{typeInfo}
		}
		tableName := mapping.tableName
		existingColumns, err := getExistingColumns(ctx, db, tableName)
		if err != nil {
			return err
		}

		tableExists, err := tableExists(ctx, db, tableName)
		if err != nil {
			return err
		}

		if !tableExists {
			_, err := db.ExecContext(ctx, mapping.createStatement)
			if err != nil {
				return err
			}
		} else {
			for k, stmt := range mapping.alterStatements {
				if _, found := existingColumns[k]; found {
					continue
				}
				_, err := db.ExecContext(ctx, stmt)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
