# Coding Conventions

## Go Coding Standards

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
)
```

### Naming Conventions
- **Types/Structs**: PascalCase (`DatabaseConnection`)
- **Interfaces**: PascalCase, `-er` suffix (`DatabaseDriver`)
- **Functions**: PascalCase exported, camelCase private
- **Variables**: camelCase, descriptive
- **Files**: snake_case (`connection_manager.go`)

### Error Handling
```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("connection failed: %w", err)
}

// Context with timeout
ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
defer cancel()
```

### Struct Definition
```go
type DatabaseConnection struct {
    ID       uuid.UUID `json:"id"`
    Name     string    `json:"name"`
    Host     string    `json:"host"`
    Port     int       `json:"port"`
    // Password NEVER in struct - use Keychain
}
```

### Concurrency
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

---

## TypeScript/React Conventions

### Import Organization
```typescript
// React and libraries
import React, { useState, useEffect } from 'react'

// Wails bindings
import { ConnectionManager } from '../wailsjs/go/main'

// Local components
import { DataGrid } from './DataGrid'

// Types
import type { DatabaseConnection } from '../types'
```

### Component Patterns
- Functional components with hooks
- TypeScript interfaces for props
- Tailwind CSS for styling
- Lucide React for icons

### State Management (Zustand)
```typescript
import { create } from 'zustand'

interface Store {
  connections: DatabaseConnection[]
  addConnection: (conn: DatabaseConnection) => void
}

export const useStore = create<Store>((set) => ({
  connections: [],
  addConnection: (conn) => set((state) => ({
    connections: [...state.connections, conn]
  }))
}))
```

### Wails RPC Calls
```typescript
async function saveConnection(conn: DatabaseConnection) {
  try {
    await ConnectionManager.Save(conn)
  } catch (err) {
    console.error('Failed:', err)
    throw err
  }
}
```

---

## General Principles

1. **No type suppression**: Never use `as any` or `@ts-ignore`
2. **Context everywhere**: All DB operations need context with timeout
3. **No passwords in structs**: Always use OS Keychain
4. **Error wrapping**: Always wrap errors with context
5. **Test coverage**: 80%+ for critical paths
