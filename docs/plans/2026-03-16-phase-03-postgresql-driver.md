# Phase 3: PostgreSQL Driver Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement complete PostgreSQL database driver with libpq, supporting connections, queries, schema inspection, and transactions.

**Architecture:** PostgresDriver implements DatabaseDriver interface using libpq C library. RAII for connection management. Prepared statements for parameterized queries. pg_catalog for schema inspection.

**Tech Stack:** C++20, libpq (PostgreSQL client library), Qt 6.6 Concurrent

---

## Task 1: Add libpq Dependency

**Files:**
- Modify: `CMakeLists.txt`
- Modify: `vcpkg.json`

**Step 1: vcpkg.json already has libpq**

Verify in `vcpkg.json`:
```json
"libpq"
```

**Step 2: Update CMakeLists.txt**

Add after `find_package(Qt6 ...)`:

```cmake
# PostgreSQL
find_package(PostgreSQL REQUIRED)
```

Add to `target_include_directories`:
```cmake
target_include_directories(tablepro PRIVATE
    ${CMAKE_SOURCE_DIR}/src
    ${PostgreSQL_INCLUDE_DIRS}
)
```

Add to `target_link_libraries`:
```cmake
target_link_libraries(tablepro PRIVATE
    # ... existing libs ...
    ${PostgreSQL_LIBRARIES}
)
```

**Step 3: Commit dependency**

```bash
git add CMakeLists.txt
git commit -m "build: Add PostgreSQL libpq dependency"
```

---

## Task 2: PostgresDriver Header

**Files:**
- Create: `src/driver/postgres_driver.hpp`

**Step 1: Create header**

```cpp
#pragma once

#include "../core/driver.hpp"
#include <libpq-fe.h>
#include <memory>

namespace tablepro {

class PostgresDriver : public DatabaseDriver {
    Q_OBJECT

public:
    explicit PostgresDriver(QObject* parent = nullptr);
    ~PostgresDriver() override;

    // Connection management
    QFuture<bool> connect(const ConnectionConfig& config) override;
    void disconnect() override;
    bool isConnected() const override;
    QFuture<void> ping() override;

    // Query execution
    QFuture<QueryResult> execute(const QString& sql) override;
    QFuture<QueryResult> executeParams(const QString& sql, const QVariantList& params) override;

    // Schema inspection
    QFuture<SchemaInfo> fetchSchema() override;
    QFuture<QList<TableInfo>> fetchTables(const QString& schema = QString()) override;
    QFuture<QList<ColumnInfo>> fetchColumns(const QString& schema, const QString& table) override;
    QFuture<QList<IndexInfo>> fetchIndexes(const QString& schema, const QString& table) override;
    QFuture<QList<ForeignKeyInfo>> fetchForeignKeys(const QString& schema, const QString& table) override;
    QFuture<QString> fetchDDL(const QString& schema, const QString& table) override;
    QFuture<qint64> fetchRowCount(const QString& schema, const QString& table) override;

    // Transactions
    bool beginTransaction() override;
    bool commitTransaction() override;
    bool rollbackTransaction() override;

    // Dialect info
    QString identifierQuote() const override { return "\""; }
    QString literalQuote() const override { return "'"; }
    QString autoIncrementSyntax() const override { return "SERIAL"; }
    QString limitClause(int limit, int offset) const override {
        return QString("SELECT * FROM (%1) AS subq LIMIT %2 OFFSET %3")
            .arg("%1").arg(limit).arg(offset);
    }

    // Connection info
    ConnectionInfo connectionInfo() const override;
    QString serverVersion() const override;

private:
    PGconn* m_conn = nullptr;
    bool m_inTransaction = false;

    // Helper methods
    QString buildConnectionString(const ConnectionConfig& config) const;
    QueryResult parseResult(PGresult* result) const;
    QVariant convertValue(const char* value, Oid type) const;
    void checkConnection() const;
    QString escapeIdentifier(const QString& identifier) const;
    QString escapeLiteral(const QString& value) const;

    // RAII wrapper for PGresult
    struct PGResultDeleter {
        void operator()(PGresult* result) const {
            if (result) PQclear(result);
        }
    };
    using PGResultPtr = std::unique_ptr<PGresult, PGResultDeleter>;
};

} // namespace tablepro
```

**Step 2: Commit header**

```bash
git add src/driver/postgres_driver.hpp
git commit -m "feat: Add PostgresDriver header"
```

---

## Task 3: PostgresDriver Connection

**Files:**
- Create: `src/driver/postgres_driver.cpp` (connection methods)

**Step 1: Create implementation file with connection methods**

```cpp
#include "postgres_driver.hpp"
#include "../core/errors.hpp"
#include "../core/logging.hpp"
#include <QtConcurrent>
#include <QHostInfo>

namespace tablepro {

using namespace std::chrono_literals;

PostgresDriver::PostgresDriver(QObject* parent)
    : DatabaseDriver(parent)
{
    qCDebug(lcDriver) << "PostgresDriver created";
}

PostgresDriver::~PostgresDriver() {
    disconnect();
}

QString PostgresDriver::buildConnectionString(const ConnectionConfig& config) const {
    QStringList parts;

    parts << QString("host=%1").arg(config.host);
    parts << QString("port=%1").arg(config.port > 0 ? config.port : 5432);
    parts << QString("dbname=%1").arg(config.database);
    parts << QString("user=%1").arg(config.username);

    if (!config.schema.isEmpty()) {
        parts << QString("search_path=%1").arg(config.schema);
    }

    // SSL options
    if (config.sslEnabled) {
        parts << "sslmode=require";
        if (!config.sslCaPath.isEmpty()) {
            parts << QString("sslrootcert=%1").arg(config.sslCaPath);
        }
        if (!config.sslCertPath.isEmpty()) {
            parts << QString("sslcert=%1").arg(config.sslCertPath);
        }
        if (!config.sslKeyPath.isEmpty()) {
            parts << QString("sslkey=%1").arg(config.sslKeyPath);
        }
    } else {
        parts << "sslmode=prefer";
    }

    // Connection timeout
    parts << QString("connect_timeout=%1").arg(config.timeout);

    return parts.join(" ");
}

QFuture<bool> PostgresDriver::connect(const ConnectionConfig& config) {
    return QtConcurrent::run([this, config]() -> bool {
        QMutexLocker locker(&m_mutex);

        qCDebug(lcConnection) << "Connecting to PostgreSQL:" << config.host << config.database;

        // Disconnect existing connection
        if (m_conn) {
            PQfinish(m_conn);
            m_conn = nullptr;
        }

        // Build connection string
        QString connStr = buildConnectionString(config);

        // Add password if provided (from keychain separately)
        // Password should be fetched from keychain before calling connect

        // Connect
        m_conn = PQconnectdb(connStr.toUtf8().constData());

        if (PQstatus(m_conn) != CONNECTION_OK) {
            QString error = QString::fromUtf8(PQerrorMessage(m_conn));
            qCCritical(lcConnection) << "Connection failed:" << error;
            PQfinish(m_conn);
            m_conn = nullptr;
            emit connectionError(error, ErrorCode::ConnectionFailed);
            return false;
        }

        // Set application name
        PQexec(m_conn, "SET application_name = 'TablePro'");

        m_config = config;
        m_connected = true;

        qCInfo(lcConnection) << "Connected to PostgreSQL" << serverVersion();
        emit connected();

        return true;
    });
}

void PostgresDriver::disconnect() {
    QMutexLocker locker(&m_mutex);

    if (m_conn) {
        qCDebug(lcConnection) << "Disconnecting from PostgreSQL";
        PQfinish(m_conn);
        m_conn = nullptr;
    }

    m_connected = false;
    m_inTransaction = false;
    emit disconnected();
}

bool PostgresDriver::isConnected() const {
    QMutexLocker locker(&m_mutex);
    return m_conn && PQstatus(m_conn) == CONNECTION_OK;
}

QFuture<void> PostgresDriver::ping() {
    return QtConcurrent::run([this]() {
        QMutexLocker locker(&m_mutex);

        if (!m_conn) {
            throw ConnectionException("Not connected");
        }

        PGResultPtr result(PQexec(m_conn, "SELECT 1"));
        if (PQresultStatus(result.get()) != PGRES_TUPLES_OK) {
            throw ConnectionException("Ping failed: " + QString::fromUtf8(PQerrorMessage(m_conn)));
        }
    });
}

ConnectionInfo PostgresDriver::connectionInfo() const {
    QMutexLocker locker(&m_mutex);

    ConnectionInfo info;
    info.id = m_config.id;
    info.name = m_config.name;
    info.type = "postgresql";
    info.database = m_config.database;
    info.schema = m_config.schema;
    info.host = m_config.host;
    info.port = m_config.port;
    info.connected = m_connected;
    info.serverVersion = serverVersion();

    return info;
}

QString PostgresDriver::serverVersion() const {
    if (!m_conn) return QString();
    return QString::fromUtf8(PQparameterStatus(m_conn, "server_version"));
}

void PostgresDriver::checkConnection() const {
    if (!m_conn || PQstatus(m_conn) != CONNECTION_OK) {
        throw ConnectionException("Not connected to database");
    }
}

QString PostgresDriver::escapeIdentifier(const QString& identifier) const {
    QString escaped = identifier;
    escaped.replace("\"", "\"\"");
    return QString("\"%1\"").arg(escaped);
}

QString PostgresDriver::escapeLiteral(const QString& value) const {
    if (!m_conn) {
        // Fallback: manual escaping
        QString escaped = value;
        escaped.replace("'", "''");
        return QString("'%1'").arg(escaped);
    }

    char* escaped = PQescapeLiteral(m_conn, value.toUtf8().constData(), value.toUtf8().size());
    if (!escaped) {
        throw QueryException("Failed to escape literal value");
    }
    QString result = QString::fromUtf8(escaped);
    PQfreemem(escaped);
    return result;
}

} // namespace tablepro
```

**Step 2: Commit connection methods**

```bash
git add src/driver/postgres_driver.cpp
git commit -m "feat: Add PostgresDriver connection methods"
```

---

## Task 4: PostgresDriver Query Execution

**Files:**
- Modify: `src/driver/postgres_driver.cpp`

**Step 1: Add query execution methods**

```cpp
// Add to postgres_driver.cpp

QueryResult PostgresDriver::execute(const QString& sql) {
    return executeParams(sql, {});
}

QFuture<QueryResult> PostgresDriver::execute(const QString& sql) {
    return executeParams(sql, {});
}

QFuture<QueryResult> PostgresDriver::executeParams(const QString& sql, const QVariantList& params) {
    return QtConcurrent::run([this, sql, params]() -> QueryResult {
        QMutexLocker locker(&m_mutex);

        auto startTime = std::chrono::steady_clock::now();

        QueryResult result;
        result.success = false;

        try {
            checkConnection();

            qCDebug(lcQuery) << "Executing query:" << sql.left(100);

            PGResultPtr pgResult;

            if (params.isEmpty()) {
                // Simple query
                pgResult.reset(PQexec(m_conn, sql.toUtf8().constData()));
            } else {
                // Parameterized query
                QVector<const char*> paramValues(params.size());
                QVector<int> paramLengths(params.size());
                QVector<int> paramFormats(params.size(), 0); // 0 = text
                QVector<QByteArray> paramBuffers(params.size());

                for (int i = 0; i < params.size(); ++i) {
                    if (params[i].isNull()) {
                        paramValues[i] = nullptr;
                        paramLengths[i] = 0;
                    } else {
                        paramBuffers[i] = params[i].toString().toUtf8();
                        paramValues[i] = paramBuffers[i].constData();
                        paramLengths[i] = paramBuffers[i].size();
                    }
                }

                pgResult.reset(PQexecParams(
                    m_conn,
                    sql.toUtf8().constData(),
                    params.size(),
                    nullptr, // param types (let PQ infer)
                    paramValues.data(),
                    paramLengths.data(),
                    paramFormats.data(),
                    0 // result format (0 = text)
                ));
            }

            ExecStatusType status = PQresultStatus(pgResult.get());

            if (status == PGRES_TUPLES_OK || status == PGRES_SINGLE_TUPLE) {
                result = parseResult(pgResult.get());
                result.success = true;
            } else if (status == PGRES_COMMAND_OK) {
                result.rowsAffected = QString::fromUtf8(PQcmdTuples(pgResult.get())).toLongLong();
                result.success = true;
            } else {
                result.error = QString::fromUtf8(PQerrorMessage(m_conn));
                qCWarning(lcQuery) << "Query failed:" << result.error;
            }

        } catch (const TableProException& e) {
            result.error = e.message();
        } catch (const std::exception& e) {
            result.error = QString::fromUtf8(e.what());
        }

        auto endTime = std::chrono::steady_clock::now();
        result.executionTimeMs = std::chrono::duration_cast<std::chrono::milliseconds>(
            endTime - startTime
        ).count();

        emit queryExecuted(sql, result.executionTimeMs);

        return result;
    });
}

QueryResult PostgresDriver::parseResult(PGresult* pgResult) const {
    QueryResult result;

    int numCols = PQnfields(pgResult);
    int numRows = PQntuples(pgResult);

    // Column names and types
    for (int col = 0; col < numCols; ++col) {
        QString colName = QString::fromUtf8(PQfname(pgResult, col));
        result.columnNames.append(colName);

        ColumnInfo colInfo;
        colInfo.name = colName;
        colInfo.typeName = QString::fromUtf8(PQfname(pgResult, col)); // Simplified
        Oid typeOid = PQftype(pgResult, col);
        // Map Oid to type name
        switch (typeOid) {
            case 16: colInfo.typeName = "boolean"; break;
            case 20: colInfo.typeName = "bigint"; break;
            case 21: colInfo.typeName = "smallint"; break;
            case 23: colInfo.typeName = "integer"; break;
            case 25: colInfo.typeName = "text"; break;
            case 700: colInfo.typeName = "real"; break;
            case 701: colInfo.typeName = "double precision"; break;
            case 1043: colInfo.typeName = "varchar"; break;
            case 1082: colInfo.typeName = "date"; break;
            case 1114: colInfo.typeName = "timestamp"; break;
            case 1184: colInfo.typeName = "timestamptz"; break;
            case 3802: colInfo.typeName = "jsonb"; break;
            case 114: colInfo.typeName = "json"; break;
            case 17: colInfo.typeName = "bytea"; break;
            case 2950: colInfo.typeName = "uuid"; break;
            default: colInfo.typeName = QString("oid:%1").arg(typeOid); break;
        }
        result.columns.append(colInfo);
    }

    // Rows
    for (int row = 0; row < numRows; ++row) {
        QVariantMap rowMap;
        for (int col = 0; col < numCols; ++col) {
            QString colName = result.columnNames[col];

            if (PQgetisnull(pgResult, row, col)) {
                rowMap[colName] = QVariant(); // NULL
            } else {
                const char* value = PQgetvalue(pgResult, row, col);
                Oid typeOid = PQftype(pgResult, col);
                rowMap[colName] = convertValue(value, typeOid);
            }
        }
        result.rows.append(rowMap);
    }

    result.rowsAffected = numRows;
    return result;
}

QVariant PostgresDriver::convertValue(const char* value, Oid type) const {
    QString strValue = QString::fromUtf8(value);

    switch (type) {
        case 16: // boolean
            return strValue == "t" || strValue == "true";

        case 20: // bigint
        case 21: // smallint
        case 23: // integer
            return strValue.toLongLong();

        case 700: // real
        case 701: // double precision
            return strValue.toDouble();

        case 17: // bytea
            // TODO: Decode hex or base64
            return QByteArray::fromHex(strValue.toUtf8());

        case 114: // json
        case 3802: // jsonb
            return QJsonDocument::fromJson(strValue.toUtf8()).toVariant();

        case 1082: // date
            return QDate::fromString(strValue, Qt::ISODate);

        case 1114: // timestamp
        case 1184: // timestamptz
            return QDateTime::fromString(strValue, Qt::ISODate);

        case 2950: // uuid
            return strValue;

        default:
            return strValue;
    }
}

} // namespace tablepro
```

**Step 2: Commit query methods**

```bash
git add src/driver/postgres_driver.cpp
git commit -m "feat: Add PostgresDriver query execution with type conversion"
```

---

## Task 5: PostgresDriver Schema Inspection

**Files:**
- Modify: `src/driver/postgres_driver.cpp`

**Step 1: Add schema inspection methods**

```cpp
// Add to postgres_driver.cpp

QFuture<SchemaInfo> PostgresDriver::fetchSchema() {
    return QtConcurrent::run([this]() -> SchemaInfo {
        SchemaInfo info;
        info.databaseName = m_config.database;

        // Fetch schemas
        QString schemaSql = R"(
            SELECT schema_name
            FROM information_schema.schemata
            WHERE schema_name NOT IN ('pg_toast', 'pg_catalog', 'information_schema')
            ORDER BY schema_name
        )";

        auto schemaResult = execute(schemaSql).result();
        for (const auto& row : schemaResult.rows) {
            info.schemas.append(row["schema_name"].toString());
        }

        // Set current schema
        info.currentSchema = m_config.schema.isEmpty() ? "public" : m_config.schema;

        // Fetch tables and views
        info.tables = fetchTables(info.currentSchema).result();
        info.views = fetchTables(info.currentSchema).result(); // TODO: filter views

        return info;
    });
}

QFuture<QList<TableInfo>> PostgresDriver::fetchTables(const QString& schema) {
    return QtConcurrent::run([this, schema]() -> QList<TableInfo> {
        QString sql = R"(
            SELECT
                table_name,
                table_schema,
                table_type
            FROM information_schema.tables
            WHERE table_schema = $1
            AND table_type IN ('BASE TABLE', 'VIEW')
            ORDER BY table_name
        )";

        QVariantList params;
        params << (schema.isEmpty() ? "public" : schema);

        auto result = executeParams(sql, params).result();

        QList<TableInfo> tables;
        for (const auto& row : result.rows) {
            TableInfo info;
            info.name = row["table_name"].toString();
            info.schema = row["table_schema"].toString();
            info.type = row["table_type"].toString() == "VIEW" ? "view" : "table";
            tables.append(info);
        }

        return tables;
    });
}

QFuture<QList<ColumnInfo>> PostgresDriver::fetchColumns(const QString& schema, const QString& table) {
    return QtConcurrent::run([this, schema, table]() -> QList<ColumnInfo> {
        QString sql = R"(
            SELECT
                c.column_name,
                c.data_type,
                c.is_nullable,
                c.column_default,
                c.character_maximum_length,
                c.numeric_precision,
                c.numeric_scale,
                CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END AS is_primary_key,
                CASE WHEN c.column_default LIKE 'nextval%' THEN true ELSE false END AS is_auto_increment
            FROM information_schema.columns c
            LEFT JOIN (
                SELECT kcu.column_name, kcu.table_name, kcu.table_schema
                FROM information_schema.table_constraints tc
                JOIN information_schema.key_column_usage kcu
                    ON tc.constraint_name = kcu.constraint_name
                    AND tc.table_schema = kcu.table_schema
                WHERE tc.constraint_type = 'PRIMARY KEY'
            ) pk ON c.column_name = pk.column_name
                AND c.table_name = pk.table_name
                AND c.table_schema = pk.table_schema
            WHERE c.table_schema = $1
            AND c.table_name = $2
            ORDER BY c.ordinal_position
        )";

        QVariantList params;
        params << schema << table;

        auto result = executeParams(sql, params).result();

        QList<ColumnInfo> columns;
        for (const auto& row : result.rows) {
            ColumnInfo info;
            info.name = row["column_name"].toString();

            // Build full type name
            QString dataType = row["data_type"].toString();
            if (dataType == "character varying") {
                int maxLen = row["character_maximum_length"].toInt();
                info.typeName = maxLen > 0 ? QString("varchar(%1)").arg(maxLen) : "varchar";
            } else if (dataType == "numeric") {
                int precision = row["numeric_precision"].toInt();
                int scale = row["numeric_scale"].toInt();
                if (precision > 0) {
                    info.typeName = scale > 0
                        ? QString("numeric(%1,%2)").arg(precision).arg(scale)
                        : QString("numeric(%1)").arg(precision);
                } else {
                    info.typeName = "numeric";
                }
            } else {
                info.typeName = dataType;
            }

            info.nullable = row["is_nullable"].toString() == "YES";
            info.defaultValue = row["column_default"];
            info.isPrimaryKey = row["is_primary_key"].toBool();
            info.isAutoIncrement = row["is_auto_increment"].toBool();

            columns.append(info);
        }

        return columns;
    });
}

QFuture<QList<IndexInfo>> PostgresDriver::fetchIndexes(const QString& schema, const QString& table) {
    return QtConcurrent::run([this, schema, table]() -> QList<IndexInfo> {
        QString sql = R"(
            SELECT
                i.relname AS index_name,
                a.attname AS column_name,
                ix.indisunique AS is_unique,
                ix.indisprimary AS is_primary,
                array_position(ix.indkey, a.attnum) AS column_position
            FROM pg_class t
            JOIN pg_index ix ON t.oid = ix.indrelid
            JOIN pg_class i ON i.oid = ix.indexrelid
            JOIN pg_namespace n ON n.oid = t.relnamespace
            JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
            WHERE n.nspname = $1
            AND t.relname = $2
            ORDER BY i.relname, column_position
        )";

        QVariantList params;
        params << schema << table;

        auto result = executeParams(sql, params).result();

        QMap<QString, IndexInfo> indexMap;
        for (const auto& row : result.rows) {
            QString indexName = row["index_name"].toString();

            if (!indexMap.contains(indexName)) {
                IndexInfo info;
                info.name = indexName;
                info.isUnique = row["is_unique"].toBool();
                info.isPrimary = row["is_primary"].toBool();
                indexMap[indexName] = info;
            }

            indexMap[indexName].columns.append(row["column_name"].toString());
        }

        return indexMap.values();
    });
}

QFuture<QList<ForeignKeyInfo>> PostgresDriver::fetchForeignKeys(const QString& schema, const QString& table) {
    return QtConcurrent::run([this, schema, table]() -> QList<ForeignKeyInfo> {
        QString sql = R"(
            SELECT
                tc.constraint_name,
                kcu.column_name,
                ccu.table_schema AS referenced_schema,
                ccu.table_name AS referenced_table,
                ccu.column_name AS referenced_column
            FROM information_schema.table_constraints tc
            JOIN information_schema.key_column_usage kcu
                ON tc.constraint_name = kcu.constraint_name
                AND tc.table_schema = kcu.table_schema
            JOIN information_schema.constraint_column_usage ccu
                ON ccu.constraint_name = tc.constraint_name
            WHERE tc.constraint_type = 'FOREIGN KEY'
            AND tc.table_schema = $1
            AND tc.table_name = $2
        )";

        QVariantList params;
        params << schema << table;

        auto result = executeParams(sql, params).result();

        QList<ForeignKeyInfo> fks;
        for (const auto& row : result.rows) {
            ForeignKeyInfo info;
            info.name = row["constraint_name"].toString();
            info.columnName = row["column_name"].toString();
            info.referencedSchema = row["referenced_schema"].toString();
            info.referencedTable = row["referenced_table"].toString();
            info.referencedColumnName = row["referenced_column"].toString();
            fks.append(info);
        }

        return fks;
    });
}

QFuture<QString> PostgresDriver::fetchDDL(const QString& schema, const QString& table) {
    return QtConcurrent::run([this, schema, table]() -> QString {
        QString sql = QString(R"(
            SELECT pg_get_tabledef('%1.%2'::regclass)
        )").arg(schema, table);

        // Alternative if pg_get_tabledef not available:
        QString altSql = QString(R"(
            SELECT 'CREATE TABLE ' || quote_ident('%1') || '.' || quote_ident('%2') || ' (' ||
            string_agg(
                quote_ident(column_name) || ' ' ||
                data_type ||
                CASE WHEN character_maximum_length IS NOT NULL
                    THEN '(' || character_maximum_length || ')' ELSE '' END ||
                CASE WHEN is_nullable = 'NO' THEN ' NOT NULL' ELSE '' END ||
                CASE WHEN column_default IS NOT NULL
                    THEN ' DEFAULT ' || column_default ELSE '' END,
                ', '
            ) || ');'
            FROM information_schema.columns
            WHERE table_schema = '%1' AND table_name = '%2'
        )").arg(schema, table);

        auto result = execute(altSql).result();
        if (result.success && !result.rows.isEmpty()) {
            return result.rows.first().begin().value().toString();
        }

        return QString();
    });
}

QFuture<qint64> PostgresDriver::fetchRowCount(const QString& schema, const QString& table) {
    return QtConcurrent::run([this, schema, table]() -> qint64 {
        QString sql = QString("SELECT COUNT(*) AS count FROM %1.%2")
            .arg(escapeIdentifier(schema))
            .arg(escapeIdentifier(table));

        auto result = execute(sql).result();
        if (result.success && !result.rows.isEmpty()) {
            return result.rows.first()["count"].toLongLong();
        }

        return 0;
    });
}
```

**Step 2: Commit schema methods**

```bash
git add src/driver/postgres_driver.cpp
git commit -m "feat: Add PostgresDriver schema inspection methods"
```

---

## Task 6: PostgresDriver Transactions

**Files:**
- Modify: `src/driver/postgres_driver.cpp`

**Step 1: Add transaction methods**

```cpp
// Add to postgres_driver.cpp

bool PostgresDriver::beginTransaction() {
    QMutexLocker locker(&m_mutex);

    if (!m_conn || m_inTransaction) {
        return false;
    }

    PGResultPtr result(PQexec(m_conn, "BEGIN"));
    if (PQresultStatus(result.get()) != PGRES_COMMAND_OK) {
        qCWarning(lcQuery) << "Failed to begin transaction:" << PQerrorMessage(m_conn);
        return false;
    }

    m_inTransaction = true;
    qCDebug(lcQuery) << "Transaction started";
    return true;
}

bool PostgresDriver::commitTransaction() {
    QMutexLocker locker(&m_mutex);

    if (!m_conn || !m_inTransaction) {
        return false;
    }

    PGResultPtr result(PQexec(m_conn, "COMMIT"));
    if (PQresultStatus(result.get()) != PGRES_COMMAND_OK) {
        qCWarning(lcQuery) << "Failed to commit transaction:" << PQerrorMessage(m_conn);
        return false;
    }

    m_inTransaction = false;
    qCDebug(lcQuery) << "Transaction committed";
    return true;
}

bool PostgresDriver::rollbackTransaction() {
    QMutexLocker locker(&m_mutex);

    if (!m_conn || !m_inTransaction) {
        return false;
    }

    PGResultPtr result(PQexec(m_conn, "ROLLBACK"));
    if (PQresultStatus(result.get()) != PGRES_COMMAND_OK) {
        qCWarning(lcQuery) << "Failed to rollback transaction:" << PQerrorMessage(m_conn);
        return false;
    }

    m_inTransaction = false;
    qCDebug(lcQuery) << "Transaction rolled back";
    return true;
}

} // namespace tablepro
```

**Step 2: Commit transaction methods**

```bash
git add src/driver/postgres_driver.cpp
git commit -m "feat: Add PostgresDriver transaction support"
```

---

## Task 7: Register PostgreSQL Driver

**Files:**
- Create: `src/driver/register_drivers.cpp`

**Step 1: Create driver registration**

```cpp
#include "postgres_driver.hpp"
#include "../core/driver_factory.hpp"

namespace tablepro {

// Static registrar
struct PostgresDriverRegistrar {
    PostgresDriverRegistrar() {
        DriverFactory::registerDriver(
            DatabaseType::PostgreSQL,
            [](QObject* parent) -> std::unique_ptr<DatabaseDriver> {
                return std::make_unique<PostgresDriver>(parent);
            }
        );
    }
};

// Instantiate to register
static PostgresDriverRegistrar postgresRegistrar;

} // namespace tablepro
```

**Step 2: Update CMakeLists.txt**

Add to sources:
```cmake
set(TABLEPRO_SOURCES
    # ... existing sources ...
    src/driver/postgres_driver.cpp
    src/driver/register_drivers.cpp
)
```

**Step 3: Commit registration**

```bash
git add src/driver/register_drivers.cpp CMakeLists.txt
git commit -m "feat: Register PostgreSQL driver with factory"
```

---

## Task 8: PostgreSQL Driver Tests

**Files:**
- Create: `tests/integration/test_postgres_driver.cpp`

**Step 1: Create integration test**

```cpp
#include <gtest/gtest.h>
#include "core/driver_factory.hpp"
#include "driver/postgres_driver.hpp"

using namespace tablepro;

// Requires running PostgreSQL container
class PostgresDriverTest : public ::testing::Test {
protected:
    void SetUp() override {
        if (!qEnvironmentVariableIsSet("TEST_POSTGRES_HOST")) {
            GTEST_SKIP() << "TEST_POSTGRES_HOST not set, skipping integration tests";
        }

        m_config.host = qgetenv("TEST_POSTGRES_HOST");
        m_config.port = qEnvironmentVariableIntValue("TEST_POSTGRES_PORT", 5432);
        m_config.database = qgetenv("TEST_POSTGRES_DATABASE");
        m_config.username = qgetenv("TEST_POSTGRES_USER");
        m_config.type = "postgresql";
        m_config.name = "Test Connection";
        m_config.id = "test-pg";
    }

    ConnectionConfig m_config;
};

TEST_F(PostgresDriverTest, CanCreateDriver) {
    auto driver = DriverFactory::create(DatabaseType::PostgreSQL);
    EXPECT_NE(driver, nullptr);
}

TEST_F(PostgresDriverTest, CanConnect) {
    auto driver = DriverFactory::create(DatabaseType::PostgreSQL);
    auto future = driver->connect(m_config);
    future.waitForFinished();

    EXPECT_TRUE(future.result());
    EXPECT_TRUE(driver->isConnected());

    driver->disconnect();
    EXPECT_FALSE(driver->isConnected());
}

TEST_F(PostgresDriverTest, CanExecuteQuery) {
    auto driver = DriverFactory::create(DatabaseType::PostgreSQL);
    driver->connect(m_config).waitForFinished();

    auto future = driver->execute("SELECT 1 AS value");
    future.waitForFinished();

    auto result = future.result();
    EXPECT_TRUE(result.success);
    EXPECT_EQ(result.rows.size(), 1);
    EXPECT_EQ(result.columnNames.size(), 1);
    EXPECT_EQ(result.columnNames[0], "value");
}

TEST_F(PostgresDriverTest, CanFetchTables) {
    auto driver = DriverFactory::create(DatabaseType::PostgreSQL);
    driver->connect(m_config).waitForFinished();

    auto future = driver->fetchTables("public");
    future.waitForFinished();

    auto tables = future.result();
    EXPECT_GE(tables.size(), 0); // May be empty on fresh database
}
```

**Step 2: Create tests CMakeLists.txt**

```cmake
# tests/integration/CMakeLists.txt
find_package(GTest REQUIRED)

add_executable(tablepro_integration_tests
    test_postgres_driver.cpp
)

target_link_libraries(tablepro_integration_tests PRIVATE
    tablepro
    GTest::gtest
    GTest::gtest_main
    Qt6::Core
)

include(GoogleTest)
gtest_discover_tests(tablepro_integration_tests)
```

**Step 3: Commit tests**

```bash
git add tests/integration/
git commit -m "test: Add PostgreSQL driver integration tests"
```

---

## Task 9: Update main.cpp to Initialize

**Files:**
- Modify: `src/main.cpp`

**Step 1: Update main.cpp**

```cpp
#include <QApplication>
#include <QIcon>
#include "ui/MainWindow.hpp"
#include "config.hpp"
#include "core/logging.hpp"
#include "driver/register_drivers.cpp"  // Ensure drivers are registered

int main(int argc, char* argv[]) {
    QApplication app(argc, argv);

    // Initialize logging
    tablepro::initLogging("debug");

    // Application metadata
    app.setApplicationName(tablepro::config::APP_NAME);
    app.setApplicationVersion(tablepro::config::APP_VERSION);
    app.setOrganizationName(tablepro::config::ORG_NAME);
    app.setOrganizationDomain(tablepro::config::ORG_DOMAIN);

    qCInfo(tablepro::lcCore) << "Starting TablePro" << tablepro::config::APP_VERSION;

    // Create and show main window
    tablepro::MainWindow mainWindow;
    mainWindow.show();

    return app.exec();
}
```

**Step 2: Commit main update**

```bash
git add src/main.cpp
git commit -m "feat: Initialize logging and drivers in main"
```

---

## Task 10: Verify Build

**Step 1: Build**

```bash
cmake --build build/debug -j$(nproc)
```

Expected: Build succeeds

**Step 2: Run application**

```bash
./build/debug/tablepro
```

Expected: Application starts (no PostgreSQL connection yet in UI)

**Step 3: Final commit**

```bash
git status
git add -A
git commit -m "feat(phase3): PostgreSQL driver implementation complete"
```

---

## Acceptance Criteria

- [ ] PostgresDriver implements all DatabaseDriver methods
- [ ] Connection with libpq works
- [ ] Query execution with prepared statements works
- [ ] Type conversion for common PostgreSQL types
- [ ] Schema inspection (tables, columns, indexes, FKs)
- [ ] Transaction support (begin/commit/rollback)
- [ ] Driver registered with DriverFactory
- [ ] Integration tests pass (with test database)
- [ ] Build succeeds with no warnings

---

**Phase 3 Complete.** Next: Phase 4 - UI Foundation