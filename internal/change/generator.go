package change

import (
	"fmt"
	"sort"
	"strings"

	"tablepro/internal/driver"
)

type SQLStatementGenerator struct {
	Dialect *Dialect
}

func NewSQLStatementGenerator(dbType driver.DatabaseType) *SQLStatementGenerator {
	return &SQLStatementGenerator{
		Dialect: GetDialect(dbType),
	}
}

func (g *SQLStatementGenerator) qualifiedTable(schemaName, tableName string) string {
	if schemaName == "" {
		return g.Dialect.QuoteIdentifier(tableName)
	}
	return g.Dialect.QuoteIdentifier(schemaName) + "." + g.Dialect.QuoteIdentifier(tableName)
}

func (g *SQLStatementGenerator) GenerateUpdate(change *CellChange, schemaName, tableName string) (string, []any, error) {
	if change == nil {
		return "", nil, fmt.Errorf("change is nil")
	}
	if len(change.PrimaryKey) == 0 {
		return "", nil, fmt.Errorf("no primary keys for UPDATE on table %s", tableName)
	}

	var b strings.Builder
	args := make([]any, 0, 1+len(change.PrimaryKey))
	paramIdx := 1

	b.WriteString("UPDATE ")
	b.WriteString(g.qualifiedTable(schemaName, tableName))
	b.WriteString(" SET ")
	b.WriteString(g.Dialect.QuoteIdentifier(change.Column))
	b.WriteString(" = ")
	b.WriteString(g.Dialect.ParamMarker(paramIdx))
	args = append(args, change.NewValue)
	paramIdx++

	b.WriteString(" WHERE ")
	pkKeys := sortedKeys(change.PrimaryKey)
	for i, pk := range pkKeys {
		if i > 0 {
			b.WriteString(" AND ")
		}
		b.WriteString(g.Dialect.QuoteIdentifier(pk))
		b.WriteString(" = ")
		b.WriteString(g.Dialect.ParamMarker(paramIdx))
		args = append(args, change.PrimaryKey[pk])
		paramIdx++
	}

	return b.String(), args, nil
}

func (g *SQLStatementGenerator) GenerateInsert(row *InsertedRow, schemaName, tableName string, columns []string, autoIncrementCols []string) (string, []any, error) {
	if row == nil {
		return "", nil, fmt.Errorf("row is nil")
	}

	autoIncSet := make(map[string]bool, len(autoIncrementCols))
	for _, col := range autoIncrementCols {
		autoIncSet[col] = true
	}

	filteredCols := make([]string, 0, len(columns))
	for _, col := range columns {
		if !autoIncSet[col] {
			filteredCols = append(filteredCols, col)
		}
	}

	if len(filteredCols) == 0 {
		return "", nil, fmt.Errorf("no columns to insert after excluding auto-increment columns")
	}

	var b strings.Builder
	args := make([]any, 0, len(filteredCols))
	paramIdx := 1

	b.WriteString("INSERT INTO ")
	b.WriteString(g.qualifiedTable(schemaName, tableName))
	b.WriteString(" (")

	for i, col := range filteredCols {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(g.Dialect.QuoteIdentifier(col))
	}

	b.WriteString(") VALUES (")

	for i, col := range filteredCols {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(g.Dialect.ParamMarker(paramIdx))
		paramIdx++
		args = append(args, row.Data[col])
	}

	b.WriteString(")")

	return b.String(), args, nil
}

func (g *SQLStatementGenerator) GenerateDelete(row *DeletedRow, schemaName, tableName string) (string, []any, error) {
	if row == nil {
		return "", nil, fmt.Errorf("row is nil")
	}
	if len(row.PrimaryKey) == 0 {
		return "", nil, fmt.Errorf("no primary keys for DELETE on table %s", tableName)
	}

	var b strings.Builder
	args := make([]any, 0, len(row.PrimaryKey))
	paramIdx := 1

	b.WriteString("DELETE FROM ")
	b.WriteString(g.qualifiedTable(schemaName, tableName))
	b.WriteString(" WHERE ")

	pkKeys := sortedKeys(row.PrimaryKey)
	for i, pk := range pkKeys {
		if i > 0 {
			b.WriteString(" AND ")
		}
		b.WriteString(g.Dialect.QuoteIdentifier(pk))
		b.WriteString(" = ")
		b.WriteString(g.Dialect.ParamMarker(paramIdx))
		args = append(args, row.PrimaryKey[pk])
		paramIdx++
	}

	return b.String(), args, nil
}

func (g *SQLStatementGenerator) GenerateAll(changes *TabChanges, autoIncrementCols []string) ([]Statement, error) {
	if changes == nil {
		return nil, fmt.Errorf("changes is nil")
	}
	if !changes.HasChanges() {
		return nil, nil
	}

	columns := inferColumns(changes)
	stmts := make([]Statement, 0, changes.ChangeCount())

	for i := range changes.DeletedRows {
		sql, args, err := g.GenerateDelete(changes.DeletedRows[i], changes.SchemaName, changes.TableName)
		if err != nil {
			return nil, fmt.Errorf("generating DELETE #%d: %w", i, err)
		}
		stmts = append(stmts, Statement{SQL: sql, Args: args, Type: "DELETE"})
	}

	for i := range changes.CellChanges {
		sql, args, err := g.GenerateUpdate(changes.CellChanges[i], changes.SchemaName, changes.TableName)
		if err != nil {
			return nil, fmt.Errorf("generating UPDATE #%d: %w", i, err)
		}
		stmts = append(stmts, Statement{SQL: sql, Args: args, Type: "UPDATE"})
	}

	for i := range changes.InsertedRows {
		sql, args, err := g.GenerateInsert(changes.InsertedRows[i], changes.SchemaName, changes.TableName, columns, autoIncrementCols)
		if err != nil {
			return nil, fmt.Errorf("generating INSERT #%d: %w", i, err)
		}
		stmts = append(stmts, Statement{SQL: sql, Args: args, Type: "INSERT"})
	}

	return stmts, nil
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func inferColumns(changes *TabChanges) []string {
	if len(changes.InsertedRows) == 0 {
		return nil
	}
	colSet := make(map[string]bool)
	for _, row := range changes.InsertedRows {
		for col := range row.Data {
			colSet[col] = true
		}
	}
	cols := make([]string, 0, len(colSet))
	for col := range colSet {
		cols = append(cols, col)
	}
	sort.Strings(cols)
	return cols
}
