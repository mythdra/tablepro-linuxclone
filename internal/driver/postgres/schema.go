package postgres

import "time"

// Row represents a single row from query results
type Row map[string]any

// QueryResult represents the result of a query execution
type QueryResult struct {
	Columns      []string
	Rows         []Row
	AffectedRows int64
}

// SchemaInfo represents database schema information
type SchemaInfo struct {
	Name       string
	Owner      string
	TableCount int
}

type TableInfo struct {
	Schema   string
	Name     string
	Type     string
	RowCount int64
	Size     int64
	Comment  string
}

// ColumnInfo represents column information
type ColumnInfo struct {
	Name             string
	DataType         string
	IsNullable       bool
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

// IndexInfo represents index information
type IndexInfo struct {
	Name        string
	TableSchema string
	TableName   string
	Columns     []string
	IsUnique    bool
	IsPrimary   bool
	Type        string
}

// ForeignKeyInfo represents foreign key information
type ForeignKeyInfo struct {
	Name               string
	TableSchema        string
	TableName          string
	ColumnName         string
	ForeignTableSchema string
	ForeignTableName   string
	ForeignColumnName  string
	OnUpdate           string
	OnDelete           string
}

// TypeMapping maps PostgreSQL types to display types
var TypeMapping = map[string]string{
	"int2":         "smallint",
	"int4":         "integer",
	"int8":         "bigint",
	"float4":       "real",
	"float8":       "double precision",
	"numeric":      "numeric",
	"bool":         "boolean",
	"bytea":        "bytea",
	"text":         "text",
	"varchar":      "varchar",
	"char":         "char",
	"date":         "date",
	"time":         "time",
	"timetz":       "time with time zone",
	"timestamp":    "timestamp",
	"timestamptz":  "timestamp with time zone",
	"interval":     "interval",
	"uuid":         "uuid",
	"json":         "json",
	"jsonb":        "jsonb",
	"xml":          "xml",
	"point":        "point",
	"line":         "line",
	"inet":         "inet",
	"cidr":         "cidr",
	"macaddr":      "macaddr",
	"bit":          "bit",
	"varbit":       "varbit",
	"oid":          "oid",
	"_int4":        "integer[]",
	"_int8":        "bigint[]",
	"_float4":      "real[]",
	"_float8":      "double precision[]",
	"_text":        "text[]",
	"_varchar":     "varchar[]",
	"_bool":        "boolean[]",
	"_date":        "date[]",
	"_timestamp":   "timestamp[]",
	"_timestamptz": "timestamp with time zone[]",
	"_uuid":        "uuid[]",
	"_jsonb":       "jsonb[]",
}

var pgTypeToGo = map[uint32]string{
	16:   "bool",
	20:   "int64",
	21:   "int32",
	23:   "int32",
	25:   "string",
	700:  "float32",
	701:  "float64",
	1042: "string",
	1043: "string",
	1082: "time.Time",
	1083: "time.Time",
	1114: "time.Time",
	1184: "time.Time",
	2950: "string",
	114:  "string",
	3802: "string",
}

func ParsePGArray(s string) []string {
	if len(s) < 2 {
		return nil
	}
	s = s[1 : len(s)-1]
	if s == "" {
		return nil
	}

	var result []string
	var current string
	inQuote := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			inQuote = !inQuote
		case ',':
			if !inQuote {
				result = append(result, current)
				current = ""
			} else {
				current += string(c)
			}
		case '\\':
			if i+1 < len(s) {
				current += string(s[i+1])
				i++
			}
		default:
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// FormatTimestamp formats time.Time to PostgreSQL timestamp string
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000")
}
