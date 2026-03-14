# Phase 2: Backend Infrastructure Tasks

Implementation checklist for Phase 2 - Core Backend Infrastructure (20 tasks)

---

## 1. Wails App Structure

- [x] 1.1 Create main App struct in cmd/main.go with context.Context field
- [x] 1.2 Implement NewApp() constructor function
- [x] 1.3 Implement startup(ctx context.Context) method with logging
- [x] 1.4 Implement shutdown(ctx context.Context) method with cleanup
- [x] 1.5 Configure Wails bindings with App struct in wails.Run()
- [x] 1.6 Configure Wails events system (runtime.EventsEmit, EventsOn)
- [x] 1.7 Test Go↔React communication with GetVersion() method

---

## 2. Logging Infrastructure

- [x] 2.1 Create internal/log package with logger.go
- [x] 2.2 Initialize slog logger with JSON handler in init()
- [x] 2.3 Define LogLevel type (Debug, Info, Warn, Error)
- [x] 2.4 Implement SetLogLevel(level LogLevel) function
- [x] 2.5 Add lumberjack logger for file rotation (100MB, 3 backups, 28 days)
- [x] 2.6 Implement context-aware logging (With() method)
- [x] 2.7 Create debugEmit() helper for EventsEmit + logging

---

## 3. Error Handling

- [x] 3.1 Create internal/errors package with errors.go
- [x] 3.2 Define Error struct (Code, Message, Cause, Context)
- [x] 3.3 Implement Error() method for error interface
- [x] 3.4 Create Wrap(err error, message string) function
- [x] 3.5 Create Wrapf(err error, format string, args ...any) function
- [x] 3.6 Define error codes constants (ErrConnectionFailed, ErrQueryFailed, etc.)
- [x] 3.7 Define APIError struct for frontend-safe errors
- [x] 3.8 Implement ToAPIError(err error) translation function
- [x] 3.9 Create ReportError(ctx, err, operation) helper

---

## 4. Configuration Management

- [x] 4.1 Create internal/config package with config.go
- [x] 4.2 Define Config struct (App, Database, Log subsections)
- [x] 4.3 Define AppConfig, DatabaseConfig, LogConfig structs
- [x] 4.4 Implement Load(path string) function for JSON file loading
- [x] 4.5 Implement DefaultConfig() function for fallback values
- [x] 4.6 Implement ApplyEnv() method for environment variable overrides
- [x] 4.7 Implement Validate() method for config validation
- [x] 4.8 Implement WatchConfig(path, onChange) for hot-reload
- [x] 4.9 Create config directory (~/.config/tablepro/) on first run

---

## Verification Checklist

Run these commands to verify Phase 2 completion:

```bash
# Build app
wails build

# Test logging
go run cmd/main.go 2>&1 | grep "App starting"

# Test config
cat ~/.config/tablepro/config.json

# Test error handling
go test ./internal/errors/...

# Test Go↔React RPC
wails dev  # Open browser console, call GetVersion()
```

---

## Acceptance Criteria

- [x] App starts with "App starting" log message
- [x] App shuts down cleanly with "App shutting down" message
- [x] Logs output in JSON format to stdout and file
- [x] Log rotation works when file exceeds 100MB
- [x] Errors include error codes and context
- [x] Config loads from ~/.config/tablepro/config.json
- [x] Environment variables override config values
- [x] Frontend can call Go methods via RPC
- [x] Frontend receives events from Go

---

## Dependencies

← [Phase 1: Project Setup](../archive/2026-03-14-phase-1-project-setup/)  
→ [Phase 3: Connection Management](../phase-3-connection-management/)

---

## Notes

- All 20 tasks must be complete before Phase 3 can begin
- Use `go test ./...` to verify packages compile correctly
- Test RPC communication manually in browser devtools
