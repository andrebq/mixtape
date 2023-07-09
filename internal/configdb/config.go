package configdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type (
	DBLike interface {
		ExecContext(context.Context, string, ...any) (sql.Result, error)
		QueryRowContext(context.Context, string, ...any) *sql.Row
	}

	Locator struct {
		tableName string
		putCmd    string
		getCmd    string
	}

	Instance struct {
		Locator
		Conn DBLike

		KeyPrefix string

		err error
	}
)

// Migrate the config table under tablename
func Migrate(ctx context.Context, conn DBLike, tablename string) (Locator, error) {
	cmds := []string{
		fmt.Sprintf(`create table if not exists %v(id integer primary key, key text not null, value text not null, last_update text not null, revisions integer);`, tablename),
		fmt.Sprintf(`create unique index unq_key on %v(key)`, tablename),
	}
	for _, c := range cmds {
		_, err := conn.ExecContext(ctx, c)
		if err != nil {
			return Locator{}, err
		}
	}
	return Locator{tableName: tablename,
		putCmd: genPutCmd(tablename),
		getCmd: genGetCmd(tablename),
	}, nil
}

func genPutCmd(table string) string {
	return fmt.Sprintf("insert into %v (key, value, last_update, revisions) values ($1, $2, $3, 1) on conflict (key) do update set value=EXCLUDED.value, last_update=EXCLUDED.value, revisions = revisions + 1", table)
}

func genGetCmd(table string) string {
	return fmt.Sprintf("select value from %v where key = $1", table)
}

// Put a value in the config
func Put(ctx context.Context, conn DBLike, loc Locator, key, value string) error {
	_, err := conn.ExecContext(ctx, loc.putCmd, key, value, time.Now().Format(time.RFC3339), 1)
	if err != nil {
		return fmt.Errorf("confidb: error while updating value for key \"%v\", caused by %w", key, err)
	}
	return err
}

// Get a value from the config, returns true to help users differentiate from a missing config
// to a config with an empty value.
//
// If a key does not exists, returns the zero value for a string
func Get(ctx context.Context, conn DBLike, loc Locator, key string) (string, bool, error) {
	var out string
	err := conn.QueryRowContext(ctx, loc.getCmd, key).Scan(&out)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
		return "", false, nil
	} else if err != nil {
		return "", false, err
	}
	return out, true, nil
}

func (i *Instance) GetString(ctx context.Context, out *string, key string, fallback string) bool {
	if i.err != nil {
		return false
	}
	var found bool
	*out, found, i.err = Get(ctx, i.Conn, i.Locator, fmt.Sprintf("%v%v", i.KeyPrefix, key))
	return found && i.err == nil
}

func (i *Instance) Err() error { return i.err }
