# Spec: SQL Editor Capability (Phase 5 Delta)

## ADDED Requirements

### Requirement: Autocomplete/Intellisense
The Monaco editor SHALL provide SQL autocomplete and intellisense powered by database schema metadata.

#### Scenario: Table name autocomplete
- **WHEN** user starts typing a table name in the editor
- **THEN** a dropdown appears with matching table names from the connected database schema

#### Scenario: Column name autocomplete
- **WHEN** user types a table name followed by a dot (e.g., "users.")
- **THEN** a dropdown appears with column names from that table

#### Scenario: SQL keyword autocomplete
- **WHEN** user starts typing an SQL keyword
- **THEN** matching keywords (SELECT, INSERT, UPDATE, etc.) are suggested

#### Scenario: Schema-aware autocomplete
- **WHEN** user switches to a different database connection
- **THEN** autocomplete suggestions update to reflect the new schema

### Requirement: Keyboard Shortcuts
The system SHALL provide keyboard shortcuts for common editor actions.

#### Scenario: Execute query with shortcut
- **WHEN** user presses Ctrl+Enter (or Cmd+Enter on macOS)
- **THEN** the current query is executed

#### Scenario: Format query with shortcut
- **WHEN** user presses Shift+Alt+F
- **THEN** the SQL query is formatted

#### Scenario: New tab with shortcut
- **WHEN** user presses Ctrl+T (or Cmd+T on macOS)
- **THEN** a new query tab is created

### Requirement: Query Cancellation UI
The ResultView SHALL display a Cancel button during query execution.

#### Scenario: Show cancel button during execution
- **WHEN** a query is executing
- **THEN** a "Cancel" button is visible in the toolbar

#### Scenario: Cancel button hides on completion
- **WHEN** query execution completes (success or error)
- **THEN** the Cancel button is hidden

#### Scenario: User cancels query
- **WHEN** user clicks the Cancel button
- **THEN** the query is cancelled and an error message indicates user cancellation

### Requirement: Multi-Statement Execution
The editor SHALL support executing multiple SQL statements in sequence.

#### Scenario: Execute multiple SELECT statements
- **WHEN** user executes a batch of SELECT statements
- **THEN** each result set is displayed in separate tabs or sections

#### Scenario: Execute mixed statement types
- **WHEN** user executes INSERT followed by SELECT
- **THEN** affected row count and result set are both displayed

#### Scenario: Statement delimiter detection
- **WHEN** user has cursor in the middle of multiple statements
- **THEN** the system identifies and executes the statement at cursor position

### Requirement: Result Streaming Integration
The editor SHALL integrate with result streaming for large queries.

#### Scenario: Streaming indicator
- **WHEN** results are being streamed
- **THEN** a loading indicator shows "Streaming results..."

#### Scenario: Partial results display
- **WHEN** first chunk of results arrives
- **THEN** results are displayed while streaming continues

#### Scenario: Streaming completion
- **WHEN** all chunks have been received
- **THEN** the loading indicator is hidden and total row count is shown
