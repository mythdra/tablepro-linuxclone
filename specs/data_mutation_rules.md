# Data Mutation & SQL Generation Rules

This document outlines exactly how TablePro tracks edits made directly inside the Data Grid and converts them into safe, dialect-specific SQL statements. The Qt/C++ rewrite must implement these exact algorithms to prevent data loss or corruption during optimistic concurrency saves.

## 1. Change Tracking (`DataChangeManager`)
TablePro tracks changes at the cell level but groups them by row. There are three states: `INSERT`, `UPDATE`, and `DELETE`.

- **Internal Data Structures (O(1) Lookups):**
  - `changeIndex`: A hash map mapping a composite key `(rowIndex, ChangeType)` to an index in the main `changes` array for constant-time lookups.
  - `modifiedCells`: `[rowIndex: Set<columnIndex>]` tracks exactly which cells are dirty.
  - `insertedRowData`: `[rowIndex: [String?]]` stores the values of newly inserted rows sparsely until saved.

- **State Transitions:**
  - *Edit an untouched row*: Creates a new `UPDATE` RowChange.
  - *Edit a dirty row*: Mutates the existing `UPDATE` RowChange. If the new value matches the `oldValue` perfectly, the cell change is discarded. If all cell changes are discarded, the row is unmarked as dirty.
  - *Edit an `INSERT` row*: Mutates the in-memory sparse array directly; never spawns an `UPDATE` change.
  - *Delete an untouched row*: Mark as `DELETE`. Stores the `originalRow` values.
  - *Delete a dirty row*: Discards the pending `UPDATE`, marks as `DELETE` using the `originalRow` values.
  - *Delete an `INSERT` row*: Completely vanishes from memory; never hits the database.

- **Undo/Redo Stack (`DataChangeUndoManager`):**
  - Maintains a specific `UndoAction` enum (`cellEdit`, `rowInsertion`, `rowDeletion`, `batchRowDeletion`, `batchRowInsertion`).
  - Editing any new cell invalidates (clears) the Redo stack entirely.

## 2. SQL Generation (`SQLStatementGenerator`)
When the user clicks "Save Changes", the tracked changes are translated into SQL.

### General Rules
- **Type Casting & Strings**: The generator receives raw Strings. It must quote/escape them properly based on the dialect (e.g., `'O''Connor'` in SQL).
- **Function Detection**: If a user types raw SQL functions (e.g., `NOW()`, `UUID()`), the generator uses regex `isSQLFunctionExpression` to prevent parameterizing it as a string literal. It injects it straight into the query: `UPDATE tbl SET updated_at = NOW()`.
- **Default Check**: If the value strictly equals `"__DEFAULT__"`, it injects the `DEFAULT` SQL keyword instead of a parameterized string.

### INSERT Rules
- Generates: `INSERT INTO "table" ("col1", "col2") VALUES ($1, $2)`
- Skips columns where the value is untouched or explicitly marked `"__DEFAULT__"`.

### UPDATE Rules
- **With Primary Key (Ideal Path):**
  - Requires the core configuration to know the `primaryKeyColumn` for the table.
  - Statement: `UPDATE "table" SET "col2" = $1 WHERE "pk" = $2`
- **Without Primary Key (Optimistic Concurrency Path):**
  - If a table lacks a primary key (or it was excluded from the `SELECT`), the generator enforces optimistic locking by matching *every single column's original value*.
  - Statement: `UPDATE "table" SET "col2" = $1 WHERE "col1" = $2 AND "col2" = $3 AND "col3" IS NULL ...`
  - If multiple rows share the exact same raw data, updating one cell updates *all* of them (intentional fallback behavior).

### DELETE Rules
- **Batching:** If multiple rows are deleted and a Primary Key exists, groups them into a single `IN` or `OR` query to reduce network overhead.
  - Statement: `DELETE FROM "table" WHERE "pk" = $1 OR "pk" = $2`
- **Without Primary Key:**
  - Same as UPDATE. Maps every single original column value in the `WHERE` clause.
  - Statement: `DELETE FROM "table" WHERE "col1" = $1 AND "col2" IS NULL AND ...`

## 3. Plugin Dispatch Fallback
If the DatabaseType is handled by a secondary plugin (e.g., MongoDB, ClickHouse), the `generateSQL` method passes the tracked changes array via the C-Bridge into the Qt Plugin. The driver plugin then returns the raw raw string commands instead of using the built-in generic SQL generator.
