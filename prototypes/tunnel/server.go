package tunnel

import (
	"context"
	"io"
	"log/slog"
	"net"

	"github.com/gliderlabs/ssh"
)

func Run(ctx context.Context) error {
	port := "127.0.0.1:2222"
	slog.Info("starting ssh server", "port", port)

	forwardHandler := &ssh.ForwardedTCPHandler{}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	d2s := DNS2Socket{}

	server := ssh.Server{
		LocalPortForwardingCallback: ssh.LocalPortForwardingCallback(func(ctx ssh.Context, dhost string, dport uint32) bool {
			return true
		}),
		DialForLocalPortForward: func(ctx ssh.Context, bindHost string, bindPort uint32) (net.Conn, error) {
			return d2s.Dial(bindHost, bindPort)
		},
		Addr: port,
		Handler: ssh.Handler(func(s ssh.Session) {
			io.WriteString(s, "Remote forwarding available...\n")
			<-s.Context().Done()
		}),
		ReversePortForwardingCallback: ssh.ReversePortForwardingCallback(func(ctx ssh.Context, host string, port uint32) bool {
			return true
		}),
		ListenerForReverseForward: func(ctx ssh.Context, bindHost string, bindPort uint32) (net.Listener, error) {
			return d2s.CreateNew(bindHost, bindPort)
		},
		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        forwardHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
		},
	}
	go func() {
		<-ctx.Done()
		server.Close()
	}()

	return server.ListenAndServe()
}
