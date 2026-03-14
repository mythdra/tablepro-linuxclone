package mongodb

import (
	"context"
	"testing"

	"tablepro/internal/connection"
)

func TestNew(t *testing.T) {
	driver := New()
	if driver == nil {
		t.Error("New() returned nil")
	}
}

func TestBuildURI(t *testing.T) {
	config := &connection.DatabaseConnection{
		Host:     "localhost",
		Port:     27017,
		Database: "testdb",
		Username: "user",
	}

	uri := buildURI(config, "password")

	expected := "mongodb://user:password@localhost:27017/testdb"
	if uri != expected {
		t.Errorf("buildURI() = %v, want %v", uri, expected)
	}
}

func TestBuildURIWithoutAuth(t *testing.T) {
	config := &connection.DatabaseConnection{
		Host:     "localhost",
		Port:     27017,
		Database: "testdb",
		Username: "",
	}

	uri := buildURI(config, "")

	expected := "mongodb://localhost:27017/testdb"
	if uri != expected {
		t.Errorf("buildURI() = %v, want %v", uri, expected)
	}
}

func TestBuildURIWithSSL(t *testing.T) {
	config := &connection.DatabaseConnection{
		Host:     "localhost",
		Port:     27017,
		Database: "testdb",
		Username: "user",
		SSL: connection.SSLConfig{
			Enabled: true,
			Mode:    "require",
		},
	}

	uri := buildURI(config, "password")

	if uri == "" {
		t.Error("buildURI() returned empty string with SSL enabled")
	}
}

func TestGetStringSafe(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"string value", "hello", "hello"},
		{"int value", 123, ""},
		{"nil value", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringSafe(tt.input)
			if result != tt.expected {
				t.Errorf("getStringSafe(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetBoolSafe(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"true value", true, true},
		{"false value", false, false},
		{"string value", "true", false},
		{"nil value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBoolSafe(tt.input)
			if result != tt.expected {
				t.Errorf("getBoolSafe(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestInferBSONType(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"string", "hello", "string"},
		{"int", 123, "int"},
		{"int64", int64(123), "int"},
		{"float", 1.5, "double"},
		{"bool", true, "bool"},
		{"[]byte", []byte("test"), "binData"},
		{"[]any", []any{1, 2, 3}, "array"},
		{"map", map[string]any{"key": "value"}, "object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferBSONType(tt.input)
			if result != tt.expected {
				t.Errorf("inferBSONType(%T) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMongoDBDriver_NotConnected(t *testing.T) {
	driver := New()

	if driver.IsConnected() {
		t.Error("IsConnected() should return false when not connected")
	}

	if driver.GetDatabase() != "" {
		t.Error("GetDatabase() should return empty string when not connected")
	}
}

func TestParseCommand(t *testing.T) {
	driver := &MongoDBDriver{}
	_ = driver

	ctx := context.Background()
	_ = ctx
}
