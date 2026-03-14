//go:build integration

// Package query provides integration tests for query execution with real databases.
// These tests require Docker to be running for testcontainers.
//
// Run with: go test -tags=integration ./internal/query/...
//
// Environment variables:
//   - INTEGRATION_TEST=1 (required to run integration tests)
//   - POSTGRES_HOST (default: localhost)
//   - POSTGRES_PORT (default: 5432)
//   - POSTGRES_USER (default: postgres)
//   - POSTGRES_PASSWORD (default: postgres)
//   - POSTGRES_DB (default: postgres)
//   - MYSQL_HOST (default: localhost)
//   - MYSQL_PORT (default: 3306)
//   - MYSQL_USER (default: root)
//   - MYSQL_PASSWORD (default: root)
//   - MYSQL_DB (default: mysql)
package query

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tablepro/internal/driver"
	"tablepro/internal/driver/mysql"
	"tablepro/internal/driver/postgres"
)

// ============================================================================
// Test Utilities
// ============================================================================

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getPostgresConfig() *postgres.Config {
	host := getEnvOrDefault("POSTGRES_HOST", "localhost")
	port, _ := strconv.Atoi(getEnvOrDefault("POSTGRES_PORT", "5432"))
	user := getEnvOrDefault("POSTGRES_USER", "postgres")
	password := getEnvOrDefault("POSTGRES_PASSWORD", "postgres")
	dbname := getEnvOrDefault("POSTGRES_DB", "postgres")

	return &postgres.Config{
		Host:     host,
		Port:     port,
		Database: dbname,
		Username: user,
		Password: password,
		SSLMode:  "disable",
	}
}

func getMySQLConfig() *mysql.Config {
	host := getEnvOrDefault("MYSQL_HOST", "localhost")
	port, _ := strconv.Atoi(getEnvOrDefault("MYSQL_PORT", "3306"))
	user := getEnvOrDefault("MYSQL_USER", "root")
	password := getEnvOrDefault("MYSQL_PASSWORD", "root")
	dbname := getEnvOrDefault("MYSQL_DB", "mysql")

	return &mysql.Config{
		Host:     host,
		Port:     port,
		Database: dbname,
		Username: user,
		Password: password,
		Charset:  "utf8mb4",
	}
}

func skipIfNotIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=1 to run.")
	}
}

// ============================================================================
// 11.1 Test Query Execution with PostgreSQL
// ============================================================================

func TestIntegration_QueryExecution_PostgreSQL(t *testing.T) {
	skipIfNotIntegration(t)
	t.Run("Simple SELECT query", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		// Execute simple SELECT query
		result, err := pgDriver.Execute(ctx, "SELECT 1 AS num, 'hello' AS str, true AS flag")
		require.NoError(t, err)
		require.Len(t, result.Rows, 1)

		// Verify result structure
		assert.Equal(t, int64(1), result.Rows[0]["num"])
		assert.Equal(t, "hello", result.Rows[0]["str"])
		assert.Equal(t, true, result.Rows[0]["flag"])
	})

	t.Run("ResultSet structure verification", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		// Create test table
		_, err = pgDriver.Execute(ctx, `
			CREATE TEMP TABLE test_resultset (
				id SERIAL PRIMARY KEY,
				name VARCHAR(100),
				value DECIMAL(10,2),
				created_at TIMESTAMP DEFAULT NOW()
			)
		`)
		require.NoError(t, err)

		// Insert test data
		_, err = pgDriver.Execute(ctx, `
			INSERT INTO test_resultset (name, value) VALUES 
			('test1', 10.50),
			('test2', 20.75)
		`)
		require.NoError(t, err)

		// Query and verify ResultSet structure
		result, err := pgDriver.Execute(ctx, "SELECT id, name, value, created_at FROM test_resultset ORDER BY id")
		require.NoError(t, err)

		// Verify columns
		assert.Len(t, result.Columns, 4)
		assert.Equal(t, "id", result.Columns[0].Name)
		assert.Equal(t, "name", result.Columns[1].Name)
		assert.Equal(t, "value", result.Columns[2].Name)
		assert.Equal(t, "created_at", result.Columns[3].Name)

		// Verify data types
		assert.Equal(t, "int4", result.Columns[0].DatabaseType)
		assert.Equal(t, "varchar", result.Columns[1].DatabaseType)
		assert.Contains(t, result.Columns[2].DatabaseType, "numeric")
		assert.Contains(t, result.Columns[3].DatabaseType, "timestamp")

		// Verify row count
		assert.Equal(t, int64(2), result.RowCount)

		// Verify data
		assert.Equal(t, int64(1), result.Rows[0]["id"])
		assert.Equal(t, "test1", result.Rows[0]["name"])
		assert.Equal(t, "10.50", result.Rows[0]["value"]) // DECIMAL as string
	})

	t.Run("Column type mapping", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		// Test various PostgreSQL types
		_, err = pgDriver.Execute(ctx, `
			CREATE TEMP TABLE test_types (
				col_bool BOOLEAN,
				col_int2 SMALLINT,
				col_int4 INTEGER,
				col_int8 BIGINT,
				col_float4 REAL,
				col_float8 DOUBLE PRECISION,
				col_numeric NUMERIC(10,2),
				col_text TEXT,
				col_varchar VARCHAR(50),
				col_date DATE,
				col_timestamp TIMESTAMP,
				col_timestamptz TIMESTAMPTZ,
				col_json JSONB,
				col_uuid UUID
			)
		`)
		require.NoError(t, err)

		// Insert test data
		_, err = pgDriver.Execute(ctx, `
			INSERT INTO test_types VALUES (
				true,
				32767,
				2147483647,
				9223372036854775807,
				3.14159,
				2.718281828459045,
				999.99,
				'text value',
				'varchar value',
				'2024-01-15',
				'2024-01-15 10:30:00',
				'2024-01-15 10:30:00+00',
				'{"key": "value"}',
				'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'
			)
		`)
		require.NoError(t, err)

		result, err := pgDriver.Execute(ctx, "SELECT * FROM test_types")
		require.NoError(t, err)

		// Verify column types are detected
		require.Len(t, result.Columns, 14)
		assert.Equal(t, "bool", result.Columns[0].DatabaseType)
		assert.Equal(t, "int2", result.Columns[1].DatabaseType)
		assert.Equal(t, "int4", result.Columns[2].DatabaseType)
		assert.Equal(t, "int8", result.Columns[3].DatabaseType)
		assert.Equal(t, "float4", result.Columns[4].DatabaseType)
		assert.Equal(t, "float8", result.Columns[5].DatabaseType)
		assert.Contains(t, result.Columns[6].DatabaseType, "numeric")
		assert.Equal(t, "text", result.Columns[7].DatabaseType)
		assert.Equal(t, "varchar", result.Columns[8].DatabaseType)
		assert.Equal(t, "date", result.Columns[9].DatabaseType)
		assert.Contains(t, result.Columns[10].DatabaseType, "timestamp")
		assert.Contains(t, result.Columns[11].DatabaseType, "timestamptz")
		assert.Equal(t, "jsonb", result.Columns[12].DatabaseType)
		assert.Equal(t, "uuid", result.Columns[13].DatabaseType)
	})

	t.Run("Data formatting", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		_, err = pgDriver.Execute(ctx, `
			CREATE TEMP TABLE test_format (
				col_bool BOOLEAN,
				col_float FLOAT8,
				col_numeric NUMERIC(10,2),
				col_json JSONB,
				col_uuid UUID,
				col_timestamp TIMESTAMP
			)
		`)
		require.NoError(t, err)

		_, err = pgDriver.Execute(ctx, `
			INSERT INTO test_format VALUES (
				true,
				3.14159,
				999.99,
				'{"nested": {"key": "value"}}',
				'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
				'2024-01-15 10:30:00.123456'
			)
		`)
		require.NoError(t, err)

		result, err := pgDriver.Execute(ctx, "SELECT * FROM test_format")
		require.NoError(t, err)

		// Verify formatted values
		row := result.Rows[0]
		assert.Equal(t, true, row["col_bool"])
		assert.Equal(t, float64(3.14159), row["col_float"])
		assert.Equal(t, "999.99", row["col_numeric"])
		assert.Contains(t, row["col_json"], "nested")
		assert.Equal(t, "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", row["col_uuid"])
		// Timestamp should be formatted as RFC3339
		assert.Contains(t, row["col_timestamp"], "2024-01-15")
	})
}

// ============================================================================
// 11.2 Test Query Execution with MySQL
// ============================================================================

func TestIntegration_QueryExecution_MySQL(t *testing.T) {
	skipIfNotIntegration(t)
	t.Run("Simple SELECT query", func(t *testing.T) {
		mysqlDriver := mysql.NewMySQLDriver()
		config := getMySQLConfig()
		ctx := context.Background()

		err := mysqlDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer mysqlDriver.Close()

		// Execute simple SELECT query
		result, err := mysqlDriver.Execute(ctx, "SELECT 1 AS num, 'hello' AS str, true AS flag")
		require.NoError(t, err)
		require.Len(t, result.Rows, 1)

		// Verify result structure
		assert.Equal(t, int64(1), result.Rows[0]["num"])
		assert.Equal(t, "hello", result.Rows[0]["str"])
		// MySQL returns bool as int8
		assert.Equal(t, int8(1), result.Rows[0]["flag"])
	})

	t.Run("ResultSet structure verification", func(t *testing.T) {
		mysqlDriver := mysql.NewMySQLDriver()
		config := getMySQLConfig()
		ctx := context.Background()

		err := mysqlDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer mysqlDriver.Close()

		// Create test table
		_, err = mysqlDriver.Execute(ctx, `
			CREATE TEMPORARY TABLE test_resultset (
				id INT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(100),
				value DECIMAL(10,2),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB
		`)
		require.NoError(t, err)

		// Insert test data
		_, err = mysqlDriver.Execute(ctx, `
			INSERT INTO test_resultset (name, value) VALUES 
			('test1', 10.50),
			('test2', 20.75)
		`)
		require.NoError(t, err)

		// Query and verify ResultSet structure
		result, err := mysqlDriver.Execute(ctx, "SELECT id, name, value, created_at FROM test_resultset ORDER BY id")
		require.NoError(t, err)

		// Verify columns
		assert.Len(t, result.Columns, 4)
		assert.Equal(t, "id", result.Columns[0].Name)
		assert.Equal(t, "name", result.Columns[1].Name)
		assert.Equal(t, "value", result.Columns[2].Name)
		assert.Equal(t, "created_at", result.Columns[3].Name)

		// Verify data types
		assert.Contains(t, result.Columns[0].DatabaseType, "int")
		assert.Equal(t, "varchar", result.Columns[1].DatabaseType)
		assert.Equal(t, "decimal", result.Columns[2].DatabaseType)
		assert.Equal(t, "timestamp", result.Columns[3].DatabaseType)

		// Verify row count
		assert.Equal(t, int64(2), result.RowCount)

		// Verify data
		assert.Equal(t, int64(1), result.Rows[0]["id"])
		assert.Equal(t, "test1", result.Rows[0]["name"])
		assert.Equal(t, "10.50", result.Rows[0]["value"])
	})

	t.Run("Column type mapping", func(t *testing.T) {
		mysqlDriver := mysql.NewMySQLDriver()
		config := getMySQLConfig()
		ctx := context.Background()

		err := mysqlDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer mysqlDriver.Close()

		// Test various MySQL types
		_, err = mysqlDriver.Execute(ctx, `
			CREATE TEMPORARY TABLE test_types (
				col_bool BOOL,
				col_tinyint TINYINT,
				col_smallint SMALLINT,
				col_mediumint MEDIUMINT,
				col_int INT,
				col_bigint BIGINT,
				col_float FLOAT,
				col_double DOUBLE,
				col_decimal DECIMAL(10,2),
				col_text TEXT,
				col_varchar VARCHAR(50),
				col_date DATE,
				col_datetime DATETIME,
				col_timestamp TIMESTAMP,
				col_json JSON,
				col_enum ENUM('a', 'b', 'c')
			) ENGINE=InnoDB
		`)
		require.NoError(t, err)

		// Insert test data
		_, err = mysqlDriver.Execute(ctx, `
			INSERT INTO test_types VALUES (
				1,
				127,
				32767,
				8388607,
				2147483647,
				9223372036854775807,
				3.14159,
				2.718281828459045,
				999.99,
				'text value',
				'varchar value',
				'2024-01-15',
				'2024-01-15 10:30:00',
				CURRENT_TIMESTAMP,
				'{"key": "value"}',
				'b'
			)
		`)
		require.NoError(t, err)

		result, err := mysqlDriver.Execute(ctx, "SELECT * FROM test_types")
		require.NoError(t, err)

		// Verify column types are detected
		require.Len(t, result.Columns, 16)
		assert.Equal(t, "tinyint", result.Columns[0].DatabaseType) // BOOL is tinyint(1)
		assert.Equal(t, "tinyint", result.Columns[1].DatabaseType)
		assert.Equal(t, "smallint", result.Columns[2].DatabaseType)
		assert.Equal(t, "mediumint", result.Columns[3].DatabaseType)
		assert.Equal(t, "int", result.Columns[4].DatabaseType)
		assert.Equal(t, "bigint", result.Columns[5].DatabaseType)
		assert.Equal(t, "float", result.Columns[6].DatabaseType)
		assert.Equal(t, "double", result.Columns[7].DatabaseType)
		assert.Equal(t, "decimal", result.Columns[8].DatabaseType)
		assert.Equal(t, "text", result.Columns[9].DatabaseType)
		assert.Equal(t, "varchar", result.Columns[10].DatabaseType)
		assert.Equal(t, "date", result.Columns[11].DatabaseType)
		assert.Equal(t, "datetime", result.Columns[12].DatabaseType)
		assert.Equal(t, "timestamp", result.Columns[13].DatabaseType)
		assert.Equal(t, "json", result.Columns[14].DatabaseType)
		assert.Equal(t, "enum", result.Columns[15].DatabaseType)
	})

	t.Run("Data formatting", func(t *testing.T) {
		mysqlDriver := mysql.NewMySQLDriver()
		config := getMySQLConfig()
		ctx := context.Background()

		err := mysqlDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer mysqlDriver.Close()

		_, err = mysqlDriver.Execute(ctx, `
			CREATE TEMPORARY TABLE test_format (
				col_bool BOOL,
				col_float DOUBLE,
				col_decimal DECIMAL(10,2),
				col_json JSON,
				col_datetime DATETIME,
				col_enum ENUM('small', 'medium', 'large')
			) ENGINE=InnoDB
		`)
		require.NoError(t, err)

		_, err = mysqlDriver.Execute(ctx, `
			INSERT INTO test_format VALUES (
				1,
				3.14159,
				999.99,
				'{"nested": {"key": "value"}}',
				'2024-01-15 10:30:00',
				'medium'
			)
		`)
		require.NoError(t, err)

		result, err := mysqlDriver.Execute(ctx, "SELECT * FROM test_format")
		require.NoError(t, err)

		// Verify formatted values
		row := result.Rows[0]
		assert.Equal(t, int8(1), row["col_bool"])
		assert.Equal(t, float64(3.14159), row["col_float"])
		assert.Equal(t, "999.99", row["col_decimal"])
		assert.Contains(t, row["col_json"], "nested")
		assert.Contains(t, row["col_datetime"], "2024-01-15")
		assert.Equal(t, "medium", row["col_enum"])
	})
}

// ============================================================================
// 11.3 Test Query Cancellation
// ============================================================================

func TestIntegration_QueryCancellation_PostgreSQL(t *testing.T) {
	skipIfNotIntegration(t)
	t.Run("Cancel long-running query with pg_sleep", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		executor := NewQueryExecutor()
		connID := uuid.New()

		// Start a query that sleeps for 10 seconds
		queryStarted := make(chan struct{})
		queryCancelled := make(chan error, 1)

		go func() {
			close(queryStarted)
			_, err := executor.Execute(ctx, connID, pgDriver, "SELECT pg_sleep(10)")
			queryCancelled <- err
		}()

		// Wait for query to start
		<-queryStarted
		time.Sleep(100 * time.Millisecond)

		// Get active query and cancel it
		activeQuery := executor.GetActiveQuery(connID)
		require.NotNil(t, activeQuery, "Query should be active")

		// Cancel the query
		err = executor.Cancel(activeQuery.ID)
		assert.NoError(t, err)

		// Wait for cancellation to propagate
		select {
		case err := <-queryCancelled:
			// Query should be cancelled
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "canceling statement due to user request")
		case <-time.After(5 * time.Second):
			t.Fatal("Query cancellation timed out")
		}
	})

	t.Run("Verify resources cleaned up after cancellation", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		executor := NewQueryExecutor()
		connID := uuid.New()

		// Execute and cancel multiple queries
		for i := 0; i < 3; i++ {
			queryStarted := make(chan struct{})

			go func() {
				close(queryStarted)
				_, _ = executor.Execute(ctx, connID, pgDriver, "SELECT pg_sleep(5)")
			}()

			<-queryStarted
			time.Sleep(50 * time.Millisecond)

			activeQuery := executor.GetActiveQuery(connID)
			require.NotNil(t, activeQuery)

			err = executor.Cancel(activeQuery.ID)
			assert.NoError(t, err)

			time.Sleep(100 * time.Millisecond)
		}

		// Verify no active queries remain
		activeQuery := executor.GetActiveQuery(connID)
		assert.Nil(t, activeQuery, "No active queries should remain")

		// Verify session still exists (for history)
		assert.Equal(t, 1, executor.GetSessionCount())
	})

	t.Run("Cancel via CancelConnection", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		executor := NewQueryExecutor()
		connID := uuid.New()

		queryStarted := make(chan struct{})
		queryCancelled := make(chan error, 1)

		go func() {
			close(queryStarted)
			_, err := executor.Execute(ctx, connID, pgDriver, "SELECT pg_sleep(10)")
			queryCancelled <- err
		}()

		<-queryStarted
		time.Sleep(100 * time.Millisecond)

		// Cancel entire connection
		err = executor.CancelConnection(connID)
		assert.NoError(t, err)

		select {
		case err := <-queryCancelled:
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "canceling statement")
		case <-time.After(5 * time.Second):
			t.Fatal("Connection cancellation timed out")
		}
	})
}

// ============================================================================
// 11.4 Test Pagination with Large Result Sets
// ============================================================================

func TestIntegration_Pagination_LargeResultSets(t *testing.T) {
	skipIfNotIntegration(t)
	t.Run("PostgreSQL - Generate and paginate 10k rows", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		// Create test table
		_, err = pgDriver.Execute(ctx, `
			CREATE TEMP TABLE test_pagination (
				id SERIAL PRIMARY KEY,
				value INTEGER,
				description TEXT
			)
		`)
		require.NoError(t, err)

		// Generate 10k rows using generate_series
		_, err = pgDriver.Execute(ctx, `
			INSERT INTO test_pagination (value, description)
			SELECT 
				g, 
				'Description for row ' || g
			FROM generate_series(1, 10000) AS g
		`)
		require.NoError(t, err)

		// Test pagination
		paginationService := NewPaginationService()
		pageSize := 100

		// Page 1
		query := paginationService.ApplySQLOffset("SELECT id, value, description FROM test_pagination ORDER BY id", 1, pageSize)
		startTime := time.Now()
		result, err := pgDriver.Execute(ctx, query)
		duration := time.Since(startTime)
		require.NoError(t, err)
		assert.Equal(t, int64(pageSize), result.RowCount)
		assert.Less(t, duration, 1*time.Second, "Page 1 should load in under 1 second")
		assert.Equal(t, int64(1), result.Rows[0]["id"])
		assert.Equal(t, int64(100), result.Rows[99]["id"])

		// Page 50
		query = paginationService.ApplySQLOffset("SELECT id, value, description FROM test_pagination ORDER BY id", 50, pageSize)
		startTime = time.Now()
		result, err = pgDriver.Execute(ctx, query)
		duration = time.Since(startTime)
		require.NoError(t, err)
		assert.Equal(t, int64(pageSize), result.RowCount)
		assert.Less(t, duration, 1*time.Second, "Page 50 should load in under 1 second")
		assert.Equal(t, int64(4901), result.Rows[0]["id"])
		assert.Equal(t, int64(5000), result.Rows[99]["id"])

		// Page 100 (last page)
		query = paginationService.ApplySQLOffset("SELECT id, value, description FROM test_pagination ORDER BY id", 100, pageSize)
		result, err = pgDriver.Execute(ctx, query)
		require.NoError(t, err)
		assert.Equal(t, int64(pageSize), result.RowCount)
		assert.Equal(t, int64(9901), result.Rows[0]["id"])
		assert.Equal(t, int64(10000), result.Rows[99]["id"])

		// Verify total count
		countResult, err := pgDriver.Execute(ctx, "SELECT COUNT(*) FROM test_pagination")
		require.NoError(t, err)
		assert.Equal(t, int64(10000), countResult.Rows[0]["count"])
	})

	t.Run("MySQL - Generate and paginate 10k rows", func(t *testing.T) {
		mysqlDriver := mysql.NewMySQLDriver()
		config := getMySQLConfig()
		ctx := context.Background()

		err := mysqlDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer mysqlDriver.Close()

		// Create test table
		_, err = mysqlDriver.Execute(ctx, `
			CREATE TEMPORARY TABLE test_pagination (
				id INT AUTO_INCREMENT PRIMARY KEY,
				value INT,
				description VARCHAR(255)
			) ENGINE=InnoDB
		`)
		require.NoError(t, err)

		// Generate 10k rows using a recursive CTE
		_, err = mysqlDriver.Execute(ctx, `
			WITH RECURSIVE numbers AS (
				SELECT 1 AS n
				UNION ALL
				SELECT n + 1 FROM numbers WHERE n < 10000
			)
			INSERT INTO test_pagination (value, description)
			SELECT n, CONCAT('Description for row ', n) FROM numbers
		`)
		require.NoError(t, err)

		// Test pagination
		paginationService := NewPaginationService()
		pageSize := 100

		// Page 1
		query := paginationService.ApplySQLOffset("SELECT id, value, description FROM test_pagination ORDER BY id", 1, pageSize)
		startTime := time.Now()
		result, err := mysqlDriver.Execute(ctx, query)
		duration := time.Since(startTime)
		require.NoError(t, err)
		assert.Equal(t, int64(pageSize), result.RowCount)
		assert.Less(t, duration, 1*time.Second, "Page 1 should load in under 1 second")
		assert.Equal(t, int64(1), result.Rows[0]["id"])
		assert.Equal(t, int64(100), result.Rows[99]["id"])

		// Page 50
		query = paginationService.ApplySQLOffset("SELECT id, value, description FROM test_pagination ORDER BY id", 50, pageSize)
		startTime = time.Now()
		result, err = mysqlDriver.Execute(ctx, query)
		duration = time.Since(startTime)
		require.NoError(t, err)
		assert.Equal(t, int64(pageSize), result.RowCount)
		assert.Less(t, duration, 1*time.Second, "Page 50 should load in under 1 second")
		assert.Equal(t, int64(4901), result.Rows[0]["id"])
		assert.Equal(t, int64(5000), result.Rows[99]["id"])

		// Verify total count
		countResult, err := mysqlDriver.Execute(ctx, "SELECT COUNT(*) as count FROM test_pagination")
		require.NoError(t, err)
		assert.Equal(t, int64(10000), countResult.Rows[0]["count"])
	})

	t.Run("Pagination with WHERE clause", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		_, err = pgDriver.Execute(ctx, `
			CREATE TEMP TABLE test_filtered (
				id SERIAL PRIMARY KEY,
				category VARCHAR(50),
				value INTEGER
			)
		`)
		require.NoError(t, err)

		_, err = pgDriver.Execute(ctx, `
			INSERT INTO test_filtered (category, value)
			SELECT 
				CASE WHEN g % 3 = 0 THEN 'A' WHEN g % 3 = 1 THEN 'B' ELSE 'C' END,
				g
			FROM generate_series(1, 3000) AS g
		`)
		require.NoError(t, err)

		paginationService := NewPaginationService()
		pageSize := 50

		// Paginate filtered results
		baseQuery := "SELECT id, category, value FROM test_filtered WHERE category = 'A' ORDER BY id"
		query := paginationService.ApplySQLOffset(baseQuery, 1, pageSize)
		result, err := pgDriver.Execute(ctx, query)
		require.NoError(t, err)
		assert.Equal(t, int64(pageSize), result.RowCount)

		// All rows should be category A
		for _, row := range result.Rows {
			assert.Equal(t, "A", row["category"])
		}
	})
}

// ============================================================================
// 11.5 Test NULL Value Handling
// ============================================================================

func TestIntegration_NULLValueHandling_PostgreSQL(t *testing.T) {
	skipIfNotIntegration(t)
	t.Run("NULL values in various column types", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		// Create table with nullable columns
		_, err = pgDriver.Execute(ctx, `
			CREATE TEMP TABLE test_nulls (
				id SERIAL PRIMARY KEY,
				col_int INTEGER,
				col_float FLOAT8,
				col_text TEXT,
				col_bool BOOLEAN,
				col_date DATE,
				col_timestamp TIMESTAMP,
				col_json JSONB,
				col_uuid UUID
			)
		`)
		require.NoError(t, err)

		// Insert row with all NULLs
		_, err = pgDriver.Execute(ctx, `
			INSERT INTO test_nulls 
			(col_int, col_float, col_text, col_bool, col_date, col_timestamp, col_json, col_uuid)
			VALUES (NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL)
		`)
		require.NoError(t, err)

		// Insert row with mixed NULLs and values
		_, err = pgDriver.Execute(ctx, `
			INSERT INTO test_nulls 
			(col_int, col_float, col_text, col_bool, col_date, col_timestamp, col_json, col_uuid)
			VALUES (42, 3.14, 'not null', true, '2024-01-15', '2024-01-15 10:30:00', '{"key": "value"}', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11')
		`)
		require.NoError(t, err)

		// Query and verify NULL handling
		result, err := pgDriver.Execute(ctx, "SELECT * FROM test_nulls ORDER BY id")
		require.NoError(t, err)
		require.Equal(t, int64(2), result.RowCount)

		// First row: all NULLs
		nullRow := result.Rows[0]
		assert.Nil(t, nullRow["col_int"], "col_int should be nil")
		assert.Nil(t, nullRow["col_float"], "col_float should be nil")
		assert.Nil(t, nullRow["col_text"], "col_text should be nil")
		assert.Nil(t, nullRow["col_bool"], "col_bool should be nil")
		assert.Nil(t, nullRow["col_date"], "col_date should be nil")
		assert.Nil(t, nullRow["col_timestamp"], "col_timestamp should be nil")
		assert.Nil(t, nullRow["col_json"], "col_json should be nil")
		assert.Nil(t, nullRow["col_uuid"], "col_uuid should be nil")

		// Second row: all values
		valueRow := result.Rows[1]
		assert.Equal(t, int64(42), valueRow["col_int"])
		assert.Equal(t, float64(3.14), valueRow["col_float"])
		assert.Equal(t, "not null", valueRow["col_text"])
		assert.Equal(t, true, valueRow["col_bool"])
		assert.Contains(t, valueRow["col_date"], "2024-01-15")
		assert.Contains(t, valueRow["col_timestamp"], "2024-01-15")
		assert.Contains(t, valueRow["col_json"], "key")
		assert.Equal(t, "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", valueRow["col_uuid"])
	})

	t.Run("NULL to JSON null conversion", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		_, err = pgDriver.Execute(ctx, `
			CREATE TEMP TABLE test_json_null (
				id SERIAL PRIMARY KEY,
				data JSONB
			)
		`)
		require.NoError(t, err)

		_, err = pgDriver.Execute(ctx, `
			INSERT INTO test_json_null (data) VALUES (NULL)
		`)
		require.NoError(t, err)

		result, err := pgDriver.Execute(ctx, "SELECT data FROM test_json_null")
		require.NoError(t, err)
		require.Equal(t, int64(1), result.RowCount)

		// NULL should be nil, not empty string or 0
		assert.Nil(t, result.Rows[0]["data"])
	})

	t.Run("ResultSet JSON serialization preserves nulls", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		_, err = pgDriver.Execute(ctx, `
			CREATE TEMP TABLE test_json_serialize (
				id INTEGER,
				value TEXT
			)
		`)
		require.NoError(t, err)

		_, err = pgDriver.Execute(ctx, `
			INSERT INTO test_json_serialize VALUES (1, NULL), (2, 'not null'), (3, NULL)
		`)
		require.NoError(t, err)

		result, err := pgDriver.Execute(ctx, "SELECT * FROM test_json_serialize ORDER BY id")
		require.NoError(t, err)

		// Serialize to JSON
		jsonBytes, err := result.MarshalJSON()
		require.NoError(t, err)

		jsonStr := string(jsonBytes)
		// Verify JSON contains "null" for NULL values
		assert.Contains(t, jsonStr, `"rows":[[1,null],[2,"not null"],[3,null]]`)
	})
}

func TestIntegration_NULLValueHandling_MySQL(t *testing.T) {
	skipIfNotIntegration(t)
	t.Run("NULL values in various column types", func(t *testing.T) {
		mysqlDriver := mysql.NewMySQLDriver()
		config := getMySQLConfig()
		ctx := context.Background()

		err := mysqlDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer mysqlDriver.Close()

		// Create table with nullable columns
		_, err = mysqlDriver.Execute(ctx, `
			CREATE TEMPORARY TABLE test_nulls (
				id INT AUTO_INCREMENT PRIMARY KEY,
				col_int INT,
				col_float DOUBLE,
				col_text TEXT,
				col_bool BOOL,
				col_date DATE,
				col_datetime DATETIME,
				col_json JSON
			) ENGINE=InnoDB
		`)
		require.NoError(t, err)

		// Insert row with all NULLs
		_, err = mysqlDriver.Execute(ctx, `
			INSERT INTO test_nulls 
			(col_int, col_float, col_text, col_bool, col_date, col_datetime, col_json)
			VALUES (NULL, NULL, NULL, NULL, NULL, NULL, NULL)
		`)
		require.NoError(t, err)

		// Insert row with mixed NULLs and values
		_, err = mysqlDriver.Execute(ctx, `
			INSERT INTO test_nulls 
			(col_int, col_float, col_text, col_bool, col_date, col_datetime, col_json)
			VALUES (42, 3.14, 'not null', 1, '2024-01-15', '2024-01-15 10:30:00', '{"key": "value"}')
		`)
		require.NoError(t, err)

		// Query and verify NULL handling
		result, err := mysqlDriver.Execute(ctx, "SELECT * FROM test_nulls ORDER BY id")
		require.NoError(t, err)
		require.Equal(t, int64(2), result.RowCount)

		// First row: all NULLs
		nullRow := result.Rows[0]
		assert.Nil(t, nullRow["col_int"], "col_int should be nil")
		assert.Nil(t, nullRow["col_float"], "col_float should be nil")
		assert.Nil(t, nullRow["col_text"], "col_text should be nil")
		assert.Nil(t, nullRow["col_bool"], "col_bool should be nil")
		assert.Nil(t, nullRow["col_date"], "col_date should be nil")
		assert.Nil(t, nullRow["col_datetime"], "col_datetime should be nil")
		assert.Nil(t, nullRow["col_json"], "col_json should be nil")
	})
}

// ============================================================================
// 11.6 Test History Tracking and Deduplication
// ============================================================================

func TestIntegration_HistoryTracking_Deduplication(t *testing.T) {
	skipIfNotIntegration(t)
	t.Run("Deduplicate identical queries", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		executor := NewQueryExecutor()
		connID := uuid.New()

		// Execute same query twice
		_, err = executor.Execute(ctx, connID, pgDriver, "SELECT 1")
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
		_, err = executor.Execute(ctx, connID, pgDriver, "SELECT 1")
		require.NoError(t, err)

		// Verify only one entry in history
		history := executor.GetHistory(connID, 10)
		assert.Len(t, history, 1, "Identical queries should be deduplicated")
	})

	t.Run("Deduplicate queries with whitespace variations", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		executor := NewQueryExecutor()
		connID := uuid.New()

		// Execute same query with different whitespace
		_, err = executor.Execute(ctx, connID, pgDriver, "SELECT * FROM users")
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
		_, err = executor.Execute(ctx, connID, pgDriver, "  SELECT * FROM users  ")
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
		_, err = executor.Execute(ctx, connID, pgDriver, "\tSELECT * FROM users\n")
		require.NoError(t, err)

		// Verify only one entry in history
		history := executor.GetHistory(connID, 10)
		assert.Len(t, history, 1, "Queries with whitespace variations should be deduplicated")
	})

	t.Run("Different queries are not deduplicated", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		executor := NewQueryExecutor()
		connID := uuid.New()

		// Execute different queries
		_, err = executor.Execute(ctx, connID, pgDriver, "SELECT 1")
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
		_, err = executor.Execute(ctx, connID, pgDriver, "SELECT 2")
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
		_, err = executor.Execute(ctx, connID, pgDriver, "SELECT 3")
		require.NoError(t, err)

		// Verify all entries in history
		history := executor.GetHistory(connID, 10)
		assert.Len(t, history, 3, "Different queries should not be deduplicated")
	})
}

func TestIntegration_HistoryTracking_LRUEviction(t *testing.T) {
	skipIfNotIntegration(t)
	t.Run("LRU eviction at 50 queries", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		executor := NewQueryExecutor()
		connID := uuid.New()

		// Execute 55 queries
		for i := 0; i < 55; i++ {
			query := fmt.Sprintf("SELECT %d", i)
			_, err = executor.Execute(ctx, connID, pgDriver, query)
			require.NoError(t, err)
			time.Sleep(time.Millisecond) // Ensure unique timestamps
		}

		// Verify only 50 entries remain (oldest evicted)
		history := executor.GetHistory(connID, 0)
		assert.Len(t, history, 50, "Should have exactly 50 entries after eviction")

		// Verify oldest entries were evicted (should start from SELECT 5)
		assert.Contains(t, history[49].Query, "SELECT 5", "Oldest entry should be SELECT 5")
		assert.Contains(t, history[0].Query, "SELECT 54", "Newest entry should be SELECT 54")
	})

	t.Run("Per-connection isolation", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		executor := NewQueryExecutor()
		connID1 := uuid.New()
		connID2 := uuid.New()

		// Execute queries on connection 1
		for i := 0; i < 25; i++ {
			query := fmt.Sprintf("SELECT conn1_%d", i)
			_, _ = executor.Execute(ctx, connID1, pgDriver, query)
		}

		// Execute different queries on connection 2
		for i := 0; i < 30; i++ {
			query := fmt.Sprintf("SELECT conn2_%d", i)
			_, _ = executor.Execute(ctx, connID2, pgDriver, query)
		}

		// Verify isolation
		history1 := executor.GetHistory(connID1, 100)
		history2 := executor.GetHistory(connID2, 100)

		assert.Len(t, history1, 25, "Connection 1 should have 25 entries")
		assert.Len(t, history2, 30, "Connection 2 should have 30 entries")

		// Verify no cross-contamination
		for _, entry := range history1 {
			assert.Contains(t, entry.Query, "conn1", "Connection 1 should only have conn1 queries")
		}
		for _, entry := range history2 {
			assert.Contains(t, entry.Query, "conn2", "Connection 2 should only have conn2 queries")
		}
	})
}

func TestIntegration_HistoryTracking_ClearHistory(t *testing.T) {
	skipIfNotIntegration(t)
	t.Run("Clear history for specific connection", func(t *testing.T) {
		pgDriver := postgres.NewPostgreSQLDriver()
		config := getPostgresConfig()
		ctx := context.Background()

		err := pgDriver.Connect(ctx, config)
		require.NoError(t, err)
		defer pgDriver.Close()

		executor := NewQueryExecutor()
		connID1 := uuid.New()
		connID2 := uuid.New()

		// Add history to both connections
		_, _ = executor.Execute(ctx, connID1, pgDriver, "SELECT 1")
		_, _ = executor.Execute(ctx, connID2, pgDriver, "SELECT 2")

		// Clear history for connection 1
		executor.ClearHistory(connID1)

		// Verify
		history1 := executor.GetHistory(connID1, 10)
		history2 := executor.GetHistory(connID2, 10)

		assert.Empty(t, history1, "Connection 1 history should be cleared")
		assert.Len(t, history2, 1, "Connection 2 history should remain")
	})
}
