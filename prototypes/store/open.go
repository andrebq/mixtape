package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
)

func OpenDB(file string) (*sqlx.DB, error) {
	err := os.MkdirAll(filepath.Dir(file), 0750)
	if err != nil {
		return nil, err
	}
	conn, err := sql.Open("sqlite3", fmt.Sprintf("file:%v", file))
	if err != nil {
		return nil, err
	}
	return sqlx.NewDb(conn, "sqlite"), nil
}
