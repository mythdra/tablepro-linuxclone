package main

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"tablepro/internal/connection"
	"tablepro/internal/driver"
	"tablepro/internal/query"
	"tablepro/internal/session"
)

// App struct
type App struct {
	ctx            context.Context
	connectionMgr  *connection.ConnectionManager
	sessionManager *session.SessionManager
	queryExecutor  *query.QueryExecutor
	sessionMap     map[uuid.UUID]*session.Session
	sessionMapMu   sync.RWMutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	connMgr, err := connection.NewConnectionManager()
	if err != nil {
		slog.Error("Failed to initialize connection manager", "error", err)
	}

	sessionConfig := session.DefaultSessionConfig()
	sessionMgr := session.NewSessionManager(sessionConfig)
	sessionMgr.SetConnectionManager(connMgr)

	queryExec := query.NewQueryExecutor()

	return &App{
		connectionMgr:  connMgr,
		sessionManager: sessionMgr,
		queryExecutor:  queryExec,
		sessionMap:     make(map[uuid.UUID]*session.Session),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	slog.Info("TablePro starting...")

	sessionConfig := session.DefaultSessionConfig()
	a.sessionManager = session.NewSessionManager(sessionConfig)
	a.sessionManager.SetConnectionManager(a.connectionMgr)

	go a.sessionManager.StartHealthCheckWorker(ctx)

	if err := a.sessionManager.CleanupOrphanedSessions(ctx); err != nil {
		slog.Error("Failed to cleanup orphaned sessions", "error", err)
	}

	slog.InfoContext(ctx, "App startup complete")
}

// shutdown is called when the app stops
func (a *App) shutdown(ctx context.Context) {
	slog.InfoContext(ctx, "TablePro shutting down...")

	if a.sessionManager != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := a.sessionManager.Shutdown(shutdownCtx); err != nil {
			slog.Error("Failed to shutdown session manager", "error", err)
		}
	}

	slog.Info("Cleanup complete")
}

// getOrCreateSession gets an existing session for a connection or creates a new one
func (a *App) getOrCreateSession(ctx context.Context, connID uuid.UUID) (*session.Session, error) {
	a.sessionMapMu.RLock()
	sess, exists := a.sessionMap[connID]
	a.sessionMapMu.RUnlock()

	if exists {
		return sess, nil
	}

	a.sessionMapMu.Lock()
	defer a.sessionMapMu.Unlock()

	if sess, exists = a.sessionMap[connID]; exists {
		return sess, nil
	}

	sess, err := a.sessionManager.CreateSession(ctx, connID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	a.sessionMap[connID] = sess
	return sess, nil
}

// executeQueryWithSession executes a query using a session's connection pool
func (a *App) executeQueryWithSession(ctx context.Context, sess *session.Session, connID uuid.UUID, drv driver.DatabaseDriver, queryStr string) (*query.QueryResult, error) {
	db, err := a.sessionManager.GetConnectionFromPool(ctx, sess)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection from pool: %w", err)
	}

	defer func() {
		if db != nil {
			a.sessionManager.ReturnConnectionToPool(sess, db)
		}
	}()

	result, err := drv.Query(ctx, queryStr)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	var resultSet *query.ResultSet
	if result != nil {
		resultSet = query.NewResultSetFromRows([]*driver.Row{result}, 0, queryStr, drv.Type())
	}

	return &query.QueryResult{
		ResultSet: resultSet,
		QueryID:   uuid.New(),
		Duration:  0,
	}, nil
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return "0.1.0-dev"
}

// ==================== Connection Management RPC Methods ====================

// TestConnection tests a database connection and returns detailed results
// Uses 10-second timeout for the entire operation
func (a *App) TestConnection(ctx context.Context, conn *connection.DatabaseConnection) (*connection.TestConnectionResult, error) {
	if a.connectionMgr == nil {
		return &connection.TestConnectionResult{
			Success: false,
			Message: "Connection manager not initialized",
		}, nil
	}

	result, err := a.connectionMgr.TestConnection(ctx, conn)
	if err != nil {
		return &connection.TestConnectionResult{
			Success: false,
			Message: fmt.Sprintf("Error testing connection: %v", err),
		}, err
	}

	return result, nil
}

// ListConnections returns all saved connections
func (a *App) ListConnections(ctx context.Context) ([]*connection.DatabaseConnection, error) {
	if a.connectionMgr == nil {
		return nil, fmt.Errorf("connection manager not initialized")
	}
	return a.connectionMgr.List(), nil
}

// GetConnection returns a connection by ID
func (a *App) GetConnection(ctx context.Context, id string) (*connection.DatabaseConnection, error) {
	if a.connectionMgr == nil {
		return nil, fmt.Errorf("connection manager not initialized")
	}

	// Parse UUID
	connID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid connection ID: %w", err)
	}

	conn, exists := a.connectionMgr.Get(connID)
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", id)
	}
	return conn, nil
}

// SaveConnection saves a new or updates an existing connection
func (a *App) SaveConnection(ctx context.Context, conn *connection.DatabaseConnection) error {
	if a.connectionMgr == nil {
		return fmt.Errorf("connection manager not initialized")
	}

	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	return a.connectionMgr.Save(conn)
}

// DeleteConnection removes a connection by ID
func (a *App) DeleteConnection(ctx context.Context, id string) error {
	if a.connectionMgr == nil {
		return fmt.Errorf("connection manager not initialized")
	}

	connID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid connection ID: %w", err)
	}

	return a.connectionMgr.Delete(connID)
}

// DuplicateConnection creates a copy of an existing connection
func (a *App) DuplicateConnection(ctx context.Context, id string) (*connection.DatabaseConnection, error) {
	if a.connectionMgr == nil {
		return nil, fmt.Errorf("connection manager not initialized")
	}

	connID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid connection ID: %w", err)
	}

	return a.connectionMgr.Duplicate(connID)
}

// ValidateConnection validates a connection configuration
func (a *App) ValidateConnection(ctx context.Context, conn *connection.DatabaseConnection) error {
	if a.connectionMgr == nil {
		return fmt.Errorf("connection manager not initialized")
	}

	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	return a.connectionMgr.Validate(conn)
}

// SavePassword saves a password to the keychain
func (a *App) SavePassword(ctx context.Context, connectionID, password string) error {
	if a.connectionMgr == nil {
		return fmt.Errorf("connection manager not initialized")
	}

	connID, err := uuid.Parse(connectionID)
	if err != nil {
		return fmt.Errorf("invalid connection ID: %w", err)
	}

	return connection.SavePassword(connID, password)
}

// GetPassword retrieves a password from the keychain
func (a *App) GetPassword(ctx context.Context, connectionID string) (string, error) {
	if a.connectionMgr == nil {
		return "", fmt.Errorf("connection manager not initialized")
	}

	connID, err := uuid.Parse(connectionID)
	if err != nil {
		return "", fmt.Errorf("invalid connection ID: %w", err)
	}

	return connection.GetPassword(connID)
}

// DeletePassword removes a password from the keychain
func (a *App) DeletePassword(ctx context.Context, connectionID string) error {
	if a.connectionMgr == nil {
		return fmt.Errorf("connection manager not initialized")
	}

	connID, err := uuid.Parse(connectionID)
	if err != nil {
		return fmt.Errorf("invalid connection ID: %w", err)
	}

	return connection.DeletePassword(connID)
}

// ==================== Query Execution RPC Methods ====================

// ExecuteQuery executes a SQL query and returns the result.
// Uses default timeout (30 seconds) for query execution.
func (a *App) ExecuteQuery(ctx context.Context, connectionID, queryStr string) (*query.QueryResult, error) {
	if a.queryExecutor == nil {
		return nil, fmt.Errorf("query executor not initialized")
	}

	if a.connectionMgr == nil {
		return nil, fmt.Errorf("connection manager not initialized")
	}

	connID, err := uuid.Parse(connectionID)
	if err != nil {
		return nil, fmt.Errorf("invalid connection ID: %w", err)
	}

	_, err = a.getOrCreateSession(ctx, connID)
	if err != nil {
		return nil, err
	}

	conn, exists := a.connectionMgr.Get(connID)
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	password, _ := connection.GetPassword(connID)
	drv, err := driver.NewDriver(driver.TypeFromString(string(conn.Type)))
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}
	defer drv.Close()

	connConfig := &driver.ConnectionConfig{
		Host:     conn.Host,
		Port:     conn.Port,
		Database: conn.Database,
		Username: conn.Username,
		Password: password,
	}
	if err := drv.Connect(ctx, connConfig); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return a.queryExecutor.Execute(ctx, connID, drv, queryStr)
}

// ExecuteQueryWithTimeout executes a SQL query with a custom timeout.
// Timeout is specified in seconds.
func (a *App) ExecuteQueryWithTimeout(ctx context.Context, connectionID, queryStr string, timeoutSeconds int) (*query.QueryResult, error) {
	if a.queryExecutor == nil {
		return nil, fmt.Errorf("query executor not initialized")
	}

	if a.connectionMgr == nil {
		return nil, fmt.Errorf("connection manager not initialized")
	}

	connID, err := uuid.Parse(connectionID)
	if err != nil {
		return nil, fmt.Errorf("invalid connection ID: %w", err)
	}

	conn, exists := a.connectionMgr.Get(connID)
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	password, _ := connection.GetPassword(connID)
	drv, err := driver.NewDriver(driver.TypeFromString(string(conn.Type)))
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	connConfig := &driver.ConnectionConfig{
		Host:     conn.Host,
		Port:     conn.Port,
		Database: conn.Database,
		Username: conn.Username,
		Password: password,
	}
	if err := drv.Connect(ctx, connConfig); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer drv.Close()

	timeout := time.Duration(timeoutSeconds) * time.Second
	return a.queryExecutor.ExecuteWithTimeout(ctx, connID, drv, queryStr, timeout)
}

// ExecuteMultiStatement executes multiple SQL statements separated by semicolons.
func (a *App) ExecuteMultiStatement(ctx context.Context, connectionID, queryStr string) (*query.MultiStatementResult, error) {
	if a.queryExecutor == nil {
		return nil, fmt.Errorf("query executor not initialized")
	}

	if a.connectionMgr == nil {
		return nil, fmt.Errorf("connection manager not initialized")
	}

	connID, err := uuid.Parse(connectionID)
	if err != nil {
		return nil, fmt.Errorf("invalid connection ID: %w", err)
	}

	conn, exists := a.connectionMgr.Get(connID)
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	password, _ := connection.GetPassword(connID)
	drv, err := driver.NewDriver(driver.TypeFromString(string(conn.Type)))
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	connConfig := &driver.ConnectionConfig{
		Host:     conn.Host,
		Port:     conn.Port,
		Database: conn.Database,
		Username: conn.Username,
		Password: password,
	}
	if err := drv.Connect(ctx, connConfig); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer drv.Close()

	return a.queryExecutor.ExecuteStatements(ctx, connID, drv, queryStr)
}

// CancelQuery cancels a running query by query ID.
func (a *App) CancelQuery(ctx context.Context, queryID string) error {
	if a.queryExecutor == nil {
		return fmt.Errorf("query executor not initialized")
	}

	qID, err := uuid.Parse(queryID)
	if err != nil {
		return fmt.Errorf("invalid query ID: %w", err)
	}

	return a.queryExecutor.Cancel(qID)
}

// GetQueryHistory returns the query history for a connection.
// Limit specifies the maximum number of entries to return (0 = all).
func (a *App) GetQueryHistory(ctx context.Context, connectionID string, limit int) ([]query.QueryHistoryEntry, error) {
	if a.queryExecutor == nil {
		return nil, fmt.Errorf("query executor not initialized")
	}

	connID, err := uuid.Parse(connectionID)
	if err != nil {
		return nil, fmt.Errorf("invalid connection ID: %w", err)
	}

	return a.queryExecutor.GetHistory(connID, limit), nil
}

// ClearQueryHistory clears the query history for a connection.
func (a *App) ClearQueryHistory(ctx context.Context, connectionID string) error {
	if a.queryExecutor == nil {
		return fmt.Errorf("query executor not initialized")
	}

	connID, err := uuid.Parse(connectionID)
	if err != nil {
		return fmt.Errorf("invalid connection ID: %w", err)
	}

	a.queryExecutor.ClearHistory(connID)
	return nil
}
