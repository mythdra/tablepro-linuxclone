# Spec: Query Pagination Capability

## ADDED Requirements

### Requirement: LIMIT/OFFSET Pagination for SQL
The system SHALL support LIMIT/OFFSET pagination for SQL database queries.

#### Scenario: First page of results
- **WHEN** user requests page 1 with page size 100
- **THEN** the query returns the first 100 rows (LIMIT 100 OFFSET 0)

#### Scenario: Second page of results
- **WHEN** user requests page 2 with page size 100
- **THEN** the query returns rows 101-200 (LIMIT 100 OFFSET 100)

#### Scenario: Last page with partial results
- **WHEN** user requests a page beyond the total row count
- **THEN** the query returns remaining rows (less than page size)

### Requirement: Cursor-Based Pagination for NoSQL
The system SHALL support cursor-based pagination for NoSQL databases (MongoDB, Redis).

#### Scenario: MongoDB cursor pagination
- **WHEN** user paginates MongoDB results
- **THEN** the system uses skip/limit with cursor for efficient pagination

#### Scenario: Redis SCAN pagination
- **WHEN** user paginates Redis results
- **THEN** the system uses SCAN command with COUNT parameter

### Requirement: PaginationService Interface
The system SHALL provide a PaginationService for managing query pagination state.

#### Scenario: Create pagination context
- **WHEN** a paginated query is initiated
- **THEN** a PaginationContext is created with page, pageSize, and totalEstimate

#### Scenario: Get next page
- **WHEN** user clicks "Next Page"
- **THEN** PaginationService calculates the correct OFFSET/cursor for the next page

#### Scenario: Get previous page
- **WHEN** user clicks "Previous Page"
- **THEN** PaginationService calculates the correct OFFSET/cursor for the previous page

### Requirement: Page Size Configuration
The system SHALL allow users to configure the page size for query results.

#### Scenario: Default page size
- **WHEN** user executes a query without specifying page size
- **THEN** the default page size (100 rows) is used

#### Scenario: Custom page size
- **WHEN** user sets page size to 500
- **THEN** subsequent queries return 500 rows per page

#### Scenario: Maximum page size limit
- **WHEN** user attempts to set page size above maximum (10000)
- **THEN** the page size is capped at the maximum value

### Requirement: Total Count Estimation
The system SHALL provide an estimate of total rows for pagination UI.

#### Scenario: Exact count for small tables
- **WHEN** a query is executed on a small table (< 10000 rows)
- **THEN** an exact COUNT(*) is performed to get total rows

#### Scenario: Estimate for large tables
- **WHEN** a query is executed on a large table (>= 10000 rows)
- **THEN** an estimate is returned using database statistics (to avoid slow COUNT)

#### Scenario: Count with WHERE clause
- **WHEN** a query has a WHERE clause
- **THEN** COUNT is performed with the same WHERE clause for accurate total

### Requirement: Pagination UI Integration
The system SHALL provide pagination controls in the ResultView component.

#### Scenario: Show current page info
- **WHEN** results are displayed
- **THEN** the UI shows "Page X of Y (Z rows)"

#### Scenario: Navigate to specific page
- **WHEN** user enters a page number in the pagination input
- **THEN** results jump to that page

#### Scenario: Disable navigation buttons
- **WHEN** user is on the first page
- **THEN** the "Previous" button is disabled

### Requirement: Infinite Scroll Support
The system SHALL support infinite scroll mode as an alternative to pagination.

#### Scenario: Scroll to load more
- **WHEN** user scrolls to the bottom of results
- **THEN** the next page of results is automatically loaded

#### Scenario: Infinite scroll with cursor
- **WHEN** using infinite scroll on NoSQL data
- **THEN** cursor-based pagination is used for efficiency

### Requirement: Pagination Performance
The system SHALL ensure pagination does not significantly impact query performance.

#### Scenario: Pagination overhead
- **WHEN** comparing paginated vs non-paginated query
- **THEN** pagination adds less than 10% overhead to query time

#### Scenario: Count query optimization
- **WHEN** getting total count for pagination
- **THEN** COUNT query uses index where possible
