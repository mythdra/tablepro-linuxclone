## ADDED Requirements

### Requirement: Cell Editing Activation
The system SHALL allow users to edit cell values by double-clicking or pressing Enter.

#### Scenario: Double-click activates edit mode
- **WHEN** user double-clicks on an editable cell
- **THEN** the cell enters edit mode with a text input
- **AND** the current value is selected for easy replacement

#### Scenario: Enter key activates edit mode
- **WHEN** a cell is focused and user presses Enter
- **THEN** the cell enters edit mode with a text input
- **AND** the cursor is positioned at the end of the value

#### Scenario: Editing non-editable cells
- **WHEN** user attempts to edit a non-editable cell (primary key or read-only table)
- **THEN** the cell does not enter edit mode
- **AND** a tooltip explains why editing is disabled

### Requirement: Save Cell Changes
The system SHALL save cell changes when user confirms the edit.

#### Scenario: Press Enter to save
- **WHEN** user is editing a cell and presses Enter
- **THEN** the new value is saved to the change tracker
- **AND** the cell exits edit mode
- **AND** the cell background turns yellow indicating pending change

#### Scenario: Click outside to save
- **WHEN** user is editing a cell and clicks outside the cell
- **THEN** the new value is saved to the change tracker
- **AND** the cell exits edit mode
- **AND** focus moves to the clicked cell

#### Scenario: Escape to cancel edit
- **WHEN** user is editing a cell and presses Escape
- **THEN** the edit is cancelled
- **AND** the cell reverts to its original value
- **AND** the cell exits edit mode

### Requirement: Data Type Validation
The system SHALL validate edited values against column data types before accepting.

#### Scenario: Valid integer input
- **WHEN** user enters a valid integer in an INTEGER column
- **THEN** the edit is accepted
- **AND** the value is saved to the change tracker

#### Scenario: Invalid integer input
- **WHEN** user enters non-numeric text in an INTEGER column
- **THEN** the edit is rejected
- **AND** an error toast displays "Invalid integer value"
- **AND** the cell reverts to its original value

#### Scenario: Date format validation
- **WHEN** user enters an invalid date in a DATE column
- **THEN** the edit is rejected
- **AND** an error toast displays "Invalid date format. Use YYYY-MM-DD"

#### Scenario: Boolean column editing
- **WHEN** user edits a BOOLEAN column
- **THEN** a dropdown appears with TRUE/FALSE options
- **AND** the selected value is saved

### Requirement: Primary Key Edit Prevention
The system SHALL prevent editing of primary key columns.

#### Scenario: Attempt to edit primary key
- **WHEN** user double-clicks a primary key cell
- **THEN** the cell does not enter edit mode
- **AND** a message displays "Primary key columns cannot be edited"

#### Scenario: Visual indication of non-editable columns
- **WHEN** a table is displayed with primary key columns
- **THEN** primary key columns have a distinct background color
- **AND** the cursor shows "not-allowed" when hovering over PK cells

### Requirement: NULL Value Editing
The system SHALL allow users to set a cell to NULL or from NULL to a value.

#### Scenario: Set cell to NULL
- **WHEN** user is editing a nullable cell
- **AND** user presses a special "Set NULL" button or shortcut
- **THEN** the cell value is set to NULL
- **AND** the change is tracked as a pending change

#### Scenario: Edit NULL cell to value
- **WHEN** user edits a cell that currently contains NULL
- **AND** enters a non-empty value
- **THEN** the NULL is replaced with the new value
- **AND** the change is tracked as a pending change
