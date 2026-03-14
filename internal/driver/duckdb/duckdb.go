package duckdb

import "tablepro/internal/driver"

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	"tablepro/internal/connection"
)

const DriverName = "duckdb"

type DuckDBDriver struct {
	db *sql.DB
}

func New() *DuckDBDriver {
	return &DuckDBDriver{}
}

func (d *DuckDBDriver) Connect(ctx context.Context, cfg connection.DatabaseConnection) error {
	dsn := cfg.Database
	if cfg.LocalFile != "" {
		dsn = cfg.LocalFile
	}

	if dsn == "" {
		dsn = ":memory:"
	}

	db, err := sql.Open(DriverName, dsn)
	if err != nil {
		return fmt.Errorf("failed to open DuckDB: %w", err)
	}

	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(10 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping DuckDB: %w", err)
	}

	d.db = db
	return nil
}

func (d *DuckDBDriver) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

func (d *DuckDBDriver) Execute(ctx context.Context, query string) ([]map[string]any, error) {
	if d.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	trimmed := strings.TrimLeft(query, " \t\n\r")
	if len(trimmed) < 6 {
		return d.executeStatement(ctx, query)
	}

	prefix := strings.ToUpper(trimmed[:min(6, len(trimmed))])
	if prefix == "SELECT" || prefix == "PRAGMA" || prefix == "EXPLAI" || prefix == "WITH" {
		return d.executeQuery(ctx, query)
	}

	return d.executeStatement(ctx, query)
}

func (d *DuckDBDriver) executeQuery(ctx context.Context, query string) ([]map[string]any, error) {
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]any)
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

func (d *DuckDBDriver) executeStatement(ctx context.Context, query string) ([]map[string]any, error) {
	result, err := d.db.ExecContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("statement execution failed: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	lastInsertID, _ := result.LastInsertId()

	return []map[string]any{
		{
			"rows_affected":  rowsAffected,
			"last_insert_id": lastInsertID,
		},
	}, nil
}

func (d *DuckDBDriver) GetSchema(ctx context.Context) (map[string]any, error) {
	if d.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	schema := make(map[string]any)
	schema["name"] = "main"
	schema["tables"] = make([]map[string]any, 0)

	tables, err := d.GetTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	schema["tables"] = tables

	return schema, nil
}

func (d *DuckDBDriver) GetTables(ctx context.Context) ([]map[string]any, error) {
	if d.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	query := `
		SELECT table_name, table_type
		FROM information_schema.tables
		WHERE table_schema = 'main'
		ORDER BY table_name
	`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []map[string]any
	for rows.Next() {
		var name, tableType string
		if err := rows.Scan(&name, &tableType); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}

		tables = append(tables, map[string]any{
			"name":      name,
			"type":      tableType,
			"row_count": nil,
		})
	}

	return tables, rows.Err()
}

func (d *DuckDBDriver) GetColumns(ctx context.Context, tableName string) ([]map[string]any, error) {
	if d.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	query := `
		SELECT 
			column_name, 
			data_type, 
			is_nullable, 
			column_default
		FROM information_schema.columns
		WHERE table_name = $1 AND table_schema = 'main'
		ORDER BY ordinal_position
	`

	rows, err := d.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []map[string]any
	for rows.Next() {
		var name, dataType, nullable string
		var defaultVal sql.NullString

		if err := rows.Scan(&name, &dataType, &nullable, &defaultVal); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		columns = append(columns, map[string]any{
			"name":          name,
			"type":          dataType,
			"not_null":      nullable == "NO",
			"default_value": defaultVal.String,
			"pk":            false,
		})
	}

	return columns, rows.Err()
}

func (d *DuckDBDriver) Ping(ctx context.Context) error {
	if d.db == nil {
		return fmt.Errorf("not connected")
	}
	return d.db.PingContext(ctx)
}

func (d *DuckDBDriver) GetDB() *sql.DB {
	return d.db
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Type returns the DatabaseType for this driver.
func (d *DuckDBDriver) Type() driver.DatabaseType {
	return driver.DatabaseTypeDuckDB
}
