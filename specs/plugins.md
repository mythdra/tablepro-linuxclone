# Driver Integration (C++20 + Qt 6)

## Overview
The C++ architecture eliminates the plugin bridge layer entirely. All database drivers implement the same `DatabaseDriver` interface directly, compiled into the same binary. Qt's plugin system (`QPluginLoader`) is available but not used — drivers are linked at compile time for better type safety and performance.

## DatabaseDriver Interface
```cpp
// driver/driver.hpp
class DatabaseDriver : public QObject {
    Q_OBJECT

public:
    explicit DatabaseDriver(QObject* parent = nullptr);
    virtual ~DatabaseDriver() = default;

    // Connection
    QFuture<bool> connect(const ConnectionConfig& config);
    virtual void disconnect() = 0;
    QFuture<bool> testConnection(const ConnectionConfig& config);
    QFuture<void> ping();

    // Query Execution
    QFuture<QueryResult> execute(const QString& sql);
    QFuture<QueryResult> executeWithParams(const QString& sql, const QVariantList& params);

    // Schema Inspection
    QFuture<QList<DatabaseInfo>> fetchDatabases();
    QFuture<QList<SchemaInfo>> fetchSchemas(const QString& database);
    QFuture<QList<TableInfo>> fetchTables(const QString& schema);
    QFuture<QList<TableInfo>> fetchViews(const QString& schema);
    QFuture<QList<RoutineInfo>> fetchRoutines(const QString& schema);
    QFuture<QList<ColumnMetadata>> fetchColumns(const QString& schema, const QString& table);
    QFuture<QList<IndexMetadata>> fetchIndexes(const QString& schema, const QString& table);
    QFuture<QList<ForeignKeyMetadata>> fetchForeignKeys(const QString& schema, const QString& table);
    QFuture<QString> fetchTableDDL(const QString& schema, const QString& table);
    QFuture<qint64> fetchTableRowCount(const QString& schema, const QString& table);

    // Transactions
    virtual bool beginTransaction() = 0;
    virtual bool commitTransaction() = 0;
    virtual bool rollbackTransaction() = 0;

    // Dialect
    virtual DialectInfo dialectInfo() const = 0;
    virtual QStringList foreignKeyDisableStatements() const = 0;
    virtual QStringList foreignKeyEnableStatements() const = 0;

signals:
    void connected();
    void disconnected();
    void connectionError(const QString& error);
    void dataRefreshed();

protected:
    ConnectionConfig m_config;
    bool m_isConnected = false;
    QMutex m_connectionMutex;
};
```

## Driver Factory
```cpp
// driver/factory.hpp
class DriverFactory : public QObject {
    Q_OBJECT

public:
    static std::unique_ptr<DatabaseDriver> createDriver(DatabaseType type, QObject* parent = nullptr);
    static QList<DatabaseType> supportedTypes();
    static bool isTypeSupported(DatabaseType type);

private:
    using DriverCreator = std::function<std::unique_ptr<DatabaseDriver>(QObject*)>;
    static const QMap<DatabaseType, DriverCreator>& getRegistry();
};
```

```cpp
// driver/factory.cpp
const QMap<DatabaseType, DriverFactory::DriverCreator>& DriverFactory::getRegistry() {
    static const QMap<DatabaseType, DriverCreator> registry = {
        {DatabaseType::PostgreSQL,   [](QObject* parent) { return std::make_unique<PostgresDriver>(parent); }},
        {DatabaseType::MySQL,        [](QObject* parent) { return std::make_unique<MySqlDriver>(parent); }},
        {DatabaseType::SQLite,       [](QObject* parent) { return std::make_unique<SQLiteDriver>(parent); }},
        {DatabaseType::DuckDB,       [](QObject* parent) { return std::make_unique<DuckDbDriver>(parent); }},
        {DatabaseType::SqlServer,    [](QObject* parent) { return std::make_unique<SqlServerDriver>(parent); }},
        {DatabaseType::ClickHouse,   [](QObject* parent) { return std::make_unique<ClickHouseDriver>(parent); }},
        {DatabaseType::MongoDB,      [](QObject* parent) { return std::make_unique<MongoDbDriver>(parent); }},
        {DatabaseType::Redis,        [](QObject* parent) { return std::make_unique<RedisDriver>(parent); }},
    };
    return registry;
}

std::unique_ptr<DatabaseDriver> DriverFactory::createDriver(DatabaseType type, QObject* parent) {
    const auto& registry = getRegistry();
    auto it = registry.find(type);
    if (it == registry.end()) {
        throw std::invalid_argument(fmt::format("Unsupported database type: {}", static_cast<int>(type)));
    }
    return it.value()(parent);
}
```

## CMake Options for Optional Drivers
Heavy drivers can be excluded from builds via CMake options:
```cmake
# CMakeLists.txt
option(ENABLE_MONGODB "Enable MongoDB driver (requires libmongocxx)" ON)
option(ENABLE_CLICKHOUSE "Enable ClickHouse driver (requires clickhouse-cpp)" ON)
option(ENABLE_DUCKDB "Enable DuckDB driver (requires duckdb)" ON)

if(ENABLE_MONGODB)
    find_package(libmongocxx REQUIRED)
    target_compile_definitions(tablepro PRIVATE ENABLE_MONGODB)
    target_sources(tablepro PRIVATE internal/driver/mongodb/mongodb.cpp)
endif()
```

```cpp
// driver/factory.cpp - Conditional compilation
const QMap<DatabaseType, DriverFactory::DriverCreator>& DriverFactory::getRegistry() {
    static const QMap<DatabaseType, DriverCreator> registry = {
        {DatabaseType::PostgreSQL, [](QObject* parent) { return std::make_unique<PostgresDriver>(parent); }},
        {DatabaseType::MySQL,      [](QObject* parent) { return std::make_unique<MySqlDriver>(parent); }},
#ifdef ENABLE_MONGODB
        {DatabaseType::MongoDB,    [](QObject* parent) { return std::make_unique<MongoDbDriver>(parent); }},
#endif
#ifdef ENABLE_CLICKHOUSE
        {DatabaseType::ClickHouse, [](QObject* parent) { return std::make_unique<ClickHouseDriver>(parent); }},
#endif
    };
    return registry;
}
```

## Driver Lifecycle
```cpp
// Creating and using a driver
auto driver = DriverFactory::createDriver(DatabaseType::PostgreSQL, this);

// Connect (async)
ConnectionConfig config;
config.host = "localhost";
config.port = 5432;
config.username = "user";
config.password = "secret";
config.database = "mydb";

connect(driver.get(), &DatabaseDriver::connected,
        this, []() { qDebug() << "Connected!"; });

driver->connect(config);

// Execute query
auto future = driver->execute("SELECT * FROM users LIMIT 10");
QFutureWatcher<QueryResult>* watcher = new QFutureWatcher<QueryResult>(this);
watcher->setFuture(future);
connect(watcher, &QFutureWatcher<QueryResult>::finished,
        this, [watcher]() {
    const auto& result = watcher->result();
    // Process result
    watcher->deleteLater();
});
```

## ConnectionManager Service
```cpp
// core/ConnectionManager.hpp
class ConnectionManager : public QObject {
    Q_OBJECT
    Q_PROPERTY(bool hasActiveConnections READ hasActiveConnections NOTIFY connectionCountChanged)

public:
    static ConnectionManager* instance();

    QFuture<bool> addConnection(const QUuid& connectionId, const ConnectionConfig& config);
    void removeConnection(const QUuid& connectionId);
    DatabaseDriver* getDriver(const QUuid& connectionId) const;
    QList<QUuid> activeConnectionIds() const;
    bool hasActiveConnections() const;

    // Health monitoring
    void startHealthCheck(const QUuid& connectionId, std::chrono::seconds interval = 30s);
    void stopHealthCheck(const QUuid& connectionId);

signals:
    void connectionAdded(const QUuid& connectionId);
    void connectionRemoved(const QUuid& connectionId);
    void connectionStatusChanged(const QUuid& connectionId, ConnectionStatus status);
    void connectionError(const QUuid& connectionId, const QString& error);
    void connectionCountChanged();

private:
    QMutableHash<QUuid, std::unique_ptr<DatabaseDriver>> m_drivers;
    QHash<QUuid, QTimer*> m_healthCheckTimers;
    mutable QMutex m_mutex;
};
```

```cpp
// core/ConnectionManager.cpp
void ConnectionManager::startHealthCheck(const QUuid& connectionId, std::chrono::seconds interval) {
    QMutexLocker locker(&m_mutex);

    auto* timer = new QTimer(this);
    timer->setInterval(interval.count());
    connect(timer, &QTimer::timeout, this, [this, connectionId]() {
        auto* driver = getDriver(connectionId);
        if (driver) {
            auto pingFuture = driver->ping();
            // Handle ping timeout/error
        } else {
            // Driver gone, stop timer
            sender()->deleteLater();
        }
    });
    timer->start();
    m_healthCheckTimers[connectionId] = timer;
}
```

## Export/Import Format Handlers
```cpp
// core/ImportExport.hpp
class ExportFormat : public QObject {
    Q_OBJECT

public:
    virtual QString id() const = 0;
    virtual QString name() const = 0;
    virtual QString extension() const = 0;
    virtual QFuture<bool> exportData(
        const QueryResult& source,
        QIODevice* output,
        const ExportOptions& options) = 0;
};

class ImportFormat : public QObject {
    Q_OBJECT

public:
    virtual QString id() const = 0;
    virtual QString name() const = 0;
    virtual QStringList extensions() const = 0;
    virtual QFuture<ImportResult> importData(
        QIODevice* input,
        DatabaseDriver* driver,
        const ImportOptions& options) = 0;
};
```

```cpp
// core/FormatRegistry.hpp
class FormatRegistry : public QObject {
    Q_OBJECT

public:
    static FormatRegistry* instance();

    void registerExportFormat(ExportFormat* format);
    void registerImportFormat(ImportFormat* format);
    ExportFormat* getExportFormat(const QString& id);
    ImportFormat* getImportFormat(const QString& id);
    QStringList supportedExportFormats() const;
    QStringList supportedImportFormats() const;

private:
    QMap<QString, ExportFormat*> m_exportFormats;
    QMap<QString, ImportFormat*> m_importFormats;
};

// Built-in formats registered at startup
class CsvExportFormat : public ExportFormat { /* ... */ };
class JsonExportFormat : public ExportFormat { /* ... */ };
class SqlExportFormat : public ExportFormat { /* ... */ };
class XlsxExportFormat : public ExportFormat { /* ... */ };
```

## Comparison with Swift Plugin System
| Aspect | Swift (Plugins) | Go (Interface) | C++20 + Qt (Interface) |
|---|---|---|---|
| Loading | Runtime `Bundle.load()` | Compile-time linking | Compile-time linking |
| Serialization | JSON/Codable across boundary | Direct Go types | Direct C++ types |
| Error handling | Plugin crashes isolated | `recover()` in same process | Exceptions + RAII cleanup |
| Distribution | Separate `.tableplugin` bundles | Single binary | Single binary + vcpkg dependencies |
| Hot-reload | Possible (reload bundle) | Requires rebuild | Requires rebuild |
| Cross-platform | macOS only | macOS + Windows + Linux | macOS + Windows + Linux |
| Type safety | Runtime cast | Compile-time | Compile-time + virtual functions |
| Memory management | ARC + bridging | GC-managed | RAII + parent-child ownership |

## Qt Plugin System (Optional Extension)
Qt's `QPluginLoader` can be used for third-party extensions if needed:
```cpp
// Optional: Load external driver plugins
class PluginDriverLoader : public QObject {
    Q_OBJECT

public:
    static std::unique_ptr<DatabaseDriver> loadDriverPlugin(
        const QString& pluginPath,
        QObject* parent);

private:
    // Plugin must export this function
    using CreateDriverFunc = DatabaseDriver* (*)(QObject*);
};
```

```cpp
// Plugin implementation (external .dylib/.so/.dll)
// mydb-driver/mydb_driver_plugin.cpp
extern "C" Q_DECL_EXPORT DatabaseDriver* createDatabaseDriver(QObject* parent) {
    return new MyCustomDriver(parent);
}
```

Usage:
```cpp
auto driver = PluginDriverLoader::loadDriverPlugin(
    QCoreApplication::applicationDirPath() + "/plugins/libmydb_driver.dylib",
    this);
```

**Note**: This is optional. Built-in drivers are preferred for stability and distribution.
