## Context

Phase 1 đã hoàn thành với project structure, Wails setup, và CI/CD. Hiện tại có Go module, directory structure (`cmd/`, `internal/`, `frontend/`), nhưng chưa có core backend infrastructure.

**Ràng buộc**:
- Must use Go 1.21+ với Wails v2
- Logging phải structured (slog), context-aware
- Error handling phải có error codes, wrapping
- Config phải load từ file + env, có validation
- Timeline: 2-3 weeks

## Goals / Non-Goals

**Goals:**
- App struct với startup/shutdown lifecycle
- Structured logging với slog, JSON output, log rotation
- Custom error types với error codes và wrapping
- Configuration system với file loading, env overrides, validation
- Wails RPC bindings và events pub/sub working
- Frontend có thể call Go methods và receive events

**Non-Goals:**
- Database connection implementation (Phase 3)
- Query execution logic (Phase 5)
- UI components beyond basic testing
- Production deployment optimization

## Decisions

### 1. Logging: slog (Go 1.21 standard library)
**Rationale**: No external dependencies, fast, structured, context-aware
**Alternatives Considered**: 
- zap: Faster nhưng cần external dependency
- logrus: Popular nhưng API cũ hơn slog

### 2. Log Rotation: lumberjack
**Rationale**: Simple, widely used, works with slog
**Alternatives**: 
- Rolling file handler trong slog: Complex hơn
- External log collector (Loki, ELK): Overkill cho desktop app

### 3. Error Handling: Custom Error struct
**Rationale**: Type-safe, includes error codes for frontend, preserves stack traces
**Alternatives**:
- pkg/errors: Đã được merge vào Go 1.13+
- errors.Wrap(): slog đã có built-in error wrapping

### 4. Config: JSON file + ENV overrides
**Rationale**: Simple, human-readable, ENV cho sensitive data
**Alternatives**:
- YAML: Complex parsing, không có lợi ích rõ ràng
- TOML: Less popular trong Go ecosystem
- Viper: Over-engineering cho simple config

### 5. Config Location: ~/.config/tablepro/
**Rationale**: XDG Base Directory spec, standard cho Linux/macOS
**Alternatives**:
- /etc/tablepro/: Cần root permission
- App bundle: Không phù hợp cho user-specific config

### 6. Wails Events: runtime.EventsEmit/On
**Rationale**: Built-in pub/sub, type-safe, auto-generated TypeScript bindings
**Alternatives**:
- WebSocket: Overkill cho desktop app
- HTTP polling: Inefficient, complex

### 7. App Lifecycle: OnStartup/OnShutdown hooks
**Rationale**: Wails cung cấp, clean separation of concerns
**Alternatives**:
- init(): Không thể return errors, khó test
- main() wrapper: Complex, không cần thiết

## Risks / Trade-offs

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Log files grow too large | Medium | Medium | Rotation với max size 100MB, max age 28 days |
| Config file corruption | Medium | Low | Backup before write, validate on load |
| Error messages leak sensitive info | High | Medium | API error translation, audit error messages |
| ENV overrides hard to debug | Low | Medium | Log effective config on startup |
| Events not received by frontend | Medium | Low | Error handling, retry logic, timeouts |
| Startup time slow due to config loading | Low | Low | Lazy loading, cache config in memory |

## Migration Plan

Not applicable - greenfield development, không có migration.

## Open Questions

1. **Log level default**: Debug cho development, Info cho production? (recommend: Info)
2. **Config hot-reload**: Có cần thiết cho Phase 2 không? (recommend: optional, làm sau)
3. **Error reporting**: Tích hợp Sentry/Crashlytics ngay hay để Phase 17? (recommend: Phase 17)
