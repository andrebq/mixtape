package testutil

import (
	"os"
	"testing"
)

func TempDir(t testing.TB) (string, func()) {
	d, err := os.MkdirTemp("", "mixtape_*")
	if err != nil {
		t.Fatal(err)
	}
	return d, func() {
		err := os.RemoveAll(d)
		if err != nil {
			t.Logf("unable to cleanup directory %v, cause: %v", d, err)
		}
	}
}
