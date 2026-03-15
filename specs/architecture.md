# TablePro System Architecture (C++20 + Qt 6.6)

## Overview
TablePro is a native cross-platform database client built with **C++20** (modern language features) and **Qt 6.6 LTS** (native widgets). It targets macOS, Windows, and Linux as native binaries (~20-40MB).

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────┐
│              Qt Widgets Frontend                        │
│  ┌──────┐ ┌──────┐ ┌────────┐ ┌──────┐                │
│  │Sidebar│ │Editor│ │DataGrid│ │Toolbar│                │
│  │QWidget│ │QSci  │ │QTableView│ │QWidget│              │
│  └──┬───┘ └──┬───┘ └───┬────┘ └──┬───┘                │
│     └────────┴─────────┴─────────┘                      │
│              Qt Signals/Slots                           │
│              QtConcurrent/QThread                       │
├─────────────────────────────────────────────────────────┤
│              C++20 Core Layer                           │
│  ┌─────────────┐  ┌──────────────────┐                 │
│  │ConnectionMgr│  │ DatabaseManager  │                 │
│  │TabManager   │  │ ExportService    │                 │
│  │SettingsMgr  │  │ ImportService    │                 │
│  │HistoryMgr   │  │ ChangeTracker    │                 │
│  └──────┬──────┘  └────────┬─────────┘                 │
│         └──────────────────┘                            │
│              Driver Interface                           │
│  ┌──────┐ ┌─────┐ ┌──────┐ ┌─────────┐                │
│  │ libpq│ │mysql│ │sqlite│ │ duckdb  │                 │
│  └──────┘ └─────┘ └──────┘ └─────────┘                 │
└─────────────────────────────────────────────────────────┘
```

## Module Organization (C++)

```
tablepro/
├── CMakeLists.txt              # Root CMake configuration
├── vcpkg.json                  # vcpkg dependencies
├── README.md                   # Project overview
├── src/
│   ├── core/                   # Business logic (no Qt GUI)
│   │   ├── DatabaseDriver.hpp  # Abstract driver interface
│   │   ├── PostgresDriver.cpp  # libpq implementation
│   │   ├── MysqlDriver.cpp     # libmysql implementation
│   │   ├── SqliteDriver.cpp    # Qt SQL SQLite
│   │   ├── DuckDbDriver.cpp    # DuckDB C API
│   │   ├── ConnectionManager.cpp
│   │   ├── QueryExecutor.cpp
│   │   ├── ChangeTracker.cpp
│   │   └── SqlGenerator.cpp
│   ├── ui/                     # Qt UI components
│   │   ├── MainWindow.cpp
│   │   ├── ConnectionDialog.cpp
│   │   ├── DataGrid/
│   │   │   ├── QueryResultView.cpp
│   │   │   └── QueryResultModel.cpp
│   │   ├── Editor/
│   │   │   ├── SqlEditor.cpp
│   │   │   └── SqlHighlighter.cpp
│   │   └── Widgets/
│   ├── services/               # Application services
│   │   ├── ExportService.cpp
│   │   ├── ImportService.cpp
│   │   └── HistoryService.cpp
│   └── main.cpp                # Application entry point
├── resources/
│   ├── icons/
│   ├── styles/
│   └── translations/
└── tests/
    ├── unit/
    ├── integration/
    └── ui/
```

## Communication: Qt Signals/Slots

### Core → UI (Signals)

```cpp
// Core class with signals
class ConnectionManager : public QObject {
    Q_OBJECT

public:
    explicit ConnectionManager(QObject* parent = nullptr);

public slots:
    void connectToDatabase(const ConnectionConfig& config);
    void disconnect();

signals:
    void connected(const ConnectionInfo& info);
    void disconnected();
    void errorOccurred(const QString& message, ErrorCode code);
    void statusChanged(ConnectionStatus status);
};

// UI connects to core
auto* manager = new ConnectionManager(this);
connect(manager, &ConnectionManager::connected,
        this, &MainWindow::onDatabaseConnected);
connect(manager, &ConnectionManager::errorOccurred,
        this, &MainWindow::onDatabaseError);
```

### UI → Core (Slots)

```cpp
// UI calls core methods directly (synchronous)
void MainWindow::onConnectClicked() {
    ConnectionConfig config = buildConfigFromForm();
    m_manager->connectToDatabase(config);
    // Returns immediately, result via signal
}

// Or async via QtConcurrent
void MainWindow::executeQueryAsync(const QString& sql) {
    auto future = QtConcurrent::run([=]() {
        return m_executor->execute(sql);
    });

    // Connect to future's finished signal
    auto* watcher = new QFutureWatcher<QueryResult>(this);
    connect(watcher, &QFutureWatcher<QueryResult>::finished,
            this, [=]() {
        emit queryResultReady(watcher->result());
        watcher->deleteLater();
    });
    watcher->setFuture(future);
}
```

## Data Flow

```
1. User action in UI (button click, key press)
         ↓
2. Slot called in UI component
         ↓
3. Core method invoked (C++ function call)
         ↓
4. Business logic executes (driver, query, etc.)
         ↓
5. Signal emitted if state changed
         ↓
6. UI slot updates widgets
         ↓
7. Qt triggers repaint
```

## Memory Management

```
┌─────────────────────────────────────────────────────────┐
│              MEMORY MANAGEMENT STRATEGY                 │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  Qt Objects (QObject subclasses)                        │
│  ═══════════════════════════════════════                │
│  • Parent-child ownership tree                          │
│  • Parent automatically deletes children                │
│  • Raw pointers OK when parent owns child               │
│                                                         │
│  auto* mainWindow = new QMainWindow(this);              │
│  auto* centralWidget = new QWidget(mainWindow);         │
│  // mainWindow owns centralWidget                       │
│                                                         │
│  ─────────────────────────────────────────────────────  │
│                                                         │
│  Non-QObject Objects (POD, services, drivers)           │
│  ═══════════════════════════════════════                │
│  • std::unique_ptr for exclusive ownership              │
│  • std::shared_ptr for shared ownership                 │
│  • Raw pointers for non-owning references               │
│                                                         │
│  std::unique_ptr<DatabaseDriver> m_driver;              │
│  DatabaseDriver* driver = m_driver.get();  // borrow    │
│                                                         │
│  ─────────────────────────────────────────────────────  │
│                                                         │
│  RAII Pattern                                           │
│  ═══════════════════════════════════════                │
│  • Resources acquired in constructor                    │
│  • Resources released in destructor                     │
│  • Exception-safe, no manual cleanup                    │
│                                                         │
│  class DatabaseConnection {                             │
│      std::unique_ptr<PQconn, PqDeleter> m_conn;         │
│      // Automatically closed on destruction             │
│  };                                                     │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## Storage & Persistence

| Data Type | Storage Mechanism | Location |
|-----------|-------------------|----------|
| Passwords | QKeychain (OS Keychain) | macOS Keychain / Windows Credential Manager / libsecret |
| Preferences | JSON file | `~/.config/tablepro/settings.json` |
| Query History | SQLite with FTS5 | `~/.config/tablepro/history.db` |
| Tab State | JSON per connection | `~/.config/tablepro/tabs/{uuid}.json` |
| SSH Keys | QKeychain + file paths | OS Keychain + `~/.ssh/` |

## Third-party Dependencies (C++)

| Library | Purpose |
|---------|---------|
| Qt 6.6 LTS | GUI framework, SQL abstraction, networking |
| libpq | PostgreSQL C client library |
| MySQL Connector/C | MySQL/MariaDB C client |
| Qt SQL SQLite | SQLite3 (built into Qt) |
| DuckDB | Embedded analytical database |
| hiredis | Redis C client library |
| libssh2 | SSH tunneling support |
| QKeychain | Cross-platform keychain wrapper |
| QScintilla | Code editor with syntax highlighting |
| spdlog | Fast C++ logging library |
| nlohmann/json | JSON library |
| Catch2 | Unit testing framework |

## Threading Model

```cpp
// Background work with QtConcurrent
QFuture<QueryResult> future = QtConcurrent::run([=]() {
    // Runs in thread pool
    return m_driver->execute(query);
});

// Monitor with QFutureWatcher
auto* watcher = new QFutureWatcher<QueryResult>(this);
connect(watcher, &QFutureWatcher::finished, this, [=]() {
    // Runs in main thread
    displayResults(watcher->result());
    watcher->deleteLater();
});
watcher->setFuture(future);

// Long-running worker with QThread
class Worker : public QObject {
    Q_OBJECT
public slots:
    void process() {
        // Long task
        emit resultReady(value);
    }
signals:
    void resultReady(const QVariant& value);
};

auto* worker = new Worker;
auto* thread = new QThread;
worker->moveToThread(thread);
connect(thread, &QThread::started, worker, &Worker::process);
connect(worker, &Worker::resultReady, this, &MainWindow::updateUi);
connect(worker, &Worker::resultReady, thread, &QThread::quit);
thread->start();

// C++20: std::jthread (joining thread)
std::jthread worker([](std::stop_token st) {
    while (!st.stop_requested()) {
        // Work with periodic stop checks
    }
});
```

## C++20 Features Used

| Feature | Usage |
|---------|-------|
| `std::optional<T>` | Values that may not exist (e.g., query results) |
| `std::variant<T, U>` | Union types for result handling |
| `std::expected<T, E>` | Error handling without exceptions (C++23 backport) |
| Concepts | Template constraints for driver interfaces |
| `std::jthread` | RAII thread management |
| Coroutines | Async query execution (experimental) |
| `std::format` | Type-safe string formatting |
| Ranges | Functional-style container operations |

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Language | C++20 | Modern features, performance, native compilation |
| GUI Framework | Qt 6.6 LTS | Mature, native widgets, excellent SQL support |
| UI Toolkit | Qt Widgets | Traditional desktop, QTableView ready-to-use |
| Build System | CMake + vcpkg | Cross-platform, excellent C++ package management |
| State Management | Qt Signal/Slot | Built-in, type-safe, thread-aware |
| Data Grid | QTableView + Custom Model | Native, virtual scrolling, editing support |
| SQL Editor | QScintilla | Syntax highlighting, autocomplete, multi-language |
| Password Storage | QKeychain | Cross-platform secure storage |
| Threading | QtConcurrent + QThread | Integrated with Qt event loop |
| Testing | Qt Test + Catch2 | Comprehensive test coverage |
