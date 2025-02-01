package mailbox

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/andrebq/mixtape/generics"
	"github.com/google/uuid"
)

type (
	Rack struct {
		l            sync.RWMutex
		msglog       chan *Message
		newFollower  chan chan<- *Message
		newConsumer  chan consumer
		dropConsumer chan consumer
		closed       chan signal
	}

	signal struct{}

	consumer struct {
		inbox  uuid.UUID
		output chan *Message
	}
)

var (
	ErrInboxNotFound = errors.New("inbox not found")
	ErrRackClosed    = errors.New("rack closed")
)

func NewRack() *Rack {
	r := &Rack{
		closed:       make(chan signal),
		msglog:       make(chan *Message, 1000),
		newFollower:  make(chan chan<- *Message),
		newConsumer:  make(chan consumer, runtime.NumCPU()*2),
		dropConsumer: make(chan consumer, runtime.NumCPU()*2),
	}
	go r.runLog()
	return r
}

func (r *Rack) runLog() {
	followers := map[chan<- *Message]struct{}{}
	consumers := map[chan<- *Message]uuid.UUID{}
	defer func() {
		for k := range followers {
			close(k)
		}
		for k := range consumers {
			close(k)
		}
	}()
	oldMessages := map[*Message]time.Time{}
	for {
		select {
		case <-r.closed:
			return
		case c := <-r.newConsumer:
			now := time.Now()
			delivered := false
			for msg, expire := range oldMessages {
				if expire.Before(now) {
					delete(oldMessages, msg)
				}
				if c.inbox == msg.To.Node {
					delivered = generics.NonBlockSend(c.output, msg)
					delete(oldMessages, msg)
					break
				}
			}
			if delivered {
				// old message already sent
				continue
			}
			consumers[c.output] = c.inbox
		case c := <-r.dropConsumer:
			delete(consumers, c.output)
		case nf := <-r.newFollower:
			followers[nf] = struct{}{}
		case m := <-r.msglog:
			for k := range followers {
				generics.NonBlockSend(k, m)
			}
			var delivered bool
			for k, v := range consumers {
				if v == m.To.Node {
					delivered = generics.NonBlockSend(k, m)
					delete(consumers, k)
				}
			}
			if !delivered {
				// add to old messages if space is available
				now := time.Now()
				for old, expire := range oldMessages {
					if expire.Before(now) {
						// TODO: delete oldest messages instead of just expired ones?
						delete(oldMessages, old)
					}
				}
				if len(oldMessages) < 1000 {
					oldMessages[m] = time.Now().Add(time.Minute)
				}
			}
		}
	}
}

func (r *Rack) Close() error {
	r.l.Lock()
	select {
	case <-r.closed:
		r.l.Unlock()
		return nil
	default:
		close(r.closed)
		r.l.Unlock()
		return nil
	}
}

func (r *Rack) Deliver(ctx context.Context, msg *Message) error {
	select {
	case r.msglog <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Rack) MessageLog(buf int) <-chan *Message {
	if buf < 0 {
		buf = 1
	}
	ch := make(chan *Message, buf)
	r.newFollower <- ch
	return ch
}

func (r *Rack) Take(ctx context.Context, node uuid.UUID) (*Message, error) {
	cons := consumer{
		inbox:  node,
		output: make(chan *Message, 1),
	}
	defer generics.NonBlockSend(r.dropConsumer, cons)
	select {
	case r.newConsumer <- cons:
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-r.closed:
			return nil, ErrRackClosed
		case m := <-cons.output:
			return m, nil
		}
	case <-r.closed:
		return nil, ErrRackClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
