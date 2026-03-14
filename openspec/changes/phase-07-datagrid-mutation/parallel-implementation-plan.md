# Phase 7 Parallel Implementation Plan

## Worktree Distribution

### Worktree 1: `.worktrees/datagrid-setup` (feature/datagrid-setup)
**Focus**: AG Grid integration, server-side row model, basic display
**Tasks**: 1-6, 11 (18 tasks total)

#### Tasks:
- [ ] 1.1 Install @ag-grid-community/react and @ag-grid-community/styles dependencies
- [ ] 1.2 Create DataGrid React component wrapper with AgGridReact
- [ ] 1.3 Configure basic column definitions from query result metadata
- [ ] 1.4 Implement virtual scrolling with rowBuffer and rowHeight settings
- [ ] 1.5 Add NULL value cell renderer (gray italic "NULL" text)
- [ ] 1.6 Create data type-specific cell formatters (date, number, boolean)
- [ ] 2.1 Implement getRows callback for AG Grid Server-Side Row Model
- [ ] 2.2 Add LIMIT/OFFSET pagination parameters to QueryExecutor
- [ ] 2.3 Implement server-side sorting with ORDER BY clause generation
- [ ] 2.4 Add pagination controls UI (Next/Previous/Page Size selector)
- [ ] 2.5 Implement status bar showing "Rows X-Y of Z"
- [ ] 2.6 Add estimated row count via EXPLAIN or COUNT(*) subquery
- [ ] 11.7 Test pagination with 10,000+ row dataset

**Dependencies**: None (can start immediately)
**Merge Order**: 1st (foundation layer)

---

### Worktree 2: `.worktrees/change-tracking` (feature/change-tracking)
**Focus**: Backend change tracking, undo/redo, data structures
**Tasks**: 4, 5, 9, 10 (23 tasks total)

#### Tasks:
- [ ] 4.1 Create DataChangeManager struct in internal/change/
- [ ] 4.2 Implement CellChange and TabChanges data structures
- [ ] 4.3 Implement UpdateCell(tabID, rowIdx, colName, newValue) method
- [ ] 4.4 Implement InsertRow(tabID, rowData) method
- [ ] 4.5 Implement DeleteRow(tabID, primaryKeyValues) method
- [ ] 4.6 Implement GetPendingChanges(tabID) method
- [ ] 4.7 Implement DiscardChanges(tabID) method
- [ ] 4.8 Implement Undo() and Redo() with stack management
- [ ] 5.1 Add yellow background styling for modified cells (#FEF3C7)
- [ ] 5.2 Add green background styling for new rows (#D1FAE5)
- [ ] 5.3 Add red background with strikethrough for deleted rows (#FEE2E2)
- [ ] 5.4 Implement cell tooltip showing original vs new value
- [ ] 5.5 Add change count badge in tab header
- [ ] 5.6 Implement Commit/Discard button enable/disable logic
- [ ] 9.1 Implement Ctrl+Z keyboard shortcut for undo
- [ ] 9.2 Implement Ctrl+Y keyboard shortcut for redo
- [ ] 9.3 Add undo/redo button toolbar actions
- [ ] 9.4 Implement undo stack limit (100 actions max)
- [ ] 9.5 Add toast notification when undo limit reached
- [ ] 10.1 Add confirmation dialog when closing tab with changes
- [ ] 10.2 Implement "You have N uncommitted changes" message
- [ ] 10.3 Discard changes on tab close confirmation
- [ ] 10.4 Prevent tab close on Cancel button click

**Dependencies**: Requires datagrid-setup for DataGrid component interface
**Merge Order**: 2nd (builds on grid foundation)

---

### Worktree 3: `.worktrees/sql-generation` (feature/sql-generation)
**Focus**: SQL statement generation, dialect handling, commit/rollback
**Tasks**: 6, 7, 12 (20 tasks total)

#### Tasks:
- [ ] 6.1 Create SQLStatementGenerator struct in internal/change/
- [ ] 6.2 Implement DialectParamMarker for parameter placeholders
- [ ] 6.3 Generate UPDATE statements from CellChange objects
- [ ] 6.4 Generate INSERT statements from InsertedRow objects
- [ ] 6.5 Generate DELETE statements from deleted row primary keys
- [ ] 6.6 Handle composite primary keys in WHERE clauses
- [ ] 6.7 Exclude auto-increment columns from INSERT statements
- [ ] 6.8 Implement batch statement ordering (DELETE → UPDATE → INSERT)
- [ ] 7.1 Implement CommitChanges(ctx, sessionID) method in DataChangeManager
- [ ] 7.2 Wrap all statements in database transaction
- [ ] 7.3 Handle transaction rollback on any statement failure
- [ ] 7.4 Implement foreign key constraint error detection
- [ ] 7.5 Add statement-level error reporting with position
- [ ] 7.6 Clear change tracker on successful commit
- [ ] 7.7 Emit "data:saved" Wails event on success
- [ ] 12.1 Test parameter markers ($1 vs ? vs @p1) for all dialects
- [ ] 12.2 Test identifier quoting ("col" vs `col` vs [col])
- [ ] 12.3 Test auto-increment column handling per dialect
- [ ] 12.4 Test NULL value representation in all drivers
- [ ] 12.5 Test transaction commit/rollback behavior

**Dependencies**: Requires change-tracking for DataChangeManager interface
**Merge Order**: 3rd (depends on change tracking structures)

---

### Worktree 4: `.worktrees/ui-integration` (feature/ui-integration)
**Focus**: Frontend UI components, SQL preview, inline editing
**Tasks**: 3, 8 (11 tasks total)

#### Tasks:
- [ ] 3.1 Enable cell editing in AG Grid with editable: true callback
- [ ] 3.2 Implement double-click and Enter key to activate edit mode
- [ ] 3.3 Add data type validation for edited values (integer, date, boolean)
- [ ] 3.4 Implement primary key column edit prevention
- [ ] 3.5 Add NULL value editing with "Set NULL" button
- [ ] 3.6 Implement Escape key to cancel edit
- [ ] 8.1 Add Monaco Editor panel for SQL preview (read-only)
- [ ] 8.2 Display generated SQL statements with syntax highlighting
- [ ] 8.3 Show change summary: "N UPDATEs, M INSERTs, K DELETEs"
- [ ] 8.4 Add Commit and Discard buttons in change panel
- [ ] 8.5 Implement loading state during commit (spinner + "Committing...")

**Dependencies**: Requires datagrid-setup for AG Grid integration
**Merge Order**: 2nd (parallel with change-tracking)

---

### Worktree 5: `.worktrees/testing` (feature/testing)
**Focus**: Integration tests, cross-dialect validation
**Tasks**: 11, 13 (13 tasks total)

#### Tasks:
- [ ] 11.1 Test cell edit → SQL generation for PostgreSQL
- [ ] 11.2 Test row insert → INSERT statement for MySQL
- [ ] 11.3 Test row delete → DELETE statement for SQLite
- [ ] 11.4 Test commit with FK constraint violation
- [ ] 11.5 Test rollback on statement failure
- [ ] 11.6 Test undo/redo stack behavior
- [ ] 13.1 Add AG Grid license note to README.md (MIT Community)
- [ ] 13.2 Document known limitations in phase README
- [ ] 13.3 Add code comments for complex SQL generation logic
- [ ] 13.4 Create troubleshooting guide for common edit errors
- [ ] 13.5 Add performance notes for large dataset handling

**Dependencies**: Requires all other worktrees to complete first
**Merge Order**: 5th (last - validates everything)

---

## Merge Strategy

### Sequential Merge Order

```
1. datagrid-setup     (foundation: AG Grid + pagination)
       ↓
2. change-tracking    (backend: change management)
   ui-integration     (frontend: editing UI)
       ↓
3. sql-generation     (backend: SQL + commit)
       ↓
4. testing            (validation: integration tests)
```

### Merge Commands (after each worktree completes)

```bash
# After worktree 1 completes
git worktree remove .worktrees/datagrid-setup
git merge feature/datagrid-setup

# After worktrees 2 & 3 complete (parallel merge)
git worktree remove .worktrees/change-tracking
git worktree remove .worktrees/ui-integration
git merge feature/change-tracking
git merge feature/ui-integration

# After worktree 4 completes
git worktree remove .worktrees/sql-generation
git merge feature/sql-generation

# After worktree 5 completes
git worktree remove .worktrees/testing
git merge feature/testing
```

### Conflict Resolution Guidelines

1. **Go files**: Changes should be in separate packages (minimal conflicts expected)
2. **React components**: May have conflicts in App.tsx - resolve by combining features
3. **Dependencies**: package.json/go.mod changes should be cumulative
4. **Wails bindings**: Ensure all new methods are exported in app.go

---

## Agent Assignment

Each worktree will be handled by a dedicated `deep` category agent with:
- Clear task list from this document
- File path boundaries (which directories to modify)
- Integration points (what interfaces to match)
- Verification criteria (tests to pass)

---

## Synchronization Protocol

### Before Merge:
1. ✅ All tasks in worktree completed
2. ✅ `go test ./...` passes in worktree
3. ✅ `npm test` passes in frontend (if applicable)
4. ✅ `go fmt ./...` and `go vet ./...` applied
5. ✅ ESLint passes for TypeScript changes

### After Merge:
1. Run full test suite: `go test ./... && cd frontend && npm test`
2. Build verification: `wails build` (single platform)
3. Check for merge conflicts in git status
4. Resolve any TypeScript type errors from merged changes

---

## Progress Tracking

| Worktree | Branch | Tasks | Status | Merge Ready |
|----------|--------|-------|--------|-------------|
| datagrid-setup | feature/datagrid-setup | 13 | ⬜ Not Started | ⬜ No |
| change-tracking | feature/change-tracking | 23 | ⬜ Not Started | ⬜ No |
| sql-generation | feature/sql-generation | 20 | ⬜ Not Started | ⬜ No |
| ui-integration | feature/ui-integration | 11 | ⬜ Not Started | ⬜ No |
| testing | feature/testing | 13 | ⬜ Not Started | ⬜ No |

---

## Communication Protocol

- Each agent works independently in its worktree
- Agents report completion via task completion notification
- Main coordinator handles merges sequentially
- If conflicts detected during merge, agent that created conflict is notified for fix
