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

type SessionManager struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*Session
	config   SessionConfig
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewSessionManager(config SessionConfig) *SessionManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &SessionManager{sessions: make(map[uuid.UUID]*Session), config: config, ctx: ctx, cancel: cancel}
}

func (sm *SessionManager) AddSession(s *Session) { sm.mu.Lock(); defer sm.mu.Unlock(); sm.sessions[s.ID] = s }
func (sm *SessionManager) GetSession(id uuid.UUID) (*Session, bool) { sm.mu.RLock(); defer sm.mu.RUnlock(); s, ok := sm.sessions[id]; return s, ok }
func (sm *SessionManager) RemoveSession(id uuid.UUID) { sm.mu.Lock(); defer sm.mu.Unlock(); delete(sm.sessions, id) }
func (sm *SessionManager) GetAllSessions() []*Session { sm.mu.RLock(); defer sm.mu.RUnlock(); r := make([]*Session, 0, len(sm.sessions)); for _, s := range sm.sessions { r = append(r, s) }; return r }
func (sm *SessionManager) Shutdown() { sm.cancel(); sm.mu.Lock(); defer sm.mu.Unlock(); for _, s := range sm.sessions { if s != nil { s.Close() } } }

func (sm *SessionManager) StartHealthCheckWorker() {
	go func() {
		ticker := time.NewTicker(sm.config.HealthCheckInterval)
		defer ticker.Stop()
		for { select { case <-sm.ctx.Done(): return; case <-ticker.C: sm.checkAllSessions() } }
	}()
}

func pingDatabase(db *sql.DB) error {
	if db == nil { return errors.New("database connection is nil") }
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.ExecContext(ctx, "SELECT 1")
	if err != nil { return fmt.Errorf("ping failed: %w", err) }
	return nil
}

func checkSessionHealth(s *Session, interval time.Duration, db *sql.DB) bool {
	if s.IsRecentlyActive(interval) { return true }
	if err := pingDatabase(db); err != nil { return false }
	s.UpdateLastActive()
	return true
}

func (sm *SessionManager) checkAllSessions() {
	for _, s := range sm.GetAllSessions() {
		if !s.IsRecentlyActive(sm.config.HealthCheckInterval) { sm.handleUnhealthySession(s) }
	}
}

func (sm *SessionManager) handleUnhealthySession(s *Session) { go sm.reconnectSession(s.ID) }

func (sm *SessionManager) calculateBackoff(n int) time.Duration {
	b := time.Duration(1<<uint(n)) * time.Second
	if b > sm.config.BackoffCap { return sm.config.BackoffCap }
	return b
}

func (sm *SessionManager) waitForBackoff(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d); defer t.Stop()
	select { case <-ctx.Done(): return false; case <-t.C: return true }
}

func (sm *SessionManager) reconnectSession(id uuid.UUID) {
	s, ok := sm.GetSession(id)
	if !ok { return }
	for s.RetryCount < sm.config.MaxRetries {
		b := sm.calculateBackoff(s.RetryCount)
		if !sm.waitForBackoff(sm.ctx, b) { return }
		if sm.establishNewConnection(s) == nil {
			s.mu.Lock(); s.RetryCount = 0; s.mu.Unlock()
			return
		}
		s.mu.Lock(); s.RetryCount++; s.mu.Unlock()
	}
	sm.closeSession(id)
}

func (sm *SessionManager) establishNewConnection(s *Session) error { return errors.New("connection establishment not yet implemented") }
func (sm *SessionManager) closeSession(id uuid.UUID) { s, ok := sm.GetSession(id); if !ok { return }; s.Close(); sm.RemoveSession(id) }
