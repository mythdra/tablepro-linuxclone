package change

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tablepro/internal/driver"
)

func makeTabChanges(dbType driver.DatabaseType, tableName string, pks []string, columns []driver.ColumnInfo) *TabChanges {
	return NewTabChanges("test-tab", tableName, "", dbType, pks, columns)
}

func TestGenerator_UpdateAllDialects(t *testing.T) {
	columns := []driver.ColumnInfo{
		{Name: "id", IsPrimaryKey: true},
		{Name: "name"},
		{Name: "email"},
	}

	tests := []struct {
		name        string
		dbType      driver.DatabaseType
		expectedSQL string
	}{
		{
			"PostgreSQL",
			driver.DatabaseTypePostgreSQL,
			`UPDATE "users" SET "name" = $1 WHERE "id" = $2`,
		},
		{
			"MySQL",
			driver.DatabaseTypeMySQL,
			"UPDATE `users` SET `name` = ? WHERE `id` = ?",
		},
		{
			"SQLite",
			driver.DatabaseTypeSQLite,
			`UPDATE "users" SET "name" = ? WHERE "id" = ?`,
		},
		{
			"DuckDB",
			driver.DatabaseTypeDuckDB,
			`UPDATE "users" SET "name" = $1 WHERE "id" = $2`,
		},
		{
			"MSSQL",
			driver.DatabaseTypeMSSQL,
			`UPDATE [users] SET [name] = @p1 WHERE [id] = @p2`,
		},
		{
			"ClickHouse",
			driver.DatabaseTypeClickHouse,
			`UPDATE "users" SET "name" = ? WHERE "id" = ?`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := makeTabChanges(tt.dbType, "users", []string{"id"}, columns)
			tc.CellChanges["0:name"] = []CellChange{
				{RowIndex: 0, ColumnName: "name", OldValue: "old", NewValue: "new"},
			}

			gen := NewSQLStatementGenerator(tt.dbType)
			stmts, err := gen.GenerateUpdate(tc, 0, map[string]any{"id": 1})
			require.NoError(t, err)
			require.Len(t, stmts, 1)
			assert.Equal(t, tt.expectedSQL, stmts[0].SQL)
			assert.Equal(t, []any{"new", 1}, stmts[0].Params)
			assert.Equal(t, ChangeActionUpdate, stmts[0].Action)
		})
	}
}

func TestGenerator_InsertAllDialects(t *testing.T) {
	columns := []driver.ColumnInfo{
		{Name: "id", IsPrimaryKey: true, IsAutoIncrement: true},
		{Name: "name"},
		{Name: "email"},
	}

	tests := []struct {
		name        string
		dbType      driver.DatabaseType
		expectedSQL string
	}{
		{
			"PostgreSQL",
			driver.DatabaseTypePostgreSQL,
			`INSERT INTO "users" ("name", "email") VALUES ($1, $2)`,
		},
		{
			"MySQL",
			driver.DatabaseTypeMySQL,
			"INSERT INTO `users` (`name`, `email`) VALUES (?, ?)",
		},
		{
			"SQLite",
			driver.DatabaseTypeSQLite,
			`INSERT INTO "users" ("name", "email") VALUES (?, ?)`,
		},
		{
			"DuckDB",
			driver.DatabaseTypeDuckDB,
			`INSERT INTO "users" ("name", "email") VALUES ($1, $2)`,
		},
		{
			"MSSQL",
			driver.DatabaseTypeMSSQL,
			`INSERT INTO [users] ([name], [email]) VALUES (@p1, @p2)`,
		},
		{
			"ClickHouse",
			driver.DatabaseTypeClickHouse,
			`INSERT INTO "users" ("name", "email") VALUES (?, ?)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := makeTabChanges(tt.dbType, "users", []string{"id"}, columns)
			row := InsertedRow{
				TempID: "tmp-1",
				Values: map[string]any{"name": "Alice", "email": "alice@test.com"},
			}

			gen := NewSQLStatementGenerator(tt.dbType)
			stmt, err := gen.GenerateInsert(tc, row)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedSQL, stmt.SQL)
			assert.Equal(t, []any{"Alice", "alice@test.com"}, stmt.Params)
			assert.Equal(t, ChangeActionInsert, stmt.Action)
		})
	}
}

func TestGenerator_DeleteAllDialects(t *testing.T) {
	columns := []driver.ColumnInfo{
		{Name: "id", IsPrimaryKey: true},
		{Name: "name"},
	}

	tests := []struct {
		name        string
		dbType      driver.DatabaseType
		expectedSQL string
	}{
		{
			"PostgreSQL",
			driver.DatabaseTypePostgreSQL,
			`DELETE FROM "users" WHERE "id" = $1`,
		},
		{
			"MySQL",
			driver.DatabaseTypeMySQL,
			"DELETE FROM `users` WHERE `id` = ?",
		},
		{
			"SQLite",
			driver.DatabaseTypeSQLite,
			`DELETE FROM "users" WHERE "id" = ?`,
		},
		{
			"DuckDB",
			driver.DatabaseTypeDuckDB,
			`DELETE FROM "users" WHERE "id" = $1`,
		},
		{
			"MSSQL",
			driver.DatabaseTypeMSSQL,
			`DELETE FROM [users] WHERE [id] = @p1`,
		},
		{
			"ClickHouse",
			driver.DatabaseTypeClickHouse,
			`DELETE FROM "users" WHERE "id" = ?`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := makeTabChanges(tt.dbType, "users", []string{"id"}, columns)
			row := DeletedRow{
				RowIndex:    0,
				PrimaryKeys: map[string]any{"id": 42},
			}

			gen := NewSQLStatementGenerator(tt.dbType)
			stmt, err := gen.GenerateDelete(tc, row)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedSQL, stmt.SQL)
			assert.Equal(t, []any{42}, stmt.Params)
			assert.Equal(t, ChangeActionDelete, stmt.Action)
		})
	}
}

func TestGenerator_CompositeKey(t *testing.T) {
	columns := []driver.ColumnInfo{
		{Name: "tenant_id", IsPrimaryKey: true},
		{Name: "user_id", IsPrimaryKey: true},
		{Name: "name"},
	}
	tc := makeTabChanges(driver.DatabaseTypePostgreSQL, "users", []string{"tenant_id", "user_id"}, columns)
	tc.CellChanges["0:name"] = []CellChange{
		{RowIndex: 0, ColumnName: "name", OldValue: "old", NewValue: "new"},
	}

	gen := NewSQLStatementGenerator(driver.DatabaseTypePostgreSQL)
	stmts, err := gen.GenerateUpdate(tc, 0, map[string]any{"tenant_id": "t1", "user_id": 5})
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, `UPDATE "users" SET "name" = $1 WHERE "tenant_id" = $2 AND "user_id" = $3`, stmts[0].SQL)
	assert.Equal(t, []any{"new", "t1", 5}, stmts[0].Params)
}

func TestGenerator_NullValue(t *testing.T) {
	columns := []driver.ColumnInfo{
		{Name: "id", IsPrimaryKey: true},
		{Name: "bio"},
	}
	tc := makeTabChanges(driver.DatabaseTypePostgreSQL, "users", []string{"id"}, columns)
	tc.CellChanges["0:bio"] = []CellChange{
		{RowIndex: 0, ColumnName: "bio", OldValue: "some text", NewValue: nil},
	}

	gen := NewSQLStatementGenerator(driver.DatabaseTypePostgreSQL)
	stmts, err := gen.GenerateUpdate(tc, 0, map[string]any{"id": 1})
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Contains(t, stmts[0].SQL, `"bio" = $1`)
	assert.Nil(t, stmts[0].Params[0])
}

func TestGenerator_ExcludeAutoIncrement(t *testing.T) {
	columns := []driver.ColumnInfo{
		{Name: "id", IsPrimaryKey: true, IsAutoIncrement: true},
		{Name: "serial_no", IsAutoIncrement: true},
		{Name: "name"},
		{Name: "email"},
	}
	tc := makeTabChanges(driver.DatabaseTypePostgreSQL, "users", []string{"id"}, columns)
	row := InsertedRow{
		TempID: "tmp-1",
		Values: map[string]any{"id": 999, "serial_no": 100, "name": "Test", "email": "t@t.com"},
	}

	gen := NewSQLStatementGenerator(driver.DatabaseTypePostgreSQL)
	stmt, err := gen.GenerateInsert(tc, row)
	require.NoError(t, err)
	assert.NotContains(t, stmt.SQL, `"id"`)
	assert.NotContains(t, stmt.SQL, `"serial_no"`)
	assert.Contains(t, stmt.SQL, `"name"`)
	assert.Contains(t, stmt.SQL, `"email"`)
	assert.Len(t, stmt.Params, 2)
}

func TestGenerator_EmptyChanges(t *testing.T) {
	columns := []driver.ColumnInfo{
		{Name: "id", IsPrimaryKey: true},
		{Name: "name"},
	}
	tc := makeTabChanges(driver.DatabaseTypePostgreSQL, "users", []string{"id"}, columns)

	gen := NewSQLStatementGenerator(driver.DatabaseTypePostgreSQL)
	stmts, err := gen.GenerateAll(tc, map[int]map[string]any{})
	require.NoError(t, err)
	assert.Empty(t, stmts)
}

func TestGenerator_BatchOrder(t *testing.T) {
	columns := []driver.ColumnInfo{
		{Name: "id", IsPrimaryKey: true, IsAutoIncrement: true},
		{Name: "name"},
	}
	tc := makeTabChanges(driver.DatabaseTypeSQLite, "users", []string{"id"}, columns)

	tc.DeletedRows = append(tc.DeletedRows, DeletedRow{
		RowIndex:    0,
		PrimaryKeys: map[string]any{"id": 1},
	})

	tc.CellChanges["1:name"] = []CellChange{
		{RowIndex: 1, ColumnName: "name", OldValue: "old", NewValue: "updated"},
	}

	tc.InsertedRows = append(tc.InsertedRows, InsertedRow{
		TempID: "tmp-1",
		Values: map[string]any{"name": "New User"},
	})

	gen := NewSQLStatementGenerator(driver.DatabaseTypeSQLite)
	rowPKs := map[int]map[string]any{
		1: {"id": 2},
	}

	stmts, err := gen.GenerateAll(tc, rowPKs)
	require.NoError(t, err)
	require.Len(t, stmts, 3)

	assert.Equal(t, ChangeActionDelete, stmts[0].Action)
	assert.Equal(t, ChangeActionUpdate, stmts[1].Action)
	assert.Equal(t, ChangeActionInsert, stmts[2].Action)
}

func TestGenerator_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name       string
		dbType     driver.DatabaseType
		columnName string
		wantQuoted string
	}{
		{"PG column with space", driver.DatabaseTypePostgreSQL, "first name", `"first name"`},
		{"MySQL column with space", driver.DatabaseTypeMySQL, "first name", "`first name`"},
		{"MSSQL column with space", driver.DatabaseTypeMSSQL, "first name", "[first name]"},
		{"PG reserved word", driver.DatabaseTypePostgreSQL, "select", `"select"`},
		{"MySQL reserved word", driver.DatabaseTypeMySQL, "select", "`select`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := GetDialect(tt.dbType)
			assert.Equal(t, tt.wantQuoted, d.QuoteIdentifier(tt.columnName))
		})
	}
}

func TestGenerator_DeleteNoPrimaryKey(t *testing.T) {
	columns := []driver.ColumnInfo{{Name: "name"}}
	tc := makeTabChanges(driver.DatabaseTypePostgreSQL, "users", []string{}, columns)

	gen := NewSQLStatementGenerator(driver.DatabaseTypePostgreSQL)
	_, err := gen.GenerateDelete(tc, DeletedRow{RowIndex: 0, PrimaryKeys: map[string]any{}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "primary keys")
}

func TestGenerator_UpdateNoChangesForRow(t *testing.T) {
	columns := []driver.ColumnInfo{
		{Name: "id", IsPrimaryKey: true},
		{Name: "name"},
	}
	tc := makeTabChanges(driver.DatabaseTypePostgreSQL, "users", []string{"id"}, columns)

	gen := NewSQLStatementGenerator(driver.DatabaseTypePostgreSQL)
	stmts, err := gen.GenerateUpdate(tc, 99, map[string]any{"id": 99})
	require.NoError(t, err)
	assert.Nil(t, stmts)
}

func TestGenerator_SchemaQualifiedTable(t *testing.T) {
	columns := []driver.ColumnInfo{
		{Name: "id", IsPrimaryKey: true},
		{Name: "name"},
	}
	tc := NewTabChanges("test-tab", "users", "public", driver.DatabaseTypePostgreSQL, []string{"id"}, columns)
	tc.CellChanges["0:name"] = []CellChange{
		{RowIndex: 0, ColumnName: "name", OldValue: "a", NewValue: "b"},
	}

	gen := NewSQLStatementGenerator(driver.DatabaseTypePostgreSQL)
	stmts, err := gen.GenerateUpdate(tc, 0, map[string]any{"id": 1})
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Contains(t, stmts[0].SQL, `"public"."users"`)
}
