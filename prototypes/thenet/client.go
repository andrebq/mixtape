package main

import (
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"net/netip"
	"time"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun/netstack"
)

func runClient(binder func() (conn.Bind, error)) {
	tun, tnet, err := netstack.CreateNetTUN(
		[]netip.Addr{netip.MustParseAddr("192.168.4.28")},
		[]netip.Addr{netip.MustParseAddr("8.8.8.8")},
		1420)
	if err != nil {
		log.Panic(err)
	}
	bind, err := binder()
	if err != nil {
		log.Panic(err)
	}
	dev := device.NewDevice(tun, bind, device.NewLogger(device.LogLevelError, ""))
	err = dev.IpcSet(`private_key=087ec6e14bbed210e7215cdc73468dfa23f080a1bfb8665b2fd809bd99d28379
public_key=c4c8e984c5322c8184c72265b92b250fdb63688705f504ba003c88f03393cf28
allowed_ip=0.0.0.0/0
endpoint=127.0.0.1:58120
`)
	err = dev.Up()
	if err != nil {
		log.Panic(err)
	}

	client := http.Client{
		Transport: &http.Transport{
			DialContext: tnet.DialContext,
		},
	}
	baseURL := "http://192.168.4.29/"
	otherURL := fmt.Sprintf("%vabc", baseURL)
	for {
		url := baseURL
		if rand.Float32() > 0.2 {
			url = otherURL
		}
		resp, err := client.Get(url)
		if err != nil {
			log.Panic(err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Panic(err)
		}
		log.Println(string(body))
		time.Sleep(time.Duration(rand.Int64N(2000)) * time.Millisecond)
	}
}
