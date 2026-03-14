# Spec: Query Execution Pipeline Capability

## ADDED Requirements

### Requirement: QueryExecutor Service
The system SHALL provide a QueryExecutor service for executing SQL queries against database connections with timeout and cancellation support.

#### Scenario: Execute query with default timeout
- **WHEN** user executes a query via QueryExecutor.Execute()
- **THEN** the query is executed with the default timeout (30 seconds) and results are returned

#### Scenario: Execute query with custom timeout
- **WHEN** user calls QueryExecutor.Execute() with a context that has a custom timeout
- **THEN** the query is cancelled if it exceeds the specified timeout

#### Scenario: Cancel running query
- **WHEN** user calls QueryExecutor.Cancel() with a valid query ID
- **THEN** the running query is cancelled and resources are cleaned up

#### Scenario: Query cancellation after completion
- **WHEN** user calls QueryExecutor.Cancel() on a query that has already completed
- **THEN** the call returns without error (no-op)

### Requirement: Context Timeout Enforcement
All query executions SHALL use Go context with timeout to prevent hanging queries.

#### Scenario: Query exceeds timeout
- **WHEN** a query takes longer than the configured timeout
- **THEN** the query is cancelled automatically and an error is returned to the user

#### Scenario: Query completes within timeout
- **WHEN** a query completes before the timeout
- **THEN** results are returned normally and context is cleaned up

### Requirement: Multi-Statement Query Support
The system SHALL support executing multiple SQL statements in a single query.

#### Scenario: Execute batch of INSERT statements
- **WHEN** user submits multiple INSERT statements separated by semicolons
- **THEN** all statements are executed in sequence and total affected rows are reported

#### Scenario: Multi-statement with mixed results
- **WHEN** user submits a batch containing SELECT and INSERT statements
- **THEN** each statement's results are returned separately with appropriate formatting

### Requirement: Query Result Streaming
The system SHALL stream query results for large datasets to avoid memory exhaustion.

#### Scenario: Stream large result set
- **WHEN** a query returns more than 1000 rows
- **THEN** results are streamed in chunks to the frontend

#### Scenario: Streaming with cancellation
- **WHEN** user cancels during streaming
- **THEN** streaming stops immediately and database cursor is closed

### Requirement: Active Query Tracking
The system SHALL track active queries per connection for monitoring and cancellation.

#### Scenario: Track query start and completion
- **WHEN** a query starts executing
- **THEN** the query is added to the active queries map with timestamp

#### Scenario: Query completion cleanup
- **WHEN** a query completes (success or error)
- **THEN** the query is removed from active queries and added to history

### Requirement: Query Execution Events
The system SHALL emit events for query lifecycle monitoring.

#### Scenario: Query start event
- **WHEN** a query begins execution
- **THEN** `query:executing` event is emitted with query ID and connection ID

#### Scenario: Query completion event
- **WHEN** a query completes successfully
- **THEN** `query:completed` event is emitted with query ID, duration, and row count

#### Scenario: Query failure event
- **WHEN** a query fails with an error
- **THEN** `query:failed` event is emitted with query ID and error message
