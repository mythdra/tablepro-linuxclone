package session

import (
	"context"
	"testing"
	"time"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPingDatabase_Success(t *testing.T) {
	db, mock, err := sqlmock.New(); require.NoError(t, err); defer db.Close()
	mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(0, 1))
	assert.NoError(t, pingDatabase(db))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPingDatabase_Failure(t *testing.T) {
	db, mock, err := sqlmock.New(); require.NoError(t, err); defer db.Close()
	mock.ExpectExec("SELECT 1").WillReturnError(sqlmock.ErrCancelled)
	err = pingDatabase(db)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ping failed")
}

func TestPingDatabase_NilConnection(t *testing.T) {
	err := pingDatabase(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is nil")
}

func TestCheckSessionHealth_SkipsRecent(t *testing.T) {
	s := &Session{ID: uuid.New(), LastActiveAt: time.Now()}
	assert.True(t, checkSessionHealth(s, 30*time.Second, nil))
}

func TestCheckSessionHealth_PingsOld(t *testing.T) {
	db, mock, err := sqlmock.New(); require.NoError(t, err); defer db.Close()
	mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(0, 1))
	s := &Session{ID: uuid.New(), LastActiveAt: time.Now().Add(-60 * time.Second)}
	assert.True(t, checkSessionHealth(s, 30*time.Second, db))
}

func TestCheckSessionHealth_Failure(t *testing.T) {
	db, mock, err := sqlmock.New(); require.NoError(t, err); defer db.Close()
	mock.ExpectExec("SELECT 1").WillReturnError(sqlmock.ErrCancelled)
	s := &Session{ID: uuid.New(), LastActiveAt: time.Now().Add(-60 * time.Second)}
	assert.False(t, checkSessionHealth(s, 30*time.Second, db))
}

func TestCalculateBackoff(t *testing.T) {
	sm := NewSessionManager(DefaultSessionConfig())
	tests := []struct{ n int; e time.Duration }{
		{0, 1 * time.Second}, {1, 2 * time.Second}, {2, 4 * time.Second},
		{3, 8 * time.Second}, {4, 16 * time.Second}, {5, 30 * time.Second}, {6, 30 * time.Second},
	}
	for _, tt := range tests {
		t.Run(string(rune('0'+tt.n)), func(t *testing.T) {
			assert.Equal(t, tt.e, sm.calculateBackoff(tt.n))
		})
	}
}

func TestCalculateBackoff_CustomCap(t *testing.T) {
	sm := NewSessionManager(SessionConfig{BackoffCap: 10 * time.Second, MaxRetries: 5})
	assert.Equal(t, 10*time.Second, sm.calculateBackoff(4))
}

func TestWaitForBackoff_Complete(t *testing.T) {
	sm := NewSessionManager(DefaultSessionConfig())
	assert.True(t, sm.waitForBackoff(context.Background(), 100*time.Millisecond))
}

func TestWaitForBackoff_Cancelled(t *testing.T) {
	sm := NewSessionManager(DefaultSessionConfig())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.False(t, sm.waitForBackoff(ctx, 1*time.Second))
}

func TestReconnectSession_MaxRetries(t *testing.T) {
	sm := NewSessionManager(SessionConfig{MaxRetries: 3, BackoffCap: 50 * time.Millisecond, HealthCheckInterval: 30 * time.Second})
	s := NewSession(uuid.New(), nil)
	s.RetryCount = 0
	sm.AddSession(s)
	go sm.reconnectSession(s.ID)
	time.Sleep(250 * time.Millisecond)
	_, ok := sm.GetSession(s.ID)
	assert.False(t, ok)
}

func TestSession_IsRecentlyActive(t *testing.T) {
	s := &Session{LastActiveAt: time.Now()}
	assert.True(t, s.IsRecentlyActive(30*time.Second))
	s.mu.Lock()
	s.LastActiveAt = time.Now().Add(-60 * time.Second)
	s.mu.Unlock()
	assert.False(t, s.IsRecentlyActive(30*time.Second))
}

func TestSession_UpdateLastActive(t *testing.T) {
	s := &Session{LastActiveAt: time.Now().Add(-1 * time.Hour)}
	before := s.GetLastActive()
	s.UpdateLastActive()
	after := s.GetLastActive()
	assert.True(t, after.After(before))
	assert.True(t, time.Since(after) < time.Second)
}

func TestSessionManager_GetAllSessions(t *testing.T) {
	sm := NewSessionManager(DefaultSessionConfig())
	sm.AddSession(NewSession(uuid.New(), nil))
	sm.AddSession(NewSession(uuid.New(), nil))
	assert.Len(t, sm.GetAllSessions(), 2)
}

func TestSessionManager_RemoveSession(t *testing.T) {
	sm := NewSessionManager(DefaultSessionConfig())
	id := uuid.New()
	sm.AddSession(&Session{ID: id})
	_, ok := sm.GetSession(id)
	assert.True(t, ok)
	sm.RemoveSession(id)
	_, ok = sm.GetSession(id)
	assert.False(t, ok)
}

func TestSessionManager_Shutdown(t *testing.T) {
	sm := NewSessionManager(DefaultSessionConfig())
	sm.AddSession(NewSession(uuid.New(), nil))
	sm.Shutdown()
	<-sm.ctx.Done()
	assert.Equal(t, context.Canceled, sm.ctx.Err())
}
