# Spec: SQL Editor Capability

## ADDED Requirements

### Requirement: Multi-tab SQL Editor Interface
The system SHALL provide a multi-tab SQL editor interface with Monaco Editor integration for writing and executing SQL queries.

#### Scenario: User opens new editor tab
- **WHEN** user clicks the "+" button in the editor toolbar
- **THEN** a new tab is created with a unique name ("Query 1", "Query 2", etc.) and an empty editor

#### Scenario: User switches between tabs
- **WHEN** user clicks on a different tab
- **THEN** that tab becomes active and displays its SQL content in the Monaco editor

#### Scenario: User closes a tab
- **WHEN** user clicks the "×" button on a tab
- **THEN** the tab is closed and focus moves to the adjacent tab

### Requirement: SQL Syntax Highlighting
The Monaco editor SHALL provide SQL syntax highlighting with support for MySQL and PostgreSQL dialects.

#### Scenario: User types SQL keywords
- **WHEN** user types SQL keywords like SELECT, FROM, WHERE
- **THEN** keywords are highlighted with appropriate colors based on the SQL dialect

#### Scenario: User switches database connection
- **WHEN** user switches from MySQL to PostgreSQL connection
- **THEN** syntax highlighting updates to reflect the correct dialect

### Requirement: Query Execution
The system SHALL execute SQL queries against the active database connection and display results.

#### Scenario: User runs a SELECT query
- **WHEN** user clicks the "Run" button or presses Ctrl+Enter with a SELECT query
- **THEN** the query is executed and results are displayed in the ResultView component

#### Scenario: User runs an INSERT query
- **WHEN** user executes an INSERT statement
- **THEN** the affected row count is displayed and no result grid is shown

#### Scenario: Query execution fails
- **WHEN** a query has a syntax error or the database returns an error
- **THEN** an error message is displayed in the ResultView with the database error details

### Requirement: Run Selected SQL
The system SHALL allow users to execute only a selected portion of SQL.

#### Scenario: User has text selected
- **WHEN** user has text selected in the editor and clicks "Run Selection"
- **THEN** only the selected text is executed as a query

#### Scenario: No text selected
- **WHEN** no text is selected and user clicks "Run Selection"
- **THEN** the system identifies and executes the SQL statement at the cursor position

### Requirement: SQL Formatting
The system SHALL format SQL queries for readability.

#### Scenario: User formats a query
- **WHEN** user clicks the "Format" button
- **THEN** the SQL in the active tab is reformatted with proper indentation and line breaks
