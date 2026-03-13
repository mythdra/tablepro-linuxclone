# TablePro State Management & Storage

## Overview
TablePro uses a combination of different storage engines to persist application state, secrets, and user data efficiently depending on the data type's characteristics.

## Application State (In-Memory)
- The app relies on the new Swift `@Observable` macro for reactive state tracking natively in SwiftUI.
- Global singleton managers (e.g., `DatabaseManager`, `AppSettingsManager`) publish their state changes to the UI seamlessly.
- State mutation happens on the `@MainActor` to prevent threading issues typical to UI frameworks.

## Storage Mechanisms (`Core/Storage`)

### 1. ConnectionStorage (Keychain)
- Uses Apple's Keychain Services to securely store:
  - Database passwords.
  - SSH private key passphrases.
  - SSH passwords.
- No sensitive keys are stored in UserDefaults. 

### 2. AppSettingsStorage (UserDefaults)
- standard `UserDefaults` is used for lightweight preferences.
- Stores UI layouts, theme preferences, font sizes, global timeouts, and "Reopen Last Session" state.
- Managed by `AppSettingsManager` which exposes observable properties.

### 3. QueryHistoryStorage (SQLite)
- All successfully executed queries are logged.
- Persisted using a local SQLite database utilizing **FTS5 (Full-Text Search)**.
- Facilitates fast text search over thousands of historical queries.
- Cleaned up periodically (e.g., deleting history older than 30 days depending on settings).

### 4. Tab State Persistence (JSON)
- The application remembers opened tabs (Table queries, Custom SQL queries) across restarts.
- `TabDiskActor` / `TabPersistenceService` serializes the tab state to JSON documents.
- **Performance Guard**: To prevent the UI from freezing during JSON serialization, queries larger than 500KB are aggressively truncated before persisting.

### 5. Other Storages
- `FilterSettingsStorage`: Remembers user-defined column filters per table.
- `LicenseStorage`: Caches the software license key and validation signatures.
- `AIChatStorage` / `AIKeyStorage`: Manages ChatGPT/Claude API keys and conversation history.

## Data Change Tracking (`Core/ChangeTracking`)
When users perform direct edits on the result grid:
- `DataChangeManager` tracks the before/after state (Delta) of every modified cell.
- Creates an undo stack.
- `SQLStatementGenerator` takes these tracked changes and safely generates dialect-specific `INSERT/UPDATE/DELETE` statements.
- `AnyChangeManager` provides a protocol-based abstraction for testing and dependency injection.
