# Spec: Filter Builder Capability

## ADDED Requirements

### Requirement: Visual Filter Builder UI
The system SHALL provide a visual interface for building WHERE clauses without writing SQL.

#### Scenario: User opens filter builder
- **WHEN** user clicks the "Filter" button in the data grid toolbar
- **THEN** a filter builder panel opens with column selection, operator dropdown, and value input

#### Scenario: User adds a filter condition
- **WHEN** user selects a column, chooses an operator, and enters a value
- **THEN** the filter condition is added to the active filter list

### Requirement: Filter Operators
The system SHALL provide 18 operators for building filters including =, ≠, <, >, ≤, ≥, LIKE, IN, BETWEEN, IS NULL, etc.

#### Scenario: User selects LIKE operator
- **WHEN** user selects the LIKE operator for a text column
- **THEN** the value input shows a hint for using % wildcards

### Requirement: Multiple Filter Conditions
The system SHALL allow combining multiple filter conditions with AND/OR logic.

#### Scenario: User adds multiple filters
- **WHEN** user adds two or more filter conditions
- **THEN** the filters are combined with AND by default, with option to switch to OR

#### Scenario: User removes a filter
- **WHEN** user clicks the "×" button on a filter condition
- **THEN** that condition is removed and the grid refreshes with updated results

### Requirement: Filter Application
The system SHALL apply filters to the displayed data.

#### Scenario: User applies filters
- **WHEN** user clicks "Apply Filters"
- **THEN** the data grid refreshes showing only rows matching the filter criteria
