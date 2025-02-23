package records

import (
	"reflect"

	"github.com/andrebq/mixtape/prototypes/store"
)

type (
	Agent struct {
		_             struct{}       `ddl:"table=t_agents"`
		ID            string         `db:"id" ddl:"type=text,primary key"`
		Name          string         `db:"name" ddl:"not null"`
		Executors     store.JSONBlob `db:"executors" ddl:"not null"`
		MaxConcurrent int64          `db:"max_concurrent" ddl:"not null"`
		Active        bool           `db:"active" ddl:"not null"`
	}

	Task struct {
		_              struct{}       `ddl:"table=t_tasks"`
		AgentID        string         `db:"agent_id" ddl:"not null"`
		TimeoutSeconds int64          `db:"timeout" ddl:"not null"`
		Instructions   string         `db:"instructions" ddl:"not null"`
		Fields         store.JSONBlob `db:"fields"`
	}
)

func RegisterTypes() {
	store.MustRegister(reflect.TypeFor[Agent]())
	store.MustRegister(reflect.TypeFor[Task]())
}
