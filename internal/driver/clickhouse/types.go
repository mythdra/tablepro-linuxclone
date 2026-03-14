package clickhouse

import (
	"fmt"
	"sync"
	"time"

	clickhouse "github.com/ClickHouse/clickhouse-go/v2"
)

// Config holds ClickHouse connection configuration
type Config struct {
	Host      string
	Port      int
	Database  string
	Username  string
	Password  string
	SSLMode   string
	SSLCert   string
	SSLKey    string
	SSLCACert string

	// Connection pool settings
	MaxOpenConnections int
	MaxIdleConnections int
	MaxConnectionLife  time.Duration

	// ClickHouse-specific settings
	Debug        bool
	Compress     string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DialTimeout  time.Duration
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLife  time.Duration
}

// DSN builds the ClickHouse connection string
func (c *Config) DSN() string {
	dsn := fmt.Sprintf(
		"clickhouse://%s:%s@%s:%d/%s",
		c.Username, c.Password, c.Host, c.Port, c.Database,
	)

	params := []string{}

	if c.SSLMode != "" && c.SSLMode != "disable" {
		params = append(params, "ssl=true")
	}

	if c.Debug {
		params = append(params, "debug=true")
	}

	if c.Compress != "" {
		params = append(params, fmt.Sprintf("compress=%s", c.Compress))
	}

	if c.ReadTimeout > 0 {
		params = append(params, fmt.Sprintf("read_timeout=%s", c.ReadTimeout))
	}

	if c.WriteTimeout > 0 {
		params = append(params, fmt.Sprintf("write_timeout=%s", c.WriteTimeout))
	}

	if c.DialTimeout > 0 {
		params = append(params, fmt.Sprintf("dial_timeout=%s", c.DialTimeout))
	}

	if len(params) > 0 {
		dsn += "?" + joinParams(params)
	}

	return dsn
}

func joinParams(params []string) string {
	result := ""
	for i, p := range params {
		if i > 0 {
			result += "&"
		}
		result += p
	}
	return result
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Host:               "localhost",
		Port:               9000,
		SSLMode:            "disable",
		Compress:           "lz4",
		ReadTimeout:        30 * time.Second,
		WriteTimeout:       30 * time.Second,
		DialTimeout:        10 * time.Second,
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
	Name     string
	Database string
	Engine   string
}

// TableInfo holds table metadata
type TableInfo struct {
	Database       string
	Name           string
	Engine         string
	RowCount       uint64
	DataSize       uint64
	IndexSize      uint64
	TotalSize      uint64
	IsMaterialized bool
	IsTemporary    bool
}

// ColumnInfo holds column metadata
type ColumnInfo struct {
	Name              string
	Type              string
	Database          string
	Table             string
	Position          int
	DefaultType       *string
	DefaultExpression *string
	Comment           *string
	IsInPartitionKey  bool
	IsInSortingKey    bool
	IsInPrimaryKey    bool
	IsSparse          *bool
}

// IndexInfo holds index metadata (ClickHouse doesn't have traditional indexes)
type IndexInfo struct {
	Name        string
	Database    string
	Table       string
	Expression  string
	Type        string
	Granularity int
}

// TypeMapping maps ClickHouse data types to Go types
var TypeMapping = map[string]string{
	"Int8":           "int32",
	"Int16":          "int32",
	"Int32":          "int32",
	"Int64":          "int64",
	"UInt8":          "uint32",
	"UInt16":         "uint32",
	"UInt32":         "uint32",
	"UInt64":         "uint64",
	"Float32":        "float32",
	"Float64":        "float64",
	"Decimal":        "float64",
	"String":         "string",
	"FixedString":    "string",
	"Date":           "time.Time",
	"Date32":         "time.Time",
	"DateTime":       "time.Time",
	"DateTime64":     "time.Time",
	"UUID":           "string",
	"Enum":           "string",
	"Enum8":          "string",
	"Enum16":         "string",
	"Array":          "[]any",
	"JSON":           "string",
	"Tuple":          "string",
	"Map":            "map[string]any",
	"Nested":         "string",
	"LowCardinality": "string",
}

// ClickHouseDriver implements the database driver interface for ClickHouse
type ClickHouseDriver struct {
	db     clickhouse.Conn
	config *Config
	mu     sync.RWMutex
}

// NewClickHouseDriver creates a new ClickHouse driver instance
func NewClickHouseDriver() *ClickHouseDriver {
	return &ClickHouseDriver{
		config: DefaultConfig(),
	}
}
