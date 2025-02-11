package store

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	typeMap = map[reflect.Type]string{
		reflect.TypeFor[int]():           "integer",
		reflect.TypeFor[int64]():         "integer",
		reflect.TypeFor[float32]():       "real",
		reflect.TypeFor[float64]():       "real",
		reflect.TypeFor[string]():        "text",
		reflect.TypeFor[[]byte]():        "blob",
		reflect.TypeFor[time.Duration](): "integer",
		reflect.TypeFor[time.Time]():     "text",
		reflect.TypeFor[bool]():          "integer",
	}
)

func generateDMLStatements(tableName string, modelType reflect.Type) (insertStmt string, deleteStmt string, lookupStmt string) {
	var columns []string
	var placeholders []string
	var updates []string
	var primaryKey string

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		dbTag := field.Tag.Get("db")
		ddlTag := field.Tag.Get("ddl")
		if dbTag == "" {
			continue
		}
		columns = append(columns, fmt.Sprintf("%q", dbTag))
		placeholders = append(placeholders, ":"+dbTag)
		if strings.Contains(ddlTag, "primary key") {
			primaryKey = dbTag
		} else {
			updates = append(updates, fmt.Sprintf("%q = excluded.%q", dbTag, dbTag))
		}
	}

	insertStmt = fmt.Sprintf(
		"INSERT INTO %q (%s) VALUES (%s) ON CONFLICT(%q) DO UPDATE SET %s;",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
		primaryKey,
		strings.Join(updates, ", "),
	)

	deleteStmt = fmt.Sprintf("DELETE FROM %q WHERE %q = :%s;", tableName, primaryKey, primaryKey)

	lookupStmt = fmt.Sprintf("SELECT * FROM %q WHERE %q = :%s;", tableName, primaryKey, primaryKey)

	return insertStmt, deleteStmt, lookupStmt
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

func parseTableName(typeInfo reflect.Type) string {
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
	return tableName
}

func genDDL(mapping *mappingData, typeInfo reflect.Type) {
	mapping.alterStatements = make(map[string]string)
	mapping.tableName = parseTableName(typeInfo)
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
