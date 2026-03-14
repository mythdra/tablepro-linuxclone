## Context

TablePro currently has query execution working (Phase 5) and session management (Phase 6), but lacks the ability to display large datasets efficiently and mutate data through the UI. Users can execute SELECT queries but cannot interact with results beyond viewing. This phase bridges the gap between query execution and full database management capabilities.

**Constraints:**
- Must handle millions of rows without performance degradation
- Cross-platform compatibility (macOS, Windows, Linux)
- Single binary deployment (~15-20MB target)
- Support for 8 database dialects with different SQL syntax

**Stakeholders:**
- End users: Database developers, analysts, DBAs needing data editing
- Development team: Maintainable code with clear separation between grid UI and mutation logic

## Goals / Non-Goals

**Goals:**
- Integrate AG Grid Community for virtualized result display
- Implement server-side pagination via LIMIT/OFFSET for all supported databases
- Enable inline cell editing with change tracking
- Generate dialect-specific SQL (UPDATE/INSERT/DELETE) from tracked changes
- Support commit/rollback in transactions
- Provide undo/redo for edits before commit

**Non-Goals:**
- Advanced filtering beyond basic WHERE clauses (Phase 8+)
- Bulk data operations (import/export handled in Phase 9-10)
- Real-time collaboration or multi-user editing
- Custom cell editors beyond text input (date pickers, dropdowns in future phases)
- Change history tracking beyond current session

## Decisions

### 1. AG Grid Community vs Alternatives

**Decision:** Use AG Grid Community (free MIT license) over TanStack Table, Handsontable, or custom virtualization.

**Rationale:**
- Battle-tested with millions of rows via Server-Side Row Model
- Built-in virtualization, column resizing, sorting UI
- Active maintenance and large community
- MIT license fits commercial use
- React integration via `@ag-grid-community/react`

**Alternatives Considered:**
- **TanStack Table**: Headless, would require building virtualization from scratch
- **Handsontable**: Excel-like but commercial license expensive
- **Custom NSTableView/DataGrid**: Already tried in Swift version; reinventing the wheel

### 2. Server-Side Row Model Strategy

**Decision:** Implement pagination with LIMIT/OFFSET for all databases; defer cursor-based pagination to Phase 8.

**Rationale:**
- LIMIT/OFFSET universally supported across PostgreSQL, MySQL, SQLite, DuckDB, MSSQL, ClickHouse
- Simpler to implement than cursor-based pagination
- Adequate for MVP (100K rows with 100-row pages = 1000 page navigations)
- Can optimize later with keyset pagination if needed

**Implementation:**
```go
type PaginationParams struct {
    Limit  int `json:"limit"`   // Default: 100
    Offset int `json:"offset"`  // Default: 0
    SortBy string `json:"sortBy"`
    SortOrder string `json:"sortOrder"` // "ASC" | "DESC"
}

func (e *QueryExecutor) ExecutePaginated(ctx context.Context, sessionID uuid.UUID, query string, params PaginationParams) (*QueryResult, error)
```

### 3. Change Tracking Architecture

**Decision:** Track changes at the Go backend level with `DataChangeManager` per session tab, not in frontend state.

**Rationale:**
- Single source of truth for pending changes
- Prevents frontend/backend sync issues
- Enables undo/redo without complex CRDT logic
- SQL generation requires knowledge of original values and primary keys

**Data Structure:**
```go
type TabChanges struct {
    SessionID   uuid.UUID
    TabID       uuid.UUID
    TableName   string
    SchemaName  string
    PrimaryKeys []string
    
    CellChanges      map[string]*CellChange  // key: "rowIdx:columnName"
    InsertedRows     []*InsertedRow
    DeletedRowIDs    map[string]bool  // key: primary key value string
    
    UndoStack []ChangeAction
    RedoStack []ChangeAction
}
```

### 4. SQL Generation Pattern

**Decision:** Use `SQLStatementGenerator` with dialect-specific parameter placeholders.

**Rationale:**
- Each database driver already has parameter binding logic
- Generated SQL must match driver's placeholder style ($1 vs ? vs @p1)
- Generator receives changes + dialect, returns []string of SQL statements

**Dialect Mapping:**
```go
type DialectParamMarker int
const (
    PostgreSQLMarker DialectParamMarker = iota // $1, $2, ...
    MySQLMarker                                  // ?
    SQLiteMarker                                 // ?
    MSSQLMarker                                  // @p1, @p2, ...
    ClickHouseMarker                             // ?
)

func (g *SQLStatementGenerator) GenerateUpdate(change *CellChange, dialect DialectParamMarker) (string, []any, error)
```

### 5. Transaction Handling

**Decision:** Wrap all commits in database transactions; rollback on any error.

**Rationale:**
- Atomicity: all changes succeed or none do
- Foreign key constraints may cause later statements to fail
- User expectation: "Commit" means all changes saved

**Implementation:**
```go
func (dcm *DataChangeManager) Commit(ctx context.Context, sessionID uuid.UUID) error {
    tx, err := session.BeginTransaction(ctx)
    if err != nil {
        return fmt.Errorf("transaction start failed: %w", err)
    }
    
    // Execute all generated statements
    for _, stmt := range statements {
        if _, err := tx.ExecContext(ctx, stmt.SQL, stmt.Args...); err != nil {
            tx.Rollback()
            return fmt.Errorf("statement %d failed: %w", i, err)
        }
    }
    
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("transaction commit failed: %w", err)
    }
    
    // Clear change tracker on success
    dcm.ClearChanges(sessionID)
    return nil
}
```

### 6. Primary Key Requirement for Editing

**Decision:** Only allow inline editing for tables with at least one primary key column.

**Rationale:**
- UPDATE/DELETE requires WHERE clause with unique identifier
- Editing tables without PKs risks modifying multiple rows
- Clear user feedback: "This table has no primary key - editing disabled"

**Detection:**
```go
func (d *Driver) GetTableInfo(schema, table string) (*TableInfo, error) {
    // Query information_schema for primary key columns
    // If len(PrimaryKeys) == 0, disable editing in UI
}
```

### 7. NULL Value Representation

**Decision:** Display NULL as gray italic "NULL" text; distinguish from empty string.

**Rationale:**
- Critical distinction for database users
- Empty string (`""`) is a valid value; NULL means "unknown/absent"
- AG Grid cell renderer can customize NULL display

**Frontend:**
```typescript
function NullCellRenderer(params: ICellRendererParams) {
  if (params.value === null || params.value === undefined) {
    return '<span style="color: #9CA3AF; font-style: italic;">NULL</span>';
  }
  return params.value?.toString() ?? '';
}
```

## Risks / Trade-offs

### [Risk] AG Grid License Confusion
**Mitigation:** AG Grid Community is MIT licensed. Avoid using Enterprise features (row grouping, pivot tables) that require commercial license. Document allowed features in `README.md`.

### [Risk] LIMIT/OFFSET Performance on Large Tables
**Mitigation:** OFFSET becomes slow for deep pagination (OFFSET 1000000). Future optimization: implement keyset pagination using WHERE id > last_seen_id.

### [Risk] SQL Injection in Generated Statements
**Mitigation:** Never interpolate values directly. Always use parameterized queries with driver's placeholder syntax. Validate column names against allowlist from schema metadata.

### [Risk] Foreign Key Constraint Failures
**Mitigation:** Execute DELETE statements before UPDATE statements (order matters). Report constraint errors clearly: "Cannot delete row - referenced by table X".

### [Risk] Memory Growth from Change Tracking
**Mitigation:** Limit undo stack depth to 100 actions. Warn users if change tracker exceeds 1000 pending changes. Clear changes after successful commit.

### [Trade-off] Cell Editing vs Row-Level Editing
**Trade-off:** Chose cell-level editing (Excel-style) over row form editing. Faster for small edits, but less discoverable. Can add row editor modal in future.

### [Trade-off] Optimistic vs Pessimistic Locking
**Trade-off:** Using optimistic approach (no locks). If two users edit same row, last commit wins. Add last_modified timestamp checking in Phase 8 for conflict detection.

## Migration Plan

**Not Applicable** - This is new feature development, not a migration. No existing data or schemas need migration.

**Deployment Steps:**
1. Add `@ag-grid-community/react` to frontend dependencies
2. Implement `DataGrid` React component
3. Create Go backend services: `DataChangeManager`, `SQLStatementGenerator`
4. Update driver interfaces to support transaction methods
5. Add integration tests for each database dialect
6. Manual testing on PostgreSQL, MySQL, SQLite

**Rollback Strategy:**
- Feature flag not implemented (not needed for initial release)
- Rollback: revert git commit, users lose editing capability but retain query execution

## Open Questions

1. **Batch Size for Bulk Inserts:** What's the optimal batch size when inserting multiple rows? Start with 100, tune based on performance.

2. **Undo Stack Limit:** Should undo/redo persist across query executions? Current design: cleared on new query.

3. **Timestamp for Optimistic Locking:** Add `last_modified` column tracking for conflict detection? Defer to Phase 8.

4. **JSON Column Editing:** Should JSON columns use a special editor ( Monaco JSON editor)? Start with text input, enhance later.
