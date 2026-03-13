# Import Service Algorithms

This document covers the architectural flow for importing potentially massive (multi-gigabyte) SQL dump files without blocking the UI thread or exhausting system memory.

## 1. Orchestration (`ImportService`)
- The UI triggers `ImportService.importFile(from:formatId:encoding:)`.
- The service uses the `TableProPluginKit` to locate an `ImportFormatPlugin` matching the `formatId` (e.g., `com.tablepro.import.sql`).
- It creates two crucial bridging concrete objects:
  - `ImportDataSinkAdapter`: A wrapper around the actual database driver (`PostgreSQLDriver`, `MySQLDriver`) that executes queries.
  - `SqlFileImportSource`: A wrapper around the chosen file on disk.
- It instantiates a `PluginImportProgress` tracker and starts a Coalescing mechanism (`ProgressUpdateCoalescer`) to rate-limit UI updates (throttling state updates so SwiftUI doesn't freeze on thousands of fast queries).

## 2. Decompression Engine (`FileDecompressor`)
- Before parsing starts, `SqlFileImportSource` checks the file extension.
- If it ends in `.gz`, it does **not** decompress it into RAM using Swift's `Compression` framework.
- Instead, it spawns a Detached Task that invokes the UNIX `/usr/bin/gunzip -c [file] > [tempFile.sql]`.
- This ensures maximum C-level C-library speed for extraction mapping directly to a temporary disk location.
- The `SqlFileImportSource.deinit` hook ensures the temporary `.sql` file is deleted when the import finishes or errors out.

## 3. Streaming SQL Parser (`SQLFileParser`)
- Multi-gigabyte SQL files cannot be loaded via `String(contentsOf:)` nor split via `.components(separatedBy: ";")`.
- `SQLFileParser` uses an asynchronous finite state machine.
- **Chunking**: It reads the file via `FileHandle` in exactly `64KB (65,536 bytes)` binary chunks.
- **State Machine**: It iterates through the chunk, parsing:
  - Normal statements.
  - Multi-line comments (`/* ... */`).
  - Single-line comments (`--`).
  - Single quotes, double quotes, and backticks.
- **Yielding**: Immediately upon hitting a safely un-quoted `;`, it yields the current string buffer via Swift `AsyncStream` back to the driver.
- **String Handling**: To achieve high performance, it processes on heavily optimized `NSString` arrays and avoids Swift's character iterators (`O(N^2)` scaling).

## 4. Execution Sink (`PluginImportDataSink`)
- The driver receives the yielded strings from the parser stream.
- The default behavior wraps the entire process in `sink.disableForeignKeyChecks()`, `sink.beginTransaction()`, loops the inserts, and then `sink.commitTransaction()`, `sink.enableForeignKeyChecks()`.
- Error handling logs intermediate progress to the UI and ensures if the 50,000th statement fails, the user is visibly told exactly where the dump crashed.

## Qt/C++ Migration Guidelines
- **Gunzip**: You can replace the UNIX subprocess call with Qt's built-in `QProcess` pointing to `gunzip`, or use `zlib` / `qCompress` directly on streams if preferred.
- **Streaming Parser**: The chunked 64KB FSM (Finite State Machine) pattern must be ported to C++ to avoid memory limits. A `QFile` reading `64 * 1024` byte blocks into a `QByteArray`, parsing looking for `;` outside of quotes/comments, is identical.
- **Throttling**: The `ProgressUpdateCoalescer` in Swift should map to a Qt mechanism where progress updates are emitted as Signals, but a `QTimer` throttles actual UI repaints to ~15-30fps.
