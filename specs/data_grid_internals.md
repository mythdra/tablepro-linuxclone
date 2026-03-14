# Data Grid Internals (AG Grid + React)

## Overview
The Data Grid replaces Swift's custom `NSTableView` wrapper with **AG Grid Community** — a battle-tested React grid component supporting millions of rows via virtual scrolling.

## 1. Virtual Scrolling
- AG Grid only renders DOM rows visible in the viewport (~30-50 rows)
- As the user scrolls, rows are recycled — keeping DOM node count constant
- **No custom virtualization code needed** — AG Grid handles this natively
- For truly massive datasets (100K+ rows), use AG Grid's **Server-Side Row Model**: React requests pages from Go backend on demand

## 2. Column Definitions
```typescript
const columnDefs: ColDef[] = columns.map(col => ({
  field: col.name,
  headerName: col.name,
  sortable: true,
  resizable: true,
  editable: isEditable && !col.isPrimaryKey,
  cellRenderer: getCellRenderer(col.type), // NULL, JSON, date renderers
  cellStyle: (params) => getCellStyle(params, pendingChanges),
}));
```

## 3. Sorting
- Click column header → `onSortChanged` callback fires
- React sends sort params to Go: `QueryManager.ExecuteWithSort(tabID, column, direction)`
- Go rebuilds query with `ORDER BY` and re-executes
- Grid refreshes with new data (not client-side sort)

## 4. Inline Editing
- Double-click cell → AG Grid activates inline editor
- On cell value change: `onCellValueChanged` → calls `DataChangeManager.UpdateCell()`
- Go tracks the delta (original vs new value)
- Cell style updates to yellow background indicating "pending change"

## 5. Visual Deltas
```typescript
function getCellStyle(params, pendingChanges) {
  const key = `${params.rowIndex}:${params.colDef.field}`;
  if (pendingChanges.updated[key]) return { backgroundColor: '#FEF3C7' };  // Yellow
  if (pendingChanges.insertedRows.has(params.rowIndex)) return { backgroundColor: '#D1FAE5' }; // Green
  if (pendingChanges.deletedRows.has(params.rowIndex)) return {
    backgroundColor: '#FEE2E2', textDecoration: 'line-through' // Red
  };
  return {};
}
```

## 6. Row Operations
| Action | Trigger | Backend Call |
|---|---|---|
| Add Row | "+" button in status bar | `DataChangeManager.InsertRow(tabID)` |
| Delete Row | Select + Delete key | `DataChangeManager.DeleteRow(tabID, rowIndices)` |
| Duplicate Row | Cmd+D | `DataChangeManager.DuplicateRow(tabID, rowIndex)` |
| Commit | Cmd+S | `DataChangeManager.Commit(tabID)` → SQL execution |
| Discard | Discard button | `DataChangeManager.Discard(tabID)` → restore originals |

## 7. Pagination
- Status bar shows: `Rows 1-500 of 12,345 | 0.045s`
- Next/Previous page buttons call `QueryManager.ExecuteWithOffset(tabID, newOffset)`
- Go modifies query with `LIMIT {pageSize} OFFSET {newOffset}`
- AG Grid replaces data with new page

## 8. Copy Operations
- Copy cell: `Cmd+C` on selected cell → clipboard as text
- Copy row: Select row → copy all column values as tab-separated
- Copy column: Right-click header → "Copy Column Values"
- Copy as INSERT: Right-click → generates INSERT statement for selected rows

## 9. NULL Handling
- NULL values rendered with a distinctive style: italic gray "NULL" text
- Editing a NULL cell: empty string saves as empty string, a special "Set NULL" button restores NULL
- AG Grid custom cell renderer handles this distinction

## Performance vs Swift NSTableView
| Aspect | Swift (NSTableView) | Go+React (AG Grid) |
|---|---|---|
| Rendering | Native AppKit | WebView DOM (virtual) |
| 10K rows | Instant | Instant (virtual scroll) |
| 1M rows | Instant | Smooth with Server-Side model |
| Memory | Native structs | JSON transfer overhead |
| Editing | Custom NSTextField | AG Grid inline editor |
| Overall | Slightly faster | More features, cross-platform |
