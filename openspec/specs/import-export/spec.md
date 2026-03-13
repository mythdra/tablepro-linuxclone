# Spec: Import/Export Capability

## ADDED Requirements

### Requirement: Data Export
The system SHALL allow users to export table data or query results to CSV, JSON, or SQL format.

#### Scenario: Export table to CSV
- **WHEN** user right-clicks a table and selects "Export" → "CSV"
- **THEN** a CSV file is downloaded with all table data

#### Scenario: Export query results to JSON
- **WHEN** user clicks "Export" in the ResultView and selects JSON format
- **THEN** a JSON file is downloaded with the query results as an array of objects

#### Scenario: Export as SQL INSERT statements
- **WHEN** user selects SQL format for export
- **THEN** a SQL file is downloaded with INSERT statements for each row

### Requirement: Data Import
The system SHALL allow users to import CSV or JSON data into a table.

#### Scenario: Import CSV into existing table
- **WHEN** user selects "Import" on a table and uploads a CSV file
- **THEN** the data is inserted into the table with column mapping confirmation

#### Scenario: Import validation fails
- **WHEN** the uploaded file has invalid data types or missing required columns
- **THEN** an error is displayed with details about the validation failure
