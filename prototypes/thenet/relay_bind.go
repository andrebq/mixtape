package main

import (
	"errors"

	"golang.zx2c4.com/wireguard/conn"
)

type (
	RelayBind struct{}
)

func (d *RelayBind) Open(port uint16) (fns []conn.ReceiveFunc, actualPort uint16, err error) {
	return nil, 0, errors.ErrUnsupported
}

func (d *RelayBind) Close() error {
	return errors.ErrUnsupported
}

func (d *RelayBind) SetMark(mark uint32) error {
	return nil
}

func (d *RelayBind) Send(bufs [][]byte, ep conn.Endpoint) error {
	return errors.ErrUnsupported
}

func (d *RelayBind) ParseEndpoint(s string) (conn.Endpoint, error) {
	return nil, errors.ErrUnsupported
}

func (d *RelayBind) BatchSize() int {
	return 10 // Arbitrary batch size
}
