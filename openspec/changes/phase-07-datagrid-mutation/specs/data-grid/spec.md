## ADDED Requirements

### Requirement: Data Grid Display
The system SHALL display query results in an AG Grid virtualized grid with server-side pagination.

#### Scenario: SELECT query returns results
- **WHEN** a SELECT query executes successfully
- **THEN** results are displayed in an AG Grid with columns matching the query result schema

#### Scenario: Large result set pagination
- **WHEN** a query returns more than 100 rows (default page size)
- **THEN** the grid displays pagination controls showing current page and total rows
- **AND** only the visible page of rows is rendered in the DOM

#### Scenario: User changes page size
- **WHEN** user selects a different page size (100, 500, 1000, 5000)
- **THEN** the grid refetches data with the new LIMIT parameter
- **AND** the status bar updates to show the new page size

### Requirement: Virtual Scrolling Performance
The system SHALL render large datasets smoothly using AG Grid's virtual scrolling.

#### Scenario: Grid displays 10,000 rows
- **WHEN** a query returns 10,000 rows
- **THEN** the grid scrolls smoothly without lag
- **AND** DOM node count remains constant (~50 rows rendered)

#### Scenario: User scrolls rapidly through data
- **WHEN** user drags the scrollbar quickly from top to bottom
- **THEN** rows are recycled and rendered progressively
- **AND** the browser remains responsive

### Requirement: Column Definitions
The system SHALL generate column definitions from query result metadata.

#### Scenario: Column headers display
- **WHEN** query results are received
- **THEN** each column displays the column name as the header
- **AND** columns are resizable by dragging the header border

#### Scenario: Data type-specific rendering
- **WHEN** a column has a specific data type (INTEGER, FLOAT, DATE, BOOLEAN)
- **THEN** values are formatted appropriately for that type
- **AND** NULL values display as gray italic "NULL" text

### Requirement: Column Sorting UI
The system SHALL allow users to sort by clicking column headers.

#### Scenario: Initial sort on column
- **WHEN** user clicks a column header
- **THEN** the grid sorts by that column in ascending order
- **AND** a sort indicator (up arrow) appears in the header

#### Scenario: Toggle sort direction
- **WHEN** user clicks a column that is already sorted ascending
- **THEN** the sort order changes to descending
- **AND** the sort indicator changes to a down arrow

#### Scenario: Clear sorting
- **WHEN** user clicks a sorted column a third time
- **THEN** sorting is cleared and original order is restored
- **AND** the sort indicator disappears

### Requirement: NULL Value Rendering
The system SHALL clearly distinguish NULL values from empty strings.

#### Scenario: Query returns NULL values
- **WHEN** a query result contains NULL values in a column
- **THEN** NULL cells display as "<NULL>" in gray italic text
- **AND** the cell is not editable to empty string

#### Scenario: Empty string vs NULL
- **WHEN** a query returns an empty string ('')
- **THEN** the cell displays as empty (blank)
- **AND** it is visually distinct from NULL cells
