# Spec: Keyboard Shortcuts Capability

## ADDED Requirements

### Requirement: Global Keyboard Shortcuts
The system SHALL provide keyboard shortcuts for all common actions across the application.

#### Scenario: Execute query with keyboard
- **WHEN** user presses Ctrl+Enter in the SQL editor
- **THEN** the current query is executed

#### Scenario: New tab with keyboard
- **WHEN** user presses Ctrl+T
- **THEN** a new editor tab is created

### Requirement: Navigation Shortcuts
The system SHALL provide shortcuts for navigating between panels.

#### Scenario: Focus sidebar
- **WHEN** user presses Ctrl+1
- **THEN** focus moves to the sidebar schema browser

#### Scenario: Focus editor
- **WHEN** user presses Ctrl+2
- **THEN** focus moves to the SQL editor

#### Scenario: Focus results
- **WHEN** user presses Ctrl+3
- **THEN** focus moves to the result view

### Requirement: Editor Shortcuts
The system SHALL provide Monaco editor standard shortcuts.

#### Scenario: Format SQL
- **WHEN** user presses Shift+Alt+F in the editor
- **THEN** the SQL is formatted

#### Scenario: Find and replace
- **WHEN** user presses Ctrl+H in the editor
- **THEN** the find/replace dialog opens

### Requirement: Customizable Shortcuts
The system SHALL allow users to customize keyboard shortcuts.

#### Scenario: User changes a shortcut
- **WHEN** user opens settings and changes the "Run Query" shortcut
- **THEN** the new shortcut is saved and used going forward
