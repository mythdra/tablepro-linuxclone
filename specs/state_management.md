# State Management & Storage (Go + React)

## Overview
State in the new architecture splits cleanly between **Go backend** (persistent data, business logic) and **React frontend** (UI state, ephemeral interactions).

## Frontend State (React — Zustand)
- **Zustand stores** replace Swift's `@Observable` pattern
- Each store is a hook: `useTabStore()`, `useConnectionStore()`, `useSettingsStore()`
- Stores call Go methods via Wails bindings for data mutations
- UI re-renders automatically when store state changes

## Backend State (Go — in-memory)
- Go services hold runtime state in concurrent-safe maps protected by `sync.RWMutex`
- `DatabaseManager.sessions` — active connections
- `TabManager.tabs` — per-connection tab state
- Push updates to React via `runtime.EventsEmit()`

## Storage Mechanisms

### 1. ConnectionStorage (OS Keychain)
- Uses `github.com/zalando/go-keyring` for cross-platform Keychain access
- Stores: database passwords, SSH passwords, SSH key passphrases, AI API keys
- Connection metadata (non-sensitive) stored as JSON file
- **Key format**: `tablepro:password:{connectionUUID}`

### 2. AppSettings (JSON file)
- Path: `~/.config/tablepro/settings.json` (Linux/Mac) / `%APPDATA%/TablePro/settings.json` (Windows)
- Uses Go's `encoding/json` + `os.ReadFile`/`os.WriteFile`
- Stores: theme, font, editor preferences, timeouts, pagination defaults

### 3. QueryHistory (SQLite + FTS5)
- Embedded SQLite via `go-sqlite3` or `modernc.org/sqlite`
- Schema: `CREATE VIRTUAL TABLE query_history USING fts5(query, connection_id, database_name, execution_time, row_count, was_successful, error_message, created_at)`
- Auto-cleanup: queries older than 30 days pruned on app startup

### 4. Tab State (JSON per connection)
- Path: `~/.config/tablepro/tabs/{connectionUUID}.json`
- Saved explicitly on: tab switch, tab close, window close, app quit
- Queries > 500KB are truncated before persisting
- Graceful sync write on app quit (Go's `os.Signal` SIGTERM handler)

### 5. Other Storages
- `FilterSettings`: JSON file per connection for saved column filters
- `LicenseStorage`: Encrypted license key + Ed25519 signature verification
- `AIChatStorage`: Conversation history as JSON

## Data Change Tracking (Go backend)
```go
type DataChangeManager struct {
    mu          sync.Mutex
    changes     map[string]*CellChange  // key: "row:col"
    insertedRows []RowData
    deletedRows  []RowIdentity
    undoStack   []ChangeAction
    redoStack   []ChangeAction
}
```
- React sends cell edits via `UpdateCell(tabID, row, col, newValue)`
- Go tracks deltas and generates dialect-specific SQL on commit
- Undo/Redo managed in Go, React reflects current state
