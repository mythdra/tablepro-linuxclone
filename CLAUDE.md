# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TablePro is a cross-platform database client built with Go + Wails v2 + React + TypeScript. It targets macOS, Windows, and Linux as a single binary (~15-20MB) and supports 8 database drivers: PostgreSQL, MySQL, SQLite, DuckDB, MSSQL, ClickHouse, MongoDB, and Redis.

## Build & Development Commands

```bash
# Development (hot reload)
wails dev

# Production build
wails build

# Build for all platforms
make build-all

# Run all tests
make test

# Run Go tests only
go test ./...

# Run specific Go test
go test -run TestName ./internal/driver/

# Run Go tests with verbose output
go test -v ./...

# Run frontend tests
cd frontend && npm test

# Run frontend tests once
cd frontend && npm run test:run

# Run frontend integration tests
cd frontend && npm run test:integration

# Run E2E tests (requires app running)
npx playwright test

# Run integration tests (requires databases)
go test -tags=integration ./internal/query/...

# Lint
make lint
```

## Requirements

- Go 1.21+
- Node.js 18+
- Wails v2.8+ (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

## Architecture

### Backend (Go)

```
internal/
├── driver/        # Database driver interface + 8 implementations
├── connection/    # Connection CRUD, OS Keychain, URL parser
├── session/       # Active connection sessions
├── query/         # Query executor, pagination, history
├── change/        # Change tracking with undo/redo, SQL generation
├── ssh/           # SSH tunnel management
└── ...
```

**Key Patterns:**
- `DatabaseDriver` interface in `internal/driver/driver.go` - all drivers implement this
- RPC methods in `app.go` are bound to frontend via Wails
- Passwords stored in OS Keychain via `go-keyring`, NEVER in structs
- All DB operations use `context.Context` with timeout

### Frontend (React + TypeScript)

```
frontend/src/
├── components/    # UI components (DataGrid, QueryEditor, ConnectionForm, etc.)
├── stores/        # Zustand stores (connectionStore, queryStore, changeStore, etc.)
├── hooks/         # Custom React hooks
├── types.ts       # Shared TypeScript types
└── wailsjs/       # Auto-generated Wails bindings
```

**Communication:**
- **RPC**: Call Go methods from TypeScript via auto-generated bindings in `wailsjs/go/main/App.js`
- **Events**: Use `EventsOn`/`EventsOff` from `wailsjs/runtime` for pub/sub

### Specifications

All specifications are in `/specs/`. Check relevant specs before implementing features. Key specs include:
- `architecture.md` - Overall system architecture
- `database_drivers.md` - Driver interface details
- `query_pipeline.md` - Query execution flow
- `data_mutation_rules.md` - Change tracking rules

## Key Dependencies

### Go
- `github.com/wailsapp/wails/v2` - Desktop framework
- `github.com/jackc/pgx/v5` - PostgreSQL
- `github.com/go-sql-driver/mysql` - MySQL
- `github.com/mattn/go-sqlite3` - SQLite
- `github.com/microsoft/go-mssqldb` - MSSQL
- `github.com/ClickHouse/clickhouse-go/v2` - ClickHouse
- `github.com/marcboeker/go-duckdb` - DuckDB
- `go.mongodb.org/mongo-driver` - MongoDB
- `github.com/redis/go-redis/v9` - Redis
- `github.com/zalando/go-keyring` - OS Keychain
- `golang.org/x/crypto/ssh` - SSH tunneling

### Frontend
- `@ag-grid-community/react` - Data grid with server-side pagination
- `@monaco-editor/react` - SQL editor with syntax highlighting
- `zustand` - State management
- `tailwindcss` - Styling
- `lucide-react` - Icons
- `react-hook-form` + `zod` - Form handling and validation

## Coding Conventions

### Go

- **Imports**: Standard library → Third-party → Internal packages
- **Naming**: PascalCase for exported, camelCase for private
- **Error Handling**: Always wrap errors with context using `fmt.Errorf("operation failed: %w", err)`
- **Context**: All DB operations MUST use context with timeout
- **Concurrency**: Use `sync.RWMutex` for shared state
- **JSON Tags**: Always use `json` tags with camelCase field names
- **Passwords**: NEVER include passwords in structs - use OS Keychain

### TypeScript/React

- Functional components with hooks
- TypeScript interfaces for props
- Zustand stores in `frontend/src/stores/`
- Wails bindings in `frontend/src/wailsjs/`

```typescript
// RPC call pattern
import { ExecuteQuery } from '../wailsjs/go/main/App'

const result = await ExecuteQuery(connectionId, queryStr)

// Event pattern
import { EventsOn, EventsOff } from '../wailsjs/runtime'

useEffect(() => {
  EventsOn('connection:status', (status) => setStatus(status))
  return () => EventsOff('connection:status')
}, [])
```

## Testing

### Go Tests
- Unit tests alongside source files
- Table-driven tests for multiple cases
- Integration tests require `-tags=integration` and running databases

### Frontend Tests
- Vitest for unit/integration tests
- Tests colocated with components (`.test.tsx` files)
- Playwright for E2E tests in `tests/e2e/`

## Phase Features

### Phase 5: Query Execution
- Monaco Editor with schema-aware autocomplete
- Server-side pagination with configurable page size
- Per-connection query history

### Phase 7: Data Grid Mutation
- Cell-level edit tracking with undo/redo (up to 100 actions)
- Dialect-aware SQL generation for 6 SQL databases
- Transaction-based commit with rollback on failure