# Database Drivers Specification (Go)

Each database driver implements the `DatabaseDriver` interface (defined in `plugins.md`). Below are the Go-specific implementation details per database.

## 1. PostgreSQL
- **Go Package**: `github.com/jackc/pgx/v5` (native, no CGo required)
- **Features**: TLS/SSL, SSH tunneling, multi-schema, JSONB, arrays, LISTEN/NOTIFY
- **Dialect**: `LIMIT $1 OFFSET $2`, identifiers quoted with `"`, params `$1, $2, ...`
- **Schema Queries**: `pg_catalog`, `information_schema`
- **Connection Pooling**: `pgxpool.Pool` for concurrent queries

## 2. MySQL / MariaDB
- **Go Package**: `github.com/go-sql-driver/mysql` (pure Go)
- **Features**: User/Pass/Host/Port, SSH tunneling, SSL, multiple databases
- **Dialect**: `` LIMIT ? OFFSET ? ``, identifiers quoted with `` ` ``, params `?`
- **Schema Queries**: `information_schema`

## 3. SQLite
- **Go Package**: `github.com/mattn/go-sqlite3` (CGo, uses sqlite3 amalgamation)
- **Alternative**: `modernc.org/sqlite` (pure Go, no CGo — recommended for easier cross-compile)
- **Features**: File-based, no network, WAL mode, PRAGMAs
- **Dialect**: `LIMIT ? OFFSET ?`, basic types (INTEGER, REAL, TEXT, BLOB)

## 4. Microsoft SQL Server
- **Go Package**: `github.com/microsoft/go-mssqldb` (pure Go)
- **Features**: SQL Server Auth, Windows Auth (NTLM), multiple databases
- **Dialect**: `TOP N`, identifiers `[name]`, offset `OFFSET ? ROWS FETCH NEXT ? ROWS ONLY`

## 5. Oracle (Optional — build tag `oracle`)
- **Go Package**: `github.com/godror/godror` (requires Oracle Instant Client)
- **Features**: SID/Service Name, TNS, Wallet
- **Dialect**: `FETCH FIRST N ROWS ONLY`, identifiers `"name"`, `ROWID` pseudo-column

## 6. ClickHouse
- **Go Package**: `github.com/ClickHouse/clickhouse-go/v2` (native protocol, pure Go)
- **Features**: Column-oriented analytics, batch inserts, compression
- **Dialect**: `` LIMIT ? OFFSET ? ``, backtick identifiers

## 7. MongoDB
- **Go Package**: `go.mongodb.org/mongo-driver` (official driver)
- **Features**: URI strings, SSH tunneling, Auth Source, Replica Sets, BSON→JSON
- **Special Handling**: MongoDB doesn't use SQL. The editor sends JSON-like filter objects. The driver translates `db.collection.find({...})` syntax into `bson.M` queries internally.

## 8. Redis
- **Go Package**: `github.com/redis/go-redis/v9`
- **Features**: Key browsing (Strings, Lists, Sets, Hashes, Sorted Sets), DB index selection (0-15)
- **Special Handling**: Editor uses Redis commands (`GET`, `HSET`, `KEYS *`), not SQL. Query results are presented as key-value pairs.

## 9. DuckDB
- **Go Package**: `github.com/marcboeker/go-duckdb` (CGo wrapper around libduckdb)
- **Features**: Local `.duckdb` files, CSV/Parquet reading, extensions
- **Dialect**: PostgreSQL-compatible syntax

## Delivery Strategy
All drivers compile into the single binary by default (except Oracle, which requires a build tag due to the Oracle Instant Client dependency). Total binary size estimate: ~20-30MB.
