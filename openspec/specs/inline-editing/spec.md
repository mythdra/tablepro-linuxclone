# Spec: Inline Editing Capability

## ADDED Requirements

### Requirement: Cell Editing
The system SHALL allow users to edit cell values directly in the data grid.

#### Scenario: User edits a cell
- **WHEN** user double-clicks on an editable cell
- **THEN** the cell enters edit mode with a text input showing the current value

#### Scenario: User saves cell edit
- **WHEN** user presses Enter or clicks outside the cell after editing
- **THEN** an UPDATE statement is generated and executed to save the change

### Requirement: Change Tracking
The system SHALL track all pending changes and display them in a change summary panel.

#### Scenario: Multiple cells edited
- **WHEN** user edits multiple cells
- **THEN** a "Pending Changes" panel shows a summary of all modifications with SQL preview

#### Scenario: User cancels changes
- **WHEN** user clicks "Discard Changes" in the pending changes panel
- **THEN** all unsaved edits are reverted and the grid refreshes

### Requirement: Primary Key Requirement
The system SHALL only allow inline editing for tables with a primary key.

#### Scenario: Table without primary key
- **WHEN** user views a table without a primary key
- **THEN** cells are not editable and a message explains why

### Requirement: Data Type Validation
The system SHALL validate edited values against column data types.

#### Scenario: Invalid data type entry
- **WHEN** user tries to enter a non-numeric value in an INTEGER column
- **THEN** the edit is rejected with an appropriate error message
