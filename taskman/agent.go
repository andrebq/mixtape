package taskman

import (
	"context"
	"time"

	"github.com/andrebq/mixtape/internal/rpc"
	"github.com/google/uuid"
)

type (
	Agent struct {
		registerAgent func(context.Context, AgentDetails) (AgentDetails, error)
		listAgents    func(context.Context, struct{}) ([]AgentDetails, error)
	}
)

var (
	rootUUID = uuid.Must(uuid.NewRandom())
)

func NewAgent(rc *rpc.Client) *Agent {
	return &Agent{
		registerAgent: rpc.RemoteCall[AgentDetails, AgentDetails](rc, "@taskman", "RegisterAgent"),
		listAgents:    rpc.RemoteCall[struct{}, []AgentDetails](rc, "@taskman", "ListAgents"),
	}
}

func (a *Agent) SelfRegister(ctx context.Context, name string) error {
	_, err := a.registerAgent(ctx, AgentDetails{
		ID:            string(uuid.NewSHA1(rootUUID, []byte(time.Now().Format(time.RFC3339Nano)+name)).String()),
		Name:          name,
		MaxConcurrent: 1,
		Executors:     []string{"shell"},
	})
	return err
}

func (a *Agent) ListPeers(ctx context.Context) ([]AgentDetails, error) {
	return a.listAgents(ctx, struct{}{})
}
