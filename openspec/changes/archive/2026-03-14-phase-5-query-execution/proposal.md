## Why

Phase 4 (Database Drivers) đã hoàn thành với 8 drivers, nhưng chưa có cách để người dùng viết và thực thi queries. Phase 5 thêm Query Execution pipeline - cho phép users viết SQL trong Monaco Editor, execute queries với timeout/cancellation, xem results với pagination, và tracking query history.

## What Changes

- **New**: Monaco Editor integration với SQL syntax highlighting cho PostgreSQL/MySQL dialects
- **New**: QueryExecutor service với context timeout và cancellation support
- **New**: Query pagination (LIMIT/OFFSET cho SQL, cursor-based cho NoSQL)
- **New**: ResultSet handling với type mapping và formatting
- **New**: In-memory query history tracking (last N queries per connection)
- **Modified**: SQL Editor spec - thêm autocomplete/intellisense, multi-statement execution, result streaming
- **Modified**: Query History spec - thêm in-memory tracking trước khi có persistent storage (Phase 11)

## Capabilities

### New Capabilities
<!-- Capabilities being introduced. Replace <name> with kebab-case identifier -->
- `query-execution-pipeline`: QueryExecutor service, Execute() với context timeout, query cancellation, multi-statement queries, result streaming
- `result-set-handling`: ResultSet struct với metadata, column type mapping, NULL handling, data formatting (dates, numbers, booleans), multiple result sets support
- `query-pagination`: LIMIT/OFFSET pagination cho SQL databases, cursor-based pagination cho NoSQL, page size configuration, total count estimation

### Modified Capabilities
<!-- Existing capabilities whose REQUIREMENTS are changing -->
- `sql-editor`: Thêm autocomplete/intellisense, keyboard shortcuts (Ctrl+Enter), query cancellation UI, multi-statement execution support, result streaming integration
- `query-history`: Thêm in-memory tracking với timestamps, query deduplication, last N queries per connection (persistent storage với SQLite FTS5 là Phase 11)

## Impact

**Affected Code:**
- Backend: `internal/query/` (QueryExecutor, ResultSet, PaginationService)
- Frontend: `frontend/src/components/QueryEditor.tsx`, `frontend/src/components/ResultView.tsx`
- Stores: `queryStore.ts` (new)
- Wails bindings: QueryExecutor methods expose sang frontend

**Dependencies:**
- `@monaco-editor/react` ^4.6.0 (new)
- `@types/monaco` (new)
- Phase 4: Database Drivers (completed)

**Systems:**
- Query execution flow: Editor → QueryExecutor → Driver → ResultView
- Memory: In-memory history tracking (cleared on app restart)
- Events: `query:executing`, `query:completed`, `query:failed`, `query:cancelled`
