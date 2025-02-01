package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/andrebq/mixtape/mailbox"
	"github.com/andrebq/mixtape/mailbox/api"
	"github.com/google/uuid"
)

func TestHandler(t *testing.T) {
	rack := mailbox.NewRack()
	defer rack.Close()
	srv := httptest.NewServer(api.New(rack))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	msg := mailbox.Message{
		ID: uuid.Must(uuid.NewRandom()),
		From: mailbox.Address{
			Node:    uuid.Must(uuid.NewRandom()),
			Process: 1,
		},
		To: mailbox.Address{
			Node:    uuid.Must(uuid.NewRandom()),
			Process: 1,
		},
		Payload: []byte("hello world"),
	}
	if err := api.Post(ctx, http.DefaultClient, srv.URL, &msg); err != nil {
		t.Fatal(err)
	}
	if actual, err := api.Get(ctx, http.DefaultClient, srv.URL, msg.To.Node); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(*actual, msg) {
		t.Fatalf("Expecting msg: \n%#v\ngot\n%#v", msg, *actual)
	}
}
