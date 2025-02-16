package relay

import (
	"context"
	"errors"
	"math"
	"net/netip"
	"sync"

	"github.com/andrebq/mixtape/generics"
)

type (
	Client struct {
		l      sync.Mutex
		nextIP netip.Addr

		chans generics.SyncMap[Endpoint, chan Packet]
	}

	Session interface {
		Endpoint() Endpoint
		IP() netip.Addr
		Read(ctx context.Context) (Packet, error)
		Write(ctx context.Context, pkt Packet) error
		Close() error
	}

	session struct {
		cli  *Client
		ep   Endpoint
		pkts <-chan Packet
	}
)

func NewClient() *Client {
	c := &Client{}
	c.nextIP = netip.MustParseAddr("10.0.0.1")
	return c
}

func (c *Client) Dial(_ context.Context) (Session, error) {
	c.l.Lock()
	defer c.l.Unlock()
	ep := EndpointFromIP(c.nextIP)
	var err error
	c.nextIP, err = incIP(c.nextIP)
	if err != nil {
		return nil, err
	}
	pkts := make(chan Packet, 1000)
	c.chans.Put(ep, pkts)
	return &session{
		cli:  c,
		ep:   ep,
		pkts: pkts,
	}, nil
}

func incIP(old netip.Addr) (netip.Addr, error) {
	parts := old.As4()
	var r byte
	// TODO: is this faster than casting to uint32 add one and overflow at 10.254.254.254?
	parts[3], r = overflowAdd(parts[3], 1, math.MaxUint8-1)
	parts[2], r = overflowAdd(parts[2], r, math.MaxUint8-1)
	parts[1], r = overflowAdd(parts[1], r, math.MaxUint8-1)
	if r != 0 {
		return old, errors.New("IPv4 overflow")
	}
	return netip.AddrFrom4(parts), nil
}

func overflowAdd(in byte, delta byte, max int) (byte, byte) {
	nv := int(in) + int(delta)
	return byte(nv % max), byte(nv / max)
}

func (s *session) Endpoint() Endpoint {
	return s.ep
}

func (s *session) IP() netip.Addr {
	return netip.AddrFrom4(s.ep.Addr)
}

func (s *session) Read(ctx context.Context) (Packet, error) {
	select {
	case val, open := <-s.pkts:
		if !open {
			return Packet{}, errors.New("session closed")
		}
		println("recv", s.ep.SrcToString(), "from", val.Endpoint.SrcToString())
		return val, nil
	case <-ctx.Done():
		return Packet{}, ctx.Err()
	}
}

func (s *session) Write(_ context.Context, pkt Packet) error {
	s.cli.chans.Use(pkt.Endpoint, func(output chan Packet) {
		// indicate here who is sending the packet
		pkt.Endpoint = s.ep
		select {
		case output <- pkt:
		default:
			// drop without blocking
		}
	})
	return nil
}

func (s *session) Close() error {
	s.cli.chans.Update(s.ep, func(v chan Packet, present bool) (newval chan Packet, keep bool) {
		if present {
			close(v)
		}
		return nil, false
	})
	return nil
}
