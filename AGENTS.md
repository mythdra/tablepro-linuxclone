# TablePro - AGENTS.md

Development guide for AI agents working on TablePro (C++20 + Qt 6.6 database client).

## Project Overview

TablePro is a native cross-platform database client targeting macOS, Windows, and Linux.

**Architecture:**
- **Language**: C++20
- **GUI Framework**: Qt 6.6 LTS (Qt Widgets)
- **Build System**: CMake 3.24+ + vcpkg
- **Database Drivers**: Qt SQL + native C libraries (libpq, libmysql, etc.)
- **Specs**: All specifications in `/specs/`

## Build & Development Commands

### Project Setup
```bash
# Install vcpkg (if not already installed)
git clone https://github.com/microsoft/vcpkg.git
cd vcpkg && ./bootstrap-vcpkg.sh

# Set VCPKG_ROOT environment variable
export VCPKG_ROOT=/path/to/vcpkg

# Install Qt6 and dependencies via vcpkg
vcpkg install qt6-base qt6-svg qt6-tools qt6-translation
vcpkg install libpq libmysql sqlite3 duckdb hiredis
vcpkg install libssh2 openssl qkeychain
```

### Configure & Build
```bash
# Create build directory
mkdir build && cd build

# Configure with CMake (Debug)
cmake .. -DCMAKE_BUILD_TYPE=Debug \
         -DCMAKE_TOOLCHAIN_FILE=$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake

# Configure with CMake (Release)
cmake .. -DCMAKE_BUILD_TYPE=Release \
         -DCMAKE_TOOLCHAIN_FILE=$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake

# Build
cmake --build . --parallel 8

# Run application
./tablepro          # macOS/Linux
Debug\tablepro.exe  # Windows
```

### Testing
```bash
# Run all tests via CTest
ctest --output-on-failure

# Run specific test
ctest -R DriverTest --verbose

# Run tests with coverage (if enabled)
cmake .. -DENABLE_COVERAGE=ON
cmake --build .
ctest --coverage
```

### Linting & Formatting
```bash
# Format C++ code
clang-format -i src/**/*.cpp src/**/*.hpp

# Check formatting without modifying
clang-format --dry-run --Werror src/**/*.cpp src/**/*.hpp

# Run clang-tidy
clang-tidy src/**/*.cpp -- -Iinclude

# CMake lint
cmake-format -i CMakeLists.txt
```

### Qt-Specific Commands
```bash
# Run Qt Designer on .ui files
designer src/ui/ConnectionDialog.ui

# Update Qt translations
lupdate src -ts translations/tablepro_en.ts

# Deploy on macOS
macdeployqt tablepro.app

# Deploy on Windows
windeployqt tablepro.exe
```

## C++ Coding Conventions

### Include Organization
```cpp
// 1. Corresponding header (for .cpp files)
#include "ConnectionManager.hpp"

// 2. Standard library
#include <memory>
#include <string>
#include <vector>
#include <optional>
#include <functional>

// 3. Qt framework
#include <QObject>
#include <QString>
#include <QVector>
#include <QSqlDatabase>

// 4. Third-party libraries
#include <pqxx/pqxx>
#include <mysql.h>

// 5. Project headers
#include "core/DatabaseDriver.hpp"
#include "ui/ConnectionDialog.hpp"
```

### Naming Conventions
- **Classes/Structs**: PascalCase (e.g., `DatabaseConnection`, `QueryResult`)
- **Interfaces**: PascalCase with `I` prefix OR pure abstract class (e.g., `IDatabaseDriver`, `DatabaseDriverInterface`)
- **Functions/Methods**: camelCase (e.g., `executeQuery`, `connectToDatabase`)
- **Variables**: camelCase, descriptive names (e.g., `connectionPool`, `activeSession`)
- **Member variables**: `m_` prefix (e.g., `m_connection`, `m_queryCache`)
- **Constants**: `kPascalCase` (e.g., `kDefaultTimeout`, `kMaxRetries`)
- **Enums**: PascalCase, values with `k` prefix (e.g., `ConnectionState::kConnected`)
- **Files**: PascalCase for classes (e.g., `ConnectionManager.hpp`, `ConnectionManager.cpp`)

### Header Guard
```cpp
// Modern C++17: #pragma once (simpler, widely supported)
#pragma once

// Alternative: traditional include guard
#ifndef TABLEPRO_CONNECTION_MANAGER_HPP
#define TABLEPRO_CONNECTION_MANAGER_HPP
#endif
```

### Smart Pointer Usage
```cpp
// Ownership: use std::unique_ptr by default
std::unique_ptr<DatabaseDriver> m_driver;

// Shared ownership: use std::shared_ptr (Qt objects use parent-child)
std::shared_ptr<QueryCache> m_cache;

// Raw pointers: for non-owning references (Qt QObject* uses parent-child)
DatabaseConnection* m_activeConnection;

// Factory function returning unique_ptr
static std::unique_ptr<DatabaseDriver> create(DatabaseType type);

// Qt objects: use parent-child memory management
auto* button = new QPushButton(this);  // this takes ownership
```

### Error Handling
```cpp
// Use exceptions for exceptional cases
try {
    m_driver->connect(config);
} catch (const DatabaseException& e) {
    emit connectionFailed(tr("Connection failed: %1").arg(e.what()));
    return;
}

// Use std::optional for values that may not exist
std::optional<Connection> getConnection(const QString& name) const;

// Use std::expected (C++23) or custom Result type for recoverable errors
// For C++20, use std::pair<value, error> or custom Result<T>
Result<QueryResult> executeQuery(const QString& sql);

// Qt-style signal for async errors
signals:
    void errorOccurred(const QString& message, ErrorCode code);
```

### Error Handling Best Practices

**1. Exception Safety**
- All database operations wrapped in try-catch
- Use RAII for resource cleanup (automatic on exception)
- Never throw across DLL/shared library boundaries

**2. User-Friendly Error Messages**
```cpp
// Bad: cryptic error
throw DatabaseException("PQconnectdb failed");

// Good: actionable error
throw DatabaseException(
    tr("Database connection failed: Host %1 port %2. "
       "Verify the database server is running and accessible.")
        .arg(config.host).arg(config.port)
);
```

**3. Qt Error Display**
```cpp
// In UI layer - convert exceptions to user messages
void ConnectionDialog::onConnectClicked() {
    try {
        m_manager->connect(m_config);
    } catch (const ConnectionException& e) {
        if (e.message().contains("connection refused")) {
            showNotification(tr("Cannot connect. Check if database is running."));
        } else if (e.message().contains("authentication")) {
            showNotification(tr("Authentication failed. Check username/password."));
        } else {
            showNotification(e.message());
        }
        qWarning() << "Connection failed:" << e.what();
    }
}
```

**4. Error Recovery Patterns**
- Retry transient errors (network timeouts) with exponential backoff
- Fail fast on permanent errors (authentication, syntax errors)
- Provide recovery actions in UI: "Retry", "Edit Query", "Close Connection"

**5. Testing Error Paths**
```cpp
TEST(ConnectionTest, handlesConnectionRefused) {
    ConnectionManager manager;
    ConnectionConfig config{"localhost", 9999, "test", "user", "pass"};

    EXPECT_THROW({
        manager.connect(config);
    }, ConnectionException);
}

TEST(ConnectionTest, timeoutCancelsQuery) {
    ConnectionManager manager;
    manager.setTimeout(std::chrono::milliseconds(100));

    EXPECT_THROW({
        manager.executeQuery("SELECT pg_sleep(1)");
    }, QueryTimeoutException);
}
```

### Class Definition Style
```cpp
#pragma once

#include <QObject>
#include <QUuid>
#include <QString>

namespace tablepro {

class DatabaseConnection {
    Q_GADGET  // Enable Qt meta-object features without QObject overhead

public:
    // Constructors
    DatabaseConnection() = default;
    explicit DatabaseConnection(const QUuid& id, const QString& name);
    ~DatabaseConnection() = default;

    // Copy/move semantics
    DatabaseConnection(const DatabaseConnection&) = default;
    DatabaseConnection& operator=(const DatabaseConnection&) = default;
    DatabaseConnection(DatabaseConnection&&) noexcept = default;
    DatabaseConnection& operator=(DatabaseConnection&&) noexcept = default;

    // Properties (Qt-style with getters/setters)
    QUuid id() const { return m_id; }
    void setId(const QUuid& id) { m_id = id; }

    QString name() const { return m_name; }
    void setName(const QString& name) { m_name = name; }

    DatabaseType type() const { return m_type; }
    void setType(DatabaseType type) { m_type = type; }

    // Connection details
    QString host() const { return m_host; }
    void setHost(const QString& host) { m_host = host; }

    int port() const { return m_port; }
    void setPort(int port) { m_port = port; }

    // Password is NEVER stored in struct - use QKeychain
    QString username() const { return m_username; }
    void setUsername(const QString& username) { m_username = username; }

private:
    QUuid m_id;
    QString m_name;
    DatabaseType m_type;
    QString m_host;
    int m_port{5432};  // Default PostgreSQL port
    QString m_username;
    // m_password intentionally excluded - use Keychain
};

} // namespace tablepro
```

### Concurrency Patterns
```cpp
// QtConcurrent for parallel algorithms
auto future = QtConcurrent::map(results, [](auto& row) {
    processRow(row);
});

// QThread for worker objects
class Worker : public QObject {
    Q_OBJECT
public slots:
    void process() {
        // Long-running task
        emit resultReady(value);
    }
signals:
    void resultReady(const QString& value);
};

// Usage
auto* worker = new Worker;
auto* thread = new QThread;
worker->moveToThread(thread);
connect(thread, &QThread::started, worker, &Worker::process);
connect(worker, &Worker::resultReady, [=](const QString& v) {
    thread->quit();
    thread->wait();
});
thread->start();

// Modern C++20: std::jthread (joining thread)
std::jthread worker([](std::stop_token st) {
    while (!st.stop_requested()) {
        // Work with periodic stop checks
    }
});

// Qt + C++20: QPromise + QFuture
QFuture<QueryResult> future = QtConcurrent::run([=]() {
    return executeLongQuery(sql);
});
```

### Qt Signal/Slot Pattern
```cpp
class QueryExecutor : public QObject {
    Q_OBJECT

public:
    explicit QueryExecutor(QObject* parent = nullptr);

public slots:
    // Called from UI to execute query
    void executeQuery(const QString& sql);

    // Called to cancel running query
    void cancelQuery();

signals:
    // Emitted when query starts
    void executionStarted();

    // Emitted when query completes
    void executionFinished(const QueryResult& result);

    // Emitted on error
    void executionError(const QString& message, ErrorCode code);

    // Emitted for progress updates
    void progressUpdated(int percent);

private:
    QSqlDatabase m_database;
    std::atomic<bool> m_running{false};
};
```

### String Handling
```cpp
// Qt QString for UI text (UTF-16 internally)
QString name = tr("Database Connection");  // tr() for translation

// std::string for internal processing (UTF-8)
std::string utf8Sql = sql.toUtf8().toStdString();

// QStringLiteral for compile-time QString (efficient)
constexpr auto kDefaultHost = QStringLiteral("localhost");

// Raw string literals for SQL (avoid escaping)
const QString query = R"(
    SELECT id, name, created_at
    FROM users
    WHERE status = 'active'
    ORDER BY created_at DESC
)";

// String concatenation
QString full = name + QStringLiteral(":") + QString::number(port);

// String formatting (Qt 6.6+)
QString formatted = tr("Connected to %1:%2").arg(host).arg(port);
```

### Logging
```cpp
// Qt logging categories
Q_LOGGING_CATEGORY(dbCategory, "tablepro.database")
Q_LOGGING_CATEGORY(uiCategory, "tablepro.ui")

// Usage
qCDebug(dbCategory) << "Executing query:" << sql;
qCInfo(dbCategory) << "Query completed in" << elapsedMs << "ms";
qCWarning(dbCategory) << "Connection pool exhausted, waiting...";
qCCritical(dbCategory) << "Database connection lost:" << error;

// Install custom message handler for file logging
void customMessageHandler(QtMsgType type, const QMessageLogContext& context, const QString& msg) {
    // Write to log file with timestamp, category, etc.
}
```

## Qt/Coding Patterns

### QObject Parent-Child Memory Management
```cpp
// Parent takes ownership of children - automatic cleanup
auto* mainWindow = new QMainWindow;
auto* centralWidget = new QWidget(mainWindow);      // mainWindow owns
auto* layout = new QVBoxLayout(centralWidget);      // centralWidget owns
auto* button = new QPushButton(tr("Click"), layout); // layout owns

// When mainWindow is deleted, all children are automatically deleted

// For non-QObject objects, use smart pointers
std::unique_ptr<DatabaseDriver> m_driver;
std::vector<std::unique_ptr<Connection>> m_connections;
```

### Qt Properties System
```cpp
class ConnectionSettings : public QObject {
    Q_OBJECT
    Q_PROPERTY(QString host READ host WRITE setHost NOTIFY hostChanged)
    Q_PROPERTY(int port READ port WRITE setPort NOTIFY portChanged)
    Q_PROPERTY(bool sslEnabled READ sslEnabled WRITE setSslEnabled NOTIFY sslChanged)

public:
    QString host() const { return m_host; }
    void setHost(const QString& host) {
        if (m_host != host) {
            m_host = host;
            emit hostChanged(host);
        }
    }

signals:
    void hostChanged(const QString& newHost);
    void portChanged(int newPort);
    void sslChanged(bool enabled);

private:
    QString m_host;
    int m_port{5432};
    bool m_sslEnabled{false};
};
```

### Model/View for Data Grid
```cpp
// Custom model for query results
class QueryResultModel : public QAbstractTableModel {
    Q_OBJECT

public:
    explicit QueryResultModel(QObject* parent = nullptr);

    // Required overrides
    int rowCount(const QModelIndex& parent = QModelIndex()) const override;
    int columnCount(const QModelIndex& parent = QModelIndex()) const override;
    QVariant data(const QModelIndex& index, int role) const override;
    QVariant headerData(int section, Qt::Orientation orientation, int role) const override;

    // For editing support
    Qt::ItemFlags flags(const QModelIndex& index) const override;
    bool setData(const QModelIndex& index, const QVariant& value, int role) override;

    // For large datasets (pagination/virtual loading)
    bool canFetchMore(const QModelIndex& parent) const override;
    void fetchMore(const QModelIndex& parent) override;

public slots:
    void setQueryResult(const QueryResult& result);

private:
    QueryResult m_result;
    int m_fetchedRows{0};
};

// Usage with QTableView
auto* tableView = new QTableView;
auto* model = new QueryResultModel(tableView);
tableView->setModel(model);
tableView->setItemDelegate(new SqlItemDelegate(tableView));
```

### QScintilla for SQL Editor
```cpp
#include <Qsci/qsciscintilla.h>
#include <Qsci/qscilexersql.h>

class SqlEditor : public QsciScintilla {
    Q_OBJECT

public:
    explicit SqlEditor(QWidget* parent = nullptr);

    void setupSyntaxHighlighting();
    void setupAutoCompletion();
    void setupBraceMatching();

public slots:
    void executeCurrentStatement();
    void executeAllStatements();
    void formatSql();

signals:
    void executeRequested(const QString& sql);
    void executeAllRequested(const QStringList& statements);

private:
    QsciLexerSQL* m_lexer;
    QsciAPIs* m_apis;
};
```

## Architecture Patterns

### Module Organization
```
tablepro/
├── CMakeLists.txt              # Root CMake configuration
├── vcpkg.json                  # vcpkg dependencies
├── README.md                   # Project overview
├── docs/                       # Documentation
│   ├── architecture.md
│   ├── drivers/
│   └── build/
├── src/
│   ├── core/                   # Business logic (no Qt GUI dependencies)
│   │   ├── DatabaseDriver.hpp
│   │   ├── PostgresDriver.cpp
│   │   ├── MysqlDriver.cpp
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

### Key Design Decisions
- **No WebView**: Native Qt Widgets for all UI
- **Password Security**: QKeychain for secure storage
- **Query History**: SQLite with FTS5 via Qt SQL
- **Tab State**: JSON files per connection UUID
- **Large Datasets**: QTableView with custom model, fetch-on-scroll
- **Memory Management**: Qt parent-child + RAII (smart pointers)
- **Threading**: QtConcurrent + QThread for background work

## Testing Guidelines

### Unit Tests with Qt Test
```cpp
#include <QtTest/QtTest>

class TestConnectionManager : public QObject {
    Q_OBJECT

private slots:
    void testSaveConnection_data();
    void testSaveConnection();

    void testInvalidHost();
    void testConnectionTimeout();

private:
    ConnectionManager* m_manager;
};

void TestConnectionManager::testSaveConnection_data() {
    QTest::addColumn<QString>("name");
    QTest::addColumn<QString>("host");
    QTest::addColumn<int>("port");

    QTest::newRow("postgres local") << "Local PG" << "localhost" << 5432;
    QTest::newRow("mysql remote") << "Remote MySQL" << "192.168.1.100" << 3306;
}

void TestConnectionManager::testSaveConnection() {
    QFETCH(QString, name);
    QFETCH(QString, host);
    QFETCH(int, port);

    ConnectionConfig config{name, host, port, "test", "user"};
    bool result = m_manager->save(config);

    QVERIFY(result);
    QCOMPARE(m_manager->connectionNames().contains(name), true);
}

QTEST_MAIN(TestConnectionManager)
#include "test_connectionmanager.moc"
```

### Test with Google Test (Alternative)
```cpp
#include <gtest/gtest.h>

TEST(DatabaseDriverTest, ConnectsSuccessfully) {
    auto driver = DatabaseDriverFactory::create(DatabaseType::PostgreSQL);

    ConnectionConfig config;
    config.host = "localhost";
    config.port = 5432;

    EXPECT_NO_THROW(driver->connect(config));
    EXPECT_TRUE(driver->isConnected());
}

TEST(DatabaseDriverTest, HandlesConnectionRefused) {
    auto driver = DatabaseDriverFactory::create(DatabaseType::PostgreSQL);

    ConnectionConfig config;
    config.host = "localhost";
    config.port = 9999;  // Invalid port

    EXPECT_THROW(driver->connect(config), ConnectionException);
}
```

### UI Tests with Qt Test
```cpp
void TestConnectionDialog::testFormValidation() {
    ConnectionDialog dialog;

    // Empty name should be invalid
    dialog.setNameField("");
    dialog.setHostField("localhost");
    QVERIFY(!dialog.isFormValid());

    // Valid input should be valid
    dialog.setNameField("Test DB");
    dialog.setHostField("localhost");
    QVERIFY(dialog.isFormValid());
}

void TestConnectionDialog::testSaveButtonClick() {
    ConnectionDialog dialog;

    QTest::mouseClick(dialog.saveButton(), Qt::LeftButton);

    // Verify save was called
    QCOMPARE(m_mockManager->saveCalled, true);
}
```

## Git & Workflow

### Branch Naming
- `feature/connection-manager`
- `bugfix/query-pagination-crash`
- `refactor/driver-interface`
- `docs/architecture-update`
- `test/add-driver-tests`

### Commit Messages
```
feat: add PostgreSQL driver with SSH tunnel support

- Implement libpq-based driver wrapper
- Add SSH tunnel via libssh2 + QtNetwork
- Store passwords in QKeychain
- Add unit tests for connection flow

Closes #42
```

### Pre-commit Checklist
- [ ] `clang-format` applied
- [ ] `cmake --build` succeeds
- [ ] `ctest` passes
- [ ] No compiler warnings (`-Wall -Wextra -Wpedantic`)
- [ ] Qt slots/signals properly connected
- [ ] Memory management verified (no leaks)

## AI Agent Guidelines

1. **Read specs first**: Always check `/specs/` before implementing
2. **Match patterns**: Follow existing C++/Qt style exactly
3. **No raw pointers**: Use smart pointers or Qt parent-child
4. **RAII always**: Resource acquisition is initialization
5. **No password in structs**: Always use QKeychain
6. **Exception handling**: Wrap all database operations
7. **Test before complete**: Run `ctest` on changed components
8. **Qt meta-object**: Remember `Q_OBJECT` macro in classes with signals/slots
9. **Thread safety**: Use `QMutex` or C++20 `std::atomic` for shared state
10. **String types**: QString for UI, std::string for internal

## CMake/vcpkg Configuration

### Root CMakeLists.txt
```cmake
cmake_minimum_required(VERSION 3.24)
project(TablePro VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_AUTOUIC ON)

find_package(Qt6 6.6 REQUIRED COMPONENTS
    Core Gui Widgets Sql Network Concurrent
)

find_package(PkgConfig REQUIRED)
pkg_check_modules(LIBPQ REQUIRED libpq)

add_executable(tablepro
    src/main.cpp
    src/core/DatabaseDriver.cpp
    src/ui/MainWindow.cpp
    # ... more sources
)

target_link_libraries(tablepro PRIVATE
    Qt6::Core
    Qt6::Gui
    Qt6::Widgets
    Qt6::Sql
    Qt6::Network
    Qt6::Concurrent
    ${LIBPQ_LIBRARIES}
)

target_include_directories(tablepro PRIVATE
    ${LIBPQ_INCLUDE_DIRS}
)
```

### vcpkg.json
```json
{
  "name": "tablepro",
  "version": "1.0.0",
  "dependencies": [
    "qt6-base",
    "qt6-svg",
    "qt6-tools",
    "libpq",
    "libmysql",
    "sqlite3",
    "duckdb",
    "hiredis",
    "libssh2",
    "openssl",
    "qkeychain"
  ],
  "features": {
    "testing": {
      "description": "Enable testing with Qt Test and Google Test",
      "dependencies": ["gtest", "qttest"]
    }
  }
}
```

## Key Dependencies

### C++ Libraries (via vcpkg)
| Library | Purpose |
|---------|---------|
| `qt6-base` | Core Qt framework (Core, Gui, Widgets) |
| `qt6-sql` | Database abstraction layer |
| `qt6-network` | Network access, SSL/TLS |
| `qt6-concurrent` | Thread pool, parallel algorithms |
| `libpq` | PostgreSQL C client library |
| `libmysql` | MySQL C client library |
| `sqlite3` | SQLite embedded database |
| `duckdb` | DuckDB analytical database |
| `hiredis` | Redis C client library |
| `libssh2` | SSH2 protocol for tunneling |
| `openssl` | SSL/TLS cryptography |
| `qkeychain` | Cross-platform secure storage |

### Qt Modules
| Module | Purpose |
|--------|---------|
| Qt Core | Foundation, containers, threading |
| Qt Gui | Rendering, fonts, images |
| Qt Widgets | Classic desktop UI components |
| Qt Sql | Database driver abstraction |
| Qt Network | HTTP, TCP, SSL |
| Qt Concurrent | Parallel processing |
| Qt Test | Unit testing framework |

## Platform-Specific Notes

### macOS
- Use `macdeployqt` for deployment
- Keychain integration via QKeychain
- Native title bar with `QMainWindow::setWindowFlags()`

### Windows
- Use `windeployqt` for deployment
- Credential Manager via QKeychain
- Native file dialogs automatically used

### Linux
- Deploy as AppImage, Flatpak, or native packages
- libsecret via QKeychain
- Test on Ubuntu 22.04+, Fedora 38+
