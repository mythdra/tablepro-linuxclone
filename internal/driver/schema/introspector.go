package schema

import (
	"context"

	"tablepro/internal/driver"
)

// SchemaIntrospector defines the interface for database schema introspection.
// Each database driver should implement this interface to provide schema metadata.
type SchemaIntrospector interface {
	// GetDatabases returns a list of all databases/schemas
	GetDatabases(ctx context.Context) ([]DatabaseInfo, error)

	// GetTables returns a list of tables in the specified database
	GetTables(ctx context.Context, database string) ([]TableInfo, error)

	// GetColumns returns column information for a specific table
	GetColumns(ctx context.Context, database, table string) ([]ColumnInfo, error)

	// GetIndexes returns index information for a specific table
	GetIndexes(ctx context.Context, database, table string) ([]IndexInfo, error)

	// GetForeignKeys returns foreign key information for a specific table
	GetForeignKeys(ctx context.Context, database, table string) ([]ForeignKeyInfo, error)

	// GetViews returns a list of views in the specified database
	GetViews(ctx context.Context, database string) ([]ViewInfo, error)

	// GetDatabase returns comprehensive schema information for a database
	GetDatabase(ctx context.Context, database string) (*DatabaseSchema, error)
}

// DatabaseInfo represents high-level database information
type DatabaseInfo struct {
	Name         string `json:"name"`
	CharacterSet string `json:"characterSet,omitempty"`
	Collation    string `json:"collation,omitempty"`
	Size         int64  `json:"size,omitempty"`
	TableCount   int    `json:"tableCount,omitempty"`
}

// TableInfo represents table metadata
type TableInfo struct {
	Name      string `json:"name"`
	Schema    string `json:"schema,omitempty"`
	Type      string `json:"type"` // "table", "view", "materialized view"
	Engine    string `json:"engine,omitempty"`
	RowCount  int64  `json:"rowCount"`
	DataSize  int64  `json:"dataSize"`
	IndexSize int64  `json:"indexSize,omitempty"`
	TotalSize int64  `json:"totalSize,omitempty"`
	Comment   string `json:"comment,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// ColumnInfo represents column metadata
type ColumnInfo struct {
	Name             string   `json:"name"`
	DataType         string   `json:"dataType"`
	NativeType       string   `json:"nativeType,omitempty"`
	IsNullable       bool     `json:"isNullable"`
	DefaultValue     *string  `json:"defaultValue,omitempty"`
	MaxLength        *int64   `json:"maxLength,omitempty"`
	NumericScale     *int     `json:"numericScale,omitempty"`
	NumericPrecision *int     `json:"numericPrecision,omitempty"`
	IsPrimaryKey     bool     `json:"isPrimaryKey"`
	IsForeignKey     bool     `json:"isForeignKey"`
	IsUnique         bool     `json:"isUnique"`
	IsAutoIncrement  bool     `json:"isAutoIncrement"`
	OrdinalPosition  int      `json:"ordinalPosition"`
	Comment          string   `json:"comment,omitempty"`
	EnumValues       []string `json:"enumValues,omitempty"`
	SetValues        []string `json:"setValues,omitempty"`
}

// IndexInfo represents index metadata
type IndexInfo struct {
	Name        string   `json:"name"`
	TableSchema string   `json:"tableSchema,omitempty"`
	TableName   string   `json:"tableName"`
	Columns     []string `json:"columns"`
	IsUnique    bool     `json:"isUnique"`
	IsPrimary   bool     `json:"isPrimary"`
	IndexType   string   `json:"indexType,omitempty"`
	Cardinality int64    `json:"cardinality,omitempty"`
}

// ForeignKeyInfo represents foreign key metadata
type ForeignKeyInfo struct {
	Name               string   `json:"name"`
	TableSchema        string   `json:"tableSchema,omitempty"`
	TableName          string   `json:"tableName"`
	Columns            []string `json:"columns"`
	ForeignTableSchema string   `json:"foreignTableSchema,omitempty"`
	ForeignTableName   string   `json:"foreignTableName"`
	ForeignColumns     []string `json:"foreignColumns"`
	OnUpdate           string   `json:"onUpdate,omitempty"`
	OnDelete           string   `json:"onDelete,omitempty"`
}

// ViewInfo represents view metadata
type ViewInfo struct {
	Name       string `json:"name"`
	Schema     string `json:"schema,omitempty"`
	Type       string `json:"type"` // "view", "materialized"
	Definition string `json:"definition,omitempty"`
	Comment    string `json:"comment,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
	UpdatedAt  string `json:"updatedAt,omitempty"`
}

// DatabaseSchema represents complete schema information for a database
type DatabaseSchema struct {
	Database    DatabaseInfo     `json:"database"`
	Tables      []TableInfo      `json:"tables"`
	Views       []ViewInfo       `json:"views"`
	Columns     []ColumnInfo     `json:"-"`
	Indexes     []IndexInfo      `json:"-"`
	ForeignKeys []ForeignKeyInfo `json:"-"`
}

// GetTableByName returns a table by name from the schema
func (d *DatabaseSchema) GetTableByName(name string) *TableInfo {
	for i := range d.Tables {
		if d.Tables[i].Name == name {
			return &d.Tables[i]
		}
	}
	return nil
}

// GetViewByName returns a view by name from the schema
func (d *DatabaseSchema) GetViewByName(name string) *ViewInfo {
	for i := range d.Views {
		if d.Views[i].Name == name {
			return &d.Views[i]
		}
	}
	return nil
}

// NewSchemaIntrospector creates a new schema introspector for the given database type
func NewSchemaIntrospector(dbType driver.DatabaseType, db interface{}) SchemaIntrospector {
	switch dbType {
	case driver.DatabaseTypePostgreSQL:
		if pg, ok := db.(PostgreSQLIntrospector); ok {
			return &PostgreSQLIntrospectorImpl{pg: pg}
		}
	case driver.DatabaseTypeMySQL:
		if mysql, ok := db.(MySQLIntrospector); ok {
			return &MySQLIntrospectorImpl{mysql: mysql}
		}
	case driver.DatabaseTypeSQLite:
		if sqlite, ok := db.(SQLiteIntrospector); ok {
			return &SQLiteIntrospectorImpl{sqlite: sqlite}
		}
	case driver.DatabaseTypeDuckDB:
		if duckdb, ok := db.(DuckDBIntrospector); ok {
			return &DuckDBIntrospectorImpl{duckdb: duckdb}
		}
	case driver.DatabaseTypeMSSQL:
		if mssql, ok := db.(MSSQLIntrospector); ok {
			return &MSSQLIntrospectorImpl{mssql: mssql}
		}
	case driver.DatabaseTypeClickHouse:
		if ch, ok := db.(ClickHouseIntrospector); ok {
			return &ClickHouseIntrospectorImpl{ch: ch}
		}
	case driver.DatabaseTypeMongoDB:
		if mongo, ok := db.(MongoDBIntrospector); ok {
			return &MongoDBIntrospectorImpl{mongo: mongo}
		}
	case driver.DatabaseTypeRedis:
		if redis, ok := db.(RedisIntrospector); ok {
			return &RedisIntrospectorImpl{redis: redis}
		}
	}
	return nil
}
