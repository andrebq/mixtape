package testutil

import (
	"os"
	"testing"

	"github.com/andrebq/mixtape/schema"
)

func TempDB(t testing.TB) (*schema.S, func(t testing.TB)) {
	tmpdir, err := os.MkdirTemp("", "mixtape-test-*")
	if err != nil {
		t.Fatal(err)
	}
	db, err := schema.Open(tmpdir)
	if err != nil {
		os.RemoveAll(tmpdir)
		t.Fatal(err)
	}
	return db, func(r testing.TB) {
		err := db.Close()
		if err != nil {
			r.Errorf("Unable to close database at %v, will try to remove the directory anyway", tmpdir)
		}
		os.RemoveAll(tmpdir)
	}
}
