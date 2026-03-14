# Phase 2: Backend Infrastructure Specifications

## ADDED Requirements

### Capability: wails-app-structure

The system SHALL provide a Wails application structure with lifecycle management and RPC bindings.

#### Scenario: App struct created
- **WHEN** application starts
- **THEN** App struct is initialized with context

#### Scenario: Startup hook executes
- **WHEN** Wails calls OnStartup
- **THEN** startup() method initializes services and logs "App starting"

#### Scenario: Shutdown hook executes
- **WHEN** Wails calls OnShutdown
- **THEN** shutdown() method cleans up resources and logs "App shutting down"

#### Scenario: RPC bindings configured
- **WHEN** Wails app runs
- **THEN** App struct is bound and methods are callable from frontend

#### Scenario: Events system configured
- **WHEN** Go code calls runtime.EventsEmit
- **THEN** Frontend receives event via EventsOn listener

#### Scenario: Go↔React communication tested
- **WHEN** frontend calls GetVersion() RPC method
- **THEN** Go method returns version string to frontend

### Capability: logging-system

The system SHALL provide structured logging with slog.

#### Scenario: Logger initialized
- **WHEN** application starts
- **THEN** slog logger is created with JSON handler

#### Scenario: Log levels configured
- **WHEN** SetLogLevel() is called
- **THEN** Logger outputs only messages at or above the specified level

#### Scenario: Log file rotation
- **WHEN** log file exceeds 100MB
- **THEN** Logger rotates to new file with max 3 backups

#### Scenario: Context-aware logging
- **WHEN** logging with context
- **THEN** Log output includes context fields (connection_id, user_id, etc.)

#### Scenario: Debug events emitted
- **WHEN** debugEmit() is called
- **THEN** Log message is written and event is emitted to frontend

### Capability: error-handling

The system SHALL provide custom error handling with error codes and wrapping.

#### Scenario: Custom error types defined
- **WHEN** error is created
- **THEN** Error includes Code, Message, Cause, and Context fields

#### Scenario: Error wrapping
- **WHEN** Wrap() is called with an error
- **THEN** Wrapped error preserves original error as Cause

#### Scenario: Error codes enumeration
- **WHEN** error is created
- **THEN** Error uses predefined code (ErrConnectionFailed, ErrQueryFailed, etc.)

#### Scenario: API error translation
- **WHEN** ToAPIError() is called
- **THEN** Internal error is converted to frontend-safe APIError

#### Scenario: Error reporting
- **WHEN** ReportError() is called
- **THEN** Error is logged with full context and operation name

### Capability: configuration

The system SHALL provide configuration management with file loading and validation.

#### Scenario: Config struct defined
- **WHEN** config package loads
- **THEN** Config struct includes App, Database, Log subsections

#### Scenario: Config loading from file
- **WHEN** Load() is called with config path
- **THEN** Config is loaded from JSON file or defaults are returned

#### Scenario: Environment variable overrides
- **WHEN** ApplyEnv() is called
- **THEN** Environment variables override file config values

#### Scenario: Config validation
- **WHEN** Validate() is called
- **THEN** Invalid config returns error (e.g., timeout < 1)

#### Scenario: Config hot-reload
- **WHEN** config file changes
- **THEN** WatchConfig() reloads config and calls onChange callback
