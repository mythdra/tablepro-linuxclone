# Query Execution Pipeline & Row Operations

This document exhaustively details the internal workings of the query builder, parsing engines, and row operation managers. The Qt/C++ rewrite must implement these specific capabilities identically for full feature parity.

## 1. Table Query Builder

The application dynamically generates SQL for standard browsing, sorting, and filtering operations. 

### Core Behaviors
- **Delegation to Plugins:** Before attempting to build native SQL, the builder first checks `pluginDriver.buildBrowseQuery()`, etc. If the specific database dialect overrides the query building process (e.g., MongoDB requires JSON strings instead of SQL), the core builder yields to the plugin.
- **Quote Sanitization**: Uses `pluginDriver.quoteIdentifier` or defaults to duplicating double quotes `""` around column/table names to prevent SQL injection or parsing errors with spaces in column names.
- **Sorting Insertion Algorithm (`buildSortedQuery`)**:
  - Automatically locates existing `ORDER BY`, `LIMIT`, and `OFFSET` clauses.
  - Strips the old `ORDER BY`.
  - Injects the new `ORDER BY "column" ASC` *before* `LIMIT` and *before* `OFFSET`, but *after* `WHERE`.
- **Search Capabilities**:
  - `buildQuickSearchQuery`: Used when the user types in the top-right search box. Generates a massive `WHERE (col1 LIKE '%x%' OR col2 LIKE '%x%')`.
  - `buildFilteredQuery`: Used for complex multi-column filters.

## 2. Row Operations Manager

This manager handles exactly what occurs when a user interacts with rows in the Data Grid.

### Batch Deletion Algorithms
Users can shift-click multiple rows and press `Delete`.
1. It separates the selection into two categories: `existingRows` (came from the DB) and `insertedRows` (created locally but not yet saved).
2. It sorts the deletion indices in *descending order* so that when it removes rows from the underlying model array, the indices don't shift out from under it.
3. `insertedRows` completely vanish from the UI and Undo stack immediately without SQL generation.
4. `existingRows` are batched into a single `changeManager.recordBatchRowDeletion` command for Undo/Redo purposes.
5. It then recalculates the *new cursor selection*, attempting to select the row immediately following the deleted clump.

### Clipboard "Paste" Pipeline
When a user copies cells from Excel and pastes them into TablePro:
1. **Auto-Detection (`detectParser`)**: Analyzes the raw string. If there are more `\t` (tabs) than `,` (commas), it utilizes the `TSVRowParser`. Otherwise, it uses `CSVRowParser`.
2. **RFC 4180 Parsing**: The CSV parser implements a state-machine that respects embedded commas inside quoted fields (`"Hello, World"`). it does not blindly split by `,`.
3. **Header Detection**: If the first row of the pasted data matches more than 50% of the destination Table's Column Names, the parser automatically discards the first row assuming it is a header row.
4. **Column Length Correction**:
   - If pasted data has fewer columns than the table: Pads the tail with `NULL` or default empty values.
   - If pasted data has more columns than the table: Truncates the trailing columns entirely.
5. **Primary Key Forgiveness**: If the destination table has an auto-increment or serial Primary Key, it automatically modifies the pasted value in that specific column index to `__DEFAULT__` regardless of what value the user copied. This ensures the database generates the key on Save, preventing unique constraint violations.

### Clipboard "Copy" Pipeline
When copying massive amounts of rows from TablePro:
- **OOM Protection**: Hardcoded to silently truncate anything over `50,000` rows maximum.
- Generates a Tab-Separated string (`\t` delimiter, `\n` linebreaks). Contains a trailing message warning if truncation occurred.
