package mssql

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

// Config holds MSSQL connection configuration
type Config struct {
	Host            string
	Port            int
	Database        string
	Username        string
	Password        string
	SSLMode         string
	SSLCert         string
	SSLKey          string
	SSLCACert       string
	AppName         string
	Encrypt         string // "disable", "false", "true"
	TrustServerCert string // "false", "true"

	// Connection pool settings
	MaxOpenConnections int
	MaxIdleConnections int
	MaxConnectionLife  time.Duration

	// Windows Authentication (optional)
	UseWindowsAuth bool
	ServerSPN      string
}

// DSN builds the MSSQL connection string
func (c *Config) DSN() string {
	// Handle Windows Authentication mode
	if c.UseWindowsAuth {
		return fmt.Sprintf(
			"server=%s;port=%d;database=%s;encrypt=%s;trustservercertificate=%s;authenticator=windows",
			c.Host, c.Port, c.Database, c.Encrypt, c.TrustServerCert,
		)
	}

	// Standard SQL Server authentication
	dsn := fmt.Sprintf(
		"server=%s;port=%d;database=%s;user id=%s;password=%s;encrypt=%s;trustservercertificate=%s",
		c.Host, c.Port, c.Database, c.Username, c.Password, c.Encrypt, c.TrustServerCert,
	)

	if c.AppName != "" {
		dsn += fmt.Sprintf(";app name=%s", c.AppName)
	}

	// SSL/TLS configuration
	if c.SSLMode != "" && c.SSLMode != "disable" {
		dsn += fmt.Sprintf(";sslmode=%s", c.SSLMode)
	}

	return dsn
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Host:               "localhost",
		Port:               1433,
		Encrypt:            "false",
		TrustServerCert:    "true",
		MaxOpenConnections: 10,
		MaxIdleConnections: 2,
		MaxConnectionLife:  time.Hour,
	}
}

// QueryResult represents the result of a query execution
type QueryResult struct {
	Columns      []string
	Rows         []Row
	AffectedRows int64
}

// Row represents a row of data returned from a query
type Row map[string]any

// SchemaInfo holds database schema information
type SchemaInfo struct {
	Name       string
	Owner      string
	TableCount int
}

// TableInfo holds table metadata
type TableInfo struct {
	Schema   string
	Name     string
	Type     string
	RowCount int64
	Size     int64
	Comment  string
	IsView   bool
}

// ColumnInfo holds column metadata
type ColumnInfo struct {
	Name             string
	DataType         string
	IsNullable       string
	DefaultValue     *string
	MaxLength        *int
	NumericScale     *int
	NumericPrecision *int
	IsPrimaryKey     bool
	IsForeignKey     bool
	IsUnique         bool
	IsAutoIncrement  bool
	Comment          string
}

// IndexInfo holds index metadata
type IndexInfo struct {
	TableSchema string
	TableName   string
	Name        string
	Columns     []string
	IsUnique    bool
	IsPrimary   bool
	Type        string
}

// ForeignKeyInfo holds foreign key metadata
type ForeignKeyInfo struct {
	TableSchema        string
	TableName          string
	Name               string
	ColumnName         string
	ForeignTableSchema string
	ForeignTableName   string
	ForeignColumnName  string
	OnUpdate           string
	OnDelete           string
}

// TypeMapping maps MSSQL data types to Go types
var TypeMapping = map[string]string{
	"bigint":           "int64",
	"binary":           "[]byte",
	"bit":              "bool",
	"char":             "string",
	"datetime":         "time.Time",
	"datetime2":        "time.Time",
	"decimal":          "float64",
	"float":            "float64",
	"image":            "[]byte",
	"int":              "int32",
	"money":            "float64",
	"nchar":            "string",
	"ntext":            "string",
	"numeric":          "float64",
	"nvarchar":         "string",
	"real":             "float32",
	"smalldatetime":    "time.Time",
	"smallint":         "int16",
	"smallmoney":       "float64",
	"text":             "string",
	"time":             "time.Time",
	"tinyint":          "uint8",
	"uniqueidentifier": "string",
	"varbinary":        "[]byte",
	"varchar":          "string",
	"xml":              "string",
}

// MSSQLDriver implements the database driver interface for MSSQL
type MSSQLDriver struct {
	db     *sql.DB
	config *Config
	mu     sync.RWMutex
}

// NewMSSQLDriver creates a new MSSQL driver instance
func NewMSSQLDriver() *MSSQLDriver {
	return &MSSQLDriver{
		config: DefaultConfig(),
	}
}
