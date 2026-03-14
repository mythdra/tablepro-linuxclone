package types

func (m *TypeMapper) mapMSSQLType(dbType string) *TypeMapping {
	mappings := map[string]*TypeMapping{
		"bigint": {
			DBType:         "bigint",
			GoType:         "int64",
			GoTypeInfo:     GoType{Name: "int64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"binary": {
			DBType:         "binary",
			GoType:         "[]byte",
			GoTypeInfo:     GoType{Name: "[]byte", IsBinary: true},
			TypeScriptType: TypeScriptType{Name: "Uint8Array"},
			ColumnFlags:    ColumnFlags{IsBinary: true},
		},
		"bit": {
			DBType:         "bit",
			GoType:         "bool",
			GoTypeInfo:     GoType{Name: "bool", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "boolean"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"char": {
			DBType:         "char",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"datetime": {
			DBType:         "datetime",
			GoType:         "time.Time",
			GoTypeInfo:     GoType{Name: "time.Time", IsTime: true},
			TypeScriptType: TypeScriptType{Name: "Date"},
			ColumnFlags:    ColumnFlags{IsTime: true},
		},
		"datetime2": {
			DBType:         "datetime2",
			GoType:         "time.Time",
			GoTypeInfo:     GoType{Name: "time.Time", IsTime: true},
			TypeScriptType: TypeScriptType{Name: "Date"},
			ColumnFlags:    ColumnFlags{IsTime: true},
		},
		"datetimeoffset": {
			DBType:         "datetimeoffset",
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
		"decimal": {
			DBType:         "decimal",
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
		"image": {
			DBType:         "image",
			GoType:         "[]byte",
			GoTypeInfo:     GoType{Name: "[]byte", IsBinary: true},
			TypeScriptType: TypeScriptType{Name: "Uint8Array"},
			ColumnFlags:    ColumnFlags{IsBinary: true},
		},
		"int": {
			DBType:         "int",
			GoType:         "int32",
			GoTypeInfo:     GoType{Name: "int32", IsNumeric: true},
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
		"nchar": {
			DBType:         "nchar",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"ntext": {
			DBType:         "ntext",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"numeric": {
			DBType:         "numeric",
			GoType:         "float64",
			GoTypeInfo:     GoType{Name: "float64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"nvarchar": {
			DBType:         "nvarchar",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"real": {
			DBType:         "real",
			GoType:         "float32",
			GoTypeInfo:     GoType{Name: "float32", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"smalldatetime": {
			DBType:         "smalldatetime",
			GoType:         "time.Time",
			GoTypeInfo:     GoType{Name: "time.Time", IsTime: true},
			TypeScriptType: TypeScriptType{Name: "Date"},
			ColumnFlags:    ColumnFlags{IsTime: true},
		},
		"smallint": {
			DBType:         "smallint",
			GoType:         "int16",
			GoTypeInfo:     GoType{Name: "int16", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"smallmoney": {
			DBType:         "smallmoney",
			GoType:         "float64",
			GoTypeInfo:     GoType{Name: "float64", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"sql_variant": {
			DBType:         "sql_variant",
			GoType:         "interface{}",
			GoTypeInfo:     GoType{Name: "interface{}", IsInterface: true},
			TypeScriptType: TypeScriptType{Name: "any"},
			ColumnFlags:    ColumnFlags{},
		},
		"text": {
			DBType:         "text",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
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
			GoType:         "[]byte",
			GoTypeInfo:     GoType{Name: "[]byte", IsBinary: true},
			TypeScriptType: TypeScriptType{Name: "Uint8Array"},
			ColumnFlags:    ColumnFlags{IsBinary: true},
		},
		"tinyint": {
			DBType:         "tinyint",
			GoType:         "uint8",
			GoTypeInfo:     GoType{Name: "uint8", IsNumeric: true},
			TypeScriptType: TypeScriptType{Name: "number"},
			ColumnFlags:    ColumnFlags{IsNumeric: true},
		},
		"uniqueidentifier": {
			DBType:         "uniqueidentifier",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true, IsUUID: true},
		},
		"varbinary": {
			DBType:         "varbinary",
			GoType:         "[]byte",
			GoTypeInfo:     GoType{Name: "[]byte", IsBinary: true},
			TypeScriptType: TypeScriptType{Name: "Uint8Array"},
			ColumnFlags:    ColumnFlags{IsBinary: true},
		},
		"varchar": {
			DBType:         "varchar",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"xml": {
			DBType:         "xml",
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

func getMSSQLMappings() []*TypeMapping {
	return []*TypeMapping{}
}
