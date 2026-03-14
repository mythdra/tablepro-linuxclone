package session

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SessionState int

const (
	StateActive SessionState = iota
	StateIdle
	StateClosed
)

type SessionConfig struct {
	MaxPoolSize         int
	PoolTimeout         time.Duration
	HealthCheckInterval time.Duration
	IdleTimeout         time.Duration
	MaxRetries          int
	BackoffCap          time.Duration
}

func DefaultSessionConfig() SessionConfig {
	return SessionConfig{
		MaxPoolSize: 5, PoolTimeout: 30 * time.Second, HealthCheckInterval: 30 * time.Second,
		IdleTimeout: 5 * time.Minute, MaxRetries: 5, BackoffCap: 30 * time.Second,
	}
}

type Session struct {
	ID, ConnectionID uuid.UUID
	State            SessionState
	mu               sync.RWMutex
	availableConns   chan *sql.DB
	currentSize      int
	waitQueue        chan struct{}
	closed           bool
	CreatedAt        time.Time
	LastActiveAt     time.Time
	RetryCount       int
	UsesSSHTunnel    bool
	SSHTunnelID      uuid.UUID
}

type SessionManager struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*Session
	config   SessionConfig
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewSessionManager(config SessionConfig) *SessionManager {
	if config.MaxPoolSize <= 0 {
		config.MaxPoolSize = 5
	}
	if config.HealthCheckInterval <= 0 {
		config.HealthCheckInterval = 30 * time.Second
	}
	if config.PoolTimeout <= 0 {
		config.PoolTimeout = 30 * time.Second
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 5
	}
	if config.BackoffCap <= 0 {
		config.BackoffCap = 30 * time.Second
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &SessionManager{sessions: make(map[uuid.UUID]*Session), config: config, ctx: ctx, cancel: cancel}
}

func (sm *SessionManager) CreateSession(ctx context.Context, connectionID uuid.UUID) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	session := &Session{
		ID: uuid.New(), ConnectionID: connectionID, State: StateActive,
		availableConns: make(chan *sql.DB, sm.config.MaxPoolSize),
		waitQueue:      make(chan struct{}, sm.config.MaxPoolSize),
		CreatedAt:      time.Now(), LastActiveAt: time.Now(),
	}
	sm.sessions[session.ID] = session
	return session, nil
}

func (sm *SessionManager) GetSession(sessionID uuid.UUID) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	session.mu.RLock()
	closed := session.closed
	session.mu.RUnlock()
	if closed {
		return nil, fmt.Errorf("session is closed")
	}
	return session, nil
}

func (sm *SessionManager) CloseSession(sessionID uuid.UUID) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}
	session.mu.Lock()
	session.closed = true
	session.State = StateClosed
	session.mu.Unlock()
	close(session.availableConns)
	for db := range session.availableConns {
		if db != nil {
			db.Close()
		}
	}
	close(session.waitQueue)
	delete(sm.sessions, sessionID)
	return nil
}

func (sm *SessionManager) GetAllSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (sm *SessionManager) GetConnection(ctx context.Context, session *Session, connFactory func(context.Context) (*sql.DB, error)) (*sql.DB, error) {
	if session == nil {
		return nil, fmt.Errorf("session is nil")
	}
	session.mu.Lock()
	if session.closed {
		session.mu.Unlock()
		return nil, fmt.Errorf("session is closed")
	}
	select {
	case db := <-session.availableConns:
		session.mu.Unlock()
		session.LastActiveAt = time.Now()
		return db, nil
	default:
	}
	if session.currentSize < sm.config.MaxPoolSize {
		session.mu.Unlock()
		db, err := connFactory(ctx)
		if err != nil {
			return nil, err
		}
		session.mu.Lock()
		if !session.closed {
			session.currentSize++
			session.LastActiveAt = time.Now()
		}
		session.mu.Unlock()
		return db, nil
	}
	session.mu.Unlock()
	timeoutCtx, cancel := context.WithTimeout(ctx, sm.config.PoolTimeout)
	defer cancel()
	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("connection pool exhausted")
		case <-session.waitQueue:
			session.mu.Lock()
			select {
			case db := <-session.availableConns:
				session.mu.Unlock()
				session.LastActiveAt = time.Now()
				return db, nil
			default:
				session.mu.Unlock()
			}
		}
	}
}

func (sm *SessionManager) ReturnConnection(session *Session, db *sql.DB) {
	if db == nil || session == nil {
		return
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.closed {
		db.Close()
		return
	}
	select {
	case session.availableConns <- db:
		select {
		case session.waitQueue <- struct{}{}:
		default:
		}
	default:
		db.Close()
	}
}

func (sm *SessionManager) Shutdown() {
	sm.cancel()
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for _, s := range sm.sessions {
		if s != nil {
			s.mu.Lock()
			s.closed = true
			s.mu.Unlock()
			close(s.availableConns)
			for db := range s.availableConns {
				if db != nil {
					db.Close()
				}
			}
			close(s.waitQueue)
		}
	}
}

func (sm *SessionManager) calculateBackoff(n int) time.Duration {
	b := time.Duration(1<<uint(n)) * time.Second
	if b > sm.config.BackoffCap {
		return sm.config.BackoffCap
	}
	return b
}

func (s *Session) IsRecentlyActive(d time.Duration) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Since(s.LastActiveAt) < d
}

func (s *Session) UpdateLastActive() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastActiveAt = time.Now()
}

func (s *Session) GetLastActive() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.LastActiveAt
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
