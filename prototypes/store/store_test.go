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
		_              struct{}       `ddl:"table=tasks"`
		ID             string         `db:"id" ddl:"primary key"`
		Script         string         `db:"script" ddl:"not null"`
		UserParameters string         `db:"user_parameters" ddl:"type=blob"`
		TTL            int64          `db:"ttl" ddl:"not null"`
		Completed      bool           `db:"completed"`
		Config         store.JSONBlob `db:"config" ddl:"not null"`
	}

	db, err := sqlx.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	typeInfo := reflect.TypeOf(Task{})
	store.MustRegister[Task]()
	err = store.Migrate(context.Background(), db, typeInfo)
	if err != nil {
		t.Fatal(err)
	}

	// usually this would be a new version, of the same type
	// but in order to properly test migrations, this is a new type
	// but mapped to the same table.
	type TaskV2 struct {
		_              struct{}       `ddl:"table=tasks"`
		ID             string         `db:"id" ddl:"primary key"`
		Script         string         `db:"script" ddl:"not null"`
		UserParameters string         `db:"user_parameters" ddl:"type=blob"`
		TTL            time.Duration  `db:"ttl" ddl:"not null"`
		Completed      bool           `db:"completed"`
		NewField       string         `db:"new_field"`
		Config         store.JSONBlob `db:"config" ddl:"not null"`
	}

	typeInfoV2 := reflect.TypeFor[TaskV2]()
	err = store.Migrate(context.Background(), db, typeInfoV2)
	if err == nil {
		t.Fatal("This should have failed, since the type was not registered")
	} else if !errors.Is(err, store.ErrNotMapped{typeInfoV2}) {
		t.Fatalf("Unexpected error type: %#v", err)
	}
	store.MustRegister[TaskV2]()
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
		Config:         store.MustJSONStr(`{"hello":"world"}`),
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

func TestPatternMatch(t *testing.T) {
	type Record struct {
		_       struct{}       `ddl:"table=tasks,default_sort=id asc"`
		ID      string         `db:"id" ddl:"primary key"`
		Integer int32          `db:"int_val"`
		Float   float32        `db:"float"`
		Time    int64          `db:"time"`
		Config  store.JSONBlob `db:"config" ddl:"not null"`
	}

	db, err := sqlx.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	typeInfo := reflect.TypeOf(Record{})
	store.MustRegister[Record]()
	err = store.Migrate(context.Background(), db, typeInfo)
	if err != nil {
		t.Fatal(err)
	}

	items := []Record{
		{ID: "1", Integer: 10, Float: 10.0, Time: time.Now().Truncate(time.Second).UnixMilli(), Config: []byte(`{"hello":"world"}`)},
		{ID: "2", Integer: 10, Float: 10.5, Time: time.Now().Truncate(time.Second).UnixMilli(), Config: []byte(`{"hello":"world2"}`)},
		{ID: "3", Integer: 12, Float: 10.5, Time: time.Now().Truncate(time.Second).UnixMilli(), Config: []byte(`{"hello":"world2"}`)},
	}
	for _, i := range items {
		if err := store.Upsert(context.Background(), db, i); err != nil {
			t.Fatal(err)
		}
	}
	matches, err := store.Match(context.Background(), []Record(nil), db, map[string]any{
		"Integer": 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	expected := []Record{items[0], items[1]}
	if !reflect.DeepEqual(expected, matches) {
		t.Fatalf("Expecting\n%v\ngot\n%v", expected, matches)
	}

	matches, err = store.Match(context.Background(), []Record(nil), db, map[string]any{
		"Integer": 10,
		"Float":   10.5,
	})
	if err != nil {
		t.Fatal(err)
	}
	expected = []Record{items[1]}
	if !reflect.DeepEqual(expected, matches) {
		t.Fatalf("Expecting\n%v\ngot\n%v", expected, matches)
	}
}
