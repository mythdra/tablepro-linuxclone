# Tab Management & Session Restoration

This document details the algorithms used to save query tabs between launches and the memory management mechanics utilized to prevent the application from crashing when dozens of tabs return millions of rows.

## 1. Disk Persistence Engine (`TabDiskActor`)
- `TabDiskActor` operates as an asynchronous, thread-safe funnel for I/O.
- Instead of using `UserDefaults` (which struggles with large binary blobs), it serializes the entire `[PersistedTab]` array for each Connection ID into individual `.json` files under `~/Library/Application Support/TablePro/TabState/`.
- **Synchronous Override**: Because macOS can terminate an app instantly on `Cmd+Q`, the actor exposes a `saveSync()` static method. `applicationWillTerminate` bypasses the asynchronous actor queue to force-flush the JSON to disk synchronously, guaranteeing tab state survival during hard quits.

## 2. Tab State Truncation (`TabPersistenceCoordinator`)
- To prevent JSON blobs from reaching hundreds of megabytes, queries that exceed `QueryTab.maxPersistableQuerySize` (currently around 500KB) are intentionally stripped from the saved state.
- `TabPersistenceCoordinator` does **not** use timers (like `debouncedSave()`). Instead, it is invoked explicitly on critical mutation events: Tab Added, Tab Closed, Text Editor Blurred, Connection Closed. This ensures maximum efficiency while avoiding race conditions on the main thread loop.

## 3. LRU Tab Eviction (`MainContentCoordinator+TabSwitch`)
When a user runs a `SELECT *` in 10 separate tabs, memory usage can balloon exponentially. TablePro implements an aggressive Tab Eviction policy to maintain a flat memory footprint:
- When a user switches tabs, `MainContentCoordinator.handleTabChange` calculates the active working set (usually the `oldTabId` and `newTabId`).
- It scans the remaining background tabs (`tabManager.tabs`).
- It filters for candidates: must not have pending unsaved edits (`!hasChanges`), must have executed previously, and must not already be evicted.
- It sorts these candidates by `lastExecutedAt` timestamp.
- If there are more than 2 inactive loaded tabs, it drops the `.resultRows` arrays of the LRU (Least Recently Used) tabs.
- **Restoration**: When the user clicks back onto an evicted tab, the `needsLazyQuery` boolean dynamically flips to `true`, and the application automatically and silently re-executes the saved SQL query to stream the data back into the grid seamlessly.

## 4. Sub-State Recovery
When switching into a tab, the Coordinator pushes the following context properties back into the global state managers before re-rendering the Data Grid:
1. **Filter State**: `filterStateManager.restoreFromTabState`
2. **Column Visibility**: `columnVisibilityManager.restoreFromColumnLayout`
3. **Change Tracking Checkums**: `changeManager.restoreState`

## Qt/C++ Migration Guidelines
- **Persistence**: Replicate `TabDiskActor` using `QJsonDocument` pointing to `QStandardPaths::AppDataLocation`. Consider using a dedicated background `QThread` or `QtConcurrent::run` to avoid blocking the UI during save, but ensure you hook into `QGuiApplication::aboutToQuit` using a blocking `QMutex` waiting condition to flush the last write.
- **LRU Eviction**: In C++, holding 2 million `QVariant` rows in background `QAbstractTableModel` instances will crash even quicker than Swift. Re-implement `evictInactiveTabs` precisely. Let `QSortFilterProxyModel` instances dump their underlying source model data and reset when hidden.
