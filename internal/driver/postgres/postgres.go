package postgres

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"tablepro/internal/driver"
)

type PostgreSQLDriver struct {
	pool   *pgxpool.Pool
	config *Config
	mu     sync.RWMutex
}

func NewPostgreSQLDriver() *PostgreSQLDriver {
	return &PostgreSQLDriver{
		config: DefaultConfig(),
	}
}

func (d *PostgreSQLDriver) Connect(ctx context.Context, config *Config) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if config != nil {
		d.config = config
	}

	poolConfig, err := d.config.PoolConfig()
	if err != nil {
		return fmt.Errorf("failed to create pool config: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	d.pool = pool
	return nil
}

func (d *PostgreSQLDriver) Execute(ctx context.Context, query string) (*QueryResult, error) {
	d.mu.RLock()
	pool := d.pool
	d.mu.RUnlock()

	if pool == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	var resultRows []Row
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		row := make(Row)
		for i, col := range columns {
			row[col] = values[i]
		}
		resultRows = append(resultRows, row)
	}

	return &QueryResult{
		Columns:      columns,
		Rows:         resultRows,
		AffectedRows: 0,
	}, nil
}

func (d *PostgreSQLDriver) GetSchema() ([]SchemaInfo, error) {
	d.mu.RLock()
	pool := d.pool
	d.mu.RUnlock()

	if pool == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		SELECT n.nspname AS schema_name,
		       r.rolname AS owner,
		       (SELECT COUNT(*) FROM pg_tables t WHERE t.schemaname = n.nspname) AS table_count
		FROM pg_namespace n
		JOIN pg_roles r ON n.nspowner = r.rolname
		WHERE n.nspname NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		  AND n.nspname NOT LIKE 'pg_temp_%'
		  AND n.nspname NOT LIKE 'pg_toast_temp_%'
		ORDER BY n.nspname;
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get schemas: %w", err)
	}
	defer rows.Close()

	var schemas []SchemaInfo
	for rows.Next() {
		var s SchemaInfo
		if err := rows.Scan(&s.Name, &s.Owner, &s.TableCount); err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, s)
	}

	return schemas, nil
}

func (d *PostgreSQLDriver) GetTables(schema string) ([]TableInfo, error) {
	d.mu.RLock()
	pool := d.pool
	d.mu.RUnlock()

	if pool == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	query := `
		SELECT
			t.table_schema,
			t.table_name,
			t.table_type,
			COALESCE((SELECT reltuples::bigint FROM pg_class WHERE relname = t.table_name AND relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = t.table_schema)), 0) AS row_count,
			COALESCE((SELECT pg_total_relation_size(schemaname||'.'||tablename)::bigint FROM pg_tables WHERE schemaname = t.table_schema AND tablename = t.table_name), 0) AS size,
			COALESCE(obj_description((t.table_schema||'.'||t.table_name)::regclass), '') AS comment
		FROM information_schema.tables t
		WHERE t.table_schema = $1
		  AND t.table_type IN ('BASE TABLE', 'VIEW', 'MATERIALIZED VIEW')
		ORDER BY t.table_name;
	`

	rows, err := pool.Query(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var t TableInfo
		var tableType string
		if err := rows.Scan(&t.Schema, &t.Name, &tableType, &t.RowCount, &t.Size, &t.Comment); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		t.Type = strings.ToLower(tableType)
		tables = append(tables, t)
	}

	return tables, nil
}

func (d *PostgreSQLDriver) GetColumns(table string) ([]ColumnInfo, error) {
	d.mu.RLock()
	pool := d.pool
	d.mu.RUnlock()

	if pool == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	parts := strings.Split(table, ".")
	var schema, tableName string
	if len(parts) == 2 {
		schema = parts[0]
		tableName = parts[1]
	} else {
		schema = "public"
		tableName = table
	}

	query := `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			c.character_maximum_length,
			c.numeric_scale,
			c.numeric_precision,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END AS is_primary_key,
			CASE WHEN fk.column_name IS NOT NULL THEN true ELSE false END AS is_foreign_key,
			CASE WHEN c.is_unique = 'YES' THEN true ELSE false END AS is_unique,
			CASE WHEN c.column_default LIKE 'nextval%' THEN true ELSE false END AS is_auto_increment,
			coalesce(cc.column_comment, '') AS comment
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT ku.table_schema, ku.table_name, ku.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
			WHERE tc.constraint_type = 'PRIMARY KEY'
		) pk ON c.table_schema = pk.table_schema AND c.table_name = pk.table_name AND c.column_name = pk.column_name
		LEFT JOIN (
			SELECT
				kcu.table_schema,
				kcu.table_name,
				kcu.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
			WHERE tc.constraint_type = 'FOREIGN KEY'
		) fk ON c.table_schema = fk.table_schema AND c.table_name = fk.table_name AND c.column_name = fk.column_name
		LEFT JOIN (
			SELECT
				col.table_schema,
				col.table_name,
				col.column_name,
				col.column_description AS column_comment
			FROM information_schema.columns col
			JOIN pg_catalog.pg_class pc ON col.table_name = pc.relname
			JOIN pg_catalog.pg_namespace pn ON col.table_schema = pn.nspname AND pc.relnamespace = pn.oid
			JOIN pg_catalog.pg_attribute pa ON pa.attname = col.column_name AND pa.attrelid = pc.oid
			WHERE col.table_schema = $1 AND col.table_name = $2
		) cc ON c.table_schema = cc.table_schema AND c.table_name = cc.table_name AND c.column_name = cc.column_name
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position;
	`

	rows, err := pool.Query(ctx, query, schema, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var c ColumnInfo
		var defaultValue *string
		var maxLength, numericScale, numericPrecision *int
		if err := rows.Scan(
			&c.Name, &c.DataType, &c.IsNullable, &defaultValue,
			&maxLength, &numericScale, &numericPrecision,
			&c.IsPrimaryKey, &c.IsForeignKey, &c.IsUnique, &c.IsAutoIncrement, &c.Comment,
		); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}
		c.DefaultValue = defaultValue
		c.MaxLength = maxLength
		c.NumericScale = numericScale
		c.NumericPrecision = numericPrecision

		if mapped, ok := TypeMapping[c.DataType]; ok {
			c.DataType = mapped
		}

		columns = append(columns, c)
	}

	return columns, nil
}

func (d *PostgreSQLDriver) GetIndexes(table string) ([]IndexInfo, error) {
	d.mu.RLock()
	pool := d.pool
	d.mu.RUnlock()

	if pool == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	parts := strings.Split(table, ".")
	var schema, tableName string
	if len(parts) == 2 {
		schema = parts[0]
		tableName = parts[1]
	} else {
		schema = "public"
		tableName = table
	}

	query := `
		SELECT
			i.relname AS index_name,
			n.nspname AS table_schema,
			t.relname AS table_name,
			ARRAY(
				SELECT a.attname
				FROM pg_index ix
				JOIN pg_attribute a ON a.attrelid = ix.indrelid AND a.attnum = ANY(ix.indkey)
				WHERE ix.indexrelid = i.oid
				ORDER BY array_position(ix.indkey, a.attnum)
			) AS columns,
			ix.indisunique AS is_unique,
			ix.indisprimary AS is_primary,
			am.amname AS type
		FROM pg_index ix
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_namespace n ON n.oid = t.relnamespace
		JOIN pg_am am ON am.oid = i.relam
		WHERE n.nspname = $1 AND t.relname = $2
		  AND NOT ix.indisprimary
		ORDER BY i.relname;
	`

	rows, err := pool.Query(ctx, query, schema, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get indexes: %w", err)
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var idx IndexInfo
		if err := rows.Scan(&idx.Name, &idx.TableSchema, &idx.TableName, (*pgtypeTextArray)(&idx.Columns), &idx.IsUnique, &idx.IsPrimary, &idx.Type); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}
		indexes = append(indexes, idx)
	}

	return indexes, nil
}

func (d *PostgreSQLDriver) GetForeignKeys(table string) ([]ForeignKeyInfo, error) {
	d.mu.RLock()
	pool := d.pool
	d.mu.RUnlock()

	if pool == nil {
		return nil, fmt.Errorf("not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	parts := strings.Split(table, ".")
	var schema, tableName string
	if len(parts) == 2 {
		schema = parts[0]
		tableName = parts[1]
	} else {
		schema = "public"
		tableName = table
	}

	query := `
		SELECT
			tc.constraint_name,
			tc.table_schema,
			tc.table_name,
			kcu.column_name,
			ccu.table_schema AS foreign_table_schema,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name,
			rc.confupdtype AS on_update,
			rc.confdeltype AS on_delete
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage ccu ON tc.constraint_name = ccu.constraint_name
		JOIN pg_catalog.pg_constraint rc ON rc.conname = tc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND tc.table_schema = $1 AND tc.table_name = $2
		ORDER BY tc.constraint_name, kcu.ordinal_position;
	`

	rows, err := pool.Query(ctx, query, schema, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get foreign keys: %w", err)
	}
	defer rows.Close()

	var fks []ForeignKeyInfo
	for rows.Next() {
		var fk ForeignKeyInfo
		var onUpdate, onDelete string
		if err := rows.Scan(
			&fk.Name, &fk.TableSchema, &fk.TableName, &fk.ColumnName,
			&fk.ForeignTableSchema, &fk.ForeignTableName, &fk.ForeignColumnName,
			&onUpdate, &onDelete,
		); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key: %w", err)
		}
		fk.OnUpdate = mapCharToConstraint(onUpdate)
		fk.OnDelete = mapCharToConstraint(onDelete)
		fks = append(fks, fk)
	}

	return fks, nil
}

func mapCharToConstraint(c string) string {
	switch c {
	case "a":
		return "NO ACTION"
	case "r":
		return "RESTRICT"
	case "c":
		return "CASCADE"
	case "n":
		return "SET NULL"
	case "d":
		return "SET DEFAULT"
	default:
		return c
	}
}

func (d *PostgreSQLDriver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.pool != nil {
		d.pool.Close()
		d.pool = nil
	}
	return nil
}

func (d *PostgreSQLDriver) Ping(ctx context.Context) error {
	d.mu.RLock()
	pool := d.pool
	d.mu.RUnlock()

	if pool == nil {
		return fmt.Errorf("not connected")
	}
	return pool.Ping(ctx)
}

type pgtypeTextArray []string

func (a *pgtypeTextArray) Scan(value any) error {
	if value == nil {
		*a = nil
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan type %T into pgtypeTextArray", value)
	}
	*a = ParsePGArray(s)
	return nil
}

// Type returns the DatabaseType for this driver.
func (d *PostgreSQLDriver) Type() driver.DatabaseType {
	return driver.DatabaseTypePostgreSQL
}
