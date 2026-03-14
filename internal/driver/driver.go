package driver

import (
	"context"
	"database/sql"
	"time"
)

// DatabaseDriver defines the interface for all database drivers.
// This interface is compatible with PostgreSQL, MySQL, SQLite, DuckDB, MSSQL, and ClickHouse.
// NoSQL drivers (MongoDB, Redis) will adapt this interface.
type DatabaseDriver interface {
	// Connect establishes a connection to the database using the provided config.
	Connect(ctx context.Context, config *ConnectionConfig) error

	// Execute runs a query and returns rows affected or results.
	Execute(ctx context.Context, query string, params ...any) (*Result, error)

	// Query runs a query and returns rows.
	Query(ctx context.Context, query string, params ...any) (*Row, error)

	// QueryContext runs a query with custom timeout.
	QueryContext(ctx context.Context, timeout time.Duration, query string, params ...any) (*Row, error)

	// GetSchema returns database schema information.
	GetSchema(ctx context.Context) (*SchemaInfo, error)

	// GetTables returns all tables in the database.
	GetTables(ctx context.Context, schemaName string) ([]TableInfo, error)

	// GetColumns returns column information for a specific table.
	GetColumns(ctx context.Context, schemaName, tableName string) ([]ColumnInfo, error)

	// GetIndexes returns index information for a specific table.
	GetIndexes(ctx context.Context, schemaName, tableName string) ([]IndexInfo, error)

	// GetForeignKeys returns foreign key information for a specific table.
	GetForeignKeys(ctx context.Context, schemaName, tableName string) ([]ForeignKeyInfo, error)

	// Ping checks if the connection is alive.
	Ping(ctx context.Context) error

	// Close closes the database connection.
	Close() error

	// GetCapabilities returns the driver capabilities.
	GetCapabilities() *DriverCapabilities

	// GetDB returns the underlying *sql.DB for advanced operations.
	GetDB() *sql.DB

	// Type returns the DatabaseType for this driver.
	Type() DatabaseType
}

// ConnectionConfig holds database connection configuration.
type ConnectionConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string

	// Optional parameters
	MaxOpenConnections int
	MaxIdleConnections int
	MaxConnectionLife  time.Duration
	QueryTimeout       time.Duration
}

// Result represents the result of an Execute operation.
type Result struct {
	LastInsertID int64
	RowsAffected int64
}

// Row represents a row of data returned from a query.
type Row struct {
	// Data contains column values keyed by column name.
	Data map[string]any

	// ColumnNames holds the ordered list of column names.
	ColumnNames []string
}

// ColumnInfo holds column metadata.
type ColumnInfo struct {
	Name            string
	DataType        string
	TypeName        string
	Nullable        bool
	DefaultValue    *string
	IsPrimaryKey    bool
	IsAutoIncrement bool
	MaxLength       *int64
	Precision       *int32
	Scale           *int32
	Comment         *string
}

// SchemaInfo holds database schema information.
type SchemaInfo struct {
	Tables     []TableInfo
	Views      []TableInfo
	Procedures []RoutineInfo
	Functions  []RoutineInfo
	Schemas    []string
}

// TableInfo holds table metadata.
type TableInfo struct {
	Name      string
	Schema    string
	Type      TableType
	Comment   *string
	RowCount  *int64
	SizeBytes *int64
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// TableType represents the type of table-like object.
type TableType string

const (
	TableTypeTable            TableType = "TABLE"
	TableTypeView             TableType = "VIEW"
	TableTypeMaterializedView TableType = "MATERIALIZED VIEW"
	TableTypeSystem           TableType = "SYSTEM TABLE"
)

// RoutineInfo holds stored procedure or function metadata.
type RoutineInfo struct {
	Name       string
	Schema     string
	Type       RoutineType
	Definition string
	Comment    *string
}

// RoutineType represents the type of routine.
type RoutineType string

const (
	RoutineTypeProcedure RoutineType = "PROCEDURE"
	RoutineTypeFunction  RoutineType = "FUNCTION"
)

// IndexInfo holds index metadata.
type IndexInfo struct {
	Name      string
	Columns   []string
	IsUnique  bool
	IsPrimary bool
	IsPartial bool
	IndexType string
	Comment   *string
}

// ForeignKeyInfo holds foreign key metadata.
type ForeignKeyInfo struct {
	Name              string
	Columns           []string
	ReferencedTable   string
	ReferencedColumns []string
	OnDelete          string
	OnUpdate          string
	MatchType         string
}

// DriverCapabilities holds driver capabilities and limits.
type DriverCapabilities struct {
	// Features indicates which features are supported by the driver.
	Features []Feature

	// MaxConnections is the maximum number of concurrent connections.
	MaxConnections int

	// MaxQueryTime is the maximum allowed query execution time (0 = unlimited).
	MaxQueryTime time.Duration

	// SupportsTransactions indicates if the driver supports transactions.
	SupportsTransactions bool

	// SupportsStoredProcedures indicates if the driver supports stored procedures.
	SupportsStoredProcedures bool

	// SupportsFunctions indicates if the driver supports functions.
	SupportsFunctions bool

	// SupportsViews indicates if the driver supports views.
	SupportsViews bool

	// SupportsMaterializedViews indicates if the driver supports materialized views.
	SupportsMaterializedViews bool

	// SupportsForeignKeys indicates if the driver supports foreign keys.
	SupportsForeignKeys bool

	// SupportsIndexes indicates if the driver supports indexes.
	SupportsIndexes bool

	// SupportsAutoIncrement indicates if the driver supports auto-increment.
	SupportsAutoIncrement bool

	// SupportsSchemas indicates if the driver supports schemas/namespaces.
	SupportsSchemas bool
}

// Feature represents an optional driver feature.
type Feature string

const (
	FeatureSSLConnection      Feature = "SSL_CONNECTION"
	FeatureSSHConnection      Feature = "SSH_TUNNEL"
	FeaturePreparedStatements Feature = "PREPARED_STATEMENTS"
	FeatureBatchStatements    Feature = "BATCH_STATEMENTS"
	FeatureCursorPagination   Feature = "CURSOR_PAGINATION"
	FeatureJSONType           Feature = "JSON_TYPE"
	FeatureArrayType          Feature = "ARRAY_TYPE"
	FeatureUUIDType           Feature = "UUID_TYPE"
	FeatureGeometricType      Feature = "GEOMETRIC_TYPE"
	FeatureFullTextSearch     Feature = "FULL_TEXT_SEARCH"
	FeatureWindowFunctions    Feature = "WINDOW_FUNCTIONS"
	FeatureCTE                Feature = "CTE"
	FeatureCTAS               Feature = "CTAS"
	FeatureMultipleSchemas    Feature = "MULTIPLE_SCHEMAS"
)
