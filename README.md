# TablePro

Cross-platform database client built with Go + Wails + React.

## Quick Start

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Install dependencies
go mod download
cd frontend && npm install

# Run development
wails dev
```

## Requirements

- Go 1.21+
- Node.js 18+
- Wails v2.8+

## Commands

```bash
# Development
wails dev

# Build for current platform
wails build

# Build for all platforms
make build-all

# Run tests
make test

# Run linters
make lint

# Clean build artifacts
make clean

# Run integration tests (requires databases)
go test -tags=integration ./internal/query/...

# Run frontend integration tests
cd frontend && npm run test:integration

# Run E2E tests
npx playwright test
```

## Project Structure

```
tablepro/
├── cmd/          # Go entry point
├── internal/     # Go packages (drivers, services)
├── frontend/     # React + TypeScript frontend
├── build/        # Build output
└── wails.json    # Wails configuration
```

## Features

- 🚀 Cross-platform (macOS, Windows, Linux)
- 🔒 Secure password storage (OS Keychain)
- 📊 8 database drivers (PostgreSQL, MySQL, SQLite, DuckDB, MSSQL, ClickHouse, MongoDB, Redis)
- 🎨 Modern UI with AG Grid and Monaco Editor
- 🔌 SSH tunneling and SSL/TLS support

## Phase 5: Query Execution

### Query Editor

- Monaco Editor-based SQL editor with syntax highlighting
- Schema-aware autocomplete (tables, columns, keywords)
- Multi-tab support with dirty state tracking
- Execute full query or selection (Ctrl/Cmd+Enter)
- Query formatting with Shift+Alt+F

### Result Pagination

- Server-side pagination for large datasets
- Configurable page size (10-10000 rows)
- Fast navigation with OFFSET/LIMIT
- Estimated vs exact row counts

### Query History

- Per-connection query history tracking
- Search and filter history entries
- Click to reload queries in editor
- Automatic deduplication
- Configurable history limit (default: 50)

### Usage Example

```typescript
// Execute a query
const result = await executeQuery(connectionId, 'SELECT * FROM users LIMIT 100');

// Navigate paginated results
const page2 = await executeQuery(connectionId, 'SELECT * FROM users LIMIT 100 OFFSET 100');

// Load from history
const history = getHistory(connectionId);
loadQuery(history[0].query);
```

### Testing

```bash
# Backend integration tests (requires PostgreSQL/MySQL)
export INTEGRATION_TEST=1
go test -tags=integration ./internal/query/...

# Frontend integration tests
cd frontend && npm run test:integration

# E2E tests (requires app running)
npx playwright install
npx playwright test
```

See [tests/README.md](tests/README.md) for detailed testing documentation.

## License

MIT
