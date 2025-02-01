package mailbox_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/andrebq/mixtape/mailbox"
	"github.com/google/uuid"
)

func TestMailbox(t *testing.T) {
	rack := mailbox.NewRack()
	defer rack.Close()

	oplog := rack.MessageLog(1)
	inbox := uuid.Must(uuid.NewRandom())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second/100)
	defer cancel()

	msg := &mailbox.Message{
		To: mailbox.Address{
			Node:    inbox,
			Process: 1,
		},
		From: mailbox.Address{
			Node:    uuid.Must(uuid.NewRandom()),
			Process: 0,
		},
	}

	rack.Deliver(ctx, msg)

	if oplm := <-oplog; !reflect.DeepEqual(oplm, msg) {
		t.Fatalf("Message from oplog does not match message sent")
	}
	v, err := rack.Take(ctx, inbox)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(v, msg) {
		t.Fatalf("Message from Take does not match message sent")
	}
	_, err = rack.Take(ctx, inbox)
	if err != ctx.Err() {
		t.Fatalf("When there are no messages, the context should expire but got %v", err)
	}
}
