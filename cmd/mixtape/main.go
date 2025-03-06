package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/andrebq/maestro"
	"github.com/andrebq/mixtape/internal/flagutil"
	"github.com/andrebq/mixtape/internal/rpc"
	"github.com/andrebq/mixtape/prototypes/gateway/ssh"
	"github.com/andrebq/mixtape/taskman"
	"github.com/urfave/cli/v2"
)

const envPrefix = "MIXTAPE_GATEWAY_SSH"

func main() {
	app := cli.App{
		Name: "mixtape",
		Commands: []*cli.Command{
			serverCmd(),
			agentCmd(),
		},
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := app.RunContext(ctx, os.Args); err != nil {
		slog.ErrorContext(ctx, "Application fault", "err", err)
		os.Exit(1)
	}
}

func agentCmd() *cli.Command {
	var name string
	return &cli.Command{
		Name: "agent",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "name", Usage: "Name of the agent", Required: true, EnvVars: []string{"MIXTAPE_TASKMAN_AGENT_NAME"}, Destination: &name},
		},
		Action: func(ctx *cli.Context) error {
			ag := taskman.NewAgent(rpc.NewClient("http://localhost:9001"))
			if err := ag.SelfRegister(ctx.Context, name); err != nil {
				return err
			}
			others, err := ag.ListPeers(ctx.Context)
			if err != nil {
				return err
			}
			slog.InfoContext(ctx.Context, "Peers", "agents", others)
			return nil
		},
	}
}

func serverCmd() *cli.Command {
	bind := "127.0.0.1:2222"
	bindHTTP := "127.0.0.1:8222"
	adminKey := ""
	kdbStoreDir := ""
	caSeed := ""
	caSeedFlag := flagutil.String(&caSeed, "ca-seed", nil, envPrefix, "32-byte, hex-encoded, seed used to generate a ed25519 private key, use the environment variable", true)
	caSeedFlag.Hidden = true

	subdomains := cli.StringSlice{}
	selfDomains := cli.StringSlice{}
	return &cli.Command{
		Name: "server",
		Flags: []cli.Flag{
			flagutil.String(&bind, "bind-addr", []string{"b"}, envPrefix, "Address to listen for incoming requests", false),
			flagutil.String(&adminKey, "admin-key", nil, envPrefix, "SSH public key file used for admin access", true),
			flagutil.String(&kdbStoreDir, "keydb-store-dir", nil, envPrefix, "Directory where key database is kept", true),
			flagutil.String(&bindHTTP, "bind-http-addr", []string{"bh"}, envPrefix, "Address to listen for HTTP Requests", false),
			flagutil.StringSlice(&selfDomains, "self-domain", []string{"self"}, envPrefix, "Address (domain:port) of the gateway itself. Must be a value recognized by clients", true),
			flagutil.StringSlice(&subdomains, "domain", []string{"d"}, envPrefix, "One or more sub-domains which can be authorized by this gateway", true),
			caSeedFlag,
		},
		Action: func(ctx *cli.Context) error {
			srv, err := taskman.NewServer(ctx.Context, "./taskman")
			if err != nil {
				return err
			}
			mctx := maestro.New(ctx.Context)
			dispatch := rpc.Dispatch{}
			dispatch.RegisterAlias(dispatch.AddActor(rpc.Wrap(srv)), "@taskman")
			mctx.Spawn(func(ctx maestro.Context) error {
				return dispatch.Run(ctx, "localhost:9001")
			})

			kdb, err := ssh.NewKeyDB("./ssh-gateway")
			if err != nil {
				return err
			}

			var buf [ed25519.SeedSize]byte
			n, err := hex.Decode(buf[:], []byte(caSeed))
			if err != nil {
				return err
			} else if n != ed25519.SeedSize {
				return errors.New("ca-seed should be 32-byte long, hex-encoded")
			}
			// clear the environment
			for _, v := range caSeedFlag.EnvVars {
				os.Setenv(v, "")
			}

			key, err := ssh.ParseAuthorizedKey(adminKey)
			if err != nil {
				return err
			}
			gateway, err := ssh.NewGateway(kdb, key, ssh.GenerateCAKey(buf))
			if err != nil {
				return err
			}
			gateway.Binding.SSH = "127.0.0.1:2222"
			gateway.Binding.HTTP = "127.0.0.1:9002"
			gateway.Binding.Domains = selfDomains.Value()
			ssh.WrapIP(gateway.Binding.Domains)

			gateway.Subdomains = subdomains.Value()
			ssh.WrapIP(gateway.Subdomains)

			mctx.Spawn(func(ctx maestro.Context) error { return gateway.Run(ctx) })
			<-mctx.Done()
			mctx.WaitChildren(maestro.TimeoutAfter(time.Minute))
			return mctx.Err()
		},
	}
}
