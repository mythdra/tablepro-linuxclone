# Spec: Table Structure Capability

## ADDED Requirements

### Requirement: Four-Tab Structure Panel
The system SHALL display table structure information in a four-tab panel: Columns, Indexes, Foreign Keys, and DDL.

#### Scenario: User selects a table
- **WHEN** user clicks on a table in the sidebar
- **THEN** the Structure Panel opens showing the Columns tab with all column definitions

#### Scenario: User switches structure tabs
- **WHEN** user clicks on different tabs (Indexes, Foreign Keys, DDL)
- **THEN** the panel displays the corresponding structure information

### Requirement: Column Information Display
The system SHALL display detailed column information including name, data type, nullability, and default values.

#### Scenario: Viewing table columns
- **WHEN** the Columns tab is displayed
- **THEN** each row shows column name, data type, nullable status, and default value

### Requirement: Index Information Display
The system SHALL display index information including name, columns, type, and uniqueness.

#### Scenario: Viewing table indexes
- **WHEN** the Indexes tab is selected
- **THEN** all indexes are listed with their columns, type (BTREE, HASH), and uniqueness

### Requirement: Foreign Key Display
The system SHALL display foreign key relationships.

#### Scenario: Viewing foreign keys
- **WHEN** the Foreign Keys tab is selected
- **THEN** each foreign key shows the source column, referenced table, and referenced column

### Requirement: DDL Display
The system SHALL display the CREATE TABLE statement for the selected table.

#### Scenario: Viewing DDL
- **WHEN** the DDL tab is selected
- **THEN** the complete CREATE TABLE statement is displayed in a read-only Monaco editor
