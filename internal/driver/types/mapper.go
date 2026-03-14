package types

import (
	"fmt"
	"strings"
	"time"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypeSQLite     DatabaseType = "sqlite"
	DatabaseTypeMSSQL      DatabaseType = "mssql"
	DatabaseTypeClickHouse DatabaseType = "clickhouse"
	DatabaseTypeMongoDB    DatabaseType = "mongodb"
	DatabaseTypeRedis      DatabaseType = "redis"
	DatabaseTypeDuckDB     DatabaseType = "duckdb"
)

// TypeMapper provides type mapping between database, Go, and TypeScript types
type TypeMapper struct {
	dbType DatabaseType
}

// NewTypeMapper creates a new TypeMapper for the specified database type
func NewTypeMapper(dbType DatabaseType) *TypeMapper {
	return &TypeMapper{dbType: dbType}
}

// GoType represents a Go type with additional metadata
type GoType struct {
	Name        string
	IsNumeric   bool
	IsString    bool
	IsTime      bool
	IsBinary    bool
	IsJSON      bool
	IsArray     bool
	IsMap       bool
	IsPointer   bool
	IsInterface bool
}

// TypeScriptType represents a TypeScript type
type TypeScriptType struct {
	Name     string
	Nullable bool
	Array    bool
	Generic  string
}

// TypeMapping contains all type mapping information
type TypeMapping struct {
	DBType         string
	GoType         string
	GoTypeInfo     GoType
	TypeScriptType TypeScriptType
	ColumnFlags    ColumnFlags
}

// ColumnFlags contains boolean flags for column classification
type ColumnFlags struct {
	IsNumeric bool
	IsString  bool
	IsTime    bool
	IsBinary  bool
	IsJSON    bool
	IsArray   bool
	IsMap     bool
	IsUUID    bool
	IsEnum    bool
}

// MapDatabaseType maps a database column type to Go type
func (m *TypeMapper) MapDatabaseType(dbType string) *TypeMapping {
	dbType = normalizeType(dbType)

	switch m.dbType {
	case DatabaseTypePostgreSQL:
		return m.mapPostgresType(dbType)
	case DatabaseTypeMySQL:
		return m.mapMySQLType(dbType)
	case DatabaseTypeSQLite:
		return m.mapSQLiteType(dbType)
	case DatabaseTypeMSSQL:
		return m.mapMSSQLType(dbType)
	case DatabaseTypeClickHouse:
		return m.mapClickHouseType(dbType)
	case DatabaseTypeMongoDB:
		return m.mapMongoDBType(dbType)
	case DatabaseTypeRedis:
		return m.mapRedisType(dbType)
	case DatabaseTypeDuckDB:
		return m.mapDuckDBType(dbType)
	default:
		return m.fallbackMapping(dbType)
	}
}

// MapGoType maps a Go type to TypeScript type
func (m *TypeMapper) MapGoType(goType string) *TypeScriptType {
	goType = strings.TrimPrefix(goType, "*")
	nullable := strings.HasPrefix(goType, "*")
	goType = strings.TrimPrefix(goType, "*")

	tsType := &TypeScriptType{
		Nullable: nullable,
	}

	switch goType {
	case "bool":
		tsType.Name = "boolean"
	case "int", "int8", "int16", "int32", "uint", "uint8", "uint16", "uint32":
		tsType.Name = "number"
	case "int64", "uint64":
		tsType.Name = "number"
	case "float32", "float64":
		tsType.Name = "number"
	case "string":
		tsType.Name = "string"
	case "time.Time":
		tsType.Name = "Date"
	case "[]byte":
		tsType.Name = "Uint8Array"
	case "[]any", "[]interface{}":
		tsType.Name = "any[]"
		tsType.Array = true
	case "map[string]any", "map[string]interface{}":
		tsType.Name = "Record<string, any>"
		tsType.Generic = "string, any"
	default:
		tsType.Name = "any"
	}

	return tsType
}

// normalizeType normalizes database type string
func normalizeType(dbType string) string {
	dbType = strings.ToLower(strings.TrimSpace(dbType))
	// Remove size specifications like varchar(255) -> varchar
	if idx := strings.Index(dbType, "("); idx > 0 {
		dbType = dbType[:idx]
	}
	// Remove unsigned prefix
	dbType = strings.TrimPrefix(dbType, "unsigned ")
	return dbType
}

// fallbackMapping returns a fallback mapping for unknown types
func (m *TypeMapper) fallbackMapping(dbType string) *TypeMapping {
	return &TypeMapping{
		DBType:         dbType,
		GoType:         "interface{}",
		GoTypeInfo:     GoType{Name: "interface{}", IsInterface: true},
		TypeScriptType: TypeScriptType{Name: "any", Nullable: true},
		ColumnFlags:    ColumnFlags{},
	}
}

// GetAllMappings returns all type mappings for the database
func (m *TypeMapper) GetAllMappings() []*TypeMapping {
	mappings := make([]*TypeMapping, 0)

	switch m.dbType {
	case DatabaseTypePostgreSQL:
		mappings = append(mappings, getPostgresMappings()...)
	case DatabaseTypeMySQL:
		mappings = append(mappings, getMySQLMappings()...)
	case DatabaseTypeSQLite:
		mappings = append(mappings, getSQLiteMappings()...)
	case DatabaseTypeMSSQL:
		mappings = append(mappings, getMSSQLMappings()...)
	case DatabaseTypeClickHouse:
		mappings = append(mappings, getClickHouseMappings()...)
	case DatabaseTypeMongoDB:
		mappings = append(mappings, getMongoDBMappings()...)
	case DatabaseTypeRedis:
		mappings = append(mappings, getRedisMappings()...)
	case DatabaseTypeDuckDB:
		mappings = append(mappings, getDuckDBMappings()...)
	}

	return mappings
}

// GetTypeScriptInterfaces generates TypeScript interface definitions
func (m *TypeMapper) GetTypeScriptInterfaces(tableName string, columns []ColumnInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("interface %s {\n", toPascalCase(tableName)))

	for _, col := range columns {
		mapping := m.MapDatabaseType(col.Type)
		sb.WriteString(fmt.Sprintf("  %s: %s;\n", toCamelCase(col.Name), mapping.TypeScriptType.Name))
	}

	sb.WriteString("}\n")
	return sb.String()
}

// ColumnInfo represents column metadata
type ColumnInfo struct {
	Name      string
	Type      string
	Nullable  bool
	IsPrimary bool
	IsUnique  bool
	Default   *string
}

// toPascalCase converts string to PascalCase
func toPascalCase(s string) string {
	if s == "" {
		return s
	}
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// toCamelCase converts string to camelCase
func toCamelCase(s string) string {
	pascal := toPascalCase(s)
	if len(pascal) == 0 {
		return pascal
	}
	return strings.ToLower(pascal[:1]) + pascal[1:]
}

// IsValidDatabaseType checks if the database type is supported
func IsValidDatabaseType(dbType string) bool {
	switch DatabaseType(strings.ToLower(dbType)) {
	case DatabaseTypePostgreSQL, DatabaseTypeMySQL, DatabaseTypeSQLite,
		DatabaseTypeMSSQL, DatabaseTypeClickHouse, DatabaseTypeMongoDB,
		DatabaseTypeRedis, DatabaseTypeDuckDB:
		return true
	case "mariadb", "sqlite3", "duck", "sqlserver", "ch", "mongo",
		"postgres", "pgsql", "pg":
		return true
	default:
		return false
	}
}

func init() {
	// Ensure time import is used
	_ = time.Time{}
}

func (m *TypeMapper) mapDuckDBType(dbType string) *TypeMapping {
	mappings := map[string]*TypeMapping{
		"bool": {
			DBType:         "bool",
			GoType:         "bool",
			GoTypeInfo:     GoType{Name: "bool", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "boolean"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"boolean": {
			DBType:         "boolean",
			GoType:         "bool",
			GoTypeInfo:     GoType{Name: "bool", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "boolean"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"tinyint": {
			DBType:         "tinyint",
			GoType:         "int8",
			GoTypeInfo:     GoType{Name: "int8", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"smallint": {
			DBType:         "smallint",
			GoType:         "int16",
			GoTypeInfo:     GoType{Name: "int16", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"int": {
			DBType:         "int",
			GoType:         "int32",
			GoTypeInfo:     GoType{Name: "int32", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"integer": {
			DBType:         "integer",
			GoType:         "int32",
			GoTypeInfo:     GoType{Name: "int32", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"bigint": {
			DBType:         "bigint",
			GoType:         "int64",
			GoTypeInfo:     GoType{Name: "int64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"utinyint": {
			DBType:         "utinyint",
			GoType:         "uint8",
			GoTypeInfo:     GoType{Name: "uint8", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"usmallint": {
			DBType:         "usmallint",
			GoType:         "uint16",
			GoTypeInfo:     GoType{Name: "uint16", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"uinteger": {
			DBType:         "uinteger",
			GoType:         "uint32",
			GoTypeInfo:     GoType{Name: "uint32", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"ubigint": {
			DBType:         "ubigint",
			GoType:         "uint64",
			GoTypeInfo:     GoType{Name: "uint64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"float": {
			DBType:         "float",
			GoType:         "float32",
			GoTypeInfo:     GoType{Name: "float32", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"double": {
			DBType:         "double",
			GoType:         "float64",
			GoTypeInfo:     GoType{Name: "float64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"decimal": {
			DBType:         "decimal",
			GoType:         "float64",
			GoTypeInfo:     GoType{Name: "float64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"numeric": {
			DBType:         "numeric",
			GoType:         "float64",
			GoTypeInfo:     GoType{Name: "float64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"real": {
			DBType:         "real",
			GoType:         "float32",
			GoTypeInfo:     GoType{Name: "float32", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"string": {
			DBType:         "string",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"varchar": {
			DBType:         "varchar",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"char": {
			DBType:         "char",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"blob": {
			DBType:         "blob",
			GoType:         "[]byte",
			GoTypeInfo:     GoType{Name: "[]byte", IsBinary: true},
			TypeScriptType: TypeScriptType{Name: "Uint8Array"},
			ColumnFlags:    ColumnFlags{IsBinary: true},
		},
		"bytea": {
			DBType:         "bytea",
			GoType:         "[]byte",
			GoTypeInfo:     GoType{Name: "[]byte", IsBinary: true},
			TypeScriptType: TypeScriptType{Name: "Uint8Array"},
			ColumnFlags:    ColumnFlags{IsBinary: true},
		},
		"date": {
			DBType:         "date",
			GoType:         "time.Time",
			GoTypeInfo:     GoType{Name: "time.Time", IsTime: true},
			TypeScriptType: TypeScriptType{Name: "Date"},
			ColumnFlags:    ColumnFlags{IsTime: true},
		},
		"time": {
			DBType:         "time",
			GoType:         "time.Time",
			GoTypeInfo:     GoType{Name: "time.Time", IsTime: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsTime: true},
		},
		"timestamp": {
			DBType:         "timestamp",
			GoType:         "time.Time",
			GoTypeInfo:     GoType{Name: "time.Time", IsTime: true},
			TypeScriptType: TypeScriptType{Name: "Date"},
			ColumnFlags:    ColumnFlags{IsTime: true},
		},
		"timestamptz": {
			DBType:         "timestamptz",
			GoType:         "time.Time",
			GoTypeInfo:     GoType{Name: "time.Time", IsTime: true},
			TypeScriptType: TypeScriptType{Name: "Date"},
			ColumnFlags:    ColumnFlags{IsTime: true},
		},
		"interval": {
			DBType:         "interval",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"uuid": {
			DBType:         "uuid",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true, IsUUID: true},
		},
		"json": {
			DBType:         "json",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsJSON: true},
			TypeScriptType: TypeScriptType{Name: "any"},
			ColumnFlags:    ColumnFlags{IsJSON: true, IsString: true},
		},
		"jsonb": {
			DBType:         "jsonb",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsJSON: true},
			TypeScriptType: TypeScriptType{Name: "any"},
			ColumnFlags:    ColumnFlags{IsJSON: true, IsString: true},
		},
		"array": {
			DBType:         "array",
			GoType:         "[]any",
			GoTypeInfo:     GoType{Name: "[]any", IsArray: true},
			TypeScriptType: TypeScriptType{Name: "any[]", Array: true},
			ColumnFlags:    ColumnFlags{IsArray: true},
		},
		"struct": {
			DBType:         "struct",
			GoType:         "map[string]any",
			GoTypeInfo:     GoType{Name: "map[string]any", IsMap: true},
			TypeScriptType: TypeScriptType{Name: "Record<string, any>", Generic: "string, any"},
			ColumnFlags:    ColumnFlags{IsMap: true},
		},
		"map": {
			DBType:         "map",
			GoType:         "map[string]any",
			GoTypeInfo:     GoType{Name: "map[string]any", IsMap: true},
			TypeScriptType: TypeScriptType{Name: "Record<string, any>", Generic: "string, any"},
			ColumnFlags:    ColumnFlags{IsMap: true},
		},
	}

	if mapping, ok := mappings[dbType]; ok {
		return mapping
	}

	if strings.HasPrefix(dbType, "array(") {
		return &TypeMapping{
			DBType:         dbType,
			GoType:         "[]any",
			GoTypeInfo:     GoType{Name: "[]any", IsArray: true},
			TypeScriptType: TypeScriptType{Name: "any[]", Array: true},
			ColumnFlags:    ColumnFlags{IsArray: true},
		}
	}

	if strings.HasPrefix(dbType, "struct(") {
		return &TypeMapping{
			DBType:         dbType,
			GoType:         "map[string]any",
			GoTypeInfo:     GoType{Name: "map[string]any", IsMap: true},
			TypeScriptType: TypeScriptType{Name: "Record<string, any>"},
			ColumnFlags:    ColumnFlags{IsMap: true},
		}
	}

	if strings.HasPrefix(dbType, "map(") {
		return &TypeMapping{
			DBType:         dbType,
			GoType:         "map[string]any",
			GoTypeInfo:     GoType{Name: "map[string]any", IsMap: true},
			TypeScriptType: TypeScriptType{Name: "Record<string, any>"},
			ColumnFlags:    ColumnFlags{IsMap: true},
		}
	}

	return m.fallbackMapping(dbType)
}

func getDuckDBMappings() []*TypeMapping {
	return []*TypeMapping{}
}
