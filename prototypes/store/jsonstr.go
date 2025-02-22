package store

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

type (
	JSONBlob json.RawMessage
)

var (
	errNotJSON = errors.New("content is not a valid JSON blob")
)

func MustJSONStr(in string) JSONBlob {
	js, err := AsJSONStr(in)
	if err != nil {
		panic(err)
	}
	return js
}

func AsJSONStr(in string) (JSONBlob, error) {
	bin := []byte(in)
	if !json.Valid(bin) {
		return JSONBlob{}, errNotJSON
	}
	return JSONBlob(in), nil
}

func (j *JSONBlob) Value() (driver.Value, error) {
	if j == nil || len(*j) == 0 {
		return nil, nil
	}
	return *j, nil
}

func (j *JSONBlob) Scan(v any) error {
	if j == nil {
		return errors.New("cannot scan into nil JSONStr")
	}
	switch v := v.(type) {
	case string:
		*j = []byte(v)
	case []byte:
		*j = v
	default:
		return fmt.Errorf("cannot cast from %T to JSONStr", v)
	}
	if !json.Valid(*j) {
		return errNotJSON
	}
	return nil
}
