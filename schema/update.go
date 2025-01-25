package schema

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/andrebq/mixtape/generics"
	"github.com/andrebq/mixtape/internal/validate"
	"github.com/jmoiron/sqlx"
)

type (
	TableName  string
	ColumnName string

	ColumnList []ColumnName
)

var (
	validIdentifer = regexp.MustCompile(`^[A-Z]+[A-Z0-9_]*$`)

	ErrInvalidName = errors.New("invalid identifier")
)

func (s *S) Merge(ctx context.Context, name TableName, columns ColumnList) error {
	name = name.Normalize()
	for i, v := range columns {
		columns[i] = v.Normalize()
	}
	if err := validate.All[validate.Valid](name, columns); err != nil {
		return err
	}
	targetName := fmt.Sprintf("t_%v", name)
	return s.writable(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		var found int
		err := tx.QueryRowxContext(ctx, `select 1 from sqlite_master where name = ?`, targetName).Scan(found)
		if errors.Is(err, sql.ErrNoRows) {
			return createTable(ctx, tx, targetName, columns)
		}
		return mergeTable(ctx, tx, targetName, columns)
	})
}

func createTable(ctx context.Context, tx *sqlx.Tx, table string, cols ColumnList) error {
	oid := ColumnName("oid")
	buf := strings.Builder{}
	colset := generics.SetOf(cols...)
	colset.PutAll(oid)
	fmt.Fprintf(&buf, "create table %v(\n", table)
	cols = colset.AppendTo(cols[:0])
	slices.Sort(cols)
	for _, c := range cols {
		fmt.Fprintf(&buf, "%v blob,\n", c)
	}
	fmt.Fprintf(&buf, "primary key (%v)", oid)
	fmt.Fprintf(&buf, ")")
	_, err := tx.ExecContext(ctx, buf.String())
	return err
}

func mergeTable(ctx context.Context, conn *sqlx.Tx, table string, newCols ColumnList) error {
	rows, err := conn.QueryxContext(ctx, fmt.Sprintf(`pragma table_info(%q)`, table))
	if err != nil {
		return err
	}
	defer rows.Close()
	type tbInfo struct {
		CID          string  `db:"cid"`
		Name         string  `db:"name"`
		Type         string  `db:"type"`
		NotNull      int     `db:"notnull"`
		DefaultValue *string `db:"dflt_value"`
		PK           int     `db:"pk"`
	}
	colset := generics.SetOf[ColumnName]()
	for rows.Next() {
		var info tbInfo
		err = rows.StructScan(&info)
		if err != nil {
			return err
		}
		colset.PutAll(ColumnName(info.Name))
	}
	rows.Close()
	for _, c := range newCols {
		if colset.Has(c) {
			continue
		}
		// TODO: there is probably a better way to batch all those changes
		_, err = conn.ExecContext(ctx, fmt.Sprintf(`alter table %v add column %v blob`, table, c))
		if err != nil {
			return fmt.Errorf("unable to add column %v to table: %w", c, err)
		}
	}
	return nil
}

func (t TableName) Valid() error {
	if !validIdentifer.MatchString(string(t)) {
		return fmt.Errorf("%v is a %v", t, ErrInvalidName)
	}
	return nil
}

func (t TableName) Normalize() TableName {
	return TableName(strings.ToUpper(string(t)))
}

func (c ColumnName) Valid() error {
	if !validIdentifer.MatchString(string(c)) {
		return fmt.Errorf("%v is a %v", c, ErrInvalidName)
	}
	return nil
}

func (c ColumnName) Normalize() ColumnName {
	return ColumnName(strings.ToUpper(string(c)))
}

func (cl ColumnList) Valid() error {
	return validate.All(cl...)
}
