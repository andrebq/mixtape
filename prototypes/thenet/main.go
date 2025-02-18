package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/andrebq/mixtape/api"
	"github.com/andrebq/mixtape/generics"
	"github.com/andrebq/mixtape/prototypes/thenet/relay"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	ctx := context.Background()
	go startServer()
	go startAlice(ctx)
	go startBob(ctx)
	select {}
}

func startServer() {
	err := relay.Run(context.Background())
	if err != nil {
		log.Panicln(err)
	}
}

func startAlice(ctx context.Context) {
	startPeer(ctx, "10.0.0.1", 1)
}

func startBob(ctx context.Context) {
	startPeer(ctx, "10.0.0.2", 2)
}

func startPeer(ctx context.Context, vip string, peerId int64) {
	// Create a connection to the server at localhost:9001
	conn, err := grpc.NewClient("localhost:9001",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           backoff.Config{MaxDelay: time.Second * 5},
			MinConnectTimeout: time.Minute,
		}))
	if err != nil {
		log.Panicf("did not connect: %v", err)
	}
	// Return an instance of the RelayClient interface provided by the protobuf package
	relayCli := api.NewRelayClient(conn)
	callCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"source", strconv.FormatInt(peerId, 10),
	))
	otherPeers, err := relayCli.RegisterPeer(callCtx, &api.Peer{
		VirtualIp: vip,
		PeerId:    peerId,
	})
	if err != nil {
		log.Panicf("Unable to register: %v", err)
	}

	stream, err := relayCli.Proxy(callCtx)
	if err != nil {
		log.Panicf("Unable to open proxy: %v", err)
	}

	peers := generics.SetOf[int64]()
	for _, v := range otherPeers.GetPeers() {
		peers.PutAll(v.GetPeerId())
	}
	peers.PutAll(1, 2)

	streamCell := generics.Cell[api.Relay_ProxyClient]{}
	streamCell.Put(stream)

	tick := time.NewTicker(time.Second)
	ping := func() {
		println("starting ping for", peerId)
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-tick.C:
				streamCell.Use(func(r api.Relay_ProxyClient) {
					content := []byte(fmt.Sprintf("ping from %v [%v]", peerId, now.Format(time.RFC3339)))
					for _, id := range peers.AppendTo(nil) {
						if id == peerId {
							continue
						}
						r.Send(&api.Packet{
							Source:      peerId,
							Destination: id,
							Content:     content,
						})
					}
				})
			}
		}
	}
	pong := func() {
		for {
			pkt, err := streamCell.Get().Recv()
			if err != nil {
				log.Panicf("unable to get data: %v", err)
			}
			content := string(pkt.GetContent())
			fmt.Fprintf(os.Stdout, "out[%v]: %v", peerId, content)
			// we received a ping, time to send a pong
			if strings.HasPrefix(content, "ping") {
				// we got a ping from another client
				peers.PutAll(pkt.Source)
				// lets send a pong
				pongMsg := []byte(fmt.Sprintf("pong: %v", content))
				streamCell.Use(func(r api.Relay_ProxyClient) {
					r.Send(&api.Packet{
						Source:      peerId,
						Destination: pkt.Source,
						Content:     pongMsg,
					})
				})
			} else if strings.HasPrefix(content, "pong") {
				// do nothing... we already printed it out before
				peers.PutAll(pkt.Source)
			}
		}
	}
	go ping()
	go pong()
	<-ctx.Done()
}
