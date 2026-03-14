package connection

import (
	"testing"
)

func TestParseConnectionURL_StandardPostgres(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    *ParsedConnection
		wantErr bool
	}{
		{
			name: "postgres full URL",
			url:  "postgres://user:pass@localhost:5432/mydb",
			want: &ParsedConnection{
				Scheme:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				Type:     TypePostgreSQL,
				Params:   map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "postgres without port",
			url:  "postgres://user:pass@localhost/mydb",
			want: &ParsedConnection{
				Scheme:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				Type:     TypePostgreSQL,
				Params:   map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "postgres without credentials",
			url:  "postgres://localhost:5432/mydb",
			want: &ParsedConnection{
				Scheme:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "",
				Password: "",
				Type:     TypePostgreSQL,
				Params:   map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "postgres without password",
			url:  "postgres://user@localhost:5432/mydb",
			want: &ParsedConnection{
				Scheme:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "user",
				Password: "",
				Type:     TypePostgreSQL,
				Params:   map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "postgres with query params",
			url:  "postgres://user:pass@localhost:5432/mydb?sslmode=require&statusColor=green",
			want: &ParsedConnection{
				Scheme:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				Type:     TypePostgreSQL,
				Params:   map[string]string{"sslmode": "require", "statusColor": "green"},
			},
			wantErr: false,
		},
		{
			name: "postgres with sslmode underscore",
			url:  "postgres://user:pass@localhost:5432/mydb?ssl_mode=require",
			want: &ParsedConnection{
				Scheme:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				Type:     TypePostgreSQL,
				Params:   map[string]string{"sslmode": "require"},
			},
			wantErr: false,
		},
		{
			name: "postgres default port",
			url:  "postgres://localhost/mydb",
			want: &ParsedConnection{
				Scheme:   "postgres",
				Host:     "localhost",
				Port:     5432,
				Database: "mydb",
				Username: "",
				Password: "",
				Type:     TypePostgreSQL,
				Params:   map[string]string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConnectionURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConnectionURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Scheme != tt.want.Scheme {
				t.Errorf("Scheme = %v, want %v", got.Scheme, tt.want.Scheme)
			}
			if got.Host != tt.want.Host {
				t.Errorf("Host = %v, want %v", got.Host, tt.want.Host)
			}
			if got.Port != tt.want.Port {
				t.Errorf("Port = %v, want %v", got.Port, tt.want.Port)
			}
			if got.Database != tt.want.Database {
				t.Errorf("Database = %v, want %v", got.Database, tt.want.Database)
			}
			if got.Username != tt.want.Username {
				t.Errorf("Username = %v, want %v", got.Username, tt.want.Username)
			}
			if got.Password != tt.want.Password {
				t.Errorf("Password = %v, want %v", got.Password, tt.want.Password)
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}
		})
	}
}

func TestParseConnectionURL_MySQL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    *ParsedConnection
		wantErr bool
	}{
		{
			name: "mysql full URL",
			url:  "mysql://user:pass@localhost:3306/mydb",
			want: &ParsedConnection{
				Scheme:   "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				Type:     TypeMySQL,
				Params:   map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "mysql without port",
			url:  "mysql://user:pass@localhost/mydb",
			want: &ParsedConnection{
				Scheme:   "mysql",
				Host:     "localhost",
				Port:     3306,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				Type:     TypeMySQL,
				Params:   map[string]string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConnectionURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConnectionURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Port != tt.want.Port {
				t.Errorf("Port = %v, want %v", got.Port, tt.want.Port)
			}
		})
	}
}

func TestParseConnectionURL_SQLite(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    *ParsedConnection
		wantErr bool
	}{
		{
			name: "sqlite URL",
			url:  "sqlite:///mydb.db",
			want: &ParsedConnection{
				Scheme:   "sqlite",
				Host:     "",
				Port:     0,
				Database: "mydb.db",
				Username: "",
				Password: "",
				Type:     TypeSQLite,
				Params:   map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "sqlite with path",
			url:  "sqlite:///path/to/mydb.db",
			want: &ParsedConnection{
				Scheme:   "sqlite",
				Host:     "",
				Port:     0,
				Database: "path/to/mydb.db",
				Username: "",
				Password: "",
				Type:     TypeSQLite,
				Params:   map[string]string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConnectionURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConnectionURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Database != tt.want.Database {
				t.Errorf("Database = %v, want %v", got.Database, tt.want.Database)
			}
		})
	}
}

func TestParseConnectionURL_MongoDB(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    *ParsedConnection
		wantErr bool
	}{
		{
			name: "mongodb full URL",
			url:  "mongodb://user:pass@localhost:27017/mydb",
			want: &ParsedConnection{
				Scheme:   "mongodb",
				Host:     "localhost",
				Port:     27017,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				Type:     TypeMongoDB,
				Params:   map[string]string{},
			},
			wantErr: false,
		},
		{
			name: "mongodb short scheme",
			url:  "mongo://user:pass@localhost:27017/mydb",
			want: &ParsedConnection{
				Scheme:   "mongo",
				Host:     "localhost",
				Port:     27017,
				Database: "mydb",
				Username: "user",
				Password: "pass",
				Type:     TypeMongoDB,
				Params:   map[string]string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConnectionURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConnectionURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}
		})
	}
}

func TestParseConnectionURL_Redis(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    *ParsedConnection
		wantErr bool
	}{
		{
			name: "redis full URL",
			url:  "redis://:pass@localhost:6379/0",
			want: &ParsedConnection{
				Scheme:   "redis",
				Host:     "localhost",
				Port:     6379,
				Database: "0",
				Username: "",
				Password: "pass",
				Type:     TypeRedis,
				Params:   map[string]string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConnectionURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConnectionURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}
		})
	}
}

func TestParseConnectionURL_SSH(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name: "postgres+ssh full URL",
			url:  "postgres+ssh://user:pass@localhost:5432/mydb",
		},
		{
			name: "postgres+ssh without port",
			url:  "postgres+ssh://user:pass@localhost/mydb",
		},
		{
			name: "postgres+ssh with query params",
			url:  "postgres+ssh://user:pass@localhost/mydb?sslmode=require&statusColor=blue",
		},
		{
			name: "mysql+ssh URL",
			url:  "mysql+ssh://user:pass@remotehost:3306/testdb",
		},
		{
			name: "redis+ssh URL",
			url:  "redis+ssh://user:pass@redishost/0",
		},
		{
			name: "mongodb+ssh URL",
			url:  "mongodb+ssh://user:pass@mgdbhost:27017/testdb",
		},
		{
			name:    "unsupported+ssh URL",
			url:     "oracle+ssh://user:pass@hosthost/db",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConnectionURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConnectionURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if !got.SSH.Enabled {
				t.Errorf("SSH.Enabled = false, want true")
			}
			if got.SSH.Host == "" {
				t.Errorf("SSH.Host = empty, want non-empty")
			}
			if got.SSH.Port != 22 {
				t.Errorf("SSH.Port = %v, want 22", got.SSH.Port)
			}
			if got.SSH.Username == "" {
				t.Errorf("SSH.Username = empty, want non-empty")
			}
		})
	}
}

func TestParseConnectionURL_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
			errMsg:  "empty URL",
		},
		{
			name:    "whitespace only",
			url:     "   ",
			wantErr: true,
			errMsg:  "empty URL",
		},
		{
			name:    "unsupported scheme",
			url:     "oracle://user:pass@localhost:1521/xe",
			wantErr: true,
			errMsg:  "unsupported database type",
		},
		{
			name:    "invalid URL format",
			url:     "not-a-valid-url",
			wantErr: true,
			errMsg:  "unsupported database type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConnectionURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConnectionURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.errMsg)
				}
			}
			if got != nil && tt.wantErr {
				t.Errorf("Expected nil result on error, got %v", got)
			}
		})
	}
}

func TestGetConnectionString(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "postgres standard",
			url:  "postgres://user:pass@localhost:5432/mydb?sslmode=require",
		},
		{
			name: "postgres ssh",
			url:  "postgres+ssh://user:pass@localhost/mydb",
		},
		{
			name: "mysql no creds",
			url:  "mysql://localhost:3306/testdb",
		},
		{
			name: "redis with db",
			url:  "redis://:pass@localhost:6379/0",
		},
		{
			name: "with multiple params",
			url:  "postgres://user@localhost:5432/mydb?a=1&b=2&c=3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseConnectionURL(tt.url)
			if err != nil {
				t.Fatalf("Failed to parse URL: %v", err)
			}
			got := parsed.GetConnectionString()
			if got == "" {
				t.Errorf("GetConnectionString() returned empty string")
			}
		})
	}
}

func TestCanonicalType(t *testing.T) {
	tests := []struct {
		scheme string
		want   DatabaseType
	}{
		{"postgres", TypePostgreSQL},
		{"postgresql", TypePostgreSQL},
		{"mysql", TypeMySQL},
		{"mongodb", TypeMongoDB},
		{"mongo", TypeMongoDB},
		{"redis", TypeRedis},
		{"sqlite", TypeSQLite},
		{"duckdb", TypeDuckDB},
		{"mssql", TypeMSSQL},
		{"sqlserver", TypeMSSQL},
		{"clickhouse", TypeClickHouse},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.scheme, func(t *testing.T) {
			got := canonicalType(tt.scheme)
			if got != tt.want {
				t.Errorf("canonicalType(%q) = %v, want %v", tt.scheme, got, tt.want)
			}
		})
	}
}

func TestIsValidScheme(t *testing.T) {
	tests := []struct {
		scheme string
		valid  bool
	}{
		{"postgres", true},
		{"postgresql", true},
		{"mysql", true},
		{"mongodb", true},
		{"mongo", true},
		{"redis", true},
		{"sqlite", true},
		{"duckdb", true},
		{"mssql", true},
		{"sqlserver", true},
		{"clickhouse", true},
		{"oracle", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.scheme, func(t *testing.T) {
			got := isValidScheme(tt.scheme)
			if got != tt.valid {
				t.Errorf("isValidScheme(%q) = %v, want %v", tt.scheme, got, tt.valid)
			}
		})
	}
}

func TestParseCredentials(t *testing.T) {
	tests := []struct {
		cred    string
		wantUsr string
		wantPw  string
	}{
		{"user:pass", "user", "pass"},
		{"user:", "user", ""},
		{"user", "user", ""},
		{"", "", ""},
		{"user:pass:with:colons", "user", "pass:with:colons"},
	}

	for _, tt := range tests {
		t.Run(tt.cred, func(t *testing.T) {
			gotUsr, gotPw := parseCredentials(tt.cred)
			if gotUsr != tt.wantUsr {
				t.Errorf("parseCredentials() username = %v, want %v", gotUsr, tt.wantUsr)
			}
			if gotPw != tt.wantPw {
				t.Errorf("parseCredentials() password = %v, want %v", gotPw, tt.wantPw)
			}
		})
	}
}

func TestParseHostPort(t *testing.T) {
	tests := []struct {
		hostPart string
		wantHost string
		wantPort int
	}{
		{"localhost:5432", "localhost", 5432},
		{"localhost", "localhost", 0},
		{"192.168.1.1:3306", "192.168.1.1", 3306},
	}

	for _, tt := range tests {
		t.Run(tt.hostPart, func(t *testing.T) {
			gotHost, gotPort := parseHostPort(tt.hostPart)
			if gotHost != tt.wantHost {
				t.Errorf("parseHostPort() host = %v, want %v", gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("parseHostPort() port = %v, want %v", gotPort, tt.wantPort)
			}
		})
	}
}

func TestExtractDatabaseWithParams(t *testing.T) {
	tests := []struct {
		path       string
		wantDB     string
		wantParams map[string]string
	}{
		{
			path:       "mydb",
			wantDB:     "mydb",
			wantParams: map[string]string{},
		},
		{
			path:       "mydb?sslmode=require",
			wantDB:     "mydb",
			wantParams: map[string]string{"sslmode": "require"},
		},
		{
			path:       "mydb?a=1&b=2",
			wantDB:     "mydb",
			wantParams: map[string]string{"a": "1", "b": "2"},
		},
		{
			path:       "mydb?empty=",
			wantDB:     "mydb",
			wantParams: map[string]string{"empty": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			params := make(map[string]string)
			got := extractDatabaseWithParams(tt.path, params)
			if got != tt.wantDB {
				t.Errorf("extractDatabaseWithParams() = %v, want %v", got, tt.wantDB)
			}
			if len(params) != len(tt.wantParams) {
				t.Errorf("params length = %v, want %v", len(params), len(tt.wantParams))
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
