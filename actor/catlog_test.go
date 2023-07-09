package actor_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/andrebq/mixtape/actor"
)

func TestClassRegistry(t *testing.T) {
	conn, err := sql.Open("sqlite3", "file::memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	reg, err := actor.NewRegistry(context.Background(), conn)
	if err != nil {
		t.Fatal("Unable to open actor registry", err)
	}
	defer reg.Close()

	_, err = reg.RegisterClass(context.Background(), "hello", `export immutable({
		double: func(msg) { return msg.value * 2 }
	})`)
	if err != nil {
		t.Fatal("Unable to register class in database", err)
	}
	reg.ClearCache()

	clz, err := reg.LoadClass(context.Background(), "hello")
	if err != nil {
		t.Fatal("Unable to load classe", err)
	}
	output, err := clz.Handle(context.Background(), "double", `{"value": 10.0}`)
	if err != nil {
		t.Fatal("Unable to process message on actor", err)
	} else if output.(float64) != 20.0 {
		t.Fatalf("Expecting output to be %v got %v", 20.0, output)
	}
}

func TestCatalog(t *testing.T) {
	conn, err := sql.Open("sqlite3", "file::memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	reg, err := actor.NewRegistry(context.Background(), conn)
	if err != nil {
		t.Fatal("Unable to open actor registry", err)
	}
	defer reg.Close()
}
