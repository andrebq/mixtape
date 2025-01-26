package schema

import (
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/andrebq/mixtape/generics"
	"github.com/andrebq/mixtape/internal/validate"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type (
	TableName  string
	ColumnName string

	ColumnList []ColumnName
)

var (
	validIdentifer = regexp.MustCompile(`^[A-Z]+[A-Z0-9_]*$`)

	ErrInvalidName    = errors.New("invalid identifier")
	ErrMissingOID     = errors.New("missing required field oid")
	ErrInvalidOIDType = errors.New("oid field must be a string or a uuid.UUID")

	oidCol = ColumnName("oid").Normalize()
)

func (s *S) Put(ctx context.Context, tupleType TableName, values map[string]any) (uuid.UUID, error) {
	cols := make(ColumnList, 0, len(values))
	cvals := make([]any, 0, len(values))
	var nextoid uuid.UUID
	for k, v := range values {
		cols = append(cols, ColumnName(k))
		cvals = append(cvals, v)
		if k == string(oidCol) {
			switch v := v.(type) {
			case string:
				var err error
				nextoid, err = uuid.Parse(v)
				if err != nil {
					return nextoid, fmt.Errorf("invalid oid: %w", err)
				}
			case uuid.UUID:
				nextoid = v
			default:
				return nextoid, ErrInvalidOIDType
			}
		}
	}
	err := s.Merge(ctx, tupleType, cols)
	if err != nil {
		return nextoid, err
	}
	colset := generics.SetOf(cols...)
	if !colset.Has(oidCol) {
		cols = append(cols, oidCol)
		var oidSeed [16]byte
		nval := atomic.AddUint64(&s.oidCount, 1)
		now := time.Now().UnixMicro()
		binary.BigEndian.AppendUint64(oidSeed[:0], nval)
		binary.BigEndian.AppendUint64(oidSeed[8:8], uint64(now))
		cvals = append(cvals, uuid.NewSHA1(s.randSeed, oidSeed[:]))
	}
	cmd := &strings.Builder{}
	fmt.Fprintf(cmd, `insert into t_%v(`, tupleType.Normalize())
	for i, c := range cols {
		if i != 0 {
			cmd.WriteString(",")
		}
		fmt.Fprintf(cmd, "%v", c.Normalize())
	}
	fmt.Fprintf(cmd, ") values (")
	for i := range cols {
		if i != 0 {
			cmd.WriteString(",")
		}
		cmd.WriteString("?")
	}
	cmd.WriteString(")")
	_, err = s.db.ExecContext(ctx, cmd.String(), cvals...)
	return nextoid, err
}

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
	buf := strings.Builder{}
	colset := generics.SetOf(cols...)
	colset.PutAll(oidCol)
	fmt.Fprintf(&buf, "create table %v(\n", table)
	cols = colset.AppendTo(cols[:0])
	slices.Sort(cols)
	for _, c := range cols {
		fmt.Fprintf(&buf, "%v blob,\n", c)
	}
	fmt.Fprintf(&buf, "primary key (%v)", oidCol)
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
