package taskman

import (
	"context"
	"errors"

	"github.com/andrebq/mixtape/api"
)

type (
	srv struct {
		api.UnsafeTaskServiceServer
	}
)

func GenAgentToken(name string, minLabels []string) (string, error) {
	return "", errors.ErrUnsupported
}

func Run(ctx context.Context) error {
	return errors.ErrUnsupported
}
