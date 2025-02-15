package relay

import "github.com/vmihailenco/msgpack/v5"

// Packet structure
type Packet struct {
	Endpoint string `msgpack:"e"`
	Buffer   []byte `msgpack:"b"`
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
