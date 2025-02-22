package store

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/andrebq/mixtape/generics"
)

var (
	typeMap = map[reflect.Type]string{
		reflect.TypeFor[int]():           "integer",
		reflect.TypeFor[int32]():         "integer",
		reflect.TypeFor[int64]():         "integer",
		reflect.TypeFor[float32]():       "real",
		reflect.TypeFor[float64]():       "real",
		reflect.TypeFor[string]():        "text",
		reflect.TypeFor[[]byte]():        "blob",
		reflect.TypeFor[time.Duration](): "integer",
		reflect.TypeFor[time.Time]():     "text",
		reflect.TypeFor[bool]():          "integer",
		reflect.TypeFor[JSONBlob]():      "text",
	}
)

func generateDMLStatements(md *mappingData, modelType reflect.Type) (insertStmt string, deleteStmt string, lookupStmt string, matchStmt func(pattern map[string]any) (string, map[string]any, error)) {
	var columns generics.Set[string]
	var placeholders []string
	var updates []string
	var primaryKey string
	tagByField := map[string]string{}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		dbTag := field.Tag.Get("db")
		ddlTag := field.Tag.Get("ddl")
		if dbTag == "" {
			continue
		}
		tagByField[field.Name] = dbTag

		columns.PutAll(fmt.Sprintf("%q", dbTag))
		placeholders = append(placeholders, ":"+dbTag)
		if strings.Contains(ddlTag, "primary key") {
			primaryKey = dbTag
		} else {
			updates = append(updates, fmt.Sprintf("%q = excluded.%q", dbTag, dbTag))
		}
	}
	slices.SortFunc(placeholders, strings.Compare)

	insertStmt = fmt.Sprintf(
		"INSERT INTO %q (%s) VALUES (%s) ON CONFLICT(%q) DO UPDATE SET %s;",
		md.tableName,
		strings.Join(columns.AppendToSorted(nil, strings.Compare), ", "),
		strings.Join(placeholders, ", "),
		primaryKey,
		strings.Join(updates, ", "),
	)

	deleteStmt = fmt.Sprintf("DELETE FROM %q WHERE %q = :%s;", md.tableName, primaryKey, primaryKey)

	lookupStmt = fmt.Sprintf("SELECT * FROM %q WHERE %q = :%s;", md.tableName, primaryKey, primaryKey)

	matchStmt = func(pattern map[string]any) (string, map[string]any, error) {
		var lookup []string
		newpattern := map[string]any{}
		for k, v := range pattern {
			if strings.HasPrefix(":", k) {
				k = k[:]
				newpattern[k] = v
			} else {
				k = tagByField[k]
				if k == "" {
					return "", nil, fmt.Errorf("field %v is not present in the struct", k)
				}
				newpattern[k] = v
			}
			quotedK := fmt.Sprintf("%q", k)
			if columns.Has(quotedK) {
				lookup = append(lookup, fmt.Sprintf("%s = :%s", quotedK, k))
			} else {
				return "", nil, fmt.Errorf("field %v is not present in the table", k)
			}
		}
		var orderBy string
		if md.defaultSort != "" {
			orderBy = fmt.Sprintf(" ORDER BY %s", md.defaultSort)
		}
		if len(lookup) == 0 {
			return fmt.Sprintf("SELECT * FROM %q%s", md.tableName, orderBy), newpattern, nil
		} else {
			return fmt.Sprintf("SELECT * FROM %q WHERE %v%s", md.tableName, strings.Join(lookup, " AND "), orderBy), newpattern, nil
		}
	}

	return
}

func goTypeToSQLiteType(goType reflect.Type, ddlTag string) string {
	if strings.Contains(ddlTag, "type=") {
		return parseTypeFromDDLTag(ddlTag)
	}
	if goType.Kind() == reflect.Pointer {
		goType = goType.Elem()
	}
	sqlType, found := typeMap[goType]
	if !found {
		panic(fmt.Sprintf("cannot generate mapping for Go type %v, use type=... to force a SQL type", goType))
	}
	return sqlType
}

func parseLookupTag(tag, lookup string) (string, bool) {
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, lookup+"=") {
			return strings.TrimPrefix(part, lookup+"="), true
		}
	}
	return "", false
}

func parseDDLTag(tag string) string {
	parts := strings.Split(tag, ",")
	var finalParts []string
	for _, part := range parts {
		if !strings.Contains(part, "=") {
			finalParts = append(finalParts, part)
		}
	}
	return strings.Join(finalParts, " ")
}

func parseTypeFromDDLTag(tag string) string {
	tp, found := parseLookupTag(tag, "type")
	if !found {
		tp = "TEXT"
	}
	return tp
}

func parseTableName(md *mappingData, typeInfo reflect.Type) {
	tableTag := ""
	for i := 0; i < typeInfo.NumField(); i++ {
		field := typeInfo.Field(i)
		ddlTag := field.Tag.Get("ddl")
		if field.Name != "_" {
			continue
		}
		tableTag = ddlTag
		break
	}
	tableName, _ := parseLookupTag(tableTag, "table")
	if tableName == "" {
		panic(fmt.Sprintf("go type %v does not have the table=... annotation", typeInfo))
	}
	md.tableName = tableName
	md.defaultSort, _ = parseLookupTag(tableTag, "default_sort")
}

func genDDL(mapping *mappingData, typeInfo reflect.Type) {
	mapping.alterStatements = make(map[string]string)
	parseTableName(mapping, typeInfo)
	var columnDefs []string

	for i := 0; i < typeInfo.NumField(); i++ {
		field := typeInfo.Field(i)
		dbTag := field.Tag.Get("db")
		ddlTag := field.Tag.Get("ddl")

		if dbTag == "" {
			continue
		}

		columnType := goTypeToSQLiteType(field.Type, ddlTag)
		columnDef := fmt.Sprintf("%q %s", dbTag, columnType)
		columnName := dbTag
		if ddlTag != "" {
			columnDef += " " + parseDDLTag(ddlTag)
		}

		mapping.alterStatements[columnName] = fmt.Sprintf("ALTER TABLE %q ADD COLUMN %s;", mapping.tableName, columnName)
		columnDefs = append(columnDefs, columnDef)
	}
	mapping.createStatement = fmt.Sprintf("CREATE TABLE %q (%s);", mapping.tableName, strings.Join(columnDefs, ", "))
}
