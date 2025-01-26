package bash

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/andrebq/mixtape/generics"
	"github.com/andrebq/mixtape/schema"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

type (
	maxWriter struct {
		allowed int
		w       io.Writer
	}
)

var (
	ErrNotFound    = errors.New("not found")
	ErrInvalidCall = errors.New("invalid call to builtin command")
	ErrWriteFaield = errors.New("unable to write data to schema")
	ErrReadFaield  = errors.New("uanble to read data from schema")
)

func (m *maxWriter) Write(b []byte) (int, error) {
	if len(b) > m.allowed {
		b = b[:m.allowed]
		n, err := m.w.Write(b)
		if err != nil {
			return n, err
		}
		return n, io.ErrShortWrite
	}
	n, err := m.w.Write(b)
	m.allowed -= n
	return n, err
}

func Eval(ctx context.Context, stdout io.Writer, s *schema.S, script string) error {
	file, _ := syntax.NewParser().Parse(strings.NewReader(script), "__user_script.sh")
	stderr := bytes.Buffer{}
	runner, _ := interp.New(
		interp.Dir("/tmp"),
		interp.Env(expand.ListEnviron("")),
		interp.ExecHandlers(sessionHandlers(s), fail),
		interp.StdIO(nil, stdout, &maxWriter{allowed: 10_000, w: &stderr}),
	)
	err := runner.Run(ctx, file)
	if err != nil {
		return fmt.Errorf("%v: %w", stderr.String(), err)
	}
	return nil
}

func fail(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
	return func(ctx context.Context, args []string) error {
		hctx := interp.HandlerCtx(ctx)
		_, _ = fmt.Fprintf(hctx.Stderr, "%v not found\n", args[0])
		// TODO: ErrNotFound might be confusing, as it could indicate the row (not the command) was not found
		return fmt.Errorf("%v %w", args[0], ErrNotFound)
	}
}

func sessionHandlers(s *schema.S) func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
	return func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
		return func(ctx context.Context, args []string) error {
			switch args[0] {
			case "match":
				return handleMatch(ctx, s, args[1:])
			case "put":
				return handlePut(ctx, s, args[1:])
			}
			return next(ctx, args)
		}
	}
}

func handleMatch(ctx context.Context, s *schema.S, args []string) error {
	hctx := interp.HandlerCtx(ctx)
	table, found, args := generics.ShiftHead(args)
	if !found {
		fmt.Fprintf(hctx.Stderr, "invalid call to match, missing tuple type, usage: match <TupleTypeName> -l <selector field> <selector value> [<selector N> <value N>] -p <projection field> [-p <projection field N>]\n")
		return ErrInvalidCall
	}
	filters := map[string]any{}
	projections := schema.ColumnList{}
	for found && len(args) > 0 {
		var operator string
		var field, filter string
		operator, found, args = generics.ShiftHead(args)
		if operator == "-l" {
			field, found, args = generics.ShiftHead(args)
			if !found {
				fmt.Fprintf(hctx.Stderr, "invalid call to match, missing tuple type, usage: match <TupleTypeName> -l <selector field> <selector value> [<selector N> <value N>] -p <projection field> [-p <projection field N>]\n")
				return ErrInvalidCall
			}
			filter, found, args = generics.ShiftHead(args)
			if !found {
				fmt.Fprintf(hctx.Stderr, "invalid call to match, missing tuple type, usage: match <TupleTypeName> -l <selector field> <selector value> [<selector N> <value N>] -p <projection field> [-p <projection field N>]\n")
				return ErrInvalidCall
			}
			filters[field] = filter
		} else if operator == "-p" {
			field, found, args = generics.ShiftHead(args)
			if !found {
				fmt.Fprintf(hctx.Stderr, "invalid call to match, missing tuple type, usage: match <TupleTypeName> -l <selector field> <selector value> [<selector N> <value N>] -p <projection field> [-p <projection field N>]\n")
				return ErrInvalidCall
			}
			projections = append(projections, schema.ColumnName(field))
		} else if found {
			// we found a value, but wasn't want onf the expected operators
			// therefore it is an error
			fmt.Fprintf(hctx.Stderr, "invalid call to match, missing tuple type, usage: match <TupleTypeName> -l <selector field> <selector value> [<selector N> <value N>] -p <projection field> [-p <projection field N>]\n")
			return ErrInvalidCall
		}
	}
	if len(filters) == 0 || len(projections) == 0 {
		fmt.Fprintf(hctx.Stderr, "invalid call to match, missing tuple type, usage: match <TupleTypeName> -l <selector field> <selector value> [<selector N> <value N>] -p <projection field> [-p <projection field N>]\n")
		return ErrInvalidCall
	}
	matches, err := s.Match(ctx, schema.TableName(table), filters, projections...)
	if err != nil {
		fmt.Fprintf(hctx.Stderr, "%v\n", err)
		return ErrReadFaield
	}
	enc := json.NewEncoder(hctx.Stdout)
	err = enc.Encode(matches)
	if err != nil {
		fmt.Fprintf(hctx.Stderr, "%v\n", err)
		return ErrReadFaield
	}
	return nil
}

func handlePut(ctx context.Context, s *schema.S, args []string) error {
	hctx := interp.HandlerCtx(ctx)
	table, found, args := generics.ShiftHead(args)
	if !found {
		fmt.Fprintf(hctx.Stderr, "invalid call to put, missing tuple type, usage: put <TupleTypeName> <field1> <value1> [<fieldN> <valueN>]\n")
		return ErrInvalidCall
	}
	values := map[string]any{}
	for found {
		var fieldName, fieldValue string
		fieldName, found, args = generics.ShiftHead(args)
		if !found {
			break
		}
		fieldValue, found, args = generics.ShiftHead(args)
		if !found {
			fmt.Fprintf(hctx.Stderr, "invalid call to put, missing tuple type, usage: put <TupleTypeName> <field1> <value1> [<fieldN> <valueN>]\n")
			return ErrInvalidCall
		}
		values[fieldName] = fieldValue
	}
	if len(values) == 0 {
		fmt.Fprintf(hctx.Stderr, "invalid call to put, missing tuple type, usage: put <TupleTypeName> <field1> <value1> [<fieldN> <valueN>]\n")
		return ErrInvalidCall
	}
	oid, err := s.Put(ctx, schema.TableName(table), values)
	if err != nil {
		fmt.Fprintf(hctx.Stderr, "%v\n", err)
		return ErrWriteFaield
	}
	_, err = fmt.Fprintf(hctx.Stdout, "%v\n", oid)
	return err
}
