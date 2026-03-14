## 1. Session Manager Core Structure

- [ ] 1.1 Create `internal/session/` package directory
- [ ] 1.2 Define Session struct with fields: ID (uuid.UUID), ConnectionID (uuid.UUID), State (enum), Pool ([]*sql.DB), CreatedAt, LastActiveAt
- [ ] 1.3 Define SessionState enum constants: StateActive, StateIdle, StateClosed
- [ ] 1.4 Define SessionManager struct with fields: mu (sync.RWMutex), sessions (map[uuid.UUID]*Session), config (SessionConfig)
- [ ] 1.5 Define SessionConfig struct with fields: MaxPoolSize (int), HealthCheckInterval (time.Duration), IdleTimeout (time.Duration)
- [ ] 1.6 Create TypeScript types: Session, SessionState, SessionConfig in frontend/src/types/

## 2. Session Lifecycle Methods

- [ ] 2.1 Implement NewSessionManager(config SessionConfig) *SessionManager constructor
- [ ] 2.2 Implement CreateSession(ctx context.Context, connectionID uuid.UUID) (*Session, error) method
- [ ] 2.3 Implement GetSession(sessionID uuid.UUID) (*Session, error) method with read lock
- [ ] 2.4 Implement CloseSession(sessionID uuid.UUID) error method with proper cleanup
- [ ] 2.5 Implement GetAllSessions() []*Session method for frontend listing
- [ ] 2.6 Add session state transition logic (active → idle → closed)
- [ ] 2.7 Bind SessionManager methods to Wails in cmd/main.go

## 3. Connection Pooling Implementation

- [ ] 3.1 Implement getConnectionFromPool(session *Session) (*sql.DB, error) function
- [ ] 3.2 Implement returnConnectionToPool(session *Session, conn *sql.DB) function
- [ ] 3.3 Add pool size tracking (current size, available connections)
- [ ] 3.4 Implement pool exhaustion queuing with 30-second timeout
- [ ] 3.5 Add connection creation logic using driver from ConnectionManager
- [ ] 3.6 Write unit tests for pool expansion and connection reuse

## 4. Health Check System

- [ ] 4.1 Implement startHealthCheckWorker(ctx context.Context) background goroutine
- [ ] 4.2 Implement pingDatabase(conn *sql.DB) error with SELECT 1 query
- [ ] 4.3 Add skip logic for recently active sessions (< 30 seconds)
- [ ] 4.4 Implement stale connection detection and removal from pool
- [ ] 4.5 Add context-based shutdown for health check worker
- [ ] 4.6 Write unit tests for health check detection

## 5. Auto-Reconnect Logic

- [ ] 5.1 Implement reconnectSession(sessionID uuid.UUID) method
- [ ] 5.2 Add exponential backoff calculation (1s, 2s, 4s, 8s, 16s, 30s cap)
- [ ] 5.3 Implement max retry limit (5 attempts) with failure notification
- [ ] 5.4 Add SSH tunnel reconnection logic (close and recreate tunnel)
- [ ] 5.5 Handle reconnection during query execution (one immediate retry)
- [ ] 5.6 Write unit tests for backoff timing and retry limits

## 6. Session Event Emission

- [ ] 6.1 Create emitSessionEvent(eventType string, payload map[string]any) helper method
- [ ] 6.2 Implement emitSessionCreated(session *Session) event emitter
- [ ] 6.3 Implement emitSessionClosed(sessionID uuid.UUID, reason string) event emitter
- [ ] 6.4 Implement emitSessionError(sessionID uuid.UUID, err error, isRecoverable bool) event emitter
- [ ] 6.5 Implement emitSessionReconnecting(sessionID uuid.UUID, retryCount int, nextRetryIn time.Duration) event emitter
- [ ] 6.6 Integrate event calls into lifecycle methods (create, close, error, reconnect)

## 7. Frontend Event Integration

- [ ] 7.1 Create useSessionEvents() custom React hook
- [ ] 7.2 Add EventsOn listeners for session:created, session:closed, session:error, session:reconnecting
- [ ] 7.3 Implement EventsOff cleanup in useEffect return function
- [ ] 7.4 Create SessionStatusToast component for error/reconnecting notifications
- [ ] 7.5 Update ConnectionStatusIndicator component to listen to session events
- [ ] 7.6 Add session event payload types to TypeScript types file

## 8. App Lifecycle Integration

- [ ] 8.1 Initialize SessionManager in app startup() method
- [ ] 8.2 Implement graceful shutdown in app shutdown() method (close all sessions)
- [ ] 8.3 Add session cleanup for closed connections on app startup
- [ ] 8.4 Integrate SessionManager with existing ConnectionManager
- [ ] 8.5 Update Phase 5 query execution to use SessionManager.GetConnection() instead of direct driver access

## 9. Testing & Verification

- [ ] 9.1 Write unit tests for SessionManager lifecycle methods
- [ ] 9.2 Write unit tests for connection pool operations
- [ ] 9.3 Write integration tests for health check with real database
- [ ] 9.4 Write integration tests for auto-reconnect behavior
- [ ] 9.5 Write frontend component tests for session event handling
- [ ] 9.6 Run `go test -race ./internal/session/...` to detect race conditions
- [ ] 9.7 Run `go test -cover ./internal/session/...` and verify >80% coverage

## 10. Documentation & Cleanup

- [ ] 10.1 Add GoDoc comments to all exported functions and types
- [ ] 10.2 Update AGENTS.md with SessionManager usage examples
- [ ] 10.3 Add session management section to architecture docs
- [ ] 10.4 Create troubleshooting guide for common session issues
- [ ] 10.5 Verify all Phase 6 acceptance criteria are met
- [ ] 10.6 Run `go fmt ./...` and `go vet ./...` on new code
