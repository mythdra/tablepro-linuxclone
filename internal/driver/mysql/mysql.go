package mysql

import "tablepro/internal/driver"

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"tablepro/internal/connection"
)

func New() *MySQLDriver {
	return &MySQLDriver{}
}

func (d *MySQLDriver) Connect(ctx context.Context, config *connection.DatabaseConnection, password string) error {
	dsn := buildDSN(config, password)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping MySQL server: %w", err)
	}

	d.db = db
	d.config = config

	if err := d.detectVersion(ctx); err != nil {
		return fmt.Errorf("failed to detect MySQL version: %w", err)
	}

	return nil
}

func buildDSN(config *connection.DatabaseConnection, password string) string {
	cfg := mysql.Config{
		User:                 config.Username,
		Passwd:               password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", config.Host, config.Port),
		DBName:               config.Database,
		AllowNativePasswords: true,
		ParseTime:            true,
		Loc:                  time.UTC,
	}

	cfg.Params = map[string]string{
		"charset": "utf8mb4",
	}

	if config.SSL.Enabled {
		switch config.SSL.Mode {
		case "require":
			cfg.TLSConfig = "skip-verify"
		case "verify-ca", "verify-full":
			cfg.TLSConfig = "verify-full"
		}
	}

	return cfg.FormatDSN()
}

func (d *MySQLDriver) detectVersion(ctx context.Context) error {
	var version string
	err := d.db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version)
	if err != nil {
		return err
	}

	d.version = &MySQLVersion{Version: version}

	if strings.Contains(version, "MariaDB") {
		d.version.IsMariaDB = true
		_, err := fmt.Sscanf(version, "MariaDB-%d.%d.%d", &d.version.Major, &d.version.Minor, &d.version.Patch)
		if err != nil {
			parts := strings.Split(version, "-")
			if len(parts) > 1 {
				_, _ = fmt.Sscanf(parts[1], "%d.%d.%d", &d.version.Major, &d.version.Minor, &d.version.Patch)
			}
		}
	} else {
		d.version.IsMariaDB = false
		_, _ = fmt.Sscanf(version, "%d.%d.%d", &d.version.Major, &d.version.Minor, &d.version.Patch)
	}

	return nil
}

func (d *MySQLDriver) Execute(ctx context.Context, query string) (*queryResult, error) {
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	result := &queryResult{
		Columns: columns,
		Rows:    make([][]any, 0),
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("failed to get column types: %w", err)
	}

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make([]any, len(columns))
		for i, colType := range columnTypes {
			val := values[i]
			row[i] = d.convertValue(val, colType.DatabaseTypeName())
		}

		result.Rows = append(result.Rows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}

func (d *MySQLDriver) ExecuteWithParams(ctx context.Context, query string, params []any) (*queryResult, error) {
	rows, err := d.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	result := &queryResult{
		Columns: columns,
		Rows:    make([][]any, 0),
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("failed to get column types: %w", err)
	}

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make([]any, len(columns))
		for i, colType := range columnTypes {
			val := values[i]
			row[i] = d.convertValue(val, colType.DatabaseTypeName())
		}

		result.Rows = append(result.Rows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}

func (d *MySQLDriver) convertValue(val any, dbType string) any {
	if val == nil {
		return nil
	}

	lowerType := strings.ToLower(dbType)

	switch lowerType {
	case "enum":
		if bytes, ok := val.([]byte); ok {
			return string(bytes)
		}
	case "set":
		if bytes, ok := val.([]byte); ok {
			return string(bytes)
		}
	case "tinyint", "smallint", "int", "integer", "bigint":
		if b, ok := val.([]byte); ok {
			var i int64
			_, _ = fmt.Sscanf(string(b), "%d", &i)
			return i
		}
	case "float", "double", "decimal":
		if b, ok := val.([]byte); ok {
			var f float64
			_, _ = fmt.Sscanf(string(b), "%f", &f)
			return f
		}
	case "blob", "tinyblob", "mediumblob", "longblob":
		if b, ok := val.([]byte); ok {
			return b
		}
	case "binary", "varbinary":
		if b, ok := val.([]byte); ok {
			return b
		}
	case "json":
		if b, ok := val.([]byte); ok {
			return string(b)
		}
	}

	return val
}

func (d *MySQLDriver) ExecuteNonQuery(ctx context.Context, query string) (int64, error) {
	result, err := d.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("non-query execution failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return rowsAffected, nil
}

func (d *MySQLDriver) Ping(ctx context.Context) error {
	if d.db == nil {
		return fmt.Errorf("not connected")
	}
	return d.db.PingContext(ctx)
}

func (d *MySQLDriver) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

func (d *MySQLDriver) BeginTransaction(ctx context.Context) error {
	if d.transaction != nil {
		return fmt.Errorf("transaction already in progress")
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	d.transaction = tx
	return nil
}

func (d *MySQLDriver) CommitTransaction() error {
	if d.transaction == nil {
		return fmt.Errorf("no transaction in progress")
	}

	err := d.transaction.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	d.transaction = nil
	return nil
}

func (d *MySQLDriver) RollbackTransaction() error {
	if d.transaction == nil {
		return fmt.Errorf("no transaction in progress")
	}

	err := d.transaction.Rollback()
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	d.transaction = nil
	return nil
}

func (d *MySQLDriver) IsMariaDB() bool {
	if d.version == nil {
		return false
	}
	return d.version.IsMariaDB
}

func (d *MySQLDriver) GetVersion() *MySQLVersion {
	return d.version
}

func (d *MySQLDriver) GetConfig() *connection.DatabaseConnection {
	return d.config
}

func (d *MySQLDriver) IsConnected() bool {
	if d.db == nil {
		return false
	}
	return true
}

func (d *MySQLDriver) GetDatabase() string {
	if d.config != nil {
		return d.config.Database
	}
	return ""
}

// Type returns the DatabaseType for this driver.
func (d *MySQLDriver) Type() driver.DatabaseType {
	return driver.DatabaseTypeMySQL
}
