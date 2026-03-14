## ADDED Requirements

### Requirement: Cell Change Tracking
The system SHALL track all cell-level changes with original and new values.

#### Scenario: User edits a cell
- **WHEN** user changes a cell value from "John" to "Jane"
- **THEN** the change is recorded with original value "John" and new value "Jane"
- **AND** the row index and column name are stored
- **AND** the primary key values for that row are captured

#### Scenario: Multiple edits to same cell
- **WHEN** user edits a cell from "John" to "Jane" then to "Bob"
- **THEN** only the latest change ("John" → "Bob") is tracked
- **AND** intermediate changes are not stored

#### Scenario: Edit then revert to original
- **WHEN** user edits a cell from "John" to "Jane" then back to "John"
- **THEN** the change is automatically removed from the change tracker
- **AND** the cell no longer shows a pending change indicator

### Requirement: Row Insertion Tracking
The system SHALL track newly inserted rows with all column values.

#### Scenario: User adds a new row
- **WHEN** user clicks the "Add Row" button
- **THEN** a new empty row is added to the grid
- **AND** the row is tracked as an inserted row with green background
- **AND** all column values are captured as the new row data

#### Scenario: User fills in new row data
- **WHEN** user enters values into cells of a new row
- **THEN** each cell edit is tracked within the inserted row context
- **AND** the complete row data is available for INSERT generation

#### Scenario: Multiple new rows
- **WHEN** user adds multiple new rows
- **THEN** each row is tracked separately with unique row indices
- **AND** rows are ordered by insertion sequence

### Requirement: Row Deletion Tracking
The system SHALL track deleted rows by their primary key values.

#### Scenario: User deletes a row
- **WHEN** user selects a row and presses Delete key
- **THEN** the row is marked as deleted with red strikethrough styling
- **AND** the primary key values are stored for DELETE generation
- **AND** the row remains visible but visually distinct

#### Scenario: Delete multiple rows
- **WHEN** user selects multiple rows and presses Delete
- **THEN** all selected rows are marked as deleted
- **AND** each row's primary key is stored

#### Scenario: Delete then undelete same row
- **WHEN** user deletes a row then immediately undeletes it (Ctrl+Z)
- **THEN** the row is removed from the deleted rows tracker
- **AND** the row styling returns to normal

### Requirement: Change Visualization
The system SHALL provide visual indicators for all pending changes.

#### Scenario: Modified cell indicator
- **WHEN** a cell has been edited but not committed
- **THEN** the cell background is yellow (#FEF3C7)
- **AND** hovering shows a tooltip with original and new values

#### Scenario: New row indicator
- **WHEN** a row is newly inserted
- **THEN** the entire row has a green background (#D1FAE5)
- **AND** all cells in the row are editable

#### Scenario: Deleted row indicator
- **WHEN** a row is marked for deletion
- **THEN** the row has a red background (#FEE2E2)
- **AND** text is displayed with strikethrough
- **AND** cells are not editable

### Requirement: Undo/Redo for Changes
The system SHALL support undo and redo operations for changes before commit.

#### Scenario: Undo cell edit
- **WHEN** user edits a cell then presses Ctrl+Z
- **THEN** the cell reverts to its original value
- **AND** the change is removed from the change tracker
- **AND** the undo action is added to the redo stack

#### Scenario: Redo cell edit
- **WHEN** user undoes a cell edit then presses Ctrl+Y
- **THEN** the cell is restored to the edited value
- **AND** the change is re-added to the change tracker
- **AND** the undo action is moved back to the undo stack

#### Scenario: Undo row insertion
- **WHEN** user adds a new row then presses Ctrl+Z
- **THEN** the new row is removed from the grid
- **AND** the row is removed from the inserted rows tracker

#### Scenario: Undo stack limit
- **WHEN** user performs more than 100 undo actions
- **THEN** the oldest undo actions are discarded
- **AND** a warning toast notifies "Undo history limit reached"

### Requirement: Discard All Changes
The system SHALL allow users to discard all pending changes at once.

#### Scenario: Discard changes button
- **WHEN** user clicks the "Discard Changes" button
- **THEN** all pending changes are reverted
- **AND** the grid refreshes to show original data
- **AND** a confirmation dialog appears if there are more than 10 changes

#### Scenario: Discard with no changes
- **WHEN** user clicks "Discard Changes" with no pending changes
- **THEN** the button is disabled
- **OR** a toast displays "No changes to discard"
