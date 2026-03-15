# State Management & Storage (C++20 + Qt 6)

## Overview
State in the C++ architecture splits cleanly between **C++ backend** (persistent data, business logic) and **Qt frontend** (UI state, ephemeral interactions). Qt's signal/slot system replaces both Go channels and React state updates.

## Frontend State (Qt — QObject Properties)
- **QObject-based managers** replace Swift's `@Observable` and React's Zustand stores
- Each manager exposes `Q_PROPERTY` bindings: `TabManager`, `ConnectionManager`, `QueryManager`
- Managers emit signals for data mutations, UI widgets update via slot connections
- Qt's meta-object system handles change notification automatically

## Backend State (C++ — in-memory)
- C++ services hold runtime state in concurrent-safe `QMap`/`QHash` protected by `QMutex`/`QMutexLocker`
- `ConnectionManager::m_sessions` — active connections
- `TabManager::m_tabStates` — per-connection tab state
- Push updates to Qt frontend via `emit` signals

## Storage Mechanisms

### 1. ConnectionStorage (OS Keychain via QKeychain)
- Uses `QKeychain` library for cross-platform Keychain access
- Stores: database passwords, SSH passwords, SSH key passphrases, AI API keys
- Connection metadata (non-sensitive) stored as JSON file
- **Key format**: `tablepro:password:{connectionUUID}`
```cpp
// connection/KeychainStorage.hpp
class KeychainStorage : public QObject {
    Q_OBJECT
public:
    QFuture<QString> readPassword(const QString& key);
    QFuture<bool> writePassword(const QString& key, const QString& password);
    QFuture<bool> deletePassword(const QString& key);
};
```

### 2. AppSettings (JSON file via QSettings)
- Path: `~/.config/tablepro/settings.json` (Linux/Mac) / `%APPDATA%/TablePro/settings.json` (Windows)
- Uses `QSettings` or `QJsonDocument` + `QFile`
- Stores: theme, font, editor preferences, timeouts, pagination defaults
```cpp
// core/Settings.hpp
class Settings : public QObject {
    Q_OBJECT
    Q_PROPERTY(QString theme READ theme WRITE setTheme NOTIFY themeChanged)
    Q_PROPERTY(int queryTimeout READ queryTimeout WRITE setQueryTimeout)
    Q_PROPERTY(int pageSize READ pageSize WRITE setPageSize)
};
```

### 3. QueryHistory (SQLite + FTS5 via Qt SQL)
- Embedded SQLite via `QSqlDatabase` with Qt SQL driver
- Schema: `CREATE VIRTUAL TABLE query_history USING fts5(query, connection_id, database_name, execution_time, row_count, was_successful, error_message, created_at)`
- Auto-cleanup: queries older than 30 days pruned on app startup via `QTimer::singleShot()`
```cpp
// query/HistoryManager.hpp
class HistoryManager : public QObject {
    Q_OBJECT
public:
    void saveQuery(const QueryExecution& execution);
    QList<QueryExecution> searchHistory(const QString& query);
    void cleanupOldEntries(int daysToKeep = 30);
};
```

### 4. Tab State (JSON per connection via QFile)
- Path: `~/.config/tablepro/tabs/{connectionUUID}.json`
- Saved explicitly on: tab switch, tab close, window close, app quit
- Queries > 500KB are truncated before persisting
- Synchronous write on app quit (via `closeEvent()` handler)
```cpp
// core/TabManager.hpp
void TabManager::saveTabsSync(const QUuid& connectionId) {
    // Synchronous save - used during app quit
    // Blocks until file is written
}
```

### 5. Other Storages
- `FilterSettings`: JSON file per connection for saved column filters
- `LicenseManager`: Encrypted license key + Ed25519 signature verification
- `AIChatHistory`: Conversation history as JSON via `QJsonDocument`

## Data Change Tracking (C++ backend)
```cpp
// core/DataChangeManager.hpp
class DataChangeManager : public QObject {
    Q_OBJECT
public:
    struct CellChange {
        int rowIndex;
        QString column;
        QVariant originalValue;
        QVariant newValue;
        QVariantMap primaryKey;  // identity for WHERE clause
    };

    void updateCell(const QUuid& tabId, int row, const QString& col, const QVariant& value);
    void insertRow(const QUuid& tabId, const QVariantMap& rowData);
    void deleteRow(const QUuid& tabId, const QVariantMap& rowIdentity);
    QFuture<bool> commit(const QUuid& tabId);
    void undo(const QUuid& tabId);
    void redo(const QUuid& tabId);

signals:
    void changesCommitted(const QUuid& tabId);
    void changesDiscarded(const QUuid& tabId);
    void undoRedoStateChanged(const QUuid& tabId);

private:
    QMutex m_mutex;
    QMap<QUuid, QList<CellChange>> m_changes;
    QMap<QUuid, QList<QVariantMap>> m_insertedRows;
    QMap<QUuid, QList<QVariantMap>> m_deletedRows;
    QMap<QUuid, QVector<ChangeAction>> m_undoStack;
    QMap<QUuid, QVector<ChangeAction>> m_redoStack;
};
```
- Qt frontend sends cell edits via `UpdateCell(tabID, row, col, newValue)`
- C++ tracks deltas and generates dialect-specific SQL on commit
- Undo/Redo managed in C++, Qt reflects current state via signals

## Memory Management

### Dual Ownership Model
```
┌─────────────────────────────────────────────────────────┐
│              MEMORY MANAGEMENT STRATEGY                 │
└─────────────────────────────────────────────────────────┘

Qt Objects (QObject subclasses):
┌──────────────────────────────────────────┐
│  Parent → Child ownership tree           │
│                                          │
│  MainWindow                              │
│  └─► QSplitter (parent: MainWindow)      │
│      ├─► SchemaTreeView (parent: split)  │
│      ├─► QTabWidget (parent: split)      │
│      └─► QDockWidget (parent: split)     │
│                                          │
│  Automatic deletion when parent dies     │
└──────────────────────────────────────────┘

Non-QObject Resources:
┌──────────────────────────────────────────┐
│  Smart pointers (C++ RAII)               │
│                                          │
│  std::unique_ptr<DatabaseDriver>         │
│  std::shared_ptr<QueryResult>            │
│  QScopedPointer<QFile>                   │
│                                          │
│  Automatic deletion when scope exits     │
└──────────────────────────────────────────┘
```

### C++20 Features Used
| Feature | Usage |
|---------|-------|
| `std::optional<T>` | Return type for nullable values |
| `std::unique_ptr<T>` | Exclusive ownership of resources |
| `std::shared_ptr<T>` | Shared ownership (query results) |
| `std::jthread` | Joinable threads for background tasks |
| Concepts | Template constraints for generic code |
| Coroutines | Async operations with `co_await` |

## Signal/Slot Communication
```cpp
// Backend emits signals
class QueryManager : public QObject {
    Q_OBJECT
public:
    void execute(const QUuid& sessionId, const QString& sql);

signals:
    void queryStarted(const QUuid& sessionId);
    void queryProgress(const QUuid& sessionId, int percent);
    void queryCompleted(const QUuid& sessionId, const QueryResult& result);
    void queryFailed(const QUuid& sessionId, const QString& error);
};

// Frontend connects via lambdas
connect(m_queryManager, &QueryManager::queryProgress,
        this, [this](const QUuid& id, int percent) {
    m_progressBar->setValue(percent);
    m_statusBar->showMessage(tr("Executing: %1%").arg(percent));
});

// Or via Qt5-style string connections (still supported)
connect(m_queryManager, SIGNAL(queryCompleted(QUuid,QueryResult)),
        this, SLOT(onQueryCompleted(QUuid,QueryResult)));
```

## Thread Safety
```cpp
// Use QMutexLocker for RAII-style locking
void ConnectionManager::updateSession(const QUuid& id, const SessionData& data) {
    QMutexLocker locker(&m_mutex);  // Lock acquired
    m_sessions[id] = data;
    // Lock released automatically when locker goes out of scope
}

// For read-heavy workloads, use QReadWriteLock
class SchemaCache {
    mutable QReadWriteLock m_lock;
    QMap<QString, SchemaData> m_cache;

public:
    SchemaData get(const QString& key) const {
        QReadLocker locker(&m_lock);  // Shared read lock
        return m_cache.value(key);
    }

    void set(const QString& key, const SchemaData& value) {
        QWriteLocker locker(&m_lock);  // Exclusive write lock
        m_cache[key] = value;
    }
};
```

## QtConcurrent for Background Tasks
```cpp
// Run heavy operations in thread pool
QFuture<QueryResult> QueryManager::executeAsync(
    const QUuid& sessionId,
    const QString& sql)
{
    return QtConcurrent::run([=]() {
        // Heavy database operation runs in thread pool
        auto* driver = getDriver(sessionId);
        return driver->execute(sql);
    });
}

// Monitor completion
auto future = m_queryManager->executeAsync(sessionId, sql);
auto* watcher = new QFutureWatcher<QueryResult>(this);
connect(watcher, &QFutureWatcher<QueryResult>::finished,
        this, [this, watcher]() {
    emit queryCompleted(watcher->result());
    watcher->deleteLater();
});
```
