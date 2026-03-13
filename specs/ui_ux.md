# UI/UX Specification

This document details the user interface components, layouts, and interaction patterns required for the C++ rewrite to maintain parity with the modern macOS TablePro application. The UI will be built primarily with Qt Quick (QML) for rapid layout and animations, backed by QObject C++ controllers.

## 1. Global Layout Structure
The main application uses a multi-pane layout pattern common in modern IDEs.

### Toolbar (Top)
- **Left Region**: Traffic lights (macOS) or standard window controls.
- **Center Region**: Connection Status Indicator (Red/Yellow/Green dot), Active Database/Schema drop-down picker, Database Server Version text.
- **Right Region**: Actions
  - Run Query (Cmd+R or F5)
  - Cancel Statement (Red stop icon)
  - Toggle Right Sidebar (AI Chat/Context)

### Sidebar (Left Pane)
- **Top Section**: "Favorites" or "Connections" list if in the welcome window. Inside a connection window, it shows the database schema tree.
- **Search Bar**: Real-time filtering of tables/views/routines.
- **Tree View**:
  - Top Level: Schemas / Databases (e.g., `public`, `information_schema`).
  - Second Level: Tables, Views, Routines organized in collapsible folders.
  - Contextual Menus (Right-click): Open table, Copy name, Drop, Truncate, DDL.

### Main Workspace (Center Pane)
Contains the Tabbed Interface (Query Tabs, Table Tabs, Structure Tabs).

### Right Sidebar (Right Pane, Optional/Collapsible)
- **AI Chat Tab**: Conversational interface communicating with the AI. Streams markdown.
- **Query History Tab**: Local searchable history of previously run queries on this connection.
- **Format Tab**: Prettier/SQL formatter configuration.

## 2. Core Views Detailed

### 2.1 Connection Editor Form
A complex modal or dedicated window view.
- **Header**: Connection Group, Connection Name, Color Picker (Colored dot).
- **Tabbed Configuration Area**:
  - `Basic`: DB Type dropdown, Host, Port, Username, Password, Database Name.
  - `SSH`: Checkbox to enable. Tunnel Host, Port, User. Auth Method Picker (Password, Key, Agent).
  - `SSL`: Checkbox to enable. Mode picker. File picker inputs for ca-cert, client-cert, client-key.
  - `Advanced`: Safe mode picker, Default Schema, Pre-connect Script textbox, Startup Commands textbox.
- **Footer**: "Test Connection" button, "Connect" button, "Save" button.

### 2.2 SQL Editor (The Query Tab)
- Needs a custom QML component wrapping a highly optimized `QQuickTextDocument` or a custom C++ text area (like `QScintilla` or custom).
- **Lines Numbers**: Left gutter.
- **Syntax Highlighting**: Fast regex or Tree-sitter integration. Must not block the UI thread on massive files.
- **Code Completion**: A floating popup list. Triggers on `Cmd+Space` or typing `.` or keywords.
- **Selection**: Cmd+Click for multiple cursors.
- **Vim Mode**: Block cursor, visual mode highlighting.

### 2.3 Data Grid (The Results View)
- Renders tabular data returned from queries. Needs to be a high-performance virtualized list/table (QML `TableView` or C++ `QTableView`).
- **Features**:
  - Resizeable columns.
  - Drag-to-reorder columns.
  - Sorting: Clicking header emits a sort signal (translating to `ORDER BY`).
  - Row numbering in the leftmost gutter.
  - **Inline Editing**: Double click a cell turns it into a `TextInput`.
  - **Visual Deltas**:
    - Edited cell text turns yellow.
    - New rows added via a "+" button appear at bottom with a green background.
    - Deleted rows appear struck-through with red background.
- **Bottom Status Bar**: Shows "Rows x-y of Total", "Query took 0.045s", Pagination controls. "Save Changes" and "Discard" buttons appear when edits exist.

### 2.4 Structure / DDL Tab
Presents table metadata.
- A tabbed sub-view: `Columns` | `Indexes` | `Foreign Keys` | `DDL`.
- Columns is a grid: Name, Type (ComboBox), Nullable (Checkbox), Auto Inc (Checkbox), Default Value (Text).

### 2.5 Export / Import Modals
- **Export**: Picker for format (CSV, JSON, SQL, XLSX, Markdown). Options toggles for "Include Headers", "Separator", "Quote Char". Destination file picker.
- **Import**: Source file picker. Target Table dropdown. Schema mapping UI (Matching CSV column index to Table Column Name). "Start Import" button with progress bar. 

## 3. Keyboard Shortcuts / Interactions
- **Cmd+T**: New Query Tab.
- **Cmd+W**: Close current tab.
- **Cmd+1...9**: Switch to nth tab.
- **Cmd+R / F5**: Execute current statement.
- **Cmd+Shift+R**: Execute all statements.
- **Cmd+E**: Run Explain plan.
- **Cmd+D**: Duplicate selected row(s).
- **Backspace/Delete**: Delete selected row(s) (Mark for deletion).
- **Cmd+S**: Commit pending data changes.
- **Cmd+K**: Quick Switcher / Command Palette (Floating middle screen popup to search tables rapidly). 

## 4. Design Aesthetics
- The app should natively follow macOS design guidelines (vibrancy, visual effect blurs, SF Symbols) despite being cross-platform with Qt.
- Utilize QML's capabilities for smooth transitions (e.g., expanding sidebar, fading in modals, bouncy scroll interactions).
- Deep Dark Mode support, ensuring contrast for syntax highlighting.
