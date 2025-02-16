package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

type (
	Node struct {
		Key  string
		Addr string
	}
)

func runServer(addr string, privateKey string, nodes []Node, bind conn.Bind) {
	tun, tnet, err := netstack.CreateNetTUN(
		[]netip.Addr{netip.MustParseAddr(addr)},
		[]netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("8.8.4.4")},
		1420,
	)
	if err != nil {
		log.Panic(err)
	}
	dev := device.NewDevice(tun, bind, device.NewLogger(device.LogLevelError, ""))
	err = dev.IpcSet(`private_key=` + privateKey)
	if err != nil {
		log.Panicln(err)
	}
	for _, n := range nodes {
		dev.IpcSet(fmt.Sprintf("public_key=%v\nallowed_ip=%v/32\nendpoint=%v\npersistent_keepalive_interval=25\n",
			n.Key, n.Addr, n.Addr))
	}
	dev.Up()
	var meter = otel.Meter("thenet/server")
	apiCallCounter, err := meter.Int64Counter("apiCalls", metric.WithDescription("API calls"), metric.WithUnit("{call}"))
	if err != nil {
		log.Panicln(err)
	}
	listener, err := tnet.ListenTCP(&net.TCPAddr{Port: 80})
	if err != nil {
		log.Panicln(err)
	}
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		apiCallCounter.Add(request.Context(),
			1, metric.WithAttributes(semconv.HTTPRoute(request.URL.Path)))

		log.Printf("> %s - %s - %s", request.RemoteAddr, request.URL.String(), request.UserAgent())
		io.WriteString(writer, "Hello from userspace TCP!")
	})
	err = http.Serve(listener, nil)
	if err != nil {
		log.Panicln(err)
	}
}
