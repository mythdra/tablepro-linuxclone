package types

import "strings"

func (m *TypeMapper) mapPostgresType(dbType string) *TypeMapping {
	mappings := map[string]*TypeMapping{
		"bool": {
			DBType:         "bool",
			GoType:         "bool",
			GoTypeInfo:     GoType{Name: "bool", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "boolean"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"int2": {
			DBType:         "int2",
			GoType:         "int16",
			GoTypeInfo:     GoType{Name: "int16", IsNumeric: true},
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
		"int4": {
			DBType:         "int4",
			GoType:         "int32",
			GoTypeInfo:     GoType{Name: "int32", IsNumeric: true},
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
		"int8": {
			DBType:         "int8",
			GoType:         "int64",
			GoTypeInfo:     GoType{Name: "int64", IsNumeric: true},
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
		"float4": {
			DBType:         "float4",
			GoType:         "float32",
			GoTypeInfo:     GoType{Name: "float32", IsNumeric: true},
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
		"float8": {
			DBType:         "float8",
			GoType:         "float64",
			GoTypeInfo:     GoType{Name: "float64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"double precision": {
			DBType:         "double precision",
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
		"decimal": {
			DBType:         "decimal",
			GoType:         "float64",
			GoTypeInfo:     GoType{Name: "float64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"money": {
			DBType:         "money",
			GoType:         "float64",
			GoTypeInfo:     GoType{Name: "float64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"char": {
			DBType:         "char",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"character": {
			DBType:         "character",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"bpchar": {
			DBType:         "bpchar",
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
		"text": {
			DBType:         "text",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
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
			GoType:         "[]byte",
			GoTypeInfo:     GoType{Name: "[]byte", IsBinary: true, IsJSON: true},
			TypeScriptType: TypeScriptType{Name: "any"},
			ColumnFlags:    ColumnFlags{IsJSON: true, IsBinary: true},
		},
		"xml": {
			DBType:         "xml",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
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
		"timetz": {
			DBType:         "timetz",
			GoType:         "time.Time",
			GoTypeInfo:     GoType{Name: "time.Time", IsTime: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsTime: true},
		},
		"interval": {
			DBType:         "interval",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"bytea": {
			DBType:         "bytea",
			GoType:         "[]byte",
			GoTypeInfo:     GoType{Name: "[]byte", IsBinary: true},
			TypeScriptType: TypeScriptType{Name: "Uint8Array"},
			ColumnFlags:    ColumnFlags{IsBinary: true},
		},
		"uuid": {
			DBType:         "uuid",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true, IsUUID: true},
		},
		"inet": {
			DBType:         "inet",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"cidr": {
			DBType:         "cidr",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"macaddr": {
			DBType:         "macaddr",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"point": {
			DBType:         "point",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"line": {
			DBType:         "line",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"lseg": {
			DBType:         "lseg",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"box": {
			DBType:         "box",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"path": {
			DBType:         "path",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"polygon": {
			DBType:         "polygon",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"circle": {
			DBType:         "circle",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"array": {
			DBType:         "array",
			GoType:         "[]any",
			GoTypeInfo:     GoType{Name: "[]any", IsArray: true},
			TypeScriptType: TypeScriptType{Name: "any[]", Array: true},
			ColumnFlags:    ColumnFlags{IsArray: true},
		},
	}

	if mapping, ok := mappings[dbType]; ok {
		return mapping
	}

	if strings.HasPrefix(dbType, "_") {
		return &TypeMapping{
			DBType:         dbType,
			GoType:         "[]any",
			GoTypeInfo:     GoType{Name: "[]any", IsArray: true},
			TypeScriptType: TypeScriptType{Name: "any[]", Array: true},
			ColumnFlags:    ColumnFlags{IsArray: true},
		}
	}

	return m.fallbackMapping(dbType)
}

func getPostgresMappings() []*TypeMapping {
	m := NewTypeMapper(DatabaseTypePostgreSQL)
	return m.GetAllMappings()
}
