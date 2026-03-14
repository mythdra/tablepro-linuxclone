package types

import "strings"

func (m *TypeMapper) mapClickHouseType(dbType string) *TypeMapping {
	mappings := map[string]*TypeMapping{
		"int8": {
			DBType:         "int8",
			GoType:         "int32",
			GoTypeInfo:     GoType{Name: "int32", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"int16": {
			DBType:         "int16",
			GoType:         "int32",
			GoTypeInfo:     GoType{Name: "int32", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"int32": {
			DBType:         "int32",
			GoType:         "int32",
			GoTypeInfo:     GoType{Name: "int32", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"int64": {
			DBType:         "int64",
			GoType:         "int64",
			GoTypeInfo:     GoType{Name: "int64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"int128": {
			DBType:         "int128",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"int256": {
			DBType:         "int256",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"uint8": {
			DBType:         "uint8",
			GoType:         "uint32",
			GoTypeInfo:     GoType{Name: "uint32", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"uint16": {
			DBType:         "uint16",
			GoType:         "uint32",
			GoTypeInfo:     GoType{Name: "uint32", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"uint32": {
			DBType:         "uint32",
			GoType:         "uint32",
			GoTypeInfo:     GoType{Name: "uint32", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"uint64": {
			DBType:         "uint64",
			GoType:         "uint64",
			GoTypeInfo:     GoType{Name: "uint64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"uint128": {
			DBType:         "uint128",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"uint256": {
			DBType:         "uint256",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"float32": {
			DBType:         "float32",
			GoType:         "float32",
			GoTypeInfo:     GoType{Name: "float32", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"float64": {
			DBType:         "float64",
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
		"string": {
			DBType:         "string",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"fixedstring": {
			DBType:         "fixedstring",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"date": {
			DBType:         "date",
			GoType:         "time.Time",
			GoTypeInfo:     GoType{Name: "time.Time", IsTime: true},
			TypeScriptType: TypeScriptType{Name: "Date"},
			ColumnFlags:    ColumnFlags{IsTime: true},
		},
		"date32": {
			DBType:         "date32",
			GoType:         "time.Time",
			GoTypeInfo:     GoType{Name: "time.Time", IsTime: true},
			TypeScriptType: TypeScriptType{Name: "Date"},
			ColumnFlags:    ColumnFlags{IsTime: true},
		},
		"datetime": {
			DBType:         "datetime",
			GoType:         "time.Time",
			GoTypeInfo:     GoType{Name: "time.Time", IsTime: true},
			TypeScriptType: TypeScriptType{Name: "Date"},
			ColumnFlags:    ColumnFlags{IsTime: true},
		},
		"datetime64": {
			DBType:         "datetime64",
			GoType:         "time.Time",
			GoTypeInfo:     GoType{Name: "time.Time", IsTime: true},
			TypeScriptType: TypeScriptType{Name: "Date"},
			ColumnFlags:    ColumnFlags{IsTime: true},
		},
		"uuid": {
			DBType:         "uuid",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true, IsUUID: true},
		},
		"enum": {
			DBType:         "enum",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true, IsEnum: true},
		},
		"enum8": {
			DBType:         "enum8",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true, IsEnum: true},
		},
		"enum16": {
			DBType:         "enum16",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true, IsEnum: true},
		},
		"array": {
			DBType:         "array",
			GoType:         "[]any",
			GoTypeInfo:     GoType{Name: "[]any", IsArray: true},
			TypeScriptType: TypeScriptType{Name: "any[]", Array: true},
			ColumnFlags:    ColumnFlags{IsArray: true},
		},
		"json": {
			DBType:         "json",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsJSON: true},
			TypeScriptType: TypeScriptType{Name: "any"},
			ColumnFlags:    ColumnFlags{IsJSON: true, IsString: true},
		},
		"tuple": {
			DBType:         "tuple",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "any"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"map": {
			DBType:         "map",
			GoType:         "map[string]any",
			GoTypeInfo:     GoType{Name: "map[string]any", IsMap: true},
			TypeScriptType: TypeScriptType{Name: "Record<string, any>", Generic: "string, any"},
			ColumnFlags:    ColumnFlags{IsMap: true},
		},
		"nested": {
			DBType:         "nested",
			GoType:         "map[string]any",
			GoTypeInfo:     GoType{Name: "map[string]any", IsMap: true},
			TypeScriptType: TypeScriptType{Name: "Record<string, any>"},
			ColumnFlags:    ColumnFlags{IsMap: true},
		},
		"ipv4": {
			DBType:         "ipv4",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"ipv6": {
			DBType:         "ipv6",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
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

	if strings.HasPrefix(dbType, "tuple(") {
		return &TypeMapping{
			DBType:         dbType,
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "any"},
			ColumnFlags:    ColumnFlags{IsString: true},
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

func getClickHouseMappings() []*TypeMapping {
	return []*TypeMapping{}
}
