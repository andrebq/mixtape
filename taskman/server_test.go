package taskman_test

import (
	"context"
	"testing"

	"github.com/andrebq/mixtape/internal/testutil"
	"github.com/andrebq/mixtape/taskman"
)

func TestServerSetup(t *testing.T) {
	dir, done := testutil.TempDir(t)
	defer done()
	srv, err := taskman.NewServer(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := srv.Close(); err != nil {
		t.Fatal(err)
	}
}
