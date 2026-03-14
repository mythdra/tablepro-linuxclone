// Package query provides query execution services with timeout, cancellation, and result streaming.
package query

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	stdriver "database/sql/driver"

	"tablepro/internal/driver"
)

// DataType represents normalized data types for frontend handling.
// Ensures consistent type handling across different database systems.
type DataType string

const (
	// DataTypeString represents string/text types
	DataTypeString DataType = "string"
	// DataTypeNumber represents generic numeric types
	DataTypeNumber DataType = "number"
	// DataTypeInteger represents integer types
	DataTypeInteger DataType = "integer"
	// DataTypeFloat represents floating-point types
	DataTypeFloat DataType = "float"
	// DataTypeBoolean represents boolean types
	DataTypeBoolean DataType = "boolean"
	// DataTypeDateTime represents timestamp/datetime types
	DataTypeDateTime DataType = "datetime"
	// DataTypeDate represents date-only types
	DataTypeDate DataType = "date"
	// DataTypeTime represents time-only types
	DataTypeTime DataType = "time"
	// DataTypeJSON represents JSON/JSONB types
	DataTypeJSON DataType = "json"
	// DataTypeBlob represents binary large object types
	DataTypeBlob DataType = "blob"
	// DataTypeArray represents array types
	DataTypeArray DataType = "array"
	// DataTypeUUID represents UUID types
	DataTypeUUID DataType = "uuid"
	// DataTypeNull represents NULL values
	DataTypeNull DataType = "null"
	// DataTypeUnknown represents unmapped types
	DataTypeUnknown DataType = "unknown"
)

// ResultSet represents the result of a query execution with metadata.
// Uses column-oriented data format for efficient frontend processing.
type ResultSet struct {
	// Columns contains metadata about each column in the result
	Columns []ColumnInfo `json:"columns"`

	// Rows contains the actual row data in column-oriented format
	// Each inner slice represents a column's values
	Rows [][]interface{} `json:"rows"`

	// RowCount is the total number of rows returned
	RowCount int64 `json:"rowCount"`

	// QueryTime is the execution duration
	QueryTime time.Duration `json:"queryTime"`

	// Statement is the executed SQL statement
	Statement string `json:"statement"`

	// HasMore indicates if there are more results (for streaming/pagination)
	HasMore bool `json:"hasMore"`

	// MultipleResultSets contains additional result sets for batch queries
	MultipleResultSets []*ResultSet `json:"multipleResultSets,omitempty"`
}

// ColumnInfo contains metadata about a result column.
// Provides type information for frontend rendering.
type ColumnInfo struct {
	// Name is the column name
	Name string `json:"name"`

	// Type is the database-specific type name (e.g., "VARCHAR", "TIMESTAMP")
	Type string `json:"type"`

	// DataType is the normalized type for frontend handling
	DataType string `json:"dataType"`

	// Nullable indicates if the column can contain NULL values
	Nullable bool `json:"nullable"`
}

// NewResultSetFromRows creates a ResultSet from driver rows.
// It handles type mapping, NULL values, and data formatting.
// Converts row-oriented data to column-oriented format for frontend.
func NewResultSetFromRows(rows []*driver.Row, queryTime time.Duration, statement string, dbType driver.DatabaseType) *ResultSet {
	if len(rows) == 0 {
		return &ResultSet{
			Columns:   make([]ColumnInfo, 0),
			Rows:      make([][]interface{}, 0),
			RowCount:  0,
			QueryTime: queryTime,
			Statement: statement,
		}
	}

	// Get column names from first row
	columnNames := rows[0].ColumnNames
	columnCount := len(columnNames)
	rowCount := len(rows)

	// Build column info with type mapping
	columns := make([]ColumnInfo, columnCount)
	for i, name := range columnNames {
		columns[i] = ColumnInfo{
			Name:     name,
			Type:     "unknown",
			DataType: string(DataTypeUnknown),
			Nullable: true,
		}
	}

	// Convert to column-oriented format
	columnData := make([][]interface{}, columnCount)
	for colIdx := range columnData {
		columnData[colIdx] = make([]interface{}, rowCount)
	}

	// Transpose row-oriented to column-oriented with formatting
	for rowIdx, row := range rows {
		for colIdx, colName := range columnNames {
			if value, ok := row.Data[colName]; ok {
				columnData[colIdx][rowIdx] = formatValue(value, columns[colIdx].Type, dbType)
			} else {
				columnData[colIdx][rowIdx] = nil
			}
		}
	}

	return &ResultSet{
		Columns:   columns,
		Rows:      columnData,
		RowCount:  int64(rowCount),
		QueryTime: queryTime,
		Statement: statement,
		HasMore:   false,
	}
}

// normalizeDataType maps database-specific types to normalized DataType enums.
// This ensures consistent frontend handling across different database systems.
// Checks both direct type names and driver-provided type mappings.
func normalizeDataType(dbType driver.DatabaseType, columnType string) DataType {
	// Normalize column type to lowercase for comparison
	normalizedType := strings.ToLower(columnType)

	// First check: direct type matching for special cases that need precise handling
	switch normalizedType {
	// Boolean types - check first to avoid misclassification
	case "bool", "boolean", "bit":
		return DataTypeBoolean

	// Date vs DateTime distinction
	case "date":
		return DataTypeDate

	// Time types (without date)
	case "time", "timetz":
		return DataTypeTime

	// Timestamp/datetime types
	case "timestamp", "timestamptz", "datetime", "datetime2", "smalldatetime":
		return DataTypeDateTime

	// JSON types
	case "json", "jsonb":
		return DataTypeJSON

	// UUID
	case "uuid", "uniqueidentifier":
		return DataTypeUUID

	// Array types
	case "array", "json_array":
		return DataTypeArray

	// Binary types
	case "bytea", "blob", "binary", "varbinary", "image", "raw", "longblob", "mediumblob", "tinyblob":
		return DataTypeBlob
	}

	// Get type mapping from driver package
	mapping := driver.GetDataTypeMapping(dbType, normalizedType)
	if mapping != nil {
		// Use mapping to determine normalized type
		if mapping.IsBinary {
			return DataTypeBlob
		}
		if mapping.IsJSON {
			return DataTypeJSON
		}
		if mapping.IsArray {
			return DataTypeArray
		}
		if mapping.IsTime {
			// Distinguish between date, time, and datetime
			switch normalizedType {
			case "date":
				return DataTypeDate
			case "time", "timetz":
				return DataTypeTime
			default:
				return DataTypeDateTime
			}
		}
		if mapping.IsNumeric {
			if mapping.GoType == "float32" || mapping.GoType == "float64" {
				return DataTypeFloat
			}
			return DataTypeInteger
		}
		if mapping.IsString {
			// Check for special string types
			switch normalizedType {
			case "uuid", "uniqueidentifier":
				return DataTypeUUID
			case "date":
				return DataTypeDate
			case "time", "timetz":
				return DataTypeTime
			case "timestamp", "timestamptz", "datetime", "datetime2", "smalldatetime":
				return DataTypeDateTime
			case "bool", "boolean", "bit":
				return DataTypeBoolean
			case "xml", "inet", "cidr", "macaddr", "interval", "enum", "set":
				return DataTypeString
			}
			return DataTypeString
		}
	}

	// Fallback: direct type matching
	switch normalizedType {
	// Integer types
	case "int", "int2", "int4", "int8", "integer", "smallint", "bigint", "tinyint", "mediumint", "serial", "bigserial":
		return DataTypeInteger

	// Float types
	case "float", "float4", "float8", "double", "double precision", "real", "numeric", "decimal", "money", "smallmoney":
		return DataTypeFloat

	// String types
	case "varchar", "text", "char", "bpchar", "character", "character varying", "string", "tinytext", "mediumtext", "longtext", "ntext", "nvarchar", "nchar":
		return DataTypeString

	// Special types
	case "xml", "inet", "cidr", "macaddr", "interval", "enum", "set":
		return DataTypeString

	default:
		return DataTypeUnknown
	}
}

// formatValue handles NULL values, type mapping, and data formatting.
// It ensures all values are JSON-serializable and frontend-friendly.
// Delegates to type-specific formatters based on normalized data type.
func formatValue(value interface{}, columnType string, dbType driver.DatabaseType) interface{} {
	// Handle NULL values - convert to nil for JSON null
	if value == nil {
		return nil
	}

	// Handle sql.NullString and similar nullable types
	if val, ok := value.(sql.NullString); ok {
		if !val.Valid {
			return nil
		}
		return val.String
	}
	if val, ok := value.(sql.NullInt64); ok {
		if !val.Valid {
			return nil
		}
		return val.Int64
	}
	if val, ok := value.(sql.NullFloat64); ok {
		if !val.Valid {
			return nil
		}
		return val.Float64
	}
	if val, ok := value.(sql.NullBool); ok {
		if !val.Valid {
			return nil
		}
		return val.Bool
	}
	if val, ok := value.(sql.NullTime); ok {
		if !val.Valid {
			return nil
		}
		return formatTime(val.Time)
	}

	// Handle driver.Valuer interface (common in database drivers)
	if valuer, ok := value.(stdriver.Valuer); ok {
		driverValue, err := valuer.Value()
		if err != nil || driverValue == nil {
			return nil
		}
		value = driverValue
	}

	// Normalize the column type for matching
	normalizedType := strings.ToLower(columnType)
	dataType := normalizeDataType(dbType, normalizedType)

	// Format based on normalized data type
	switch dataType {
	case DataTypeDateTime:
		return formatDateTimeValue(value)
	case DataTypeDate:
		return formatDateValue(value)
	case DataTypeTime:
		return formatTimeValue(value)
	case DataTypeBoolean:
		return formatBooleanValue(value)
	case DataTypeFloat, DataTypeInteger:
		return formatNumericValue(value)
	case DataTypeJSON:
		return formatJSONValue(value)
	case DataTypeBlob:
		return formatBlobValue(value)
	case DataTypeArray:
		return formatArrayValue(value)
	case DataTypeUUID:
		return formatUUIDValue(value)
	default:
		return formatStringValue(value)
	}
}

// formatDateTimeValue formats time.Time and datetime values to RFC3339 string.
// Handles both time.Time objects and string representations.
func formatDateTimeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case time.Time:
		return v.Format(time.RFC3339)
	case string:
		// Try to parse and re-format for consistency
		if t, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", v); err == nil {
			return t.Format(time.RFC3339)
		}
		if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			return t.Format(time.RFC3339)
		}
		return v
	default:
		return fmt.Sprintf("%v", value)
	}
}

// formatDateValue formats date values to ISO 8601 date string (YYYY-MM-DD).
func formatDateValue(value interface{}) interface{} {
	switch v := value.(type) {
	case time.Time:
		return v.Format("2006-01-02")
	case string:
		return v
	default:
		return fmt.Sprintf("%v", value)
	}
}

// formatTimeValue formats time values to ISO 8601 time string with timezone.
func formatTimeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case time.Time:
		return v.Format("15:04:05.999999999Z07:00")
	case string:
		return v
	default:
		return fmt.Sprintf("%v", value)
	}
}

// formatTime is a helper for sql.NullTime handling.
// Returns RFC3339 formatted time string.
func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// formatBooleanValue ensures booleans are true/false, not 1/0.
// Handles various input types including integers and strings.
func formatBooleanValue(value interface{}) interface{} {
	switch v := value.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int8:
		return v != 0
	case int16:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case uint:
		return v != 0
	case uint8:
		return v != 0
	case uint16:
		return v != 0
	case uint32:
		return v != 0
	case uint64:
		return v != 0
	case string:
		// Handle PostgreSQL text booleans
		lower := strings.ToLower(v)
		return lower == "true" || lower == "t" || lower == "yes" || lower == "y" || v == "1"
	default:
		// Fallback: truthy check
		return value != nil && value != "" && value != 0
	}
}

// formatNumericValue preserves precision and avoids scientific notation.
// Returns integers as int64, floats with appropriate precision.
// Numeric strings (from DECIMAL) are returned as-is for exact precision.
func formatNumericValue(value interface{}) interface{} {
	switch v := value.(type) {
	case float32:
		// Check if it's actually an integer value
		if v == float32(int64(v)) && !math.IsInf(float64(v), 0) && !math.IsNaN(float64(v)) {
			return int64(v)
		}
		// Avoid scientific notation for reasonable ranges
		if math.Abs(float64(v)) < 1e15 && math.Abs(float64(v)) > 1e-6 {
			return strconv.FormatFloat(float64(v), 'f', -1, 32)
		}
		return v
	case float64:
		// Check if it's actually an integer value
		if v == float64(int64(v)) && !math.IsInf(v, 0) && !math.IsNaN(v) {
			return int64(v)
		}
		// Avoid scientific notation for reasonable ranges
		if math.Abs(v) < 1e15 && math.Abs(v) > 1e-6 {
			return strconv.FormatFloat(v, 'f', -1, 64)
		}
		return v
	case int, int8, int16, int32, int64:
		return v
	case uint, uint8, uint16, uint32, uint64:
		return v
	case string:
		// Handle numeric strings (from DECIMAL/NUMERIC types)
		// Return as string to preserve exact precision
		if _, err := strconv.ParseFloat(v, 64); err == nil {
			return v
		}
		return v
	case json.Number:
		// json.Number preserves precision
		return v.String()
	default:
		return value
	}
}

// formatJSONValue handles JSON/JSONB column values.
// Attempts to parse and return parsed JSON, or base64 encode if invalid.
func formatJSONValue(value interface{}) interface{} {
	switch v := value.(type) {
	case []byte:
		// Try to parse as JSON first
		var parsed interface{}
		if err := json.Unmarshal(v, &parsed); err == nil {
			return parsed
		}
		// If not valid JSON, base64 encode
		return base64.StdEncoding.EncodeToString(v)
	case string:
		// Try to parse as JSON first
		var parsed interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err == nil {
			return parsed
		}
		return v
	default:
		// Return as-is, let frontend handle
		return value
	}
}

// formatBlobValue handles BLOB/BYTEA column values with base64 encoding.
// Ensures binary data is safely transmitted as JSON string.
func formatBlobValue(value interface{}) interface{} {
	switch v := value.(type) {
	case []byte:
		return base64.StdEncoding.EncodeToString(v)
	case string:
		return base64.StdEncoding.EncodeToString([]byte(v))
	default:
		// Fallback: convert to string representation
		return fmt.Sprintf("%v", value)
	}
}

// formatArrayValue handles array-type values.
// Recursively formats each element using formatValue.
func formatArrayValue(value interface{}) interface{} {
	switch v := value.(type) {
	case []interface{}:
		// Recursively format array elements
		result := make([]interface{}, len(v))
		for i, elem := range v {
			result[i] = formatValue(elem, "unknown", driver.DatabaseTypeUnknown)
		}
		return result
	case []string:
		result := make([]interface{}, len(v))
		for i, elem := range v {
			result[i] = elem
		}
		return result
	case []int:
		result := make([]interface{}, len(v))
		for i, elem := range v {
			result[i] = elem
		}
		return result
	case []int64:
		result := make([]interface{}, len(v))
		for i, elem := range v {
			result[i] = elem
		}
		return result
	case []float64:
		result := make([]interface{}, len(v))
		for i, elem := range v {
			result[i] = elem
		}
		return result
	default:
		// Return as-is for other array types
		return value
	}
}

// formatUUIDValue ensures UUID values are lowercase strings.
func formatUUIDValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return strings.ToLower(v)
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", value)
	}
}

// formatStringValue handles string-type values.
// Converts []byte to string, uses fmt.Stringer when available.
func formatStringValue(value interface{}) interface{} {
	switch v := value.(type) {
	case []byte:
		return string(v)
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		// Convert to string for display
		return fmt.Sprintf("%v", value)
	}
}

// handleNULLValue converts values to JSON-safe format.
// Deprecated: Use formatValue instead for comprehensive handling.
func handleNULLValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	return value
}
