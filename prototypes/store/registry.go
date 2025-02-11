package store

import (
	"reflect"

	"github.com/andrebq/mixtape/generics"
)

type (
	mappingData struct {
		tableName       string
		alterStatements map[string]string
		createStatement string

		insert, delete, lookup string
	}
)

var (
	registry = generics.SyncMap[reflect.Type, *mappingData]{}
)

// MustRegister takes a given type and adds the mapping from Go to SQL
// into the global registry.
//
// It is safe to register the same type multiple times.
//
// When not specified in tags, table names are derived from the name of the Go type
// without considering package information. Therefore it is mandatory to specify table
// name using struct tags.
//
// Eg.:
//
//	type TaskV2 struct {
//		_              struct{}      `ddl:"table=tasks"`
//		ID             string        `db:"id" ddl:"primary key"`
//		Script         string        `db:"script" ddl:"not null"`
//		UserParameters string        `db:"user_parameters" ddl:"type=blob"`
//		TTL            time.Duration `db:"ttl" ddl:"not null"`
//		Completed      bool          `db:"completed"`
//		NewField       string        `db:"new_field"`
//	}
func MustRegister(tp reflect.Type) {
	md := mappingData{}
	genDDL(&md, tp)
	md.insert, md.delete, md.lookup = generateDMLStatements(md.tableName, tp)
	registry.Put(tp, &md)
}
