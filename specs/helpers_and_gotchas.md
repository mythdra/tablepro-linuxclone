# Helpers & Gotchas (C++20 + Qt 6)

## C++-Specific Helpers

### 1. Connection URL Parser
```cpp
// connection/ConnectionUrlParser.hpp
class ConnectionUrlParser {
public:
    struct ParsedConnection {
        bool isSshTunnel = false;
        QString sshHost;
        int sshPort = 22;
        QString sshUser;
        QString sshPassword;
        QString dbHost;
        int dbPort = 5432;
        QString dbUser;
        QString dbPassword;
        QString database;
        DatabaseType dbType;
    };

    static std::optional<ParsedConnection> parse(const QString& rawUrl);

private:
    // Handle dual-@ SSH URLs: postgres+ssh://sshuser@bastion:22/dbuser:pass@10.0.0.1:5432/mydb
    // Qt's QUrl::parsed() breaks on dual @, so use custom regex:
    static inline const QRegularExpression s_sshPattern{
        R"(^(\w+)\+ssh://([^@]+)@([^:/]+):?(\d*)/(.+)$)"
    };
};
```

```cpp
// connection/ConnectionUrlParser.cpp
std::optional<ConnectionUrlParser::ParsedConnection> ConnectionUrlParser::parse(const QString& rawUrl) {
    ParsedConnection result;

    // Check for SSH tunnel URL
    auto sshMatch = s_sshPattern.match(rawUrl);
    if (sshMatch.hasMatch()) {
        result.isSshTunnel = true;
        result.dbType = parseDatabaseType(sshMatch.captured(1));
        result.sshUser = sshMatch.captured(2);
        result.sshHost = sshMatch.captured(3);
        result.sshPort = sshMatch.captured(4).toInt();
        if (result.sshPort == 0) result.sshPort = 22;

        // Parse inner URL (after SSH part)
        QString innerUrl = sshMatch.captured(5);
        return parseInnerUrl(innerUrl, result);
    }

    // Standard URL without SSH
    return parseStandardUrl(rawUrl);
}
```
> **Gotcha**: Qt's `QUrl::parsed()` cannot handle `scheme+ssh://user@host/otheruser@otherhost` — must use manual regex splitting (same issue as Swift/Go).

### 2. RAII for C Libraries
```cpp
// driver/postgres/PostgresDriver.hpp

// Custom deleter for PQconn*
struct PqConnDeleter {
    void operator()(PGconn* conn) const {
        if (conn) PQfinish(conn);
    }
};

// Custom deleter for PGresult*
struct PqResultDeleter {
    void operator()(PGresult* result) const {
        if (result) PQclear(result);
    }
};

class PostgresDriver : public DatabaseDriver {
private:
    std::unique_ptr<PGconn, PqConnDeleter> m_connection;

    // Usage:
    // m_connection.reset(PQconnectdb(conninfo));
    // Automatic cleanup when driver is destroyed
};
```
> **Best Practice**: Always use custom deleters for C library resources — ensures proper cleanup even on exceptions.

### 3. QtConcurrent with Timeout
```cpp
// Use QFutureWatcher with QTimer for timeout
QFuture<QueryResult> executeWithTimeout(
    const QString& sql,
    std::chrono::milliseconds timeout)
{
    auto future = QtConcurrent::run([=]() {
        return driver->execute(sql);
    });

    auto* watcher = new QFutureWatcher<QueryResult>();
    watcher->setFuture(future);

    QEventLoop loop;
    QTimer::singleShot(timeout.count(), &loop, [&]() {
        if (!future.isFinished()) {
            future.cancel();
            loop.exit(1);  // Timeout
        }
    });

    connect(watcher, &QFutureWatcher<QueryResult>::finished,
            &loop, &QEventLoop::quit);

    int result = loop.exec();
    watcher->deleteLater();

    if (result == 1) {
        return QueryResult{.error = "Query timeout"};
    }

    return future.result();
}
```
> **Gotcha**: `QFuture::cancel()` doesn't forcibly stop the thread — the task must check `QFuture::isCanceled()` periodically.

### 4. QVariant Type Handling
```cpp
// Database drivers return QVariant for cell values
// Qt's QVariant handles most types, but:
// - QDateTime → format as ISO 8601 string for display
// - QByteArray → hex encode for binary display
// - Null QVariant → render as "NULL" in grid

QString formatCellValue(const QVariant& value) {
    if (!value.isValid() || value.isNull()) {
        return "NULL";
    }

    if (value.userType() == QMetaType::QDateTime) {
        return value.toDateTime().toString(Qt::ISODate);
    }

    if (value.userType() == QMetaType::QByteArray) {
        return value.toByteArray().toHex();
    }

    return value.toString();
}
```

### 5. Large Query Truncation
```cpp
// core/TabManager.cpp
constexpr qint64 MaxPersistableQuerySize = 500 * 1024;  // 500KB

QString truncateForPersistence(const QString& query) {
    if (query.size() > MaxPersistableQuerySize) {
        return {};  // Don't persist oversized queries
    }
    return query;
}

void TabManager::truncateLargeQueries(QList<PersistedTabInfo>& tabs, qint64 maxSize) {
    for (auto& tab : tabs) {
        if (tab.query.size() > maxSize) {
            tab.query.clear();
        }
    }
}
```

## Qt-Specific Helpers

### 1. Debounced Search
```cpp
// Use QTimer for debouncing
class Debouncer : public QObject {
    Q_OBJECT

public:
    explicit Debouncer(int delayMs, QObject* parent = nullptr);

    template<typename Func>
    void call(Func func) {
        m_timer->stop();
        m_timer->start();
        connect(m_timer, &QTimer::timeout, this, [this, func]() {
            m_timer->stop();
            func();
        }, Qt::SingleShotConnection);
    }

signals:
    void triggered();

private:
    QTimer* m_timer;
};

// Usage in SchemaTreeView:
m_searchDebouncer = new Debouncer(300, this);
connect(m_searchEdit, &QLineEdit::textChanged,
        this, [this](const QString& text) {
    m_searchDebouncer->call([this, text]() {
        m_filterModel->setFilterText(text);
    });
});
```

### 2. Progress Throttling
```cpp
// Import/Query progress - throttle UI updates to ~15fps
class ThrottledProgress : public QObject {
    Q_OBJECT

public:
    explicit ThrottledProgress(int intervalMs = 66, QObject* parent = nullptr);

public slots:
    void update(const ImportProgress& progress);

private:
    QElapsedTimer m_lastUpdate;
    int m_intervalMs;
};

void ThrottledProgress::update(const ImportProgress& progress) {
    if (m_lastUpdate.elapsed() < m_intervalMs) {
        return;  // Throttle
    }
    m_lastUpdate.restart();
    emit progressUpdated(progress);
}
```

### 3. Clipboard Handling
```cpp
// Copy cell or row data to clipboard
void GridCopyActions::copyToClipboard(QWidget* parent, const QString& text) {
    auto* clipboard = QApplication::clipboard();
    clipboard->setText(text);

    // Optional: Also use QKeychain for secure clipboard on some platforms
    // clipboard->setImage(pixmap.toImage()); for rich text
}
```

### 4. Date Formatting
```cpp
// Display database timestamps in user's locale
QString formatDate(const QVariant& value, const QLocale& locale) {
    if (!value.isValid() || value.isNull()) {
        return "NULL";
    }

    QDateTime dt = value.toDateTime();
    return locale.toString(dt, QLocale::ShortFormat);
}

// Or use Qt's built-in formatting
QString formatted = QLocale::system().toString(dt, "yyyy-MM-dd HH:mm:ss");
```

## Architecture Gotchas

### QtConcurrent Memory
- `QTableView` with `QAbstractTableModel` handles millions of rows via virtual scrolling
- Only visible rows are rendered — model holds data in C++ backend
- For 100K rows × 20 columns, keep `QueryResult` under 100MB
- **Mitigation**: Paginate results (default 500 rows per page)
- **Mitigation**: Use `QSharedMemory` for very large result sets

### JSON Transfer Overhead
- `QJsonDocument` serialization for persistence
- For tab state persistence, limit query text to 500KB
- **Mitigation**: Use `QDataStream` for binary serialization (faster, smaller)
- **Mitigation**: Compress with qCompress() for large state

### Thread Safety in Qt
- All GUI operations MUST run on main thread
- Use `QMetaObject::invokeMethod()` to marshal to main thread:
```cpp
// From worker thread:
QMetaObject::invokeMethod(mainWidget, [result]() {
    mainWidget->updateUi(result);  // Runs on main thread
}, Qt::QueuedConnection);
```
- Never hold `QMutex` while emitting signals (deadlock risk)
- Use `QMutexLocker` for RAII-style locking

### Parent-Child Ownership
```cpp
// CORRECT: Qt parent-child ownership
auto* widget = new QWidget(this);  // Deleted when parent dies
auto* layout = new QVBoxLayout(widget);  // Deleted when widget dies

// WRONG: Double ownership
std::unique_ptr<QWidget> widget = std::make_unique<QWidget>(this);  // DON'T
// Qt will try to delete, unique_ptr will try to delete = crash
```
> **Gotcha**: Don't use smart pointers for Qt widgets with parents — Qt manages lifetime.

### Signal/Slot Connection Types
```cpp
// Qt::DirectConnection: Slot called immediately (in emitter's thread)
// Qt::QueuedConnection: Slot called later (in receiver's thread)
// Qt::AutoConnection: Queued if crossing threads, Direct otherwise

// Cross-thread connection (automatically queued)
connect(worker, &Worker::resultReady,
        ui, &UiWidget::updateUi);  // QueuedConnection automatically

// Same-thread connection (direct)
connect(button, &QPushButton::clicked,
        this, &MainWindow::onButtonClicked);  // DirectConnection
```

### Event Loop Gotchas
```cpp
// DON'T: Blocking the main thread
QThread::sleep(5);  // UI freezes!

// DO: Use async patterns
QTimer::singleShot(5000, this, []() {
    // Called after 5 seconds without blocking
});

// DO: Use QEventLoop for local async coordination
QEventLoop loop;
QTimer::singleShot(5000, &loop, &QEventLoop::quit);
loop.exec();  // Blocks this context, not necessarily main thread
```

### QString vs std::string
```cpp
// Prefer QString for Qt APIs (implicit Unicode support)
// Prefer std::string for C libraries and network protocols

// Conversion:
QString qtStr = QString::fromStdString(stdStr);
std::string stdStr = qtStr.toStdString();

// UTF-8 conversion (common for network/DB):
QByteArray utf8 = qtStr.toUtf8();
QString fromUtf8 = QString::fromUtf8(utf8);
```

### File Path Handling
```cpp
// Use QDir for cross-platform paths
QString configPath = QStandardPaths::writableLocation(
    QStandardPaths::AppConfigLocation);

// Use QFile for file operations
QFile file(path);
if (file.open(QIODevice::ReadOnly)) {
    QByteArray data = file.readAll();
}

// Use QFileInfo for metadata
QFileInfo info(path);
if (info.exists() && info.isReadable()) { ... }
```
