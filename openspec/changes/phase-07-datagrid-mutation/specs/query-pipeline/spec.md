## MODIFIED Requirements

### Requirement: Server-Side Pagination
The system SHALL support server-side pagination for query results.

#### Scenario: Pagination parameters passed to backend
- **WHEN** frontend requests query results with pagination
- **THEN** the request includes `limit` and `offset` parameters
- **AND** the backend applies LIMIT/OFFSET to the query

#### Scenario: Navigate to next page
- **WHEN** user clicks "Next Page" in the grid
- **THEN** the backend receives a request with `offset = currentOffset + limit`
- **AND** results for the new page are returned
- **AND** the grid replaces the displayed data

#### Scenario: Navigate to specific page
- **WHEN** user enters page number "10" in the pagination control
- **THEN** the backend calculates `offset = (10 - 1) * pageSize`
- **AND** results for page 10 are returned

### Requirement: Server-Side Sorting
The system SHALL handle sorting on the backend to avoid loading all data.

#### Scenario: Sort by column
- **WHEN** user clicks a column header to sort
- **THEN** the frontend sends sort parameters: `{ column: "name", order: "ASC" }`
- **AND** the backend rebuilds the query with `ORDER BY "name" ASC`
- **AND** the first page of sorted results is returned

#### Scenario: Sort with pagination
- **WHEN** user sorts and paginates simultaneously
- **THEN** the backend applies ORDER BY before LIMIT/OFFSET
- **AND** results are consistent across pages

### Requirement: Estimated Row Count
The system SHALL provide an estimated total row count for pagination UI.

#### Scenario: Large result set count
- **WHEN** a query returns more than the page size
- **THEN** the backend provides an estimated total count (via EXPLAIN or COUNT(*) subquery)
- **AND** the grid displays "Showing 1-100 of ~50,000 rows"

#### Scenario: Exact count on demand
- **WHEN** user clicks "Get exact count"
- **THEN** the backend executes `SELECT COUNT(*)` on the base query
- **AND** the exact total is displayed (may take time for large datasets)
