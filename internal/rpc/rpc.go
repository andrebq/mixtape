package rpc

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"path"
	"strconv"
	"sync/atomic"

	"github.com/andrebq/mixtape/generics"
)

type (
	UserError interface {
		Status() int
		UserError() string
	}

	ActorID  uint64
	Dispatch struct {
		nextActorID uint64
		actors      generics.SyncMap[ActorID, ActorHandler]
		alias       generics.SyncMap[string, uint64]
	}

	Stream interface {
		Read([]byte) ([]byte, error)
		Send([]byte) error
	}

	ActorHandler interface {
		Dispatch(ctx context.Context, method string, payload []byte) ([]byte, error)
		Stream(ctx context.Context, method string, stream Stream) error
	}
)

var (
	_ ActorHandler = (*Actor)(nil)
)

func (d *Dispatch) AddActor(handler ActorHandler) ActorID {
	// TODO: think of something more random
	nid := atomic.AddUint64(&d.nextActorID, 1)
	d.actors.Put(ActorID(nid), handler)
	return ActorID(nid)
}

func (d *Dispatch) RegisterAlias(id ActorID, alias string) {
	d.alias.Put(alias, uint64(id))
}

func (d *Dispatch) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "only POST", http.StatusMethodNotAllowed)
		return
	}
	kindAndActor, method := path.Split(req.URL.Path)
	kind, actor := path.Split(path.Clean(kindAndActor))
	id, found := d.alias.Get(actor)
	if !found {
		var err error
		id, err = strconv.ParseUint(actor, 36, 64)
		if err != nil {
			http.Error(w, "actor not found", http.StatusNotFound)
			return
		}
	}
	actorHandler, _ := d.actors.Get(ActorID(id))
	if actorHandler == nil {
		http.Error(w, "actor not found", http.StatusNotFound)
		return
	}
	switch path.Base(kind) {
	case "call":
		d.handleCall(w, req, method, actorHandler)
	case "stream":
		http.Error(w, "operation kind not supported", http.StatusBadRequest)
		return
	}
}

func (d *Dispatch) handleCall(w http.ResponseWriter, req *http.Request, method string, actor ActorHandler) {
	buf, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "operation kind not supported", http.StatusBadRequest)
		return
	}
	buf, err = actor.Dispatch(req.Context(), method, buf)
	if err != nil {
		if ue, ok := err.(UserError); ok {
			http.Error(w, ue.UserError(), ue.Status())
			return
		} else {
			slog.ErrorContext(req.Context(),
				"Error while processing actor request",
				"path", req.URL.Path,
				"err", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Add("Content-Length", strconv.Itoa(len(buf)))
	w.Header().Add("Content-Type", "application/vnd.msgpack")
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}

func (d *Dispatch) Run(ctx context.Context, bind string) error {
	lst, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "Starting dispatch server", "listener", lst.Addr())
	srv := http.Server{
		Handler:     d,
		BaseContext: func(l net.Listener) context.Context { return ctx },
	}
	shutdownCompleted := make(chan struct{})
	go func() {
		defer close(shutdownCompleted)
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	err = srv.Serve(lst)
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	if err != nil {
		slog.Error("Server failure", "listener", lst.Addr(), "err", err)
	}
	<-shutdownCompleted
	return err
}
