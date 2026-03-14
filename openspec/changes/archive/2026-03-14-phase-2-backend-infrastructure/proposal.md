## Why

Phase 1 đã thiết lập project structure thành công. Phase 2 tiếp tục xây dựng core backend infrastructure - nền tảng cho tất cả các features sau này. Không có logging, error handling, và configuration system, việc develop các phases tiếp theo sẽ gặp nhiều khó khăn và không có consistency.

## What Changes

- **Wails App Structure**: App struct, startup/shutdown lifecycle, RPC bindings, events system
- **Logging Infrastructure**: Structured logging với slog, log levels, file rotation, context-aware logging
- **Error Handling System**: Custom error types, error wrapping, error codes, API error translation
- **Configuration Management**: Config structs, file loading, env overrides, validation, hot-reload

## Capabilities

### New Capabilities
- `wails-app-structure`: App lifecycle, RPC bindings, events pub/sub
- `logging-system`: Structured logging với slog, log levels, rotation, context-aware
- `error-handling`: Custom error types, wrapping, codes, API translation
- `configuration`: Config loading, validation, env overrides, hot-reload

### Modified Capabilities
- (None - đây là foundational infrastructure, không modify existing capabilities)

## Impact

- **Code**: Tạo 4 packages mới trong `internal/`: log/, errors/, config/, và update cmd/
- **Dependencies**: 
  - Go: gopkg.in/natefinch/lumberjack.v2 (log rotation)
  - Frontend: Wails events integration
- **Systems**: File-based config (~/.config/tablepro/), structured logging
- **Timeline**: 2-3 weeks cho complete implementation
- **Downstream**: Tất cả phases sau đều phụ thuộc vào infrastructure này
