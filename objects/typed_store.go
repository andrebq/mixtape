package objects

import (
	"context"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
)

func Put[T any](ctx context.Context, s Session, val T) (Ref, error) {
	if s.Err() != nil {
		return Ref{}, s.Err()
	}
	buf, err := msgpack.Marshal(val)
	if err != nil {
		return Ref{}, err
	}
	r, _ := s.Put(ctx, msgpack.RawMessage(buf))
	if s.Err() != nil {
		return Ref{}, fmt.Errorf("unable to update: %w", s.Err())
	}
	return r, nil
}

func Get[T any](ctx context.Context, out *T, s Session, ref Ref) error {
	if s.Err() != nil {
		return s.Err()
	}
	buf, ok := s.Get(ctx, ref)
	if s.Err() != nil {
		return fmt.Errorf("unable to get: %w", s.Err())
	} else if !ok {
		return ErrNotFound
	}
	return msgpack.Unmarshal(buf, out)
}
