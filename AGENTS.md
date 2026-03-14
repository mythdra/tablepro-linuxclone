# TablePro - AGENTS.md

Development guide for AI agents working on TablePro (Go + Wails + React database client).

## Project Overview

TablePro is a cross-platform database client targeting macOS, Windows, and Linux as a single binary (~15-20MB).

**Architecture:**
- **Backend**: Go with Wails v2 runtime
- **Frontend**: React + TypeScript rendered in Wails WebView
- **Communication**: Wails RPC (Go methods bound to TypeScript) + Events (pub/sub)
- **Specs**: All specifications in `/specs/` and `/openspec/specs/`

## Build & Development Commands

### Project Setup
```bash
# Initialize Go module (when starting implementation)
go mod init github.com/tablepro/tablepro

# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Initialize Wails project
wails init -n tablepro -t react

# Install frontend dependencies
cd frontend && npm install
```

### Build Commands
```bash
# Development mode with hot reload
wails dev

# Production build
wails build

# Production build for specific platforms
wails build -platform darwin
wails build -platform windows
wails build -platform linux
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run single test
go test -run TestName ./internal/driver/

# Run tests with verbose output
go test -v ./...

# Frontend tests (when implemented)
cd frontend && npm test
```

### Linting & Formatting
```bash
# Go formatting
go fmt ./...

# Go linting
go vet ./...
golangci-lint run

# Frontend formatting
cd frontend && npm run format

# Frontend linting
cd frontend && npm run lint
```

## Go Coding Conventions

### Import Organization
```go
import (
    // Standard library
    "context"
    "database/sql"
    "encoding/json"
    "fmt"

    // Third-party packages
    "github.com/jackc/pgx/v5"
    "github.com/wailsapp/wails/v2/pkg/runtime"

    // Internal packages
    "github.com/tablepro/tablepro/internal/driver"
    "github.com/tablepro/tablepro/internal/connection"
)
```

### Naming Conventions
- **Types/Structs**: PascalCase (e.g., `DatabaseConnection`, `QueryTab`)
- **Interfaces**: PascalCase, often with `-er` suffix (e.g., `DatabaseDriver`, `Formatter`)
- **Functions/Methods**: PascalCase for exported, camelCase for private
- **Variables**: camelCase, descriptive names
- **Constants**: PascalCase or ALL_CAPS for compile-time constants
- **Files**: snake_case lowercase (e.g., `connection_manager.go`)

### Error Handling
```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("connection failed: %w", err)
}

// Use context with timeout for all DB operations
ctx, cancel := context.WithTimeout(parentCtx, time.Duration(timeout)*time.Second)
defer cancel()

// Never ignore errors
_, err = driver.Execute(query)
if err != nil {
    log.Printf("Query execution failed: %v", err)
    return err
}
```

### Error Handling Best Practices

**1. Error Wrapping (Go)**
- Always wrap errors with context using `fmt.Errorf("operation failed: %w", err)`
- Use `%w` verb for wrapping to enable `errors.Is()` and `errors.As()`
- Add meaningful context: what operation failed, not just the error

**2. Context Timeouts**
- All database operations MUST use context with timeout
- Default timeout: 30 seconds for queries
- Use `context.WithTimeout()` and always call `defer cancel()`
- Handle `context.DeadlineExceeded` and `context.Canceled` appropriately

**3. User-Friendly Error Messages**
```go
// Bad: cryptic error
return fmt.Errorf("dial tcp: connection refused")

// Good: actionable error
return fmt.Errorf("database connection failed: host %s port %d - check if database is running", host, port)
```

**4. Frontend Error Display (TypeScript)**
```typescript
// Extract user-friendly message from wrapped errors
function getUserMessage(error: Error): string {
  if (error.message.includes('connection refused')) {
    return 'Cannot connect to database. Check host and port.';
  }
  if (error.message.includes('timeout')) {
    return 'Query timed out. Try a smaller dataset or increase timeout.';
  }
  return error.message;
}

// Always log full error for debugging
console.error('Query failed:', error);
showToast(getUserMessage(error), 'error');
```

**5. Error Recovery Patterns**
- Retry transient errors (network timeouts) with exponential backoff
- Fail fast on permanent errors (authentication, syntax errors)
- Provide recovery actions in UI: "Retry", "Edit Query", "Close Connection"

**6. Testing Error Paths**
```go
// Test error wrapping
t.Run("error includes context", func(t *testing.T) {
    err := executor.Execute(ctx, connID, failingDriver, query)
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    if !strings.Contains(err.Error(), "query execution failed") {
        t.Errorf("error missing context: %v", err)
    }
})

// Test timeout handling
t.Run("timeout cancels query", func(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    err := executor.Execute(ctx, connID, slowDriver, query)
    if !errors.Is(err, context.DeadlineExceeded) {
        t.Errorf("expected timeout error, got: %v", err)
    }
})
```

### Struct Definition Style
```go
type DatabaseConnection struct {
    ID       uuid.UUID        `json:"id"`
    Name     string           `json:"name"`
    Type     DatabaseType     `json:"type"`
    Group    string           `json:"group"`
    ColorTag string           `json:"colorTag"`

    // Core connection
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Database string `json:"database"`
    Username string `json:"username"`
    // Password NEVER in struct - stored in OS Keychain
}
```

### Concurrency Patterns
```go
// Use sync.RWMutex for shared state
type ConnectionManager struct {
    mu    sync.RWMutex
    conns map[uuid.UUID]*Connection
}

// Goroutine with context
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    // ... work
}()
```

### JSON Tags
- Always use `json` tags on exported struct fields
- Use camelCase for JSON field names
- Use `-` tag to exclude sensitive fields (passwords, API keys)

### Logging
```go
// Use structured logging (log/slog or zap)
log.Printf("Connection established: host=%s, port=%d", host, port)

// For Wails events (debugging)
runtime.EventsEmit(ctx, "debug:info", map[string]any{
    "message": "Query executed",
    "duration": executionTime,
})
```

## TypeScript/React Conventions

### Import Organization
```typescript
// React and libraries
import React, { useState, useEffect } from 'react'

// Wails bindings
import { ConnectionManager } from '../wailsjs/go/main'

// Local components
import { DataGrid } from './DataGrid'
import { Toolbar } from './Toolbar'

// Types and utilities
import type { DatabaseConnection } from '../types'
import { formatDate } from '../utils/format'
```

### State Management (Zustand)
```typescript
import { create } from 'zustand'

interface ConnectionStore {
  connections: DatabaseConnection[]
  activeConnection: uuid.UUID | null
  addConnection: (conn: DatabaseConnection) => void
  setActiveConnection: (id: uuid.UUID) => void
}

export const useConnectionStore = create<ConnectionStore>((set) => ({
  connections: [],
  activeConnection: null,
  addConnection: (conn) => set((state) => ({
    connections: [...state.connections, conn]
  })),
  setActiveConnection: (id) => set({ activeConnection: id })
}))
```

### Wails RPC Calls
```typescript
// Go methods are async in TypeScript
async function saveConnection(conn: DatabaseConnection) {
  try {
    const result = await ConnectionManager.Save(conn)
    return result
  } catch (err) {
    console.error('Failed to save connection:', err)
    throw err
  }
}

// Wails Events (pub/sub)
import { EventsOn, EventsOff } from '../wailsjs/runtime'

useEffect(() => {
  EventsOn('connection:status', (status: ConnectionStatus) => {
    setStatus(status)
  })
  return () => EventsOff('connection:status')
}, [])
```

### Component Patterns
- Use functional components with hooks
- TypeScript interfaces for props
- Tailwind CSS for styling
- Lucide React for icons

## Architecture Patterns

### Module Organization
```
cmd/
  main.go              # Wails entry point
internal/
  driver/              # Database driver interface + implementations
  connection/          # Connection CRUD, Keychain, URL parser
  session/             # Active connection sessions
  query/               # Query builder, pagination, dialects
  change/              # Change tracking, SQL generation
  export/              # Export/Import services
  history/             # Query history (SQLite FTS5)
  settings/            # App preferences
  tab/                 # Tab persistence
  ssh/                 # SSH tunnel management
  license/             # License validation
frontend/
  src/
    components/        # React UI components
    stores/            # Zustand stores
    hooks/             # Custom React hooks
    lib/               # Utilities
```

### Key Design Decisions
- **No REST API**: Wails provides native IPC over WebView
- **Password Security**: OS Keychain via `go-keyring`
- **Query History**: Embedded SQLite with FTS5
- **Tab State**: JSON files per connection UUID
- **Large Datasets**: AG Grid with server-side row model

## Testing Guidelines

### Go Tests
```go
func TestConnectionManager_Save(t *testing.T) {
    // Arrange
    manager := NewConnectionManager()
    conn := &DatabaseConnection{
        Name: "Test DB",
        Type: "postgres",
    }

    // Act
    err := manager.Save(conn)

    // Assert
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
}

// Table-driven tests for drivers
func TestDriver_Query(t *testing.T) {
    tests := []struct {
        name    string
        query   string
        wantErr bool
    }{
        {"valid SELECT", "SELECT 1", false},
        {"invalid syntax", "SELEC 1", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

### Frontend Tests
```typescript
import { render, screen, fireEvent } from '@testing-library/react'

test('saves connection on form submit', async () => {
  render(<ConnectionForm />)
  
  fireEvent.change(screen.getByLabelText('Name'), {
    target: { value: 'Test DB' }
  })
  
  fireEvent.click(screen.getByText('Save'))
  
  expect(await screen.findByText('Connection saved')).toBeInTheDocument()
})
```

## Git & Workflow

### Branch Naming
- `feature/connection-manager`
- `bugfix/query-pagination`
- `refactor/driver-interface`
- `docs/api-reference`

### Commit Messages
```
feat: add PostgreSQL driver with SSH tunnel support
- Implement pgx-based driver
- Add SSH tunnel via golang.org/x/crypto/ssh
- Store passwords in OS Keychain
```

### Pre-commit Checklist
- [ ] `go fmt ./...` applied
- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes
- [ ] TypeScript compiles without errors
- [ ] ESLint passes

## AI Agent Guidelines

1. **Read specs first**: Always check `/specs/` before implementing
2. **Match patterns**: Follow existing code style exactly
3. **No type suppression**: Never use `as any` or `@ts-ignore`
4. **Context everywhere**: All DB operations need context with timeout
5. **No password in structs**: Always use OS Keychain
6. **Error wrapping**: Always wrap errors with context
7. **Test before complete**: Run `go test` on changed packages

## Cursor/Copilot Rules

No Cursor rules (`.cursor/rules/` or `.cursorrules`) or GitHub Copilot instructions (`.github/copilot-instructions.md`) exist in this repository yet.

## Key Dependencies

### Go
- `github.com/wailsapp/wails/v2` - Desktop framework
- `github.com/jackc/pgx/v5` - PostgreSQL
- `github.com/go-sql-driver/mysql` - MySQL
- `github.com/mattn/go-sqlite3` - SQLite
- `golang.org/x/crypto/ssh` - SSH tunneling
- `github.com/zalando/go-keyring` - OS Keychain

### React
- `@ag-grid-community/react` - Data grid
- `@monaco-editor/react` - SQL editor
- `zustand` - State management
- `@radix-ui/react-*` - UI primitives
- `tailwindcss` - Styling
