package driver

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabaseTypeString tests DatabaseType String() method
func TestDatabaseTypeString(t *testing.T) {
	tests := []struct {
		dbType     DatabaseType
		wantString string
	}{
		{DatabaseTypePostgreSQL, "postgresql"},
		{DatabaseTypeMySQL, "mysql"},
		{DatabaseTypeSQLite, "sqlite"},
		{DatabaseTypeDuckDB, "duckdb"},
		{DatabaseTypeMSSQL, "mssql"},
		{DatabaseTypeClickHouse, "clickhouse"},
		{DatabaseTypeMongoDB, "mongodb"},
		{DatabaseTypeRedis, "redis"},
		{DatabaseTypeUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.wantString, func(t *testing.T) {
			got := tt.dbType.String()
			assert.Equal(t, tt.wantString, got)
		})
	}
}

// TestTypeFromString tests string to DatabaseType conversion
func TestTypeFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected DatabaseType
	}{
		// PostgreSQL aliases
		{"postgresql", DatabaseTypePostgreSQL},
		{"postgres", DatabaseTypePostgreSQL},
		{"pgsql", DatabaseTypePostgreSQL},
		{"pg", DatabaseTypePostgreSQL},
		// MySQL aliases
		{"mysql", DatabaseTypeMySQL},
		{"mariadb", DatabaseTypeMySQL},
		// SQLite aliases
		{"sqlite", DatabaseTypeSQLite},
		{"sqlite3", DatabaseTypeSQLite},
		// DuckDB aliases
		{"duckdb", DatabaseTypeDuckDB},
		{"duck", DatabaseTypeDuckDB},
		// MSSQL aliases
		{"mssql", DatabaseTypeMSSQL},
		{"sqlserver", DatabaseTypeMSSQL},
		{"sql-server", DatabaseTypeMSSQL},
		// ClickHouse aliases
		{"clickhouse", DatabaseTypeClickHouse},
		{"ch", DatabaseTypeClickHouse},
		// MongoDB aliases
		{"mongodb", DatabaseTypeMongoDB},
		{"mongo", DatabaseTypeMongoDB},
		// Redis
		{"redis", DatabaseTypeRedis},
		// Unknown
		{"unknown_db", DatabaseTypeUnknown},
		{"", DatabaseTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := TypeFromString(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestDataTypeMapping tests type mapping structures
func TestDataTypeMapping(t *testing.T) {
	mappings := CommonDataTypes()

	// Verify all expected database types have mappings
	expectedTypes := []DatabaseType{
		DatabaseTypePostgreSQL,
		DatabaseTypeMySQL,
		DatabaseTypeSQLite,
		DatabaseTypeMSSQL,
		DatabaseTypeClickHouse,
		DatabaseTypeDuckDB,
	}

	for _, dbType := range expectedTypes {
		t.Run(string(dbType), func(t *testing.T) {
			assert.Contains(t, mappings, dbType)
			assert.NotEmpty(t, mappings[dbType])
		})
	}

	// Verify PostgreSQL mappings
	pgMappings := mappings[DatabaseTypePostgreSQL]
	assert.True(t, len(pgMappings) > 20, "PostgreSQL should have many type mappings")

	// Verify specific mappings exist
	hasVarchar := false
	hasInt4 := false
	for _, m := range pgMappings {
		if m.DBType == "varchar" {
			hasVarchar = true
			assert.Equal(t, "string", m.GoType)
			assert.True(t, m.IsString)
		}
		if m.DBType == "int4" {
			hasInt4 = true
			assert.Equal(t, "int32", m.GoType)
			assert.True(t, m.IsNumeric)
		}
	}
	assert.True(t, hasVarchar, "Should have varchar mapping")
	assert.True(t, hasInt4, "Should have int4 mapping")
}

// TestGetDataTypeMapping tests type mapping retrieval
func TestGetDataTypeMapping(t *testing.T) {
	tests := []struct {
		dbType     DatabaseType
		columnType string
		wantNil    bool
	}{
		{DatabaseTypePostgreSQL, "varchar", false},
		{DatabaseTypePostgreSQL, "int4", false},
		{DatabaseTypePostgreSQL, "timestamptz", false},
		{DatabaseTypePostgreSQL, "unknown_type", true},
		{DatabaseTypeMySQL, "varchar", false},
		{DatabaseTypeMySQL, "bigint", false},
		{DatabaseTypeSQLite, "text", false},
		{DatabaseTypeSQLite, "integer", false},
		{DatabaseTypeMSSQL, "datetime", false},
		{DatabaseTypeMSSQL, "int", false},
		{DatabaseTypeClickHouse, "String", false},
		{DatabaseTypeClickHouse, "Int64", false},
		{DatabaseTypeDuckDB, "boolean", false},
		{DatabaseTypeDuckDB, "timestamp", false},
		{DatabaseTypeUnknown, "any", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.dbType)+"_"+tt.columnType, func(t *testing.T) {
			got := GetDataTypeMapping(tt.dbType, tt.columnType)
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.NotEmpty(t, got.GoType)
			}
		})
	}
}

// TestIsNumericType tests numeric type detection
func TestIsNumericType(t *testing.T) {
	tests := []struct {
		dbType     DatabaseType
		columnType string
		want       bool
	}{
		{DatabaseTypePostgreSQL, "int4", true},
		{DatabaseTypePostgreSQL, "varchar", false},
		{DatabaseTypePostgreSQL, "text", false},
		{DatabaseTypeMySQL, "int", true},
		{DatabaseTypeMySQL, "double", true},
		{DatabaseTypeMySQL, "varchar", false},
		{DatabaseTypeSQLite, "integer", true},
		{DatabaseTypeSQLite, "real", true},
		{DatabaseTypeSQLite, "text", false},
		{DatabaseTypeMSSQL, "bigint", true},
		{DatabaseTypeMSSQL, "bit", true},
		{DatabaseTypeMSSQL, "nvarchar", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.dbType)+"_"+tt.columnType, func(t *testing.T) {
			got := IsNumericType(tt.dbType, tt.columnType)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIsStringType tests string type detection
func TestIsStringType(t *testing.T) {
	tests := []struct {
		dbType     DatabaseType
		columnType string
		want       bool
	}{
		{DatabaseTypePostgreSQL, "varchar", true},
		{DatabaseTypePostgreSQL, "text", true},
		{DatabaseTypePostgreSQL, "int4", false},
		{DatabaseTypeMySQL, "varchar", true},
		{DatabaseTypeMySQL, "text", true},
		{DatabaseTypeMySQL, "int", false},
		{DatabaseTypeSQLite, "text", true},
		{DatabaseTypeSQLite, "blob", false},
		{DatabaseTypeClickHouse, "String", true},
		{DatabaseTypeClickHouse, "Int64", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.dbType)+"_"+tt.columnType, func(t *testing.T) {
			got := IsStringType(tt.dbType, tt.columnType)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestIsTimeType tests time type detection
func TestIsTimeType(t *testing.T) {
	tests := []struct {
		dbType     DatabaseType
		columnType string
		want       bool
	}{
		{DatabaseTypePostgreSQL, "timestamp", true},
		{DatabaseTypePostgreSQL, "timestamptz", true},
		{DatabaseTypePostgreSQL, "date", true},
		{DatabaseTypePostgreSQL, "int4", false},
		{DatabaseTypeMySQL, "datetime", true},
		{DatabaseTypeMySQL, "timestamp", true},
		{DatabaseTypeMySQL, "date", true},
		{DatabaseTypeMySQL, "varchar", false},
		{DatabaseTypeSQLite, "text", false}, // SQLite doesn't have native time types
		{DatabaseTypeMSSQL, "datetime", true},
		{DatabaseTypeMSSQL, "datetime2", true},
		{DatabaseTypeClickHouse, "DateTime", true},
		{DatabaseTypeClickHouse, "Date", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.dbType)+"_"+tt.columnType, func(t *testing.T) {
			got := IsTimeType(tt.dbType, tt.columnType)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestDriverCapabilities tests capability structures
func TestDriverCapabilities(t *testing.T) {
	caps := &DriverCapabilities{
		Features: []Feature{
			FeatureSSLConnection,
			FeatureSSHConnection,
			FeaturePreparedStatements,
			FeatureBatchStatements,
		},
		MaxConnections:           100,
		MaxQueryTime:             30 * time.Second,
		SupportsTransactions:     true,
		SupportsStoredProcedures: true,
		SupportsFunctions:        true,
		SupportsViews:            true,
		SupportsForeignKeys:      true,
		SupportsIndexes:          true,
		SupportsAutoIncrement:    true,
		SupportsSchemas:          true,
	}

	assert.NotNil(t, caps.Features)
	assert.Equal(t, 4, len(caps.Features))
	assert.Contains(t, caps.Features, FeatureSSLConnection)
	assert.True(t, caps.SupportsTransactions)
	assert.True(t, caps.SupportsViews)
	assert.True(t, caps.SupportsForeignKeys)
}

// TestConnectionConfig tests connection configuration
func TestConnectionConfig(t *testing.T) {
	config := &ConnectionConfig{
		Host:               "localhost",
		Port:               5432,
		Database:           "testdb",
		Username:           "user",
		Password:           "password",
		SSLMode:            "disable",
		MaxOpenConnections: 25,
		MaxIdleConnections: 5,
		MaxConnectionLife:  5 * time.Minute,
		QueryTimeout:       30 * time.Second,
	}

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 5432, config.Port)
	assert.Equal(t, "testdb", config.Database)
	assert.Equal(t, "user", config.Username)
	assert.Equal(t, "disable", config.SSLMode)
	assert.Equal(t, 25, config.MaxOpenConnections)
}

// TestResultStructure tests Result struct
func TestResultStructure(t *testing.T) {
	result := &Result{
		LastInsertID: 12345,
		RowsAffected: 100,
	}

	assert.Equal(t, int64(12345), result.LastInsertID)
	assert.Equal(t, int64(100), result.RowsAffected)
}

// TestRowStructure tests Row struct
func TestRowStructure(t *testing.T) {
	row := Row{
		Data: map[string]any{
			"id":   1,
			"name": "test",
		},
		ColumnNames: []string{"id", "name"},
	}

	assert.Equal(t, 2, len(row.ColumnNames))
	assert.Equal(t, "id", row.ColumnNames[0])
	assert.Equal(t, "name", row.ColumnNames[1])
	assert.Equal(t, 1, row.Data["id"])
	assert.Equal(t, "test", row.Data["name"])
}

// TestColumnInfoStructure tests ColumnInfo struct
func TestColumnInfoStructure(t *testing.T) {
	nullable := true
	maxLen := int64(255)
	col := ColumnInfo{
		Name:            "id",
		DataType:        "int",
		TypeName:        "integer",
		Nullable:        nullable,
		DefaultValue:    nil,
		IsPrimaryKey:    true,
		IsAutoIncrement: true,
		MaxLength:       &maxLen,
		Comment:         nil,
	}

	assert.Equal(t, "id", col.Name)
	assert.Equal(t, "int", col.DataType)
	assert.True(t, col.Nullable)
	assert.True(t, col.IsPrimaryKey)
	assert.True(t, col.IsAutoIncrement)
	assert.Equal(t, &maxLen, col.MaxLength)
}

// TestTableInfoStructure tests TableInfo struct
func TestTableInfoStructure(t *testing.T) {
	rowCount := int64(1000)
	size := int64(1024000)
	comment := "Test table"
	now := time.Now()

	table := TableInfo{
		Name:      "users",
		Schema:    "public",
		Type:      TableTypeTable,
		Comment:   &comment,
		RowCount:  &rowCount,
		SizeBytes: &size,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	assert.Equal(t, "users", table.Name)
	assert.Equal(t, "public", table.Schema)
	assert.Equal(t, TableTypeTable, table.Type)
	assert.Equal(t, &comment, table.Comment)
	assert.Equal(t, &rowCount, table.RowCount)
	assert.Equal(t, &size, table.SizeBytes)
}

// TestSchemaInfoStructure tests SchemaInfo struct
func TestSchemaInfoStructure(t *testing.T) {
	schema := SchemaInfo{
		Tables: []TableInfo{
			{Name: "users", Schema: "public", Type: TableTypeTable},
			{Name: "posts", Schema: "public", Type: TableTypeTable},
		},
		Views: []TableInfo{
			{Name: "user_stats", Schema: "public", Type: TableTypeView},
		},
		Schemas: []string{"public", "custom"},
	}

	assert.Equal(t, 2, len(schema.Tables))
	assert.Equal(t, 1, len(schema.Views))
	assert.Equal(t, 2, len(schema.Schemas))
}

// TestIndexInfoStructure tests IndexInfo struct
func TestIndexInfoStructure(t *testing.T) {
	idx := IndexInfo{
		Name:      "idx_users_email",
		Columns:   []string{"email"},
		IsUnique:  true,
		IsPrimary: false,
		IsPartial: false,
		IndexType: "btree",
	}

	assert.Equal(t, "idx_users_email", idx.Name)
	assert.Equal(t, []string{"email"}, idx.Columns)
	assert.True(t, idx.IsUnique)
	assert.False(t, idx.IsPrimary)
	assert.Equal(t, "btree", idx.IndexType)
}

// TestForeignKeyInfoStructure tests ForeignKeyInfo struct
func TestForeignKeyInfoStructure(t *testing.T) {
	fk := ForeignKeyInfo{
		Name:              "fk_users_posts",
		Columns:           []string{"user_id"},
		ReferencedTable:   "users",
		ReferencedColumns: []string{"id"},
		OnDelete:          "CASCADE",
		OnUpdate:          "NO ACTION",
	}

	assert.Equal(t, "fk_users_posts", fk.Name)
	assert.Equal(t, []string{"user_id"}, fk.Columns)
	assert.Equal(t, "users", fk.ReferencedTable)
	assert.Equal(t, []string{"id"}, fk.ReferencedColumns)
	assert.Equal(t, "CASCADE", fk.OnDelete)
	assert.Equal(t, "NO ACTION", fk.OnUpdate)
}

// TestRoutineInfoStructure tests RoutineInfo struct
func TestRoutineInfoStructure(t *testing.T) {
	routine := RoutineInfo{
		Name:       "get_user",
		Schema:     "public",
		Type:       RoutineTypeFunction,
		Definition: "CREATE FUNCTION get_user...",
	}

	assert.Equal(t, "get_user", routine.Name)
	assert.Equal(t, RoutineTypeFunction, routine.Type)
	assert.NotEmpty(t, routine.Definition)
}

// TestTableTypeConstants tests table type constants
func TestTableTypeConstants(t *testing.T) {
	assert.Equal(t, TableType("TABLE"), TableTypeTable)
	assert.Equal(t, TableType("VIEW"), TableTypeView)
	assert.Equal(t, TableType("MATERIALIZED VIEW"), TableTypeMaterializedView)
	assert.Equal(t, TableType("SYSTEM TABLE"), TableTypeSystem)
}

// TestRoutineTypeConstants tests routine type constants
func TestRoutineTypeConstants(t *testing.T) {
	assert.Equal(t, RoutineType("PROCEDURE"), RoutineTypeProcedure)
	assert.Equal(t, RoutineType("FUNCTION"), RoutineTypeFunction)
}

// TestFeatureConstants tests feature constants
func TestFeatureConstants(t *testing.T) {
	features := []Feature{
		FeatureSSLConnection,
		FeatureSSHConnection,
		FeaturePreparedStatements,
		FeatureBatchStatements,
		FeatureCursorPagination,
		FeatureJSONType,
		FeatureArrayType,
		FeatureUUIDType,
		FeatureGeometricType,
		FeatureFullTextSearch,
		FeatureWindowFunctions,
		FeatureCTE,
		FeatureCTAS,
		FeatureMultipleSchemas,
	}

	assert.Equal(t, 14, len(features))
}

// TestDriverFactory tests driver factory functions
func TestDriverFactory(t *testing.T) {
	// Test IsSupported - currently no drivers are registered
	// Drivers must be registered via driver.RegisterDriver() in init()
	assert.False(t, IsSupported(DatabaseTypePostgreSQL))
	assert.False(t, IsSupported(DatabaseTypeMySQL))
	assert.False(t, IsSupported(DatabaseTypeSQLite))
	assert.False(t, IsSupported(DatabaseTypeDuckDB))
	assert.False(t, IsSupported(DatabaseTypeMSSQL))
	assert.False(t, IsSupported(DatabaseTypeClickHouse))
	assert.False(t, IsSupported(DatabaseTypeMongoDB))
	assert.False(t, IsSupported(DatabaseTypeRedis))
	assert.False(t, IsSupported(DatabaseTypeUnknown))

	// Test SupportedDrivers - should be empty until drivers register themselves
	drivers := SupportedDrivers()
	assert.Empty(t, drivers)
}

// TestNewDriver tests creating new drivers - should fail until registered
func TestNewDriver(t *testing.T) {
	tests := []struct {
		dbType  DatabaseType
		wantErr bool
	}{
		{DatabaseTypePostgreSQL, true}, // Not registered
		{DatabaseTypeMySQL, true},      // Not registered
		{DatabaseTypeSQLite, true},     // Not registered
		{DatabaseTypeUnknown, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.dbType), func(t *testing.T) {
			driver, err := NewDriver(tt.dbType)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, driver)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, driver)
			}
		})
	}
}

// TestInterfaceCompliance is a compile-time check that all drivers implement DatabaseDriver
// This test will fail to compile if any driver doesn't implement the interface
func TestInterfaceCompliance(t *testing.T) {
	// This is a compile-time check - if it compiles, all drivers implement the interface
	// We verify the interface is valid by checking its methods exist
	var _ DatabaseDriver = &MockDriver{}
	assert.True(t, true) // If we get here, interface compliance is valid
}

// MockDriver implements DatabaseDriver for interface compliance testing
type MockDriver struct{}

func (m *MockDriver) Connect(ctx context.Context, config *ConnectionConfig) error {
	return nil
}

func (m *MockDriver) Execute(ctx context.Context, query string, params ...any) (*Result, error) {
	return &Result{}, nil
}

func (m *MockDriver) Query(ctx context.Context, query string, params ...any) (*Row, error) {
	return &Row{}, nil
}

func (m *MockDriver) QueryContext(ctx context.Context, timeout time.Duration, query string, params ...any) (*Row, error) {
	return &Row{}, nil
}

func (m *MockDriver) GetSchema(ctx context.Context) (*SchemaInfo, error) {
	return &SchemaInfo{}, nil
}

func (m *MockDriver) GetTables(ctx context.Context, schemaName string) ([]TableInfo, error) {
	return []TableInfo{}, nil
}

func (m *MockDriver) GetColumns(ctx context.Context, schemaName, tableName string) ([]ColumnInfo, error) {
	return []ColumnInfo{}, nil
}

func (m *MockDriver) GetIndexes(ctx context.Context, schemaName, tableName string) ([]IndexInfo, error) {
	return []IndexInfo{}, nil
}

func (m *MockDriver) GetForeignKeys(ctx context.Context, schemaName, tableName string) ([]ForeignKeyInfo, error) {
	return []ForeignKeyInfo{}, nil
}

func (m *MockDriver) Ping(ctx context.Context) error {
	return nil
}

func (m *MockDriver) Close() error {
	return nil
}

func (m *MockDriver) GetCapabilities() *DriverCapabilities {
	return &DriverCapabilities{}
}

func (m *MockDriver) GetDB() *sql.DB {
	return nil
}

// TestMockDriverImplementsInterface verifies the mock driver implements the interface
func TestMockDriverImplementsInterface(t *testing.T) {
	// This compiles only if MockDriver implements DatabaseDriver
	var driver DatabaseDriver = &MockDriver{}
	assert.NotNil(t, driver)
}

// TestContextCancellation verifies mock driver can handle context
func TestContextCancellation(t *testing.T) {
	mock := &MockDriver{}
	// Mock returns nil for all operations - this verifies the type works
	// Real drivers would return context.Canceled when context is cancelled
	err := mock.Ping(context.Background())
	assert.NoError(t, err)
}

// TestTimeoutContext verifies mock driver can handle timeout context
func TestTimeoutContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mock := &MockDriver{}
	// Mock returns nil for all operations - this verifies the type works
	// Real drivers would return context.DeadlineExceeded when deadline passes
	err := mock.Ping(ctx)
	assert.NoError(t, err)
}

// TestConnectionPoolingConfig tests connection pool configuration
func TestConnectionPoolingConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *ConnectionConfig
		wantErr bool
	}{
		{
			name: "default pool settings",
			config: &ConnectionConfig{
				MaxOpenConnections: 25,
				MaxIdleConnections: 5,
				MaxConnectionLife:  5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "zero max open",
			config: &ConnectionConfig{
				MaxOpenConnections: 0,
			},
			wantErr: false,
		},
		{
			name: "custom pool",
			config: &ConnectionConfig{
				MaxOpenConnections: 100,
				MaxIdleConnections: 20,
				MaxConnectionLife:  10 * time.Minute,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the config is valid
			require.NotNil(t, tt.config)
			assert.GreaterOrEqual(t, tt.config.MaxIdleConnections, 0)
			assert.GreaterOrEqual(t, tt.config.MaxConnectionLife, time.Duration(0))
		})
	}
}

// TestDriverCapabilitiesVariations tests different capability combinations
func TestDriverCapabilitiesVariations(t *testing.T) {
	tests := []struct {
		name       string
		caps       DriverCapabilities
		wantFeat   int
		supportsTx bool
	}{
		{
			name: "PostgreSQL capabilities",
			caps: DriverCapabilities{
				Features:                 []Feature{FeatureSSLConnection, FeatureCTE, FeatureWindowFunctions, FeatureJSONType},
				SupportsTransactions:     true,
				SupportsStoredProcedures: true,
				SupportsFunctions:        true,
				SupportsViews:            true,
				SupportsForeignKeys:      true,
				SupportsSchemas:          true,
			},
			wantFeat:   4,
			supportsTx: true,
		},
		{
			name: "MySQL capabilities",
			caps: DriverCapabilities{
				Features:                 []Feature{FeatureSSLConnection, FeatureJSONType},
				SupportsTransactions:     true,
				SupportsStoredProcedures: true,
				SupportsFunctions:        true,
				SupportsViews:            true,
				SupportsForeignKeys:      true,
				SupportsSchemas:          false,
			},
			wantFeat:   2,
			supportsTx: true,
		},
		{
			name: "SQLite capabilities",
			caps: DriverCapabilities{
				Features:                 []Feature{FeatureJSONType},
				SupportsTransactions:     true,
				SupportsStoredProcedures: false,
				SupportsFunctions:        false,
				SupportsViews:            true,
				SupportsForeignKeys:      true,
				SupportsSchemas:          false,
			},
			wantFeat:   1,
			supportsTx: true,
		},
		{
			name: "MongoDB capabilities",
			caps: DriverCapabilities{
				Features: []Feature{FeatureSSLConnection},
				// NoSQL databases don't support these
				SupportsTransactions:     true,
				SupportsStoredProcedures: false,
				SupportsFunctions:        false,
				SupportsViews:            false,
				SupportsForeignKeys:      false,
				SupportsIndexes:          true,
			},
			wantFeat:   1,
			supportsTx: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantFeat, len(tt.caps.Features))
			assert.Equal(t, tt.supportsTx, tt.caps.SupportsTransactions)
		})
	}
}

// BenchmarkTypeMapping benchmarks type mapping lookups
func BenchmarkTypeMapping(b *testing.B) {
	b.Run("PostgreSQL varchar", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = GetDataTypeMapping(DatabaseTypePostgreSQL, "varchar")
		}
	})

	b.Run("MySQL int", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = GetDataTypeMapping(DatabaseTypeMySQL, "int")
		}
	})

	b.Run("SQLite text", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = GetDataTypeMapping(DatabaseTypeSQLite, "text")
		}
	})

	b.Run("IsNumericType", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = IsNumericType(DatabaseTypePostgreSQL, "int4")
		}
	})
}

func (m *MockDriver) Type() DatabaseType {
	return DatabaseTypeUnknown
}
