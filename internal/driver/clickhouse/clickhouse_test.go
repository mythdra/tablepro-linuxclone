package clickhouse

import (
	"testing"
)

func TestClickHouseDriver_NewClickHouseDriver(t *testing.T) {
	driver := NewClickHouseDriver()
	if driver == nil {
		t.Error("Expected non-nil ClickHouseDriver")
	}
	if driver.config == nil {
		t.Error("Expected non-nil config")
	}
}

func TestConfig_DefaultConfigClickHouse(t *testing.T) {
	config := DefaultConfig()

	if config.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", config.Host)
	}
	if config.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", config.Port)
	}
	if config.SSLMode != "disable" {
		t.Errorf("Expected SSL mode 'disable', got '%s'", config.SSLMode)
	}
	if config.Compress != "lz4" {
		t.Errorf("Expected compress 'lz4', got '%s'", config.Compress)
	}
}

func TestConfig_DSNClickHouse(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "standard",
			config: &Config{
				Host:     "localhost",
				Port:     9000,
				Database: "default",
				Username: "default",
				Password: "",
			},
			expected: "clickhouse://default:@localhost:9000/default",
		},
		{
			name: "with password",
			config: &Config{
				Host:     "localhost",
				Port:     9000,
				Database: "mydb",
				Username: "user",
				Password: "pass123",
			},
			expected: "clickhouse://user:pass123@localhost:9000/mydb",
		},
		{
			name: "with SSL",
			config: &Config{
				Host:     "localhost",
				Port:     9440,
				Database: "secure",
				Username: "user",
				Password: "pass",
				SSLMode:  "enable",
			},
			expected: "clickhouse://user:pass@localhost:9440/secure?ssl=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := tt.config.DSN()
			if dsn != tt.expected {
				t.Errorf("Expected DSN:\n%s\nGot:\n%s", tt.expected, dsn)
			}
		})
	}
}

func TestClickHouseTypeMapping(t *testing.T) {
	expectedMappings := map[string]string{
		"Int8":           "int32",
		"Int16":          "int32",
		"Int32":          "int32",
		"Int64":          "int64",
		"UInt8":          "uint32",
		"UInt64":         "uint64",
		"Float32":        "float32",
		"Float64":        "float64",
		"String":         "string",
		"FixedString":    "string",
		"Date":           "time.Time",
		"DateTime":       "time.Time",
		"DateTime64":     "time.Time",
		"UUID":           "string",
		"Array":          "[]any",
		"Map":            "map[string]any",
		"JSON":           "string",
		"Enum":           "string",
		"LowCardinality": "string",
	}

	for dbType, expectedGoType := range expectedMappings {
		if goType, ok := TypeMapping[dbType]; !ok {
			t.Errorf("Missing type mapping for %s", dbType)
		} else if goType != expectedGoType {
			t.Errorf("For %s, expected %s, got %s", dbType, expectedGoType, goType)
		}
	}
}

func TestClickHouseDriver_GetCapabilities(t *testing.T) {
	driver := NewClickHouseDriver()
	caps := driver.GetCapabilities()

	if caps == nil {
		t.Error("Expected non-nil capabilities")
	}

	if caps.MaxConnections != 10 {
		t.Errorf("Expected MaxConnections 10, got %d", caps.MaxConnections)
	}

	if caps.SupportsTransactions {
		t.Error("Expected SupportsTransactions to be false")
	}

	if caps.SupportsStoredProcedures {
		t.Error("Expected SupportsStoredProcedures to be false")
	}

	if !caps.SupportsFunctions {
		t.Error("Expected SupportsFunctions to be true")
	}

	if !caps.SupportsViews {
		t.Error("Expected SupportsViews to be true")
	}

	if !caps.SupportsMaterializedViews {
		t.Error("Expected SupportsMaterializedViews to be true")
	}

	if caps.SupportsForeignKeys {
		t.Error("Expected SupportsForeignKeys to be false")
	}

	if !caps.SupportsSchemas {
		t.Error("Expected SupportsSchemas to be true")
	}

	if caps.SupportsAutoIncrement {
		t.Error("Expected SupportsAutoIncrement to be false")
	}
}
