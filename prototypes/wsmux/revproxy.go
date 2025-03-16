package wsmux

import (
	"errors"
	"net"
	"time"
)

type (
	wsListener struct {
		id string
	}

	wsMuxConn struct {
		laddr    wsAddr
		raddr    wsAddr
		buf      []byte
		rdl, wdl time.Time
	}

	wsAddr struct {
		id string
	}
)

func NewListener(remoteServerWS string) (net.Listener, error) {
	return &wsListener{
		id: remoteServerWS,
	}, errors.ErrUnsupported
}

func (w *wsListener) Addr() net.Addr {
	return wsAddr{
		id: w.id,
	}
}

func (w *wsListener) Close() error {
	return errors.ErrUnsupported
}

func (w *wsListener) Accept() (net.Conn, error) {
	connid := ""
	// connid would be acquired later
	return &wsMuxConn{
		laddr: wsAddr{id: connid},
		raddr: wsAddr{id: w.id},
	}, errors.ErrUnsupported
}

func (wa wsAddr) Network() string { return "ws-proxy" }
func (wa wsAddr) String() string  { return wa.id }

func (w *wsMuxConn) Close() error         { return errors.ErrUnsupported }
func (w *wsMuxConn) LocalAddr() net.Addr  { return w.laddr }
func (w *wsMuxConn) RemoteAddr() net.Addr { return w.raddr }
func (w *wsMuxConn) SetDeadline(val time.Time) error {
	w.SetWriteDeadline(val)
	w.SetReadDeadline(val)
	return nil
}
func (w *wsMuxConn) SetReadDeadline(val time.Time) error {
	w.rdl = val
	return nil
}
func (w *wsMuxConn) SetWriteDeadline(val time.Time) error {
	w.wdl = val
	return nil
}
func (w *wsMuxConn) Write(out []byte) (int, error) {
	return 0, errors.ErrUnsupported
}
func (w *wsMuxConn) Read(out []byte) (int, error) {
	var n int
	if len(w.buf) > 0 {
		n = copy(out, w.buf[:min(len(w.buf), len(out))])
		w.buf = w.buf[n:]
		return n, nil
	}
	var err error
	w.buf, err = w.remotePkt()
	if len(w.buf) > 0 {
		n = copy(out, w.buf[:min(len(w.buf), len(out))])
		w.buf = w.buf[n:]
	}
	return n, err
}
func (w *wsMuxConn) remotePkt() ([]byte, error) {
	return nil, errors.ErrUnsupported
}
