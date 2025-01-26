package schema_test

import (
	"context"
	"testing"
	"time"

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

	_, err = db.Put(context.Background(), "Pages", map[string]any{
		"title":        "test",
		"content":      "it works",
		"author":       "test",
		"publish_date": time.Now().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Put(context.Background(), "NewTuple", map[string]any{
		"oid":        "28459049-d5cc-5b6e-9e71-9a0bb410ef3a",
		"field":      "value",
		"otherField": "otherValue",
	})
	if err != nil {
		t.Fatal(err)
	}
}
