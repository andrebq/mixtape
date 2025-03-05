package store

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type (
	TypedJSON[T any] struct {
		V T
	}
)

func AsJSON[T any](v T) TypedJSON[T] {
	return TypedJSON[T]{v}
}

func (j *TypedJSON[T]) Value() (driver.Value, error) {
	return json.Marshal(j.V)
}

func (j *TypedJSON[T]) Scan(v any) error {
	if j == nil {
		return errors.New("cannot scan into nil JSONStr")
	}
	var input io.Reader
	switch v := v.(type) {
	case string:
		input = strings.NewReader(v)
	case []byte:
		input = bytes.NewReader(v)
	default:
		return fmt.Errorf("cannot cast from %T to JSONStr", v)
	}
	return json.NewDecoder(input).Decode(&j.V)
}
