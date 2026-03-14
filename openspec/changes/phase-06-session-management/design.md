## Context

TablePro currently manages database connections through ConnectionManager (Phase 3), but lacks proper session lifecycle management. Each query execution creates ad-hoc connections without pooling, health monitoring, or automatic reconnection. This leads to:

- **Connection leaks**: Sessions not properly closed on app shutdown
- **Stale connections**: No detection of dropped database connections
- **Poor resource usage**: No connection reuse between queries
- **No session visibility**: Frontend cannot display session status

Phase 5 (Query Execution) established query execution patterns but did not include session abstraction. Phase 6 adds the SessionManager layer between ConnectionManager and query execution.

**Constraints**:
- Must integrate with existing ConnectionManager (Phase 3)
- Must support all 8 database drivers (Phase 4)
- Must work with Wails event system (Phase 2)
- Zero external dependencies beyond Go stdlib

## Goals / Non-Goals

**Goals:**
- Session lifecycle management (create, get, close) with proper cleanup
- Connection pooling per session with configurable size (default: 5 connections)
- Health check system with periodic ping (default: 30 seconds)
- Auto-reconnect on connection failure with exponential backoff
- Real-time session events emitted to frontend
- Session state tracking (active, idle, closed)
- Graceful shutdown on app exit

**Non-Goals:**
- Session persistence across app restarts (handled by tab persistence spec)
- Query cancellation (Phase 5 already handles context-based cancellation)
- Load balancing across multiple database servers
- Connection pooling across different users/sessions

## Decisions

### 1. SessionManager as Singleton Service

**Decision**: SessionManager is a singleton service bound to Wails app lifecycle.

**Rationale**: 
- Single point of control for all database sessions
- Consistent with ConnectionManager pattern (Phase 3)
- Simplifies shutdown/cleanup logic

**Alternatives Considered**:
- Per-tab session managers → Too much overhead, complex coordination
- Session factory pattern → Unnecessary abstraction for single-user app

### 2. Connection Pool Per Session (Not Global)

**Decision**: Each session maintains its own connection pool.

**Rationale**:
- Sessions map 1:1 with database connections in most cases
- Simpler to track ownership and cleanup
- Avoids cross-session connection contamination

**Alternatives Considered**:
- Global connection pool shared across sessions → Complex bookkeeping, potential security issues
- No pooling (create per query) → Poor performance, resource waste

### 3. Health Check via Periodic Ping

**Decision**: Background goroutine pings each session every 30 seconds.

**Rationale**:
- Simple, database-agnostic approach
- Detects network issues, server timeouts, SSH tunnel failures
- Uses standard SQL `SELECT 1` ping query

**Alternatives Considered**:
- Health check on-demand (before each query) → Adds latency to every query
- Event-driven health checks → Doesn't detect idle connection drops

### 4. Auto-Reconnect with Exponential Backoff

**Decision**: Failed health checks trigger reconnect with backoff (1s, 2s, 4s, 8s, max 30s).

**Rationale**:
- Prevents reconnection storms on network issues
- Matches industry standard patterns
- User retains control (can manually close session)

**Alternatives Considered**:
- Immediate reconnect → Overwhelms network/database during outages
- No auto-reconnect → Forces manual reconnection for transient issues

### 5. Session Events via Wails EventsOn/EventsEmit

**Decision**: Use Wails event system for session status updates.

**Rationale**:
- Consistent with Phase 2 event patterns
- Decoupled communication (frontend can listen/unlisten)
- Supports multiple listeners (UI, logging, analytics)

**Event Names**:
- `session:created` - Session successfully created
- `session:closed` - Session closed (user or system)
- `session:error` - Session encountered error (includes error message)
- `session:reconnecting` - Attempting to reconnect

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| **Connection pool exhaustion** | Configurable pool size, queue requests when exhausted, emit events when pool low |
| **Health check overhead** | Configurable interval, skip ping for active sessions (recent query = healthy) |
| **Auto-reconnect loops** | Exponential backoff, max retry count, user notification after 3 failures |
| **Goroutine leaks** | Context-based cancellation, defer cleanup on SessionManager shutdown |
| **Race conditions in pool** | sync.RWMutex for pool access, test with `-race` flag |
| **SSH tunnel reconnection complexity** | Close and recreate entire tunnel, not just DB connection |

## Migration Plan

Phase 6 is a greenfield implementation with no migration required:

1. Create `internal/session/` package with SessionManager
2. Implement session lifecycle methods (Create, Get, Close)
3. Add connection pooling logic
4. Implement health check goroutine
5. Bind SessionManager to Wails app
6. Update Phase 5 query execution to use SessionManager
7. Add frontend event listeners for session status

**Rollback**: Not applicable - Phase 6 is additive. If issues arise, disable health checks or auto-reconnect via config flags.

## Open Questions

1. **Pool size default**: 5 connections per session - should this be user-configurable in settings?
2. **Health check interval**: 30 seconds reasonable, or should it vary by database type?
3. **Session timeout**: Should idle sessions auto-close after N minutes? If so, what's the default?
