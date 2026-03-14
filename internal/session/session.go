//go:build !test_session

// Package session provides database session management with connection pooling,
// health checking, and automatic reconnection capabilities.
package session

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SessionState represents the current state of a database session.
type SessionState int

const (
	// StateActive indicates the session is actively being used.
	StateActive SessionState = iota
	// StateIdle indicates the session has no active queries.
	StateIdle
	// StateClosed indicates the session has been closed.
	StateClosed
)

// SessionConfig holds configuration for session management.
type SessionConfig struct {
	// MaxPoolSize is the maximum number of connections in the pool (default: 5).
	MaxPoolSize int
	// HealthCheckInterval is how often to check connection health (default: 30s).
	HealthCheckInterval time.Duration
	// IdleTimeout is how long a session can be idle before cleanup consideration (default: 5m).
	IdleTimeout time.Duration
}

// DefaultSessionConfig returns a SessionConfig with sensible defaults.
func DefaultSessionConfig() SessionConfig {
	return SessionConfig{
		MaxPoolSize:         5,
		HealthCheckInterval: 30 * time.Second,
		IdleTimeout:         5 * time.Minute,
	}
}

// Session represents an active database session with connection metadata.
type Session struct {
	// ID is the unique identifier for this session.
	ID uuid.UUID
	// ConnectionID is the ID of the connection this session belongs to.
	ConnectionID uuid.UUID
	// State is the current state of the session.
	State SessionState
	// Pool is the connection pool for this session.
	Pool []*sql.DB
	// CreatedAt is when the session was created.
	CreatedAt time.Time
	// LastActiveAt is when the session was last used.
	LastActiveAt time.Time
	config SessionConfig
	mu     sync.Mutex
}

// SessionManager manages database sessions with connection pooling and lifecycle management.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*Session
	config   SessionConfig
	ctx      context.Context
}

// NewSessionManager creates a new session manager with the given configuration.
func NewSessionManager(config SessionConfig) *SessionManager {
	if config.MaxPoolSize <= 0 {
		config.MaxPoolSize = 5
	}
	if config.HealthCheckInterval <= 0 {
		config.HealthCheckInterval = 30 * time.Second
	}
	if config.IdleTimeout <= 0 {
		config.IdleTimeout = 5 * time.Minute
	}

	return &SessionManager{
		sessions: make(map[uuid.UUID]*Session),
		config:   config,
		ctx:      context.Background(),
	}
}

// CreateSession creates a new session for the given connection ID.
// Initializes pool with 1 connection (stub - full implementation in Task Group 3).
func (sm *SessionManager) CreateSession(ctx context.Context, connectionID uuid.UUID) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := &Session{
		ID:           uuid.New(),
		ConnectionID: connectionID,
		State:        StateActive,
		Pool:         make([]*sql.DB, 0),
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
		config:       sm.config,
	}

	session.Pool = append(session.Pool, nil)
	sm.sessions[session.ID] = session

	sm.emitSessionCreated(ctx, session)
	return session, nil
}

// GetSession retrieves a session by ID.
// Returns an error if the session is not found or is closed.
func (sm *SessionManager) GetSession(sessionID uuid.UUID) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if session.State == StateClosed {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}

// CloseSession closes a session and removes it from the manager.
// All pooled connections are closed before removal.
func (sm *SessionManager) CloseSession(sessionID uuid.UUID) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	for _, conn := range session.Pool {
		if conn != nil {
			conn.Close()
		}
	}

	session.State = StateClosed
	delete(sm.sessions, sessionID)

	sm.emitSessionClosed(sm.ctx, sessionID, "user requested")
	return nil
}

// GetAllSessions returns a copy of all sessions managed by this SessionManager.
func (sm *SessionManager) GetAllSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// emitSessionEvent emits a session event to the Wails frontend.
func (sm *SessionManager) emitSessionEvent(ctx context.Context, eventType string, payload map[string]any) {
	runtime.EventsEmit(ctx, "session:"+eventType, payload)
}

// emitSessionCreated emits an event when a session is created.
func (sm *SessionManager) emitSessionCreated(ctx context.Context, session *Session) {
	sm.emitSessionEvent(ctx, "created", map[string]any{
		"sessionID":    session.ID.String(),
		"connectionID": session.ConnectionID.String(),
		"state":        getStateString(session.State),
	})
}

// emitSessionClosed emits an event when a session is closed.
func (sm *SessionManager) emitSessionClosed(ctx context.Context, sessionID uuid.UUID, reason string) {
	sm.emitSessionEvent(ctx, "closed", map[string]any{
		"sessionID": sessionID.String(),
		"reason":    reason,
	})
}

// emitSessionError emits an event when a session error occurs.
func (sm *SessionManager) emitSessionError(ctx context.Context, sessionID uuid.UUID, err error, isRecoverable bool) {
	sm.emitSessionEvent(ctx, "error", map[string]any{
		"sessionID":     sessionID.String(),
		"error":         err.Error(),
		"isRecoverable": isRecoverable,
	})
}

// emitSessionReconnecting emits an event when a session is reconnecting.
func (sm *SessionManager) emitSessionReconnecting(ctx context.Context, sessionID uuid.UUID, retryCount int, nextRetryIn time.Duration) {
	sm.emitSessionEvent(ctx, "reconnecting", map[string]any{
		"sessionID":   sessionID.String(),
		"retryCount":  retryCount,
		"nextRetryIn": nextRetryIn.String(),
	})
}

// getStateString returns the string representation of a SessionState.
func getStateString(state SessionState) string {
	switch state {
	case StateActive:
		return "active"
	case StateIdle:
		return "idle"
	case StateClosed:
		return "closed"
	default:
		return "unknown"
	}
}
