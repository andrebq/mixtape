package main

import (
	"context"
	"errors"
	"sync"

	"github.com/andrebq/mixtape/prototypes/thenet/relay"
	"golang.zx2c4.com/wireguard/conn"
)

type (
	RelayBind struct {
		sync.RWMutex
		Session relay.Session
		Ctx     context.Context
	}
)

var (
	errClosed = errors.New("closed")
)

func locked(m *sync.RWMutex, fn func()) {
	m.Lock()
	defer m.Unlock()
	fn()
}

func (r *RelayBind) Open(port uint16) (fns []conn.ReceiveFunc, actualPort uint16, err error) {
	var session relay.Session
	locked(&r.RWMutex, func() {
		session = r.Session
	})
	if session == nil {
		err = errClosed
		return
	}
	return []conn.ReceiveFunc{func(packets [][]byte, sizes []int, eps []conn.Endpoint) (n int, err error) {
		p, err := session.Read(r.Ctx)
		if err != nil {
			return 0, err
		}
		println(r.Session.Endpoint().SrcToString(), "recv", len(p.Buffer), "from", p.Endpoint.SrcToString())
		packets[0] = p.Buffer
		sizes[0] = len(p.Buffer)
		eps[0] = p.Endpoint
		return 1, nil
	}}, port, nil
}

func (r *RelayBind) Close() (err error) {
	// the session should be closed directly, not via RelayBind
	return nil
}

func (r *RelayBind) SetMark(mark uint32) error {
	return nil
}

func (r *RelayBind) Send(bufs [][]byte, ep conn.Endpoint) error {
	var session relay.Session
	locked(&r.RWMutex, func() {
		session = r.Session
	})
	if session == nil {
		return errClosed
	}
	for _, b := range bufs {
		if len(b) == 0 {
			continue
		}
		println(r.Session.Endpoint().SrcToString(), "sent", len(b), "to", ep.SrcToString())
		session.Write(r.Ctx, relay.Packet{
			Endpoint: ep.(relay.Endpoint),
			Buffer:   b,
		})
	}
	return nil
}

func (r *RelayBind) ParseEndpoint(s string) (conn.Endpoint, error) {
	return relay.ParseEndpoint(s)
}

func (r *RelayBind) BatchSize() int {
	return 10 // Arbitrary batch size
}
