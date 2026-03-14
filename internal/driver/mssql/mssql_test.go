package mssql

import (
	"testing"
)

func TestMSSQLDriver_NewMSSQLDriver(t *testing.T) {
	driver := NewMSSQLDriver()
	if driver == nil {
		t.Error("Expected non-nil MSSQLDriver")
	}
	if driver.config == nil {
		t.Error("Expected non-nil config")
	}
}

func TestConfig_DefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", config.Host)
	}
	if config.Port != 1433 {
		t.Errorf("Expected port 1433, got %d", config.Port)
	}
	if config.Encrypt != "false" {
		t.Errorf("Expected encrypt 'false', got '%s'", config.Encrypt)
	}
	if config.TrustServerCert != "true" {
		t.Errorf("Expected trustservercertificate 'true', got '%s'", config.TrustServerCert)
	}
}

func TestConfig_DSN(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "standard SQL auth",
			config: &Config{
				Host:            "localhost",
				Port:            1433,
				Database:        "testdb",
				Username:        "sa",
				Password:        "password",
				Encrypt:         "false",
				TrustServerCert: "true",
			},
			expected: "server=localhost;port=1433;database=testdb;user id=sa;password=password;encrypt=false;trustservercertificate=true",
		},
		{
			name: "with app name",
			config: &Config{
				Host:            "localhost",
				Port:            1433,
				Database:        "testdb",
				Username:        "sa",
				Password:        "password",
				Encrypt:         "false",
				TrustServerCert: "true",
				AppName:         "TablePro",
			},
			expected: "server=localhost;port=1433;database=testdb;user id=sa;password=password;encrypt=false;trustservercertificate=true;app name=TablePro",
		},
		{
			name: "windows auth",
			config: &Config{
				Host:            "localhost",
				Port:            1433,
				Database:        "testdb",
				Encrypt:         "false",
				TrustServerCert: "true",
				UseWindowsAuth:  true,
			},
			expected: "server=localhost;port=1433;database=testdb;encrypt=false;trustservercertificate=true;authenticator=windows",
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

func TestTypeMapping(t *testing.T) {
	expectedMappings := map[string]string{
		"bigint":           "int64",
		"binary":           "[]byte",
		"bit":              "bool",
		"char":             "string",
		"datetime":         "time.Time",
		"datetime2":        "time.Time",
		"int":              "int32",
		"nvarchar":         "string",
		"uniqueidentifier": "string",
		"varbinary":        "[]byte",
		"varchar":          "string",
		"xml":              "string",
	}

	for dbType, expectedGoType := range expectedMappings {
		if goType, ok := TypeMapping[dbType]; !ok {
			t.Errorf("Missing type mapping for %s", dbType)
		} else if goType != expectedGoType {
			t.Errorf("For %s, expected %s, got %s", dbType, expectedGoType, goType)
		}
	}
}

func TestMSSQLDriver_GetCapabilities(t *testing.T) {
	driver := NewMSSQLDriver()
	caps := driver.GetCapabilities()

	if caps == nil {
		t.Error("Expected non-nil capabilities")
	}

	if caps.MaxConnections != 10 {
		t.Errorf("Expected MaxConnections 10, got %d", caps.MaxConnections)
	}

	if !caps.SupportsTransactions {
		t.Error("Expected SupportsTransactions to be true")
	}

	if !caps.SupportsStoredProcedures {
		t.Error("Expected SupportsStoredProcedures to be true")
	}

	if !caps.SupportsIndexes {
		t.Error("Expected SupportsIndexes to be true")
	}

	if !caps.SupportsForeignKeys {
		t.Error("Expected SupportsForeignKeys to be true")
	}

	if !caps.SupportsSchemas {
		t.Error("Expected SupportsSchemas to be true")
	}

	if caps.SupportsMaterializedViews {
		t.Error("Expected SupportsMaterializedViews to be false")
	}
}
