# Data Grid Internals

This document exhaustively details the internal workings of the Data Grid component (`DataGridView`), which is the most complex UI surface in the application. The Qt/C++ rewrite must achieve feature parity with its performance characteristics.

## 1. Core Architecture Map
The current application wraps a native `NSTableView` inside a SwiftUI `NSViewRepresentable`. 
**Qt Equivalent:** `QTableView` with a custom `QAbstractTableModel` and a custom `QStyledItemDelegate`.

### Identity-Based Redrawing (`DataGridIdentity`)
To achieve high performance, the grid blocks the declarative UI loop from forcing expensive full-table reloads using an *Identity Snapshot*.
The snapshot consists of:
- `reloadVersion` (Number of data mutations)
- `resultVersion` (New query results)
- `metadataVersion` (Schema updates like ForeignKey discovery)
- `rowCount` & `columnCount`
- `hiddenColumns` set

If the identity has *not* changed since the last frame, the grid skips the `updateNSView` body entirely.

### Granular Rendering & Caching
The grid does not render every column for every row at once (Virtualization).
When `changeManager.reloadVersion` increments, it checks exactly *which* row indices mutated. If `changedRows.count < 500`, it requests a granular `tableView.reloadData(forRowIndexes: ...)` rather than a full table reset.

## 2. Dynamic Column Sizing Algorithm
Calculating perfect column widths for millions of rows is impossible. The application uses a bounded O(1) heuristic instead of an O(N) font-measurement lookup.

#### Algorithm (`calculateOptimalColumnWidth`):
1. **Header Baseline:** Starts by estimating the header title width (`headerCharCount * monoCharWidth * 0.75 + 48`).
2. **Sampling:** It limits the measurement to exactly `30` rows total (or `10` if there are >50 columns). It strides through the dataset (e.g., measuring row 0, 1000, 2000... 30000).
3. **Character Cap**: Even within the 30 rows, it stops measuring a string if it exceeds `50` characters.
4. **Calculated Width**: It uses `max(headerBaseline, sampleRowsCellWidths)`.
5. **Clamping**: The final width is hard-clamped between `minColumnWidth: 60` and `maxColumnWidth: 800`.

## 3. Custom Cell Factories
Every rendered cell is managed by `DataGridCellFactory`. It generates distinct cell UI based on the Data Type and context:

- **Row Number Cell**: Gray text `#` aligned right.
- **FK Arrow Cell (`FKArrowButton`)**: If a foreign key relationship exists for that specific column name, it injects a tiny `arrow.right.circle.fill` button on the trailing edge. Clicking it triggers a jump to the referenced row.
- **Dropdown Cell**: If the column datatype is an `ENUM`, `SET`, or boolean-like, it injects a tiny chevron `chevron.up.chevron.down` indicating that double-clicking will open a specialized combo-box rather than a text editor.
- **Layer-Backed Backgrounds (`RowVisualState`)**: 
  - Yellow background: `isModified` locally.
  - Red background: `isDeleted` locally.
  - Green background: `isInserted` locally.

### Large Dataset Degradation
If the dataset row count exceeds `5,000`, the cell factory disables several expensive visual operations:
- Focus ring drawing is disabled.
- VoiceOver cell-description accessibility labels are bypassed to prevent memory explosion.
- Placeholder text calculations ("NULL", "DEFAULT") are bypassed in favor of simple empty strings.

### The Single-Line Sanitizer
Database payloads often contain `\n` or carriage returns deep within text. Since cells must remain one-line high, the string extension `sanitizedForCellDisplay` scans UTF-16 characters and dynamically swaps `[0x0A, 0x0D, 0x0B, 0x0C, 0x85, 0x2028, 0x2029]` for spaces *only* for the purpose of the cell label drawing. The original multi-line string is stored in memory.

## 4. Keyboard Navigation & Editing

### Dual Editors
The grid utilizes two distinct editing surfaces depending on the target coordinate:
1. **Inline Field Editor**: Used for standard numbers/short strings. Native text boundary.
2. **Cell Overlay Editor**: If the underlying text contains `containsLineBreak`, the grid suppresses the inline editor and spawns a huge multi-line popover floating directly above the target cell bounds (`showOverlayEditor`).

### Tab Traversal Algorithm (`insertTab`, `insertBacktab`)
When an editor is active and the user presses `Tab`:
1. Focus attempts `currentColumn + 1`.
2. If `currentColumn >= totalColumns`: It wraps around to `column = 1` and `currentRow += 1`.
3. If it hits the end of the table entirely, it stops.
4. *Crucially*, it checks if the *next* destination cell contains multiline text. If it does, it dynamically jumps from the Inline Editor into the Overlay Editor mode mid-tab.

## 5. Real-Time Settings Observations
The `DataGridCellFactory` subscribes to global settings changes for `rowHeight`, `fontFamily`, and `dateFormat`. 
Instead of discarding the native table elements (which causes UI flicker), it loops over the **visible cells only** (`tableView.visibleRect`), swapping their fonts in place.

## Qt/C++ Migration Guidelines
- **Models**: Implement `QAbstractTableModel`. Handle the `flags()` override to block editability on Foreign Key columns (which require double-click UI). Make use of `data(const QModelIndex &index, int role)` where `Qt::BackgroundRole` serves the Yellow/Red/Green state colors, rather than manually painting CALayers.
- **Delegates**: Subclass `QStyledItemDelegate`. Implement `paint()` for drawing the FK arrows and dropdown chevrons. Implement `createEditor()` to spawn a `QTextEdit` overlay for multiline strings and a standard `QLineEdit` for basic strings.
- **Short-circuiting**: Reimplement the Identity-Based rendering bailout by tracking state revisions in the C++ model, emitting `dataChanged` signals *only* for dirty cell indices rather than `layoutChanged`.
