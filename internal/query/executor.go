// Package query provides query execution services with timeout, cancellation, and result streaming.
package query

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"tablepro/internal/driver"
)

// QueryExecutor manages query execution with timeout and cancellation support.
// It tracks active queries per connection and maintains query history.
// Thread-safe for concurrent use across multiple connections.
type QueryExecutor struct {
	mu             sync.RWMutex
	sessions       map[uuid.UUID]*QuerySession
	defaultTimeout time.Duration
}

// QuerySession represents an active query session for a connection.
// Contains the active query (if any), history, and database type.
type QuerySession struct {
	ConnectionID uuid.UUID
	ActiveQuery  *ActiveQuery
	History      []QueryHistoryEntry
	DBType       driver.DatabaseType
}

// ActiveQuery represents a currently executing query.
// Holds the query context for cancellation and timing information.
type ActiveQuery struct {
	ID        uuid.UUID
	Query     string
	Context   context.Context
	Cancel    context.CancelFunc
	StartedAt time.Time
}

// QueryResult represents the result of a query execution.
// Contains the result set, query ID for tracking, and execution duration.
type QueryResult struct {
	ResultSet  *ResultSet         `json:"resultSet"`
	QueryID    uuid.UUID          `json:"queryId"`
	Duration   time.Duration      `json:"duration"`
	Pagination *PaginationContext `json:"pagination,omitempty"`
}

// StatementResult represents the result of a single statement in a multi-statement query.
// Used when executing multiple SQL statements separated by semicolons.
type StatementResult struct {
	Statement    string        `json:"statement"`
	ResultSet    *ResultSet    `json:"resultSet,omitempty"`
	RowsAffected int64         `json:"rowsAffected,omitempty"`
	Error        string        `json:"error,omitempty"`
	Success      bool          `json:"success"`
	Duration     time.Duration `json:"duration"`
}

// MultiStatementResult represents results from executing multiple statements.
// Aggregates individual StatementResults with total execution time.
type MultiStatementResult struct {
	QueryID       uuid.UUID         `json:"queryId"`
	TotalDuration time.Duration     `json:"totalDuration"`
	Results       []StatementResult `json:"results"`
	PartialFail   bool              `json:"partialFail"`
}

// StreamingChunk represents a chunk of streamed query results.
// Used for large result sets that are sent incrementally to the frontend.
type StreamingChunk struct {
	ResultSet *ResultSet `json:"resultSet"`
	ChunkID   int        `json:"chunkId"`
	IsLast    bool       `json:"isLast"`
	Error     string     `json:"error,omitempty"`
}

// NewQueryExecutor creates a new QueryExecutor with default timeout.
// The default timeout is 30 seconds.
func NewQueryExecutor() *QueryExecutor {
	return &QueryExecutor{
		sessions:       make(map[uuid.UUID]*QuerySession),
		defaultTimeout: 30 * time.Second,
	}
}

// SetDefaultTimeout sets the default query timeout.
// Applied to all queries unless overridden with ExecuteWithTimeout.
func (e *QueryExecutor) SetDefaultTimeout(timeout time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.defaultTimeout = timeout
}

// Execute executes a query with the default timeout.
// Uses the driver to run the query and tracks it in the session.
// Returns QueryResult with result set and execution duration.
func (e *QueryExecutor) Execute(ctx context.Context, connID uuid.UUID, drv driver.DatabaseDriver, query string) (*QueryResult, error) {
	return e.ExecuteWithTimeout(ctx, connID, drv, query, e.defaultTimeout)
}

// ExecuteWithTimeout executes a query with a custom timeout.
// Creates a query session, tracks active query, and records history.
// Returns error wrapped with execution context.
func (e *QueryExecutor) ExecuteWithTimeout(ctx context.Context, connID uuid.UUID, drv driver.DatabaseDriver, query string, timeout time.Duration) (*QueryResult, error) {
	e.mu.Lock()

	session, exists := e.sessions[connID]
	if !exists {
		session = &QuerySession{
			ConnectionID: connID,
			History:      make([]QueryHistoryEntry, 0),
			DBType:       driver.DatabaseTypeUnknown,
		}
		e.sessions[connID] = session
	}

	queryCtx, cancel := context.WithTimeout(ctx, timeout)
	queryID := uuid.New()
	activeQuery := &ActiveQuery{
		ID:        queryID,
		Query:     query,
		Context:   queryCtx,
		Cancel:    cancel,
		StartedAt: time.Now(),
	}
	session.ActiveQuery = activeQuery

	e.mu.Unlock()

	startTime := time.Now()
	row, err := drv.Query(queryCtx, query)
	duration := time.Since(startTime)

	var resultSet *ResultSet
	if err == nil && row != nil {
		resultSet = NewResultSetFromRows([]*driver.Row{row}, duration, query, session.DBType)
	}

	historyEntry := QueryHistoryEntry{
		ID:         queryID,
		Query:      query,
		Timestamp:  time.Now(),
		Duration:   duration,
		Connection: connID,
	}

	if err != nil {
		historyEntry.Success = false
		historyEntry.Error = err.Error()
	} else {
		historyEntry.Success = true
		if resultSet != nil {
			historyEntry.RowCount = resultSet.RowCount
		}
	}

	e.mu.Lock()
	session.ActiveQuery = nil
	session.History = append(session.History, historyEntry)
	e.mu.Unlock()

	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return &QueryResult{
		ResultSet: resultSet,
		QueryID:   queryID,
		Duration:  duration,
	}, nil
}

// Cancel cancels a running query by ID.
// Searches all sessions for the active query and calls context.CancelFunc.
// Returns error if query not found or already completed.
func (e *QueryExecutor) Cancel(queryID uuid.UUID) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, session := range e.sessions {
		if session.ActiveQuery != nil && session.ActiveQuery.ID == queryID {
			session.ActiveQuery.Cancel()
			return nil
		}
	}

	return fmt.Errorf("query %s not found or already completed", queryID)
}

// CancelConnection cancels all active queries for a connection.
// Safe to call even if no queries are active.
func (e *QueryExecutor) CancelConnection(connID uuid.UUID) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	session, exists := e.sessions[connID]
	if !exists {
		return nil
	}

	if session.ActiveQuery != nil {
		session.ActiveQuery.Cancel()
	}

	return nil
}

// GetActiveQuery returns the active query for a connection.
// Returns nil if no query is currently executing.
func (e *QueryExecutor) GetActiveQuery(connID uuid.UUID) *ActiveQuery {
	e.mu.RLock()
	defer e.mu.RUnlock()

	session, exists := e.sessions[connID]
	if !exists {
		return nil
	}

	return session.ActiveQuery
}

// GetHistory returns the query history for a connection.
// If limit is positive, returns only the last N entries.
// Returns empty slice if connection has no history.
func (e *QueryExecutor) GetHistory(connID uuid.UUID, limit int) []QueryHistoryEntry {
	e.mu.RLock()
	defer e.mu.RUnlock()

	session, exists := e.sessions[connID]
	if !exists {
		return []QueryHistoryEntry{}
	}

	if limit <= 0 || limit > len(session.History) {
		return session.History
	}

	start := len(session.History) - limit
	return session.History[start:]
}

// ClearHistory clears the query history for a connection.
// Does not affect active queries.
func (e *QueryExecutor) ClearHistory(connID uuid.UUID) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, exists := e.sessions[connID]
	if exists {
		session.History = make([]QueryHistoryEntry, 0)
	}
}

// CloseConnection closes a connection and cleans up its session.
// Cancels any active query before removing the session.
func (e *QueryExecutor) CloseConnection(connID uuid.UUID) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if session, exists := e.sessions[connID]; exists && session.ActiveQuery != nil {
		session.ActiveQuery.Cancel()
	}

	delete(e.sessions, connID)
}

// GetSessionCount returns the number of active sessions.
// Useful for monitoring and debugging.
func (e *QueryExecutor) GetSessionCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.sessions)
}

// ExecuteStatements executes multiple SQL statements separated by semicolons.
// Uses default timeout for entire batch.
// Returns MultiStatementResult with individual statement results.
func (e *QueryExecutor) ExecuteStatements(ctx context.Context, connID uuid.UUID, drv driver.DatabaseDriver, queries string) (*MultiStatementResult, error) {
	return e.ExecuteStatementsWithTimeout(ctx, connID, drv, queries, e.defaultTimeout)
}

// ExecuteStatementsWithTimeout executes multiple SQL statements with a custom timeout.
// Each statement executes sequentially. Partial failures are tracked.
func (e *QueryExecutor) ExecuteStatementsWithTimeout(ctx context.Context, connID uuid.UUID, drv driver.DatabaseDriver, queries string, timeout time.Duration) (*MultiStatementResult, error) {
	statements := parseStatements(queries)
	if len(statements) == 0 {
		return nil, fmt.Errorf("no valid statements found")
	}

	e.mu.Lock()

	session, exists := e.sessions[connID]
	if !exists {
		session = &QuerySession{
			ConnectionID: connID,
			History:      make([]QueryHistoryEntry, 0),
			DBType:       driver.DatabaseTypeUnknown,
		}
		e.sessions[connID] = session
	}

	queryCtx, cancel := context.WithTimeout(ctx, timeout)
	queryID := uuid.New()
	activeQuery := &ActiveQuery{
		ID:        queryID,
		Query:     queries,
		Context:   queryCtx,
		Cancel:    cancel,
		StartedAt: time.Now(),
	}
	session.ActiveQuery = activeQuery

	e.mu.Unlock()

	startTime := time.Now()
	results := make([]StatementResult, 0, len(statements))
	partialFail := false

	for _, stmt := range statements {
		stmtStartTime := time.Now()
		result := StatementResult{
			Statement: stmt,
		}

		if queryCtx.Err() != nil {
			result.Success = false
			result.Error = "query cancelled"
			result.Duration = time.Since(stmtStartTime)
			results = append(results, result)
			partialFail = true
			continue
		}

		row, err := drv.Query(queryCtx, stmt)
		result.Duration = time.Since(stmtStartTime)

		if err != nil {
			result.Success = false
			result.Error = err.Error()
			partialFail = true
		} else if row != nil {
			result.Success = true
			result.ResultSet = NewResultSetFromRows([]*driver.Row{row}, result.Duration, stmt, session.DBType)
			result.RowsAffected = result.ResultSet.RowCount
		} else {
			result.Success = true
		}

		results = append(results, result)
	}

	totalDuration := time.Since(startTime)

	historyEntry := QueryHistoryEntry{
		ID:         queryID,
		Query:      queries,
		Timestamp:  time.Now(),
		Duration:   totalDuration,
		Connection: connID,
	}

	allSuccess := !partialFail
	historyEntry.Success = allSuccess
	if !allSuccess {
		historyEntry.Error = "partial failure in multi-statement execution"
	}

	e.mu.Lock()
	session.ActiveQuery = nil
	session.History = append(session.History, historyEntry)
	e.mu.Unlock()

	return &MultiStatementResult{
		QueryID:       queryID,
		TotalDuration: totalDuration,
		Results:       results,
		PartialFail:   partialFail,
	}, nil
}

// parseStatements splits a query string into individual statements by semicolons.
// Trims whitespace and filters empty statements.
func parseStatements(queries string) []string {
	parts := strings.Split(queries, ";")
	statements := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			statements = append(statements, trimmed)
		}
	}

	return statements
}

// ExecuteStreaming executes a query and streams results in chunks of 1000 rows.
// Returns a channel of StreamingChunk for progressive rendering.
func (e *QueryExecutor) ExecuteStreaming(ctx context.Context, connID uuid.UUID, drv driver.DatabaseDriver, query string) (<-chan *StreamingChunk, error) {
	return e.ExecuteStreamingWithChunkSize(ctx, connID, drv, query, 1000)
}

// ExecuteStreamingWithChunkSize executes a query and streams results with custom chunk size.
// Chunk size is clamped to minimum of 1 if invalid.
func (e *QueryExecutor) ExecuteStreamingWithChunkSize(ctx context.Context, connID uuid.UUID, drv driver.DatabaseDriver, query string, chunkSize int) (<-chan *StreamingChunk, error) {
	if chunkSize <= 0 {
		chunkSize = 1000
	}

	chunkChan := make(chan *StreamingChunk, 2)

	e.mu.Lock()
	session, exists := e.sessions[connID]
	if !exists {
		session = &QuerySession{
			ConnectionID: connID,
			History:      make([]QueryHistoryEntry, 0),
			DBType:       driver.DatabaseTypeUnknown,
		}
		e.sessions[connID] = session
	}

	queryCtx, cancel := context.WithTimeout(ctx, e.defaultTimeout)
	queryID := uuid.New()
	activeQuery := &ActiveQuery{
		ID:        queryID,
		Query:     query,
		Context:   queryCtx,
		Cancel:    cancel,
		StartedAt: time.Now(),
	}
	session.ActiveQuery = activeQuery
	e.mu.Unlock()

	startTime := time.Now()

	go func() {
		defer close(chunkChan)
		defer func() {
			e.mu.Lock()
			session.ActiveQuery = nil
			e.mu.Unlock()
		}()

		rows, err := drv.Query(queryCtx, query)

		if err != nil {
			chunkChan <- &StreamingChunk{
				Error:  err.Error(),
				IsLast: true,
			}

			historyEntry := QueryHistoryEntry{
				ID:         queryID,
				Query:      query,
				Timestamp:  time.Now(),
				Duration:   time.Since(startTime),
				Success:    false,
				Error:      err.Error(),
				Connection: connID,
			}
			e.mu.Lock()
			session.History = append(session.History, historyEntry)
			e.mu.Unlock()
			return
		}

		if rows == nil {
			chunkChan <- &StreamingChunk{
				ResultSet: &ResultSet{
					Columns:   make([]ColumnInfo, 0),
					Rows:      make([][]interface{}, 0),
					RowCount:  0,
					QueryTime: time.Since(startTime),
					Statement: query,
				},
				ChunkID: 0,
				IsLast:  true,
			}
			return
		}

		rowSlice := []*driver.Row{rows}
		resultSet := NewResultSetFromRows(rowSlice, time.Since(startTime), query, session.DBType)

		chunkChan <- &StreamingChunk{
			ResultSet: resultSet,
			ChunkID:   0,
			IsLast:    true,
		}

		historyEntry := QueryHistoryEntry{
			ID:         queryID,
			Query:      query,
			Timestamp:  time.Now(),
			Duration:   time.Since(startTime),
			Success:    true,
			RowCount:   resultSet.RowCount,
			Connection: connID,
		}
		e.mu.Lock()
		session.History = append(session.History, historyEntry)
		e.mu.Unlock()
	}()

	return chunkChan, nil
}

var streamingIDCounter uint64

// ExecuteMultiStreaming executes multiple statements and streams each result set.
// Each statement result is sent as a separate chunk on the returned channel.
func (e *QueryExecutor) ExecuteMultiStreaming(ctx context.Context, connID uuid.UUID, drv driver.DatabaseDriver, queries string) (<-chan *StreamingChunk, error) {
	statements := parseStatements(queries)
	if len(statements) == 0 {
		ch := make(chan *StreamingChunk, 1)
		ch <- &StreamingChunk{
			Error:  "no valid statements found",
			IsLast: true,
		}
		close(ch)
		return ch, nil
	}

	chunkChan := make(chan *StreamingChunk, len(statements)+1)

	e.mu.Lock()
	session, exists := e.sessions[connID]
	if !exists {
		session = &QuerySession{
			ConnectionID: connID,
			History:      make([]QueryHistoryEntry, 0),
			DBType:       driver.DatabaseTypeUnknown,
		}
		e.sessions[connID] = session
	}
	e.mu.Unlock()

	go func() {
		defer close(chunkChan)

		for i, stmt := range statements {
			select {
			case <-ctx.Done():
				chunkChan <- &StreamingChunk{
					Error:  "query cancelled",
					IsLast: true,
				}
				return
			default:
			}

			stmtCtx, cancel := context.WithTimeout(ctx, e.defaultTimeout)
			resultSet, err := drv.Query(stmtCtx, stmt)
			cancel()

			isLast := (i == len(statements)-1)

			if err != nil {
				chunkChan <- &StreamingChunk{
					Error:   err.Error(),
					ChunkID: int(atomic.AddUint64(&streamingIDCounter, 1)),
					IsLast:  isLast,
				}
				continue
			}

			var rs *ResultSet
			if resultSet != nil {
				rs = NewResultSetFromRows([]*driver.Row{resultSet}, 0, stmt, session.DBType)
			} else {
				rs = &ResultSet{
					Columns:   make([]ColumnInfo, 0),
					Rows:      make([][]interface{}, 0),
					RowCount:  0,
					QueryTime: 0,
					Statement: stmt,
				}
			}

			chunkChan <- &StreamingChunk{
				ResultSet: rs,
				ChunkID:   int(atomic.AddUint64(&streamingIDCounter, 1)),
				IsLast:    isLast,
			}
		}
	}()

	return chunkChan, nil
}
