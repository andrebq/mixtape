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

func runClient(addr string, privateKey string, nodes []Node, bind conn.Bind) {
	tun, tnet, err := netstack.CreateNetTUN(
		[]netip.Addr{netip.MustParseAddr(addr)},
		[]netip.Addr{netip.MustParseAddr("8.8.8.8")},
		1420)
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
