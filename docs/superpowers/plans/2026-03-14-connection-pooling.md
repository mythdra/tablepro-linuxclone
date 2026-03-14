# Connection Pooling Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement connection pooling per session to enable connection reuse and handle concurrent queries efficiently.

**Architecture:** Create a connection pool within each QuerySession that manages multiple `*sql.DB` connections. The pool uses a buffered channel for available connections and implements blocking with timeout when exhausted. Pool operations are protected by sync.RWMutex for thread safety.

**Tech Stack:** Go stdlib (sync, channel, context, time, database/sql), existing QueryExecutor from Phase 5.

---

## File Structure

**Files to Create:**
- `internal/session/pool.go` - Connection pool implementation
- `internal/session/pool_test.go` - Unit tests for pool functionality

**Files to Modify:**
- `internal/query/executor.go:16-23` - Add pool to QuerySession struct
- `internal/query/executor.go:101-170` - Integrate pool into ExecuteWithTimeout
- `internal/query/executor.go:258-270` - Add CloseConnection to drain pool

**Dependencies:**
- Coordinate with Task Group 1 (SessionManager) - do NOT modify core SessionManager structure
- Integration with ConnectionManager happens in Task Group 8 - for now, use mock/simulated connections

---

## Task 1: Create Session Package Structure

**Files:**
- Create: `internal/session/pool.go`
- Create: `internal/session/pool_test.go`

- [ ] **Step 1: Create session directory**

```bash
mkdir -p /Users/can/code/tablepro-fork/.worktrees/phase-06-sessions/internal/session
```

- [ ] **Step 2: Create pool.go with package declaration and imports**

```go
// Package session manages database connection pooling per session.
package session

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)
```

- [ ] **Step 3: Commit**

```bash
cd /Users/can/code/tablepro-fork/.worktrees/phase-06-sessions
git add internal/session/
git commit -m "feat: add session package structure"
```

---

## Task 2: Define Session Struct with Pool Fields

**Files:**
- Modify: `internal/session/pool.go:8-15`
- Test: `internal/session/pool_test.go`

- [ ] **Step 1: Write the failing test**

```go
// In pool_test.go
func TestSession_CreatesWithDefaultPoolSize(t *testing.T) {
	connID := uuid.New()
	session := NewSession(connID, 5)
	
	if session == nil {
		t.Fatal("expected non-nil session")
	}
	if session.ConnectionID != connID {
		t.Errorf("expected connection ID %s, got %s", connID, session.ConnectionID)
	}
	if session.MaxPoolSize != 5 {
		t.Errorf("expected max pool size 5, got %d", session.MaxPoolSize)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/session/... -v`
Expected: FAIL with "undefined: NewSession"

- [ ] **Step 3: Define Session struct and NewSession constructor**

```go
// In pool.go

// SessionConfig holds configuration for a connection pool session.
type SessionConfig struct {
	// MaxPoolSize is the maximum number of connections in the pool.
	// Default is 5 if not specified.
	MaxPoolSize int
	
	// PoolTimeout is the timeout for acquiring a connection from the pool.
	// Default is 30 seconds if not specified.
	PoolTimeout time.Duration
}

// Session represents a database session with connection pooling.
// It manages a pool of reusable connections for a specific database connection.
// Thread-safe for concurrent use across multiple goroutines.
type Session struct {
	// ConnectionID is the unique identifier for this session's connection.
	ConnectionID uuid.UUID
	
	// MaxPoolSize is the maximum number of connections in the pool.
	MaxPoolSize int
	
	// PoolTimeout is the timeout for acquiring a connection from the pool.
	PoolTimeout time.Duration
	
	// mu protects access to the pool and size counters.
	mu sync.RWMutex
	
	// availableConns is a buffered channel holding available connections.
	// When a connection is needed, it's taken from this channel.
	// When returned, it's sent back to this channel.
	availableConns chan *sql.DB
	
	// currentSize tracks the total number of connections (available + in use).
	currentSize int
	
	// waitQueue signals when a connection becomes available.
	waitQueue chan struct{}
	
	// closed indicates if the session has been closed.
	closed bool
}

// DefaultPoolSize is the default maximum pool size.
const DefaultPoolSize = 5

// DefaultPoolTimeout is the default timeout for acquiring a connection.
const DefaultPoolTimeout = 30 * time.Second

// NewSession creates a new Session with the given connection ID and configuration.
// If config is nil, uses default values (5 connections, 30s timeout).
func NewSession(connID uuid.UUID, config *SessionConfig) *Session {
	maxSize := DefaultPoolSize
	timeout := DefaultPoolTimeout
	
	if config != nil {
		if config.MaxPoolSize > 0 {
			maxSize = config.MaxPoolSize
		}
		if config.PoolTimeout > 0 {
			timeout = config.PoolTimeout
		}
	}
	
	return &Session{
		ConnectionID:   connID,
		MaxPoolSize:    maxSize,
		PoolTimeout:    timeout,
		availableConns: make(chan *sql.DB, maxSize),
		currentSize:    0,
		waitQueue:      make(chan struct{}, 1),
		closed:         false,
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/session/... -v -run TestSession_CreatesWithDefaultPoolSize`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/session/pool.go internal/session/pool_test.go
git commit -m "feat: define Session struct with pool fields"
```

---

## Task 3: Implement createConnection Method

**Files:**
- Modify: `internal/session/pool.go`
- Test: `internal/session/pool_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestCreateConnection_AddsToPool(t *testing.T) {
	connID := uuid.New()
	session := NewSession(connID, &SessionConfig{MaxPoolSize: 5})
	
	// Create a mock sql.DB (in real implementation, this comes from ConnectionManager)
	mockDB := &sql.DB{}
	
	err := session.addConnectionToPool(mockDB)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	
	// Verify connection is in available pool
	if len(session.availableConns) != 1 {
		t.Errorf("expected 1 available connection, got %d", len(session.availableConns))
	}
	if session.currentSize != 1 {
		t.Errorf("expected currentSize 1, got %d", session.currentSize)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/session/... -v -run TestCreateConnection_AddsToPool`
Expected: FAIL with "undefined: addConnectionToPool"

- [ ] **Step 3: Write minimal implementation**

```go
// addConnectionToPool adds a connection to the available pool.
// Increments currentSize counter.
// Must be called with session.mu locked.
func (s *Session) addConnectionToPool(db *sql.DB) error {
	if s.closed {
		return fmt.Errorf("session is closed")
	}
	
	select {
	case s.availableConns <- db:
		s.currentSize++
		return nil
	default:
		return fmt.Errorf("pool is full")
	}
}

// createConnection creates a new database connection and adds it to the pool.
// In production, this delegates to ConnectionManager to create the actual connection.
// For now, accepts a connection factory function for testing.
func (s *Session) createConnection(ctx context.Context, connFactory func(context.Context) (*sql.DB, error)) (*sql.DB, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.closed {
		return nil, fmt.Errorf("session is closed")
	}
	
	if s.currentSize >= s.MaxPoolSize {
		return nil, fmt.Errorf("pool at maximum capacity")
	}
	
	db, err := connFactory(ctx)
	if err != nil {
		return nil, fmt.Errorf("connection creation failed: %w", err)
	}
	
	if err := s.addConnectionToPool(db); err != nil {
		db.Close()
		return nil, err
	}
	
	return db, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/session/... -v -run TestCreateConnection_AddsToPool`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/session/pool.go internal/session/pool_test.go
git commit -m "feat: implement createConnection method"
```

---

## Task 4: Implement getConnectionFromPool Method

**Files:**
- Modify: `internal/session/pool.go`
- Test: `internal/session/pool_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestGetConnectionFromPool_CreatesWhenEmpty(t *testing.T) {
	connID := uuid.New()
	session := NewSession(connID, &SessionConfig{MaxPoolSize: 5})
	
	connFactory := func(ctx context.Context) (*sql.DB, error) {
		return &sql.DB{}, nil
	}
	
	ctx := context.Background()
	db, err := session.GetConnection(ctx, connFactory)
	
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if db == nil {
		t.Fatal("expected non-nil connection")
	}
	if session.currentSize != 1 {
		t.Errorf("expected currentSize 1, got %d", session.currentSize)
	}
}

func TestGetConnectionFromPool_ReusesExisting(t *testing.T) {
	connID := uuid.New()
	session := NewSession(connID, &SessionConfig{MaxPoolSize: 5})
	
	creationCount := 0
	connFactory := func(ctx context.Context) (*sql.DB, error) {
		creationCount++
		return &sql.DB{}, nil
	}
	
	ctx := context.Background()
	
	// Get first connection (should create)
	db1, err := session.GetConnection(ctx, connFactory)
	if err != nil {
		t.Fatalf("first GetConnection failed: %v", err)
	}
	
	// Return it
	session.ReturnConnection(db1)
	
	// Get second connection (should reuse)
	db2, err := session.GetConnection(ctx, connFactory)
	if err != nil {
		t.Fatalf("second GetConnection failed: %v", err)
	}
	
	if creationCount != 1 {
		t.Errorf("expected 1 connection creation, got %d", creationCount)
	}
	if db1 != db2 {
		t.Error("expected same connection to be reused")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/session/... -v -run "TestGetConnectionFromPool"`
Expected: FAIL with "undefined: GetConnection"

- [ ] **Step 3: Write minimal implementation**

```go
// ErrPoolExhausted is returned when the pool is exhausted and timeout occurs.
var ErrPoolExhausted = errors.New("connection pool exhausted: timeout waiting for available connection")

// GetConnection acquires a connection from the pool.
// If no connections are available and pool is not at max capacity, creates a new one.
// If pool is exhausted, waits up to PoolTimeout for a connection to be returned.
// Returns error if timeout expires or session is closed.
func (s *Session) GetConnection(ctx context.Context, connFactory func(context.Context) (*sql.DB, error)) (*sql.DB, error) {
	s.mu.Lock()
	
	if s.closed {
		s.mu.Unlock()
		return nil, fmt.Errorf("session is closed")
	}
	
	// Try to get available connection (non-blocking)
	select {
	case db := <-s.availableConns:
		s.mu.Unlock()
		return db, nil
	default:
		// No available connections
	}
	
	// If pool not at max, create new connection
	if s.currentSize < s.MaxPoolSize {
		s.mu.Unlock()
		return s.createConnection(ctx, connFactory)
	}
	
	// Pool exhausted - need to wait
	s.mu.Unlock()
	
	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, s.PoolTimeout)
	defer cancel()
	
	// Wait for connection or timeout
	for {
		select {
		case <-timeoutCtx.Done():
			return nil, ErrPoolExhausted
		case <-s.waitQueue:
			// Connection might be available now
			s.mu.Lock()
			select {
			case db := <-s.availableConns:
				s.mu.Unlock()
				return db, nil
			default:
				s.mu.Unlock()
				// Spurious wakeup, continue waiting
			}
		}
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/session/... -v -run "TestGetConnectionFromPool"`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/session/pool.go internal/session/pool_test.go
git commit -m "feat: implement getConnectionFromPool with reuse logic"
```

---

## Task 5: Implement returnConnectionToPool Method

**Files:**
- Modify: `internal/session/pool.go`
- Test: `internal/session/pool_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestReturnConnectionToPool(t *testing.T) {
	connID := uuid.New()
	session := NewSession(connID, &SessionConfig{MaxPoolSize: 5})
	
	connFactory := func(ctx context.Context) (*sql.DB, error) {
		return &sql.DB{}, nil
	}
	
	ctx := context.Background()
	db, _ := session.GetConnection(ctx, connFactory)
	
	if session.currentSize != 1 {
		t.Fatalf("expected currentSize 1, got %d", session.currentSize)
	}
	if len(session.availableConns) != 0 {
		t.Fatalf("expected 0 available (connection in use), got %d", len(session.availableConns))
	}
	
	session.ReturnConnection(db)
	
	if len(session.availableConns) != 1 {
		t.Errorf("expected 1 available connection, got %d", len(session.availableConns))
	}
	if session.currentSize != 1 {
		t.Errorf("expected currentSize 1 (unchanged), got %d", session.currentSize)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/session/... -v -run TestReturnConnectionToPool`
Expected: FAIL with "undefined: ReturnConnection"

- [ ] **Step 3: Write minimal implementation**

```go
// ReturnConnection returns a connection to the pool.
// Signals waiting goroutines that a connection is available.
// Safe to call even if session is closed (will just discard connection).
func (s *Session) ReturnConnection(db *sql.DB) {
	if db == nil {
		return
	}
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.closed {
		// Session closed, just close the connection
		db.Close()
		return
	}
	
	// Try to return to pool (non-blocking)
	select {
	case s.availableConns <- db:
		// Successfully returned
		// Signal waiting goroutines
		select {
		case s.waitQueue <- struct{}{}:
		default:
			// No one waiting, that's fine
		}
	default:
		// Pool is full (shouldn't happen in normal operation)
		// Just close the connection
		db.Close()
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/session/... -v -run TestReturnConnectionToPool`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/session/pool.go internal/session/pool_test.go
git commit -m "feat: implement returnConnectionToPool with signaling"
```

---

## Task 6: Implement Pool Exhaustion Timeout Test

**Files:**
- Modify: `internal/session/pool_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestGetConnectionFromPool_ExhaustionTimeout(t *testing.T) {
	connID := uuid.New()
	session := NewSession(connID, &SessionConfig{
		MaxPoolSize:   2,
		PoolTimeout:   100 * time.Millisecond, // Short timeout for test
	})
	
	connFactory := func(ctx context.Context) (*sql.DB, error) {
		return &sql.DB{}, nil
	}
	
	ctx := context.Background()
	
	// Exhaust the pool
	db1, _ := session.GetConnection(ctx, connFactory)
	db2, _ := session.GetConnection(ctx, connFactory)
	
	// Don't return connections - pool should be exhausted
	start := time.Now()
	_, err := session.GetConnection(ctx, connFactory)
	elapsed := time.Since(start)
	
	if err != ErrPoolExhausted {
		t.Errorf("expected ErrPoolExhausted, got: %v", err)
	}
	
	// Verify timeout occurred (should be close to 100ms, not instant)
	if elapsed < 90*time.Millisecond {
		t.Errorf("expected timeout ~100ms, got %v", elapsed)
	}
	if elapsed > 200*time.Millisecond {
		t.Errorf("timeout took too long: %v", elapsed)
	}
	
	// Cleanup
	session.ReturnConnection(db1)
	session.ReturnConnection(db2)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/session/... -v -run TestGetConnectionFromPool_ExhaustionTimeout`
Expected: This should actually pass if implementation is correct, but run to verify

- [ ] **Step 3: If test fails, debug and fix**

Expected behavior: Test should pass with current implementation. If it doesn't, check timeout logic.

- [ ] **Step 4: Commit**

```bash
git add internal/session/pool_test.go
git commit -m "test: add pool exhaustion timeout test"
```

---

## Task 7: Implement Pool Size Limit Test

**Files:**
- Modify: `internal/session/pool_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestPoolSizeLimit(t *testing.T) {
	connID := uuid.New()
	session := NewSession(connID, &SessionConfig{MaxPoolSize: 3})
	
	connFactory := func(ctx context.Context) (*sql.DB, error) {
		return &sql.DB{}, nil
	}
	
	ctx := context.Background()
	
	// Create 3 connections (at limit)
	connections := make([]*sql.DB, 3)
	for i := 0; i < 3; i++ {
		db, err := session.GetConnection(ctx, connFactory)
		if err != nil {
			t.Fatalf("GetConnection %d failed: %v", i, err)
		}
		connections[i] = db
	}
	
	if session.currentSize != 3 {
		t.Errorf("expected currentSize 3, got %d", session.currentSize)
	}
	
	// Return all connections
	for _, db := range connections {
		session.ReturnConnection(db)
	}
	
	// Verify all are available
	if len(session.availableConns) != 3 {
		t.Errorf("expected 3 available connections, got %d", len(session.availableConns))
	}
	
	// Try to get a 4th connection (should reuse, not create new)
	creationCount := 0
	countingFactory := func(ctx context.Context) (*sql.DB, error) {
		creationCount++
		return &sql.DB{}, nil
	}
	
	db4, err := session.GetConnection(ctx, countingFactory)
	if err != nil {
		t.Fatalf("GetConnection 4 failed: %v", err)
	}
	
	if creationCount != 0 {
		t.Errorf("expected 0 new creations (should reuse), got %d", creationCount)
	}
	
	session.ReturnConnection(db4)
}
```

- [ ] **Step 2: Run test to verify it passes**

Run: `go test ./internal/session/... -v -run TestPoolSizeLimit`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/session/pool_test.go
git commit -m "test: add pool size limit test"
```

---

## Task 8: Implement Close Method

**Files:**
- Modify: `internal/session/pool.go`
- Test: `internal/session/pool_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestSession_Close(t *testing.T) {
	connID := uuid.New()
	session := NewSession(connID, &SessionConfig{MaxPoolSize: 3})
	
	connFactory := func(ctx context.Context) (*sql.DB, error) {
		return &sql.DB{}, nil
	}
	
	ctx := context.Background()
	
	// Get some connections
	db1, _ := session.GetConnection(ctx, connFactory)
	db2, _ := session.GetConnection(ctx, connFactory)
	session.ReturnConnection(db1)
	
	// Close the session
	session.Close()
	
	if !session.closed {
		t.Error("expected session to be closed")
	}
	
	// Trying to get connection should fail
	_, err := session.GetConnection(ctx, connFactory)
	if err == nil {
		t.Error("expected error when getting connection from closed session")
	}
	
	// Returning connection should not panic (just discard)
	session.ReturnConnection(db2) // Should not panic
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/session/... -v -run TestSession_Close`
Expected: FAIL with "undefined: Close"

- [ ] **Step 3: Write minimal implementation**

```go
// Close closes the session and all pooled connections.
// After Close, GetConnection will return an error.
// Returned connections are discarded (not pooled).
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.closed {
		return
	}
	
	s.closed = true
	
	// Close all available connections
	close(s.availableConns)
	for db := range s.availableConns {
		db.Close()
	}
	
	// Wake up any waiting goroutines
	close(s.waitQueue)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/session/... -v -run TestSession_Close`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/session/pool.go internal/session/pool_test.go
git commit -m "feat: implement Close method for session cleanup"
```

---

## Task 9: Integrate Pool with QueryExecutor

**Files:**
- Modify: `internal/query/executor.go:24` - Add Session field to QuerySession
- Modify: `internal/query/executor.go:108-170` - Use pool in ExecuteWithTimeout
- Modify: `internal/query/executor.go:258-270` - Close pool in CloseConnection

- [ ] **Step 1: Add Session to QuerySession struct**

```go
// In internal/query/executor.go

import (
	// ... existing imports
	"tablepro/internal/session"
)

// QuerySession represents an active query session for a connection.
// Contains the active query (if any), history, and database type.
type QuerySession struct {
	ConnectionID uuid.UUID
	ActiveQuery  *ActiveQuery
	History      []QueryHistoryEntry
	DBType       driver.DatabaseType
	Pool         *session.Session  // Add this line
}
```

- [ ] **Step 2: Initialize pool when creating QuerySession**

```go
// In ExecuteWithTimeout, around line 113:
session, exists := e.sessions[connID]
if !exists {
	session = &QuerySession{
		ConnectionID: connID,
		History:      make([]QueryHistoryEntry, 0),
		DBType:       driver.DatabaseTypeUnknown,
		Pool:         session.NewSession(connID, &session.SessionConfig{MaxPoolSize: 5}),
	}
	e.sessions[connID] = session
}
```

- [ ] **Step 3: Use pool in ExecuteWithTimeout**

```go
// Replace the direct driver.Query call with pool connection:
// Around line 134-140

// Get connection from pool
db, err := session.Pool.GetConnection(ctx, func(ctx context.Context) (*sql.DB, error) {
	// Connection factory - create new connection via ConnectionManager
	// For now, use driver to create connection
	password, _ := connection.GetPassword(connID)
	connConfig := &driver.ConnectionConfig{
		Host:     conn.Host,
		Port:     conn.Port,
		Database: conn.Database,
		Username: conn.Username,
		Password: password,
	}
	
	drv, err := driver.NewDriver(driver.TypeFromString(string(conn.Type)))
	if err != nil {
		return nil, err
	}
	
	if err := drv.Connect(ctx, connConfig); err != nil {
		return nil, err
	}
	
	// Return the underlying *sql.DB from the driver
	return drv.(*driver.BaseDriver).DB(), nil
})

if err != nil {
	return nil, fmt.Errorf("failed to acquire connection from pool: %w", err)
}

// Use db to execute query
row, err := db.QueryContext(queryCtx, query)
// ... rest of execution

// Return connection to pool when done
defer session.Pool.ReturnConnection(db)
```

- [ ] **Step 4: Update CloseConnection to close pool**

```go
// In CloseConnection (around line 258-270):
func (e *QueryExecutor) CloseConnection(connID uuid.UUID) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if session, exists := e.sessions[connID]; exists {
		if session.ActiveQuery != nil {
			session.ActiveQuery.Cancel()
		}
		// Close the connection pool
		if session.Pool != nil {
			session.Pool.Close()
		}
		delete(e.sessions, connID)
	}
}
```

- [ ] **Step 5: Run tests to verify integration**

Run: `go test ./internal/query/... -v`
Expected: Compilation errors initially (need to fix integration)

- [ ] **Step 6: Fix compilation errors**

Address any type mismatches or import issues.

- [ ] **Step 7: Commit**

```bash
git add internal/query/executor.go
git commit -m "feat: integrate connection pool with QueryExecutor"
```

---

## Task 10: Run Race Detector and Final Tests

**Files:**
- All session and query test files

- [ ] **Step 1: Run all session tests**

```bash
cd /Users/can/code/tablepro-fork/.worktrees/phase-06-sessions
go test ./internal/session/... -v
```

Expected: All tests pass

- [ ] **Step 2: Run race detector on session tests**

```bash
go test ./internal/session/... -race -v
```

Expected: No race conditions detected

- [ ] **Step 3: Run all query tests**

```bash
go test ./internal/query/... -v
```

Expected: All tests pass (may need mock adjustments for pool integration)

- [ ] **Step 4: Run race detector on query tests**

```bash
go test ./internal/query/... -race -v
```

Expected: No race conditions detected

- [ ] **Step 5: Run full test suite**

```bash
go test ./... -v
```

Expected: All tests pass

- [ ] **Step 6: Run full test suite with race detector**

```bash
go test ./... -race
```

Expected: No race conditions, all tests pass

- [ ] **Step 7: Final commit**

```bash
git add .
git commit -m "feat: complete connection pooling implementation with tests"
```

---

## Success Criteria Checklist

- [ ] `go test ./internal/session/...` passes including all pool tests
- [ ] `go test -race ./internal/session/...` detects no race conditions
- [ ] Pool correctly limits to MaxPoolSize
- [ ] Timeout works correctly (30 seconds default, not blocking forever)
- [ ] GoDoc comments on all exported functions
- [ ] `getConnectionFromPool()` returns available connection or creates new one up to max
- [ ] `returnConnectionToPool()` returns connection to available pool
- [ ] Pool exhaustion queuing with 30-second timeout implemented
- [ ] Unit tests cover all required scenarios:
  - TestGetConnectionFromPool_CreatesWhenEmpty
  - TestGetConnectionFromPool_ReusesExisting
  - TestGetConnectionFromPool_ExhaustionTimeout
  - TestReturnConnectionToPool
  - TestPoolSizeLimit

---

## Notes for Implementation

1. **Mock Strategy**: Since ConnectionManager integration is Task Group 8, use a connection factory function for testing. This allows testing pool logic without actual database connections.

2. **Thread Safety**: All pool operations must be protected by mutex. The channel operations provide their own synchronization, but size tracking needs mutex protection.

3. **Timeout Handling**: Use context.WithTimeout for the 30-second pool exhaustion timeout. This ensures clean cancellation and proper error propagation.

4. **Error Messages**: Include helpful context in error messages. "connection pool exhausted" should indicate it's a resource limitation, not a database error.

5. **Integration Coordination**: This implementation assumes QuerySession from Phase 5. If the structure differs, adapt the integration steps accordingly.

6. **GoDoc Comments**: Every exported function must have a GoDoc comment following Go conventions. This is mandatory for completion.
