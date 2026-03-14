# Core Services (Go Backend)

## Overview
Business logic lives in Go packages under `internal/`. Each service is a Go struct that gets bound to Wails for frontend access.

## 1. Export & Import Services (`internal/export/`)
- `ExportService` and `ImportService` structs bound to Wails
- Formats implemented as Go interfaces (CSV, JSON, SQL, XLSX, Markdown)
- Streaming: large datasets written via `io.Writer` pipes to avoid RAM spikes
- **XLSX**: Use `github.com/xuri/excelize/v2` for Excel file generation
- Frontend calls: `ExportTable(connectionID, tableName, format, options)`
- Progress pushed via `runtime.EventsEmit(ctx, "export:progress", pct)`

## 2. Formatting Services (`internal/formatting/`)
- `SQLFormatterService`: SQL beautification using regex-based tokenizer or `github.com/k0kubun/sqldef`
- `DateFormatter`: Go's `time.Format()` for locale-aware date display
- Called from React when user clicks "Format SQL" in editor

## 3. Infrastructure (`internal/infra/`)
- **DeepLinkHandler**: Parses `tablepro://` URLs, extracts connection params
  - On macOS: Wails handles URL scheme registration
  - Queues deep links if app not fully loaded yet
- **WindowManager**: Wails manages windows natively
  - Multi-window via `runtime.WindowSetTitle()`, `runtime.WindowShow()`
- **Updater**: Self-update via `github.com/rhysd/go-github-selfupdate` or custom HTTP check

## 4. Query Builders (`internal/query/`)
- **DialectProvider**: Maps `DatabaseType` → quoting rules, param style, pagination syntax
- **TableQueryBuilder**: Generates parameterized `SELECT/INSERT/UPDATE/DELETE`
  - Handles filters, sorting, pagination per dialect
  - Example: PostgreSQL `SELECT * FROM "users" WHERE "age" > $1 ORDER BY "name" LIMIT $2 OFFSET $3`
- **RowParser**: Converts raw `[]any` from Go drivers into typed frontend-consumable JSON
- **SQLStatementGenerator**: Converts `DataChangeManager` deltas into executable SQL

## 5. Licensing (`internal/license/`)
- **LicenseManager**: Checks license state (Free vs Pro)
- **LicenseAPIClient**: HTTP calls to Lemon Squeezy or custom backend
- **SignatureVerifier**: Ed25519 public key verification via Go's `crypto/ed25519`
- License key stored in OS Keychain via `go-keyring`
