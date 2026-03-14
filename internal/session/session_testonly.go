//go:build test_session

// Package session provides database session management - test-only version
package session

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SessionState represents the current state of a database session.
type SessionState int

const (
	StateActive SessionState = iota
	StateIdle
	StateClosed
)

// SessionConfig holds configuration for session management.
type SessionConfig struct {
	MaxPoolSize         int
	HealthCheckInterval time.Duration
	IdleTimeout         time.Duration
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
	ID           uuid.UUID
	ConnectionID uuid.UUID
	State        SessionState
	Pool         []*sql.DB
	CreatedAt    time.Time
	LastActiveAt time.Time
	config       SessionConfig
	mu           sync.Mutex
}

// SessionManager manages database sessions.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*Session
	config   SessionConfig
	ctx      context.Context
}

// NewSessionManager creates a new session manager.
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

// CreateSession creates a new session.
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

// CloseSession closes a session.
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

// GetAllSessions returns all sessions.
func (sm *SessionManager) GetAllSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (sm *SessionManager) emitSessionEvent(ctx context.Context, eventType string, payload map[string]any) {
	// No-op for tests
}

func (sm *SessionManager) emitSessionCreated(ctx context.Context, session *Session) {
	sm.emitSessionEvent(ctx, "created", map[string]any{
		"sessionID":    session.ID.String(),
		"connectionID": session.ConnectionID.String(),
		"state":        getStateString(session.State),
	})
}

func (sm *SessionManager) emitSessionClosed(ctx context.Context, sessionID uuid.UUID, reason string) {
	sm.emitSessionEvent(ctx, "closed", map[string]any{
		"sessionID": sessionID.String(),
		"reason":    reason,
	})
}

func (sm *SessionManager) emitSessionError(ctx context.Context, sessionID uuid.UUID, err error, isRecoverable bool) {
	sm.emitSessionEvent(ctx, "error", map[string]any{
		"sessionID":     sessionID.String(),
		"error":         err.Error(),
		"isRecoverable": isRecoverable,
	})
}

func (sm *SessionManager) emitSessionReconnecting(ctx context.Context, sessionID uuid.UUID, retryCount int, nextRetryIn time.Duration) {
	sm.emitSessionEvent(ctx, "reconnecting", map[string]any{
		"sessionID":   sessionID.String(),
		"retryCount":  retryCount,
		"nextRetryIn": nextRetryIn.String(),
	})
}

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
