package ssh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/andrebq/mixtape/prototypes/store"
	"github.com/gliderlabs/ssh"
	"github.com/jmoiron/sqlx"
	gossh "golang.org/x/crypto/ssh"
)

type (
	DynKDB struct {
		conn *sqlx.DB
	}

	KeyConfig struct {
		ID           string                    `db:"id" ddl:"primary key"`
		ExpiresAt    time.Time                 `db:"expires_at"`
		ValidFrom    time.Time                 `db:"valid_from"`
		AllowedHosts store.TypedJSON[[]string] `db:"allowed_hosts"`
		Description  string                    `db:"description"`
	}

	KeyPermissions struct {
		ID      string                        `db:"id" ddl:"primary key"`
		Entries store.TypedJSON[[]Permission] `db:"entries"`
	}

	Permission struct {
		Operation string
		Resource  string
		Action    string
	}

	KeyRegistration struct {
		ID          string    `db:"id" ddl:"primary key"`
		PublicKey   SSHPubKey `json:"pubkey"`
		UseCases    []string  `json:"useCase"`
		Hosts       []string  `json:"hosts"`
		Description string    `json:"description"`
		Owner       string    `json:"owner"`
	}

	SSHPubKey struct {
		ssh.PublicKey
	}
)

func (s *SSHPubKey) MarshalJSON() ([]byte, error) {
	if s.PublicKey == nil {
		return nil, nil
	}
	return json.Marshal(string(gossh.MarshalAuthorizedKey(*s)))
}

func (s *SSHPubKey) UnmarshalJSON(buf []byte) error {
	s.PublicKey = nil
	var str string
	err := json.Unmarshal(buf, &str)
	if err != nil {
		return err
	}
	key, _, _, _, err := gossh.ParseAuthorizedKey([]byte(str))
	if err != nil {
		return err
	}
	s.PublicKey = key
	return nil
}

var (
	errNotAuthorized = errors.New("ssh: not authorized")
)

func (d *DynKDB) RegisterKey(ctx context.Context, key ssh.PublicKey, validFrom, expiresAt time.Time, allowedHosts []string) error {
	lookupKey := d.computeKeyLookup(key)
	cfg := KeyConfig{
		ID:           lookupKey,
		ExpiresAt:    expiresAt,
		ValidFrom:    validFrom,
		AllowedHosts: store.AsJSON(allowedHosts),
		Description:  string(gossh.MarshalAuthorizedKey(key)),
	}
	return store.Upsert(ctx, d.conn, cfg)
}

func (d *DynKDB) SetPermission(ctx context.Context, key ssh.PublicKey, operation, resource, action string) error {
	lookupKey := d.computeKeyPermissionLookup(key)

	permissions := KeyPermissions{
		ID: lookupKey,
	}
	// TODO: instead of d.conn, use a transaction...
	permissions, err := store.LookupOne(ctx, d.conn, permissions)
	if store.IsNotFound(err) {
		err = nil
	} else if err != nil {
		return err
	}
	foundIdx := -1
	for i, perm := range permissions.Entries.V {
		if perm.Operation == operation && perm.Resource == resource {
			if action == "deny" {
				permissions.Entries.V[i] = Permission{}
				foundIdx = i
				break
			}
		}
	}
	if foundIdx == -1 && action != "deny" {
		permissions.Entries.V = append(permissions.Entries.V, Permission{
			Action:    action,
			Operation: operation,
			Resource:  resource,
		})
	} else if foundIdx != -1 {
		// TODO: excessive use of Entries.V shows I was converting this in a rush...
		switch {
		case foundIdx == len(permissions.Entries.V)-1:
			permissions.Entries.V = permissions.Entries.V[:foundIdx-1]
		case foundIdx == 0:
			permissions.Entries.V = permissions.Entries.V[1:]
		default:
			permissions.Entries.V[foundIdx] = permissions.Entries.V[len(permissions.Entries.V)-1]
			permissions.Entries.V = permissions.Entries.V[:len(permissions.Entries.V)-1]
		}
		sort.Slice(permissions.Entries, func(i, j int) bool {
			pi, pj := permissions.Entries.V[i], permissions.Entries.V[j]
			return pi.Resource < pj.Resource &&
				pi.Operation < pj.Operation &&
				pi.Action < pj.Action
		})
	}
	return store.Upsert(ctx, d.conn, permissions)
}

func (d *DynKDB) AuthN(ctx context.Context, key ssh.PublicKey) error {
	_, err := d.lookupAndVerifyConfig(ctx, key)
	return err
}

func (d *DynKDB) AuthZ(ctx context.Context, key ssh.PublicKey, operation, resource string) error {
	if operation != "expose-endpoint" {
		// for now, this is the only action, other than admin, that can run
		// and admin sessions are authorized by a different key
		return errNotAuthorized
	}
	cfg, err := d.lookupAndVerifyConfig(ctx, key)
	if err != nil {
		return err
	}
	for _, s := range cfg.AllowedHosts.V {
		if s == resource {
			return nil
		}
	}
	return errNotAuthorized
}

func (d *DynKDB) RequestKeyRegistration(ctx context.Context, key KeyRegistration) (KeyRegistration, error) {
	key.ID = ""
	regLookup := d.computeKeyLookupRegistration(key.PublicKey)

	oldreg := KeyRegistration{
		ID: regLookup,
	}

	oldreg, err := store.LookupOne(ctx, d.conn, oldreg)
	if err == nil {
		oldLookup := d.computeKeyLookupRegistration(oldreg.PublicKey.PublicKey)
		if oldLookup != regLookup {
			slog.Error("Key registration on database does not match its body", "lookupKey", regLookup, "bodyKey", oldLookup)
			return key, errors.New("unexpected state")
		}
	}
	key.ID = regLookup
	return key, store.Upsert(ctx, d.conn, key)
}

func (d *DynKDB) lookupAndVerifyConfig(ctx context.Context, key ssh.PublicKey) (KeyConfig, error) {
	lookupKey := d.computeKeyLookup(key)
	ops := d.Store.Ops(false)
	defer ops.Close()
	val := ops.KV().GetBytes(ctx, nil, lookupKey)
	if val == nil {
		return KeyConfig{}, errNotAuthorized
	}
	var cfg KeyConfig
	if err := json.Unmarshal(val, &cfg); err != nil {
		slog.Error("Invalid key-config from database", "lookup", lookupKey)
		return KeyConfig{}, errNotAuthorized
	}
	now := time.Now()
	if cfg.ValidFrom.After(now) || cfg.ExpiresAt.Before(now) {
		return KeyConfig{}, errNotAuthorized
	}
	return cfg, nil
}

func (d *DynKDB) computeKeyLookup(key ssh.PublicKey) string {
	sig := gossh.FingerprintSHA256(key)
	return fmt.Sprintf("kdb:key:%v", sig)
}

func (d *DynKDB) computeKeyPermissionLookup(key ssh.PublicKey) string {
	return fmt.Sprintf("kdb:key-perm:%v", gossh.FingerprintSHA256(key))
}

func (d *DynKDB) computeKeyLookupRegistration(key ssh.PublicKey) string {
	return fmt.Sprintf("kdb:key-reg:%v", gossh.FingerprintSHA256(key))
}
