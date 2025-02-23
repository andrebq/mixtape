package taskman

import (
	"context"
	"encoding/json"
	"log/slog"
	"path/filepath"
	"reflect"

	"github.com/andrebq/mixtape/generics"
	"github.com/andrebq/mixtape/internal/rpc"
	"github.com/andrebq/mixtape/prototypes/store"
	"github.com/andrebq/mixtape/taskman/records"
	"github.com/jmoiron/sqlx"
)

type (
	Server struct {
		db *sqlx.DB
	}
)

func init() {
	records.RegisterTypes()
}

func NewServer(ctx context.Context, dir string) (*Server, error) {
	var err error
	s := Server{}
	s.db, err = store.OpenDB(filepath.Join(dir, "taskman", "core.db"))
	if err != nil {
		return nil, err
	}
	err = store.Migrate(ctx, s.db, reflect.TypeFor[records.Agent]())
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *Server) AsActor(actor *rpc.Actor) {
	rpc.AddMethod(actor, "ListAgents", s.ListAgents)
	rpc.AddMethod(actor, "RegisterAgent", s.RegisterAgent)
}

func (s *Server) Close() error {
	return s.db.Close()
}

func (s *Server) ListAgents(ctx context.Context, _ struct{}) ([]AgentDetails, error) {
	activeAgents, err := store.Match(ctx, []records.Agent(nil), s.db, map[string]any{
		"Active": true,
	})
	if err != nil {
		return nil, err
	}
	return generics.Transform(activeAgents, []AgentDetails(nil), func(a records.Agent) AgentDetails {
		var executors []string
		if err := json.Unmarshal(a.Executors, &executors); err != nil {
			slog.ErrorContext(ctx, "Unable to decode executor list for agent, will default to empty", "agent.ID", a.ID, "err", err)
		}
		return AgentDetails{
			Name:          a.Name,
			ID:            a.ID,
			Executors:     executors,
			MaxConcurrent: a.MaxConcurrent,
		}
	}), nil
}

func (s *Server) RegisterAgent(ctx context.Context, details *AgentDetails) (*AgentDetails, error) {
	// TODO: add some validation here
	a := records.Agent{
		Name:          details.Name,
		ID:            details.ID,
		MaxConcurrent: details.MaxConcurrent,
		Active:        true,
	}
	buf, err := json.Marshal(details.Executors)
	if err != nil {
		return nil, err
	}
	a.Executors = buf
	err = store.Upsert(ctx, s.db, a)
	if err != nil {
		return nil, err
	}
	return details, nil
}
