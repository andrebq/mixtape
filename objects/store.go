package objects

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vmihailenco/msgpack/v5"
)

type (
	Ref struct {
		Kind string `msgpack:"_kind"`
		ID   OID    `msgpack:"_id"`
	}

	OID [16]byte

	Session interface {
		io.Closer
		Put(ctx context.Context, obj msgpack.RawMessage) (Ref, bool)
		Get(ctx context.Context, ref Ref) (msgpack.RawMessage, bool)
		//Tag(ctx context.Context, target Ref, tags map[string]string)

		Err() error
		Commit() error
	}

	Storage interface {
		io.Closer
		Session(ctx context.Context) Session
	}

	sqlStore struct {
		db *sql.DB

		opcount uint64
		seed    uuid.UUID
	}

	sqlSession struct {
		tx    *sql.Tx
		err   error
		store *sqlStore
	}
)

var (
	ErrClosed      = errors.New("already closed")
	ErrFailed      = errors.New("session has an error, cannot commit")
	ErrMissingKind = errors.New("missing kind property")
	ErrNotFound    = errors.New("not found")
)

func (o *OID) Scan(val any) error {
	if val == nil {
		return nil
	}
	switch val := val.(type) {
	case string:
		id, err := uuid.Parse(val)
		if err != nil {
			return err
		}
		*o = [16]byte(id)
	case []byte:
		if len(val) == 16 {
			copy((*o)[:], val)
		}
		id, err := uuid.ParseBytes(val)
		if err != nil {
			return err
		}
		*o = [16]byte(id)
	default:
		return fmt.Errorf("cannot cast %T to OID", val)
	}
	return nil
}

func (o OID) Value() (driver.Value, error) {
	return o[:], nil
}

func MemoryStorage() (Storage, error) {
	conn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}
	err = initDB(context.Background(), conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	oidseed, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	return &sqlStore{
		db:   conn,
		seed: oidseed,
	}, nil
}

func initDB(ctx context.Context, conn *sql.DB) error {
	_, err := conn.ExecContext(ctx, `create table if not exists t_objects(_id blob, _kind text, content blob, primary key(_kind, _id))`)
	if err != nil {
		return err
	}
	return nil
}

func (s *sqlStore) Close() error {
	return s.db.Close()
}

func (s *sqlStore) Session(ctx context.Context) Session {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
	return &sqlSession{
		tx:    tx,
		err:   err,
		store: s,
	}
}

func (s *sqlStore) newOID() OID {
	val := atomic.AddUint64(&s.opcount, 1)
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], val)
	return OID(uuid.NewSHA1(s.seed, buf[:]))
}

func (s *sqlSession) Close() error {
	if s.tx == nil {
		// already closed
		return s.err
	}
	s.err = s.tx.Rollback()
	return s.err
}

func (s *sqlSession) Commit() error {
	if s.tx == nil {
		return ErrClosed
	} else if s.err != nil {
		return ErrFailed
	}
	s.err = s.tx.Commit()
	s.tx = nil
	return s.err
}

func (s *sqlSession) Err() error { return s.err }

func (s *sqlSession) Get(ctx context.Context, ref Ref) (msgpack.RawMessage, bool) {
	if s.err != nil {
		return nil, false
	}
	var buf msgpack.RawMessage
	err := s.tx.QueryRowContext(ctx, `select content from t_objects where _kind = ? and _id = ?`, ref.Kind, ref.ID[:]).Scan(&buf)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false
	} else if err != nil {
		s.err = err
		return nil, false
	}
	return buf, true
}

func (s *sqlSession) Put(ctx context.Context, obj msgpack.RawMessage) (Ref, bool) {
	if s.err != nil {
		return Ref{}, false
	}
	var ref Ref
	s.err = msgpack.Unmarshal(obj, &ref)
	if s.err != nil {
		return Ref{}, false
	}
	if ref.Kind == "" {
		s.err = ErrMissingKind
	}
	if ref.ID.IsZero() {
		// slow path, need to decode the whole object
		// and insert a new
		var out map[any]any
		msgpack.Unmarshal(obj, &out)
		ref.ID = s.store.newOID()
		out["_id"] = ref.ID
		obj, s.err = msgpack.Marshal(out)
	}
	_, s.err = s.tx.ExecContext(ctx, `insert into t_objects (_kind, _id, content) values (?, ?, ?)`, ref.Kind, ref.ID, obj)
	return ref, s.err == nil
}

func (o *OID) IsZero() bool {
	return *o == OID{}
}
