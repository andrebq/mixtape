package rpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

type (
	Client struct {
		base string
	}

	RemoteProcedure[T any, V any] func(context.Context, T) (V, error)
)

func NewClient(base string) *Client {
	return &Client{
		base: strings.TrimSuffix(base, "/"),
	}
}

func RemoteCall[T any, O any](c *Client, actorIdOrAlias string, method string) RemoteProcedure[T, O] {
	isPtr := reflect.TypeFor[O]().Kind() == reflect.Pointer
	return func(ctx context.Context, t T) (O, error) {
		var o O
		buf, err := msgpack.Marshal(t)
		if err != nil {
			return o, err
		}
		buf, err = c.post(ctx, actorIdOrAlias, method, buf)
		if err != nil {
			return o, err
		}
		if isPtr {
			err = msgpack.Unmarshal(buf, o)
		} else {
			err = msgpack.Unmarshal(buf, &o)
		}
		return o, err
	}
}

func (c *Client) post(ctx context.Context, actor, method string, payload []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%v/call/%v/%v", c.base, actor, method), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/vnd.msgpack")
	req.Header.Add("Content-Length", strconv.Itoa(len(payload)))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %v when calling method %v on remote actor",
			res.StatusCode, method)
	}
	return io.ReadAll(res.Body)
}
