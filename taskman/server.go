package taskman

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/andrebq/mixtape/api"
	"google.golang.org/grpc"
)

type (
	TaskManagerServer struct {
		api.UnsafeTaskManagerServer
	}
)

func (s *TaskManagerServer) RegisterSupervisor(ctx context.Context, req *api.SupervisorStats) (*api.SupervisorConfig, error) {
	slog.InfoContext(ctx, "RegisterSupervisor called")
	return &api.SupervisorConfig{}, nil
}

func (s *TaskManagerServer) FetchTask(ctx context.Context, req *api.RunnerSpec) (*api.NextTask, error) {
	slog.InfoContext(ctx, "FetchTask called")
	return &api.NextTask{}, nil
}

func (s *TaskManagerServer) AppendLog(ctx context.Context, req *api.LogEntry) (*api.Empty, error) {
	slog.InfoContext(ctx, "Appendslog called")
	return &api.Empty{}, nil
}

func (s *TaskManagerServer) UploadAsset(ctx context.Context, req *api.Asset) (*api.AssetRef, error) {
	slog.InfoContext(ctx, "UploadAsset called")
	return &api.AssetRef{}, nil
}

func (s *TaskManagerServer) WaitForInput(ctx context.Context, req *api.InputRequest) (*api.InputResponse, error) {
	slog.InfoContext(ctx, "WaitForInput called")
	return &api.InputResponse{}, nil
}

func GenAgentToken(name string, minLabels []string) (string, error) {
	return "", errors.ErrUnsupported
}

func Handler() (http.Handler, error) {
	server := grpc.NewServer()
	api.RegisterTaskManagerServer(server, &TaskManagerServer{})
	return server, nil
}
