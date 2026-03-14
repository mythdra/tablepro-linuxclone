# Data Mutation Rules (Go Backend)

## Overview
When users edit cells in the AG Grid, changes are tracked by the Go `DataChangeManager` and compiled into dialect-specific SQL statements on commit.

## Change Tracking Flow
```
React (AG Grid)               Go Backend
─────────────                  ──────────
User edits cell ──────────→  DataChangeManager.UpdateCell(tabID, rowIdx, colName, newVal)
User adds row   ──────────→  DataChangeManager.InsertRow(tabID, rowData)
User deletes row ─────────→  DataChangeManager.DeleteRow(tabID, rowIdentity)
User presses Cmd+S ───────→  DataChangeManager.Commit(tabID)
                               ↓
                           SQLStatementGenerator.Generate(changes, dialect)
                               ↓
                           Driver.Execute(statements) in a transaction
                               ↓
                           EventsEmit("data:saved") → React refreshes grid
```

## Cell Change Tracking
```go
type CellChange struct {
    RowIndex     int
    Column       string
    OriginalValue any
    NewValue      any
    PrimaryKey    map[string]any // identity for WHERE clause
}
```

## SQL Generation Rules (DialectProvider)

### UPDATE
```sql
-- PostgreSQL
UPDATE "schema"."table" SET "column" = $1 WHERE "pk_col" = $2;
-- MySQL
UPDATE `schema`.`table` SET `column` = ? WHERE `pk_col` = ?;
-- SQL Server
UPDATE [schema].[table] SET [column] = @p1 WHERE [pk_col] = @p2;
```

### INSERT
```sql
-- PostgreSQL
INSERT INTO "schema"."table" ("col1", "col2") VALUES ($1, $2);
-- MySQL
INSERT INTO `schema`.`table` (`col1`, `col2`) VALUES (?, ?);
```

### DELETE
```sql
-- PostgreSQL
DELETE FROM "schema"."table" WHERE "pk_col" = $1;
```

## Safe Mode Enforcement
```go
type SafeModeLevel int
const (
    SafeModeOff      SafeModeLevel = iota // No restrictions
    SafeModeRequireWhere                   // Block UPDATE/DELETE without WHERE
    SafeModeReadOnly                       // Block all mutations
)
```
- Before executing any generated SQL, `DataChangeManager.Commit()` checks the connection's `SafeMode` setting
- `SafeModeRequireWhere`: the generator always includes WHERE with primary key — this is inherently satisfied
- `SafeModeReadOnly`: `Commit()` returns error immediately

## Undo/Redo
- Go maintains `undoStack []ChangeAction` and `redoStack []ChangeAction`
- React calls `DataChangeManager.Undo(tabID)` / `DataChangeManager.Redo(tabID)`
- After undo/redo, Go emits updated pending changes to React for visual delta rendering
