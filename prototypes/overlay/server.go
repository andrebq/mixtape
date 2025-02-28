package overlay

import (
	"context"
	"net/http"
	"net/netip"
	"sync/atomic"

	"github.com/andrebq/mixtape/generics"
	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"
)

type (
	Switch struct {
		peers generics.SyncMap[netip.Addr, *Peer]

		connid int64
	}

	Peer struct {
		s      *Switch
		output chan<- Packet
		conns  generics.SyncMap[int64, *websocket.Conn]
	}

	Packet struct {
		To   []byte `msgpack:"t"`
		From []byte `msgpack:"f"`
		Buf  []byte `msgpack:"b"`
	}
)

func Handler(s *Switch) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/overlay/{ip}/conn", s)
	return mux
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *Switch) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ip, err := netip.ParseAddr(req.PathValue("ip"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	toSend := make(chan Packet, 1000)
	s.peers.Update(ip, func(v *Peer, present bool) (newval *Peer, keep bool) {
		newval = v
		keep = true
		if !present {
			newval = &Peer{s: s, output: toSend}
		}
		return
	})
	peer, _ := s.peers.Get(ip)
	if peer == nil {
		conn.Close()
		return
	}
	connid := atomic.AddInt64(&s.connid, 1)
	peer.conns.Put(connid, conn)
	ctx, cancel := context.WithCancel(req.Context())

	conn.SetPingHandler(nil)
	conn.SetPongHandler(nil)
	conn.SetCloseHandler(func(code int, text string) error {
		cancel()
		return nil
	})

	go peer.pumpOut(ctx, conn, toSend)
	peer.pumpIn(ctx, conn)

	peer.conns.Delete(connid)
	if peer.conns.Empty() {
		s.peers.Delete(ip)
	}
}

func (p *Peer) pumpOut(ctx context.Context, conn *websocket.Conn, toSend <-chan Packet) {
	for {
		select {
		case <-ctx.Done():
			return
		case p := <-toSend:
			// TODO: add proper error handling here
			_ = Send(conn, p)
		}
	}
}

func (p *Peer) pumpIn(ctx context.Context, conn *websocket.Conn) {
	for {
		_, _, open := generics.NonBlockRecv(ctx.Done())
		if !open {
			return
		}

		kind, msg, err := conn.ReadMessage()
		println(kind, msg, err)
		if websocket.IsUnexpectedCloseError(err) {
			return
		} else if err != nil {
			// TODO: add proper error handling here
			_ = err
			continue
		}
		println("message from socket")
		if kind != websocket.BinaryMessage {
			continue
		}
		var pkt Packet
		err = msgpack.Unmarshal(msg, &pkt)
		if err != nil {
			// TODO: add proper error handling here
			_ = err
			continue
		}
		ipaddr, _ := netip.AddrFromSlice(pkt.To)
		// TODO: update code to avoid ip-spoofing
		other, _ := p.s.peers.Get(ipaddr)
		if other == nil {
			continue
		}
		println("sending packet to", ipaddr.String())
		if !generics.NonBlockSend(other.output, pkt) {
			println("packet drop to", ipaddr.String())
		} else {
			println("packet sent")
		}
	}
}
