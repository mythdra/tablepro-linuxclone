# Database Drivers Specification

TablePro supports a wide variety of databases using a plugin-based architecture. To maintain parity in the Qt/C++ rewrite, each supported database driver must be implemented as a separate dynamically loaded Qt Plugin (`QPluginLoader`) adhering to the `DatabaseDriver` C++ interface.

## Core Plugin Interface Requirements

Every database driver plugin must implement:
- **Connection Management:** Establish, test, and terminate connections.
- **Query Execution:** Synchronous/asynchronous query execution returning structured results (`[String: Any]` equivalents like `QVariantMap` or heavily optimized binary representations).
- **Schema Extractor:** Fetch database structure (schemas, tables, views, routines).
- **Table Metadata:** Fetch specific table definitions (columns, indexes, primary/foreign keys).
- **Data Mutation:** Generate dialect-specific `INSERT`, `UPDATE`, `DELETE` scripts.
- **EXPLAIN/AST support:** Generate or structure dialect-specific EXPLAIN query plans.

## 1. PostgreSQL (Relational)
- **C++ Dependency:** `libpq` / `libpqxx` (Standard C/C++ library for pg)
- **Features Required:**
  - TLS/SSL native support.
  - SSH Tunneling (either local port forwarding or proxy command).
  - Multi-schema support (`public`, etc.)
  - Fetching arrays and complex JSONB/JSON data.
  - Exposing specific server variables (`search_path`).
- **Dialect Specifics:** `LIMIT $1 OFFSET $2`, quotes using `"`, parameter binding via `$1`, `$2`. Requires understanding of `pg_catalog`.

## 2. MySQL / MariaDB (Relational)
- **C++ Dependency:** `libmysqlclient` / `mariadb-client`
- **Features Required:**
  - Standard User/Pass/Host/Port connect.
  - SSH / SSL capabilities.
- **Dialect Specifics:** `LIMIT $1 OFFSET $2`, quotes using ```, parameter binding via `?`. Uses `information_schema` heavily.

## 3. SQLite (Embedded Relational)
- **C++ Dependency:** `sqlite3` amalgamation / Qt's built-in `QSQLITE`.
- **Features Required:**
  - File-based connections (path picker). No networking configs needed unless custom PRAGMAs are executed on connection.
- **Dialect Specifics:** `LIMIT $1 OFFSET $2`, strings use `"`, basic typing (INTEGER, REAL, TEXT, BLOB).

## 4. Microsoft SQL Server (Relational)
- **C++ Dependency:** `FreeTDS` + `unixODBC` OR Microsoft ODBC Driver for SQL Server.
- **Features Required:**
  - NTLM Auth / Domain auth options.
  - Multiple databases per instance.
- **Dialect Specifics:** `TOP N`, quotes using `[ ]`, offset uses `OFFSET $1 ROWS FETCH NEXT $2 ROWS ONLY`. 

## 5. Oracle (Relational - Optional Download Plugin)
- **C++ Dependency:** `OCI` (Oracle Call Interface).
- **Features Required:**
  - Connect via SID or Service Name.
  - Complex Wallet/TNSNames config.
- **Dialect Specifics:** Quotes `""`, `FETCH FIRST N ROWS ONLY`, pseudo-columns like `ROWID`.

## 6. ClickHouse (OLAP)
- **C++ Dependency:** `clickhouse-cpp` (native protocol) OR HTTP/s REST.
- **Features Required:**
  - Handling massive column-oriented analytical datasets.
  - Session parameters.
- **Dialect Specifics:** EXPLAIN capabilities are massive, specific formats (AST, Pipeline, Plan). Quotes using ```.

## 7. MongoDB (NoSQL)
- **C++ Dependency:** `mongocxx` driver built on `libmongoc`.
- **Features Required:**
  - SSH Tunnels.
  - URI connection strings.
  - Auth databases (Auth Source), Replica Sets.
  - Outputting heavily nested BSON as formatted JSON text.
- **Dialect Specifics:** Converting JS-like queries (`db.collection.find(...)`) inside the SQL Editor into actual Mongo operations. Alternatively, providing true shell integration.

## 8. Redis (Key-Value)
- **C++ Dependency:** `hiredis` or `redis-plus-plus`.
- **Features Required:**
  - Viewing keys, values (Strings, Lists, Sets, Hashes, ZSets).
  - DB indexes (0-15 typical).
- **Dialect Specifics:** The command editor uses Redis commands (`GET foo`, `HSET bar b 2`) not SQL. 

## 9. DuckDB (Embedded OLAP)
- **C++ Dependency:** `libduckdb`.
- **Features Required:**
  - Handling heavy local `.duckdb` files, reading standard CSV/Parquet files fast.
- **Dialect Specifics:** Postgres-like dialect, analytic functions, extensions (`INSTALL httpfs; LOAD httpfs;`).

## Delivery Strategy
To reduce the core DMG size, drivers like Oracle, MS SQL Server, and MongoDB might be modular plugins downloaded on demand. The core app should ship with built-in SQLite, PostgreSQL, MySQL, and DuckDB.
