## Why

Phase 6 introduces session management to handle active database connections with connection pooling, health monitoring, and automatic reconnection. This solves the problem of managing long-lived database sessions efficiently, preventing connection leaks, detecting stale connections, and providing real-time session status updates to the frontend.

## What Changes

- **New SessionManager service** in Go backend to track and manage active database sessions
- **Connection pooling** per session with configurable pool size limits
- **Health check system** with periodic ping and auto-reconnect on failure
- **Session lifecycle events** emitted to frontend via Wails events (session:created, session:closed, session:error)
- **Session state tracking** (active, idle, closed) with proper cleanup on shutdown
- **New Go package** at `internal/session/` for session management logic

## Capabilities

### New Capabilities
- `session-management`: Manages active database session lifecycle, connection pooling, health checks, and auto-reconnection
- `session-events`: Real-time session status updates emitted to frontend via Wails events

### Modified Capabilities
- `query-execution`: Sessions now provide pooled connections for query execution instead of direct driver access

## Impact

- **Affected Code**: Query execution service (Phase 5) will use SessionManager to acquire/release connections
- **New Dependencies**: None (uses standard Go sync package and context)
- **Frontend Changes**: New event listeners for session status updates
- **Backend Services**: SessionManager integrates with ConnectionManager (Phase 3) and Database Drivers (Phase 4)
- **Wails Bindings**: SessionManager methods exposed to TypeScript frontend
