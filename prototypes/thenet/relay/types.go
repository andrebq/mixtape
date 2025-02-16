package relay

import (
	"net/netip"

	"github.com/vmihailenco/msgpack/v5"
)

// Packet structure
type Packet struct {
	Endpoint Endpoint `msgpack:"e"`
	Buffer   []byte   `msgpack:"b"`
}

type Endpoint struct {
	Addr [4]byte `msgpack:"l"`
}

// Encode packet into a byte slice (placeholder function)
func encodePacket(pkt Packet) ([]byte, error) {
	return msgpack.Marshal(pkt)
}

// Decode received data into a packet (placeholder function)
func decodePacket(data []byte) (out Packet, err error) {
	err = msgpack.Unmarshal(data, &out)
	return
}

func EndpointFromIP(ip netip.Addr) Endpoint {
	return Endpoint{Addr: ip.As4()}
}

func ParseEndpoint(s string) (Endpoint, error) {
	ip, err := netip.ParseAddr(s)
	if err != nil {
		return Endpoint{}, err
	}
	return Endpoint{Addr: ip.As4()}, nil
}

func (e Endpoint) ClearSrc() {
	for i := range e.Addr {
		e.Addr[i] = 0
	}
}

func (e Endpoint) SrcToString() string {
	return netip.AddrFrom4(e.Addr).String()
}

func (e Endpoint) DstToString() string {
	return netip.AddrFrom4(e.Addr).String()
}

func (e Endpoint) DstToBytes() []byte {
	ret := e.Addr
	return ret[:]
}

func (e Endpoint) DstIP() netip.Addr {
	return netip.AddrFrom4(e.Addr)
}

func (e Endpoint) SrcIP() netip.Addr {
	return netip.AddrFrom4(e.Addr)
}
