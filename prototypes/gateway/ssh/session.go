package ssh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

func (g *Gateway) sessionHandler(s ssh.Session) {
	fmt.Fprintf(s, "Successful authentication, but your credentials do not allow interactive access\n")
	s.Exit(0)
	s.Close()
}

func (g *Gateway) sessionHandleWhoami(s ssh.Session) {
	json.NewEncoder(s).Encode(struct {
		User        string    `json:"user"`
		Key         string    `json:"key"`
		Fingerprint string    `json:"fingerprint"`
		Now         time.Time `json:"now"`
	}{
		User:        s.User(),
		Key:         string(bytes.TrimSpace(gossh.MarshalAuthorizedKey(s.PublicKey()))),
		Fingerprint: gossh.FingerprintSHA256(s.PublicKey()),
		Now:         time.Now(),
	})
	s.Exit(0)
}
