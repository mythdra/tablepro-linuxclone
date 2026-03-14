package types

func (m *TypeMapper) mapMongoDBType(dbType string) *TypeMapping {
	mappings := map[string]*TypeMapping{
		"objectid": {
			DBType:         "objectid",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true, IsUUID: true},
		},
		"objectId": {
			DBType:         "objectId",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true, IsUUID: true},
		},
		"string": {
			DBType:         "string",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"int": {
			DBType:         "int",
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
		"long": {
			DBType:         "long",
			GoType:         "int64",
			GoTypeInfo:     GoType{Name: "int64", IsNumeric: true},
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
		"double": {
			DBType:         "double",
			GoType:         "float64",
			GoTypeInfo:     GoType{Name: "float64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"float": {
			DBType:         "float",
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
		"date": {
			DBType:         "date",
			GoType:         "time.Time",
			GoTypeInfo:     GoType{Name: "time.Time", IsTime: true},
			TypeScriptType: TypeScriptType{Name: "Date"},
			ColumnFlags:    ColumnFlags{IsTime: true},
		},
		"timestamp": {
			DBType:         "timestamp",
			GoType:         "time.Time",
			GoTypeInfo:     GoType{Name: "time.Time", IsTime: true},
			TypeScriptType: TypeScriptType{Name: "Date"},
			ColumnFlags:    ColumnFlags{IsTime: true},
		},
		"bindata": {
			DBType:         "bindata",
			GoType:         "[]byte",
			GoTypeInfo:     GoType{Name: "[]byte", IsBinary: true},
			TypeScriptType: TypeScriptType{Name: "Uint8Array"},
			ColumnFlags:    ColumnFlags{IsBinary: true},
		},
		"binary": {
			DBType:         "binary",
			GoType:         "[]byte",
			GoTypeInfo:     GoType{Name: "[]byte", IsBinary: true},
			TypeScriptType: TypeScriptType{Name: "Uint8Array"},
			ColumnFlags:    ColumnFlags{IsBinary: true},
		},
		"undefined": {
			DBType:         "undefined",
			GoType:         "interface{}",
			GoTypeInfo:     GoType{Name: "interface{}", IsInterface: true},
			TypeScriptType: TypeScriptType{Name: "undefined"},
			ColumnFlags:    ColumnFlags{},
		},
		"null": {
			DBType:         "null",
			GoType:         "interface{}",
			GoTypeInfo:     GoType{Name: "interface{}", IsInterface: true},
			TypeScriptType: TypeScriptType{Name: "null"},
			ColumnFlags:    ColumnFlags{},
		},
		"regex": {
			DBType:         "regex",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"dbpointer": {
			DBType:         "dbpointer",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"javascript": {
			DBType:         "javascript",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"javascriptwithscope": {
			DBType:         "javascriptwithscope",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"symbol": {
			DBType:         "symbol",
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
		"object": {
			DBType:         "object",
			GoType:         "map[string]any",
			GoTypeInfo:     GoType{Name: "map[string]any", IsMap: true},
			TypeScriptType: TypeScriptType{Name: "Record<string, any>", Generic: "string, any"},
			ColumnFlags:    ColumnFlags{IsMap: true},
		},
		"geopoint": {
			DBType:         "geopoint",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"geoshape": {
			DBType:         "geoshape",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
	}

	if mapping, ok := mappings[dbType]; ok {
		return mapping
	}

	return m.fallbackMapping(dbType)
}

func getMongoDBMappings() []*TypeMapping {
	return []*TypeMapping{}
}
