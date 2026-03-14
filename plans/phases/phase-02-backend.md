# Phase 2: Core Backend Infrastructure

**Duration**: 2-3 weeks | **Priority**: 🔴 Critical | **Tasks**: 20

---

## Overview

Establish the core backend infrastructure that all features will build upon. This phase creates the application skeleton, logging, error handling, and configuration systems.

---

## Task List

### 2.1 Wails App Structure

#### 2.1.1 Create main App struct
```go
// cmd/main.go
type App struct {
    ctx context.Context
}

func NewApp() *App {
    return &App{}
}

func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
    log.Println("App starting...")
}

func (a *App) shutdown(ctx context.Context) {
    log.Println("App shutting down...")
}
```
- **Output**: `cmd/main.go` with App struct
- **Verification**: App compiles and runs

#### 2.1.2 Implement startup() and shutdown()
- Add logging
- Initialize services in order
- Clean up resources on shutdown
- **Verification**: Logs show startup/shutdown sequence

#### 2.1.3 Set up Wails bindings
```go
err := wails.Run(&options.App{
    Title:     "TablePro",
    Width:     1280,
    Height:    720,
    MinWidth:  1024,
    MinHeight: 600,
    OnStartup:  app.startup,
    OnShutdown: app.shutdown,
    Bind: []interface{}{
        app,
    },
})
```
- **Verification**: Go methods callable from frontend

#### 2.1.4 Configure Wails events system
```go
// Emit event
runtime.EventsEmit(a.ctx, "app:ready", map[string]any{
    "version": "0.1.0",
})

// Frontend listens
EventsOn('app:ready', (data) => console.log(data))
```
- **Verification**: Events flow Go → React

#### 2.1.5 Test Go↔React communication
- Create test method `GetVersion()` in Go
- Call from React component
- Display result in UI
- **Verification**: End-to-end RPC working

---

### 2.2 Logging Infrastructure

#### 2.2.1 Set up slog structured logging
```go
import "log/slog"

var log *slog.Logger

func init() {
    log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
}
```
- **Output**: Centralized logger package
- **Verification**: Logs output in JSON format

#### 2.2.2 Create log levels configuration
```go
type LogLevel string
const (
    LevelDebug LogLevel = "debug"
    LevelInfo  LogLevel = "info"
    LevelWarn  LogLevel = "warn"
    LevelError LogLevel = "error"
)

func SetLogLevel(level LogLevel) {
    // Update logger level
}
```
- **Verification**: Log level changeable at runtime

#### 2.2.3 Implement log file rotation
```go
import "gopkg.in/natefinch/lumberjack.v2"

writer := &lumberjack.Logger{
    Filename:   "/var/log/tablepro/app.log",
    MaxSize:    100, // MB
    MaxBackups: 3,
    MaxAge:     28,  // days
}
```
- **Output**: Logs rotate automatically
- **Verification**: Old logs archived/deleted

#### 2.2.4 Add context-aware logging
```go
func (a *App) Connect(ctx context.Context, id string) error {
    logger := log.With("connection_id", id)
    logger.InfoContext(ctx, "Connecting to database")
}
```
- **Verification**: Log lines include context fields

#### 2.2.5 Create debug event emitters
```go
func debugEmit(ctx context.Context, event string, data any) {
    log.Debug("Event emitted", "event", event, "data", data)
    runtime.EventsEmit(ctx, "debug:"+event, data)
}
```
- **Output**: Debug events visible in frontend console
- **Verification**: Debug panel shows events

---

### 2.3 Error Handling

#### 2.3.1 Define custom error types
```go
type Error struct {
    Code    string
    Message string
    Cause   error
    Context map[string]any
}

func (e *Error) Error() string {
    return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
}
```
- **Output**: `internal/errors/errors.go`
- **Verification**: Errors include code and context

#### 2.3.2 Create error wrapping utilities
```go
func Wrap(err error, message string) error {
    return &Error{
        Code:    "INTERNAL",
        Message: message,
        Cause:   err,
    }
}

func Wrapf(err error, format string, args ...any) error {
    return Wrap(err, fmt.Sprintf(format, args...))
}
```
- **Verification**: Stack traces preserved

#### 2.3.3 Implement error codes enumeration
```go
const (
    ErrConnectionFailed  = "CONNECTION_FAILED"
    ErrQueryFailed       = "QUERY_FAILED"
    ErrAuthFailed        = "AUTH_FAILED"
    ErrNotFound          = "NOT_FOUND"
    ErrValidationFailed  = "VALIDATION_FAILED"
)
```
- **Output**: Standardized error codes
- **Verification**: Consistent error handling

#### 2.3.4 Set up error translation for frontend
```go
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func ToAPIError(err error) *APIError {
    // Map internal errors to frontend-safe errors
}
```
- **Verification**: Frontend receives clean errors

#### 2.3.5 Create error reporting utilities
```go
func ReportError(ctx context.Context, err error, operation string) {
    log.ErrorContext(ctx, operation+" failed", "error", err)
    // In future: send to Sentry/crash reporting
}
```
- **Verification**: Errors logged with full context

---

### 2.4 Configuration Management

#### 2.4.1 Define config struct
```go
type Config struct {
    App      AppConfig      `json:"app"`
    Database DatabaseConfig `json:"database"`
    Log      LogConfig      `json:"log"`
}

type AppConfig struct {
    Name    string `json:"name"`
    Version string `json:"version"`
    Debug   bool   `json:"debug"`
}
```
- **Output**: `internal/config/config.go`
- **Verification**: Config loads from file

#### 2.4.2 Implement config loading
```go
func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return DefaultConfig(), nil
    }
    var cfg Config
    return &cfg, json.Unmarshal(data, &cfg)
}
```
- **Verification**: Missing file uses defaults

#### 2.4.3 Add environment variable overrides
```go
func (c *Config) ApplyEnv() {
    if v := os.Getenv("TABLEPRO_DEBUG"); v != "" {
        c.App.Debug = v == "true"
    }
}
```
- **Verification**: ENV vars override file config

#### 2.4.4 Create config validation
```go
func (c *Config) Validate() error {
    if c.Database.Timeout < 1 {
        return errors.New("timeout must be >= 1")
    }
    return nil
}
```
- **Verification**: Invalid config rejected on startup

#### 2.4.5 Set up hot-reload
```go
func WatchConfig(path string, onChange func(*Config)) {
    // Watch file for changes
    // Reload and call onChange
}
```
- **Verification**: Config changes apply without restart

---

## Deliverables

| Item | Location | Status |
|------|----------|--------|
| App struct | `cmd/main.go` | ⬜ Not Started |
| Logging | `internal/log/` | ⬜ Not Started |
| Error handling | `internal/errors/` | ⬜ Not Started |
| Config system | `internal/config/` | ⬜ Not Started |

---

## Acceptance Criteria

- [ ] App starts and shuts down cleanly
- [ ] Logging outputs structured JSON
- [ ] Errors include codes and context
- [ ] Config loads from file + env
- [ ] Go↔React RPC working

---

## Next Phase

→ [Phase 3: Connection Management](phase-03-connections.md)
