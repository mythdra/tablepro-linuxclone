# Helper Functions & Architectural Gotchas

This document catalogs critical utility functions, caching strategies, and subtle behaviors found throughout the TablePro codebase. These must be faithfully replicated in the Qt/C++ rewrite to maintain performance and feature parity.

## 1. `DateFormattingService` (High-Performance Caching)

Rendering 10,000 rows x 20 columns in a Data Grid means processing potentially hundreds of thousands of date strings. Standard date parsing is extremely slow.

### The Gotchas:
- **String Caching**: The app maintains an `NSCache` (max 10,000 items) mapping raw database strings (e.g., `"2024-03-01 12:00:00"`) directly to their localized, formatted display strings.
- **Timezone Handling**:
  - If a format string contains a timezone marker (e.g. `ISO8601`), it parses the offset.
  - If a format string is "naive" (e.g., `yyyy-MM-dd HH:mm:ss`), it parses the time utilizing the user's `TimeZone.current` rather than UTC. This ensures that displaying `"12:00:00"` does not accidentally shift to `"17:00:00"` just because the local system is in EST.

### Qt Implementation:
Use a thread-safe `QCache<QString, QString>` inside a global Singleton or Data Model context. For date manipulation, use `QDateTime::fromString` ensuring naive dates invoke `Qt::LocalTime`.

## 2. `SQLFormatterService` (Safe Parsing & Cursor Tracking)

The SQL Formatter attempts to beautify messy queries but must be extremely defensive to prevent data corruption or UI freezing.

### The Gotchas:
- **10MB DoS Protection**: Prevents freezing on massive dumps.
- **Placeholder Replacement**: Before running the regex to uppercase keywords like `SELECT`, the engine temporarily strips out all String Literals (`'...'`) and Comments (`-- ...`) and replaces them with a UUID placeholder (`__STRING_0__`). This prevents the formatter from modifying string data (e.g., changing `'please select a user'` to `'please SELECT a user'`).
- **Cursor Ratio Preservation**: When a user highlights code and clicks "Format", the engine calculates a float ratio `(original_cursor_index / original_string_length)` and applies it to the newly formatted string length so their cursor doesn't jump to the beginning of the file.

### Qt Implementation:
Use `QRegularExpression` caching. The string replacement logic can be replicated natively, ensuring `QRegularExpression` is instantiated once and reused to minimize CPU load.

## 3. `ExportService` (Progress Coalescing & Batched Queries)

When exporting 20 tables, calculating the exact progress bar denominator natively requires running `SELECT COUNT(*)` 20 times, which is slow.

### The Gotchas:
- **UNION ALL Batching**: The service clumps up to 50 `COUNT(*)` checks into a single query via `UNION ALL` to prevent database round-trips.
- **Progress Coalescing**: The C-plugins emit progress updates (e.g., `processedRows = 120530`) thousands of times a second. The app utilizes a `ProgressUpdateCoalescer` that drops rapid frames and only dispatches updates to the MainUI Thread at a manageable framerate (e.g., 30fps), preventing the UI from locking up during a massive export.

### Qt Implementation:
Progress signals from the `QThread` or `QtConcurrent` must be debounced/throttled or passed through a `QTimer`-controlled throttler before hitting the QML UI bindings.

## 4. `UserDefaults` Extensions

- **Recent Databases History**: Connects history arrays to specific `connectionId` UUIDs, bounding the history to a `maxRecentCount = 5` via `Array.prefix`.
- **Qt Implementation**: Replicate using `QSettings` utilizing nested `beginGroup(uuid)` logic.

## 5. `String` Extensions
- **JSON Pretty Printing**: Extracts strings -> JSON Objects -> Serializes with `.prettyPrinted`, `.sortedKeys`, and `.withoutEscapingSlashes`.
- **SHA256**: Utilizes `CryptoKit` (macOS native) to hash strings securely.
- **Qt Implementation**: Replicate using `QJsonDocument(QJsonDocument::fromJson).toJson(QJsonDocument::Indented)` and `QCryptographicHash::hash(data, QCryptographicHash::Sha256).toHex()`.
