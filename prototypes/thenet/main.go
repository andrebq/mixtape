package main

import (
	"context"
	"time"
)

func main() {
	mustBackgroundProvider(context.Background())
	go runServer()
	time.Sleep(time.Second)
	go runClient()
	select {}
}
