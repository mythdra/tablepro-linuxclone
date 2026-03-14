## ADDED Requirements

### Requirement: UPDATE Statement Generation
The system SHALL generate UPDATE statements from cell changes.

#### Scenario: Single cell update
- **WHEN** a single cell is changed in a row
- **THEN** an UPDATE statement is generated: `UPDATE "schema"."table" SET "column" = $1 WHERE "pk" = $2`
- **AND** parameters include the new value and primary key value

#### Scenario: Multiple cell updates in same row
- **WHEN** multiple cells are changed in the same row
- **THEN** a single UPDATE statement is generated with multiple SET clauses
- **AND** parameters are ordered correctly for the database dialect

#### Scenario: UPDATE with NULL value
- **WHEN** a cell is changed from a value to NULL
- **THEN** the UPDATE statement sets the column to NULL (not an empty string)
- **AND** the parameter is properly bound as NULL

#### Scenario: Primary key excluded from SET
- **WHEN** generating an UPDATE statement
- **THEN** primary key columns are never included in the SET clause
- **AND** only non-PK columns can be updated

### Requirement: INSERT Statement Generation
The system SHALL generate INSERT statements from new rows.

#### Scenario: Single row insertion
- **WHEN** a new row is added with values
- **THEN** an INSERT statement is generated: `INSERT INTO "schema"."table" ("col1", "col2") VALUES ($1, $2)`
- **AND** all non-auto-generated columns are included

#### Scenario: INSERT with NULL values
- **WHEN** a new row has some cells set to NULL
- **THEN** the INSERT statement includes NULL for those columns
- **AND** the statement does not insert empty strings for NULL cells

#### Scenario: Auto-increment column handling
- **WHEN** a table has an auto-increment primary key
- **THEN** the auto-increment column is excluded from the INSERT column list
- **AND** the database generates the PK value automatically

#### Scenario: Multiple row insertions
- **WHEN** multiple new rows are added
- **THEN** separate INSERT statements are generated for each row
- **OR** a single batch INSERT if the database supports it (e.g., `VALUES ($1, $2), ($3, $4)`)

### Requirement: DELETE Statement Generation
The system SHALL generate DELETE statements from deleted rows.

#### Scenario: Single row deletion
- **WHEN** a row is marked for deletion
- **THEN** a DELETE statement is generated: `DELETE FROM "schema"."table" WHERE "pk" = $1`
- **AND** the WHERE clause uses the primary key value

#### Scenario: Composite primary key deletion
- **WHEN** a table has a composite primary key (multiple columns)
- **THEN** the DELETE WHERE clause includes all PK columns with AND
- **AND** all PK values are bound as parameters

#### Scenario: Multiple row deletions
- **WHEN** multiple rows are marked for deletion
- **THEN** separate DELETE statements are generated for each row
- **AND** statements are executed in order of selection

### Requirement: Dialect-Specific Parameter Markers
The system SHALL generate SQL with correct parameter placeholders for each database dialect.

#### Scenario: PostgreSQL parameters
- **WHEN** generating SQL for PostgreSQL
- **THEN** parameters use $1, $2, $3... format
- **AND** identifiers are double-quoted ("column")

#### Scenario: MySQL parameters
- **WHEN** generating SQL for MySQL
- **THEN** parameters use ? format
- **AND** identifiers are backtick-quoted (`column`)

#### Scenario: SQLite parameters
- **WHEN** generating SQL for SQLite
- **THEN** parameters use ? format
- **AND** identifiers are double-quoted ("column")

#### Scenario: SQL Server parameters
- **WHEN** generating SQL for MSSQL
- **THEN** parameters use @p1, @p2, @p3... format
- **AND** identifiers are bracket-quoted ([column])

### Requirement: Batch Statement Generation
The system SHALL group statements for efficient batch execution.

#### Scenario: Order statements correctly
- **WHEN** generating statements from mixed changes
- **THEN** DELETE statements execute first (to handle FK constraints)
- **AND** UPDATE statements execute second
- **AND** INSERT statements execute last

#### Scenario: Statement grouping
- **WHEN** there are 50 pending changes
- **THEN** statements are grouped into batches of 100 or fewer
- **AND** each batch is executed as a unit

### Requirement: SQL Preview Before Commit
The system SHALL display generated SQL to the user before committing.

#### Scenario: Preview panel shows SQL
- **WHEN** there are pending changes and user views the change panel
- **THEN** the generated SQL statements are displayed in a read-only editor
- **AND** statements are syntax-highlighted using Monaco Editor

#### Scenario: SQL shows change count
- **WHEN** previewing changes
- **THEN** a summary shows: "3 UPDATEs, 2 INSERTs, 1 DELETE"
- **AND** the user can review before committing
