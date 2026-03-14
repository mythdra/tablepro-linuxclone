## ADDED Requirements

### Requirement: Transaction-Based Commit
The system SHALL execute all changes within a database transaction.

#### Scenario: Begin transaction on commit
- **WHEN** user clicks the "Commit Changes" button
- **THEN** a database transaction is started on the active connection
- **AND** all generated SQL statements execute within this transaction

#### Scenario: Commit all statements
- **WHEN** all generated SQL statements execute successfully
- **THEN** the transaction is committed
- **AND** changes are permanently saved to the database
- **AND** the change tracker is cleared

#### Scenario: Rollback on error
- **WHEN** any SQL statement fails during execution
- **THEN** the transaction is rolled back
- **AND** all changes are reverted in the database
- **AND** the change tracker retains all pending changes
- **AND** an error message identifies the failed statement

### Requirement: Foreign Key Constraint Handling
The system SHALL handle foreign key constraint violations gracefully.

#### Scenario: Delete blocked by FK
- **WHEN** a DELETE statement fails due to a foreign key constraint
- **THEN** the error message includes: "Cannot delete row - referenced by [child_table].[child_column]"
- **AND** the transaction is rolled back
- **AND** the user can see which rows failed

#### Scenario: Statement-level error reporting
- **WHEN** statement 3 of 10 fails
- **THEN** the error identifies statement 3 specifically
- **AND** statements 1-2 are noted as successful (but rolled back)
- **AND** the user can choose to retry or discard

### Requirement: Commit Success Feedback
The system SHALL provide clear feedback when commit succeeds.

#### Scenario: Success toast notification
- **WHEN** all changes commit successfully
- **THEN** a green success toast displays: "5 changes committed successfully"
- **AND** the change summary panel is cleared
- **AND** the grid refreshes to show updated data

#### Scenario: Emit success event
- **WHEN** commit completes successfully
- **THEN** a Wails event "data:saved" is emitted
- **AND** frontend components can listen for this event
- **AND** query history is updated with the mutation

### Requirement: Partial Commit Prevention
The system SHALL prevent partial commits where only some changes succeed.

#### Scenario: All-or-nothing semantics
- **WHEN** committing 10 changes and 1 fails
- **THEN** zero changes are persisted to the database
- **AND** the user sees the error for the failed statement
- **AND** all 10 changes remain in the pending state for retry

#### Scenario: User can retry after error
- **WHEN** a commit fails
- **THEN** the user can click "Commit" again to retry
- **OR** the user can click "Discard Changes" to clear all

### Requirement: Long-Running Commit Handling
The system SHALL handle commits that take longer than expected.

#### Scenario: Commit timeout warning
- **WHEN** a commit takes longer than 10 seconds
- **THEN** a loading spinner displays with "Committing changes..."
- **AND** a timeout is set to 30 seconds maximum

#### Scenario: Commit timeout exceeded
- **WHEN** a commit exceeds 30 seconds
- **THEN** the commit is cancelled
- **THEN** an error displays: "Commit timed out. Changes may be partial - check database state."
- **AND** the user is advised to check connection status

### Requirement: Commit Button State Management
The system SHALL manage the Commit button state based on pending changes.

#### Scenario: Commit button disabled when no changes
- **WHEN** there are no pending changes
- **THEN** the "Commit" button is disabled
- **AND** the button shows "(0 changes)" text

#### Scenario: Commit button enabled with changes
- **WHEN** there is at least one pending change
- **THEN** the "Commit" button is enabled
- **AND** the button shows "(N changes)" text with the count

#### Scenario: Commit button shows loading state
- **WHEN** commit is in progress
- **THEN** the button is disabled and shows a spinner
- **AND** the text changes to "Committing..."
