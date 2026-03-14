# Spec: Query History Capability (Phase 5 Delta)

## ADDED Requirements

### Requirement: In-Memory History Tracking
The system SHALL track executed queries in memory for the current session.

#### Scenario: Query added to history
- **WHEN** a query is executed successfully
- **THEN** the query text and execution metadata are added to in-memory history

#### Scenario: History includes timestamp
- **WHEN** a query is executed
- **THEN** the exact timestamp (date and time) is recorded with the query

#### Scenario: History includes duration
- **WHEN** a query completes
- **THEN** the execution duration is recorded in the history entry

### Requirement: Query Deduplication
The system SHALL deduplicate queries in history to avoid redundant entries.

#### Scenario: Duplicate query detection
- **WHEN** user executes the same query twice (ignoring whitespace)
- **THEN** the existing history entry is updated with new timestamp instead of creating duplicate

#### Scenario: Whitespace normalization
- **WHEN** comparing queries for deduplication
- **THEN** leading/trailing whitespace and case are normalized before comparison

### Requirement: Last N Queries Per Connection
The system SHALL retain only the last N queries per database connection.

#### Scenario: Default history limit
- **WHEN** no limit is configured
- **THEN** the last 50 queries per connection are retained

#### Scenario: History limit enforcement
- **WHEN** the 51st query is executed for a connection
- **THEN** the oldest query for that connection is removed from history

#### Scenario: Per-connection history
- **WHEN** user switches to a different connection
- **THEN** history is filtered to show only queries for the active connection

### Requirement: History UI Panel
The system SHALL provide a UI panel for viewing query history.

#### Scenario: Display history entries
- **WHEN** user opens the history panel
- **THEN** executed queries are listed with timestamp and duration

#### Scenario: Click to load query
- **WHEN** user clicks on a history entry
- **THEN** the query text is loaded into a new editor tab

#### Scenario: Filter history by search
- **WHEN** user types in the history search box
- **THEN** the history list filters to show only matching queries

### Requirement: History Events
The system SHALL emit events when history is modified.

#### Scenario: History entry added
- **WHEN** a new query is added to history
- **THEN** `history:added` event is emitted

#### Scenario: History cleared
- **WHEN** user clears history for a connection
- **THEN** `history:cleared` event is emitted with connection ID
