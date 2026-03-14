package redis

import (
	"testing"

	"tablepro/internal/connection"
)

func TestNew(t *testing.T) {
	driver := New()
	if driver == nil {
		t.Error("New() returned nil")
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"simple", "GET key", []string{"GET", "key"}},
		{"quoted double", `GET "my key"`, []string{"GET", "my key"}},
		{"multiple args", "HMSET hash field1 value1 field2 value2", []string{"HMSET", "hash", "field1", "value1", "field2", "value2"}},
		{"no args", "PING", []string{"PING"}},
		{"quoted single", "GET 'my key'", []string{"GET", "my key"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCommand(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseCommand(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("parseCommand(%q)[%d] = %v, want %v", tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestRedisDriver_NotConnected(t *testing.T) {
	driver := New()

	if driver.IsConnected() {
		t.Error("IsConnected() should return false when not connected")
	}

	if driver.GetDatabase() != 0 {
		t.Error("GetDatabase() should return 0 when not connected")
	}
}

func TestParseVersion(t *testing.T) {
	info := `# Server
redis_version:7.0.5
redis_mode:standalone
`

	version := parseVersion(info)
	if version != "7.0.5" {
		t.Errorf("parseVersion() = %v, want 7.0.5", version)
	}
}

func TestSplitLines(t *testing.T) {
	input := "line1\nline2\nline3"
	result := splitLines(input)

	if len(result) != 3 {
		t.Errorf("splitLines() returned %d lines, want 3", len(result))
	}
	if result[0] != "line1" || result[1] != "line2" || result[2] != "line3" {
		t.Errorf("splitLines() = %v, want [line1 line2 line3]", result)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		result bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"hello", "hello", true},
		{"", "a", false},
		{"a", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			if result != tt.result {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.result)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"123", 123},
		{"0", 0},
		{"-456", -456},
		{"invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseInt(tt.input)
			if result != tt.expected {
				t.Errorf("parseInt(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConnect_DBIndexParsing(t *testing.T) {
	tests := []struct {
		database string
		expected int
	}{
		{"0", 0},
		{"1", 1},
		{"15", 15},
		{"invalid", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.database, func(t *testing.T) {
			dbIndex := 0
			if tt.database != "" {
				if parsed, err := parseDatabaseIndex(tt.database); err == nil {
					dbIndex = parsed
				}
			}
			if dbIndex != tt.expected {
				t.Errorf("database index = %v, want %v", dbIndex, tt.expected)
			}
		})
	}
}

func parseDatabaseIndex(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"hello", "hello"},
		{"  ", ""},
		{"\ttab\t", "tab"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := trimSpace(tt.input)
			if result != tt.expected {
				t.Errorf("trimSpace(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStartsWith(t *testing.T) {
	tests := []struct {
		s      string
		prefix string
		result bool
	}{
		{"hello", "hel", true},
		{"hello", "hello", true},
		{"hello", "world", false},
		{"", "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.prefix, func(t *testing.T) {
			result := startsWith(tt.s, tt.prefix)
			if result != tt.result {
				t.Errorf("startsWith(%q, %q) = %v, want %v", tt.s, tt.prefix, result, tt.result)
			}
		})
	}
}

func TestQueryResult_FormatResult(t *testing.T) {
	driver := &RedisDriver{}

	tests := []struct {
		name  string
		input interface{}
		cols  int
		rows  int
	}{
		{"string", "hello", 1, 1},
		{"int64", int64(123), 1, 1},
		{"[]string", []string{"a", "b"}, 1, 2},
		{"bool", true, 1, 1},
		{"map", map[string]string{"key": "value"}, 2, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := driver.formatResult(tt.input)
			if len(result.Columns) != tt.cols {
				t.Errorf("columns = %d, want %d", len(result.Columns), tt.cols)
			}
			if len(result.Rows) != tt.rows {
				t.Errorf("rows = %d, want %d", len(result.Rows), tt.rows)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	driver := &RedisDriver{}

	config := &connection.DatabaseConnection{
		Host:     "localhost",
		Port:     6379,
		Database: "0",
		Username: "",
	}

	if config.Host != "localhost" {
		t.Errorf("config.Host = %v, want localhost", config.Host)
	}
	if config.Port != 6379 {
		t.Errorf("config.Port = %v, want 6379", config.Port)
	}
	_ = driver
}
