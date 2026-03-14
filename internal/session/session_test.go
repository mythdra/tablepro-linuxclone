package session

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewSessionManager(t *testing.T) {
	t.Run("creates manager with default config", func(t *testing.T) {
		config := SessionConfig{
			MaxPoolSize:         5,
			HealthCheckInterval: 30 * time.Second,
			IdleTimeout:         5 * time.Minute,
		}
		manager := NewSessionManager(config)
		if manager == nil {
			t.Fatal("expected SessionManager, got nil")
		}
		if manager.config.MaxPoolSize != 5 {
			t.Errorf("expected MaxPoolSize=5, got %d", manager.config.MaxPoolSize)
		}
	})
	t.Run("creates manager with zero config uses defaults", func(t *testing.T) {
		manager := NewSessionManager(SessionConfig{})
		if manager == nil {
			t.Fatal("expected SessionManager, got nil")
		}
		if manager.config.MaxPoolSize != 5 {
			t.Errorf("expected default MaxPoolSize=5, got %d", manager.config.MaxPoolSize)
		}
	})
	t.Run("creates manager with DefaultSessionConfig", func(t *testing.T) {
		config := DefaultSessionConfig()
		manager := NewSessionManager(config)
		if manager == nil {
			t.Fatal("expected SessionManager, got nil")
		}
		if manager.config.MaxPoolSize != 5 {
			t.Errorf("expected MaxPoolSize=5, got %d", manager.config.MaxPoolSize)
		}
		if manager.config.HealthCheckInterval != 30*time.Second {
			t.Errorf("expected HealthCheckInterval=30s, got %v", manager.config.HealthCheckInterval)
		}
		if manager.config.IdleTimeout != 5*time.Minute {
			t.Errorf("expected IdleTimeout=5m, got %v", manager.config.IdleTimeout)
		}
	})
}

func TestCreateSession(t *testing.T) {
	t.Run("creates new session with active state", func(t *testing.T) {
		ctx := context.Background()
		manager := NewSessionManager(SessionConfig{})
		connectionID := uuid.New()
		session, err := manager.CreateSession(ctx, connectionID)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if session.ConnectionID != connectionID {
			t.Errorf("expected ConnectionID=%v, got %v", connectionID, session.ConnectionID)
		}
		if session.State != StateActive {
			t.Errorf("expected State=StateActive, got %v", session.State)
		}
		if len(session.Pool) != 1 {
			t.Errorf("expected Pool size=1, got %d", len(session.Pool))
		}
		if session.CreatedAt.IsZero() {
			t.Error("expected CreatedAt to be set")
		}
		if session.LastActiveAt.IsZero() {
			t.Error("expected LastActiveAt to be set")
		}
	})
	t.Run("session has unique ID", func(t *testing.T) {
		ctx := context.Background()
		manager := NewSessionManager(SessionConfig{})
		s1, _ := manager.CreateSession(ctx, uuid.New())
		s2, _ := manager.CreateSession(ctx, uuid.New())
		if s1.ID == s2.ID {
			t.Error("expected unique session IDs")
		}
	})
	t.Run("adds session to manager map", func(t *testing.T) {
		ctx := context.Background()
		manager := NewSessionManager(SessionConfig{})
		session, _ := manager.CreateSession(ctx, uuid.New())
		allSessions := manager.GetAllSessions()
		if len(allSessions) != 1 {
			t.Errorf("expected 1 session, got %d", len(allSessions))
		}
		if allSessions[0].ID != session.ID {
			t.Errorf("expected session ID %v, got %v", session.ID, allSessions[0].ID)
		}
	})
}

func TestGetSession(t *testing.T) {
	t.Run("returns error for non-existent session", func(t *testing.T) {
		manager := NewSessionManager(SessionConfig{})
		_, err := manager.GetSession(uuid.New())
		if err == nil || err.Error() != "session not found" {
			t.Errorf("expected 'session not found' error, got: %v", err)
		}
	})
	t.Run("returns session for existing ID", func(t *testing.T) {
		ctx := context.Background()
		manager := NewSessionManager(SessionConfig{})
		created, _ := manager.CreateSession(ctx, uuid.New())
		retrieved, err := manager.GetSession(created.ID)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if retrieved.ID != created.ID {
			t.Errorf("expected session ID %v, got %v", created.ID, retrieved.ID)
		}
	})
	t.Run("returns error for closed session", func(t *testing.T) {
		ctx := context.Background()
		manager := NewSessionManager(SessionConfig{})
		created, _ := manager.CreateSession(ctx, uuid.New())
		manager.CloseSession(created.ID)
		_, err := manager.GetSession(created.ID)
		if err == nil || err.Error() != "session not found" {
			t.Errorf("expected 'session not found' error, got: %v", err)
		}
	})
}

func TestCloseSession(t *testing.T) {
	t.Run("closes session and removes from map", func(t *testing.T) {
		ctx := context.Background()
		manager := NewSessionManager(SessionConfig{})
		session, _ := manager.CreateSession(ctx, uuid.New())
		manager.CloseSession(session.ID)
		if len(manager.GetAllSessions()) != 0 {
			t.Errorf("expected 0 sessions after close, got %d", len(manager.GetAllSessions()))
		}
	})
	t.Run("returns error for non-existent session", func(t *testing.T) {
		manager := NewSessionManager(SessionConfig{})
		err := manager.CloseSession(uuid.New())
		if err == nil {
			t.Error("expected error for non-existent session")
		}
	})
}

func TestGetAllSessions(t *testing.T) {
	t.Run("returns empty slice when no sessions", func(t *testing.T) {
		manager := NewSessionManager(SessionConfig{})
		sessions := manager.GetAllSessions()
		if sessions == nil || len(sessions) != 0 {
			t.Errorf("expected empty slice")
		}
	})
	t.Run("returns all active sessions", func(t *testing.T) {
		ctx := context.Background()
		manager := NewSessionManager(SessionConfig{})
		for i := 0; i < 3; i++ {
			manager.CreateSession(ctx, uuid.New())
		}
		if len(manager.GetAllSessions()) != 3 {
			t.Errorf("expected 3 sessions, got %d", len(manager.GetAllSessions()))
		}
	})
	t.Run("returns copy not reference", func(t *testing.T) {
		ctx := context.Background()
		manager := NewSessionManager(SessionConfig{})
		manager.CreateSession(ctx, uuid.New())
		s1 := manager.GetAllSessions()
		s2 := manager.GetAllSessions()
		if &s1[0] == &s2[0] {
			t.Error("expected GetAllSessions to return copy, not reference")
		}
	})
}

func TestGetStateString(t *testing.T) {
	t.Run("active", func(t *testing.T) {
		if getStateString(StateActive) != "active" {
			t.Error("StateActive should be 'active'")
		}
	})
	t.Run("idle", func(t *testing.T) {
		if getStateString(StateIdle) != "idle" {
			t.Error("StateIdle should be 'idle'")
		}
	})
	t.Run("closed", func(t *testing.T) {
		if getStateString(StateClosed) != "closed" {
			t.Error("StateClosed should be 'closed'")
		}
	})
}
