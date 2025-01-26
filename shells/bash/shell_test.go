package bash_test

import (
	"context"
	"strings"
	"testing"

	"github.com/andrebq/mixtape/internal/testutil"
	"github.com/andrebq/mixtape/shells/bash"
)

func TestSimpleShell(t *testing.T) {
	db, done := testutil.TempDB(t)
	defer done(t)
	output := strings.Builder{}
	err := bash.Eval(context.Background(), &output, db, `
	put Pages author "Bob Bobson" publishDate "2025-01-01" wordCount 200 oid "6a5da2a8-78e0-aaaa-9a3f-ac537c64256a"
	match Pages -l author "Bob Bobson" -p publishDate -p oid
	`)
	if err != nil {
		t.Fatal(err)
	}
	expectedOutput := "6a5da2a8-78e0-aaaa-9a3f-ac537c64256a\n[{\"oid\":\"6a5da2a8-78e0-aaaa-9a3f-ac537c64256a\",\"publishDate\":\"2025-01-01\"}]\n"
	if output.String() != expectedOutput {
		t.Fatalf("Expecting output to be:\n%v\ngot\n%v", expectedOutput, output.String())
	}
}
