package change

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tablepro/internal/driver"
)

func TestAllDialects_ParamMarkers(t *testing.T) {
	tests := []struct {
		name     string
		dbType   driver.DatabaseType
		index    int
		expected string
	}{
		{"PostgreSQL $1", driver.DatabaseTypePostgreSQL, 1, "$1"},
		{"PostgreSQL $2", driver.DatabaseTypePostgreSQL, 2, "$2"},
		{"PostgreSQL $10", driver.DatabaseTypePostgreSQL, 10, "$10"},
		{"MySQL ?", driver.DatabaseTypeMySQL, 1, "?"},
		{"MySQL ? index 5", driver.DatabaseTypeMySQL, 5, "?"},
		{"SQLite ?", driver.DatabaseTypeSQLite, 1, "?"},
		{"SQLite ? index 3", driver.DatabaseTypeSQLite, 3, "?"},
		{"DuckDB $1", driver.DatabaseTypeDuckDB, 1, "$1"},
		{"DuckDB $3", driver.DatabaseTypeDuckDB, 3, "$3"},
		{"MSSQL @p1", driver.DatabaseTypeMSSQL, 1, "@p1"},
		{"MSSQL @p5", driver.DatabaseTypeMSSQL, 5, "@p5"},
		{"MSSQL @p100", driver.DatabaseTypeMSSQL, 100, "@p100"},
		{"ClickHouse ?", driver.DatabaseTypeClickHouse, 1, "?"},
		{"ClickHouse ? index 7", driver.DatabaseTypeClickHouse, 7, "?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialect := GetDialect(tt.dbType)
			result := dialect.ParamMarker(tt.index)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAllDialects_QuoteIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		dbType   driver.DatabaseType
		input    string
		expected string
	}{
		{"PostgreSQL double quotes", driver.DatabaseTypePostgreSQL, "col", `"col"`},
		{"MySQL backticks", driver.DatabaseTypeMySQL, "col", "`col`"},
		{"SQLite double quotes", driver.DatabaseTypeSQLite, "col", `"col"`},
		{"DuckDB double quotes", driver.DatabaseTypeDuckDB, "col", `"col"`},
		{"MSSQL brackets", driver.DatabaseTypeMSSQL, "col", "[col]"},
		{"ClickHouse double quotes", driver.DatabaseTypeClickHouse, "col", `"col"`},
		{"PostgreSQL with space", driver.DatabaseTypePostgreSQL, "my column", `"my column"`},
		{"MySQL with space", driver.DatabaseTypeMySQL, "my column", "`my column`"},
		{"MSSQL with space", driver.DatabaseTypeMSSQL, "my column", "[my column]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialect := GetDialect(tt.dbType)
			result := dialect.QuoteIdentifier(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDialect_PostgreSQL_Specifics(t *testing.T) {
	d := GetDialect(driver.DatabaseTypePostgreSQL)

	assert.Equal(t, driver.DatabaseTypePostgreSQL, d.DatabaseType)
	assert.Equal(t, `"`, d.QuoteChar)
	assert.Equal(t, "dollar", d.ParamStyle)
	assert.Equal(t, "$", d.ParamPrefix)
	assert.True(t, d.SupportsSchema)

	tableName := d.QualifiedTableName("public", "users")
	assert.Equal(t, `"public"."users"`, tableName)

	tableNameNoSchema := d.QualifiedTableName("", "users")
	assert.Equal(t, `"users"`, tableNameNoSchema)

	where, nextIdx := d.BuildWhereClause([]string{"id", "tenant_id"}, 1)
	assert.Equal(t, `"id" = $1 AND "tenant_id" = $2`, where)
	assert.Equal(t, 3, nextIdx)
}

func TestDialect_MySQL_Specifics(t *testing.T) {
	d := GetDialect(driver.DatabaseTypeMySQL)

	assert.Equal(t, driver.DatabaseTypeMySQL, d.DatabaseType)
	assert.Equal(t, "`", d.QuoteChar)
	assert.Equal(t, "question", d.ParamStyle)
	assert.False(t, d.SupportsSchema)

	tableName := d.QualifiedTableName("mydb", "users")
	assert.Equal(t, "`users`", tableName)

	where, nextIdx := d.BuildWhereClause([]string{"id"}, 1)
	assert.Equal(t, "`id` = ?", where)
	assert.Equal(t, 2, nextIdx)
}

func TestDialect_MSSQL_Specifics(t *testing.T) {
	d := GetDialect(driver.DatabaseTypeMSSQL)

	assert.Equal(t, driver.DatabaseTypeMSSQL, d.DatabaseType)
	assert.Equal(t, "[", d.QuoteChar)
	assert.Equal(t, "at", d.ParamStyle)
	assert.Equal(t, "@p", d.ParamPrefix)
	assert.True(t, d.SupportsSchema)

	tableName := d.QualifiedTableName("dbo", "users")
	assert.Equal(t, "[dbo].[users]", tableName)

	where, nextIdx := d.BuildWhereClause([]string{"id", "version"}, 3)
	assert.Equal(t, "[id] = @p3 AND [version] = @p4", where)
	assert.Equal(t, 5, nextIdx)
}

func TestDialect_SQLite_Specifics(t *testing.T) {
	d := GetDialect(driver.DatabaseTypeSQLite)

	assert.Equal(t, driver.DatabaseTypeSQLite, d.DatabaseType)
	assert.Equal(t, `"`, d.QuoteChar)
	assert.Equal(t, "question", d.ParamStyle)
	assert.False(t, d.SupportsSchema)

	tableName := d.QualifiedTableName("main", "users")
	assert.Equal(t, `"users"`, tableName)

	where, _ := d.BuildWhereClause([]string{"rowid"}, 1)
	assert.Equal(t, `"rowid" = ?`, where)
}

func TestDialect_DuckDB_Specifics(t *testing.T) {
	d := GetDialect(driver.DatabaseTypeDuckDB)

	assert.Equal(t, driver.DatabaseTypeDuckDB, d.DatabaseType)
	assert.Equal(t, `"`, d.QuoteChar)
	assert.Equal(t, "dollar", d.ParamStyle)
	assert.True(t, d.SupportsSchema)

	tableName := d.QualifiedTableName("main", "data")
	assert.Equal(t, `"main"."data"`, tableName)
}

func TestDialect_ClickHouse_Specifics(t *testing.T) {
	d := GetDialect(driver.DatabaseTypeClickHouse)

	assert.Equal(t, driver.DatabaseTypeClickHouse, d.DatabaseType)
	assert.Equal(t, `"`, d.QuoteChar)
	assert.Equal(t, "question", d.ParamStyle)
	assert.True(t, d.SupportsSchema)

	tableName := d.QualifiedTableName("default", "events")
	assert.Equal(t, `"default"."events"`, tableName)
}

func TestDialect_Unknown_Defaults(t *testing.T) {
	d := GetDialect(driver.DatabaseTypeUnknown)

	assert.Equal(t, `"`, d.QuoteChar)
	assert.Equal(t, "question", d.ParamStyle)
	assert.False(t, d.SupportsSchema)
}
