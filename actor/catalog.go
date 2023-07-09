package actor

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/andrebq/mixtape/internal/configdb"
)

type (
	Registry struct {
		lock sync.Mutex
		conn *sql.DB

		config configdb.Locator

		version string

		classCache map[string]*ActorClass
	}
)

func NewRegistry(ctx context.Context, conn *sql.DB) (*Registry, error) {
	r := &Registry{conn: conn}
	if err := r.verifyAndMigrate(ctx); err != nil {
		return nil, fmt.Errorf("actor:registry could not load registry: %v", err)
	}
	return r, nil
}

func (r *Registry) Close() error {
	if r.conn == nil {
		return nil
	}
	err := r.conn.Close()
	r.conn = nil
	return err
}

func (r *Registry) ClearCache() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.classCache = map[string]*ActorClass{}
}

func (r *Registry) RegisterClass(ctx context.Context, name, code string) (*ActorClass, error) {
	clz, err := LoadClass(code)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	r.lock.Lock()
	defer r.lock.Unlock()

	_, err = r.conn.ExecContext(ctx, `insert into _mt_actor_class(name, code, last_update, revisions) values ($1, $2, $3, $4)
	on conflict do update set code = EXCLUDED.code, last_update = EXCLUDED.last_update, revisions = revisions + 1`, name, code, time.Now().Format(time.RFC3339), 1)
	if err != nil {
		return nil, err
	}

	if r.classCache == nil {
		r.classCache = map[string]*ActorClass{}
	}
	r.classCache[name] = clz
	return clz, nil
}

func (r *Registry) LoadClass(ctx context.Context, name string) (*ActorClass, error) {
	clz, _, err := r.loadClass(ctx, name)
	return clz, err
}

func (r *Registry) loadClass(ctx context.Context, name string) (*ActorClass, int64, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if r.classCache == nil {
		r.classCache = map[string]*ActorClass{}
	} else if clz, found := r.classCache[name]; found {
		return clz, 0, nil
	}

	var code string
	err := r.conn.QueryRowContext(ctx, "select code from _mt_actor_class where name = $1", name).Scan(&code)
	if err != nil {
		return nil, 0, err
	}
	clz, err := LoadClass(code)
	if err != nil {
		return nil, 0, err
	}
	r.classCache[name] = clz
	return clz, 0, nil
}

// RegisterActor if the actor does not exist.
//
// If the actor exists, then its class will change to clz
func (r *Registry) RegisterActor(ctx context.Context, actorUID, clz string) error {
	clzObj, actorClassID, err := r.loadClass(ctx, clz)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	r.lock.Lock()
	defer r.lock.Unlock()

	result, err := r.conn.ExecContext(ctx,
		`insert into _mt_actor_state(actor_uid, actor_class_id, state, last_update, revisions) values ($1, $2, $3, $4, $5) on conflict do nothing`,
		actorUID, actorClassID, "", time.Now().Format(time.RFC3339), 1)
	if err != nil {
		return err
	}
	newrows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if newrows == 0 {
		return nil
	}
	return r.initState(ctx, actorUID, clzObj)
}

func (r *Registry) initState(ctx context.Context, actorUID string, clz *ActorClass) error {
	state, revision, err := r.loadState(ctx, actorUID)
	if err != nil {
		return err
	}
	_, state, err = clz.HandleStateful(ctx, "initializeState", state, "{}")
	if err != nil {
		return err
	}
	_, err = r.saveState(ctx, actorUID, state, revision)
	return err
}

func (r *Registry) loadState(ctx context.Context, actorUID string) (string, int64, error) {
	var state string
	var revs int64
	err := r.conn.QueryRowContext(ctx, "select state, revisions from _mt_actor_state where actor_uid = $1", actorUID).Scan(&state, &revs)
	return state, revs, err
}

func (r *Registry) saveState(ctx context.Context, actorUID, state string, prevRev int64) (int64, error) {
	var newrev int64
	err := r.conn.QueryRowContext(ctx,
		"update _mt_actor_state set state = $1, revisions = revisions + 1, last_update = $2 where actor_uid = $3 and revisions = $4 returning revisions",
		state,
		time.Now().Format(time.RFC3339),
		actorUID,
		prevRev).Scan(&newrev)
	return newrev, err
}

func (r *Registry) verifyAndMigrate(ctx context.Context) error {
	// TODO: implemente proper migration in the future
	cmds := []string{
		`create table if not exists _mt_actor_class(id integer primary key, name text not null, code text not null, last_update text not null, revisions integer);`,
		`create table if not exists _mt_actor_state(
			id integer primary key,
			actor_uid text not null,
			actor_class_id integer,
			state text,
			last_update not null,
			revisions integer,
			foreign key (actor_class_id) references _mt_actor_class(id));`,
		`create unique index unq_actor_state on _mt_actor_state(actor_uid);`,
	}
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	if r.config, err = configdb.Migrate(ctx, tx, "_mt_actor_registry_config"); err != nil {
		return err
	}
	for _, c := range cmds {
		_, err := tx.ExecContext(ctx, c)
		if err != nil {
			return err
		}
	}

	if err := configdb.Put(ctx, tx, r.config, "version", "0.0.1"); err != nil {
		return err
	}

	i := configdb.Instance{Conn: tx, Locator: r.config}
	if !i.GetString(ctx, &r.version, "version", "") {
		if i.Err() != nil {
			return fmt.Errorf("actor:registry could not check version of database, caused by %w", i.Err())
		} else {
			return fmt.Errorf("actor:registry unexpected error, version not found")
		}
	} else if r.version != "0.0.1" {
		return fmt.Errorf("actor:registry version from database %v does not match expected value of %v", r.version, "0.0.1")
	}

	err = tx.Commit()
	tx = nil
	return err
}
