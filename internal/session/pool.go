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
	ConnectionID   uuid.UUID
	MaxPoolSize    int
	PoolTimeout    time.Duration
	mu             sync.RWMutex
	availableConns chan *sql.DB
	currentSize    int
	waitQueue      chan struct{}
	closed         bool
	ID             uuid.UUID
	CreatedAt      time.Time
	LastActiveAt   time.Time
	RetryCount     int
	MaxRetries     int
	UsesSSHTunnel  bool
	SSHTunnelID    uuid.UUID
}

const DefaultPoolSize = 5
const DefaultPoolTimeout = 30 * time.Second
var ErrPoolExhausted = errors.New("connection pool exhausted")

func NewSession(connID uuid.UUID, config *SessionConfig) *Session {
	maxSize, timeout := DefaultPoolSize, DefaultPoolTimeout
	if config != nil {
		if config.MaxPoolSize > 0 { maxSize = config.MaxPoolSize }
		if config.PoolTimeout > 0 { timeout = config.PoolTimeout }
	}
	return &Session{
		ID: uuid.New(), ConnectionID: connID, MaxPoolSize: maxSize, PoolTimeout: timeout,
		availableConns: make(chan *sql.DB, maxSize), CreatedAt: time.Now(), LastActiveAt: time.Now(),
		waitQueue: make(chan struct{}, 1),
	}
}

func (s *Session) GetConnection(ctx context.Context, connFactory func(context.Context) (*sql.DB, error)) (*sql.DB, error) {
	s.mu.Lock()
	if s.closed { s.mu.Unlock(); return nil, fmt.Errorf("session is closed") }
	select { case db := <-s.availableConns: s.mu.Unlock(); return db, nil; default: }
	if s.currentSize < s.MaxPoolSize {
		s.mu.Unlock()
		db, err := connFactory(ctx)
		if err != nil { return nil, err }
		s.mu.Lock()
		if !s.closed {
			select { case s.availableConns <- db: s.currentSize++; default: }
		}
		s.mu.Unlock()
		return db, nil
	}
	s.mu.Unlock()
	timeoutCtx, cancel := context.WithTimeout(ctx, s.PoolTimeout)
	defer cancel()
	for {
		select {
		case <-timeoutCtx.Done(): return nil, ErrPoolExhausted
		case <-s.waitQueue:
			s.mu.Lock()
			select { case db := <-s.availableConns: s.mu.Unlock(); return db, nil; default: s.mu.Unlock() }
		}
	}
}

func (s *Session) ReturnConnection(db *sql.DB) {
	if db == nil || s.closed { return }
	s.mu.Lock()
	defer s.mu.Unlock()
	select { case s.availableConns <- db: select { case s.waitQueue <- struct{}{}: default: }; default: db.Close() }
}

func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed { return }
	s.closed = true
	close(s.availableConns)
	for db := range s.availableConns { db.Close() }
	close(s.waitQueue)
}

func (s *Session) UpdateLastActive() { s.mu.Lock(); defer s.mu.Unlock(); s.LastActiveAt = time.Now() }
func (s *Session) GetLastActive() time.Time { s.mu.RLock(); defer s.mu.RUnlock(); return s.LastActiveAt }
func (s *Session) IsRecentlyActive(d time.Duration) bool { s.mu.RLock(); defer s.mu.RUnlock(); return time.Since(s.LastActiveAt) < d }
