# Phase 7: Data Grid & Mutation

**Duration**: 3-4 weeks | **Priority**: 🟠 High | **Tasks**: 35

---

## Overview

Implement the data grid with AG Grid, including inline editing, change tracking, and SQL generation for commits.

---

## Key Features

1. **Virtual Scrolling**: Millions of rows via Server-Side Row Model
2. **Inline Editing**: Double-click to edit, visual change indicators
3. **Change Tracking**: Track edits, new rows, deleted rows
4. **SQL Generation**: Generate UPDATE/INSERT/DELETE from changes
5. **Commit/Rollback**: Transaction support for changes

---

## Task Summary

### 7.1 AG Grid Setup (6 tasks)
- [ ] 7.1.1 Install @ag-grid-community/react
- [ ] 7.1.2 Create DataGrid component wrapper
- [ ] 7.1.3 Configure virtual scrolling
- [ ] 7.1.4 Set up column definitions
- [ ] 7.1.5 Implement row data mapping
- [ ] 7.1.6 Handle NULL value rendering

### 7.2 Server-Side Row Model (5 tasks)
- [ ] 7.2.1 Implement getRows callback
- [ ] 7.2.2 Fetch data with LIMIT/OFFSET
- [ ] 7.2.3 Handle sorting server-side
- [ ] 7.2.4 Handle filtering server-side
- [ ] 7.2.5 Implement infinite scrolling

### 7.3 Inline Editing (6 tasks)
- [ ] 7.3.1 Enable cell editing in AG Grid
- [ ] 7.3.2 Track edited cells in ChangeTracker
- [ ] 7.3.3 Implement visual indicators (yellow)
- [ ] 7.3.4 Handle new rows (green)
- [ ] 7.3.5 Handle deleted rows (red strikethrough)
- [ ] 7.3.6 Implement undo/redo for edits

### 7.4 Change Tracking (6 tasks)
- [ ] 7.4.1 Create DataChangeManager struct
- [ ] 7.4.2 Track cell edits with original values
- [ ] 7.4.3 Track new rows
- [ ] 7.4.4 Track deleted rows
- [ ] 7.4.5 Implement DiscardChanges()
- [ ] 7.4.6 Implement GetPendingChanges()

### 7.5 SQL Generation (7 tasks)
- [ ] 7.5.1 Create SQLStatementGenerator
- [ ] 7.5.2 Generate UPDATE from edits
- [ ] 7.5.3 Generate INSERT from new rows
- [ ] 7.5.4 Generate DELETE from removed rows
- [ ] 7.5.5 Handle primary key in WHERE
- [ ] 7.5.6 Generate batch statements
- [ ] 7.5.7 Validate generated SQL

### 7.6 Commit Changes (5 tasks)
- [ ] 7.6.1 Implement CommitChanges() method
- [ ] 7.6.2 Wrap in transaction
- [ ] 7.6.3 Handle foreign key constraints
- [ ] 7.6.4 Report errors per statement
- [ ] 7.6.5 Clear change tracker on success
- [ ] 7.6.6 Emit success events

---

## Acceptance Criteria

- [ ] Grid displays millions of rows smoothly
- [ ] Inline editing works with visual feedback
- [ ] Changes tracked correctly
- [ ] SQL generation produces valid SQL
- [ ] Commit/rollback working
- [ ] Undo/redo for edits working

---

## Dependencies

← [Phase 6: Session Management](phase-06-sessions.md)  
→ [Phase 8: Tab Management](phase-08-tabs.md)
