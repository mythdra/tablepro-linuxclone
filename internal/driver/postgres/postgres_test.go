package postgres

import (
	"context"
	"testing"
	"time"
)

func TestConfig_PoolConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "default config",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "pass",
				SSLMode:  "disable",
			},
			wantErr: false,
		},
		{
			name: "config with SSL",
			config: &Config{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
				Username: "user",
				Password: "pass",
				SSLMode:  "disable",
			},
			wantErr: false,
		},
		{
			name: "config with pool settings",
			config: &Config{
				Host:            "localhost",
				Port:            5432,
				Database:        "testdb",
				Username:        "user",
				Password:        "pass",
				SSLMode:         "disable",
				MinConnections:  5,
				MaxConnections:  20,
				MaxConnLifetime: time.Hour,
				MaxConnIdleTime: 15 * time.Minute,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.config.PoolConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.PoolConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("Expected port 5432, got %d", cfg.Port)
	}
	if cfg.SSLMode != "prefer" {
		t.Errorf("Expected SSLMode prefer, got %s", cfg.SSLMode)
	}
}

func TestParsePGArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple array",
			input:    "{a,b,c}",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty array",
			input:    "{}",
			expected: nil,
		},
		{
			name:     "array with quotes",
			input:    `{"hello,world","test value"}`,
			expected: []string{"hello,world", "test value"},
		},
		{
			name:     "single element",
			input:    "{only}",
			expected: []string{"only"},
		},
		{
			name:     "numbers",
			input:    "{1,2,3}",
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "too short",
			input:    "a",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParsePGArray(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("ParsePGArray(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("ParsePGArray(%q)[%d] = %v, want %v", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestFormatTimestamp(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 45, 123000000, time.UTC)
	expected := "2024-01-15 10:30:45.123"
	result := FormatTimestamp(testTime)
	if result != expected {
		t.Errorf("FormatTimestamp() = %v, want %v", result, expected)
	}
}

func TestPostgreSQLDriver_NewPostgreSQLDriver(t *testing.T) {
	driver := NewPostgreSQLDriver()
	if driver == nil {
		t.Error("NewPostgreSQLDriver() returned nil")
	}
	if driver.config == nil {
		t.Error("NewPostgreSQLDriver() did not set default config")
	}
}

func TestPostgreSQLDriver_Connect_NotConnected(t *testing.T) {
	driver := NewPostgreSQLDriver()
	ctx := context.Background()
	_, err := driver.Execute(ctx, "SELECT 1")
	if err == nil {
		t.Error("Expected error when executing on non-connected driver")
	}
}

func TestPostgreSQLDriver_Close_NotConnected(t *testing.T) {
	driver := NewPostgreSQLDriver()
	err := driver.Close()
	if err != nil {
		t.Errorf("Close() on non-connected driver returned error: %v", err)
	}
}

func TestPostgreSQLDriver_Ping_NotConnected(t *testing.T) {
	driver := NewPostgreSQLDriver()
	ctx := context.Background()
	err := driver.Ping(ctx)
	if err == nil {
		t.Error("Expected error when pinging non-connected driver")
	}
}

func TestTypeMapping(t *testing.T) {
	tests := []struct {
		pgType     string
		wantMapped string
	}{
		{"int4", "integer"},
		{"int8", "bigint"},
		{"float4", "real"},
		{"timestamptz", "timestamp with time zone"},
		{"jsonb", "jsonb"},
		{"_int4", "integer[]"},
		{"unknown_type", "unknown_type"},
	}

	for _, tt := range tests {
		t.Run(tt.pgType, func(t *testing.T) {
			mapped, ok := TypeMapping[tt.pgType]
			if tt.pgType == "unknown_type" {
				if ok {
					t.Errorf("Expected no mapping for %s", tt.pgType)
				}
			} else if mapped != tt.wantMapped {
				t.Errorf("TypeMapping[%s] = %v, want %v", tt.pgType, mapped, tt.wantMapped)
			}
		})
	}
}

func TestMapCharToConstraint(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"a", "NO ACTION"},
		{"r", "RESTRICT"},
		{"c", "CASCADE"},
		{"n", "SET NULL"},
		{"d", "SET DEFAULT"},
		{"x", "x"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapCharToConstraint(tt.input)
			if result != tt.expected {
				t.Errorf("mapCharToConstraint(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
