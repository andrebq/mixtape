package api

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/andrebq/mixtape/generics"
	"github.com/andrebq/mixtape/mailbox"
	"github.com/google/uuid"
	"github.com/tinylib/msgp/msgp"
)

func New(rack *mailbox.Rack) http.Handler {
	listeners := generics.SyncMap[uuid.UUID, int]{}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /{id}", func(w http.ResponseWriter, r *http.Request) {
		inbox, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var msg mailbox.Message
		err = msg.DecodeMsg(msgp.NewReader(r.Body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = rack.Deliver(r.Context(), &msg)
		if err != nil {
			// TODO handle internal errors here
			slog.ErrorContext(r.Context(), "Error delivering message for inbox", "inbox", inbox, "error", err, "messageId", msg.ID, "ReplyTo", msg.ReplyTo)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /{id}", func(w http.ResponseWriter, r *http.Request) {
		inbox, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var tooMany bool
		listeners.Update(inbox, func(v int, present bool) (newval int, keep bool) {
			if v == 5 {
				tooMany = true
			}
			return v + 1, true
		})
		if tooMany {
			http.Error(w, "Too many listeners for the given inbox", http.StatusTooManyRequests)
			return
		}
		msg, err := rack.Take(r.Context(), inbox)
		listeners.Update(inbox, func(v int, present bool) (newval int, keep bool) {
			v = v - 1
			return v, v > 0
		})
		if err != nil {
			// TODO handle internal errors here
			slog.ErrorContext(r.Context(), "Error fetching message for inbox", "inbox", inbox, "error", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		// TODO: handle accept headers properly, for now, assume msgpack is acceptable
		buf, err := msg.MarshalMsg(nil)
		if err != nil {
			// TODO handle internal errors here
			slog.ErrorContext(r.Context(), "Error encoding message for inbox", "inbox", inbox, "error", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/vnd.msgpack") // https://www.iana.org/assignments/media-types/application/vnd.msgpack
		w.Header().Add("Content-Length", strconv.Itoa(len(buf)))
		w.WriteHeader(http.StatusOK)
		w.Write(buf)
	})
	return mux
}

func Post(ctx context.Context, cli *http.Client, urlPrefix string, msg *mailbox.Message) error {
	urlPrefix = strings.TrimSuffix(urlPrefix, "/")
	buf, err := msg.MarshalMsg(nil)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%v/%v", urlPrefix, msg.To.Node), bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	res, err := cli.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", res.StatusCode)
	}
	return nil
}

func Get(ctx context.Context, cli *http.Client, urlPrefix string, node uuid.UUID) (*mailbox.Message, error) {
	urlPrefix = strings.TrimSuffix(urlPrefix, "/")
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%v/%v", urlPrefix, node), nil)
	if err != nil {
		return nil, err
	}
	res, err := cli.Do(req)
	if err != nil {
		return nil, err
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %v", res.StatusCode)
	}
	defer res.Body.Close()
	var out mailbox.Message
	println("headers", res.Header.Get("Content-Length"))
	err = out.DecodeMsg(msgp.NewReader(res.Body))
	if err != nil {
		return nil, err
	}
	return &out, nil
}
