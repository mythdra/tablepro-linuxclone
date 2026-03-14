package change

import (
	"strings"
	"testing"

	"tablepro/internal/driver"
)

func TestGenerateUpdate_PostgreSQL(t *testing.T) {
	gen := NewSQLStatementGenerator(driver.DatabaseTypePostgreSQL)
	change := &CellChange{
		RowIndex:   0,
		ColumnName: "email",
		OldValue:   "old@test.com",
		NewValue:   "new@test.com",
		PrimaryKeys: map[string]any{
			"id": 42,
		},
	}

	sql, args, err := gen.GenerateUpdate(change, "public", "users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSQL := `UPDATE "public"."users" SET "email" = $1 WHERE "id" = $2`
	if sql != expectedSQL {
		t.Errorf("SQL mismatch\ngot:  %s\nwant: %s", sql, expectedSQL)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "new@test.com" {
		t.Errorf("args[0] = %v, want new@test.com", args[0])
	}
	if args[1] != 42 {
		t.Errorf("args[1] = %v, want 42", args[1])
	}
}

func TestGenerateUpdate_MySQL(t *testing.T) {
	gen := NewSQLStatementGenerator(driver.DatabaseTypeMySQL)
	change := &CellChange{
		RowIndex:   0,
		ColumnName: "name",
		OldValue:   "old",
		NewValue:   "new",
		PrimaryKeys: map[string]any{
			"id": 1,
		},
	}

	sql, args, err := gen.GenerateUpdate(change, "", "users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSQL := "UPDATE `users` SET `name` = ? WHERE `id` = ?"
	if sql != expectedSQL {
		t.Errorf("SQL mismatch\ngot:  %s\nwant: %s", sql, expectedSQL)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
}

func TestGenerateUpdate_MSSQL(t *testing.T) {
	gen := NewSQLStatementGenerator(driver.DatabaseTypeMSSQL)
	change := &CellChange{
		RowIndex:   0,
		ColumnName: "status",
		OldValue:   "active",
		NewValue:   "inactive",
		PrimaryKeys: map[string]any{
			"id": 99,
		},
	}

	sql, args, err := gen.GenerateUpdate(change, "dbo", "orders")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSQL := "UPDATE [dbo].[orders] SET [status] = @p1 WHERE [id] = @p2"
	if sql != expectedSQL {
		t.Errorf("SQL mismatch\ngot:  %s\nwant: %s", sql, expectedSQL)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "inactive" {
		t.Errorf("args[0] = %v, want inactive", args[0])
	}
	if args[1] != 99 {
		t.Errorf("args[1] = %v, want 99", args[1])
	}
}

func TestGenerateInsert_ExcludesAutoIncrement(t *testing.T) {
	gen := NewSQLStatementGenerator(driver.DatabaseTypePostgreSQL)
	row := &InsertedRow{
		TempID: "tmp-1",
		Values: map[string]any{
			"id":    nil,
			"name":  "Alice",
			"email": "alice@test.com",
		},
	}

	columns := []string{"id", "name", "email"}
	autoInc := []string{"id"}

	sql, args, err := gen.GenerateInsert(row, "public", "users", columns, autoInc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(sql, `"id"`) {
		t.Errorf("SQL should not contain auto-increment column 'id': %s", sql)
	}
	if !strings.Contains(sql, `"name"`) || !strings.Contains(sql, `"email"`) {
		t.Errorf("SQL missing non-auto-increment columns: %s", sql)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
}

func TestGenerateInsert_WithNullValues(t *testing.T) {
	gen := NewSQLStatementGenerator(driver.DatabaseTypeSQLite)
	row := &InsertedRow{
		TempID: "tmp-2",
		Values: map[string]any{
			"name":  "Bob",
			"email": nil,
		},
	}

	columns := []string{"name", "email"}

	sql, args, err := gen.GenerateInsert(row, "", "contacts", columns, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSQL := `INSERT INTO "contacts" ("name", "email") VALUES (?, ?)`
	if sql != expectedSQL {
		t.Errorf("SQL mismatch\ngot:  %s\nwant: %s", sql, expectedSQL)
	}

	foundNil := false
	for _, arg := range args {
		if arg == nil {
			foundNil = true
			break
		}
	}
	if !foundNil {
		t.Error("expected nil in args for NULL column value")
	}
}

func TestGenerateDelete_CompositeKey(t *testing.T) {
	gen := NewSQLStatementGenerator(driver.DatabaseTypePostgreSQL)
	row := &DeletedRow{
		RowIndex: 5,
		PrimaryKeys: map[string]any{
			"order_id":   100,
			"product_id": 200,
		},
	}

	sql, args, err := gen.GenerateDelete(row, "public", "order_items")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSQL := `DELETE FROM "public"."order_items" WHERE "order_id" = $1 AND "product_id" = $2`
	if sql != expectedSQL {
		t.Errorf("SQL mismatch\ngot:  %s\nwant: %s", sql, expectedSQL)
	}
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != 100 {
		t.Errorf("args[0] = %v, want 100", args[0])
	}
	if args[1] != 200 {
		t.Errorf("args[1] = %v, want 200", args[1])
	}
}

func TestGenerateAll_OrderIsDeleteUpdateInsert(t *testing.T) {
	gen := NewSQLStatementGenerator(driver.DatabaseTypePostgreSQL)
	changes := &TabChanges{
		SchemaName: "public",
		TableName:  "users",
		CellChanges: []CellChange{
			{
				ColumnName:  "name",
				NewValue:    "updated",
				PrimaryKeys: map[string]any{"id": 1},
			},
		},
		InsertedRows: []InsertedRow{
			{
				TempID: "tmp-1",
				Values: map[string]any{"name": "new-user"},
			},
		},
		DeletedRows: []DeletedRow{
			{
				RowIndex:    2,
				PrimaryKeys: map[string]any{"id": 3},
			},
		},
	}

	stmts, err := gen.GenerateAll(changes, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stmts) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(stmts))
	}
	if stmts[0].Type != "DELETE" {
		t.Errorf("statement[0].Type = %s, want DELETE", stmts[0].Type)
	}
	if stmts[1].Type != "UPDATE" {
		t.Errorf("statement[1].Type = %s, want UPDATE", stmts[1].Type)
	}
	if stmts[2].Type != "INSERT" {
		t.Errorf("statement[2].Type = %s, want INSERT", stmts[2].Type)
	}
}

func TestAllDialects(t *testing.T) {
	dialects := []struct {
		dbType          driver.DatabaseType
		wantParamMarker string
		wantQuotedCol   string
	}{
		{driver.DatabaseTypePostgreSQL, "$1", `"col"`},
		{driver.DatabaseTypeMySQL, "?", "`col`"},
		{driver.DatabaseTypeSQLite, "?", `"col"`},
		{driver.DatabaseTypeDuckDB, "?", `"col"`},
		{driver.DatabaseTypeMSSQL, "@p1", "[col]"},
		{driver.DatabaseTypeClickHouse, "?", `"col"`},
	}

	for _, tt := range dialects {
		t.Run(string(tt.dbType), func(t *testing.T) {
			d := GetDialect(tt.dbType)

			gotParam := d.ParamMarker(1)
			if gotParam != tt.wantParamMarker {
				t.Errorf("ParamMarker(1) = %s, want %s", gotParam, tt.wantParamMarker)
			}

			gotQuote := d.QuoteIdentifier("col")
			if gotQuote != tt.wantQuotedCol {
				t.Errorf("QuoteIdentifier(col) = %s, want %s", gotQuote, tt.wantQuotedCol)
			}
		})
	}
}

func TestGenerateUpdate_NilChange(t *testing.T) {
	gen := NewSQLStatementGenerator(driver.DatabaseTypePostgreSQL)
	_, _, err := gen.GenerateUpdate(nil, "public", "users")
	if err == nil {
		t.Fatal("expected error for nil change")
	}
}

func TestGenerateUpdate_NoPrimaryKeys(t *testing.T) {
	gen := NewSQLStatementGenerator(driver.DatabaseTypePostgreSQL)
	change := &CellChange{
		ColumnName:  "name",
		NewValue:    "test",
		PrimaryKeys: map[string]any{},
	}
	_, _, err := gen.GenerateUpdate(change, "public", "users")
	if err == nil {
		t.Fatal("expected error for empty primary keys")
	}
}

func TestGenerateAll_EmptyChanges(t *testing.T) {
	gen := NewSQLStatementGenerator(driver.DatabaseTypePostgreSQL)
	changes := &TabChanges{
		SchemaName: "public",
		TableName:  "users",
	}

	stmts, err := gen.GenerateAll(changes, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stmts != nil {
		t.Errorf("expected nil for empty changes, got %d statements", len(stmts))
	}
}
