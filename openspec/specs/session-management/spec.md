## ADDED Requirements

### Requirement: Session Lifecycle Management
The system SHALL provide session lifecycle management for active database connections.

#### Scenario: Create new session
- **WHEN** user connects to a database successfully
- **THEN** a new session is created with a unique UUID and tracked by SessionManager

#### Scenario: Get existing session
- **WHEN** frontend requests session by UUID
- **THEN** SessionManager returns the session if it exists and is active

#### Scenario: Close session
- **WHEN** user disconnects or app shuts down
- **THEN** the session is closed, all pooled connections are released, and session is removed from tracking

#### Scenario: Get non-existent session
- **WHEN** frontend requests a session that doesn't exist or is closed
- **THEN** SessionManager returns an error indicating session not found

### Requirement: Session State Tracking
The system SHALL track session state throughout its lifecycle.

#### Scenario: Session created state
- **WHEN** a session is created
- **THEN** its state is set to "active"

#### Scenario: Session idle detection
- **WHEN** a session has no active queries for 60 seconds
- **THEN** its state changes to "idle"

#### Scenario: Session closed state
- **WHEN** a session is closed
- **THEN** its state changes to "closed" and no further operations are allowed

### Requirement: Connection Pooling Per Session
The system SHALL maintain a connection pool for each session to enable connection reuse.

#### Scenario: Pool initialization
- **WHEN** a session is created
- **THEN** a connection pool is initialized with 1 connection (minimum pool size)

#### Scenario: Pool expansion
- **WHEN** all pooled connections are in use and a new query is executed
- **THEN** the pool expands up to the maximum size (default: 5 connections)

#### Scenario: Pool exhaustion handling
- **WHEN** pool is at maximum size and all connections are in use
- **THEN** the request queues until a connection becomes available (timeout: 30 seconds)

#### Scenario: Connection reuse
- **WHEN** a query completes and returns a connection to the pool
- **THEN** the connection is marked as available for reuse by subsequent queries

### Requirement: Connection Health Monitoring
The system SHALL monitor connection health and detect stale connections.

#### Scenario: Periodic ping
- **WHEN** a session is active
- **THEN** a background goroutine pings the database every 30 seconds with `SELECT 1`

#### Scenario: Stale connection detection
- **WHEN** a ping fails or times out
- **THEN** the connection is marked as stale and removed from the pool

#### Scenario: Skip ping for active sessions
- **WHEN** a session has executed a query within the last 30 seconds
- **THEN** the health check skips the ping (recent activity = healthy)

### Requirement: Automatic Reconnection
The system SHALL attempt to automatically reconnect when connections fail.

#### Scenario: Reconnect on health check failure
- **WHEN** a health check detects a stale connection
- **THEN** the system attempts to reconnect with exponential backoff (1s, 2s, 4s, 8s, max 30s)

#### Scenario: Reconnect success
- **WHEN** reconnection succeeds
- **THEN** the session state returns to "active" and normal operations resume

#### Scenario: Reconnect failure after max retries
- **WHEN** reconnection fails 5 times
- **THEN** the session is closed and user is notified to reconnect manually

#### Scenario: Reconnect during query execution
- **WHEN** a query fails due to connection loss
- **THEN** the system attempts one immediate reconnect before returning an error
