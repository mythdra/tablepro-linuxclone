package schema

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"tablepro/internal/driver"
)

type PostgreSQLIntrospector interface {
	GetDB() *sql.DB
}

type PostgreSQLIntrospectorImpl struct {
	pg PostgreSQLIntrospector
}

func (i *PostgreSQLIntrospectorImpl) GetDatabases(ctx context.Context) ([]DatabaseInfo, error) {
	query := `
		SELECT datname, pg_encoding_to_char(encoding), datcollate
		FROM pg_database
		WHERE datistemplate = false
		ORDER BY datname
	`
	rows, err := i.pg.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	var dbs []DatabaseInfo
	for rows.Next() {
		var db DatabaseInfo
		if err := rows.Scan(&db.Name, &db.CharacterSet, &db.Collation); err != nil {
			return nil, fmt.Errorf("failed to scan database: %w", err)
		}
		dbs = append(dbs, db)
	}
	return dbs, rows.Err()
}

func (i *PostgreSQLIntrospectorImpl) GetTables(ctx context.Context, database string) ([]TableInfo, error) {
	query := `
		SELECT 
			t.table_name,
			t.table_schema,
			COALESCE(pg.total_relation_size(quote_ident(t.table_schema) || '.' || quote_ident(t.table_name)), 0) as total_size,
			COALESCE(pg_relation_size(quote_ident(t.table_schema) || '.' || quote_ident(t.table_name)), 0) as data_size,
			COALESCE(pg_indexes_size(quote_ident(t.table_schema) || '.' || quote_ident(t.table_name)), 0) as index_size,
			COALESCE((SELECT reltuples::bigint FROM pg_class WHERE relname = t.table_name AND relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = t.table_schema)), 0) as row_count,
			obj_description((quote_ident(t.table_schema) || '.' || quote_ident(t.table_name))::regclass, 'pg_class') as comment
		FROM information_schema.tables t
		LEFT JOIN pg_class ON relname = t.table_name
		LEFT JOIN pg_namespace ON relnamespace = pg_namespace.oid AND nspname = t.table_schema
		WHERE t.table_schema = 'public' AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_name
	`
	rows, err := i.pg.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var t TableInfo
		if err := rows.Scan(&t.Name, &t.Schema, &t.TotalSize, &t.DataSize, &t.IndexSize, &t.RowCount, &t.Comment); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		t.Type = "table"
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (i *PostgreSQLIntrospectorImpl) GetColumns(ctx context.Context, database, table string) ([]ColumnInfo, error) {
	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			c.character_maximum_length,
			c.numeric_precision,
			c.numeric_scale,
			c.ordinal_position,
			c.column_comment,
			cd.constraint_type,
			pg_get_serial_sequence(c.table_schema || '.' || c.table_name, c.column_name) IS NOT NULL as is_auto_increment
		FROM information_schema.columns c
		LEFT JOIN information_schema.table_constraints tc 
			ON tc.table_name = c.table_name AND tc.table_schema = c.table_schema AND tc.constraint_type = 'PRIMARY KEY'
		LEFT JOIN information_schema.key_column_usage kcu 
			ON kcu.table_name = c.table_name AND kcu.table_schema = c.table_schema AND kcu.column_name = c.column_name
		LEFT JOIN information_schema.constraint_column_usage ccu 
			ON ccu.constraint_name = tc.constraint_name AND ccu.table_schema = tc.table_schema AND ccu.column_name = c.column_name
		LEFT JOIN (
			SELECT tc.table_schema, tc.table_name, kcu.column_name, 'PRIMARY KEY' as constraint_type
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu ON kcu.constraint_name = tc.constraint_name
			WHERE tc.constraint_type = 'PRIMARY KEY'
		) cd ON cd.table_schema = c.table_schema AND cd.table_name = c.table_name AND cd.column_name = c.column_name
		WHERE c.table_name = $1 AND c.table_schema = 'public'
		ORDER BY c.ordinal_position
	`
	rows, err := i.pg.GetDB().QueryContext(ctx, query, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var c ColumnInfo
		var nullable, isAutoInc sql.NullString
		var defaultVal sql.NullString
		var maxLen, numPrec, numScale, ordinal sql.NullInt64
		var comment sql.NullString
		var constraintType sql.NullString

		if err := rows.Scan(&c.Name, &c.DataType, &nullable, &defaultVal, &maxLen, &numPrec, &numScale, &ordinal, &comment, &constraintType, &isAutoInc); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		c.IsNullable = nullable.String == "YES"
		c.IsPrimaryKey = constraintType.String == "PRIMARY KEY"
		c.IsAutoIncrement = isAutoInc.String == "YES"
		if defaultVal.Valid {
			c.DefaultValue = &defaultVal.String
		}
		if maxLen.Valid {
			c.MaxLength = &maxLen.Int64
		}
		if numPrec.Valid {
			prec := int(numPrec.Int64)
			c.NumericPrecision = &prec
		}
		if numScale.Valid {
			scale := int(numScale.Int64)
			c.NumericScale = &scale
		}
		if ordinal.Valid {
			c.OrdinalPosition = int(ordinal.Int64)
		}
		if comment.Valid {
			c.Comment = comment.String
		}
		columns = append(columns, c)
	}
	return columns, rows.Err()
}

func (i *PostgreSQLIntrospectorImpl) GetIndexes(ctx context.Context, database, table string) ([]IndexInfo, error) {
	query := `
		SELECT 
			ix.relname as index_name,
			am.amname as index_type,
			ARRAY_AGG(a.attname ORDER BY array_position(ix.indkey, a.attnum)) as columns,
			ix.indisunique as is_unique,
			ix.indisprimary as is_primary
		FROM pg_index ix
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_namespace n ON n.oid = t.relnamespace
		JOIN pg_class ix ON ix.oid = ix.indexrelid
		JOIN pg_am am ON am.oid = ix.relam
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		WHERE t.relname = $1 AND n.nspname = 'public'
		GROUP BY ix.relname, am.amname, ix.indisunique, ix.indisprimary
		ORDER BY ix.relname
	`
	rows, err := i.pg.GetDB().QueryContext(ctx, query, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var idx IndexInfo
		var columns []string
		if err := rows.Scan(&idx.Name, &idx.IndexType, (*[]string)(&columns), &idx.IsUnique, &idx.IsPrimary); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}
		idx.TableName = table
		idx.Columns = columns
		indexes = append(indexes, idx)
	}
	return indexes, rows.Err()
}

func (i *PostgreSQLIntrospectorImpl) GetForeignKeys(ctx context.Context, database, table string) ([]ForeignKeyInfo, error) {
	query := `
		SELECT 
			con.conname as constraint_name,
			con.relname as table_name,
			ARRAY_AGG(a.attname ORDER BY array_position(con.conkey, a.attnum)) as columns,
			ref_t.relname as foreign_table_name,
			ARRAY_AGG(ref_a.attname ORDER BY array_position(con.confkey, ref_a.attnum)) as foreign_columns,
			con.confdeltype::text as on_delete,
			con.confupdtype::text as on_update
		FROM pg_constraint con
		JOIN pg_class t ON t.oid = con.conrelid
		JOIN pg_namespace n ON n.oid = t.relnamespace
		JOIN pg_attribute a ON a.attrelid = con.conrelid AND a.attnum = ANY(con.conkey)
		JOIN pg_class ref_t ON ref_t.oid = con.confrelid
		JOIN pg_attribute ref_a ON ref_a.attrelid = con.confrelid AND ref_a.attnum = ANY(con.confkey)
		WHERE t.relname = $1 AND n.nspname = 'public' AND con.contype = 'f'
		GROUP BY con.conname, con.relname, ref_t.relname, con.confdeltype, con.confupdtype
		ORDER BY con.conname
	`
	rows, err := i.pg.GetDB().QueryContext(ctx, query, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys: %w", err)
	}
	defer rows.Close()

	var fks []ForeignKeyInfo
	for rows.Next() {
		var fk ForeignKeyInfo
		var columns, foreignColumns []string
		if err := rows.Scan(&fk.Name, &fk.TableName, (*[]string)(&columns), &fk.ForeignTableName, (*[]string)(&foreignColumns), &fk.OnDelete, &fk.OnUpdate); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key: %w", err)
		}
		fk.Columns = columns
		fk.ForeignColumns = foreignColumns
		fks = append(fks, fk)
	}
	return fks, rows.Err()
}

func (i *PostgreSQLIntrospectorImpl) GetViews(ctx context.Context, database string) ([]ViewInfo, error) {
	query := `
		SELECT 
			t.table_name,
			t.table_schema,
			pg_get_viewdef(quote_ident(t.table_schema) || '.' || quote_ident(t.table_name), true) as definition
		FROM information_schema.views t
		WHERE t.table_schema = 'public'
		ORDER BY t.table_name
	`
	rows, err := i.pg.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query views: %w", err)
	}
	defer rows.Close()

	var views []ViewInfo
	for rows.Next() {
		var v ViewInfo
		if err := rows.Scan(&v.Name, &v.Schema, &v.Definition); err != nil {
			return nil, fmt.Errorf("failed to scan view: %w", err)
		}
		v.Type = "view"
		views = append(views, v)
	}
	return views, rows.Err()
}

func (i *PostgreSQLIntrospectorImpl) GetDatabase(ctx context.Context, database string) (*DatabaseSchema, error) {
	schema := &DatabaseSchema{
		Database: DatabaseInfo{Name: database},
	}

	tables, err := i.GetTables(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	schema.Tables = tables

	views, err := i.GetViews(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("failed to get views: %w", err)
	}
	schema.Views = views

	for idx := range schema.Tables {
		t := &schema.Tables[idx]
		cols, err := i.GetColumns(ctx, database, t.Name)
		if err != nil {
			continue
		}
		schema.Columns = append(schema.Columns, cols...)
	}

	return schema, nil
}

type MySQLIntrospector interface {
	GetDB() *sql.DB
}

type MySQLIntrospectorImpl struct {
	mysql MySQLIntrospector
}

func (i *MySQLIntrospectorImpl) GetDatabases(ctx context.Context) ([]DatabaseInfo, error) {
	query := "SHOW DATABASES"
	rows, err := i.mysql.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	var dbs []DatabaseInfo
	for rows.Next() {
		var db DatabaseInfo
		if err := rows.Scan(&db.Name); err != nil {
			return nil, fmt.Errorf("failed to scan database: %w", err)
		}
		if db.Name != "information_schema" && db.Name != "performance_schema" && db.Name != "mysql" && db.Name != "sys" {
			dbs = append(dbs, db)
		}
	}
	return dbs, rows.Err()
}

func (i *MySQLIntrospectorImpl) GetTables(ctx context.Context, database string) ([]TableInfo, error) {
	query := `
		SELECT 
			TABLE_NAME,
			TABLE_TYPE,
			ENGINE,
			TABLE_ROWS,
			DATA_LENGTH,
			INDEX_LENGTH,
			TABLE_COMMENT
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`
	rows, err := i.mysql.GetDB().QueryContext(ctx, query, database)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var t TableInfo
		if err := rows.Scan(&t.Name, &t.Type, &t.Engine, &t.RowCount, &t.DataSize, &t.IndexSize, &t.Comment); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		t.TotalSize = t.DataSize + t.IndexSize
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (i *MySQLIntrospectorImpl) GetColumns(ctx context.Context, database, table string) ([]ColumnInfo, error) {
	query := `
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			COLUMN_TYPE,
			IS_NULLABLE,
			COLUMN_KEY,
			EXTRA,
			COLUMN_DEFAULT,
			COLUMN_COMMENT,
			CHARACTER_MAXIMUM_LENGTH,
			NUMERIC_SCALE,
			NUMERIC_PRECISION,
			ORDINAL_POSITION
		FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`
	rows, err := i.mysql.GetDB().QueryContext(ctx, query, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var c ColumnInfo
		var columnKey, extra string
		var defaultVal, comment sql.NullString
		var maxLen, numScale, numPrec, ordinal sql.NullInt64

		if err := rows.Scan(&c.Name, &c.DataType, &c.NativeType, &c.IsNullable, &columnKey, &extra, &defaultVal, &comment, &maxLen, &numScale, &numPrec, &ordinal); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		c.IsPrimaryKey = columnKey == "PRI"
		c.IsAutoIncrement = strings.Contains(extra, "auto_increment")
		if defaultVal.Valid {
			c.DefaultValue = &defaultVal.String
		}
		if maxLen.Valid {
			c.MaxLength = &maxLen.Int64
		}
		if numScale.Valid {
			scale := int(numScale.Int64)
			c.NumericScale = &scale
		}
		if numPrec.Valid {
			prec := int(numPrec.Int64)
			c.NumericPrecision = &prec
		}
		if ordinal.Valid {
			c.OrdinalPosition = int(ordinal.Int64)
		}
		if comment.Valid {
			c.Comment = comment.String
		}
		columns = append(columns, c)
	}
	return columns, rows.Err()
}

func (i *MySQLIntrospectorImpl) GetIndexes(ctx context.Context, database, table string) ([]IndexInfo, error) {
	query := `
		SELECT 
			INDEX_NAME,
			NON_UNIQUE,
			INDEX_TYPE,
			CARDINALITY,
			GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX)
		FROM information_schema.STATISTICS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		GROUP BY INDEX_NAME, NON_UNIQUE, INDEX_TYPE, CARDINALITY
		ORDER BY INDEX_NAME
	`
	rows, err := i.mysql.GetDB().QueryContext(ctx, query, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var idx IndexInfo
		var columns string
		if err := rows.Scan(&idx.Name, &idx.IsUnique, &idx.IndexType, &idx.Cardinality, &columns); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}
		idx.IsUnique = !idx.IsUnique
		idx.IsPrimary = idx.Name == "PRIMARY"
		idx.TableName = table
		idx.Columns = strings.Split(columns, ",")
		indexes = append(indexes, idx)
	}
	return indexes, rows.Err()
}

func (i *MySQLIntrospectorImpl) GetForeignKeys(ctx context.Context, database, table string) ([]ForeignKeyInfo, error) {
	query := `
		SELECT 
			CONSTRAINT_NAME,
			TABLE_NAME,
			COLUMN_NAME,
			REFERENCED_TABLE_NAME,
			REFERENCED_COLUMN_NAME,
			DELETE_RULE,
			UPDATE_RULE
		FROM information_schema.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? 
		AND REFERENCED_TABLE_NAME IS NOT NULL
		ORDER BY CONSTRAINT_NAME
	`
	rows, err := i.mysql.GetDB().QueryContext(ctx, query, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys: %w", err)
	}
	defer rows.Close()

	fkMap := make(map[string]*ForeignKeyInfo)
	for rows.Next() {
		var fk ForeignKeyInfo
		var tableName, columnName, refTable, refColumn string

		if err := rows.Scan(&fk.Name, &tableName, &columnName, &refTable, &refColumn, &fk.OnDelete, &fk.OnUpdate); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key: %w", err)
		}

		if existing, exists := fkMap[fk.Name]; exists {
			existing.Columns = append(existing.Columns, columnName)
			existing.ForeignColumns = append(existing.ForeignColumns, refColumn)
		} else {
			fk.ForeignTableName = refTable
			fk.TableName = tableName
			fk.Columns = []string{columnName}
			fk.ForeignColumns = []string{refColumn}
			fkMap[fk.Name] = &fk
		}
	}

	fks := make([]ForeignKeyInfo, 0, len(fkMap))
	for _, fk := range fkMap {
		fks = append(fks, *fk)
	}
	return fks, rows.Err()
}

func (i *MySQLIntrospectorImpl) GetViews(ctx context.Context, database string) ([]ViewInfo, error) {
	query := `
		SELECT 
			TABLE_NAME,
			TABLE_SCHEMA,
			VIEW_DEFINITION
		FROM information_schema.VIEWS 
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME
	`
	rows, err := i.mysql.GetDB().QueryContext(ctx, query, database)
	if err != nil {
		return nil, fmt.Errorf("failed to query views: %w", err)
	}
	defer rows.Close()

	var views []ViewInfo
	for rows.Next() {
		var v ViewInfo
		if err := rows.Scan(&v.Name, &v.Schema, &v.Definition); err != nil {
			return nil, fmt.Errorf("failed to scan view: %w", err)
		}
		v.Type = "view"
		views = append(views, v)
	}
	return views, rows.Err()
}

func (i *MySQLIntrospectorImpl) GetDatabase(ctx context.Context, database string) (*DatabaseSchema, error) {
	schema := &DatabaseSchema{
		Database: DatabaseInfo{Name: database},
	}

	tables, err := i.GetTables(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	schema.Tables = tables

	views, err := i.GetViews(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("failed to get views: %w", err)
	}
	schema.Views = views

	return schema, nil
}

type SQLiteIntrospector interface {
	GetDB() *sql.DB
}

type SQLiteIntrospectorImpl struct {
	sqlite SQLiteIntrospector
}

func (i *SQLiteIntrospectorImpl) GetDatabases(ctx context.Context) ([]DatabaseInfo, error) {
	return []DatabaseInfo{{Name: "main"}}, nil
}

func (i *SQLiteIntrospectorImpl) GetTables(ctx context.Context, database string) ([]TableInfo, error) {
	query := `SELECT name, type FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`
	rows, err := i.sqlite.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var t TableInfo
		if err := rows.Scan(&t.Name, &t.Type); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (i *SQLiteIntrospectorImpl) GetColumns(ctx context.Context, database, table string) ([]ColumnInfo, error) {
	query := fmt.Sprintf("PRAGMA table_info('%s')", table)
	rows, err := i.sqlite.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var c ColumnInfo
		var cid, notNull, pk int
		var defaultVal sql.NullString

		if err := rows.Scan(&cid, &c.Name, &c.NativeType, &defaultVal, &notNull, &pk); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		c.IsNullable = notNull == 0
		c.IsPrimaryKey = pk == 1
		if defaultVal.Valid {
			c.DefaultValue = &defaultVal.String
		}
		c.OrdinalPosition = cid
		columns = append(columns, c)
	}
	return columns, rows.Err()
}

func (i *SQLiteIntrospectorImpl) GetIndexes(ctx context.Context, database, table string) ([]IndexInfo, error) {
	query := fmt.Sprintf("PRAGMA index_list('%s')", table)
	rows, err := i.sqlite.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var idx IndexInfo
		var origin string

		if err := rows.Scan(&idx.Name, &idx.IsUnique, &origin, &idx.IsPrimary); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}
		idx.TableName = table
		indexes = append(indexes, idx)
	}
	return indexes, rows.Err()
}

func (i *SQLiteIntrospectorImpl) GetForeignKeys(ctx context.Context, database, table string) ([]ForeignKeyInfo, error) {
	query := fmt.Sprintf("PRAGMA foreign_key_list('%s')", table)
	rows, err := i.sqlite.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys: %w", err)
	}
	defer rows.Close()

	var fks []ForeignKeyInfo
	for rows.Next() {
		var fk ForeignKeyInfo
		var id, seq int

		if err := rows.Scan(&id, &seq, &fk.ForeignTableName, &fk.ForeignColumns, &fk.OnUpdate, &fk.OnDelete); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key: %w", err)
		}
		fk.TableName = table
		fk.Name = fmt.Sprintf("fk_%d_%d", id, seq)
		fks = append(fks, fk)
	}
	return fks, rows.Err()
}

func (i *SQLiteIntrospectorImpl) GetViews(ctx context.Context, database string) ([]ViewInfo, error) {
	query := `SELECT name, sql FROM sqlite_master WHERE type='view' ORDER BY name`
	rows, err := i.sqlite.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query views: %w", err)
	}
	defer rows.Close()

	var views []ViewInfo
	for rows.Next() {
		var v ViewInfo
		if err := rows.Scan(&v.Name, &v.Definition); err != nil {
			return nil, fmt.Errorf("failed to scan view: %w", err)
		}
		v.Type = "view"
		views = append(views, v)
	}
	return views, rows.Err()
}

func (i *SQLiteIntrospectorImpl) GetDatabase(ctx context.Context, database string) (*DatabaseSchema, error) {
	schema := &DatabaseSchema{
		Database: DatabaseInfo{Name: database},
	}

	tables, err := i.GetTables(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	schema.Tables = tables

	views, err := i.GetViews(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("failed to get views: %w", err)
	}
	schema.Views = views

	return schema, nil
}

type DuckDBIntrospector interface {
	GetDB() *sql.DB
}

type DuckDBIntrospectorImpl struct {
	duckdb DuckDBIntrospector
}

func (i *DuckDBIntrospectorImpl) GetDatabases(ctx context.Context) ([]DatabaseInfo, error) {
	query := `SELECT database_name FROM duckdb_databases() ORDER BY database_name`
	rows, err := i.duckdb.GetDB().QueryContext(ctx, query)
	if err != nil {
		return []DatabaseInfo{{Name: "main"}}, nil
	}
	defer rows.Close()

	var dbs []DatabaseInfo
	for rows.Next() {
		var db DatabaseInfo
		if err := rows.Scan(&db.Name); err != nil {
			return nil, fmt.Errorf("failed to scan database: %w", err)
		}
		dbs = append(dbs, db)
	}
	if len(dbs) == 0 {
		dbs = []DatabaseInfo{{Name: "main"}}
	}
	return dbs, rows.Err()
}

func (i *DuckDBIntrospectorImpl) GetTables(ctx context.Context, database string) ([]TableInfo, error) {
	query := `
		SELECT table_name, table_type
		FROM information_schema.tables
		WHERE table_schema = 'main'
		ORDER BY table_name
	`
	rows, err := i.duckdb.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var t TableInfo
		if err := rows.Scan(&t.Name, &t.Type); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (i *DuckDBIntrospectorImpl) GetColumns(ctx context.Context, database, table string) ([]ColumnInfo, error) {
	query := `
		SELECT 
			column_name, 
			data_type, 
			is_nullable, 
			column_default,
			character_maximum_length,
			ordinal_position
		FROM information_schema.columns
		WHERE table_name = $1 AND table_schema = 'main'
		ORDER BY ordinal_position
	`
	rows, err := i.duckdb.GetDB().QueryContext(ctx, query, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var c ColumnInfo
		var nullable string
		var defaultVal sql.NullString
		var maxLen, ordinal sql.NullInt64

		if err := rows.Scan(&c.Name, &c.DataType, &nullable, &defaultVal, &maxLen, &ordinal); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		c.IsNullable = nullable == "YES"
		if defaultVal.Valid {
			c.DefaultValue = &defaultVal.String
		}
		if maxLen.Valid {
			c.MaxLength = &maxLen.Int64
		}
		if ordinal.Valid {
			c.OrdinalPosition = int(ordinal.Int64)
		}
		columns = append(columns, c)
	}
	return columns, rows.Err()
}

func (i *DuckDBIntrospectorImpl) GetIndexes(ctx context.Context, database, table string) ([]IndexInfo, error) {
	return []IndexInfo{}, nil
}

func (i *DuckDBIntrospectorImpl) GetForeignKeys(ctx context.Context, database, table string) ([]ForeignKeyInfo, error) {
	return []ForeignKeyInfo{}, nil
}

func (i *DuckDBIntrospectorImpl) GetViews(ctx context.Context, database string) ([]ViewInfo, error) {
	query := `
		SELECT table_name, table_schema
		FROM information_schema.views
		WHERE table_schema = 'main'
		ORDER BY table_name
	`
	rows, err := i.duckdb.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query views: %w", err)
	}
	defer rows.Close()

	var views []ViewInfo
	for rows.Next() {
		var v ViewInfo
		if err := rows.Scan(&v.Name, &v.Schema); err != nil {
			return nil, fmt.Errorf("failed to scan view: %w", err)
		}
		v.Type = "view"
		views = append(views, v)
	}
	return views, rows.Err()
}

func (i *DuckDBIntrospectorImpl) GetDatabase(ctx context.Context, database string) (*DatabaseSchema, error) {
	schema := &DatabaseSchema{
		Database: DatabaseInfo{Name: database},
	}

	tables, err := i.GetTables(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	schema.Tables = tables

	views, err := i.GetViews(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("failed to get views: %w", err)
	}
	schema.Views = views

	return schema, nil
}

type MSSQLIntrospector interface {
	GetDB() *sql.DB
}

type MSSQLIntrospectorImpl struct {
	mssql MSSQLIntrospector
}

func (i *MSSQLIntrospectorImpl) GetDatabases(ctx context.Context) ([]DatabaseInfo, error) {
	query := `SELECT name FROM sys.databases WHERE name NOT IN ('master', 'tempdb', 'model', 'msdb') ORDER BY name`
	rows, err := i.mssql.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	var dbs []DatabaseInfo
	for rows.Next() {
		var db DatabaseInfo
		if err := rows.Scan(&db.Name); err != nil {
			return nil, fmt.Errorf("failed to scan database: %w", err)
		}
		dbs = append(dbs, db)
	}
	return dbs, rows.Err()
}

func (i *MSSQLIntrospectorImpl) GetTables(ctx context.Context, database string) ([]TableInfo, error) {
	query := `
		SELECT 
			t.NAME,
			t.create_date,
			t.modify_date,
			p.rows as row_count,
			sa.total_space * 8 as total_size,
			sa.used_space * 8 as used_space
		FROM sys.tables t
		INNER JOIN sys.indexes i ON t.object_id = i.object_id
		INNER JOIN sys.partitions p ON i.object_id = p.object_id AND i.index_id = p.index_id
		OUTER APPLY (
			SELECT SUM(a.total_pages) * 8 AS total_space, SUM(a.used_pages) * 8 AS used_space
			FROM sys.allocation_units a
			WHERE p.partition_id = a.container_id
		) sa
		WHERE i.index_id IN (0, 1)
		ORDER BY t.NAME
	`
	rows, err := i.mssql.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var t TableInfo
		if err := rows.Scan(&t.Name, &t.CreatedAt, &t.UpdatedAt, &t.RowCount, &t.TotalSize, &t.DataSize); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		t.Type = "table"
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (i *MSSQLIntrospectorImpl) GetColumns(ctx context.Context, database, table string) ([]ColumnInfo, error) {
	query := `
		SELECT 
			c.name,
			t.name as data_type,
			c.is_nullable,
			dc.definition as default_value,
			c.max_length,
			c.precision,
			c.scale,
			c.column_id,
			ccol.value as comment,
			CASE WHEN ic.column_id IS NOT NULL THEN 1 ELSE 0 END as is_identity
		FROM sys.columns c
		INNER JOIN sys.types t ON c.user_type_id = t.user_type_id
		LEFT JOIN sys.default_constraints dc ON c.default_object_id = dc.object_id
		LEFT JOIN sys.extended_properties ccol ON c.object_id = ccol.major_id AND c.column_id = ccol.minor_id AND ccol.name = 'MS_Description'
		LEFT JOIN sys.identity_columns ic ON c.object_id = ic.object_id AND c.column_id = ic.column_id
		WHERE c.object_id = OBJECT_ID(@tableName)
		ORDER BY c.column_id
	`
	rows, err := i.mssql.GetDB().QueryContext(ctx, query, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var c ColumnInfo
		var defaultVal, comment sql.NullString
		var maxLen, precision, scale, ordinal sql.NullInt64

		if err := rows.Scan(&c.Name, &c.DataType, &c.IsNullable, &defaultVal, &maxLen, &precision, &scale, &ordinal, &comment, &c.IsAutoIncrement); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		if defaultVal.Valid {
			c.DefaultValue = &defaultVal.String
		}
		c.NativeType = c.DataType
		if maxLen.Valid {
			c.MaxLength = &maxLen.Int64
		}
		if precision.Valid {
			prec := int(precision.Int64)
			c.NumericPrecision = &prec
		}
		if scale.Valid {
			scaleVal := int(scale.Int64)
			c.NumericScale = &scaleVal
		}
		if ordinal.Valid {
			c.OrdinalPosition = int(ordinal.Int64)
		}
		if comment.Valid {
			c.Comment = comment.String
		}
		columns = append(columns, c)
	}
	return columns, rows.Err()
}

func (i *MSSQLIntrospectorImpl) GetIndexes(ctx context.Context, database, table string) ([]IndexInfo, error) {
	query := `
		SELECT 
			i.name as index_name,
			i.type_desc as index_type,
			i.is_unique,
			i.is_primary_key,
			STUFF((
				SELECT ', ' + c.name
				FROM sys.index_columns ic
				INNER JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
				WHERE ic.object_id = i.object_id AND ic.index_id = i.index_id
				ORDER BY ic.key_ordinal
				FOR XML PATH('')
			), 1, 2, '') as columns
		FROM sys.indexes i
		WHERE i.object_id = OBJECT_ID(@tableName) AND i.type > 0
		ORDER BY i.name
	`
	rows, err := i.mssql.GetDB().QueryContext(ctx, query, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var idx IndexInfo
		var columns string

		if err := rows.Scan(&idx.Name, &idx.IndexType, &idx.IsUnique, &idx.IsPrimary, &columns); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}
		idx.TableName = table
		idx.Columns = strings.Split(columns, ", ")
		indexes = append(indexes, idx)
	}
	return indexes, rows.Err()
}

func (i *MSSQLIntrospectorImpl) GetForeignKeys(ctx context.Context, database, table string) ([]ForeignKeyInfo, error) {
	query := `
		SELECT 
			fk.name as constraint_name,
			tp.name as table_name,
			cp.name as column_name,
			tr.name as foreign_table_name,
			cr.name as foreign_column_name,
			fk.delete_referential_action_desc,
			fk.update_referential_action_desc
		FROM sys.foreign_keys fk
		INNER JOIN sys.foreign_key_columns fkc ON fk.object_id = fkc.constraint_object_id
		INNER JOIN sys.tables tp ON fkc.parent_object_id = tp.object_id
		INNER JOIN sys.columns cp ON fkc.parent_object_id = cp.object_id AND fkc.parent_column_id = cp.column_id
		INNER JOIN sys.tables tr ON fkc.referenced_object_id = tr.object_id
		INNER JOIN sys.columns cr ON fkc.referenced_object_id = cr.object_id AND fkc.referenced_column_id = cr.column_id
		WHERE tp.name = @tableName
		ORDER BY fk.name, fkc.constraint_column_id
	`
	rows, err := i.mssql.GetDB().QueryContext(ctx, query, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys: %w", err)
	}
	defer rows.Close()

	fkMap := make(map[string]*ForeignKeyInfo)
	for rows.Next() {
		var fk ForeignKeyInfo
		var tableName, columnName, refTable, refColumn string

		if err := rows.Scan(&fk.Name, &tableName, &columnName, &refTable, &refColumn, &fk.OnDelete, &fk.OnUpdate); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key: %w", err)
		}

		if existing, exists := fkMap[fk.Name]; exists {
			existing.Columns = append(existing.Columns, columnName)
			existing.ForeignColumns = append(existing.ForeignColumns, refColumn)
		} else {
			fk.ForeignTableName = refTable
			fk.TableName = tableName
			fk.Columns = []string{columnName}
			fk.ForeignColumns = []string{refColumn}
			fkMap[fk.Name] = &fk
		}
	}

	fks := make([]ForeignKeyInfo, 0, len(fkMap))
	for _, fk := range fkMap {
		fks = append(fks, *fk)
	}
	return fks, rows.Err()
}

func (i *MSSQLIntrospectorImpl) GetViews(ctx context.Context, database string) ([]ViewInfo, error) {
	query := `
		SELECT 
			v.name,
			v.create_date,
			v.modify_date,
			sm.definition
		FROM sys.views v
		LEFT JOIN sys.sql_modules sm ON v.object_id = sm.object_id
		ORDER BY v.name
	`
	rows, err := i.mssql.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query views: %w", err)
	}
	defer rows.Close()

	var views []ViewInfo
	for rows.Next() {
		var v ViewInfo
		if err := rows.Scan(&v.Name, &v.CreatedAt, &v.UpdatedAt, &v.Definition); err != nil {
			return nil, fmt.Errorf("failed to scan view: %w", err)
		}
		v.Type = "view"
		views = append(views, v)
	}
	return views, rows.Err()
}

func (i *MSSQLIntrospectorImpl) GetDatabase(ctx context.Context, database string) (*DatabaseSchema, error) {
	schema := &DatabaseSchema{
		Database: DatabaseInfo{Name: database},
	}

	tables, err := i.GetTables(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	schema.Tables = tables

	views, err := i.GetViews(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("failed to get views: %w", err)
	}
	schema.Views = views

	return schema, nil
}

type ClickHouseIntrospector interface {
	GetDB() *sql.DB
}

type ClickHouseIntrospectorImpl struct {
	ch ClickHouseIntrospector
}

func (i *ClickHouseIntrospectorImpl) GetDatabases(ctx context.Context) ([]DatabaseInfo, error) {
	query := `SELECT name FROM system.databases ORDER BY name`
	rows, err := i.ch.GetDB().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query databases: %w", err)
	}
	defer rows.Close()

	var dbs []DatabaseInfo
	for rows.Next() {
		var db DatabaseInfo
		if err := rows.Scan(&db.Name); err != nil {
			return nil, fmt.Errorf("failed to scan database: %w", err)
		}
		dbs = append(dbs, db)
	}
	return dbs, rows.Err()
}

func (i *ClickHouseIntrospectorImpl) GetTables(ctx context.Context, database string) ([]TableInfo, error) {
	query := `
		SELECT 
			name,
			engine,
			total_rows,
			total_bytes,
			data_compressed_bytes,
			data_uncompressed_bytes,
			comment
		FROM system.tables
		WHERE database = ? AND is_temporary = 0
		ORDER BY name
	`
	rows, err := i.ch.GetDB().QueryContext(ctx, query, database)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var t TableInfo
		if err := rows.Scan(&t.Name, &t.Engine, &t.RowCount, &t.TotalSize, &t.DataSize, &t.IndexSize, &t.Comment); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		t.Type = "table"
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (i *ClickHouseIntrospectorImpl) GetColumns(ctx context.Context, database, table string) ([]ColumnInfo, error) {
	query := `
		SELECT 
			name,
			type,
			default_expression,
			is_nullable,
			comment,
			ordinal
		FROM system.columns
		WHERE database = ? AND table = ?
		ORDER BY ordinal
	`
	rows, err := i.ch.GetDB().QueryContext(ctx, query, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var c ColumnInfo
		var defaultExpr, comment sql.NullString

		if err := rows.Scan(&c.Name, &c.NativeType, &defaultExpr, &c.IsNullable, &comment, &c.OrdinalPosition); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		c.DataType = c.NativeType
		if defaultExpr.Valid {
			c.DefaultValue = &defaultExpr.String
		}
		if comment.Valid {
			c.Comment = comment.String
		}
		columns = append(columns, c)
	}
	return columns, rows.Err()
}

func (i *ClickHouseIntrospectorImpl) GetIndexes(ctx context.Context, database, table string) ([]IndexInfo, error) {
	return []IndexInfo{}, nil
}

func (i *ClickHouseIntrospectorImpl) GetForeignKeys(ctx context.Context, database, table string) ([]ForeignKeyInfo, error) {
	return []ForeignKeyInfo{}, nil
}

func (i *ClickHouseIntrospectorImpl) GetViews(ctx context.Context, database string) ([]ViewInfo, error) {
	query := `
		SELECT name, engine
		FROM system.tables
		WHERE database = ? AND is_temporary = 0 AND engine LIKE '%View'
		ORDER BY name
	`
	rows, err := i.ch.GetDB().QueryContext(ctx, query, database)
	if err != nil {
		return nil, fmt.Errorf("failed to query views: %w", err)
	}
	defer rows.Close()

	var views []ViewInfo
	for rows.Next() {
		var v ViewInfo
		if err := rows.Scan(&v.Name, &v.Type); err != nil {
			return nil, fmt.Errorf("failed to scan view: %w", err)
		}
		views = append(views, v)
	}
	return views, rows.Err()
}

func (i *ClickHouseIntrospectorImpl) GetDatabase(ctx context.Context, database string) (*DatabaseSchema, error) {
	schema := &DatabaseSchema{
		Database: DatabaseInfo{Name: database},
	}

	tables, err := i.GetTables(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	schema.Tables = tables

	views, err := i.GetViews(ctx, database)
	if err != nil {
		return nil, fmt.Errorf("failed to get views: %w", err)
	}
	schema.Views = views

	return schema, nil
}

type MongoDBIntrospector interface {
	GetDatabase() interface{}
}

type MongoDBIntrospectorImpl struct {
	mongo MongoDBIntrospector
}

func (i *MongoDBIntrospectorImpl) GetDatabases(ctx context.Context) ([]DatabaseInfo, error) {
	return []DatabaseInfo{{Name: "default"}}, nil
}

func (i *MongoDBIntrospectorImpl) GetTables(ctx context.Context, database string) ([]TableInfo, error) {
	return []TableInfo{{Name: "collections", Type: "collection"}}, nil
}

func (i *MongoDBIntrospectorImpl) GetColumns(ctx context.Context, database, table string) ([]ColumnInfo, error) {
	return []ColumnInfo{}, nil
}

func (i *MongoDBIntrospectorImpl) GetIndexes(ctx context.Context, database, table string) ([]IndexInfo, error) {
	return []IndexInfo{}, nil
}

func (i *MongoDBIntrospectorImpl) GetForeignKeys(ctx context.Context, database, table string) ([]ForeignKeyInfo, error) {
	return []ForeignKeyInfo{}, nil
}

func (i *MongoDBIntrospectorImpl) GetViews(ctx context.Context, database string) ([]ViewInfo, error) {
	return []ViewInfo{}, nil
}

func (i *MongoDBIntrospectorImpl) GetDatabase(ctx context.Context, database string) (*DatabaseSchema, error) {
	schema := &DatabaseSchema{
		Database: DatabaseInfo{Name: database},
	}
	return schema, nil
}

type RedisIntrospector interface {
	GetClient() interface{}
}

type RedisIntrospectorImpl struct {
	redis RedisIntrospector
}

func (i *RedisIntrospectorImpl) GetDatabases(ctx context.Context) ([]DatabaseInfo, error) {
	return []DatabaseInfo{{Name: "default"}}, nil
}

func (i *RedisIntrospectorImpl) GetTables(ctx context.Context, database string) ([]TableInfo, error) {
	return []TableInfo{}, nil
}

func (i *RedisIntrospectorImpl) GetColumns(ctx context.Context, database, table string) ([]ColumnInfo, error) {
	return []ColumnInfo{}, nil
}

func (i *RedisIntrospectorImpl) GetIndexes(ctx context.Context, database, table string) ([]IndexInfo, error) {
	return []IndexInfo{}, nil
}

func (i *RedisIntrospectorImpl) GetForeignKeys(ctx context.Context, database, table string) ([]ForeignKeyInfo, error) {
	return []ForeignKeyInfo{}, nil
}

func (i *RedisIntrospectorImpl) GetViews(ctx context.Context, database string) ([]ViewInfo, error) {
	return []ViewInfo{}, nil
}

func (i *RedisIntrospectorImpl) GetDatabase(ctx context.Context, database string) (*DatabaseSchema, error) {
	schema := &DatabaseSchema{
		Database: DatabaseInfo{Name: database},
	}
	return schema, nil
}

func GetDriverForSchema(dbType driver.DatabaseType) string {
	switch dbType {
	case driver.DatabaseTypePostgreSQL:
		return "PostgreSQL"
	case driver.DatabaseTypeMySQL:
		return "MySQL"
	case driver.DatabaseTypeSQLite:
		return "SQLite"
	case driver.DatabaseTypeDuckDB:
		return "DuckDB"
	case driver.DatabaseTypeMSSQL:
		return "MSSQL"
	case driver.DatabaseTypeClickHouse:
		return "ClickHouse"
	case driver.DatabaseTypeMongoDB:
		return "MongoDB"
	case driver.DatabaseTypeRedis:
		return "Redis"
	default:
		return "Unknown"
	}
}

func IsKeyValueStore(dbType driver.DatabaseType) bool {
	return dbType == driver.DatabaseTypeRedis || dbType == driver.DatabaseTypeMongoDB
}

func SupportsSchemaIntrospection(dbType driver.DatabaseType) bool {
	return !IsKeyValueStore(dbType)
}
