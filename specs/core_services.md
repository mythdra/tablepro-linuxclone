# TablePro Core Services

## Overview
The `Core/Services` directory encapsulates business logic that doesn't belong directly in UI layers or raw database drivers. 

## 1. Export & Import Services (`Core/Services/Export/`)
TablePro allows exporting query results or entire tables into various formats.
- Uses `ExportService` and `ImportService` coordinators.
- Formats are actually independent plugins loaded by `PluginManager` (e.g., CSV, JSON, MQL, SQL, XLSX).
- To stream massive datasets without maxing out RAM, data is passed as Streams or batched arrays via `PluginExportDataSource`.
- **XLSXExportPlugin**: Uses native Swift XLSX manipulation (or Core-written `XLSXWriter`) to emit proper Excel files. 

## 2. Formatting Services (`Core/Services/Formatting/`)
- `SQLFormatterService`: Parses dirty, unstructured SQL queries and formats them with proper indentation and casing according to standard styles.
- `DateFormattingService`: Utility to parse standard SQL-styled dates to locale-aware localized string presentations on the native macOS UI.

## 3. Infrastructure & Routing (`Core/Services/Infrastructure/`)
- `AppNotifications`: Central point for macOS Notification Center alerts (e.g., Export complete, Query successful).
- `DeeplinkHandler`: Parses incoming URL schemes (`tablepro://`) to automatically open specific connections or run queries.
- `WindowOpener` / `WindowManager`: AppKit abstractions bridging macOS Window management to SwiftUI `WindowGroup`.
- `UpdaterBridge`: Connects the Sparkle (Obj-C) standard framework to the Swift UI for "Check for Updates" actions.

## 4. Query Builders (`Core/Services/Query/`)
- **SQLDialectProvider**: Maps `DatabaseType` to specific string escaping and formatting rules for SQL query generation.
- **TableQueryBuilder**: Generates safe, parameterized `SELECT`, `INSERT`, `UPDATE`, `DELETE` queries dynamically taking pagination, filtering, grouping, and ordering into account.
- **RowParser**: Translates raw `ResultRow` values stringified from C-DB drivers into typed native structures conforming to the table schema definitions.
- **RowOperationsManager**: The UI bridge utilizing `TableQueryBuilder` for mutating table rows safely.

## 5. Licensing (`Core/Services/Licensing/`)
- **LicenseManager**: Exposes current app license state (Free vs Pro).
- **LicenseAPIClient**: Communicates with standard Lemon Squeezy or custom backend REST protocols for validation.
- **LicenseSignatureVerifier**: Uses a cryptographic public key to verify that the cached license key hasn't been tampered with or tampered offline (prevents basic pirate bypasses).
