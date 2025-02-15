package main

import (
	"context"
	"time"

	"golang.zx2c4.com/wireguard/conn"
)

func main() {
	mustBackgroundProvider(context.Background())
	binderfunc := func() (conn.Bind, error) {
		return NewDebugBind(conn.NewDefaultBind()), nil
		//return &RelayBind{}, nil
	}
	go runServer(binderfunc)
	time.Sleep(time.Second)
	go runClient(binderfunc)
	select {}
}
