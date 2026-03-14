package query

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"tablepro/internal/driver"
)

// MockDriver is a mock implementation of driver.DatabaseDriver for testing.
type MockDriver struct {
	queryFunc func(ctx context.Context, query string, params ...any) (*driver.Row, error)
}

func (m *MockDriver) Connect(ctx context.Context, config *driver.ConnectionConfig) error {
	return nil
}

func (m *MockDriver) Execute(ctx context.Context, query string, params ...any) (*driver.Result, error) {
	return &driver.Result{}, nil
}

func (m *MockDriver) Query(ctx context.Context, query string, params ...any) (*driver.Row, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, query, params...)
	}
	return &driver.Row{
		Data:        map[string]any{"result": "mock"},
		ColumnNames: []string{"result"},
	}, nil
}

func (m *MockDriver) QueryContext(ctx context.Context, timeout time.Duration, query string, params ...any) (*driver.Row, error) {
	return m.Query(ctx, query, params...)
}

func (m *MockDriver) GetSchema(ctx context.Context) (*driver.SchemaInfo, error) {
	return &driver.SchemaInfo{}, nil
}

func (m *MockDriver) GetTables(ctx context.Context, schemaName string) ([]driver.TableInfo, error) {
	return []driver.TableInfo{}, nil
}

func (m *MockDriver) GetColumns(ctx context.Context, schemaName, tableName string) ([]driver.ColumnInfo, error) {
	return []driver.ColumnInfo{}, nil
}

func (m *MockDriver) GetIndexes(ctx context.Context, schemaName, tableName string) ([]driver.IndexInfo, error) {
	return []driver.IndexInfo{}, nil
}

func (m *MockDriver) GetForeignKeys(ctx context.Context, schemaName, tableName string) ([]driver.ForeignKeyInfo, error) {
	return []driver.ForeignKeyInfo{}, nil
}

func (m *MockDriver) Ping(ctx context.Context) error {
	return nil
}

func (m *MockDriver) Close() error {
	return nil
}

func (m *MockDriver) GetCapabilities() *driver.DriverCapabilities {
	return &driver.DriverCapabilities{}
}

func (m *MockDriver) GetDB() *sql.DB {
	return nil
}

func (m *MockDriver) Type() driver.DatabaseType {
	return driver.DatabaseTypeUnknown
}

// TestQueryExecutor_Execute tests the Execute method with timeout.
func TestQueryExecutor_Execute(t *testing.T) {
	executor := NewQueryExecutor()
	connID := uuid.New()

	mockDriver := &MockDriver{
		queryFunc: func(ctx context.Context, query string, params ...any) (*driver.Row, error) {
			return &driver.Row{
				Data:        map[string]any{"id": 1, "name": "test"},
				ColumnNames: []string{"id", "name"},
			}, nil
		},
	}

	ctx := context.Background()
	result, err := executor.Execute(ctx, connID, mockDriver, "SELECT * FROM test")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.ResultSet)
	assert.Equal(t, int64(1), result.ResultSet.RowCount)
	assert.NotEmpty(t, result.QueryID)
}

// TestQueryExecutor_ExecuteWithTimeout tests ExecuteWithTimeout.
func TestQueryExecutor_ExecuteWithTimeout(t *testing.T) {
	executor := NewQueryExecutor()
	executor.SetDefaultTimeout(100 * time.Millisecond)
	connID := uuid.New()

	mockDriver := &MockDriver{
		queryFunc: func(ctx context.Context, query string, params ...any) (*driver.Row, error) {
			// Simulate slow query
			time.Sleep(50 * time.Millisecond)
			return &driver.Row{
				Data:        map[string]any{"result": "success"},
				ColumnNames: []string{"result"},
			}, nil
		},
	}

	ctx := context.Background()
	result, err := executor.ExecuteWithTimeout(ctx, connID, mockDriver, "SELECT * FROM test", 200*time.Millisecond)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestQueryExecutor_Execute_Cancel tests query cancellation.
func TestQueryExecutor_Execute_Cancel(t *testing.T) {
	executor := NewQueryExecutor()
	connID := uuid.New()

	queryStarted := make(chan struct{})
	queryCancelled := make(chan struct{})

	mockDriver := &MockDriver{
		queryFunc: func(ctx context.Context, query string, params ...any) (*driver.Row, error) {
			close(queryStarted)
			<-ctx.Done()
			close(queryCancelled)
			return nil, ctx.Err()
		},
	}

	ctx := context.Background()

	// Start query in goroutine
	go func() {
		_, _ = executor.Execute(ctx, connID, mockDriver, "SELECT * FROM test")
	}()

	// Wait for query to start
	<-queryStarted

	// Give it a moment for the query ID to be set
	time.Sleep(10 * time.Millisecond)

	// Cancel the query
	activeQuery := executor.GetActiveQuery(connID)
	if activeQuery != nil {
		err := executor.Cancel(activeQuery.ID)
		assert.NoError(t, err)
	}

	// Wait for cancellation to propagate
	select {
	case <-queryCancelled:
		// Success - query was cancelled
	case <-time.After(2 * time.Second):
		t.Fatal("Query cancellation timed out")
	}
}

// TestQueryExecutor_ExecuteStatements tests multi-statement execution.
func TestQueryExecutor_ExecuteStatements(t *testing.T) {
	executor := NewQueryExecutor()
	connID := uuid.New()

	statementCount := 0
	mockDriver := &MockDriver{
		queryFunc: func(ctx context.Context, query string, params ...any) (*driver.Row, error) {
			statementCount++
			return &driver.Row{
				Data:        map[string]any{"stmt": statementCount},
				ColumnNames: []string{"stmt"},
			}, nil
		},
	}

	ctx := context.Background()
	result, err := executor.ExecuteStatements(ctx, connID, mockDriver, "SELECT 1; SELECT 2; SELECT 3")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Results, 3)
	assert.False(t, result.PartialFail)
	assert.Equal(t, int64(1), result.Results[0].ResultSet.RowCount)
	assert.Equal(t, int64(1), result.Results[1].ResultSet.RowCount)
	assert.Equal(t, int64(1), result.Results[2].ResultSet.RowCount)
}

// TestQueryExecutor_ExecuteStatements_PartialFailure tests partial failure handling.
func TestQueryExecutor_ExecuteStatements_PartialFailure(t *testing.T) {
	executor := NewQueryExecutor()
	connID := uuid.New()

	callCount := 0
	mockDriver := &MockDriver{
		queryFunc: func(ctx context.Context, query string, params ...any) (*driver.Row, error) {
			callCount++
			if callCount == 2 {
				return nil, assert.AnError
			}
			return &driver.Row{
				Data:        map[string]any{"result": callCount},
				ColumnNames: []string{"result"},
			}, nil
		},
	}

	ctx := context.Background()
	result, err := executor.ExecuteStatements(ctx, connID, mockDriver, "SELECT 1; SELECT 2; SELECT 3")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.PartialFail)
	assert.Len(t, result.Results, 3)
	assert.True(t, result.Results[0].Success)
	assert.False(t, result.Results[1].Success)
	assert.Equal(t, "assert.AnError general error for testing", result.Results[1].Error)
	assert.True(t, result.Results[2].Success)
}

// TestQueryExecutor_ExecuteStatements_Empty tests empty query handling.
func TestQueryExecutor_ExecuteStatements_Empty(t *testing.T) {
	executor := NewQueryExecutor()
	connID := uuid.New()
	mockDriver := &MockDriver{}

	ctx := context.Background()
	result, err := executor.ExecuteStatements(ctx, connID, mockDriver, "   ;  ;  ")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no valid statements found")
}

// TestQueryExecutor_GetHistory tests history tracking.
func TestQueryExecutor_GetHistory(t *testing.T) {
	executor := NewQueryExecutor()
	connID := uuid.New()

	mockDriver := &MockDriver{
		queryFunc: func(ctx context.Context, query string, params ...any) (*driver.Row, error) {
			return &driver.Row{
				Data:        map[string]any{"result": "test"},
				ColumnNames: []string{"result"},
			}, nil
		},
	}

	ctx := context.Background()

	// Execute multiple queries
	_, _ = executor.Execute(ctx, connID, mockDriver, "SELECT 1")
	_, _ = executor.Execute(ctx, connID, mockDriver, "SELECT 2")
	_, _ = executor.Execute(ctx, connID, mockDriver, "SELECT 3")

	// Get all history
	history := executor.GetHistory(connID, 0)
	assert.Len(t, history, 3)

	// Get last 2
	history = executor.GetHistory(connID, 2)
	assert.Len(t, history, 2)
	assert.Contains(t, history[0].Query, "SELECT 2")
	assert.Contains(t, history[1].Query, "SELECT 3")
}

// TestQueryExecutor_ClearHistory tests history clearing.
func TestQueryExecutor_ClearHistory(t *testing.T) {
	executor := NewQueryExecutor()
	connID := uuid.New()

	mockDriver := &MockDriver{
		queryFunc: func(ctx context.Context, query string, params ...any) (*driver.Row, error) {
			return &driver.Row{
				Data:        map[string]any{"result": "test"},
				ColumnNames: []string{"result"},
			}, nil
		},
	}

	ctx := context.Background()
	_, _ = executor.Execute(ctx, connID, mockDriver, "SELECT 1")
	_, _ = executor.Execute(ctx, connID, mockDriver, "SELECT 2")

	// Verify history exists
	history := executor.GetHistory(connID, 0)
	assert.Len(t, history, 2)

	// Clear history
	executor.ClearHistory(connID)

	// Verify history is empty
	history = executor.GetHistory(connID, 0)
	assert.Empty(t, history)
}

// TestQueryExecutor_ExecuteStreaming tests streaming execution.
func TestQueryExecutor_ExecuteStreaming(t *testing.T) {
	executor := NewQueryExecutor()
	connID := uuid.New()

	mockDriver := &MockDriver{
		queryFunc: func(ctx context.Context, query string, params ...any) (*driver.Row, error) {
			return &driver.Row{
				Data:        map[string]any{"stream": "data"},
				ColumnNames: []string{"stream"},
			}, nil
		},
	}

	ctx := context.Background()
	chunkChan, err := executor.ExecuteStreaming(ctx, connID, mockDriver, "SELECT * FROM test")

	assert.NoError(t, err)
	assert.NotNil(t, chunkChan)

	// Read chunks
	var chunks []*StreamingChunk
	for chunk := range chunkChan {
		chunks = append(chunks, chunk)
	}

	assert.Len(t, chunks, 1)
	assert.NotNil(t, chunks[0].ResultSet)
	assert.True(t, chunks[0].IsLast)
	assert.Empty(t, chunks[0].Error)
}

// TestQueryExecutor_ExecuteStreaming_Error tests streaming error handling.
func TestQueryExecutor_ExecuteStreaming_Error(t *testing.T) {
	executor := NewQueryExecutor()
	connID := uuid.New()

	mockDriver := &MockDriver{
		queryFunc: func(ctx context.Context, query string, params ...any) (*driver.Row, error) {
			return nil, assert.AnError
		},
	}

	ctx := context.Background()
	chunkChan, err := executor.ExecuteStreaming(ctx, connID, mockDriver, "SELECT * FROM test")

	assert.NoError(t, err)

	// Read error chunk
	chunk := <-chunkChan
	assert.NotNil(t, chunk)
	assert.Contains(t, chunk.Error, "assert.AnError")
	assert.True(t, chunk.IsLast)
}

// TestQueryExecutor_ExecuteMultiStreaming tests multi-statement streaming.
func TestQueryExecutor_ExecuteMultiStreaming(t *testing.T) {
	executor := NewQueryExecutor()
	connID := uuid.New()

	statementIndex := 0
	mockDriver := &MockDriver{
		queryFunc: func(ctx context.Context, query string, params ...any) (*driver.Row, error) {
			statementIndex++
			return &driver.Row{
				Data:        map[string]any{"stmt": statementIndex},
				ColumnNames: []string{"stmt"},
			}, nil
		},
	}

	ctx := context.Background()
	chunkChan, err := executor.ExecuteMultiStreaming(ctx, connID, mockDriver, "SELECT 1; SELECT 2; SELECT 3")

	assert.NoError(t, err)

	var chunks []*StreamingChunk
	for chunk := range chunkChan {
		chunks = append(chunks, chunk)
	}

	assert.Len(t, chunks, 3)
	for i, chunk := range chunks {
		assert.NotNil(t, chunk.ResultSet)
		if i == 2 {
			assert.True(t, chunk.IsLast)
		} else {
			assert.False(t, chunk.IsLast)
		}
	}
}

// TestQueryExecutor_CancelConnection tests connection-wide cancellation.
func TestQueryExecutor_CancelConnection(t *testing.T) {
	executor := NewQueryExecutor()
	connID := uuid.New()

	queryCancelled := make(chan struct{})

	mockDriver := &MockDriver{
		queryFunc: func(ctx context.Context, query string, params ...any) (*driver.Row, error) {
			<-ctx.Done()
			close(queryCancelled)
			return nil, ctx.Err()
		},
	}

	ctx := context.Background()

	// Start query
	go func() {
		_, _ = executor.Execute(ctx, connID, mockDriver, "SELECT * FROM test")
	}()

	// Give it time to start
	time.Sleep(10 * time.Millisecond)

	// Cancel connection
	err := executor.CancelConnection(connID)
	assert.NoError(t, err)

	// Wait for cancellation
	select {
	case <-queryCancelled:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Connection cancellation timed out")
	}
}

// TestQueryExecutor_CloseConnection tests connection cleanup.
func TestQueryExecutor_CloseConnection(t *testing.T) {
	executor := NewQueryExecutor()
	connID := uuid.New()

	// Create a session by executing a query
	mockDriver := &MockDriver{
		queryFunc: func(ctx context.Context, query string, params ...any) (*driver.Row, error) {
			return &driver.Row{
				Data:        map[string]any{"result": "test"},
				ColumnNames: []string{"result"},
			}, nil
		},
	}

	ctx := context.Background()
	_, _ = executor.Execute(ctx, connID, mockDriver, "SELECT 1")

	// Verify session exists
	assert.Equal(t, 1, executor.GetSessionCount())

	// Close connection
	executor.CloseConnection(connID)

	// Session should be removed
	assert.Equal(t, 0, executor.GetSessionCount())
}

// TestParseStatements tests the parseStatements helper function.
func TestParseStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"single statement", "SELECT 1", 1},
		{"multiple statements", "SELECT 1; SELECT 2; SELECT 3", 3},
		{"statements with extra whitespace", "  SELECT 1  ;   SELECT 2  ", 2},
		{"empty statements", "SELECT 1;; SELECT 2;", 2},
		{"only whitespace", "   ;   ;   ", 0},
		{"empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statements := parseStatements(tt.input)
			assert.Len(t, statements, tt.expected)
		})
	}
}
