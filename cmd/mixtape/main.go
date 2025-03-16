package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/andrebq/mixtape/internal/rpc"
	"github.com/andrebq/mixtape/taskman"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.App{
		Name: "mixtape",
		Commands: []*cli.Command{
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
