package sqlite

import "tablepro/internal/driver"

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"tablepro/internal/connection"
)

const DriverName = "sqlite3"

type SQLiteDriver struct {
	db *sql.DB
}

func New() *SQLiteDriver {
	return &SQLiteDriver{}
}

func (d *SQLiteDriver) Connect(ctx context.Context, cfg connection.DatabaseConnection) error {
	dsn := cfg.Database
	if cfg.LocalFile != "" {
		dsn = cfg.LocalFile
	}

	if dsn == "" {
		dsn = ":memory:"
	}

	db, err := sql.Open(DriverName, dsn)
	if err != nil {
		return fmt.Errorf("failed to open SQLite database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	d.db = db
	return nil
}

func (d *SQLiteDriver) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

func (d *SQLiteDriver) Execute(ctx context.Context, query string) ([]map[string]any, error) {
	if d.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	trimmed := strings.TrimLeft(query, " \t\n\r")
	if len(trimmed) < 6 {
		return d.executeStatement(ctx, query)
	}

	prefix := strings.ToUpper(trimmed[:min(6, len(trimmed))])
	if prefix == "SELECT" || prefix == "PRAGMA" || prefix == "EXPLAI" {
		return d.executeQuery(ctx, query)
	}

	return d.executeStatement(ctx, query)
}

func (d *SQLiteDriver) executeQuery(ctx context.Context, query string) ([]map[string]any, error) {
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

func (d *SQLiteDriver) executeStatement(ctx context.Context, query string) ([]map[string]any, error) {
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

func (d *SQLiteDriver) GetSchema(ctx context.Context) (map[string]any, error) {
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

func (d *SQLiteDriver) GetTables(ctx context.Context) ([]map[string]any, error) {
	if d.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	rows, err := d.db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []map[string]any
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}

		tables = append(tables, map[string]any{
			"name":      name,
			"type":      "table",
			"row_count": nil,
		})
	}

	return tables, rows.Err()
}

func (d *SQLiteDriver) GetColumns(ctx context.Context, tableName string) ([]map[string]any, error) {
	if d.db == nil {
		return nil, fmt.Errorf("not connected")
	}

	query := fmt.Sprintf("PRAGMA table_info('%s')", tableName)

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
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
			val := values[i]
			if val == nil {
				row[col] = nil
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	// Convert to expected format
	var columnsResult []map[string]any
	for _, r := range results {
		columnsResult = append(columnsResult, map[string]any{
			"cid":           r["cid"],
			"name":          r["name"],
			"type":          r["type"],
			"not_null":      r["notnull"].(int64) == 1,
			"default_value": r["dflt_value"],
			"pk":            r["pk"].(int64) == 1,
		})
	}

	return columnsResult, rows.Err()
}

func (d *SQLiteDriver) Ping(ctx context.Context) error {
	if d.db == nil {
		return fmt.Errorf("not connected")
	}
	return d.db.PingContext(ctx)
}

func (d *SQLiteDriver) GetDB() *sql.DB {
	return d.db
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Type returns the DatabaseType for this driver.
func (d *SQLiteDriver) Type() driver.DatabaseType {
	return driver.DatabaseTypeSQLite
}
