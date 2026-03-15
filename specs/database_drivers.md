# Database Drivers Specification (C++20)

Each database driver implements the `DatabaseDriver` interface. Drivers wrap native C/C++ client libraries with RAII semantics.

## 1. PostgreSQL
- **Library**: `libpq` (via `libpq-fe.h`, C API)
- **CMake Target**: `PQ::PQ`
- **Features**: TLS/SSL, SSH tunneling, multi-schema, JSONB, arrays, LISTEN/NOTIFY
- **Dialect**: `LIMIT $1 OFFSET $2`, identifiers quoted with `"`, params `$1, $2, ...`
- **Schema Queries**: `pg_catalog`, `information_schema`
- **Connection Pooling**: Custom pool using `std::vector<std::unique_ptr<PQconn>>`

```cpp
class PostgresDriver : public DatabaseDriver {
public:
    bool connect(const ConnectionConfig& config) override;
    void disconnect() override;
    QueryResult execute(const QString& sql) override;
    SchemaInfo introspectSchema() override;

private:
    std::unique_ptr<PQconn, PqDeleter> m_connection;  // RAII wrapper
    QString m_currentDatabase;
};
```

## 2. MySQL / MariaDB
- **Library**: MySQL Connector/C (`libmysqlclient`)
- **CMake Target**: `MySQL::MySQL`
- **Features**: User/Pass/Host/Port, SSH tunneling, SSL, multiple databases
- **Dialect**: `` LIMIT ? OFFSET ? ``, identifiers quoted with `` ` ``, params `?`
- **Schema Queries**: `information_schema`

```cpp
class MysqlDriver : public DatabaseDriver {
public:
    bool connect(const ConnectionConfig& config) override;
    void disconnect() override;
    QueryResult execute(const QString& sql) override;
    SchemaInfo introspectSchema() override;

private:
    std::unique_ptr<MYSQL, MysqlDeleter> m_connection;  // RAII wrapper
    QString m_currentDatabase;
};
```

## 3. SQLite
- **Library**: Qt SQL module (built-in SQLite backend)
- **CMake Target**: `Qt6::Sql`
- **Features**: File-based, no network, WAL mode, PRAGMAs
- **Dialect**: `LIMIT ? OFFSET ?`, basic types (INTEGER, REAL, TEXT, BLOB)

```cpp
class SqliteDriver : public DatabaseDriver {
public:
    bool connect(const ConnectionConfig& config) override;
    void disconnect() override;
    QueryResult execute(const QString& sql) override;
    SchemaInfo introspectSchema() override;

private:
    QSqlDatabase m_database;  // Qt SQL handles cleanup
    QString m_filePath;
};
```

## 4. Microsoft SQL Server
- **Library**: FreeTDS (`libsybdb`) or Microsoft ODBC
- **CMake Target**: `FreeTDS::FreeTDS`
- **Features**: SQL Server Auth, Windows Auth (NTLM), multiple databases
- **Dialect**: `TOP N`, identifiers `[name]`, offset `OFFSET ? ROWS FETCH NEXT ? ROWS ONLY`

## 5. DuckDB
- **Library**: DuckDB C API (`libduckdb`)
- **CMake Target**: `duckdb::duckdb`
- **Features**: Local `.duckdb` files, CSV/Parquet reading, extensions
- **Dialect**: PostgreSQL-compatible syntax

```cpp
class DuckDbDriver : public DatabaseDriver {
public:
    bool connect(const ConnectionConfig& config) override;
    void disconnect() override;
    QueryResult execute(const QString& sql) override;
    SchemaInfo introspectSchema() override;

private:
    duckdb_database m_database{nullptr};
    duckdb_connection m_connection{nullptr};
};
```

## 6. ClickHouse
- **Library**: `clickhouse-cpp` (C++ client)
- **CMake Target**: `clickhouse-cpp::clickhouse-cpp`
- **Features**: Column-oriented analytics, batch inserts, compression
- **Dialect**: `` LIMIT ? OFFSET ? ``, backtick identifiers

## 7. MongoDB
- **Library**: `libmongoc` (MongoDB C driver)
- **CMake Target**: `mongo::mongoc_shared`
- **Features**: URI strings, SSH tunneling, Auth Source, Replica Sets, BSON→JSON
- **Special Handling**: Editor sends BSON queries. Driver translates to `bson_t` internally.

## 8. Redis
- **Library**: `hiredis` (Redis C client)
- **CMake Target**: `hiredis::hiredis`
- **Features**: Key browsing (Strings, Lists, Sets, Hashes, Sorted Sets), DB index selection (0-15)
- **Special Handling**: Editor uses Redis commands (`GET`, `HSET`, `KEYS *`), not SQL. Results as key-value pairs.

## Driver Interface

```cpp
// src/core/DatabaseDriver.hpp
#pragma once

#include <QObject>
#include <QString>
#include <memory>
#include "../specs/data_models.md"

namespace tablepro {

enum class DatabaseType {
    PostgreSQL,
    MySQL,
    SQLite,
    DuckDB,
    MSSQL,
    ClickHouse,
    MongoDB,
    Redis,
    Unknown
};

class DatabaseDriver : public QObject {
    Q_OBJECT

public:
    virtual ~DatabaseDriver() = default;

    // Connection lifecycle
    virtual bool connect(const ConnectionConfig& config) = 0;
    virtual void disconnect() = 0;
    virtual bool isConnected() const = 0;

    // Query execution
    virtual QueryResult execute(const QString& sql) = 0;
    virtual QueryResult executeWithPagination(
        const QString& sql, int offset, int limit,
        const QString& orderBy = {}, const QString& orderDir = "ASC") = 0;

    // Schema introspection
    virtual QList<DatabaseInfo> listDatabases() = 0;
    virtual SchemaInfo introspectSchema(const QString& database) = 0;
    virtual QString getDdl(const QString& tableName) = 0;

    // Capabilities
    virtual DatabaseType type() const = 0;
    virtual QString name() const = 0;
    virtual bool supportsTransactions() const { return true; }
    virtual bool supportsSchemas() const { return false; }

signals:
    void connected();
    void disconnected();
    void errorOccurred(const QString& message);
};

} // namespace tablepro
```

## Dialect Handling

```cpp
// src/core/SqlGenerator.hpp
#pragma once

#include "DatabaseDriver.hpp"

namespace tablepro {

class SqlDialect {
public:
    virtual ~SqlDialect() = default;

    // Quoting
    virtual QString quoteIdentifier(const QString& name) const = 0;
    virtual QString quoteString(const QString& value) const = 0;

    // Pagination
    virtual QString wrapWithPagination(
        const QString& sql, int offset, int limit,
        const QString& orderBy, const QString& orderDir) const = 0;

    // Count wrapper
    virtual QString wrapWithCount(const QString& sql) const = 0;

    // EXPLAIN
    virtual QString wrapWithExplain(const QString& sql) const = 0;

    // Upsert (INSERT ... ON CONFLICT / ON DUPLICATE KEY)
    virtual QString buildUpsert(
        const QString& table,
        const QStringList& columns,
        const QVariantList& values,
        const QStringList& keyColumns) const = 0;
};

// Concrete dialects
class PostgresDialect : public SqlDialect {
    QString quoteIdentifier(const QString& name) const override {
        return "\"" + name + "\"";
    }
    QString wrapWithPagination(...) const override {
        return sql + " ORDER BY " + orderBy + " " + orderDir +
               " LIMIT " + QString::number(limit) +
               " OFFSET " + QString::number(offset);
    }
    // ...
};

class MySqlDialect : public SqlDialect {
    QString quoteIdentifier(const QString& name) const override {
        return "`" + name + "`";
    }
    // ...
};

class SqlServerDialect : public SqlDialect {
    QString quoteIdentifier(const QString& name) const override {
        return "[" + name + "]";
    }
    QString wrapWithPagination(...) const override {
        return sql + " ORDER BY " + orderBy + " " + orderDir +
               " OFFSET " + QString::number(offset) + " ROWS " +
               "FETCH NEXT " + QString::number(limit) + " ROWS ONLY";
    }
};

} // namespace tablepro
```

## Driver Factory

```cpp
// src/core/DriverFactory.hpp
#pragma once

#include "DatabaseDriver.hpp"
#include <memory>

namespace tablepro {

class DriverFactory {
public:
    static std::unique_ptr<DatabaseDriver> create(DatabaseType type);
};

} // namespace tablepro

// src/core/DriverFactory.cpp
#include "PostgresDriver.hpp"
#include "MysqlDriver.hpp"
#include "SqliteDriver.hpp"
#include "DuckDbDriver.hpp"

std::unique_ptr<DatabaseDriver> DriverFactory::create(DatabaseType type) {
    switch (type) {
        case DatabaseType::PostgreSQL:
            return std::make_unique<PostgresDriver>();
        case DatabaseType::MySQL:
            return std::make_unique<MysqlDriver>();
        case DatabaseType::SQLite:
            return std::make_unique<SqliteDriver>();
        case DatabaseType::DuckDB:
            return std::make_unique<DuckDbDriver>();
        // ... other drivers
        default:
            return nullptr;
    }
}
```

## Build Configuration

```cmake
# CMakeLists.txt
find_package(PQ REQUIRED)
find_package(MySQL REQUIRED)
find_package(duckdb REQUIRED)
find_package(hiredis REQUIRED)
find_package(Libssh2 REQUIRED)

# Driver sources
set(DRIVER_SOURCES
    src/core/DatabaseDriver.cpp
    src/core/PostgresDriver.cpp
    src/core/MysqlDriver.cpp
    src/core/SqliteDriver.cpp
    src/core/DuckDbDriver.cpp
    src/core/SqlGenerator.cpp
    src/core/DriverFactory.cpp
)

add_library(drivers STATIC ${DRIVER_SOURCES})
target_link_libraries(drivers PUBLIC
    Qt6::Core
    Qt6::Sql
    PQ::PQ
    MySQL::MySQL
    duckdb::duckdb
    hiredis::hiredis
)
```

## RAII Deleters

```cpp
// src/core/drivers/PqDeleter.hpp
#pragma once
#include <libpq-fe.h>

struct PqDeleter {
    void operator()(PQconn* conn) const {
        if (conn) {
            PQfinish(conn);
        }
    }
};

// src/core/drivers/MysqlDeleter.hpp
#pragma once
#include <mysql.h>

struct MysqlDeleter {
    void operator()(MYSQL* mysql) const {
        if (mysql) {
            mysql_close(mysql);
            // mysql_init() allocates, mysql_close() frees
        }
    }
};
```
