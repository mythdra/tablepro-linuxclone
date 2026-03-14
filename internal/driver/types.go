package driver

type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypeSQLite     DatabaseType = "sqlite"
	DatabaseTypeDuckDB     DatabaseType = "duckdb"
	DatabaseTypeMSSQL      DatabaseType = "mssql"
	DatabaseTypeClickHouse DatabaseType = "clickhouse"
	DatabaseTypeMongoDB    DatabaseType = "mongodb"
	DatabaseTypeRedis      DatabaseType = "redis"
	DatabaseTypeUnknown    DatabaseType = "unknown"
)

func (d DatabaseType) String() string {
	return string(d)
}

func TypeFromString(s string) DatabaseType {
	switch s {
	case "postgresql", "postgres", "pgsql", "pg":
		return DatabaseTypePostgreSQL
	case "mysql", "mariadb":
		return DatabaseTypeMySQL
	case "sqlite", "sqlite3":
		return DatabaseTypeSQLite
	case "duckdb", "duck":
		return DatabaseTypeDuckDB
	case "mssql", "sqlserver", "sql-server":
		return DatabaseTypeMSSQL
	case "clickhouse", "ch":
		return DatabaseTypeClickHouse
	case "mongodb", "mongo":
		return DatabaseTypeMongoDB
	case "redis":
		return DatabaseTypeRedis
	default:
		return DatabaseTypeUnknown
	}
}

type DataTypeMapping struct {
	DBType    string
	GoType    string
	IsNumeric bool
	IsString  bool
	IsTime    bool
	IsBinary  bool
	IsJSON    bool
	IsArray   bool
}

func CommonDataTypes() map[DatabaseType][]DataTypeMapping {
	return map[DatabaseType][]DataTypeMapping{
		DatabaseTypePostgreSQL: {
			{"bool", "bool", true, false, false, false, false, false},
			{"int2", "int16", true, false, false, false, false, false},
			{"int4", "int32", true, false, false, false, false, false},
			{"int8", "int64", true, false, false, false, false, false},
			{"float4", "float32", true, false, false, false, false, false},
			{"float8", "float64", true, false, false, false, false, false},
			{"numeric", "float64", true, false, false, false, false, false},
			{"decimal", "float64", true, false, false, false, false, false},
			{"money", "float64", true, false, false, false, false, false},
			{"varchar", "string", false, true, false, false, false, false},
			{"text", "string", false, true, false, false, false, false},
			{"char", "string", false, true, false, false, false, false},
			{"bpchar", "string", false, true, false, false, false, false},
			{"json", "string", false, false, false, false, true, false},
			{"jsonb", "[]byte", false, false, false, true, false, false},
			{"xml", "string", false, true, false, false, false, false},
			{"timestamp", "time.Time", false, false, true, false, false, false},
			{"timestamptz", "time.Time", false, false, true, false, false, false},
			{"date", "time.Time", false, false, true, false, false, false},
			{"time", "time.Time", false, false, true, false, false, false},
			{"timetz", "time.Time", false, false, true, false, false, false},
			{"interval", "string", false, true, false, false, false, false},
			{"bytea", "[]byte", false, false, false, true, false, false},
			{"uuid", "string", false, true, false, false, false, false},
			{"inet", "string", false, true, false, false, false, false},
			{"cidr", "string", false, true, false, false, false, false},
			{"macaddr", "string", false, true, false, false, false, false},
		},
		DatabaseTypeMySQL: {
			{"tinyint", "int32", true, false, false, false, false, false},
			{"smallint", "int32", true, false, false, false, false, false},
			{"mediumint", "int32", true, false, false, false, false, false},
			{"int", "int32", true, false, false, false, false, false},
			{"integer", "int32", true, false, false, false, false, false},
			{"bigint", "int64", true, false, false, false, false, false},
			{"float", "float32", true, false, false, false, false, false},
			{"double", "float64", true, false, false, false, false, false},
			{"real", "float64", true, false, false, false, false, false},
			{"decimal", "float64", true, false, false, false, false, false},
			{"numeric", "float64", true, false, false, false, false, false},
			{"char", "string", false, true, false, false, false, false},
			{"varchar", "string", false, true, false, false, false, false},
			{"tinytext", "string", false, true, false, false, false, false},
			{"text", "string", false, true, false, false, false, false},
			{"mediumtext", "string", false, true, false, false, false, false},
			{"longtext", "string", false, true, false, false, false, false},
			{"json", "string", false, false, false, false, true, false},
			{"enum", "string", false, true, false, false, false, false},
			{"set", "string", false, true, false, false, false, false},
			{"date", "time.Time", false, false, true, false, false, false},
			{"datetime", "time.Time", false, false, true, false, false, false},
			{"timestamp", "time.Time", false, false, true, false, false, false},
			{"time", "time.Time", false, false, true, false, false, false},
			{"year", "int32", true, false, false, false, false, false},
			{"binary", "[]byte", false, false, false, true, false, false},
			{"varbinary", "[]byte", false, false, false, true, false, false},
		},
		DatabaseTypeSQLite: {
			{"integer", "int64", true, false, false, false, false, false},
			{"int", "int64", true, false, false, false, false, false},
			{"real", "float64", true, false, false, false, false, false},
			{"double", "float64", true, false, false, false, false, false},
			{"float", "float64", true, false, false, false, false, false},
			{"numeric", "float64", true, false, false, false, false, false},
			{"decimal", "float64", true, false, false, false, false, false},
			{"text", "string", false, true, false, false, false, false},
			{"character", "string", false, true, false, false, false, false},
			{"varchar", "string", false, true, false, false, false, false},
			{"clob", "string", false, true, false, false, false, false},
			{"blob", "[]byte", false, false, false, true, false, false},
			{"none", "nil", false, false, false, false, false, false},
		},
		DatabaseTypeMSSQL: {
			{"bigint", "int64", true, false, false, false, false, false},
			{"binary", "[]byte", false, false, false, true, false, false},
			{"bit", "bool", true, false, false, false, false, false},
			{"char", "string", false, true, false, false, false, false},
			{"datetime", "time.Time", false, false, true, false, false, false},
			{"datetime2", "time.Time", false, false, true, false, false, false},
			{"decimal", "float64", true, false, false, false, false, false},
			{"float", "float64", true, false, false, false, false, false},
			{"image", "[]byte", false, false, false, true, false, false},
			{"int", "int32", true, false, false, false, false, false},
			{"money", "float64", true, false, false, false, false, false},
			{"nchar", "string", false, true, false, false, false, false},
			{"ntext", "string", false, true, false, false, false, false},
			{"numeric", "float64", true, false, false, false, false, false},
			{"nvarchar", "string", false, true, false, false, false, false},
			{"real", "float32", true, false, false, false, false, false},
			{"smalldatetime", "time.Time", false, false, true, false, false, false},
			{"smallint", "int16", true, false, false, false, false, false},
			{"smallmoney", "float64", true, false, false, false, false, false},
			{"text", "string", false, true, false, false, false, false},
			{"time", "time.Time", false, false, true, false, false, false},
			{"tinyint", "uint8", true, false, false, false, false, false},
			{"uniqueidentifier", "string", false, true, false, false, false, false},
			{"varbinary", "[]byte", false, false, false, true, false, false},
			{"varchar", "string", false, true, false, false, false, false},
			{"xml", "string", false, true, false, false, false, false},
		},
		DatabaseTypeClickHouse: {
			{"Int8", "int32", true, false, false, false, false, false},
			{"Int16", "int32", true, false, false, false, false, false},
			{"Int32", "int32", true, false, false, false, false, false},
			{"Int64", "int64", true, false, false, false, false, false},
			{"UInt8", "uint32", true, false, false, false, false, false},
			{"UInt16", "uint32", true, false, false, false, false, false},
			{"UInt32", "uint32", true, false, false, false, false, false},
			{"UInt64", "uint64", true, false, false, false, false, false},
			{"Float32", "float32", true, false, false, false, false, false},
			{"Float64", "float64", true, false, false, false, false, false},
			{"Decimal", "float64", true, false, false, false, false, false},
			{"String", "string", false, true, false, false, false, false},
			{"FixedString", "string", false, true, false, false, false, false},
			{"Date", "time.Time", false, false, true, false, false, false},
			{"Date32", "time.Time", false, false, true, false, false, false},
			{"DateTime", "time.Time", false, false, true, false, false, false},
			{"DateTime64", "time.Time", false, false, true, false, false, false},
			{"UUID", "string", false, true, false, false, false, false},
			{"Enum", "string", false, true, false, false, false, false},
			{"Enum8", "string", false, true, false, false, false, false},
			{"Enum16", "string", false, true, false, false, false, false},
			{"Array", "[]any", false, false, false, false, false, true},
			{"JSON", "string", false, false, false, false, true, false},
			{"Tuple", "string", false, true, false, false, false, false},
			{"Map", "string", false, true, false, false, false, false},
		},
		DatabaseTypeDuckDB: {
			{"bool", "bool", true, false, false, false, false, false},
			{"boolean", "bool", true, false, false, false, false, false},
			{"tinyint", "int8", true, false, false, false, false, false},
			{"smallint", "int16", true, false, false, false, false, false},
			{"int", "int32", true, false, false, false, false, false},
			{"integer", "int32", true, false, false, false, false, false},
			{"bigint", "int64", true, false, false, false, false, false},
			{"utinyint", "uint8", true, false, false, false, false, false},
			{"usmallint", "uint16", true, false, false, false, false, false},
			{"uinteger", "uint32", true, false, false, false, false, false},
			{"ubigint", "uint64", true, false, false, false, false, false},
			{"float", "float32", true, false, false, false, false, false},
			{"double", "float64", true, false, false, false, false, false},
			{"decimal", "float64", true, false, false, false, false, false},
			{"numeric", "float64", true, false, false, false, false, false},
			{"real", "float32", true, false, false, false, false, false},
			{"string", "string", false, true, false, false, false, false},
			{"varchar", "string", false, true, false, false, false, false},
			{"char", "string", false, true, false, false, false, false},
			{"blob", "[]byte", false, false, false, true, false, false},
			{"bytea", "[]byte", false, false, false, true, false, false},
			{"date", "time.Time", false, false, true, false, false, false},
			{"time", "time.Time", false, false, true, false, false, false},
			{"timestamp", "time.Time", false, false, true, false, false, false},
			{"timestamptz", "time.Time", false, false, true, false, false, false},
			{"interval", "string", false, true, false, false, false, false},
			{"uuid", "string", false, true, false, false, false, false},
			{"json", "string", false, false, false, false, true, false},
			{"jsonb", "string", false, false, false, false, true, false},
			{"array", "[]any", false, false, false, false, false, true},
			{"struct", "map[string]any", false, false, false, false, false, false},
			{"map", "map[string]any", false, false, false, false, false, false},
		},
	}
}

func GetDataTypeMapping(dbType DatabaseType, columnType string) *DataTypeMapping {
	mappings, ok := CommonDataTypes()[dbType]
	if !ok {
		return nil
	}

	for i := range mappings {
		if mappings[i].DBType == columnType {
			return &mappings[i]
		}
	}

	return nil
}

func IsNumericType(dbType DatabaseType, columnType string) bool {
	mapping := GetDataTypeMapping(dbType, columnType)
	return mapping != nil && mapping.IsNumeric
}

func IsStringType(dbType DatabaseType, columnType string) bool {
	mapping := GetDataTypeMapping(dbType, columnType)
	return mapping != nil && mapping.IsString
}

func IsTimeType(dbType DatabaseType, columnType string) bool {
	mapping := GetDataTypeMapping(dbType, columnType)
	return mapping != nil && mapping.IsTime
}
