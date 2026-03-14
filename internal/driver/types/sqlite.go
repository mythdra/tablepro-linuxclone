package types

func (m *TypeMapper) mapSQLiteType(dbType string) *TypeMapping {
	mappings := map[string]*TypeMapping{
		"integer": {
			DBType:         "integer",
			GoType:         "int64",
			GoTypeInfo:     GoType{Name: "int64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"int": {
			DBType:         "int",
			GoType:         "int64",
			GoTypeInfo:     GoType{Name: "int64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"tinyint": {
			DBType:         "tinyint",
			GoType:         "int64",
			GoTypeInfo:     GoType{Name: "int64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"smallint": {
			DBType:         "smallint",
			GoType:         "int64",
			GoTypeInfo:     GoType{Name: "int64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"mediumint": {
			DBType:         "mediumint",
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
		"unsigned big int": {
			DBType:         "unsigned big int",
			GoType:         "int64",
			GoTypeInfo:     GoType{Name: "int64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"int2": {
			DBType:         "int2",
			GoType:         "int64",
			GoTypeInfo:     GoType{Name: "int64", IsNumeric: true},
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
		"real": {
			DBType:         "real",
			GoType:         "float64",
			GoTypeInfo:     GoType{Name: "float64", IsNumeric: true},
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
		"text": {
			DBType:         "text",
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
		"varchar": {
			DBType:         "varchar",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"varying character": {
			DBType:         "varying character",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"nchar": {
			DBType:         "nchar",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"native character": {
			DBType:         "native character",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"nvarchar": {
			DBType:         "nvarchar",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"clob": {
			DBType:         "clob",
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
		"none": {
			DBType:         "none",
			GoType:         "interface{}",
			GoTypeInfo:     GoType{Name: "interface{}", IsInterface: true},
			TypeScriptType: TypeScriptType{Name: "null"},
			ColumnFlags:    ColumnFlags{},
		},
		"null": {
			DBType:         "null",
			GoType:         "interface{}",
			GoTypeInfo:     GoType{Name: "interface{}", IsInterface: true},
			TypeScriptType: TypeScriptType{Name: "null"},
			ColumnFlags:    ColumnFlags{},
		},
	}

	if mapping, ok := mappings[dbType]; ok {
		return mapping
	}

	return m.fallbackMapping(dbType)
}

func getSQLiteMappings() []*TypeMapping {
	return []*TypeMapping{}
}
