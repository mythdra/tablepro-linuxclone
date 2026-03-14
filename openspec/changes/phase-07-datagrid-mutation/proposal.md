## Why

Phase 7 implements the core data interaction layer: a high-performance data grid with inline editing, change tracking, and SQL generation for committing mutations. This is the MVP milestone that transforms TablePro from a query viewer into a full database management tool where users can view, edit, and modify data directly.

## What Changes

- **AG Grid Integration**: Replace basic result display with AG Grid Community for virtual scrolling and column management
- **Server-Side Row Model**: Implement pagination and infinite scrolling for large datasets via LIMIT/OFFSET
- **Inline Editing**: Enable double-click cell editing with visual change indicators (yellow for edits, green for new rows, red for deleted)
- **Change Tracking System**: Track cell edits, new rows, and deleted rows with original value preservation
- **SQL Generation Engine**: Generate dialect-specific UPDATE/INSERT/DELETE statements from tracked changes
- **Commit/Rollback**: Transaction-based commit with foreign key constraint handling and error reporting
- **Undo/Redo**: Maintain undo/redo stacks for cell edits before commit

## Capabilities

### New Capabilities
- `data-grid`: AG Grid wrapper with virtual scrolling, column definitions, and server-side row model
- `inline-editing`: Cell editing with validation, change tracking, and visual delta indicators
- `change-tracking`: DataChangeManager for tracking cell edits, row insertions, and deletions
- `sql-generation`: SQLStatementGenerator for producing dialect-specific mutation SQL
- `commit-rollback`: Transaction-based commit with constraint handling and rollback support

### Modified Capabilities
- `session-management`: ADD tab-level change tracking state and pending changes query interface
- `query-pipeline`: ADD server-side sorting, filtering, and pagination parameters to query execution

## Impact

- **Frontend**: AG Grid React component integration, change tracking UI, commit/discord toolbar actions
- **Backend**: DataChangeManager service, SQLStatementGenerator, transaction handling in drivers
- **Dependencies**: AG Grid Community license (MIT), existing driver transaction support
- **Performance**: Server-side row model required for 100K+ row datasets
- **Testing**: Integration tests for SQL generation across all supported database dialects
