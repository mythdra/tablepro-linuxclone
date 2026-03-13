# Spec: Data Grid Capability

## ADDED Requirements

### Requirement: Query Result Display
The system SHALL display query results in a virtualized data grid with pagination.

#### Scenario: SELECT query returns results
- **WHEN** a SELECT query executes successfully
- **THEN** results are displayed in a grid with columns matching the query result

#### Scenario: Large result set pagination
- **WHEN** a query returns more than 100 rows
- **THEN** the grid displays pagination controls showing 100 rows per page

### Requirement: Column Sorting
The system SHALL allow users to sort result columns.

#### Scenario: User clicks column header
- **WHEN** user clicks on a column header
- **THEN** the grid sorts by that column in ascending order

#### Scenario: User clicks sorted column again
- **WHEN** user clicks a column that is already sorted
- **THEN** the sort order toggles between ascending and descending

### Requirement: NULL Value Display
The system SHALL clearly distinguish NULL values from empty strings.

#### Scenario: Query returns NULL values
- **WHEN** a query result contains NULL values
- **THEN** NULL cells display as "<NULL>" in a distinct gray color

### Requirement: Copy Cell Values
The system SHALL allow users to copy cell values to the clipboard.

#### Scenario: User copies a cell
- **WHEN** user double-clicks on a cell or presses Ctrl+C while focused
- **THEN** the cell value is copied to the clipboard
