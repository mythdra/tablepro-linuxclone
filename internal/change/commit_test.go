package change

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"tablepro/internal/driver"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT
	)`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	_, err = db.Exec(`INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@test.com'), (2, 'Bob', 'bob@test.com'), (3, 'Charlie', 'charlie@test.com')`)
	if err != nil {
		t.Fatalf("failed to seed data: %v", err)
	}
	return db
}

func TestCommitChanges_Success(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	changes := &TabChanges{
		SchemaName: "",
		TableName:  "users",
		CellChanges: []CellChange{
			{
				ColumnName:  "name",
				OldValue:    "Alice",
				NewValue:    "Alice Updated",
				PrimaryKeys: map[string]any{"id": 1},
			},
		},
		InsertedRows: []InsertedRow{
			{
				TempID: "tmp-1",
				Values: map[string]any{"name": "Dave", "email": "dave@test.com"},
			},
		},
		DeletedRows: []DeletedRow{
			{
				RowIndex:    2,
				PrimaryKeys: map[string]any{"id": 3},
			},
		},
	}

	err := CommitChanges(context.Background(), db, changes, driver.DatabaseTypeSQLite, []string{"id"})
	if err != nil {
		t.Fatalf("CommitChanges failed: %v", err)
	}

	var name string
	err = db.QueryRow("SELECT name FROM users WHERE id = 1").Scan(&name)
	if err != nil {
		t.Fatalf("query after commit failed: %v", err)
	}
	if name != "Alice Updated" {
		t.Errorf("expected 'Alice Updated', got '%s'", name)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE id = 3").Scan(&count)
	if err != nil {
		t.Fatalf("query after delete failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected row id=3 deleted, but count=%d", count)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE name = 'Dave'").Scan(&count)
	if err != nil {
		t.Fatalf("query after insert failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 inserted row for Dave, got %d", count)
	}
}

func TestCommitChanges_RollbackOnError(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	changes := &TabChanges{
		SchemaName: "",
		TableName:  "users",
		CellChanges: []CellChange{
			{
				ColumnName:  "name",
				OldValue:    "Alice",
				NewValue:    "Alice Updated",
				PrimaryKeys: map[string]any{"id": 1},
			},
		},
		InsertedRows: []InsertedRow{
			{
				TempID: "tmp-bad",
				Values: map[string]any{"name": nil},
			},
		},
	}

	err := CommitChanges(context.Background(), db, changes, driver.DatabaseTypeSQLite, nil)
	if err == nil {
		t.Fatal("expected error due to NOT NULL constraint, got nil")
	}

	var name string
	qErr := db.QueryRow("SELECT name FROM users WHERE id = 1").Scan(&name)
	if qErr != nil {
		t.Fatalf("query after rollback failed: %v", qErr)
	}
	if name != "Alice" {
		t.Errorf("expected 'Alice' (rollback should preserve original), got '%s'", name)
	}
}

func TestCommitChanges_EmptyChanges(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	changes := &TabChanges{
		SchemaName: "",
		TableName:  "users",
	}

	err := CommitChanges(context.Background(), db, changes, driver.DatabaseTypeSQLite, nil)
	if err != nil {
		t.Fatalf("CommitChanges with empty changes should succeed, got: %v", err)
	}
}

func TestCommitChanges_NilChanges(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	err := CommitChanges(context.Background(), db, nil, driver.DatabaseTypeSQLite, nil)
	if err == nil {
		t.Fatal("expected error for nil changes")
	}
}

func TestCommitChanges_CancelledContext(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	changes := &TabChanges{
		SchemaName: "",
		TableName:  "users",
		CellChanges: []CellChange{
			{
				ColumnName:  "name",
				NewValue:    "test",
				PrimaryKeys: map[string]any{"id": 1},
			},
		},
	}

	err := CommitChanges(ctx, db, changes, driver.DatabaseTypeSQLite, nil)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	fmt.Println("cancelled context error:", err)
}
