package types

func (m *TypeMapper) mapRedisType(dbType string) *TypeMapping {
	mappings := map[string]*TypeMapping{
		"string": {
			DBType:         "string",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"list": {
			DBType:         "list",
			GoType:         "[]string",
			GoTypeInfo:     GoType{Name: "[]string", IsArray: true},
			TypeScriptType: TypeScriptType{Name: "string[]", Array: true},
			ColumnFlags:    ColumnFlags{IsArray: true},
		},
		"set": {
			DBType:         "set",
			GoType:         "map[string]struct{}",
			GoTypeInfo:     GoType{Name: "map[string]struct{}", IsMap: true},
			TypeScriptType: TypeScriptType{Name: "Set<string>"},
			ColumnFlags:    ColumnFlags{IsMap: true},
		},
		"zset": {
			DBType:         "zset",
			GoType:         "map[string]float64",
			GoTypeInfo:     GoType{Name: "map[string]float64", IsMap: true},
			TypeScriptType: TypeScriptType{Name: "Record<string, number>"},
			ColumnFlags:    ColumnFlags{IsMap: true},
		},
		"sortedset": {
			DBType:         "sortedset",
			GoType:         "map[string]float64",
			GoTypeInfo:     GoType{Name: "map[string]float64", IsMap: true},
			TypeScriptType: TypeScriptType{Name: "Record<string, number>"},
			ColumnFlags:    ColumnFlags{IsMap: true},
		},
		"hash": {
			DBType:         "hash",
			GoType:         "map[string]string",
			GoTypeInfo:     GoType{Name: "map[string]string", IsMap: true},
			TypeScriptType: TypeScriptType{Name: "Record<string, string>"},
			ColumnFlags:    ColumnFlags{IsMap: true},
		},
		"hmap": {
			DBType:         "hmap",
			GoType:         "map[string]string",
			GoTypeInfo:     GoType{Name: "map[string]string", IsMap: true},
			TypeScriptType: TypeScriptType{Name: "Record<string, string>"},
			ColumnFlags:    ColumnFlags{IsMap: true},
		},
		"stream": {
			DBType:         "stream",
			GoType:         "map[string]any",
			GoTypeInfo:     GoType{Name: "map[string]any", IsMap: true},
			TypeScriptType: TypeScriptType{Name: "Record<string, any>"},
			ColumnFlags:    ColumnFlags{IsMap: true},
		},
		"hyperloglog": {
			DBType:         "hyperloglog",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"bitmap": {
			DBType:         "bitmap",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"bitfield": {
			DBType:         "bitfield",
			GoType:         "string",
			GoTypeInfo:     GoType{Name: "string", IsString: true},
			TypeScriptType: TypeScriptType{Name: "string"},
			ColumnFlags:    ColumnFlags{IsString: true},
		},
		"none": {
			DBType:         "none",
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

func getRedisMappings() []*TypeMapping {
	return []*TypeMapping{}
}
