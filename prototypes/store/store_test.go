package store_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/andrebq/mixtape/prototypes/store"
	"github.com/jmoiron/sqlx"
)

func TestSQLGen(t *testing.T) {
	type Task struct {
		_              struct{} `ddl:"table=tasks"`
		ID             string   `db:"id" ddl:"primary key"`
		Script         string   `db:"script" ddl:"not null"`
		UserParameters string   `db:"user_parameters" ddl:"type=blob"`
		TTL            int64    `db:"ttl" ddl:"not null"`
		Completed      bool     `db:"completed"`
	}

	db, err := sqlx.Open("sqlite", "file::memory:?cache=shared")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	typeInfo := reflect.TypeOf(Task{})
	store.MustRegister(typeInfo)
	err = store.Migrate(context.Background(), db, typeInfo)
	if err != nil {
		t.Fatal(err)
	}

	// usually this would be a new version, of the same type
	// but in order to properly test migrations, this is a new type
	// but mapped to the same table.
	type TaskV2 struct {
		_              struct{}      `ddl:"table=tasks"`
		ID             string        `db:"id" ddl:"primary key"`
		Script         string        `db:"script" ddl:"not null"`
		UserParameters string        `db:"user_parameters" ddl:"type=blob"`
		TTL            time.Duration `db:"ttl" ddl:"not null"`
		Completed      bool          `db:"completed"`
		NewField       string        `db:"new_field"`
	}

	typeInfoV2 := reflect.TypeFor[TaskV2]()
	err = store.Migrate(context.Background(), db, typeInfoV2)
	if err == nil {
		t.Fatal("This should have failed, since the type was not registered")
	} else if !errors.Is(err, store.ErrNotMapped{typeInfoV2}) {
		t.Fatalf("Unexpected error type: %#v", err)
	}
	store.MustRegister(typeInfoV2)
	err = store.Migrate(context.Background(), db, typeInfoV2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	task := TaskV2{
		ID:             "abc123",
		Script:         "something",
		UserParameters: `{"key":"value"}`,
		TTL:            time.Minute,
		Completed:      false,
		NewField:       "new field",
	}
	err = store.Upsert(context.Background(), db, task)
	if err != nil {
		t.Fatalf("Unable to insert objec: %v", err)
	}

	if found, err := store.LookupOne(context.Background(), db, TaskV2{ID: task.ID}); err != nil {
		t.Fatalf("Unable to perform single row lookup: %v", err)
	} else if !reflect.DeepEqual(found, task) {
		t.Fatalf("Data mismatch, expecting \n%#v\ngot\n%#v", task, found)
	}
}
