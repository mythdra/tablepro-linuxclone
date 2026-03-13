# Connection Management & History Flow

This document outlines the mechanics used within TablePro for managing, storing, parsing, and testing database connection credentials.

## 1. Credentials Storage (`ConnectionStorage`)
- **Metadata**: Connection properties (Host, Port, User, DB Type, SSL config) are serialized into a master JSON blob via `UserDefaults` under key `com.TablePro.connections`.
- **Sensitive Data**: Passwords, SSH passwords, and SSH private key passphrases are **never** stored in UserDefaults. They are shifted automatically to the OS Keychain.
  - The lookup key relies on the Connection's UUID: `com.TablePro.password.[UUID]`.
  - The C++ port MUST utilize `qtkeychain` (Cross-Platform Keychain wrapper for Qt) to mirror this seamless segregation of metadata and sensitive auth data.
- **Copying Connections**: When repeating a connection via "Duplicate", TablePro intercepts the copy event, regenerates a new `UUID()`, suffixes the name with " (Copy)", and actively iterates over the Keychain APIs to fetch and re-store the secrets for the new UUID.

## 2. Dynamic Form Fields & Pgpass (`ConnectionFormView`)
- Since connection payloads diverge radically per database type (e.g. MongoDB requiring `AuthSource` while Oracle requires `ServiceName`), the form observes the selected `DatabaseType` and queries the `PluginManager` for extra properties.
- **Pgpass Integration**: For PostgreSQL, there is an asynchronous filesystem check that scans `~/.pgpass`. If the file exists but has permissions wider than `0600`, the connection form automatically displays a red UI alert instructing the user to fix their permissions.

## 3. Deep Linking & URL Parser (`ConnectionURLParser` & `AppDelegate`)
- **Drag & Drop / URL Import**: Users can drag complex strings like `postgres+ssh://ec2-user@bastion:22/dbuser:pass@10.0.0.1:5432/main_db`.
- The parser destructs this into:
  - Outer scheme (e.g., SSH jump host creds).
  - Inner scheme (internal VPC database credentials).
  - Extra Query parameters `?sslmode=require&statusColor=#FF0000&env=Production`.
- **Deep Link Execution**: If TablePro is invoked via system deep link (`tablepro://...`), the `AppDelegate` catches the event.
  - It searches `ConnectionStorage` to see if a connection *exactly matching* the host, port, db name, and user already exists.
  - If a match exists, it fires the tab.
  - If not, it builds a `Transient Connection` purely in memory. The connection is discarded on app close unless the user saves it.
  - It handles delays gracefully. If the deep link requested a specific `?table=Users&column=id&operator=eq&value=5`, it awaits the `databaseDidConnect` broadcast, opens the Table Tab, waits 300ms, and fires the filter notification.

## 4. Connection Testing Algorithm
- The UI contains a "Test Connection" button.
- When clicked:
  - It constructs a temporary `DatabaseConnection` object using the dirty form state.
  - It temporarily writes the in-memory form password to the Keychain under the temporary UUID.
  - It requests the active `PluginDriver` to initiate a blocking connection sequence.
  - **Cleanup**: It does not actively wipe the temporary Keychain entry on failure (a minor leak the Qt version might want to plug), but updates the UI with a native alert.
  - **Auto-Installation Prompt**: If testing fails due to `.pluginNotInstalled`, the system overrides the generic failure dialog and redirects the flow to the Plugin Downloader interface.

## Qt/C++ Migration Guidelines
- **Storage**: Standardize around QSettings for metadata (JSON or INI format depending on preference), but proxy all password gets/sets to `QKeychain::ReadPasswordJob` / `QKeychain::WritePasswordJob`.
- **URLs**: C++ lacks Swift's custom URL parser tolerances for dual `@` symbols. Consider building a manual regex sequence (as seen in `ConnectionURLParser.swift`) rather than relying purely on `QUrl`, which breaks on `scheme+ssh://user@host/dbuser@dbhost` syntax.
- **Queueing Engine**: Reproduce the `queuedURLEntries` array mechanism to buffer deep links during the `QApplication` splash screen / loading phase before the `QMainWindow` is fully instantiated.
