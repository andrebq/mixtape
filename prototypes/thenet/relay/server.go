package relay

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/andrebq/mixtape/api"
	"github.com/andrebq/mixtape/generics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// relayServer is used to implement api.RelayService.
type relayServer struct {
	api.UnimplementedRelayServer

	endpoints generics.SyncMap[int64, api.Relay_ProxyServer]
	peers     generics.SyncMap[string, *api.Peer]
}

func (s *relayServer) RegisterPeer(ctx context.Context, peer *api.Peer) (*api.PeerList, error) {
	// TODO: check authentication here
	old, _ := s.peers.Put(peer.GetVirtualIp(), peer)
	if old != nil {
		// TODO: how to signal that and old connection should be removed?
		_, _ = s.endpoints.Delete(old.PeerId)
	}
	lst := &api.PeerList{}
	for _, p := range s.peers.LockedIter() {
		lst.Peers = append(lst.Peers, p)
	}
	return lst, nil
}

// Proxy implements the RelayProxy RPC method.
func (s *relayServer) Proxy(stream api.Relay_ProxyServer) error {
	source, err := s.checkAuth(stream.Context())
	if err != nil {
		return fmt.Errorf("invalid authentication: %v", err)
	}
	for {
		packet, err := stream.Recv()
		if err != nil {
			log.Printf("Failed to receive a packet : %v", err)
			return err
		}
		if packet.Source != source {
			// TODO: should we drop the connection instead?
			continue
		}
		conn, _ := s.endpoints.Get(packet.Destination)
		if conn != nil {
			conn.Send(packet)
		}
	}
}

func (s *relayServer) checkAuth(ctx context.Context) (source int64, err error) {
	// TODO: this id should be part of an auth token...
	srcid := metadata.ValueFromIncomingContext(ctx, "source")
	if len(srcid) != 1 {
		err = errors.New("invalid source header")
		return
	}
	source, err = strconv.ParseInt(srcid[0], 16, 64)
	return
}

func Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// Set up the server and start listening on port 50051.
	lis, err := net.Listen("tcp", ":9001")
	if err != nil {
		return fmt.Errorf("unable to listen to port 9001")
	}

	go func() {
		<-ctx.Done()
		// not graceful... but works
		lis.Close()
	}()

	s := grpc.NewServer()
	api.RegisterRelayServer(s, &relayServer{})

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}
