package db

import (
	"strings"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type ColumnDefinition struct {
	Name string
	Type string
}

func (c ColumnDefinition) String() string {
	return c.Name + " " + c.Type
}

func CreateTableIfNotExists(conn *sqlite.Conn, tableName string, cols []ColumnDefinition) error {
	var columnDef = make([]string, len(cols))
	for i, col := range cols {
		columnDef[i] = col.String()
	}

	columns := strings.Join(columnDef, ",")

	query := `CREATE TABLE IF NOT EXISTS ` + tableName + ` (` + columns + `);`

	return sqlitex.ExecuteTransient(conn, query, nil)
}
