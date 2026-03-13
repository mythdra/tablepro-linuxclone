# TablePro System Architecture

## Overview
TablePro is a native macOS database client designed as a fast, lightweight alternative to TablePlus. 
It requires macOS 14.0+ and is built entirely in Swift 5.9, utilizing a mix of SwiftUI and AppKit. It is distributed as a Universal Binary (arm64 + x86_64).

## High-Level Modules
The application codebase is strictly organized into distinct layers and modules:

- **Core** (`TablePro/Core/`): Contains the business logic, services, database management, change tracking, and export/import functionalities.
- **Views** (`TablePro/Views/`): Contains all SwiftUI user interface components. Organized by feature (e.g., Editor, Connection, Sidebar, Main, Results, Toolbar).
- **Models** (`TablePro/Models/`): Contains simple data structures and domain objects (Connection, Database, Query, UI state, clickhouse, AI, etc.).
- **ViewModels** (`TablePro/ViewModels/`): Presentation logic that binds Models to Views using SwiftUI's latest `@Observable` macro approach.
- **Plugins** (`Plugins/`): Independent bundles (`.tableplugin`) for Database Drivers (MySQL, PostgreSQL, etc.) and Export/Import formats.
- **Theme/Extensions**: Contains shared styling, syntax highlighting themes, and Swift extensions.
- **Libs/**: Pre-built static libraries for database C-clients (e.g., `libmariadb.a`, `libpq.a`) tracked via Git LFS.

## Third-party Integrations (SPM)
- **CodeEditSourceEditor**: Used for the SQL editor, featuring tree-sitter based syntax highlighting and multi-cursor support.
- **Sparkle**: Used for automatic in-app updates.
- **OracleNIO**: Swift implementation of Oracle DB driver.

## Core Coordinator Pattern
The primary state and flow of the application are orchestrated by the `MainContentCoordinator`. 
Due to its complexity, it is divided into multiple extension files (e.g., `+Alerts`, `+Filtering`, `+Pagination`, `+RowOperations`) to keep the source clean. 

## Change Tracking System
The app uses a sophisticated tracking and undo mechanism when users edit table cells directly:
1. User modifies a cell -> `DataChangeManager` stores the delta.
2. User hits Save -> `SQLStatementGenerator` creates appropriate `INSERT/UPDATE/DELETE` commands safely handling primary keys.
3. Undo/Redo is supported via `DataChangeUndoManager`.

## Storage & Persistence
- **Passwords**: Apple Keychain via `ConnectionStorage`.
- **Preferences**: UserDefaults via `AppSettingsStorage` / `AppSettingsManager`.
- **Query History**: SQLite FTS5 for fast full-text searching via `QueryHistoryStorage`.
- **Session/Tab State**: Persisted as JSON via `TabPersistenceService`. To prevent performance hits, huge queries (>500KB) are truncated.

## Agent Strategy & Development Rules
- Strict adherence to SwiftLint (line length < 120, max 1100 lines/file, max 160 lines/func).
- Localization is mandatory using `String(localized:)` for user-facing texts.
- Performance: O(1) string access required (`NSString`), avoiding `String.count` heavily in tight loops.
