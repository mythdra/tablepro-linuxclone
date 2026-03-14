package change

import (
	"fmt"
	"strings"

	"tablepro/internal/driver"
)

type SQLStatementGenerator struct {
	dialect *Dialect
}

func NewSQLStatementGenerator(dbType driver.DatabaseType) *SQLStatementGenerator {
	return &SQLStatementGenerator{
		dialect: GetDialect(dbType),
	}
}

func (g *SQLStatementGenerator) GenerateUpdate(tabChanges *TabChanges, rowIndex int, primaryKeyValues map[string]any) ([]Statement, error) {
	changes, exists := g.getChangesForRow(tabChanges, rowIndex)
	if !exists || len(changes) == 0 {
		return nil, nil
	}

	var setClauses []string
	var params []any
	paramIdx := 1

	for _, ch := range changes {
		setClauses = append(setClauses, fmt.Sprintf("%s = %s",
			g.dialect.QuoteIdentifier(ch.ColumnName),
			g.dialect.ParamMarker(paramIdx)))
		params = append(params, ch.NewValue)
		paramIdx++
	}

	whereClause, _ := g.dialect.BuildWhereClause(tabChanges.PrimaryKeys, paramIdx)
	for _, pk := range tabChanges.PrimaryKeys {
		params = append(params, primaryKeyValues[pk])
	}

	tableName := g.dialect.QualifiedTableName(tabChanges.SchemaName, tabChanges.TableName)

	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		tableName,
		strings.Join(setClauses, ", "),
		whereClause)

	return []Statement{{SQL: sql, Params: params, Action: ChangeActionUpdate}}, nil
}

func (g *SQLStatementGenerator) GenerateInsert(tabChanges *TabChanges, insertedRow InsertedRow) (Statement, error) {
	var columnNames []string
	var placeholders []string
	var params []any
	paramIdx := 1

	for _, col := range tabChanges.Columns {
		if col.IsAutoIncrement {
			continue
		}
		val, exists := insertedRow.Values[col.Name]
		if !exists {
			continue
		}
		columnNames = append(columnNames, g.dialect.QuoteIdentifier(col.Name))
		placeholders = append(placeholders, g.dialect.ParamMarker(paramIdx))
		params = append(params, val)
		paramIdx++
	}

	if len(columnNames) == 0 {
		return Statement{}, fmt.Errorf("no columns to insert")
	}

	tableName := g.dialect.QualifiedTableName(tabChanges.SchemaName, tabChanges.TableName)

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columnNames, ", "),
		strings.Join(placeholders, ", "))

	return Statement{SQL: sql, Params: params, Action: ChangeActionInsert}, nil
}

func (g *SQLStatementGenerator) GenerateDelete(tabChanges *TabChanges, deletedRow DeletedRow) (Statement, error) {
	if len(tabChanges.PrimaryKeys) == 0 {
		return Statement{}, fmt.Errorf("cannot generate DELETE without primary keys")
	}

	var params []any
	whereClause, _ := g.dialect.BuildWhereClause(tabChanges.PrimaryKeys, 1)
	for _, pk := range tabChanges.PrimaryKeys {
		params = append(params, deletedRow.PrimaryKeys[pk])
	}

	tableName := g.dialect.QualifiedTableName(tabChanges.SchemaName, tabChanges.TableName)

	sql := fmt.Sprintf("DELETE FROM %s WHERE %s", tableName, whereClause)

	return Statement{SQL: sql, Params: params, Action: ChangeActionDelete}, nil
}

func (g *SQLStatementGenerator) GenerateAll(tabChanges *TabChanges, rowPrimaryKeys map[int]map[string]any) ([]Statement, error) {
	var statements []Statement

	for _, deletedRow := range tabChanges.DeletedRows {
		stmt, err := g.GenerateDelete(tabChanges, deletedRow)
		if err != nil {
			return nil, fmt.Errorf("generate delete failed: %w", err)
		}
		statements = append(statements, stmt)
	}

	processedRows := make(map[int]bool)
	for key := range tabChanges.CellChanges {
		parts := strings.SplitN(key, ":", 2)
		if len(parts) != 2 {
			continue
		}
		rowIdx := 0
		fmt.Sscanf(parts[0], "%d", &rowIdx)
		if processedRows[rowIdx] {
			continue
		}
		processedRows[rowIdx] = true

		pkValues, ok := rowPrimaryKeys[rowIdx]
		if !ok {
			return nil, fmt.Errorf("missing primary key values for row %d", rowIdx)
		}

		stmts, err := g.GenerateUpdate(tabChanges, rowIdx, pkValues)
		if err != nil {
			return nil, fmt.Errorf("generate update failed: %w", err)
		}
		statements = append(statements, stmts...)
	}

	for _, insertedRow := range tabChanges.InsertedRows {
		stmt, err := g.GenerateInsert(tabChanges, insertedRow)
		if err != nil {
			return nil, fmt.Errorf("generate insert failed: %w", err)
		}
		statements = append(statements, stmt)
	}

	return statements, nil
}

func (g *SQLStatementGenerator) getChangesForRow(tabChanges *TabChanges, rowIndex int) ([]CellChange, bool) {
	var result []CellChange
	prefix := fmt.Sprintf("%d:", rowIndex)
	found := false
	for key, changes := range tabChanges.CellChanges {
		if strings.HasPrefix(key, prefix) {
			found = true
			result = append(result, changes...)
		}
	}
	return result, found
}
