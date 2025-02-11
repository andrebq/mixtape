package store

import (
	"context"
	"reflect"

	"github.com/jmoiron/sqlx"
)

func rtype(val any) reflect.Type {
	rt := reflect.TypeOf(val)
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	return rt
}

func mappingFor(val any) (*mappingData, error) {
	tp := rtype(val)
	mapping, _ := registry.Get(tp)
	if mapping == nil {
		return nil, ErrNotMapped{tp}
	}
	return mapping, nil
}

// Upsert val on tx using the type registry.
//
// Callers are responsible for creating the required objects in tx,
// see MustRegister and Migrate for more information.
func Upsert(ctx context.Context, tx sqlx.ExtContext, val interface{}) error {
	m, err := mappingFor(val)
	if err != nil {
		return err
	}
	_, err = sqlx.NamedExecContext(ctx, tx, m.insert, val)
	return err
}
