package rpc

import (
	"context"
	"errors"

	"github.com/andrebq/mixtape/generics"
	"github.com/vmihailenco/msgpack/v5"
)

type (
	Actor struct {
		methodMap generics.SyncMap[string, func(context.Context, []byte) ([]byte, error)]
	}

	AsActor interface {
		AsActor(a *Actor)
	}
)

func Wrap(val AsActor) *Actor {
	a := &Actor{}
	val.AsActor(a)
	return a
}

func AddMethod[T any, V any](a *Actor, method string, fn func(context.Context, T) (V, error)) {
	//isPtr := reflect.TypeFor[T]().Kind() == reflect.Pointer
	a.methodMap.Put(method, func(ctx context.Context, b []byte) ([]byte, error) {
		var input T
		var err error
		err = msgpack.Unmarshal(b, &input)
		// if isPtr {
		// 	err = msgpack.Unmarshal(b, input)
		// } else {
		// 	err = msgpack.Unmarshal(b, &input)
		// }
		if err != nil {
			// TODO: make this a user error
			return nil, err
		}
		out, err := fn(ctx, input)
		if err != nil {
			return nil, err
		}
		return msgpack.Marshal(out)
	})
}

func (a *Actor) Dispatch(ctx context.Context, method string, p []byte) ([]byte, error) {
	fn, _ := a.methodMap.Get(method)
	if fn == nil {
		return nil, errors.New("method not found 404")
	}
	return fn(ctx, p)
}

func (a *Actor) Stream(ctx context.Context, method string, stream Stream) error {
	return errors.ErrUnsupported
}
