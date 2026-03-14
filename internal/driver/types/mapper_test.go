package types

import (
	"testing"
)

func TestTypeMapper_MapDatabaseType_PostgreSQL(t *testing.T) {
	tests := []struct {
		name     string
		dbType   string
		wantGo   string
		wantTS   string
		wantJSON bool
		wantTime bool
		wantNum  bool
	}{
		{"bool", "bool", "bool", "boolean", false, false, true},
		{"int4", "int4", "int32", "number", false, false, true},
		{"int8", "int8", "int64", "number", false, false, true},
		{"float4", "float4", "float32", "number", false, false, true},
		{"float8", "float8", "float64", "number", false, false, true},
		{"numeric", "numeric", "float64", "number", false, false, true},
		{"varchar", "varchar", "string", "string", false, false, false},
		{"text", "text", "string", "string", false, false, false},
		{"json", "json", "string", "any", true, false, false},
		{"jsonb", "jsonb", "[]byte", "any", true, false, false},
		{"timestamp", "timestamp", "time.Time", "Date", false, true, false},
		{"timestamptz", "timestamptz", "time.Time", "Date", false, true, false},
		{"date", "date", "time.Time", "Date", false, true, false},
		{"bytea", "bytea", "[]byte", "Uint8Array", false, false, false},
		{"uuid", "uuid", "string", "string", false, false, false},
		{"inet", "inet", "string", "string", false, false, false},
		{"array", "array", "[]any", "any[]", false, false, false},
		{"unknown", "unknown_custom_type", "interface{}", "any", false, false, false},
	}

	m := NewTypeMapper(DatabaseTypePostgreSQL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.MapDatabaseType(tt.dbType)
			if got.GoType != tt.wantGo {
				t.Errorf("MapDatabaseType() GoType = %v, want %v", got.GoType, tt.wantGo)
			}
			if got.TypeScriptType.Name != tt.wantTS {
				t.Errorf("MapDatabaseType() TypeScript = %v, want %v", got.TypeScriptType.Name, tt.wantTS)
			}
			if got.GoTypeInfo.IsJSON != tt.wantJSON {
				t.Errorf("MapDatabaseType() IsJSON = %v, want %v", got.GoTypeInfo.IsJSON, tt.wantJSON)
			}
			if got.GoTypeInfo.IsTime != tt.wantTime {
				t.Errorf("MapDatabaseType() IsTime = %v, want %v", got.GoTypeInfo.IsTime, tt.wantTime)
			}
			if got.GoTypeInfo.IsNumeric != tt.wantNum {
				t.Errorf("MapDatabaseType() IsNumeric = %v, want %v", got.GoTypeInfo.IsNumeric, tt.wantNum)
			}
		})
	}
}

func TestTypeMapper_MapDatabaseType_MySQL(t *testing.T) {
	tests := []struct {
		name   string
		dbType string
		wantGo string
		wantTS string
	}{
		{"tinyint", "tinyint", "int32", "number"},
		{"int", "int", "int32", "number"},
		{"bigint", "bigint", "int64", "number"},
		{"float", "float", "float32", "number"},
		{"double", "double", "float64", "number"},
		{"decimal", "decimal", "float64", "number"},
		{"varchar", "varchar", "string", "string"},
		{"text", "text", "string", "string"},
		{"json", "json", "string", "any"},
		{"enum", "enum", "string", "string"},
		{"set", "set", "string", "string"},
		{"date", "date", "time.Time", "Date"},
		{"datetime", "datetime", "time.Time", "Date"},
		{"timestamp", "timestamp", "time.Time", "Date"},
		{"binary", "binary", "[]byte", "Uint8Array"},
		{"blob", "blob", "[]byte", "Uint8Array"},
		{"tinytext", "tinytext", "string", "string"},
		{"mediumtext", "mediumtext", "string", "string"},
		{"longtext", "longtext", "string", "string"},
	}

	m := NewTypeMapper(DatabaseTypeMySQL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.MapDatabaseType(tt.dbType)
			if got.GoType != tt.wantGo {
				t.Errorf("MapDatabaseType() GoType = %v, want %v", got.GoType, tt.wantGo)
			}
			if got.TypeScriptType.Name != tt.wantTS {
				t.Errorf("MapDatabaseType() TypeScript = %v, want %v", got.TypeScriptType.Name, tt.wantTS)
			}
		})
	}
}

func TestTypeMapper_MapDatabaseType_SQLite(t *testing.T) {
	tests := []struct {
		name   string
		dbType string
		wantGo string
		wantTS string
	}{
		{"integer", "integer", "int64", "number"},
		{"int", "int", "int64", "number"},
		{"real", "real", "float64", "number"},
		{"double", "double", "float64", "number"},
		{"float", "float", "float64", "number"},
		{"numeric", "numeric", "float64", "number"},
		{"text", "text", "string", "string"},
		{"varchar", "varchar", "string", "string"},
		{"clob", "clob", "string", "string"},
		{"blob", "blob", "[]byte", "Uint8Array"},
		{"null", "null", "interface{}", "null"},
		{"none", "none", "interface{}", "null"},
	}

	m := NewTypeMapper(DatabaseTypeSQLite)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.MapDatabaseType(tt.dbType)
			if got.GoType != tt.wantGo {
				t.Errorf("MapDatabaseType() GoType = %v, want %v", got.GoType, tt.wantGo)
			}
			if got.TypeScriptType.Name != tt.wantTS {
				t.Errorf("MapDatabaseType() TypeScript = %v, want %v", got.TypeScriptType.Name, tt.wantTS)
			}
		})
	}
}

func TestTypeMapper_MapDatabaseType_MSSQL(t *testing.T) {
	tests := []struct {
		name   string
		dbType string
		wantGo string
		wantTS string
	}{
		{"bigint", "bigint", "int64", "number"},
		{"int", "int", "int32", "number"},
		{"smallint", "smallint", "int16", "number"},
		{"tinyint", "tinyint", "uint8", "number"},
		{"bit", "bit", "bool", "boolean"},
		{"float", "float", "float64", "number"},
		{"real", "real", "float32", "number"},
		{"decimal", "decimal", "float64", "number"},
		{"money", "money", "float64", "number"},
		{"varchar", "varchar", "string", "string"},
		{"nvarchar", "nvarchar", "string", "string"},
		{"char", "char", "string", "string"},
		{"nchar", "nchar", "string", "string"},
		{"text", "text", "string", "string"},
		{"ntext", "ntext", "string", "string"},
		{"datetime", "datetime", "time.Time", "Date"},
		{"datetime2", "datetime2", "time.Time", "Date"},
		{"smalldatetime", "smalldatetime", "time.Time", "Date"},
		{"time", "time", "time.Time", "string"},
		{"date", "date", "time.Time", "Date"},
		{"binary", "binary", "[]byte", "Uint8Array"},
		{"varbinary", "varbinary", "[]byte", "Uint8Array"},
		{"image", "image", "[]byte", "Uint8Array"},
		{"uniqueidentifier", "uniqueidentifier", "string", "string"},
		{"xml", "xml", "string", "string"},
	}

	m := NewTypeMapper(DatabaseTypeMSSQL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.MapDatabaseType(tt.dbType)
			if got.GoType != tt.wantGo {
				t.Errorf("MapDatabaseType() GoType = %v, want %v", got.GoType, tt.wantGo)
			}
			if got.TypeScriptType.Name != tt.wantTS {
				t.Errorf("MapDatabaseType() TypeScript = %v, want %v", got.TypeScriptType.Name, tt.wantTS)
			}
		})
	}
}

func TestTypeMapper_MapDatabaseType_ClickHouse(t *testing.T) {
	tests := []struct {
		name   string
		dbType string
		wantGo string
		wantTS string
	}{
		{"int8", "int8", "int32", "number"},
		{"int16", "int16", "int32", "number"},
		{"int32", "int32", "int32", "number"},
		{"int64", "int64", "int64", "number"},
		{"uint8", "uint8", "uint32", "number"},
		{"uint16", "uint16", "uint32", "number"},
		{"uint32", "uint32", "uint32", "number"},
		{"uint64", "uint64", "uint64", "number"},
		{"float32", "float32", "float32", "number"},
		{"float64", "float64", "float64", "number"},
		{"decimal", "decimal", "float64", "number"},
		{"string", "string", "string", "string"},
		{"fixedstring", "fixedstring", "string", "string"},
		{"date", "date", "time.Time", "Date"},
		{"date32", "date32", "time.Time", "Date"},
		{"datetime", "datetime", "time.Time", "Date"},
		{"datetime64", "datetime64", "time.Time", "Date"},
		{"uuid", "uuid", "string", "string"},
		{"enum", "enum", "string", "string"},
		{"array", "array", "[]any", "any[]"},
		{"json", "json", "string", "any"},
		{"tuple", "tuple", "string", "any"},
		{"map", "map", "map[string]any", "Record<string, any>"},
		{"ipv4", "ipv4", "string", "string"},
		{"ipv6", "ipv6", "string", "string"},
	}

	m := NewTypeMapper(DatabaseTypeClickHouse)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.MapDatabaseType(tt.dbType)
			if got.GoType != tt.wantGo {
				t.Errorf("MapDatabaseType() GoType = %v, want %v", got.GoType, tt.wantGo)
			}
			if got.TypeScriptType.Name != tt.wantTS {
				t.Errorf("MapDatabaseType() TypeScript = %v, want %v", got.TypeScriptType.Name, tt.wantTS)
			}
		})
	}
}

func TestTypeMapper_MapDatabaseType_MongoDB(t *testing.T) {
	tests := []struct {
		name   string
		dbType string
		wantGo string
		wantTS string
	}{
		{"objectid", "objectid", "string", "string"},
		{"string", "string", "string", "string"},
		{"int", "int", "int32", "number"},
		{"int32", "int32", "int32", "number"},
		{"long", "long", "int64", "number"},
		{"int64", "int64", "int64", "number"},
		{"double", "double", "float64", "number"},
		{"float", "float", "float64", "number"},
		{"decimal", "decimal", "float64", "number"},
		{"bool", "bool", "bool", "boolean"},
		{"boolean", "boolean", "bool", "boolean"},
		{"date", "date", "time.Time", "Date"},
		{"timestamp", "timestamp", "time.Time", "Date"},
		{"binData", "binData", "[]byte", "Uint8Array"},
		{"binary", "binary", "[]byte", "Uint8Array"},
		{"null", "null", "interface{}", "null"},
		{"undefined", "undefined", "interface{}", "undefined"},
		{"regex", "regex", "string", "string"},
		{"javascript", "javascript", "string", "string"},
		{"array", "array", "[]any", "any[]"},
		{"object", "object", "map[string]any", "Record<string, any>"},
	}

	m := NewTypeMapper(DatabaseTypeMongoDB)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.MapDatabaseType(tt.dbType)
			if got.GoType != tt.wantGo {
				t.Errorf("MapDatabaseType() GoType = %v, want %v", got.GoType, tt.wantGo)
			}
			if got.TypeScriptType.Name != tt.wantTS {
				t.Errorf("MapDatabaseType() TypeScript = %v, want %v", got.TypeScriptType.Name, tt.wantTS)
			}
		})
	}
}

func TestTypeMapper_MapDatabaseType_Redis(t *testing.T) {
	tests := []struct {
		name   string
		dbType string
		wantGo string
		wantTS string
	}{
		{"string", "string", "string", "string"},
		{"list", "list", "[]string", "string[]"},
		{"set", "set", "map[string]struct{}", "Set<string>"},
		{"zset", "zset", "map[string]float64", "Record<string, number>"},
		{"sortedset", "sortedset", "map[string]float64", "Record<string, number>"},
		{"hash", "hash", "map[string]string", "Record<string, string>"},
		{"hmap", "hmap", "map[string]string", "Record<string, string>"},
		{"stream", "stream", "map[string]any", "Record<string, any>"},
		{"hyperloglog", "hyperloglog", "string", "string"},
		{"bitmap", "bitmap", "string", "string"},
		{"none", "none", "interface{}", "null"},
	}

	m := NewTypeMapper(DatabaseTypeRedis)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.MapDatabaseType(tt.dbType)
			if got.GoType != tt.wantGo {
				t.Errorf("MapDatabaseType() GoType = %v, want %v", got.GoType, tt.wantGo)
			}
			if got.TypeScriptType.Name != tt.wantTS {
				t.Errorf("MapDatabaseType() TypeScript = %v, want %v", got.TypeScriptType.Name, tt.wantTS)
			}
		})
	}
}

func TestTypeMapper_MapGoType(t *testing.T) {
	tests := []struct {
		name     string
		goType   string
		wantTS   string
		wantNull bool
	}{
		{"bool", "bool", "boolean", false},
		{"int", "int", "number", false},
		{"int32", "int32", "number", false},
		{"int64", "int64", "number", false},
		{"float32", "float32", "number", false},
		{"float64", "float64", "number", false},
		{"string", "string", "string", false},
		{"time.Time", "time.Time", "Date", false},
		{"[]byte", "[]byte", "Uint8Array", false},
		{"[]any", "[]any", "any[]", false},
		{"map[string]any", "map[string]any", "Record<string, any>", false},
	}

	m := NewTypeMapper(DatabaseTypePostgreSQL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.MapGoType(tt.goType)
			if got.Name != tt.wantTS {
				t.Errorf("MapGoType() TypeScript = %v, want %v", got.Name, tt.wantTS)
			}
			if got.Nullable != tt.wantNull {
				t.Errorf("MapGoType() Nullable = %v, want %v", got.Nullable, tt.wantNull)
			}
		})
	}
}

func TestIsValidDatabaseType(t *testing.T) {
	tests := []struct {
		name   string
		dbType string
		want   bool
	}{
		{"postgresql lowercase", "postgresql", true},
		{"postgres", "postgres", true},
		{"mysql", "mysql", true},
		{"mariadb", "mariadb", true},
		{"sqlite", "sqlite", true},
		{"duckdb", "duckdb", true},
		{"mssql", "mssql", true},
		{"clickhouse", "clickhouse", true},
		{"mongodb", "mongodb", true},
		{"redis", "redis", true},
		{"invalid", "invalid", false},
		{"unknown", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidDatabaseType(tt.dbType)
			if got != tt.want {
				t.Errorf("IsValidDatabaseType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeType(t *testing.T) {
	tests := []struct {
		name   string
		dbType string
		want   string
	}{
		{"simple", "varchar", "varchar"},
		{"with size", "varchar(255)", "varchar"},
		{"with size int", "int(11)", "int"},
		{"uppercase", "VARCHAR", "varchar"},
		{"unsigned", "unsigned int", "int"},
		{"unsigned with size", "unsigned bigint(20)", "bigint"},
		{"decimal with precision", "decimal(10,2)", "decimal"},
		{"numeric with precision", "numeric(18,6)", "numeric"},
		{"char with size", "char(10)", "char"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeType(tt.dbType)
			if got != tt.want {
				t.Errorf("normalizeType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "user", "User"},
		{"with underscore", "user_name", "UserName"},
		{"multiple underscores", "user_name_id", "UserNameId"},
		{"already pascal", "UserName", "UserName"},
		{"empty", "", ""},
		{"single char", "a", "A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toPascalCase(tt.input)
			if got != tt.want {
				t.Errorf("toPascalCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "user", "user"},
		{"with underscore", "user_name", "userName"},
		{"multiple underscores", "user_name_id", "userNameId"},
		{"already camel", "userName", "userName"},
		{"empty", "", ""},
		{"single char", "A", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toCamelCase(tt.input)
			if got != tt.want {
				t.Errorf("toCamelCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeMapper_GetTypeScriptInterfaces(t *testing.T) {
	m := NewTypeMapper(DatabaseTypePostgreSQL)

	columns := []ColumnInfo{
		{Name: "id", Type: "int4", Nullable: false},
		{Name: "createdAt", Type: "timestamptz", Nullable: false},
	}

	got := m.GetTypeScriptInterfaces("users", columns)

	expected := `interface Users {
  id: number;
  createdAt: Date;
}
`

	if got != expected {
		t.Errorf("GetTypeScriptInterfaces() = %v, want %v", got, expected)
	}
}

func TestColumnFlags(t *testing.T) {
	m := NewTypeMapper(DatabaseTypePostgreSQL)

	tests := []struct {
		name      string
		dbType    string
		wantNum   bool
		wantStr   bool
		wantTime  bool
		wantBin   bool
		wantJSON  bool
		wantArray bool
		wantUUID  bool
	}{
		{"int4", "int4", true, false, false, false, false, false, false},
		{"varchar", "varchar", false, true, false, false, false, false, false},
		{"timestamp", "timestamp", false, false, true, false, false, false, false},
		{"bytea", "bytea", false, false, false, true, false, false, false},
		{"array", "array", false, false, false, false, false, true, false},
		{"uuid", "uuid", false, true, false, false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := m.MapDatabaseType(tt.dbType)
			if mapping.ColumnFlags.IsNumeric != tt.wantNum {
				t.Errorf("IsNumeric = %v, want %v", mapping.ColumnFlags.IsNumeric, tt.wantNum)
			}
			if mapping.ColumnFlags.IsString != tt.wantStr {
				t.Errorf("IsString = %v, want %v", mapping.ColumnFlags.IsString, tt.wantStr)
			}
			if mapping.ColumnFlags.IsTime != tt.wantTime {
				t.Errorf("IsTime = %v, want %v", mapping.ColumnFlags.IsTime, tt.wantTime)
			}
			if mapping.ColumnFlags.IsBinary != tt.wantBin {
				t.Errorf("IsBinary = %v, want %v", mapping.ColumnFlags.IsBinary, tt.wantBin)
			}
			if mapping.ColumnFlags.IsJSON != tt.wantJSON {
				t.Errorf("IsJSON = %v, want %v", mapping.ColumnFlags.IsJSON, tt.wantJSON)
			}
			if mapping.ColumnFlags.IsArray != tt.wantArray {
				t.Errorf("IsArray = %v, want %v", mapping.ColumnFlags.IsArray, tt.wantArray)
			}
			if mapping.ColumnFlags.IsUUID != tt.wantUUID {
				t.Errorf("IsUUID = %v, want %v", mapping.ColumnFlags.IsUUID, tt.wantUUID)
			}
		})
	}
}

func TestDuckDBTypeMappings(t *testing.T) {
	m := NewTypeMapper(DatabaseTypeDuckDB)

	tests := []struct {
		name   string
		dbType string
		wantGo string
		wantTS string
	}{
		{"bool", "bool", "bool", "boolean"},
		{"boolean", "boolean", "bool", "boolean"},
		{"tinyint", "tinyint", "int8", "number"},
		{"smallint", "smallint", "int16", "number"},
		{"int", "int", "int32", "number"},
		{"integer", "integer", "int32", "number"},
		{"bigint", "bigint", "int64", "number"},
		{"utinyint", "utinyint", "uint8", "number"},
		{"usmallint", "usmallint", "uint16", "number"},
		{"uinteger", "uinteger", "uint32", "number"},
		{"ubigint", "ubigint", "uint64", "number"},
		{"float", "float", "float32", "number"},
		{"double", "double", "float64", "number"},
		{"decimal", "decimal", "float64", "number"},
		{"string", "string", "string", "string"},
		{"varchar", "varchar", "string", "string"},
		{"blob", "blob", "[]byte", "Uint8Array"},
		{"date", "date", "time.Time", "Date"},
		{"time", "time", "time.Time", "string"},
		{"timestamp", "timestamp", "time.Time", "Date"},
		{"timestamptz", "timestamptz", "time.Time", "Date"},
		{"interval", "interval", "string", "string"},
		{"uuid", "uuid", "string", "string"},
		{"json", "json", "string", "any"},
		{"jsonb", "jsonb", "string", "any"},
		{"array", "array", "[]any", "any[]"},
		{"struct", "struct", "map[string]any", "Record<string, any>"},
		{"map", "map", "map[string]any", "Record<string, any>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.MapDatabaseType(tt.dbType)
			if got.GoType != tt.wantGo {
				t.Errorf("MapDatabaseType() GoType = %v, want %v", got.GoType, tt.wantGo)
			}
			if got.TypeScriptType.Name != tt.wantTS {
				t.Errorf("MapDatabaseType() TypeScript = %v, want %v", got.TypeScriptType.Name, tt.wantTS)
			}
		})
	}
}
