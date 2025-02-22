package taskman

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"

	"github.com/andrebq/mixtape/api"
	"github.com/andrebq/mixtape/prototypes/store"
	"github.com/andrebq/mixtape/taskman/records"
	"github.com/jmoiron/sqlx"
)

type (
	Server struct {
		api.UnimplementedTaskManagerServer
		db *sqlx.DB
	}

	agentRecord struct {
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

func (s *Server) Close() error {
	return s.db.Close()
}

func (s *Server) ListAgents(context.Context, *api.Empty) (*api.AgentList, error) {
	return nil, errors.ErrUnsupported
}

func (s *Server) ScheduleTask(context.Context, *api.NewTask) (*api.Empty, error) {
	return nil, errors.ErrUnsupported
}
func (s *Server) RegisterAgent(context.Context, *api.AgentDetails) (*api.AgentDetails, error) {
	return nil, errors.ErrUnsupported
}
func (s *Server) NextTask(context.Context, *api.AgentIdentity) (*api.TaskDetails, error) {
	return nil, errors.ErrUnsupported
}
func (s *Server) AppendLog(context.Context, *api.LogEntry) (*api.Empty, error) {
	return nil, errors.ErrUnsupported
}
