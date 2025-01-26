package schema

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type (
	S struct {
		db *sqlx.DB

		randSeed uuid.UUID

		oidCount uint64
	}
)

func Open(dest string) (*S, error) {
	dbpath := filepath.Join(dest, "index.db")
	dsn := fmt.Sprintf("file:%v?cache=shared", dbpath)
	seed, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	conn, err := sqlx.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	return &S{db: conn,
		randSeed: seed,
	}, nil
}

func (s *S) Close() error {
	return s.db.Close()
}

func (s *S) newTx(ctx context.Context) (*sqlx.Tx, func(error) error, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	return tx, func(err error) error {
		if err != nil {
			return tx.Rollback()
		} else {
			return tx.Commit()
		}
	}, nil
}

func (s *S) writable(ctx context.Context, fn func(context.Context, *sqlx.Tx) error) error {
	tx, done, err := s.newTx(ctx)
	if err != nil {
		return err
	}
	var fnErr error
	defer func() {
		fnErr = done(fnErr)
	}()
	fnErr = fn(ctx, tx)
	return fnErr
}
