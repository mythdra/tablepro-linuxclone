# Spec: Query History Capability

## ADDED Requirements

### Requirement: Query History Persistence
The system SHALL store all executed queries in a SQLite database for later retrieval.

#### Scenario: Query is executed
- **WHEN** a user executes a SQL query
- **THEN** the query text, timestamp, duration, and connection info are saved to the history database

#### Scenario: Application restart
- **WHEN** the user closes and reopens the application
- **THEN** previously executed queries are still available in the history panel

### Requirement: History Search
The system SHALL provide full-text search for querying history.

#### Scenario: User searches history
- **WHEN** user types in the history search box
- **THEN** the history list filters to show only queries containing the search term

### Requirement: History Re-execution
The system SHALL allow users to re-execute queries from history.

#### Scenario: Re-execute from history
- **WHEN** user double-clicks on a query in the history panel
- **THEN** the query is loaded into a new editor tab and ready to execute

### Requirement: History Retention Policy
The system SHALL retain query history with configurable limits.

#### Scenario: History reaches limit
- **WHEN** the history database exceeds 1000 queries
- **THEN** the oldest queries are automatically deleted to maintain the limit
