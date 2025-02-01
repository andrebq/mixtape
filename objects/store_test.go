package objects_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/andrebq/mixtape/objects"
)

func TestSimpleOperations(t *testing.T) {
	st, err := objects.MemoryStorage()
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	sess := st.Session(context.TODO())
	type Person struct {
		ID   objects.OID `msgpack:"_id"`
		Kind string      `msgpack:"_kind"`
		Name string
	}
	np := Person{
		Kind: "Person",
		Name: "Bob",
	}
	ref, err := objects.Put(context.TODO(), sess, np)
	if err != nil {
		t.Fatal(err)
	}
	np.ID = ref.ID
	var found Person
	err = objects.Get(context.TODO(), &found, sess, ref)
	if err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(found, np) {
		t.Fatalf("expecting %#v but got %#v", np, found)
	}
}

func TestTransactionCommit(t *testing.T) {
	st, err := objects.MemoryStorage()
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	sess := st.Session(context.TODO())
	type Person struct {
		ID   objects.OID `msgpack:"_id"`
		Kind string      `msgpack:"_kind"`
		Name string
	}
	np := Person{
		Kind: "Person",
		Name: "Bob",
	}
	ref, err := objects.Put(context.TODO(), sess, np)
	if err != nil {
		t.Fatal(err)
	}
	if err := sess.Commit(); err != nil {
		t.Fatal(err)
	}
	np.ID = ref.ID
	sess = st.Session(context.TODO())
	var found Person
	if err := objects.Get(context.TODO(), &found, sess, ref); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(np, found) {
		t.Fatalf("Expecting %#v got %#v", np, found)
	}
	sess.Close()
}

func TestTransactionRollback(t *testing.T) {
	st, err := objects.MemoryStorage()
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	sess := st.Session(context.TODO())
	type Person struct {
		ID   objects.OID `msgpack:"_id"`
		Kind string      `msgpack:"_kind"`
		Name string
	}
	np := Person{
		Kind: "Person",
		Name: "Bob",
	}
	ref, err := objects.Put(context.TODO(), sess, np)
	if err != nil {
		t.Fatal(err)
	}
	sess.Close()
	sess = st.Session(context.TODO())
	var found Person
	err = objects.Get(context.TODO(), &found, sess, ref)
	if !errors.Is(err, objects.ErrNotFound) {
		t.Fatalf("Should have returned not found since the session ended without a commit, but got %v", err)
	}
}
