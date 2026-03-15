# Phase 9: Additional Drivers Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement remaining 7 database drivers (MySQL, SQLite, DuckDB, SQL Server, ClickHouse, MongoDB, Redis).

**Architecture:** Each driver implements DatabaseDriver interface. Follow PostgreSQL driver patterns. Register with DriverFactory.

**Tech Stack:** C++20, libmysql, Qt SQL (SQLite), duckdb, ODBC, libmongocxx, hiredis

---

## Task 1: MySQL Driver

**Files:**
- Create: `src/driver/mysql_driver.hpp`
- Create: `src/driver/mysql_driver.cpp`

**Step 1: Add MySQL dependency to vcpkg.json**

```json
"libmysql"
```

**Step 2: Create MySQL driver following PostgreSQL pattern**

Key differences:
- Use `mysql_real_connect()` for connection
- Use `mysql_stmt_prepare()` / `mysql_stmt_execute()` for prepared statements
- Use `information_schema` for metadata queries
- Quote identifiers with backticks `` ` ``

**Step 3: Register MySQL driver**

```cpp
DriverFactory::registerDriver(DatabaseType::MySQL, [](QObject* p) {
    return std::make_unique<MySqlDriver>(p);
});
```

**Commit:**

```bash
git add src/driver/mysql_driver.*
git commit -m "feat: Add MySQL driver"
```

---

## Task 2: SQLite Driver

**Files:**
- Create: `src/driver/sqlite_driver.hpp`
- Create: `src/driver/sqlite_driver.cpp`

**Step 1: Use Qt SQL module (already available)**

Key points:
- Use `QSqlDatabase::addDatabase("QSQLITE")`
- Use `QSqlQuery` for execution
- Query `sqlite_master` for metadata
- Handle ATTACH DATABASE for multiple databases

**Step 2: SQLite-specific features**

- WAL mode support
- `sqlite_master` for schema
- No network support (file-based only)
- `last_insert_rowid()` for auto-increment

**Commit:**

```bash
git add src/driver/sqlite_driver.*
git commit -m "feat: Add SQLite driver using Qt SQL"
```

---

## Task 3: DuckDB Driver

**Files:**
- Create: `src/driver/duckdb_driver.hpp`
- Create: `src/driver/duckdb_driver.cpp`

**Step 1: Add DuckDB dependency to vcpkg.json**

```json
"duckdb"
```

**Step 2: Implement DuckDB driver**

Key points:
- Use `duckdb_open()` / `duckdb_connect()`
- Use `duckdb_prepare()` / `duckdb_execute_prepared()`
- Columnar result handling
- Support for LIST, STRUCT, MAP types

**Commit:**

```bash
git add src/driver/duckdb_driver.*
git commit -m "feat: Add DuckDB driver"
```

---

## Task 4: SQL Server Driver

**Files:**
- Create: `src/driver/sqlserver_driver.hpp`
- Create: `src/driver/sqlserver_driver.cpp`

**Step 1: Add ODBC dependency**

```json
"odbc"
```

**Step 2: Implement SQL Server driver**

Key points:
- Use ODBC API (`SQLConnect`, `SQLExecDirect`, etc.)
- Connection string format: `Driver={ODBC Driver 17 for SQL Server};Server=...`
- Query `sys.tables`, `sys.columns` for metadata
- Handle `NVARCHAR`, `DATETIME2`, `UNIQUEIDENTIFIER`

**Commit:**

```bash
git add src/driver/sqlserver_driver.*
git commit -m "feat: Add SQL Server driver via ODBC"
```

---

## Task 5: ClickHouse Driver

**Files:**
- Create: `src/driver/clickhouse_driver.hpp`
- Create: `src/driver/clickhouse_driver.cpp`

**Step 1: Add dependency**

```json
"clickhouse-cpp"
```

**Step 2: Implement ClickHouse driver**

Key points:
- HTTP or TCP protocol
- Column-oriented result handling
- Query `system.tables`, `system.columns`
- Handle Array, Tuple, Map, Nested types

**Commit:**

```bash
git add src/driver/clickhouse_driver.*
git commit -m "feat: Add ClickHouse driver"
```

---

## Task 6: MongoDB Driver

**Files:**
- Create: `src/driver/mongodb_driver.hpp`
- Create: `src/driver/mongodb_driver.cpp`

**Step 1: Add MongoDB dependencies**

```json
"mongo-cxx-driver"
```

**Step 2: Implement MongoDB driver**

Key points:
- Use `mongocxx::client`
- BSON document handling
- Collections as "tables"
- `listCollections()` for schema
- Aggregation pipeline support

**Commit:**

```bash
git add src/driver/mongodb_driver.*
git commit -m "feat: Add MongoDB driver"
```

---

## Task 7: Redis Driver

**Files:**
- Create: `src/driver/redis_driver.hpp`
- Create: `src/driver/redis_driver.cpp`

**Step 1: Add hiredis dependency**

```json
"hiredis"
```

**Step 2: Implement Redis driver**

Key points:
- Use `redisConnect()` / `redisCommand()`
- Execute commands as queries: `GET key`, `HGETALL hash`
- `INFO` for server info
- `KEYS pattern` for listing
- Type-aware value parsing (String, Hash, List, Set, ZSet)

**Commit:**

```bash
git add src/driver/redis_driver.*
git commit -m "feat: Add Redis driver"
```

---

## Task 8: Update Driver Registration

**Files:**
- Modify: `src/driver/register_drivers.cpp`

**Step 1: Register all drivers**

```cpp
#include "postgres_driver.hpp"
#include "mysql_driver.hpp"
#include "sqlite_driver.hpp"
#include "duckdb_driver.hpp"
#include "sqlserver_driver.hpp"
#include "clickhouse_driver.hpp"
#include "mongodb_driver.hpp"
#include "redis_driver.hpp"

namespace tablepro {

struct DriverRegistrar {
    DriverRegistrar() {
        DriverFactory::registerDriver(DatabaseType::PostgreSQL,
            [](QObject* p) { return std::make_unique<PostgresDriver>(p); });
        DriverFactory::registerDriver(DatabaseType::MySQL,
            [](QObject* p) { return std::make_unique<MySqlDriver>(p); });
        DriverFactory::registerDriver(DatabaseType::SQLite,
            [](QObject* p) { return std::make_unique<SqliteDriver>(p); });
        DriverFactory::registerDriver(DatabaseType::DuckDB,
            [](QObject* p) { return std::make_unique<DuckDbDriver>(p); });
        DriverFactory::registerDriver(DatabaseType::SQLServer,
            [](QObject* p) { return std::make_unique<SqlServerDriver>(p); });
        DriverFactory::registerDriver(DatabaseType::ClickHouse,
            [](QObject* p) { return std::make_unique<ClickHouseDriver>(p); });
        DriverFactory::registerDriver(DatabaseType::MongoDB,
            [](QObject* p) { return std::make_unique<MongoDbDriver>(p); });
        DriverFactory::registerDriver(DatabaseType::Redis,
            [](QObject* p) { return std::make_unique<RedisDriver>(p); });
    }
};

static DriverRegistrar registrar;

} // namespace tablepro
```

**Commit:**

```bash
git add src/driver/register_drivers.cpp
git commit -m "feat: Register all 8 database drivers"
```

---

## Task 9: Update CMakeLists.txt

**Step 1: Add all driver sources and dependencies**

```cmake
# MySQL
find_package(MySQL REQUIRED)
target_link_libraries(tablepro PRIVATE ${MYSQL_LIBRARIES})

# DuckDB
find_package(duckdb REQUIRED)
target_link_libraries(tablepro PRIVATE duckdb::duckdb)

# ODBC (SQL Server)
find_package(ODBC REQUIRED)
target_link_libraries(tablepro PRIVATE ODBC::ODBC)

# MongoDB
find_package(mongocxx REQUIRED)
target_link_libraries(tablepro PRIVATE mongo::mongocxx_static)

# Redis
find_package(hiredis REQUIRED)
target_link_libraries(tablepro PRIVATE hiredis::hiredis)

# Driver sources
set(TABLEPRO_SOURCES
    # ... existing ...
    src/driver/postgres_driver.cpp
    src/driver/mysql_driver.cpp
    src/driver/sqlite_driver.cpp
    src/driver/duckdb_driver.cpp
    src/driver/sqlserver_driver.cpp
    src/driver/clickhouse_driver.cpp
    src/driver/mongodb_driver.cpp
    src/driver/redis_driver.cpp
    src/driver/register_drivers.cpp
)
```

**Commit:**

```bash
git add CMakeLists.txt
git commit -m "build: Add all driver dependencies and sources"
```

---

## Acceptance Criteria

- [ ] All 8 drivers implement DatabaseDriver interface
- [ ] MySQL driver works with libmysql
- [ ] SQLite driver works with Qt SQL
- [ ] DuckDB driver works with embedded engine
- [ ] SQL Server driver works via ODBC
- [ ] ClickHouse driver connects and queries
- [ ] MongoDB driver handles BSON documents
- [ ] Redis driver executes commands
- [ ] All drivers registered with factory

---

**Phase 9 Complete.** Next: Phase 10 - SSH/SSL & Security