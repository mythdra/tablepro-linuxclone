package duckdb

import (
	"context"
	"testing"

	"tablepro/internal/connection"
)

func TestDuckDBDriver_Connect_InMemory(t *testing.T) {
	driver := New()
	cfg := connection.DatabaseConnection{
		Database: ":memory:",
	}

	err := driver.Connect(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Close()

	if err := driver.Ping(context.Background()); err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
}

func TestDuckDBDriver_Execute_Select(t *testing.T) {
	driver := New()
	cfg := connection.DatabaseConnection{
		Database: ":memory:",
	}

	err := driver.Connect(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Close()

	_, err = driver.Execute(context.Background(), "SELECT 1 AS num, 'hello' AS str")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestDuckDBDriver_Execute_Insert(t *testing.T) {
	driver := New()
	cfg := connection.DatabaseConnection{
		Database: ":memory:",
	}

	err := driver.Connect(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Close()

	_, err = driver.Execute(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, name VARCHAR)")
	if err != nil {
		t.Fatalf("Create table failed: %v", err)
	}

	result, err := driver.Execute(context.Background(), "INSERT INTO test (id, name) VALUES (1, 'test')")
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected result from INSERT")
	}

	if result[0]["rows_affected"].(int64) != 1 {
		t.Errorf("Expected 1 row affected, got %v", result[0]["rows_affected"])
	}
}

func TestDuckDBDriver_GetSchema(t *testing.T) {
	driver := New()
	cfg := connection.DatabaseConnection{
		Database: ":memory:",
	}

	err := driver.Connect(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Close()

	_, err = driver.Execute(context.Background(), "CREATE TABLE users (id INTEGER, name VARCHAR)")
	if err != nil {
		t.Fatalf("Create table failed: %v", err)
	}

	schema, err := driver.GetSchema(context.Background())
	if err != nil {
		t.Fatalf("GetSchema failed: %v", err)
	}

	if schema["name"] != "main" {
		t.Errorf("Expected schema name 'main', got %v", schema["name"])
	}

	tables := schema["tables"].([]map[string]any)
	if len(tables) != 1 {
		t.Errorf("Expected 1 table, got %d", len(tables))
	}

	if tables[0]["name"] != "users" {
		t.Errorf("Expected table name 'users', got %v", tables[0]["name"])
	}
}

func TestDuckDBDriver_GetTables(t *testing.T) {
	driver := New()
	cfg := connection.DatabaseConnection{
		Database: ":memory:",
	}

	err := driver.Connect(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Close()

	_, err = driver.Execute(context.Background(), "CREATE TABLE test1 (id INTEGER)")
	if err != nil {
		t.Fatalf("Create table failed: %v", err)
	}

	_, err = driver.Execute(context.Background(), "CREATE TABLE test2 (id INTEGER)")
	if err != nil {
		t.Fatalf("Create table failed: %v", err)
	}

	tables, err := driver.GetTables(context.Background())
	if err != nil {
		t.Fatalf("GetTables failed: %v", err)
	}

	if len(tables) != 2 {
		t.Errorf("Expected 2 tables, got %d", len(tables))
	}
}

func TestDuckDBDriver_GetColumns(t *testing.T) {
	driver := New()
	cfg := connection.DatabaseConnection{
		Database: ":memory:",
	}

	err := driver.Connect(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer driver.Close()

	_, err = driver.Execute(context.Background(), "CREATE TABLE users (id INTEGER, name VARCHAR NOT NULL, age INTEGER DEFAULT 18)")
	if err != nil {
		t.Fatalf("Create table failed: %v", err)
	}

	columns, err := driver.GetColumns(context.Background(), "users")
	if err != nil {
		t.Fatalf("GetColumns failed: %v", err)
	}

	if len(columns) != 3 {
		t.Errorf("Expected 3 columns, got %d", len(columns))
	}

	if columns[0]["name"] != "id" {
		t.Errorf("Expected first column 'id', got %v", columns[0]["name"])
	}

	if columns[1]["name"] != "name" {
		t.Errorf("Expected second column 'name', got %v", columns[1]["name"])
	}
}

func TestDuckDBDriver_Close(t *testing.T) {
	driver := New()
	cfg := connection.DatabaseConnection{
		Database: ":memory:",
	}

	err := driver.Connect(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if err := driver.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if err := driver.Ping(context.Background()); err == nil {
		t.Error("Expected error after close")
	}
}
