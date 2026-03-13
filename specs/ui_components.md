# TablePro UI Components & Navigation

## Overview
TablePro's user interface is built primarily with SwiftUI, integrated with AppKit where necessary (e.g., window management, complex text editors). The UI architecture strongly follows the Coordinator pattern to separate navigation/business logic from view definitions.

## Main Layout & Navigation
The application window layout is divided into standard database client sections:
1. **Sidebar (`Views/Sidebar`)**: Displays connection groups and active connections.
2. **Database Switcher / Quick Switcher**: Allows rapid switching between schemas and databases.
3. **Main Content (`Views/Main`)**: The central working area.
4. **Right Sidebar**: Inspector panel for table metadata, column details, and foreign keys.

### MainContentCoordinator
The central brain of the UI is `MainContentCoordinator`. Due to its immense responsibility, it's chunked into multiple extensions (`+Alerts`, `+Filtering`, `+Pagination`, `+RowOperations`, etc.). 
It manages:
- The active selected table or query tab.
- Pagination state and data fetching calls.
- Filter and sort states for grid views.
- Toggling of UI panels.

## SQL Editor (`Views/Editor`)
The SQL Editor doesn't use standard SwiftUI TextEdit due to performance requirements.
- **Engine**: CodeEditSourceEditor (via SPM), enabling tree-sitter based syntax highlighting, blazing fast text layout, and multi-cursor editing.
- **Theme**: `SQLEditorTheme` defining colors/fonts, bridged by `TableProEditorTheme`.
- **Autocomplete**: Handled by `CompletionEngine` and bridged to the editor via `SQLCompletionAdapter`.
- **Query Execution**: `QueryEditorView` handles splitting multiple statements and executing the selected text or current statement under cursor.
- **AI Integration**: Editor context menus feature inline AI assistance for query generation and explanation.

## Data Grid & Results (`Views/Results`)
The result set from a SQL query or table browse is displayed in a highly optimized native grid.
- Capable of inline cell editing, handled via `DataChangeManager`.
- Supports pagination and infinite scrolling conventions.

## Connections UI
The `Connection` views handle forms for SSH tunneling, SSL configurations, server addresses, and custom driver flags (like MongoDB connection strings or Redis databases).
