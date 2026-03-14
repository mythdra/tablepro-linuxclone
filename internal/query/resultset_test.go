package query

import (
	"encoding/base64"
	"fmt"
	"math"
	"testing"
	"time"

	"tablepro/internal/driver"
)

func TestNormalizeDataType_PostgreSQL(t *testing.T) {
	tests := []struct {
		name       string
		columnType string
		want       DataType
	}{
		{"boolean", "bool", DataTypeBoolean},
		{"boolean full", "boolean", DataTypeBoolean},
		{"integer smallint", "smallint", DataTypeInteger},
		{"integer int2", "int2", DataTypeInteger},
		{"integer int4", "int4", DataTypeInteger},
		{"integer int8", "int8", DataTypeInteger},
		{"integer bigint", "bigint", DataTypeInteger},
		{"float float4", "float4", DataTypeFloat},
		{"float float8", "float8", DataTypeFloat},
		{"float numeric", "numeric", DataTypeFloat},
		{"float decimal", "decimal", DataTypeFloat},
		{"varchar", "varchar", DataTypeString},
		{"text", "text", DataTypeString},
		{"json", "json", DataTypeJSON},
		{"jsonb", "jsonb", DataTypeJSON},
		{"timestamp", "timestamp", DataTypeDateTime},
		{"timestamptz", "timestamptz", DataTypeDateTime},
		{"date", "date", DataTypeDate},
		{"time", "time", DataTypeTime},
		{"bytea", "bytea", DataTypeBlob},
		{"uuid", "uuid", DataTypeUUID},
		{"array", "array", DataTypeArray},
		{"inet", "inet", DataTypeString},
		{"unknown type", "unknown_type", DataTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeDataType(driver.DatabaseTypePostgreSQL, tt.columnType)
			if got != tt.want {
				t.Errorf("normalizeDataType(PostgreSQL, %q) = %v, want %v", tt.columnType, got, tt.want)
			}
		})
	}
}

func TestNormalizeDataType_MySQL(t *testing.T) {
	tests := []struct {
		name       string
		columnType string
		want       DataType
	}{
		{"tinyint", "tinyint", DataTypeInteger},
		{"int", "int", DataTypeInteger},
		{"bigint", "bigint", DataTypeInteger},
		{"float", "float", DataTypeFloat},
		{"double", "double", DataTypeFloat},
		{"decimal", "decimal", DataTypeFloat},
		{"varchar", "varchar", DataTypeString},
		{"text", "text", DataTypeString},
		{"json", "json", DataTypeJSON},
		{"enum", "enum", DataTypeString},
		{"datetime", "datetime", DataTypeDateTime},
		{"timestamp", "timestamp", DataTypeDateTime},
		{"date", "date", DataTypeDate},
		{"time", "time", DataTypeTime},
		{"blob", "blob", DataTypeBlob},
		{"binary", "binary", DataTypeBlob},
		{"unknown", "unknown", DataTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeDataType(driver.DatabaseTypeMySQL, tt.columnType)
			if got != tt.want {
				t.Errorf("normalizeDataType(MySQL, %q) = %v, want %v", tt.columnType, got, tt.want)
			}
		})
	}
}

func TestNormalizeDataType_SQLite(t *testing.T) {
	tests := []struct {
		name       string
		columnType string
		want       DataType
	}{
		{"integer", "integer", DataTypeInteger},
		{"real", "real", DataTypeFloat},
		{"float", "float", DataTypeFloat},
		{"text", "text", DataTypeString},
		{"blob", "blob", DataTypeBlob},
		{"numeric", "numeric", DataTypeFloat},
		{"unknown", "unknown", DataTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeDataType(driver.DatabaseTypeSQLite, tt.columnType)
			if got != tt.want {
				t.Errorf("normalizeDataType(SQLite, %q) = %v, want %v", tt.columnType, got, tt.want)
			}
		})
	}
}

func TestFormatValue_NULLHandling(t *testing.T) {
	// Test NULL value returns nil
	result := formatValue(nil, "varchar", driver.DatabaseTypePostgreSQL)
	if result != nil {
		t.Errorf("formatValue(nil) = %v, want nil", result)
	}
}

func TestFormatValue_Boolean(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  interface{}
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"int 1", int(1), true},
		{"int 0", int(0), false},
		{"int64 1", int64(1), true},
		{"int64 0", int64(0), false},
		{"string true", "true", true},
		{"string false", "false", false},
		{"string t", "t", true},
		{"string 1", "1", true},
		{"string 0", "0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValue(tt.value, "boolean", driver.DatabaseTypePostgreSQL)
			if got != tt.want {
				t.Errorf("formatBooleanValue(%v) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestFormatValue_DateTime(t *testing.T) {
	testTime := time.Date(2026, 3, 14, 10, 30, 0, 0, time.UTC)

	// Test time.Time formatting
	got := formatValue(testTime, "timestamp", driver.DatabaseTypePostgreSQL)
	if gotStr, ok := got.(string); ok {
		// Should be RFC3339 format
		if _, err := time.Parse(time.RFC3339, gotStr); err != nil {
			t.Errorf("formatDateTimeValue(time.Time) returned invalid RFC3339: %v", got)
		}
	} else {
		t.Errorf("formatDateTimeValue(time.Time) = %T, want string", got)
	}

	// Test string datetime
	got = formatValue("2026-03-14 10:30:00", "datetime", driver.DatabaseTypeMySQL)
	if gotStr, ok := got.(string); ok {
		if _, err := time.Parse(time.RFC3339, gotStr); err != nil {
			t.Errorf("formatDateTimeValue(string) should convert to RFC3339, got: %v", got)
		}
	}
}

func TestFormatValue_Numeric(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{"int", int(42)},
		{"int64", int64(42)},
		{"float32", float32(3.14)},
		{"float64", float64(3.14159)},
		{"large int", int64(9223372036854775807)},
		{"small float", float64(0.000001)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValue(tt.value, "numeric", driver.DatabaseTypePostgreSQL)
			// Should not be scientific notation for reasonable ranges
			if gotStr, ok := got.(string); ok {
				if gotStr == "NaN" || gotStr == "Infinity" || gotStr == "-Infinity" {
					t.Errorf("formatNumericValue(%v) returned special value: %v", tt.value, got)
				}
				// String representation should preserve precision
			}
		})
	}
}

func TestFormatValue_JSON(t *testing.T) {
	// Test JSON object
	jsonObj := `{"key": "value", "number": 42}`
	got := formatValue(jsonObj, "jsonb", driver.DatabaseTypePostgreSQL)
	if gotMap, ok := got.(map[string]interface{}); ok {
		if gotMap["key"] != "value" {
			t.Errorf("formatJSONValue(object) key mismatch")
		}
		if gotMap["number"].(float64) != 42 {
			t.Errorf("formatJSONValue(object) number mismatch")
		}
	} else {
		t.Errorf("formatJSONValue(object) = %T, want map[string]interface{}", got)
	}

	// Test JSON array
	jsonArr := `[1, 2, 3]`
	got = formatValue(jsonArr, "json", driver.DatabaseTypeMySQL)
	if gotArr, ok := got.([]interface{}); ok {
		if len(gotArr) != 3 {
			t.Errorf("formatJSONValue(array) length mismatch")
		}
	} else {
		t.Errorf("formatJSONValue(array) = %T, want []interface{}", got)
	}
}

func TestFormatValue_Blob(t *testing.T) {
	testData := []byte("Hello, World!")
	expected := base64.StdEncoding.EncodeToString(testData)

	got := formatValue(testData, "bytea", driver.DatabaseTypePostgreSQL)
	if gotStr, ok := got.(string); ok {
		if gotStr != expected {
			t.Errorf("formatBlobValue([]byte) = %v, want %v", gotStr, expected)
		}
	} else {
		t.Errorf("formatBlobValue([]byte) = %T, want string", got)
	}
}

func TestFormatValue_UUID(t *testing.T) {
	uuid := "A0EEBC99-9C0B-4EF8-BB6D-6BB9BD380A11"
	expected := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"

	got := formatValue(uuid, "uuid", driver.DatabaseTypePostgreSQL)
	if gotStr, ok := got.(string); ok {
		if gotStr != expected {
			t.Errorf("formatUUIDValue(%q) = %v, want %v", uuid, gotStr, expected)
		}
	} else {
		t.Errorf("formatUUIDValue(string) = %T, want string", got)
	}
}

func TestFormatValue_Array(t *testing.T) {
	testArray := []interface{}{"a", "b", "c"}

	got := formatValue(testArray, "array", driver.DatabaseTypePostgreSQL)
	if gotArr, ok := got.([]interface{}); ok {
		if len(gotArr) != 3 {
			t.Errorf("formatArrayValue length mismatch")
		}
	} else {
		t.Errorf("formatArrayValue = %T, want []interface{}", got)
	}
}

func TestFormatNumericValue_Precision(t *testing.T) {
	// Test large numbers without scientific notation
	largeNum := float64(123456789012345)
	got := formatNumericValue(largeNum)
	if gotStr, ok := got.(string); ok {
		if gotStr == "1.23456789012345e+14" {
			t.Errorf("formatNumericValue(%v) returned scientific notation: %v", largeNum, got)
		}
	}

	// Test that integers are returned as int64
	intVal := float64(42)
	got = formatNumericValue(intVal)
	if _, ok := got.(int64); !ok {
		t.Errorf("formatNumericValue(%v) should return int64 for integer values, got %T", intVal, got)
	}
}

func TestFormatNumericValue_SpecialCases(t *testing.T) {
	// Test NaN - note: NaN != NaN in Go, so we use math.IsNaN
	nan := math.NaN()
	got := formatNumericValue(nan)
	if !math.IsNaN(got.(float64)) {
		t.Errorf("formatNumericValue(NaN) should preserve NaN, got %v", got)
	}

	// Test Infinity
	inf := math.Inf(1)
	got = formatNumericValue(inf)
	if !math.IsInf(got.(float64), 1) {
		t.Errorf("formatNumericValue(Inf) should preserve Inf, got %v", got)
	}
}

func TestNewResultSetFromRows_Empty(t *testing.T) {
	rows := []*driver.Row{}
	result := NewResultSetFromRows(rows, time.Millisecond*100, "SELECT 1", driver.DatabaseTypePostgreSQL)

	if result.RowCount != 0 {
		t.Errorf("NewResultSetFromRows(empty) RowCount = %d, want 0", result.RowCount)
	}
	if len(result.Columns) != 0 {
		t.Errorf("NewResultSetFromRows(empty) Columns = %v, want empty", result.Columns)
	}
	if len(result.Rows) != 0 {
		t.Errorf("NewResultSetFromRows(empty) Rows should be empty")
	}
}

func TestNewResultSetFromRows_SingleRow(t *testing.T) {
	rows := []*driver.Row{
		{
			Data:        map[string]interface{}{"id": 1, "name": "Alice", "active": true},
			ColumnNames: []string{"id", "name", "active"},
		},
	}

	result := NewResultSetFromRows(rows, time.Millisecond*50, "SELECT * FROM users", driver.DatabaseTypePostgreSQL)

	if result.RowCount != 1 {
		t.Errorf("NewResultSetFromRows RowCount = %d, want 1", result.RowCount)
	}
	if len(result.Columns) != 3 {
		t.Errorf("NewResultSetFromRows Columns = %d, want 3", len(result.Columns))
	}
	if result.Statement != "SELECT * FROM users" {
		t.Errorf("NewResultSetFromRows Statement = %q, want %q", result.Statement, "SELECT * FROM users")
	}
}

func TestNewResultSetFromRows_MultipleRows(t *testing.T) {
	rows := []*driver.Row{
		{
			Data:        map[string]interface{}{"id": 1, "name": "Alice"},
			ColumnNames: []string{"id", "name"},
		},
		{
			Data:        map[string]interface{}{"id": 2, "name": "Bob"},
			ColumnNames: []string{"id", "name"},
		},
		{
			Data:        map[string]interface{}{"id": 3, "name": "Charlie"},
			ColumnNames: []string{"id", "name"},
		},
	}

	result := NewResultSetFromRows(rows, time.Millisecond*100, "SELECT * FROM users", driver.DatabaseTypePostgreSQL)

	if result.RowCount != 3 {
		t.Errorf("NewResultSetFromRows RowCount = %d, want 3", result.RowCount)
	}
	// Column-oriented: 2 columns with 3 values each
	if len(result.Rows) != 2 {
		t.Errorf("NewResultSetFromRows Rows columns = %d, want 2", len(result.Rows))
	}
	if len(result.Rows[0]) != 3 {
		t.Errorf("NewResultSetFromRows first column values = %d, want 3", len(result.Rows[0]))
	}
}

func TestNewResultSetFromRows_NULLValues(t *testing.T) {
	rows := []*driver.Row{
		{
			Data:        map[string]interface{}{"id": 1, "name": nil, "active": true},
			ColumnNames: []string{"id", "name", "active"},
		},
		{
			Data:        map[string]interface{}{"id": 2, "name": "Bob", "active": nil},
			ColumnNames: []string{"id", "name", "active"},
		},
	}

	result := NewResultSetFromRows(rows, time.Millisecond*50, "SELECT * FROM users", driver.DatabaseTypePostgreSQL)

	// Check that NULL values are preserved as nil
	if result.Rows[1][0] != nil {
		t.Errorf("NULL name in row 0 should be nil, got %v", result.Rows[1][0])
	}
	if result.Rows[2][1] != nil {
		t.Errorf("NULL active in row 1 should be nil, got %v", result.Rows[2][1])
	}
}

func TestNewResultSetFromRows_ColumnOriented(t *testing.T) {
	rows := []*driver.Row{
		{
			Data:        map[string]interface{}{"a": 1, "b": 2, "c": 3},
			ColumnNames: []string{"a", "b", "c"},
		},
		{
			Data:        map[string]interface{}{"a": 4, "b": 5, "c": 6},
			ColumnNames: []string{"a", "b", "c"},
		},
	}

	result := NewResultSetFromRows(rows, time.Millisecond*100, "SELECT * FROM test", driver.DatabaseTypePostgreSQL)

	// Column-oriented storage: each row in result.Rows is a column
	if len(result.Rows) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(result.Rows))
	}

	// First column 'a' should have values [1, 4]
	if len(result.Rows[0]) != 2 {
		t.Fatalf("Expected 2 values in first column, got %d", len(result.Rows[0]))
	}
	if fmt.Sprintf("%v", result.Rows[0][0]) != "1" || fmt.Sprintf("%v", result.Rows[0][1]) != "4" {
		t.Errorf("First column values incorrect: %v", result.Rows[0])
	}

	// Second column 'b' should have values [2, 5]
	if fmt.Sprintf("%v", result.Rows[1][0]) != "2" || fmt.Sprintf("%v", result.Rows[1][1]) != "5" {
		t.Errorf("Second column values incorrect: %v", result.Rows[1])
	}

	// Third column 'c' should have values [3, 6]
	if fmt.Sprintf("%v", result.Rows[2][0]) != "3" || fmt.Sprintf("%v", result.Rows[2][1]) != "6" {
		t.Errorf("Third column values incorrect: %v", result.Rows[2])
	}
}
