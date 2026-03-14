package mssql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"tablepro/internal/driver"
)

// Connect establishes a connection to MSSQL database
func (d *MSSQLDriver) Connect(ctx context.Context, config *Config) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if config != nil {
		d.config = config
	}

	// Build DSN
	dsn := d.config.DSN()

	// Open connection
	db, err := sql.Open("mssql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	if d.config.MaxOpenConnections > 0 {
		db.SetMaxOpenConns(d.config.MaxOpenConnections)
	}
	if d.config.MaxIdleConnections > 0 {
		db.SetMaxIdleConns(d.config.MaxIdleConnections)
	}
	if d.config.MaxConnectionLife > 0 {
		db.SetConnMaxLifetime(d.config.MaxConnectionLife)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	d.db = db
	return nil
}

// Execute runs a query and returns results
func (d *MSSQLDriver) Execute(ctx context.Context, query string, params ...any) (*driver.Result, error) {
	d.mu.RLock()
	db := d.db
	d.mu.RUnlock()

	if db == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := db.ExecContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	lastInsertID, _ := result.LastInsertId()

	return &driver.Result{
		LastInsertID: lastInsertID,
		RowsAffected: rowsAffected,
	}, nil
}

// Query runs a query and returns rows
func (d *MSSQLDriver) Query(ctx context.Context, query string, params ...any) (*driver.Row, error) {
	d.mu.RLock()
	db := d.db
	d.mu.RUnlock()

	if db == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Scan first row
	if rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := &driver.Row{
			Data:        make(map[string]any),
			ColumnNames: columns,
		}
		for i, col := range columns {
			row.Data[col] = values[i]
		}

		return row, nil
	}

	return nil, sql.ErrNoRows
}

// QueryContext runs a query with custom timeout
func (d *MSSQLDriver) QueryContext(ctx context.Context, timeout time.Duration, query string, params ...any) (*driver.Row, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return d.Query(ctx, query, params...)
}

// GetSchema returns database schema information
func (d *MSSQLDriver) GetSchema(ctx context.Context) (*driver.SchemaInfo, error) {
	d.mu.RLock()
	db := d.db
	d.mu.RUnlock()

	if db == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	query := `
		SELECT name AS schema_name
		FROM sys.schemas
		WHERE name NOT IN ('sys', 'INFORMATION_SCHEMA', 'guest')
		ORDER BY name;
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get schemas: %w", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, schema)
	}

	return &driver.SchemaInfo{
		Schemas: schemas,
	}, nil
}

// GetTables returns all tables in the database
func (d *MSSQLDriver) GetTables(ctx context.Context, schemaName string) ([]driver.TableInfo, error) {
	d.mu.RLock()
	db := d.db
	d.mu.RUnlock()

	if db == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Default to dbo schema if not specified
	if schemaName == "" {
		schemaName = "dbo"
	}

	query := `
		SELECT 
			t.TABLE_SCHEMA,
			t.TABLE_NAME,
			t.TABLE_TYPE,
			(SELECT SUM(p.rows) FROM sys.partitions p 
			 INNER JOIN sys.tables st ON p.object_id = st.object_id 
			 WHERE st.name = t.TABLE_NAME AND p.index_id IN (0, 1)) AS row_count
		FROM INFORMATION_SCHEMA.TABLES t
		WHERE t.TABLE_SCHEMA = @p1
		  AND t.TABLE_TYPE IN ('BASE TABLE', 'VIEW')
		ORDER BY t.TABLE_NAME;
	`

	rows, err := db.QueryContext(ctx, query, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	defer rows.Close()

	var tables []driver.TableInfo
	for rows.Next() {
		var t driver.TableInfo
		var tableType string
		if err := rows.Scan(&t.Schema, &t.Name, &tableType, &t.RowCount); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		t.Type = driver.TableType(strings.TrimPrefix(tableType, "BASE "))
		tables = append(tables, t)
	}

	return tables, nil
}

// GetColumns returns column information for a specific table
func (d *MSSQLDriver) GetColumns(ctx context.Context, schemaName, tableName string) ([]driver.ColumnInfo, error) {
	d.mu.RLock()
	db := d.db
	d.mu.RUnlock()

	if db == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Default to dbo schema if not specified
	if schemaName == "" {
		schemaName = "dbo"
	}

	query := `
		SELECT 
			c.COLUMN_NAME,
			c.DATA_TYPE,
			c.IS_NULLABLE,
			c.COLUMN_DEFAULT,
			c.CHARACTER_MAXIMUM_LENGTH,
			c.NUMERIC_SCALE,
			c.NUMERIC_PRECISION,
			CASE WHEN pk.COLUMN_NAME IS NOT NULL THEN 1 ELSE 0 END AS is_primary_key,
			CASE WHEN fk.COLUMN_NAME IS NOT NULL THEN 1 ELSE 0 END AS is_foreign_key
		FROM INFORMATION_SCHEMA.COLUMNS c
		LEFT JOIN (
			SELECT ku.TABLE_SCHEMA, ku.TABLE_NAME, ku.COLUMN_NAME
			FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
			JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE ku ON tc.CONSTRAINT_NAME = ku.CONSTRAINT_NAME
			WHERE tc.CONSTRAINT_TYPE = 'PRIMARY KEY'
		) pk ON c.TABLE_SCHEMA = pk.TABLE_SCHEMA AND c.TABLE_NAME = pk.TABLE_NAME AND c.COLUMN_NAME = pk.COLUMN_NAME
		LEFT JOIN (
			SELECT 
				kcu.TABLE_SCHEMA,
				kcu.TABLE_NAME,
				kcu.COLUMN_NAME
			FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
			JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu ON tc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME
			WHERE tc.CONSTRAINT_TYPE = 'FOREIGN KEY'
		) fk ON c.TABLE_SCHEMA = fk.TABLE_SCHEMA AND c.TABLE_NAME = fk.TABLE_NAME AND c.COLUMN_NAME = fk.COLUMN_NAME
		WHERE c.TABLE_SCHEMA = @p1 AND c.TABLE_NAME = @p2
		ORDER BY c.ORDINAL_POSITION;
	`

	rows, err := db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	defer rows.Close()

	var columns []driver.ColumnInfo
	for rows.Next() {
		var c driver.ColumnInfo
		var defaultValue *string
		var maxLength, numericScale, numericPrecision *int
		var isPrimaryKey, isForeignKey int
		var isNullable string

		if err := rows.Scan(
			&c.Name, &c.DataType, &isNullable, &defaultValue,
			&maxLength, &numericScale, &numericPrecision,
			&isPrimaryKey, &isForeignKey,
		); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		c.Nullable = isNullable == "YES"
		c.DefaultValue = defaultValue
		c.IsPrimaryKey = isPrimaryKey == 1
		if mapped, ok := TypeMapping[c.DataType]; ok {
			c.TypeName = mapped
		} else {
			c.TypeName = c.DataType
		}

		columns = append(columns, c)
	}

	return columns, nil
}

// GetIndexes returns index information for a specific table
func (d *MSSQLDriver) GetIndexes(ctx context.Context, schemaName, tableName string) ([]driver.IndexInfo, error) {
	d.mu.RLock()
	db := d.db
	d.mu.RUnlock()

	if db == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if schemaName == "" {
		schemaName = "dbo"
	}

	query := `
		SELECT 
			i.name AS index_name,
			STUFF((
				SELECT ', ' + c.name
				FROM sys.index_columns ic
				JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
				WHERE ic.object_id = i.object_id AND ic.is_included_column = 0
				ORDER BY ic.key_ordinal
				FOR XML PATH('')
			), 1, 2, '') AS columns,
			i.is_unique,
			i.is_primary_key,
			i.type_desc
		FROM sys.indexes i
		JOIN sys.tables t ON i.object_id = t.object_id
		JOIN sys.schemas s ON t.schema_id = s.schema_id
		WHERE s.name = @p1 AND t.name = @p2
		  AND i.is_primary_key = 0
		ORDER BY i.name;
	`

	rows, err := db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get indexes: %w", err)
	}
	defer rows.Close()

	var indexes []driver.IndexInfo
	for rows.Next() {
		var idx driver.IndexInfo
		var columnsStr string
		if err := rows.Scan(&idx.Name, &columnsStr, &idx.IsUnique, &idx.IsPrimary, &idx.IndexType); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}
		idx.Columns = strings.Split(columnsStr, ", ")
		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// GetForeignKeys returns foreign key information for a specific table
func (d *MSSQLDriver) GetForeignKeys(ctx context.Context, schemaName, tableName string) ([]driver.ForeignKeyInfo, error) {
	d.mu.RLock()
	db := d.db
	d.mu.RUnlock()

	if db == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if schemaName == "" {
		schemaName = "dbo"
	}

	query := `
		SELECT 
			fk.name AS constraint_name,
			STUFF((
				SELECT ', ' + c.name
				FROM sys.foreign_key_columns fkc2
				JOIN sys.columns c ON fkc2.parent_column_id = c.column_id AND fkc2.parent_object_id = c.object_id
				WHERE fkc2.constraint_object_id = fk.object_id
				ORDER BY fkc2.constraint_column_id
				FOR XML PATH('')
			), 1, 2, '') AS columns,
			STUFF((
				SELECT ', ' + cr.name
				FROM sys.foreign_key_columns fkc2
				JOIN sys.columns cr ON fkc2.referenced_column_id = cr.column_id AND fkc2.referenced_object_id = cr.object_id
				WHERE fkc2.constraint_object_id = fk.object_id
				ORDER BY fkc2.constraint_column_id
				FOR XML PATH('')
			), 1, 2, '') AS referenced_columns,
			tr.name AS referenced_table,
			fk.update_referential_action_desc AS on_update,
			fk.delete_referential_action_desc AS on_delete
		FROM sys.foreign_keys fk
		JOIN sys.tables t ON fk.parent_object_id = t.object_id
		JOIN sys.schemas ts ON t.schema_id = ts.schema_id
		JOIN sys.tables tr ON fk.referenced_object_id = tr.object_id
		WHERE ts.name = @p1 AND t.name = @p2
		GROUP BY fk.name, tr.name, fk.update_referential_action_desc, fk.delete_referential_action_desc
		ORDER BY fk.name;
	`

	rows, err := db.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get foreign keys: %w", err)
	}
	defer rows.Close()

	var fks []driver.ForeignKeyInfo
	for rows.Next() {
		var fk driver.ForeignKeyInfo
		var columnsStr, refColumnsStr string
		if err := rows.Scan(
			&fk.Name, &columnsStr, &refColumnsStr, &fk.ReferencedTable,
			&fk.OnDelete, &fk.OnUpdate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key: %w", err)
		}
		fk.Columns = strings.Split(columnsStr, ", ")
		fk.ReferencedColumns = strings.Split(refColumnsStr, ", ")
		fks = append(fks, fk)
	}

	return fks, nil
}

// Ping checks if the connection is alive
func (d *MSSQLDriver) Ping(ctx context.Context) error {
	d.mu.RLock()
	db := d.db
	d.mu.RUnlock()

	if db == nil {
		return fmt.Errorf("not connected")
	}
	return db.PingContext(ctx)
}

// Close closes the database connection
func (d *MSSQLDriver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db != nil {
		d.db.Close()
		d.db = nil
	}
	return nil
}

// GetCapabilities returns the driver capabilities
func (d *MSSQLDriver) GetCapabilities() *driver.DriverCapabilities {
	return &driver.DriverCapabilities{
		Features: []driver.Feature{
			driver.FeatureSSLConnection,
			driver.FeatureSSHConnection,
			driver.FeaturePreparedStatements,
			driver.FeatureBatchStatements,
			driver.FeatureCursorPagination,
			driver.FeatureJSONType,
			driver.FeatureArrayType,
			driver.FeatureUUIDType,
		},
		MaxConnections:            10,
		SupportsTransactions:      true,
		SupportsStoredProcedures:  true,
		SupportsFunctions:         true,
		SupportsViews:             true,
		SupportsMaterializedViews: false,
		SupportsForeignKeys:       true,
		SupportsIndexes:           true,
		SupportsAutoIncrement:     true,
		SupportsSchemas:           true,
	}
}

// GetDB returns the underlying *sql.DB for advanced operations
func (d *MSSQLDriver) GetDB() *sql.DB {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.db
}

// Type returns the DatabaseType for this driver.
func (d *MSSQLDriver) Type() driver.DatabaseType {
	return driver.DatabaseTypeMSSQL
}
