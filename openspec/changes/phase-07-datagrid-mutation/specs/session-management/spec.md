## MODIFIED Requirements

### Requirement: Tab-Level Change Tracking State
The system SHALL track pending data changes per query tab.

#### Scenario: Tab tracks its own changes
- **WHEN** a user edits cells in Tab A
- **THEN** only Tab A shows pending change indicators
- **AND** Tab B (if open) remains unaffected

#### Scenario: Change count in tab header
- **WHEN** a tab has pending changes
- **THEN** the tab header displays a badge with the change count (e.g., "Query 1 (3)")
- **AND** the badge is removed after commit or discard

#### Scenario: Close tab with pending changes
- **WHEN** user attempts to close a tab with uncommitted changes
- **THEN** a confirmation dialog appears: "You have N uncommitted changes. Close anyway?"
- **AND** clicking "Close" discards all changes
- **AND** clicking "Cancel" keeps the tab open

### Requirement: Session Change Manager Access
The system SHALL provide a per-session change manager accessible from frontend.

#### Scenario: Get pending changes
- **WHEN** frontend requests pending changes for a session
- **THEN** the backend returns a structured object with:
  - `cellChanges`: array of cell-level changes
  - `insertedRows`: array of new row data
  - `deletedRows`: array of primary key values
- **AND** the response includes schema and table names

#### Scenario: Clear changes on query execution
- **WHEN** a new query is executed in a tab
- **THEN** all pending changes for that tab are discarded
- **AND** a warning appears if there were uncommitted changes

### Requirement: Transaction per Session
The system SHALL manage transactions at the session level.

#### Scenario: Begin transaction for commit
- **WHEN** user initiates a commit
- **THEN** the session manager begins a transaction on the active connection
- **AND** the transaction is associated with the session ID

#### Scenario: Transaction isolation
- **WHEN** a transaction is active
- **THEN** other queries in the same session see uncommitted changes
- **AND** other sessions do not see uncommitted changes (database default isolation)
