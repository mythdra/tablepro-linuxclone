# Schema Sidebar Internals

This document details the mechanics of the left-hand Sidebar, which displays Database Tables, Views, and handles batch operations.

## 1. Lazy Loading & Caching (`TableFetcher`)
When a database connection succeeds, the sidebar does not instantly spam queries to fetch every table unless the user actually opens that connection tab.
- **Provider Mechanism**: `LiveTableFetcher` delegates to `SQLSchemaProvider`. It first checks an in-memory application cache to see if the table list was already fetched during this session.
- **Query Fallback**: If missing, it asks the active `PluginDatabaseDriver` to run its specific metadata query (e.g., querying `information_schema.tables` or `SHOW TABLES`).

## 2. Visual Representation & Badging (`TableRow`)
- **Icons**: Standard tables use a system `tablecells` icon (blue). Views use an `eye` icon (purple).
- **Status Badges**: If a user queues a Table for a destructive operation, an overlay badge appears over the icon:
  - **Pending Delete/Drop**: `minus.circle.fill` (Red)
  - **Pending Truncate**: `exclamationmark.circle.fill` (Orange)
- **Accessibility**: VoiceOver automatically merges the status badge state into the row description (e.g., "Table: USERS, pending delete").

## 3. Client-Side Search (`SidebarViewModel`)
- The "Filter Tables" search box does not issue SQL `LIKE` queries to the database.
- It operates strictly as an in-memory `.filter { $0.name.localizedCaseInsensitiveContains(debouncedSearchText) }`.
- Empty state transitions gracefully between "No tables exist in database" vs "No matching tables for search".

## 4. Batch Operations Pipeline
Users can multi-select (`Command`+Click or `Shift`+Click) rows in the `QTreeView`/`QListView`.
- **Queuing**: Right-clicking and selecting `Truncate` or `Delete` does *not* execute SQL. It adds the table names to `pendingTruncates` or `pendingDeletes` Sets.
- **Toggle Cancellation**: If a user selects tables that are *already* pending truncation, and clicks Truncate again, it silently un-queues them (removes from the Set).
- **Confirmation Dialog**: Adding a new pending operation spawns a `TableOperationDialog`, asking for parameters (e.g., "CASCADE" options for PostgreSQL).

## 5. Context Menu Specifications
The Context Menu explicitly disables actions depending on context (e.g., `isReadOnly` Safe Mode, or whether the item is a View).
Required Actions:
1. **Create New View**: Always available unless `isReadOnly`.
2. **Edit View Definition**: Only visible if clicked item `type == .view`.
3. **Show Structure**: Opens the table configuration tab instead of the data grid.
4. **Copy Name**: Comma-separates all selected table names to clipboard.
5. **Export...**: Spawns Export Dialog for the selected tables.
6. **Import...**: Only visible if the Plugin Driver explicitly `supportsImport()`.
7. **Truncate**: Disabled for Views. Available for multi-selection.
8. **Drop View / Delete Table**: Contextually renamed based on entity type.

## Qt/C++ Migration Guidelines
- Use `QTreeView` or `QListView` with a `QSortFilterProxyModel` for the client-side filtering.
- Context menus should be built dynamically in `contextMenuEvent` intercepting the clicked `QModelIndex`.
- Badges and Text Colors (Red/Orange pending states) should be handled inside `QAbstractListModel::data()` returning colors for `Qt::ForegroundRole` and combining icons in `Qt::DecorationRole`.
