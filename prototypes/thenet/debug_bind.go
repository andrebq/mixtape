package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"golang.zx2c4.com/wireguard/conn"
)

type (
	debugBind struct {
		bind       conn.Bind
		upstream   metric.Int64Counter
		downstream metric.Int64Counter
	}
)

func NewDebugBind(actual conn.Bind) conn.Bind {
	var meter = otel.Meter("thenet/udpConn")
	upstream, err := meter.Int64Counter("pktSend", metric.WithDescription("Upstream"), metric.WithUnit("bytes"))
	if err != nil {
		log.Panicln(err)
	}
	downstream, err := meter.Int64Counter("pktRecv", metric.WithDescription("Downstream"), metric.WithUnit("bytes"))
	if err != nil {
		log.Panicln(err)
	}
	return &debugBind{actual, upstream, downstream}
}

func (d *debugBind) Open(port uint16) (fns []conn.ReceiveFunc, actualPort uint16, err error) {
	fns, actualPort, err = d.bind.Open(port)
	for i, fn := range fns {
		fns[i] = func(packets [][]byte, sizes []int, eps []conn.Endpoint) (n int, err error) {
			for i, sz := range sizes {
				if sz == 0 {
					continue
				}
				attribute := semconv.NetPeerName(eps[i].DstToString())
				d.downstream.Add(context.Background(), int64(sz), metric.WithAttributes(attribute))
			}
			return fn(packets, sizes, eps)
		}
	}
	return
}

func (d *debugBind) Close() error {
	return d.bind.Close()
}

func (d *debugBind) SetMark(mark uint32) error {
	return d.bind.SetMark(mark)
}

func (d *debugBind) Send(bufs [][]byte, ep conn.Endpoint) error {
	attribute := semconv.NetPeerName(ep.DstToString())
	var total int64
	for _, b := range bufs {
		total += int64(len(b))
	}
	d.upstream.Add(context.Background(), total, metric.WithAttributes(attribute))
	return d.bind.Send(bufs, ep)
}

func (d *debugBind) ParseEndpoint(s string) (conn.Endpoint, error) {
	return d.bind.ParseEndpoint(s)
}

func (d *debugBind) BatchSize() int {
	return 10 // Arbitrary batch size
}
