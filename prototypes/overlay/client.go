package overlay

import (
	"errors"

	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"
)

func Send[T any](out *websocket.Conn, v T) error {
	buf, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}
	return out.WriteMessage(websocket.BinaryMessage, buf)
}

func Recv[T any](out *T, conn *websocket.Conn) error {
	for {
		kind, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		if kind == websocket.CloseMessage {
			return errors.New("closed")
		} else if kind != websocket.BinaryMessage {
			continue
		}
		return msgpack.Unmarshal(msg, out)
	}
}
