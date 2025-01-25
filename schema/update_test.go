package schema_test

import (
	"context"
	"testing"

	"github.com/andrebq/mixtape/internal/testutil"
	"github.com/andrebq/mixtape/schema"
)

func TestSchemaUpdate(t *testing.T) {
	db, done := testutil.TempDB(t)
	defer done(t)
	err := db.Merge(context.Background(), "Pages", schema.ColumnList{
		"title", "content", "author",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = db.Merge(context.Background(), "Pages", schema.ColumnList{
		"publish_date",
	})
	if err != nil {
		t.Fatal(err)
	}
}
