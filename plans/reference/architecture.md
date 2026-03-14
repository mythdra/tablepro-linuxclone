# Architecture Reference

## System Overview

TablePro is a cross-platform database client built with:
- **Backend**: Go + Wails v2
- **Frontend**: React + TypeScript
- **Communication**: Wails RPC + Events

---

## High-Level Architecture

```
┌─────────────────────────────────────────┐
│           React Frontend (WebView)       │
│  ┌──────┐ ┌──────┐ ┌────────┐ ┌──────┐ │
│  │Sidebar│ │Editor│ │DataGrid│ │Toolbar│ │
│  └──┬───┘ └──┬───┘ └───┬────┘ └──┬───┘ │
│     └────────┴─────────┴─────────┘      │
│              wails.EventsEmit            │
│              wails.Bind (RPC)            │
├─────────────────────────────────────────┤
│           Go Backend (Wails Runtime)     │
│  ┌─────────────┐  ┌──────────────────┐  │
│  │ ConnectionMgr│  │ DatabaseManager  │  │
│  │ TabManager   │  │ ExportService    │  │
│  │ SettingsMgr  │  │ ImportService    │  │
│  │ HistoryMgr   │  │ ChangeTracker    │  │
│  └──────┬──────┘  └────────┬─────────┘  │
│         └──────────────────┘             │
│              Driver Interface            │
│  ┌──────┐ ┌─────┐ ┌──────┐ ┌─────────┐ │
│  │ pgx  │ │mysql│ │sqlite│ │ duckdb  │ │
│  └──────┘ └─────┘ └──────┘ └─────────┘ │
└─────────────────────────────────────────┘
```

---

## Module Organization

```
cmd/
  main.go              # Wails entry point
internal/
  driver/              # Database driver interface + implementations
  connection/          # Connection CRUD, Keychain, URL parser
  session/             # Active connection sessions & health monitoring
  query/               # Query builder, pagination, dialect provider
  change/              # DataChangeManager, SQLStatementGenerator
  export/              # Export/Import orchestrators
  history/             # Query history (SQLite FTS5)
  settings/            # App preferences (JSON file)
  tab/                 # Tab persistence (JSON per connection)
  ssh/                 # SSH tunnel management
  license/             # License validation (Ed25519)
frontend/
  src/
    components/        # React UI components
    stores/            # Zustand state stores
    hooks/             # Custom React hooks
    lib/               # Utility functions
```

---

## Communication Patterns

### Go → React (RPC)
```go
// Go: Bound method
func (a *App) GetVersion() string {
    return "1.0.0"
}

// React: Async call
const version = await GetVersion()
```

### Go → React (Events)
```go
// Go: Emit event
runtime.EventsEmit(ctx, "connection:status", status)

// React: Listen
EventsOn('connection:status', (status) => setStatus(status))
```

---

## Data Flow

1. User action in UI
2. React calls Go method via Wails RPC
3. Go executes business logic
4. Go emits event if state changed
5. React updates Zustand store
6. UI re-renders

---

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Desktop Framework | Wails | Single binary, Go backend |
| State Management | Zustand | Minimal boilerplate |
| Data Grid | AG Grid | Virtual scrolling, performance |
| SQL Editor | Monaco | VS Code engine |
| Password Storage | OS Keychain | Security, native integration |

---

See `/specs/architecture.md` for complete architecture specification.
