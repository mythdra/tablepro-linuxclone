package clickhouse

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"time"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
	chdriver "github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"tablepro/internal/driver"
)

// Connect establishes a connection to ClickHouse database
func (d *ClickHouseDriver) Connect(ctx context.Context, config *Config) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if config != nil {
		d.config = config
	}

	// Build options
	options := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", d.config.Host, d.config.Port)},
		Auth: clickhouse.Auth{
			Database: d.config.Database,
			Username: d.config.Username,
			Password: d.config.Password,
		},
		Debug:       d.config.Debug,
		DialTimeout: d.config.DialTimeout,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	}

	if d.config.SSLMode != "" && d.config.SSLMode != "disable" {
		options.TLS = &tls.Config{}
	}

	// Open connection
	conn, err := clickhouse.Open(options)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	d.db = conn
	return nil
}

// Execute runs a query and returns results
func (d *ClickHouseDriver) Execute(ctx context.Context, query string, params ...any) (*driver.Result, error) {
	d.mu.RLock()
	conn := d.db
	d.mu.RUnlock()

	if conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := conn.Exec(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &driver.Result{
		LastInsertID: 0,
		RowsAffected: 0,
	}, nil
}

// Query runs a query and returns rows
func (d *ClickHouseDriver) Query(ctx context.Context, query string, params ...any) (*driver.Row, error) {
	d.mu.RLock()
	conn := d.db
	d.mu.RUnlock()

	if conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := conn.Query(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns := rows.Columns()

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
func (d *ClickHouseDriver) QueryContext(ctx context.Context, timeout time.Duration, query string, params ...any) (*driver.Row, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return d.Query(ctx, query, params...)
}

// GetSchema returns database schema information
func (d *ClickHouseDriver) GetSchema(ctx context.Context) (*driver.SchemaInfo, error) {
	d.mu.RLock()
	conn := d.db
	d.mu.RUnlock()

	if conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	query := `
		SELECT name
		FROM system.databases
		WHERE name NOT IN ('system', 'information_schema', 'performance_schema')
		ORDER BY name;
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get schemas: %w", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, name)
	}

	return &driver.SchemaInfo{
		Schemas: schemas,
	}, nil
}

// GetTables returns all tables in the database
func (d *ClickHouseDriver) GetTables(ctx context.Context, schemaName string) ([]driver.TableInfo, error) {
	d.mu.RLock()
	conn := d.db
	d.mu.RUnlock()

	if conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Default to current database if not specified
	if schemaName == "" {
		schemaName = d.config.Database
	}

	query := `
		SELECT 
			database,
			name,
			engine,
			total_rows AS row_count,
			data_compressed_bytes AS data_size,
			index_compressed_bytes AS index_size
		FROM system.tables
		WHERE database = ?
		  AND is_temporary = 0
		ORDER BY name;
	`

	rows, err := conn.Query(ctx, query, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	defer rows.Close()

	var tables []driver.TableInfo
	for rows.Next() {
		var t driver.TableInfo
		var engine string
		if err := rows.Scan(&t.Schema, &t.Name, &engine, &t.RowCount, &t.SizeBytes, &t.SizeBytes); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		t.Type = driver.TableTypeTable
		tables = append(tables, t)
	}

	return tables, nil
}

// GetColumns returns column information for a specific table
func (d *ClickHouseDriver) GetColumns(ctx context.Context, schemaName, tableName string) ([]driver.ColumnInfo, error) {
	d.mu.RLock()
	conn := d.db
	d.mu.RUnlock()

	if conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Default to current database if not specified
	if schemaName == "" {
		schemaName = d.config.Database
	}

	query := `
		SELECT 
			name,
			type,
			default_type,
			default_expression,
			comment,
			is_in_partition_key,
			is_in_sorting_key,
			is_in_primary_key
		FROM system.columns
		WHERE database = ? AND table = ?
		ORDER BY position;
	`

	rows, err := conn.Query(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	defer rows.Close()

	var columns []driver.ColumnInfo
	for rows.Next() {
		var c driver.ColumnInfo
		var defaultType, defaultExpr, comment *string
		var isInPartitionKey, isInSortingKey, isInPrimaryKey int

		if err := rows.Scan(
			&c.Name, &c.TypeName,
			&defaultType, &defaultExpr, &comment,
			&isInPartitionKey, &isInSortingKey, &isInPrimaryKey,
		); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		c.DefaultValue = defaultExpr
		c.Comment = comment
		c.IsPrimaryKey = isInPrimaryKey == 1

		// Map to Go type
		if mapped, ok := TypeMapping[c.TypeName]; ok {
			c.TypeName = mapped
		}

		columns = append(columns, c)
	}

	return columns, nil
}

// GetIndexes returns index information for a specific table
// ClickHouse uses skip indexes instead of traditional indexes
func (d *ClickHouseDriver) GetIndexes(ctx context.Context, schemaName, tableName string) ([]driver.IndexInfo, error) {
	d.mu.RLock()
	conn := d.db
	d.mu.RUnlock()

	if conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if schemaName == "" {
		schemaName = d.config.Database
	}

	query := `
		SELECT 
			name,
			expr,
			type,
			granularity
		FROM system.skip_indices
		WHERE database = ? AND table = ?
		ORDER BY name;
	`

	rows, err := conn.Query(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get indexes: %w", err)
	}
	defer rows.Close()

	var indexes []driver.IndexInfo
	for rows.Next() {
		var idx driver.IndexInfo
		var expr, idxType string
		var granularity int
		if err := rows.Scan(&idx.Name, &expr, &idxType, &granularity); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}
		idx.IndexType = idxType
		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// GetForeignKeys returns foreign key information
// ClickHouse doesn't support foreign keys in the traditional sense
func (d *ClickHouseDriver) GetForeignKeys(ctx context.Context, schemaName, tableName string) ([]driver.ForeignKeyInfo, error) {
	// ClickHouse doesn't support traditional foreign keys
	return []driver.ForeignKeyInfo{}, nil
}

// Ping checks if the connection is alive
func (d *ClickHouseDriver) Ping(ctx context.Context) error {
	d.mu.RLock()
	conn := d.db
	d.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}
	return conn.Ping(ctx)
}

// Close closes the database connection
func (d *ClickHouseDriver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db != nil {
		d.db.Close()
		d.db = nil
	}
	return nil
}

// GetCapabilities returns the driver capabilities
func (d *ClickHouseDriver) GetCapabilities() *driver.DriverCapabilities {
	return &driver.DriverCapabilities{
		Features: []driver.Feature{
			driver.FeatureSSLConnection,
			driver.FeatureJSONType,
			driver.FeatureArrayType,
			driver.FeatureUUIDType,
			driver.FeatureWindowFunctions,
			driver.FeatureCTE,
		},
		MaxConnections:            10,
		SupportsTransactions:      false,
		SupportsStoredProcedures:  false,
		SupportsFunctions:         true,
		SupportsViews:             true,
		SupportsMaterializedViews: true,
		SupportsForeignKeys:       false,
		SupportsIndexes:           true,
		SupportsAutoIncrement:     false,
		SupportsSchemas:           true,
	}
}

// GetDB returns the underlying ClickHouse connection for advanced operations
func (d *ClickHouseDriver) GetDB() chdriver.Conn {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.db
}

// Batch implements batch operations for ClickHouse
type Batch struct {
	conn  chdriver.Conn
	query string
	batch chdriver.Batch
}

// NewBatch creates a new batch for inserting data
func (d *ClickHouseDriver) NewBatch(ctx context.Context, query string) (*Batch, error) {
	d.mu.RLock()
	conn := d.db
	d.mu.RUnlock()

	if conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	batch, err := conn.PrepareBatch(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch: %w", err)
	}

	return &Batch{
		conn:  conn,
		query: query,
		batch: batch,
	}, nil
}

// Append adds a row to the batch
func (b *Batch) Append(values ...any) error {
	return b.batch.Append(values...)
}

// Send sends the batch to ClickHouse
func (b *Batch) Send() error {
	return b.batch.Send()
}

// Type returns the DatabaseType for this driver.
func (d *ClickHouseDriver) Type() driver.DatabaseType {
	return driver.DatabaseTypeClickHouse
}
