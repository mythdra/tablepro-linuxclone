package postgres

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestPostgreSQLDriver_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	host := getEnvOrDefault("POSTGRES_HOST", "localhost")
	port, _ := strconv.Atoi(getEnvOrDefault("POSTGRES_PORT", "5432"))
	user := getEnvOrDefault("POSTGRES_USER", "postgres")
	password := getEnvOrDefault("POSTGRES_PASSWORD", "postgres")
	dbname := getEnvOrDefault("POSTGRES_DB", "postgres")

	driver := NewPostgreSQLDriver()
	config := &Config{
		Host:     host,
		Port:     port,
		Database: dbname,
		Username: user,
		Password: password,
		SSLMode:  "disable",
	}

	ctx := context.Background()

	err := driver.Connect(ctx, config)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer driver.Close()

	t.Run("Execute simple query", func(t *testing.T) {
		result, err := driver.Execute(ctx, "SELECT 1 AS num, 'hello' AS str")
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if len(result.Rows) != 1 {
			t.Fatalf("Expected 1 row, got %d", len(result.Rows))
		}
		if result.Rows[0]["num"] != int64(1) {
			t.Errorf("Expected num=1, got %v", result.Rows[0]["num"])
		}
		if result.Rows[0]["str"] != "hello" {
			t.Errorf("Expected str=hello, got %v", result.Rows[0]["str"])
		}
	})

	t.Run("Execute with multiple rows", func(t *testing.T) {
		driver.Execute(ctx, "CREATE TABLE IF NOT EXISTS test_table (id SERIAL PRIMARY KEY, name TEXT)")
		driver.Execute(ctx, "INSERT INTO test_table (name) VALUES ('a'), ('b'), ('c')")

		result, err := driver.Execute(ctx, "SELECT id, name FROM test_table ORDER BY id")
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if len(result.Rows) != 3 {
			t.Fatalf("Expected 3 rows, got %d", len(result.Rows))
		}

		driver.Execute(ctx, "DROP TABLE test_table")
	})

	t.Run("GetSchema", func(t *testing.T) {
		schemas, err := driver.GetSchema()
		if err != nil {
			t.Fatalf("GetSchema failed: %v", err)
		}
		if len(schemas) == 0 {
			t.Error("Expected at least one schema")
		}
	})

	t.Run("Ping", func(t *testing.T) {
		err := driver.Ping(ctx)
		if err != nil {
			t.Fatalf("Ping failed: %v", err)
		}
	})
}

func TestPostgreSQLDriver_Integration_Schema(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	driver := getConnectedDriver(t)
	defer driver.Close()
	ctx := context.Background()

	driver.Execute(ctx, "CREATE SCHEMA IF NOT EXISTS test_schema")
	driver.Execute(ctx, "CREATE TABLE IF NOT EXISTS test_schema.users (id SERIAL PRIMARY KEY, name TEXT, email VARCHAR(255))")
	driver.Execute(ctx, "CREATE TABLE IF NOT EXISTS test_schema.orders (id SERIAL PRIMARY KEY, user_id INTEGER REFERENCES test_schema.users(id), amount DECIMAL(10,2))")

	t.Run("GetTables", func(t *testing.T) {
		tables, err := driver.GetTables("test_schema")
		if err != nil {
			t.Fatalf("GetTables failed: %v", err)
		}
		if len(tables) != 2 {
			t.Errorf("Expected 2 tables, got %d", len(tables))
		}
	})

	t.Run("GetColumns", func(t *testing.T) {
		columns, err := driver.GetColumns("test_schema.users")
		if err != nil {
			t.Fatalf("GetColumns failed: %v", err)
		}
		if len(columns) != 3 {
			t.Errorf("Expected 3 columns, got %d", len(columns))
		}
		for _, col := range columns {
			if col.Name == "id" && !col.IsPrimaryKey {
				t.Error("Expected id to be primary key")
			}
		}
	})

	t.Run("GetForeignKeys", func(t *testing.T) {
		fks, err := driver.GetForeignKeys("test_schema.orders")
		if err != nil {
			t.Fatalf("GetForeignKeys failed: %v", err)
		}
		if len(fks) != 1 {
			t.Errorf("Expected 1 foreign key, got %d", len(fks))
		}
	})

	driver.Execute(ctx, "DROP TABLE test_schema.orders")
	driver.Execute(ctx, "DROP TABLE test_schema.users")
	driver.Execute(ctx, "DROP SCHEMA test_schema")
}

func TestPostgreSQLDriver_Integration_Types(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	driver := getConnectedDriver(t)
	defer driver.Close()
	ctx := context.Background()

	driver.Execute(ctx, `
		CREATE TABLE IF NOT EXISTS type_tests (
			id SERIAL PRIMARY KEY,
			val_uuid UUID,
			val_json JSONB,
			val_array INTEGER[],
			val_ts TIMESTAMPTZ,
			val_bool BOOLEAN,
			val_int INTEGER,
			val_text TEXT
		)
	`)

	result, err := driver.Execute(ctx, "SELECT * FROM type_tests LIMIT 1")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(result.Columns) != 7 {
		t.Errorf("Expected 7 columns, got %d", len(result.Columns))
	}

	driver.Execute(ctx, "DROP TABLE type_tests")
}

func TestPostgreSQLDriver_ConnectionPool(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}

	driver := NewPostgreSQLDriver()
	config := &Config{
		Host:            getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:            5432,
		Database:        getEnvOrDefault("POSTGRES_DB", "postgres"),
		Username:        getEnvOrDefault("POSTGRES_USER", "postgres"),
		Password:        getEnvOrDefault("POSTGRES_PASSWORD", "postgres"),
		SSLMode:         "disable",
		MinConnections:  2,
		MaxConnections:  5,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 15 * time.Minute,
	}

	ctx := context.Background()
	err := driver.Connect(ctx, config)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer driver.Close()

	for i := 0; i < 3; i++ {
		_, err := driver.Execute(ctx, "SELECT 1")
		if err != nil {
			t.Fatalf("Execute %d failed: %v", i, err)
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getConnectedDriver(t *testing.T) *PostgreSQLDriver {
	driver := NewPostgreSQLDriver()
	config := &Config{
		Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:     5432,
		Database: getEnvOrDefault("POSTGRES_DB", "postgres"),
		Username: getEnvOrDefault("POSTGRES_USER", "postgres"),
		Password: getEnvOrDefault("POSTGRES_PASSWORD", "postgres"),
		SSLMode:  "disable",
	}

	ctx := context.Background()
	err := driver.Connect(ctx, config)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	return driver
}
