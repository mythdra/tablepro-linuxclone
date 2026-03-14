package session

import (
	"context"
	"testing"
	"time"
	"github.com/google/uuid"
)

func TestNewSessionManager(t *testing.T) {
	config := DefaultSessionConfig()
	sm := NewSessionManager(config)
	if sm == nil {
		t.Fatal("SessionManager is nil")
	}
	if sm.config.MaxPoolSize != 5 {
		t.Errorf("Expected MaxPoolSize 5, got %d", sm.config.MaxPoolSize)
	}
	sm.Shutdown()
}

func TestCreateSession(t *testing.T) {
	sm := NewSessionManager(DefaultSessionConfig())
	defer sm.Shutdown()
	
	session, err := sm.CreateSession(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if session == nil {
		t.Fatal("Session is nil")
	}
	if session.State != StateActive {
		t.Errorf("Expected StateActive, got %v", session.State)
	}
}

func TestGetSession(t *testing.T) {
	sm := NewSessionManager(DefaultSessionConfig())
	defer sm.Shutdown()
	
	_, err := sm.GetSession(uuid.New())
	if err == nil {
		t.Fatal("Expected error for non-existent session")
	}
	
	session, _ := sm.CreateSession(context.Background(), uuid.New())
	retrieved, err := sm.GetSession(session.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if retrieved.ID != session.ID {
		t.Error("Retrieved session ID mismatch")
	}
}

func TestCloseSession(t *testing.T) {
	sm := NewSessionManager(DefaultSessionConfig())
	defer sm.Shutdown()
	
	session, _ := sm.CreateSession(context.Background(), uuid.New())
	
	err := sm.CloseSession(session.ID)
	if err != nil {
		t.Fatalf("CloseSession failed: %v", err)
	}
	
	_, err = sm.GetSession(session.ID)
	if err == nil {
		t.Fatal("Expected error after close")
	}
}

func TestGetAllSessions(t *testing.T) {
	sm := NewSessionManager(DefaultSessionConfig())
	defer sm.Shutdown()
	
	sessions := sm.GetAllSessions()
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions, got %d", len(sessions))
	}
	
	sm.CreateSession(context.Background(), uuid.New())
	sm.CreateSession(context.Background(), uuid.New())
	
	sessions = sm.GetAllSessions()
	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}
}

func TestIsRecentlyActive(t *testing.T) {
	session := &Session{LastActiveAt: time.Now()}
	if !session.IsRecentlyActive(time.Minute) {
		t.Error("Expected recently active")
	}
	
	session.LastActiveAt = time.Now().Add(-2 * time.Minute)
	if session.IsRecentlyActive(time.Minute) {
		t.Error("Expected not recently active")
	}
}

func TestCalculateBackoff(t *testing.T) {
	sm := NewSessionManager(DefaultSessionConfig())
	
	tests := []struct {
		count int
		want  time.Duration
	}{
		{0, time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{10, 30 * time.Second}, // capped
	}
	
	for _, tt := range tests {
		got := sm.calculateBackoff(tt.count)
		if got != tt.want {
			t.Errorf("calculateBackoff(%d) = %v, want %v", tt.count, got, tt.want)
		}
	}
}
