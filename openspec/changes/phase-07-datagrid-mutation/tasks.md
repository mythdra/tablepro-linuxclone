## 1. AG Grid Setup & Integration

- [ ] 1.1 Install @ag-grid-community/react and @ag-grid-community/styles dependencies
- [ ] 1.2 Create DataGrid React component wrapper with AgGridReact
- [ ] 1.3 Configure basic column definitions from query result metadata
- [ ] 1.4 Implement virtual scrolling with rowBuffer and rowHeight settings
- [ ] 1.5 Add NULL value cell renderer (gray italic "NULL" text)
- [ ] 1.6 Create data type-specific cell formatters (date, number, boolean)

## 2. Server-Side Row Model

- [ ] 2.1 Implement getRows callback for AG Grid Server-Side Row Model
- [ ] 2.2 Add LIMIT/OFFSET pagination parameters to QueryExecutor
- [ ] 2.3 Implement server-side sorting with ORDER BY clause generation
- [ ] 2.4 Add pagination controls UI (Next/Previous/Page Size selector)
- [ ] 2.5 Implement status bar showing "Rows X-Y of Z"
- [ ] 2.6 Add estimated row count via EXPLAIN or COUNT(*) subquery

## 3. Inline Editing Implementation

- [ ] 3.1 Enable cell editing in AG Grid with editable: true callback
- [ ] 3.2 Implement double-click and Enter key to activate edit mode
- [ ] 3.3 Add data type validation for edited values (integer, date, boolean)
- [ ] 3.4 Implement primary key column edit prevention
- [ ] 3.5 Add NULL value editing with "Set NULL" button
- [ ] 3.6 Implement Escape key to cancel edit

## 4. Change Tracking Backend

- [ ] 4.1 Create DataChangeManager struct in internal/change/
- [ ] 4.2 Implement CellChange and TabChanges data structures
- [ ] 4.3 Implement UpdateCell(tabID, rowIdx, colName, newValue) method
- [ ] 4.4 Implement InsertRow(tabID, rowData) method
- [ ] 4.5 Implement DeleteRow(tabID, primaryKeyValues) method
- [ ] 4.6 Implement GetPendingChanges(tabID) method
- [ ] 4.7 Implement DiscardChanges(tabID) method
- [ ] 4.8 Implement Undo() and Redo() with stack management

## 5. Visual Change Indicators

- [ ] 5.1 Add yellow background styling for modified cells (#FEF3C7)
- [ ] 5.2 Add green background styling for new rows (#D1FAE5)
- [ ] 5.3 Add red background with strikethrough for deleted rows (#FEE2E2)
- [ ] 5.4 Implement cell tooltip showing original vs new value
- [ ] 5.5 Add change count badge in tab header
- [ ] 5.6 Implement Commit/Discard button enable/disable logic

## 6. SQL Generation Engine

- [ ] 6.1 Create SQLStatementGenerator struct in internal/change/
- [ ] 6.2 Implement DialectParamMarker for parameter placeholders
- [ ] 6.3 Generate UPDATE statements from CellChange objects
- [ ] 6.4 Generate INSERT statements from InsertedRow objects
- [ ] 6.5 Generate DELETE statements from deleted row primary keys
- [ ] 6.6 Handle composite primary keys in WHERE clauses
- [ ] 6.7 Exclude auto-increment columns from INSERT statements
- [ ] 6.8 Implement batch statement ordering (DELETE → UPDATE → INSERT)

## 7. Commit/Rollback Implementation

- [ ] 7.1 Implement CommitChanges(ctx, sessionID) method in DataChangeManager
- [ ] 7.2 Wrap all statements in database transaction
- [ ] 7.3 Handle transaction rollback on any statement failure
- [ ] 7.4 Implement foreign key constraint error detection
- [ ] 7.5 Add statement-level error reporting with position
- [ ] 7.6 Clear change tracker on successful commit
- [ ] 7.7 Emit "data:saved" Wails event on success

## 8. SQL Preview UI

- [ ] 8.1 Add Monaco Editor panel for SQL preview (read-only)
- [ ] 8.2 Display generated SQL statements with syntax highlighting
- [ ] 8.3 Show change summary: "N UPDATEs, M INSERTs, K DELETEs"
- [ ] 8.4 Add Commit and Discard buttons in change panel
- [ ] 8.5 Implement loading state during commit (spinner + "Committing...")

## 9. Undo/Redo Frontend

- [ ] 9.1 Implement Ctrl+Z keyboard shortcut for undo
- [ ] 9.2 Implement Ctrl+Y keyboard shortcut for redo
- [ ] 9.3 Add undo/redo button toolbar actions
- [ ] 9.4 Implement undo stack limit (100 actions max)
- [ ] 9.5 Add toast notification when undo limit reached

## 10. Close Tab with Pending Changes

- [ ] 10.1 Add confirmation dialog when closing tab with changes
- [ ] 10.2 Implement "You have N uncommitted changes" message
- [ ] 10.3 Discard changes on tab close confirmation
- [ ] 10.4 Prevent tab close on Cancel button click

## 11. Integration Tests

- [ ] 11.1 Test cell edit → SQL generation for PostgreSQL
- [ ] 11.2 Test row insert → INSERT statement for MySQL
- [ ] 11.3 Test row delete → DELETE statement for SQLite
- [ ] 11.4 Test commit with FK constraint violation
- [ ] 11.5 Test rollback on statement failure
- [ ] 11.6 Test undo/redo stack behavior
- [ ] 11.7 Test pagination with 10,000+ row dataset

## 12. Cross-Dialect Testing

- [ ] 12.1 Test parameter markers ($1 vs ? vs @p1) for all dialects
- [ ] 12.2 Test identifier quoting ("col" vs `col` vs [col])
- [ ] 12.3 Test auto-increment column handling per dialect
- [ ] 12.4 Test NULL value representation in all drivers
- [ ] 12.5 Test transaction commit/rollback behavior

## 13. Documentation & Polish

- [ ] 13.1 Add AG Grid license note to README.md (MIT Community)
- [ ] 13.2 Document known limitations in phase README
- [ ] 13.3 Add code comments for complex SQL generation logic
- [ ] 13.4 Create troubleshooting guide for common edit errors
- [ ] 13.5 Add performance notes for large dataset handling
