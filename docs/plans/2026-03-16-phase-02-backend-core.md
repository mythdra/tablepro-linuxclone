# Phase 2: Backend Core Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build core backend infrastructure including database driver interface, connection manager, query executor, and error handling framework.

**Architecture:** Abstract DatabaseDriver interface with Qt signals/slots for async operations. ConnectionManager handles connection lifecycle. QueryExecutor provides async query execution with QFuture.

**Tech Stack:** C++20, Qt 6.6 (Core, Concurrent), std::optional, std::variant

---

## Task 1: Error Types

**Files:**
- Create: `src/core/errors.hpp`
- Create: `src/core/errors.cpp`

**Step 1: Create errors.hpp**

```cpp
#pragma once

#include <QException>
#include <QString>
#include <QDebug>

namespace tablepro {

enum class ErrorCode {
    Unknown = 0,
    ConnectionFailed = 100,
    ConnectionTimeout = 101,
    AuthenticationFailed = 102,
    QueryFailed = 200,
    QueryTimeout = 201,
    QuerySyntaxError = 202,
    SchemaNotFound = 300,
    TableNotFound = 301,
    DriverNotSupported = 400,
};

class TableProException : public QException {
public:
    explicit TableProException(const QString& message, ErrorCode code = ErrorCode::Unknown);

    const char* what() const noexcept override;
    QString message() const { return m_message; }
    ErrorCode code() const { return m_code; }

    TableProException* clone() const override;
    void raise() const override;

private:
    QString m_message;
    ErrorCode m_code;
    QByteArray m_what;
};

// Specific exception types
class ConnectionException : public TableProException {
public:
    explicit ConnectionException(const QString& message, ErrorCode code = ErrorCode::ConnectionFailed)
        : TableProException(message, code) {}
};

class QueryException : public TableProException {
public:
    explicit QueryException(const QString& message, ErrorCode code = ErrorCode::QueryFailed)
        : TableProException(message, code) {}
};

class SchemaException : public TableProException {
public:
    explicit SchemaException(const QString& message, ErrorCode code = ErrorCode::SchemaNotFound)
        : TableProException(message, code) {}
};

} // namespace tablepro
```

**Step 2: Create errors.cpp**

```cpp
#include "errors.hpp"

namespace tablepro {

TableProException::TableProException(const QString& message, ErrorCode code)
    : m_message(message)
    , m_code(code)
{
    m_what = QString("TablePro Error [%1]: %2")
        .arg(static_cast<int>(code))
        .arg(message)
        .toUtf8();
}

const char* TableProException::what() const noexcept {
    return m_what.constData();
}

TableProException* TableProException::clone() const {
    return new TableProException(m_message, m_code);
}

void TableProException::raise() const {
    throw *this;
}

} // namespace tablepro
```

**Step 3: Commit error types**

```bash
git add src/core/errors.hpp src/core/errors.cpp
git commit -m "feat: Add error types and exception hierarchy"
```

---

## Task 2: Query Result Types

**Files:**
- Create: `src/core/types.hpp`

**Step 1: Create types.hpp**

```cpp
#pragma once

#include <QString>
#include <QStringList>
#include <QVariant>
#include <QVariantMap>
#include <QList>
#include <QJsonObject>
#include <optional>

namespace tablepro {

struct ColumnInfo {
    QString name;
    QString typeName;
    bool nullable = true;
    bool isPrimaryKey = false;
    bool isAutoIncrement = false;
    QVariant defaultValue;

    QJsonObject toJson() const;
    static ColumnInfo fromJson(const QJsonObject& json);
};

struct TableInfo {
    QString name;
    QString schema;
    QString type;  // "table", "view", "materialized_view"
    qint64 rowCount = 0;
    QString comment;

    QJsonObject toJson() const;
    static TableInfo fromJson(const QJsonObject& json);
};

struct IndexInfo {
    QString name;
    QStringList columns;
    bool isUnique = false;
    bool isPrimary = false;

    QJsonObject toJson() const;
    static IndexInfo fromJson(const QJsonObject& json);
};

struct ForeignKeyInfo {
    QString name;
    QString columnName;
    QString referencedSchema;
    QString referencedTable;
    QString referencedColumnName;

    QJsonObject toJson() const;
    static ForeignKeyInfo fromJson(const QJsonObject& json);
};

struct QueryResult {
    QStringList columnNames;
    QList<ColumnInfo> columns;
    QList<QVariantMap> rows;
    qint64 rowsAffected = 0;
    qint64 executionTimeMs = 0;
    QString error;
    bool success = true;

    bool isEmpty() const { return rows.isEmpty(); }
    int rowCount() const { return rows.size(); }
    int columnCount() const { return columnNames.size(); }

    QJsonObject toJson() const;
    static QueryResult fromJson(const QJsonObject& json);
};

struct SchemaInfo {
    QString databaseName;
    QString currentSchema;
    QStringList schemas;
    QList<TableInfo> tables;
    QList<TableInfo> views;

    QJsonObject toJson() const;
    static SchemaInfo fromJson(const QJsonObject& json);
};

struct ConnectionConfig {
    QString id;
    QString name;
    QString type;  // "postgresql", "mysql", "sqlite", etc.
    QString host;
    int port = 0;
    QString database;
    QString username;
    QString schema;

    // SSL options
    bool sslEnabled = false;
    QString sslCertPath;
    QString sslKeyPath;
    QString sslCaPath;

    // SSH tunnel options
    bool sshEnabled = false;
    QString sshHost;
    int sshPort = 22;
    QString sshUsername;
    QString sshKeyPath;

    // Other options
    int timeout = 30;
    QString startupQuery;

    QJsonObject toJson() const;
    static ConnectionConfig fromJson(const QJsonObject& json);
};

struct ConnectionInfo {
    QString id;
    QString name;
    QString type;
    QString database;
    QString schema;
    QString host;
    int port = 0;
    bool connected = false;
    QString serverVersion;

    QJsonObject toJson() const;
};

} // namespace tablepro
```

**Step 2: Create types.cpp**

```cpp
#include "types.hpp"

namespace tablepro {

// ColumnInfo
QJsonObject ColumnInfo::toJson() const {
    return {
        {"name", name},
        {"typeName", typeName},
        {"nullable", nullable},
        {"isPrimaryKey", isPrimaryKey},
        {"isAutoIncrement", isAutoIncrement},
        {"defaultValue", defaultValue.toString()}
    };
}

ColumnInfo ColumnInfo::fromJson(const QJsonObject& json) {
    ColumnInfo info;
    info.name = json["name"].toString();
    info.typeName = json["typeName"].toString();
    info.nullable = json["nullable"].toBool(true);
    info.isPrimaryKey = json["isPrimaryKey"].toBool();
    info.isAutoIncrement = json["isAutoIncrement"].toBool();
    info.defaultValue = json["defaultValue"].toVariant();
    return info;
}

// TableInfo
QJsonObject TableInfo::toJson() const {
    return {
        {"name", name},
        {"schema", schema},
        {"type", type},
        {"rowCount", rowCount},
        {"comment", comment}
    };
}

TableInfo TableInfo::fromJson(const QJsonObject& json) {
    TableInfo info;
    info.name = json["name"].toString();
    info.schema = json["schema"].toString();
    info.type = json["type"].toString();
    info.rowCount = json["rowCount"].toVariant().toLongLong();
    info.comment = json["comment"].toString();
    return info;
}

// IndexInfo
QJsonObject IndexInfo::toJson() const {
    QJsonArray cols;
    for (const auto& c : columns) cols.append(c);
    return {
        {"name", name},
        {"columns", cols},
        {"isUnique", isUnique},
        {"isPrimary", isPrimary}
    };
}

IndexInfo IndexInfo::fromJson(const QJsonObject& json) {
    IndexInfo info;
    info.name = json["name"].toString();
    info.isUnique = json["isUnique"].toBool();
    info.isPrimary = json["isPrimary"].toBool();
    for (const auto& c : json["columns"].toArray()) {
        info.columns.append(c.toString());
    }
    return info;
}

// ForeignKeyInfo
QJsonObject ForeignKeyInfo::toJson() const {
    return {
        {"name", name},
        {"columnName", columnName},
        {"referencedSchema", referencedSchema},
        {"referencedTable", referencedTable},
        {"referencedColumnName", referencedColumnName}
    };
}

ForeignKeyInfo ForeignKeyInfo::fromJson(const QJsonObject& json) {
    ForeignKeyInfo info;
    info.name = json["name"].toString();
    info.columnName = json["columnName"].toString();
    info.referencedSchema = json["referencedSchema"].toString();
    info.referencedTable = json["referencedTable"].toString();
    info.referencedColumnName = json["referencedColumnName"].toString();
    return info;
}

// QueryResult
QJsonObject QueryResult::toJson() const {
    QJsonArray rowsArray;
    for (const auto& row : rows) {
        QJsonObject rowObj;
        for (auto it = row.begin(); it != row.end(); ++it) {
            rowObj[it.key()] = QJsonValue::fromVariant(it.value());
        }
        rowsArray.append(rowObj);
    }

    QJsonArray colsArray;
    for (const auto& col : columns) {
        colsArray.append(col.toJson());
    }

    return {
        {"columnNames", QJsonArray::fromStringList(columnNames)},
        {"columns", colsArray},
        {"rows", rowsArray},
        {"rowsAffected", rowsAffected},
        {"executionTimeMs", executionTimeMs},
        {"error", error},
        {"success", success}
    };
}

QueryResult QueryResult::fromJson(const QJsonObject& json) {
    QueryResult result;
    for (const auto& name : json["columnNames"].toArray()) {
        result.columnNames.append(name.toString());
    }
    for (const auto& col : json["columns"].toArray()) {
        result.columns.append(ColumnInfo::fromJson(col.toObject()));
    }
    for (const auto& row : json["rows"].toArray()) {
        QVariantMap rowMap;
        for (auto it = row.toObject().begin(); it != row.toObject().end(); ++it) {
            rowMap[it.key()] = it.value().toVariant();
        }
        result.rows.append(rowMap);
    }
    result.rowsAffected = json["rowsAffected"].toInteger();
    result.executionTimeMs = json["executionTimeMs"].toInteger();
    result.error = json["error"].toString();
    result.success = json["success"].toBool(true);
    return result;
}

// SchemaInfo
QJsonObject SchemaInfo::toJson() const {
    QJsonArray tablesArray;
    for (const auto& t : tables) tablesArray.append(t.toJson());
    QJsonArray viewsArray;
    for (const auto& v : views) viewsArray.append(v.toJson());

    return {
        {"databaseName", databaseName},
        {"currentSchema", currentSchema},
        {"schemas", QJsonArray::fromStringList(schemas)},
        {"tables", tablesArray},
        {"views", viewsArray}
    };
}

SchemaInfo SchemaInfo::fromJson(const QJsonObject& json) {
    SchemaInfo info;
    info.databaseName = json["databaseName"].toString();
    info.currentSchema = json["currentSchema"].toString();
    for (const auto& s : json["schemas"].toArray()) {
        info.schemas.append(s.toString());
    }
    for (const auto& t : json["tables"].toArray()) {
        info.tables.append(TableInfo::fromJson(t.toObject()));
    }
    for (const auto& v : json["views"].toArray()) {
        info.views.append(TableInfo::fromJson(v.toObject()));
    }
    return info;
}

// ConnectionConfig
QJsonObject ConnectionConfig::toJson() const {
    return {
        {"id", id},
        {"name", name},
        {"type", type},
        {"host", host},
        {"port", port},
        {"database", database},
        {"username", username},
        {"schema", schema},
        {"sslEnabled", sslEnabled},
        {"sslCertPath", sslCertPath},
        {"sslKeyPath", sslKeyPath},
        {"sslCaPath", sslCaPath},
        {"sshEnabled", sshEnabled},
        {"sshHost", sshHost},
        {"sshPort", sshPort},
        {"sshUsername", sshUsername},
        {"sshKeyPath", sshKeyPath},
        {"timeout", timeout},
        {"startupQuery", startupQuery}
    };
}

ConnectionConfig ConnectionConfig::fromJson(const QJsonObject& json) {
    ConnectionConfig config;
    config.id = json["id"].toString();
    config.name = json["name"].toString();
    config.type = json["type"].toString();
    config.host = json["host"].toString();
    config.port = json["port"].toInt();
    config.database = json["database"].toString();
    config.username = json["username"].toString();
    config.schema = json["schema"].toString();
    config.sslEnabled = json["sslEnabled"].toBool();
    config.sslCertPath = json["sslCertPath"].toString();
    config.sslKeyPath = json["sslKeyPath"].toString();
    config.sslCaPath = json["sslCaPath"].toString();
    config.sshEnabled = json["sshEnabled"].toBool();
    config.sshHost = json["sshHost"].toString();
    config.sshPort = json["sshPort"].toInt(22);
    config.sshUsername = json["sshUsername"].toString();
    config.sshKeyPath = json["sshKeyPath"].toString();
    config.timeout = json["timeout"].toInt(30);
    config.startupQuery = json["startupQuery"].toString();
    return config;
}

// ConnectionInfo
QJsonObject ConnectionInfo::toJson() const {
    return {
        {"id", id},
        {"name", name},
        {"type", type},
        {"database", database},
        {"schema", schema},
        {"host", host},
        {"port", port},
        {"connected", connected},
        {"serverVersion", serverVersion}
    };
}

} // namespace tablepro
```

**Step 3: Commit types**

```bash
git add src/core/types.hpp src/core/types.cpp
git commit -m "feat: Add core data types for queries, schemas, connections"
```

---

## Task 3: Database Driver Interface

**Files:**
- Create: `src/core/driver.hpp`

**Step 1: Create driver.hpp**

```cpp
#pragma once

#include <QObject>
#include <QFuture>
#include <QMutex>
#include "types.hpp"

namespace tablepro {

class DatabaseDriver : public QObject {
    Q_OBJECT

public:
    explicit DatabaseDriver(QObject* parent = nullptr);
    ~DatabaseDriver() override;

    // Connection management
    virtual QFuture<bool> connect(const ConnectionConfig& config) = 0;
    virtual void disconnect() = 0;
    virtual bool isConnected() const = 0;
    virtual QFuture<void> ping() = 0;

    // Query execution
    virtual QFuture<QueryResult> execute(const QString& sql) = 0;
    virtual QFuture<QueryResult> executeParams(const QString& sql, const QVariantList& params) = 0;

    // Schema inspection
    virtual QFuture<SchemaInfo> fetchSchema() = 0;
    virtual QFuture<QList<TableInfo>> fetchTables(const QString& schema = QString()) = 0;
    virtual QFuture<QList<ColumnInfo>> fetchColumns(const QString& schema, const QString& table) = 0;
    virtual QFuture<QList<IndexInfo>> fetchIndexes(const QString& schema, const QString& table) = 0;
    virtual QFuture<QList<ForeignKeyInfo>> fetchForeignKeys(const QString& schema, const QString& table) = 0;
    virtual QFuture<QString> fetchDDL(const QString& schema, const QString& table) = 0;
    virtual QFuture<qint64> fetchRowCount(const QString& schema, const QString& table) = 0;

    // Transactions
    virtual bool beginTransaction() = 0;
    virtual bool commitTransaction() = 0;
    virtual bool rollbackTransaction() = 0;

    // Dialect info
    virtual QString identifierQuote() const = 0;
    virtual QString literalQuote() const = 0;
    virtual QString autoIncrementSyntax() const = 0;
    virtual QString limitClause(int limit, int offset) const = 0;

    // Connection info
    virtual ConnectionInfo connectionInfo() const = 0;
    virtual QString serverVersion() const = 0;

signals:
    void connected();
    void disconnected();
    void connectionError(const QString& message, ErrorCode code);
    void queryExecuted(const QString& sql, qint64 durationMs);

protected:
    ConnectionConfig m_config;
    mutable QMutex m_mutex;
    bool m_connected = false;
};

} // namespace tablepro
```

**Step 2: Create driver.cpp (base implementation)**

```cpp
#include "driver.hpp"

namespace tablepro {

DatabaseDriver::DatabaseDriver(QObject* parent)
    : QObject(parent)
{
}

DatabaseDriver::~DatabaseDriver() = default;

} // namespace tablepro
```

**Step 3: Commit driver interface**

```bash
git add src/core/driver.hpp src/core/driver.cpp
git commit -m "feat: Add abstract DatabaseDriver interface"
```

---

## Task 4: Driver Factory

**Files:**
- Create: `src/core/driver_factory.hpp`
- Create: `src/core/driver_factory.cpp`

**Step 1: Create driver_factory.hpp**

```cpp
#pragma once

#include <QObject>
#include <memory>
#include <functional>
#include <QMap>
#include "driver.hpp"

namespace tablepro {

enum class DatabaseType {
    PostgreSQL,
    MySQL,
    SQLite,
    DuckDB,
    SQLServer,
    ClickHouse,
    MongoDB,
    Redis
};

class DriverFactory {
public:
    using DriverCreator = std::function<std::unique_ptr<DatabaseDriver>(QObject*)>;

    static std::unique_ptr<DatabaseDriver> create(DatabaseType type, QObject* parent = nullptr);
    static std::unique_ptr<DatabaseDriver> create(const QString& typeName, QObject* parent = nullptr);

    static QList<DatabaseType> supportedTypes();
    static QStringList supportedTypeNames();
    static QString typeName(DatabaseType type);
    static DatabaseType typeFromName(const QString& name);
    static bool isSupported(DatabaseType type);
    static bool isSupported(const QString& typeName);

    static void registerDriver(DatabaseType type, DriverCreator creator);

private:
    static QMap<DatabaseType, DriverCreator>& registry();
};

} // namespace tablepro
```

**Step 2: Create driver_factory.cpp**

```cpp
#include "driver_factory.hpp"
#include <stdexcept>

namespace tablepro {

QMap<DatabaseType, DriverFactory::DriverCreator>& DriverFactory::registry() {
    static QMap<DatabaseType, DriverCreator> reg;
    return reg;
}

void DriverFactory::registerDriver(DatabaseType type, DriverCreator creator) {
    registry().insert(type, std::move(creator));
}

std::unique_ptr<DatabaseDriver> DriverFactory::create(DatabaseType type, QObject* parent) {
    const auto& reg = registry();
    if (!reg.contains(type)) {
        throw std::invalid_argument(
            QString("Unsupported database type: %1").arg(typeName(type)).toStdString()
        );
    }
    return reg[type](parent);
}

std::unique_ptr<DatabaseDriver> DriverFactory::create(const QString& typeName, QObject* parent) {
    return create(typeFromName(typeName), parent);
}

QList<DatabaseType> DriverFactory::supportedTypes() {
    return registry().keys();
}

QStringList DriverFactory::supportedTypeNames() {
    QStringList names;
    for (auto type : supportedTypes()) {
        names.append(typeName(type));
    }
    return names;
}

QString DriverFactory::typeName(DatabaseType type) {
    switch (type) {
        case DatabaseType::PostgreSQL: return "postgresql";
        case DatabaseType::MySQL: return "mysql";
        case DatabaseType::SQLite: return "sqlite";
        case DatabaseType::DuckDB: return "duckdb";
        case DatabaseType::SQLServer: return "sqlserver";
        case DatabaseType::ClickHouse: return "clickhouse";
        case DatabaseType::MongoDB: return "mongodb";
        case DatabaseType::Redis: return "redis";
        default: return "unknown";
    }
}

DatabaseType DriverFactory::typeFromName(const QString& name) {
    QString lower = name.toLower();
    if (lower == "postgresql" || lower == "postgres" || lower == "psql") {
        return DatabaseType::PostgreSQL;
    }
    if (lower == "mysql" || lower == "mariadb") {
        return DatabaseType::MySQL;
    }
    if (lower == "sqlite") {
        return DatabaseType::SQLite;
    }
    if (lower == "duckdb" || lower == "duck") {
        return DatabaseType::DuckDB;
    }
    if (lower == "sqlserver" || lower == "mssql") {
        return DatabaseType::SQLServer;
    }
    if (lower == "clickhouse") {
        return DatabaseType::ClickHouse;
    }
    if (lower == "mongodb" || lower == "mongo") {
        return DatabaseType::MongoDB;
    }
    if (lower == "redis") {
        return DatabaseType::Redis;
    }
    throw std::invalid_argument(QString("Unknown database type: %1").arg(name).toStdString());
}

bool DriverFactory::isSupported(DatabaseType type) {
    return registry().contains(type);
}

bool DriverFactory::isSupported(const QString& typeName) {
    try {
        return isSupported(typeFromName(typeName));
    } catch (...) {
        return false;
    }
}

} // namespace tablepro
```

**Step 3: Commit driver factory**

```bash
git add src/core/driver_factory.hpp src/core/driver_factory.cpp
git commit -m "feat: Add DriverFactory for driver instantiation"
```

---

## Task 5: Connection Manager

**Files:**
- Create: `src/core/connection_manager.hpp`
- Create: `src/core/connection_manager.cpp`

**Step 1: Create connection_manager.hpp**

```cpp
#pragma once

#include <QObject>
#include <QMap>
#include <QUuid>
#include <memory>
#include "types.hpp"
#include "driver.hpp"

namespace tablepro {

class ConnectionManager : public QObject {
    Q_OBJECT

public:
    static ConnectionManager* instance();

    // Connection lifecycle
    QFuture<ConnectionInfo> connect(const ConnectionConfig& config);
    void disconnect(const QString& connectionId);
    void disconnectAll();

    // Connection queries
    ConnectionInfo connectionInfo(const QString& connectionId) const;
    QList<ConnectionInfo> activeConnections() const;
    bool isConnected(const QString& connectionId) const;
    DatabaseDriver* driver(const QString& connectionId) const;

    // Configuration persistence
    void saveConnection(const ConnectionConfig& config);
    void deleteConnection(const QString& id);
    QList<ConnectionConfig> savedConnections() const;
    ConnectionConfig savedConnection(const QString& id) const;

    // File operations
    void loadConnections();
    void saveConnections();

signals:
    void connectionEstablished(const ConnectionInfo& info);
    void connectionClosed(const QString& connectionId);
    void connectionError(const QString& connectionId, const QString& error);
    void connectionsLoaded();
    void connectionsSaved();

private:
    explicit ConnectionManager(QObject* parent = nullptr);

    QString configFilePath() const;
    QString generateId() const;

    QMap<QString, std::unique_ptr<DatabaseDriver>> m_connections;
    QMap<QString, ConnectionConfig> m_savedConfigs;
    mutable QMutex m_mutex;
};

} // namespace tablepro
```

**Step 2: Create connection_manager.cpp**

```cpp
#include "connection_manager.hpp"
#include "driver_factory.hpp"
#include "errors.hpp"
#include <QFile>
#include <QJsonDocument>
#include <QJsonArray>
#include <QStandardPaths>
#include <QDir>
#include <QtConcurrent>

namespace tablepro {

ConnectionManager* ConnectionManager::instance() {
    static ConnectionManager* inst = new ConnectionManager();
    return inst;
}

ConnectionManager::ConnectionManager(QObject* parent)
    : QObject(parent)
{
    loadConnections();
}

QString ConnectionManager::configFilePath() const {
    QString configPath = QStandardPaths::writableLocation(QStandardPaths::AppConfigLocation);
    QDir dir(configPath);
    if (!dir.exists()) {
        dir.mkpath(".");
    }
    return configPath + "/connections.json";
}

QString ConnectionManager::generateId() const {
    return QUuid::createUuid().toString(QUuid::WithoutBraces);
}

QFuture<ConnectionInfo> ConnectionManager::connect(const ConnectionConfig& config) {
    return QtConcurrent::run([this, config]() -> ConnectionInfo {
        QMutexLocker locker(&m_mutex);

        QString connId = config.id.isEmpty() ? generateId() : config.id;

        try {
            auto driverType = DriverFactory::typeFromName(config.type);
            auto driver = DriverFactory::create(driverType, this);

            // Connect synchronously for now
            auto future = driver->connect(config);
            future.waitForFinished();

            if (!future.result()) {
                throw ConnectionException(
                    QString("Failed to connect to %1 at %2:%3")
                        .arg(config.name, config.host)
                        .arg(config.port)
                );
            }

            m_connections.insert(connId, std::move(driver));

            ConnectionInfo info;
            info.id = connId;
            info.name = config.name;
            info.type = config.type;
            info.database = config.database;
            info.schema = config.schema;
            info.host = config.host;
            info.port = config.port;
            info.connected = true;

            emit connectionEstablished(info);
            return info;

        } catch (const TableProException& e) {
            emit connectionError(config.id, e.message());
            throw;
        } catch (const std::exception& e) {
            emit connectionError(config.id, QString::fromUtf8(e.what()));
            throw ConnectionException(QString::fromUtf8(e.what()));
        }
    });
}

void ConnectionManager::disconnect(const QString& connectionId) {
    QMutexLocker locker(&m_mutex);

    if (m_connections.contains(connectionId)) {
        m_connections[connectionId]->disconnect();
        m_connections.remove(connectionId);
        emit connectionClosed(connectionId);
    }
}

void ConnectionManager::disconnectAll() {
    QMutexLocker locker(&m_mutex);

    for (auto it = m_connections.begin(); it != m_connections.end(); ++it) {
        it.value()->disconnect();
        emit connectionClosed(it.key());
    }
    m_connections.clear();
}

ConnectionInfo ConnectionManager::connectionInfo(const QString& connectionId) const {
    QMutexLocker locker(&m_mutex);

    if (m_connections.contains(connectionId)) {
        return m_connections[connectionId]->connectionInfo();
    }
    return ConnectionInfo();
}

QList<ConnectionInfo> ConnectionManager::activeConnections() const {
    QMutexLocker locker(&m_mutex);

    QList<ConnectionInfo> infos;
    for (const auto& driver : m_connections) {
        infos.append(driver->connectionInfo());
    }
    return infos;
}

bool ConnectionManager::isConnected(const QString& connectionId) const {
    QMutexLocker locker(&m_mutex);
    return m_connections.contains(connectionId) &&
           m_connections[connectionId]->isConnected();
}

DatabaseDriver* ConnectionManager::driver(const QString& connectionId) const {
    QMutexLocker locker(&m_mutex);
    return m_connections.value(connectionId).get();
}

void ConnectionManager::saveConnection(const ConnectionConfig& config) {
    QString id = config.id.isEmpty() ? generateId() : config.id;
    ConnectionConfig configWithId = config;
    configWithId.id = id;

    m_savedConfigs.insert(id, configWithId);
    saveConnections();
}

void ConnectionManager::deleteConnection(const QString& id) {
    m_savedConfigs.remove(id);
    saveConnections();
}

QList<ConnectionConfig> ConnectionManager::savedConnections() const {
    return m_savedConfigs.values();
}

ConnectionConfig ConnectionManager::savedConnection(const QString& id) const {
    return m_savedConfigs.value(id);
}

void ConnectionManager::loadConnections() {
    QFile file(configFilePath());
    if (!file.open(QIODevice::ReadOnly)) {
        return;
    }

    QJsonParseError error;
    QJsonDocument doc = QJsonDocument::fromJson(file.readAll(), &error);
    file.close();

    if (error.error != QJsonParseError::NoError) {
        qWarning() << "Failed to parse connections file:" << error.errorString();
        return;
    }

    QJsonArray array = doc.array();
    for (const auto& value : array) {
        ConnectionConfig config = ConnectionConfig::fromJson(value.toObject());
        m_savedConfigs.insert(config.id, config);
    }

    emit connectionsLoaded();
}

void ConnectionManager::saveConnections() {
    QJsonArray array;
    for (const auto& config : m_savedConfigs) {
        array.append(config.toJson());
    }

    QFile file(configFilePath());
    if (!file.open(QIODevice::WriteOnly)) {
        qWarning() << "Failed to open connections file for writing";
        return;
    }

    file.write(QJsonDocument(array).toJson());
    file.close();

    emit connectionsSaved();
}

} // namespace tablepro
```

**Step 3: Commit connection manager**

```bash
git add src/core/connection_manager.hpp src/core/connection_manager.cpp
git commit -m "feat: Add ConnectionManager for connection lifecycle"
```

---

## Task 6: Query Executor

**Files:**
- Create: `src/core/query_executor.hpp`
- Create: `src/core/query_executor.cpp`

**Step 1: Create query_executor.hpp**

```cpp
#pragma once

#include <QObject>
#include <QFuture>
#include <QUuid>
#include "types.hpp"
#include "driver.hpp"

namespace tablepro {

class QueryExecutor : public QObject {
    Q_OBJECT

public:
    explicit QueryExecutor(QObject* parent = nullptr);

    // Execute query on a connection
    QFuture<QueryResult> execute(const QString& connectionId, const QString& sql);
    QFuture<QueryResult> executeParams(
        const QString& connectionId,
        const QString& sql,
        const QVariantList& params
    );

    // Execute with pagination
    QFuture<QueryResult> executePaginated(
        const QString& connectionId,
        const QString& sql,
        int limit,
        int offset
    );

    // Cancel running query
    void cancel(const QString& queryId);

    // Query history
    struct HistoryItem {
        QString id;
        QString connectionId;
        QString database;
        QString sql;
        qint64 executionTimeMs;
        bool success;
        QString error;
        QDateTime timestamp;
    };

    QList<HistoryItem> recentHistory(int limit = 100) const;
    void clearHistory();

signals:
    void queryStarted(const QString& queryId, const QString& sql);
    void queryFinished(const QString& queryId, const QueryResult& result);
    void queryFailed(const QString& queryId, const QString& error);
    void queryCancelled(const QString& queryId);

private:
    QString generateQueryId() const;
    void addToHistory(const QString& connectionId, const QString& sql,
                      const QueryResult& result, qint64 durationMs);

    QList<HistoryItem> m_history;
    mutable QMutex m_mutex;
};

} // namespace tablepro
```

**Step 2: Create query_executor.cpp**

```cpp
#include "query_executor.hpp"
#include "connection_manager.hpp"
#include <QtConcurrent>
#include <QDateTime>

namespace tablepro {

QueryExecutor::QueryExecutor(QObject* parent)
    : QObject(parent)
{
}

QString QueryExecutor::generateQueryId() const {
    return QUuid::createUuid().toString(QUuid::WithoutBraces);
}

QFuture<QueryResult> QueryExecutor::execute(const QString& connectionId, const QString& sql) {
    return executeParams(connectionId, sql, {});
}

QFuture<QueryResult> QueryExecutor::executeParams(
    const QString& connectionId,
    const QString& sql,
    const QVariantList& params
) {
    QString queryId = generateQueryId();

    return QtConcurrent::run([this, connectionId, sql, params, queryId]() -> QueryResult {
        emit queryStarted(queryId, sql);

        auto startTime = std::chrono::steady_clock::now();

        QueryResult result;

        try {
            auto* driver = ConnectionManager::instance()->driver(connectionId);
            if (!driver) {
                throw QueryException("Connection not found: " + connectionId);
            }

            QFuture<QueryResult> future;
            if (params.isEmpty()) {
                future = driver->execute(sql);
            } else {
                future = driver->executeParams(sql, params);
            }
            future.waitForFinished();
            result = future.result();

        } catch (const TableProException& e) {
            result.success = false;
            result.error = e.message();
            emit queryFailed(queryId, e.message());
        } catch (const std::exception& e) {
            result.success = false;
            result.error = QString::fromUtf8(e.what());
            emit queryFailed(queryId, result.error);
        }

        auto endTime = std::chrono::steady_clock::now();
        result.executionTimeMs = std::chrono::duration_cast<std::chrono::milliseconds>(
            endTime - startTime
        ).count();

        addToHistory(connectionId, sql, result, result.executionTimeMs);
        emit queryFinished(queryId, result);

        return result;
    });
}

QFuture<QueryResult> QueryExecutor::executePaginated(
    const QString& connectionId,
    const QString& sql,
    int limit,
    int offset
) {
    auto* driver = ConnectionManager::instance()->driver(connectionId);
    if (!driver) {
        return QtConcurrent::run([]() -> QueryResult {
            QueryResult result;
            result.success = false;
            result.error = "Connection not found";
            return result;
        });
    }

    QString paginatedSql = sql;
    if (!sql.contains("LIMIT", Qt::CaseInsensitive)) {
        paginatedSql = driver->limitClause(limit, offset).arg(sql);
    }

    return execute(connectionId, paginatedSql);
}

void QueryExecutor::cancel(const QString& queryId) {
    // TODO: Implement query cancellation
    emit queryCancelled(queryId);
}

void QueryExecutor::addToHistory(
    const QString& connectionId,
    const QString& sql,
    const QueryResult& result,
    qint64 durationMs
) {
    QMutexLocker locker(&m_mutex);

    HistoryItem item;
    item.id = generateQueryId();
    item.connectionId = connectionId;
    item.sql = sql;
    item.executionTimeMs = durationMs;
    item.success = result.success;
    item.error = result.error;
    item.timestamp = QDateTime::currentDateTime();

    m_history.prepend(item);

    // Keep only last 1000 items
    while (m_history.size() > 1000) {
        m_history.removeLast();
    }
}

QList<QueryExecutor::HistoryItem> QueryExecutor::recentHistory(int limit) const {
    QMutexLocker locker(&m_mutex);
    return m_history.mid(0, qMin(limit, m_history.size()));
}

void QueryExecutor::clearHistory() {
    QMutexLocker locker(&m_mutex);
    m_history.clear();
}

} // namespace tablepro
```

**Step 3: Commit query executor**

```bash
git add src/core/query_executor.hpp src/core/query_executor.cpp
git commit -m "feat: Add QueryExecutor for async query execution"
```

---

## Task 7: Logging System

**Files:**
- Create: `src/core/logging.hpp`
- Create: `src/core/logging.cpp`

**Step 1: Create logging.hpp**

```cpp
#pragma once

#include <QLoggingCategory>
#include <QString>

namespace tablepro {

// Declare logging categories
Q_DECLARE_LOGGING_CATEGORY(lcCore)
Q_DECLARE_LOGGING_CATEGORY(lcConnection)
Q_DECLARE_LOGGING_CATEGORY(lcQuery)
Q_DECLARE_LOGGING_CATEGORY(lcUI)
Q_DECLARE_LOGGING_CATEGORY(lcDriver)

// Initialize logging system
void initLogging(const QString& logLevel = "info");

// Set log level
void setLogLevel(const QString& level);

// Log to file
void enableFileLogging(const QString& filePath = QString());

} // namespace tablepro
```

**Step 2: Create logging.cpp**

```cpp
#include "logging.hpp"
#include <QFile>
#include <QTextStream>
#include <QDateTime>
#include <QDir>
#include <QStandardPaths>
#include <cstdio>

QT_BEGIN_NAMESPACE
Q_LOGGING_CATEGORY(lcCore, "tablepro.core")
Q_LOGGING_CATEGORY(lcConnection, "tablepro.connection")
Q_LOGGING_CATEGORY(lcQuery, "tablepro.query")
Q_LOGGING_CATEGORY(lcUI, "tablepro.ui")
Q_LOGGING_CATEGORY(lcDriver, "tablepro.driver")
QT_END_NAMESPACE

namespace tablepro {

static QFile* logFile = nullptr;

static void messageHandler(QtMsgType type, const QMessageLogContext& context, const QString& msg) {
    QString levelStr;
    switch (type) {
        case QtDebugMsg:    levelStr = "DEBUG"; break;
        case QtInfoMsg:     levelStr = "INFO"; break;
        case QtWarningMsg:  levelStr = "WARN"; break;
        case QtCriticalMsg: levelStr = "ERROR"; break;
        case QtFatalMsg:    levelStr = "FATAL"; break;
    }

    QString timestamp = QDateTime::currentDateTime().toString("yyyy-MM-dd HH:mm:ss.zzz");
    QString category = context.category ? context.category : "default";
    QString file = context.file ? QFileInfo(context.file).fileName() : "";
    int line = context.line;

    QString formatted = QString("[%1] [%2] [%3] %4")
        .arg(timestamp, levelStr, category, msg);

    // Output to stderr
    fprintf(stderr, "%s\n", formatted.toUtf8().constData());

    // Output to file if enabled
    if (logFile && logFile->isOpen()) {
        QTextStream stream(logFile);
        stream << formatted << "\n";
        stream.flush();
    }
}

void initLogging(const QString& logLevel) {
    setLogLevel(logLevel);
    qInstallMessageHandler(messageHandler);
}

void setLogLevel(const QString& level) {
    QString filterRules;

    if (level == "debug") {
        filterRules = "tablepro.*=true";
    } else if (level == "info") {
        filterRules = "tablepro.*=info";
    } else if (level == "warning") {
        filterRules = "tablepro.*=warning";
    } else if (level == "error") {
        filterRules = "tablepro.*=critical";
    } else {
        filterRules = "tablepro.core=info\ntablepro.connection=info\ntablepro.query=info\ntablepro.ui=warning\ntablepro.driver=info";
    }

    QLoggingCategory::setFilterRules(filterRules);
}

void enableFileLogging(const QString& filePath) {
    QString path = filePath;
    if (path.isEmpty()) {
        QString logDir = QStandardPaths::writableLocation(QStandardPaths::AppDataLocation);
        QDir dir(logDir);
        if (!dir.exists()) {
            dir.mkpath(".");
        }
        path = logDir + "/tablepro.log";
    }

    logFile = new QFile(path);
    if (logFile->open(QIODevice::WriteOnly | QIODevice::Append)) {
        qCInfo(lcCore) << "Logging to file:" << path;
    } else {
        qCWarning(lcCore) << "Failed to open log file:" << path;
        delete logFile;
        logFile = nullptr;
    }
}

} // namespace tablepro
```

**Step 3: Commit logging**

```bash
git add src/core/logging.hpp src/core/logging.cpp
git commit -m "feat: Add categorized logging system"
```

---

## Task 8: Update CMakeLists.txt

**Files:**
- Modify: `CMakeLists.txt`

**Step 1: Add new source files to CMakeLists.txt**

Add after the existing `set(TABLEPRO_SOURCES ...)` block:

```cmake
set(TABLEPRO_SOURCES
    src/main.cpp
    src/ui/MainWindow.cpp
    src/core/errors.cpp
    src/core/types.cpp
    src/core/driver.cpp
    src/core/driver_factory.cpp
    src/core/connection_manager.cpp
    src/core/query_executor.cpp
    src/core/logging.cpp
)
```

**Step 2: Commit CMakeLists update**

```bash
git add CMakeLists.txt
git commit -m "build: Add core sources to CMakeLists.txt"
```

---

## Task 9: Unit Tests Setup

**Files:**
- Create: `tests/unit/test_types.cpp`
- Create: `tests/unit/test_connection_manager.cpp`

**Step 1: Create test_types.cpp**

```cpp
#include <gtest/gtest.h>
#include "core/types.hpp"

using namespace tablepro;

TEST(TypesTest, ColumnInfoToJsonAndBack) {
    ColumnInfo original;
    original.name = "id";
    original.typeName = "integer";
    original.nullable = false;
    original.isPrimaryKey = true;
    original.isAutoIncrement = true;
    original.defaultValue = 0;

    QJsonObject json = original.toJson();
    ColumnInfo restored = ColumnInfo::fromJson(json);

    EXPECT_EQ(restored.name, original.name);
    EXPECT_EQ(restored.typeName, original.typeName);
    EXPECT_EQ(restored.nullable, original.nullable);
    EXPECT_EQ(restored.isPrimaryKey, original.isPrimaryKey);
    EXPECT_EQ(restored.isAutoIncrement, original.isAutoIncrement);
}

TEST(TypesTest, QueryResultToJsonAndBack) {
    QueryResult original;
    original.columnNames = {"id", "name"};
    original.rowsAffected = 5;
    original.executionTimeMs = 123;
    original.success = true;

    QVariantMap row;
    row["id"] = 1;
    row["name"] = "test";
    original.rows.append(row);

    QJsonObject json = original.toJson();
    QueryResult restored = QueryResult::fromJson(json);

    EXPECT_EQ(restored.columnNames, original.columnNames);
    EXPECT_EQ(restored.rowsAffected, original.rowsAffected);
    EXPECT_EQ(restored.executionTimeMs, original.executionTimeMs);
    EXPECT_EQ(restored.success, original.success);
    EXPECT_EQ(restored.rows.size(), 1);
}

TEST(TypesTest, ConnectionConfigToJsonAndBack) {
    ConnectionConfig original;
    original.id = "test-conn";
    original.name = "Test Connection";
    original.type = "postgresql";
    original.host = "localhost";
    original.port = 5432;
    original.database = "testdb";
    original.username = "testuser";

    QJsonObject json = original.toJson();
    ConnectionConfig restored = ConnectionConfig::fromJson(json);

    EXPECT_EQ(restored.id, original.id);
    EXPECT_EQ(restored.name, original.name);
    EXPECT_EQ(restored.type, original.type);
    EXPECT_EQ(restored.host, original.host);
    EXPECT_EQ(restored.port, original.port);
    EXPECT_EQ(restored.database, original.database);
    EXPECT_EQ(restored.username, original.username);
}
```

**Step 2: Create CMakeLists.txt for tests**

```cmake
# tests/unit/CMakeLists.txt
find_package(GTest REQUIRED)

add_executable(tablepro_unit_tests
    test_types.cpp
)

target_link_libraries(tablepro_unit_tests PRIVATE
    tablepro
    GTest::gtest
    GTest::gtest_main
)

include(GoogleTest)
gtest_discover_tests(tablepro_unit_tests)
```

**Step 3: Update root CMakeLists.txt**

Add at the end:

```cmake
# Testing
if(BUILD_TESTING)
    add_subdirectory(tests/unit)
endif()
```

**Step 4: Commit tests**

```bash
git add tests/
git commit -m "test: Add unit tests for core types"
```

---

## Task 10: Verify Build and Tests

**Step 1: Configure with tests enabled**

```bash
cmake --preset debug -DBUILD_TESTING=ON
```

**Step 2: Build**

```bash
cmake --build build/debug -j$(nproc)
```

**Step 3: Run tests**

```bash
cd build/debug && ctest --output-on-failure
```

Expected: Tests pass

**Step 4: Final commit if needed**

```bash
git status
```

---

## Acceptance Criteria

- [ ] All core types defined with JSON serialization
- [ ] DatabaseDriver abstract interface complete
- [ ] DriverFactory can create drivers by type
- [ ] ConnectionManager saves/loads connections
- [ ] QueryExecutor executes queries asynchronously
- [ ] Logging system with categorized output
- [ ] Unit tests pass for types
- [ ] Build succeeds with no warnings

---

**Phase 2 Complete.** Next: Phase 3 - PostgreSQL Driver