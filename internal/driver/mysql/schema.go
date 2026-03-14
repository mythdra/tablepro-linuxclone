package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type SchemaInfo struct {
	Name   string
	Tables []TableInfo
	Views  []TableInfo
}

type TableInfo struct {
	Name      string
	Type      string
	Engine    string
	Rows      int64
	DataSize  int64
	IndexSize int64
	Comment   string
}

type ColumnMetadata struct {
	Name             string
	DataType         string
	Type             string
	IsNullable       bool
	IsAutoIncrement  bool
	IsPrimaryKey     bool
	DefaultValue     any
	Comment          string
	MaxLength        int64
	NumericScale     int
	NumericPrecision int
	EnumValues       []string
	SetValues        []string
}

type IndexMetadata struct {
	Name        string
	Columns     []string
	IsUnique    bool
	IsPrimary   bool
	IndexType   string
	Cardinality int64
}

type ForeignKeyMetadata struct {
	Name              string
	Columns           []string
	ReferencedTable   string
	ReferencedColumns []string
	OnDelete          string
	OnUpdate          string
}

func (d *MySQLDriver) FetchDatabases(ctx context.Context) ([]string, error) {
	query := "SHOW DATABASES"
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch databases: %w", err)
	}
	defer rows.Close()

	databases := make([]string, 0)
	for rows.Next() {
		var db string
		if err := rows.Scan(&db); err != nil {
			return nil, fmt.Errorf("failed to scan database: %w", err)
		}
		if db != "information_schema" && db != "performance_schema" && db != "mysql" && db != "sys" {
			databases = append(databases, db)
		}
	}

	return databases, nil
}

func (d *MySQLDriver) FetchTables(ctx context.Context, database string) ([]TableInfo, error) {
	query := `SELECT TABLE_NAME, TABLE_TYPE, ENGINE, TABLE_ROWS, 
		DATA_LENGTH + INDEX_LENGTH AS total_size, TABLE_COMMENT
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME`

	rows, err := d.db.QueryContext(ctx, query, database)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tables: %w", err)
	}
	defer rows.Close()

	tables := make([]TableInfo, 0)
	for rows.Next() {
		var t TableInfo
		if err := rows.Scan(&t.Name, &t.Type, &t.Engine, &t.Rows, &t.DataSize, &t.Comment); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		tables = append(tables, t)
	}

	return tables, nil
}

func (d *MySQLDriver) FetchViews(ctx context.Context, database string) ([]TableInfo, error) {
	query := `SELECT TABLE_NAME, TABLE_TYPE, ENGINE, TABLE_ROWS, 
		DATA_LENGTH + INDEX_LENGTH AS total_size, TABLE_COMMENT
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'VIEW'
		ORDER BY TABLE_NAME`

	rows, err := d.db.QueryContext(ctx, query, database)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch views: %w", err)
	}
	defer rows.Close()

	views := make([]TableInfo, 0)
	for rows.Next() {
		var v TableInfo
		if err := rows.Scan(&v.Name, &v.Type, &v.Engine, &v.Rows, &v.DataSize, &v.Comment); err != nil {
			return nil, fmt.Errorf("failed to scan view: %w", err)
		}
		views = append(views, v)
	}

	return views, nil
}

func (d *MySQLDriver) FetchColumns(ctx context.Context, database, table string) ([]ColumnMetadata, error) {
	query := `SELECT 
		COLUMN_NAME, DATA_TYPE, COLUMN_TYPE, IS_NULLABLE, 
		COLUMN_KEY, EXTRA, COLUMN_DEFAULT, COLUMN_COMMENT,
		CHARACTER_MAXIMUM_LENGTH, NUMERIC_SCALE, NUMERIC_PRECISION,
		IF(COLUMN_TYPE LIKE 'enum(%)', SUBSTRING(COLUMN_TYPE, 6, LENGTH(COLUMN_TYPE) - 6), NULL) AS enum_values,
		IF(COLUMN_TYPE LIKE 'set(%)', SUBSTRING(COLUMN_TYPE, 5, LENGTH(COLUMN_TYPE) - 5), NULL) AS set_values
		FROM information_schema.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION`

	rows, err := d.db.QueryContext(ctx, query, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch columns: %w", err)
	}
	defer rows.Close()

	columns := make([]ColumnMetadata, 0)
	for rows.Next() {
		var c ColumnMetadata
		var columnKey, extra string
		var enumValues, setValues sql.NullString

		err := rows.Scan(
			&c.Name, &c.DataType, &c.Type, &c.IsNullable,
			&columnKey, &extra, &c.DefaultValue, &c.Comment,
			&c.MaxLength, &c.NumericScale, &c.NumericPrecision,
			&enumValues, &setValues,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		c.IsPrimaryKey = columnKey == "PRI"
		c.IsAutoIncrement = strings.Contains(extra, "auto_increment")

		if enumValues.Valid && enumValues.String != "" {
			c.EnumValues = parseEnumSetValues(enumValues.String)
		}
		if setValues.Valid && setValues.String != "" {
			c.SetValues = parseEnumSetValues(setValues.String)
		}

		columns = append(columns, c)
	}

	return columns, nil
}

func parseEnumSetValues(s string) []string {
	s = strings.Trim(s, "'\"")
	parts := strings.Split(s, "','")
	result := make([]string, len(parts))
	for i, p := range parts {
		result[i] = strings.Trim(p, "'")
	}
	return result
}

func (d *MySQLDriver) FetchIndexes(ctx context.Context, database, table string) ([]IndexMetadata, error) {
	query := `SELECT INDEX_NAME, NON_UNIQUE, SEQ_IN_INDEX, INDEX_TYPE, CARDINALITY, COLUMN_NAME
		FROM information_schema.STATISTICS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY INDEX_NAME, SEQ_IN_INDEX`

	rows, err := d.db.QueryContext(ctx, query, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch indexes: %w", err)
	}
	defer rows.Close()

	indexMap := make(map[string]*IndexMetadata)
	for rows.Next() {
		var indexName string
		var nonUnique bool
		var seq int
		var indexType string
		var cardinality int64
		var columnName string

		if err := rows.Scan(&indexName, &nonUnique, &seq, &indexType, &cardinality, &columnName); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}

		if idx, exists := indexMap[indexName]; exists {
			idx.Columns = append(idx.Columns, columnName)
		} else {
			indexMap[indexName] = &IndexMetadata{
				Name:        indexName,
				IsUnique:    !nonUnique,
				IsPrimary:   indexName == "PRIMARY",
				IndexType:   indexType,
				Cardinality: cardinality,
				Columns:     []string{columnName},
			}
		}
	}

	indexes := make([]IndexMetadata, 0, len(indexMap))
	for _, idx := range indexMap {
		indexes = append(indexes, *idx)
	}

	return indexes, nil
}

func (d *MySQLDriver) FetchForeignKeys(ctx context.Context, database, table string) ([]ForeignKeyMetadata, error) {
	query := `SELECT 
		CONSTRAINT_NAME, TABLE_NAME, COLUMN_NAME,
		REFERENCED_TABLE_NAME, REFERENCED_COLUMN_NAME,
		DELETE_RULE, UPDATE_RULE
		FROM information_schema.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? 
		AND REFERENCED_TABLE_NAME IS NOT NULL
		ORDER BY CONSTRAINT_NAME, ORDINAL_POSITION`

	rows, err := d.db.QueryContext(ctx, query, database, table)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch foreign keys: %w", err)
	}
	defer rows.Close()

	fkMap := make(map[string]*ForeignKeyMetadata)
	for rows.Next() {
		var fk ForeignKeyMetadata
		var tableName, columnName, refTable, refColumn string

		if err := rows.Scan(&fk.Name, &tableName, &columnName, &refTable, &refColumn, &fk.OnDelete, &fk.OnUpdate); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key: %w", err)
		}

		if existing, exists := fkMap[fk.Name]; exists {
			existing.Columns = append(existing.Columns, columnName)
			existing.ReferencedColumns = append(existing.ReferencedColumns, refColumn)
		} else {
			fk.ReferencedTable = refTable
			fk.Columns = []string{columnName}
			fk.ReferencedColumns = []string{refColumn}
			fkMap[fk.Name] = &fk
		}
	}

	fks := make([]ForeignKeyMetadata, 0, len(fkMap))
	for _, fk := range fkMap {
		fks = append(fks, *fk)
	}

	return fks, nil
}

func (d *MySQLDriver) FetchTableDDL(ctx context.Context, database, table string) (string, error) {
	query := fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`", database, table)
	var name, ddl string
	err := d.db.QueryRowContext(ctx, query).Scan(&name, &ddl)
	if err != nil {
		return "", fmt.Errorf("failed to fetch table DDL: %w", err)
	}
	return ddl, nil
}

func (d *MySQLDriver) FetchTableRowCount(ctx context.Context, database, table string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM `%s`.`%s`", database, table)
	var count int64
	err := d.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch row count: %w", err)
	}
	return count, nil
}

func (d *MySQLDriver) FetchSchemas(ctx context.Context, database string) ([]SchemaInfo, error) {
	schemas, err := d.FetchDatabases(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]SchemaInfo, 0, len(schemas))
	for _, schema := range schemas {
		tables, err := d.FetchTables(ctx, schema)
		if err != nil {
			return nil, err
		}
		views, err := d.FetchViews(ctx, schema)
		if err != nil {
			return nil, err
		}
		result = append(result, SchemaInfo{
			Name:   schema,
			Tables: tables,
			Views:  views,
		})
	}

	return result, nil
}
