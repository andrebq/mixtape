package overlay_test

import (
	"fmt"
	"net/http/httptest"
	"net/netip"
	"reflect"
	"strings"
	"testing"

	"github.com/andrebq/mixtape/prototypes/overlay"
	"github.com/gorilla/websocket"
)

func TestServer(t *testing.T) {
	aliceIp := netip.MustParseAddr("10.0.0.1")
	bobIp := netip.MustParseAddr("10.0.0.2")
	srv := httptest.NewServer(overlay.Handler(&overlay.Switch{}))
	defer srv.Close()

	alicePacket := []byte("hello from Alice")
	bobPacket := []byte("hello from Bob")

	bobStarted := make(chan struct{})

	go func() {
		c, _, err := websocket.DefaultDialer.Dial(
			strings.Replace(fmt.Sprintf("%v/overlay/%v/conn", srv.URL, bobIp.String()), "http", "ws", -1), nil)
		if err != nil {
			t.Error("dial:", err)
			return
		}
		defer c.Close()
		var pkt overlay.Packet
		close(bobStarted)
		err = overlay.Recv(&pkt, c)
		if err != nil {
			t.Error("Bob recv error", err)
			return
		}
		if !reflect.DeepEqual(pkt.Buf, alicePacket) {
			t.Errorf("Expecting %v got %v", string(alicePacket), string(pkt.Buf))
		}
		pkt.To, pkt.From = pkt.From, pkt.To
		pkt.Buf = bobPacket
		err = overlay.Send(c, pkt)
		if err != nil {
			t.Error("Bob send error", err)
		}
	}()
	c, _, err := websocket.DefaultDialer.Dial(
		strings.Replace(fmt.Sprintf("%v/overlay/%v/conn", srv.URL, aliceIp.String()), "http", "ws", -1), nil)
	if err != nil {
		t.Fatal("dial:", err)
	}
	defer c.Close()
	var pkt overlay.Packet
	pkt.To, _ = bobIp.MarshalBinary()
	pkt.From, _ = aliceIp.MarshalBinary()
	pkt.Buf = alicePacket
	// wait until bob starts listening
	<-bobStarted
	err = overlay.Send(c, pkt)
	if err != nil {
		t.Fatal(err)
	}

	err = overlay.Recv(&pkt, c)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Alice recv", string(pkt.Buf))
	if !reflect.DeepEqual(pkt.Buf, bobPacket) {
		t.Errorf("Expecting %q got %q", string(bobPacket), string(pkt.Buf))
	}
}
