# Spec: Result Set Handling Capability

## ADDED Requirements

### Requirement: ResultSet Data Structure
The system SHALL return query results in a ResultSet struct with metadata and column-oriented data storage.

#### Scenario: SELECT query returns ResultSet
- **WHEN** a SELECT query is executed successfully
- **THEN** a ResultSet is returned with Columns metadata and Rows data

#### Scenario: ResultSet includes row count
- **WHEN** a query completes
- **THEN** ResultSet.RowCount contains the total number of rows returned

#### Scenario: ResultSet includes query time
- **WHEN** a query completes
- **THEN** ResultSet.QueryTime contains the execution duration

### Requirement: Column Metadata
The system SHALL provide column metadata including name, type, and nullability for each column in the result.

#### Scenario: Column info from PostgreSQL
- **WHEN** a query returns columns from PostgreSQL
- **THEN** ColumnInfo includes PostgreSQL type names (e.g., "timestamp with time zone", "jsonb")

#### Scenario: Nullable column detection
- **WHEN** a column is defined as NULLABLE in the database schema
- **THEN** ColumnInfo.Nullable is set to true

### Requirement: Data Type Mapping
The system SHALL map database-specific types to normalized DataType enums for consistent frontend handling.

#### Scenario: PostgreSQL types mapped
- **WHEN** PostgreSQL returns types like TIMESTAMP, BYTEA, UUID
- **THEN** they are mapped to normalized types: datetime, blob, uuid

#### Scenario: MySQL ENUM type
- **WHEN** MySQL returns an ENUM column
- **THEN** it is mapped to the string type with enum values preserved in metadata

#### Scenario: MongoDB BSON types
- **WHEN** MongoDB returns BSON types (ObjectId, Date, Array)
- **THEN** they are converted to JSON-serializable Go types

### Requirement: NULL Value Handling
The system SHALL correctly handle NULL values in query results.

#### Scenario: NULL in result set
- **WHEN** a query returns a row with NULL values
- **THEN** NULL values are represented as nil in Go and null in JSON response

#### Scenario: NULL in numeric column
- **WHEN** a numeric column contains NULL
- **THEN** the value is null (not 0) in the frontend

#### Scenario: NULL in string column
- **WHEN** a string column contains NULL
- **THEN** the value is null (not empty string) in the frontend

### Requirement: Data Formatting
The system SHALL format special data types for display in the frontend.

#### Scenario: Date/time formatting
- **WHEN** a datetime column is returned
- **THEN** it is formatted as ISO 8601 string (e.g., "2026-03-14T10:30:00Z")

#### Scenario: Boolean formatting
- **WHEN** a boolean column is returned
- **THEN** it is represented as true/false (not 1/0)

#### Scenario: Numeric formatting
- **WHEN** a numeric column with large numbers is returned
- **THEN** precision is preserved without scientific notation

#### Scenario: JSON/BLOB formatting
- **WHEN** a JSON or BLOB column is returned
- **THEN** it is base64-encoded or pretty-printed for display

### Requirement: Multiple Result Sets Support
The system SHALL handle queries that return multiple result sets (e.g., stored procedures, batch queries).

#### Scenario: Batch query with multiple SELECTs
- **WHEN** a query contains multiple SELECT statements
- **THEN** each result set is returned separately with its own metadata

#### Scenario: Mixed statement batch
- **WHEN** a batch contains INSERT followed by SELECT
- **THEN** affected row count and result set are both returned

### Requirement: Error Result Handling
The system SHALL return structured error information when queries fail.

#### Scenario: Syntax error
- **WHEN** a query has invalid SQL syntax
- **THEN** an error is returned with the database error message and query position

#### Scenario: Connection lost during query
- **WHEN** a connection is lost while executing a query
- **THEN** an error is returned indicating the connection failure

#### Scenario: Permission denied
- **WHEN** a user lacks permission for a query
- **THEN** an error is returned with the permission error details
