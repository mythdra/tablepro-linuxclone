//go:build integration

package change

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tablepro/internal/driver"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		age INTEGER
	)`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO users (name, email, age) VALUES
		('Alice', 'alice@test.com', 30),
		('Bob', 'bob@test.com', 25),
		('Charlie', 'charlie@test.com', 35)`)
	require.NoError(t, err)

	return db
}

func setupTestDBWithFK(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE departments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	)`)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE employees (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		dept_id INTEGER NOT NULL,
		FOREIGN KEY (dept_id) REFERENCES departments(id)
	)`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO departments (name) VALUES ('Engineering')`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO employees (name, dept_id) VALUES ('Alice', 1)`)
	require.NoError(t, err)

	return db
}

func TestEndToEnd_EditCommit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewDataChangeManager()
	columns := []driver.ColumnInfo{
		{Name: "id", IsPrimaryKey: true, IsAutoIncrement: true},
		{Name: "name", IsPrimaryKey: false},
		{Name: "email", IsPrimaryKey: false},
		{Name: "age", IsPrimaryKey: false},
	}
	manager.EnsureTab("tab1", "users", "", driver.DatabaseTypeSQLite, []string{"id"}, columns)

	err := manager.TrackCellChange("tab1", 0, "name", "Alice", "Alice Updated", "text")
	require.NoError(t, err)

	err = manager.TrackCellChange("tab1", 1, "age", 25, 26, "integer")
	require.NoError(t, err)

	gen := NewSQLStatementGenerator(driver.DatabaseTypeSQLite)
	tabChanges := manager.GetChanges("tab1")
	require.NotNil(t, tabChanges)

	rowPKs := map[int]map[string]any{
		0: {"id": 1},
		1: {"id": 2},
	}

	statements, err := gen.GenerateAll(tabChanges, rowPKs)
	require.NoError(t, err)
	assert.Len(t, statements, 2)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := CommitChanges(ctx, db, statements)
	require.NoError(t, err)
	assert.Equal(t, 2, result.StatementsExecuted)
	assert.Equal(t, int64(2), result.RowsAffected)

	var name string
	err = db.QueryRow("SELECT name FROM users WHERE id = 1").Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "Alice Updated", name)

	var age int
	err = db.QueryRow("SELECT age FROM users WHERE id = 2").Scan(&age)
	require.NoError(t, err)
	assert.Equal(t, 26, age)
}

func TestEndToEnd_UndoRedoCommit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewDataChangeManager()
	columns := []driver.ColumnInfo{
		{Name: "id", IsPrimaryKey: true, IsAutoIncrement: true},
		{Name: "name"},
		{Name: "email"},
		{Name: "age"},
	}
	manager.EnsureTab("tab1", "users", "", driver.DatabaseTypeSQLite, []string{"id"}, columns)

	err := manager.TrackCellChange("tab1", 0, "name", "Alice", "Alice V1", "text")
	require.NoError(t, err)

	err = manager.TrackCellChange("tab1", 1, "name", "Bob", "Bob V1", "text")
	require.NoError(t, err)

	err = manager.TrackCellChange("tab1", 2, "name", "Charlie", "Charlie V1", "text")
	require.NoError(t, err)

	assert.Equal(t, 3, manager.GetUndoStackSize())

	entry, err := manager.Undo()
	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, "Charlie", entry.OldValue)

	assert.Equal(t, 2, manager.GetChangeCount("tab1"))

	gen := NewSQLStatementGenerator(driver.DatabaseTypeSQLite)
	tabChanges := manager.GetChanges("tab1")

	rowPKs := map[int]map[string]any{
		0: {"id": 1},
		1: {"id": 2},
	}

	statements, err := gen.GenerateAll(tabChanges, rowPKs)
	require.NoError(t, err)
	assert.Len(t, statements, 2)

	ctx := context.Background()
	result, err := CommitChanges(ctx, db, statements)
	require.NoError(t, err)
	assert.Equal(t, 2, result.StatementsExecuted)

	var name string
	err = db.QueryRow("SELECT name FROM users WHERE id = 3").Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "Charlie", name)
}

func TestEndToEnd_ForeignKeyViolation(t *testing.T) {
	db := setupTestDBWithFK(t)
	defer db.Close()

	deleteStmt := Statement{
		SQL:    `DELETE FROM "departments" WHERE "id" = ?`,
		Params: []any{1},
		Action: ChangeActionDelete,
	}

	ctx := context.Background()
	_, err := CommitChanges(ctx, db, []Statement{deleteStmt})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rolled back")

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM departments").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestEndToEnd_RollbackOnError(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	statements := []Statement{
		{
			SQL:    `UPDATE "users" SET "name" = ? WHERE "id" = ?`,
			Params: []any{"Updated Name", 1},
			Action: ChangeActionUpdate,
		},
		{
			SQL:    `UPDATE "nonexistent_table" SET "name" = ? WHERE "id" = ?`,
			Params: []any{"Fail", 1},
			Action: ChangeActionUpdate,
		},
	}

	ctx := context.Background()
	_, err := CommitChanges(ctx, db, statements)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rolled back")

	var name string
	err = db.QueryRow("SELECT name FROM users WHERE id = 1").Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "Alice", name)
}

func TestEndToEnd_EmptyCommit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	result, err := CommitChanges(ctx, db, []Statement{})
	require.NoError(t, err)
	assert.Equal(t, 0, result.StatementsExecuted)
	assert.Equal(t, int64(0), result.RowsAffected)
}
