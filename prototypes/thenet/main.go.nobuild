package main

import (
	"context"
	"log"
	"time"

	"github.com/andrebq/mixtape/prototypes/thenet/relay"
)

func main() {
	mustBackgroundProvider(context.Background())
	cli := relay.NewClient()
	server, err := cli.Dial(context.TODO())
	if err != nil {
		log.Panicln(err)
	}
	serverPrivateKey := "003ed5d73b55806c30de3f8a7bdab38af13539220533055e635690b8b87ad641"
	serverNode := Node{
		Addr: server.IP().String(),
		Key:  "c4c8e984c5322c8184c72265b92b250fdb63688705f504ba003c88f03393cf28",
	}
	client, err := cli.Dial(context.TODO())
	clientPrivateKey := "087ec6e14bbed210e7215cdc73468dfa23f080a1bfb8665b2fd809bd99d28379"
	clientNode := Node{
		Addr: client.IP().String(),
		Key:  "f928d4f6c1b86c12f2562c10b07c555c5c57fd00f59e90c8d8d88767271cbf7c",
	}
	if err != nil {
		log.Panicln(err)
	}
	go runServer(server.IP().String(), serverPrivateKey, []Node{clientNode}, &RelayBind{
		Ctx:     context.TODO(),
		Session: server,
	})
	time.Sleep(time.Second)
	go runClient(client.IP().String(), clientPrivateKey, []Node{serverNode}, &RelayBind{
		Ctx:     context.TODO(),
		Session: client,
	})
	select {}
}
