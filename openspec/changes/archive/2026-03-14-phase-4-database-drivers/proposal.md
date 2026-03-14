## Why

Phase 3 đã hoàn thành Connection Management với đầy đủ data models, CRUD operations, SSH/SSL support, và UI. Phase 4 tiếp tục xây dựng Database Drivers - thành phần cốt lõi để kết nối và thực thi queries trên 8 database systems khác nhau (PostgreSQL, MySQL, SQLite, DuckDB, MSSQL, ClickHouse, MongoDB, Redis). Đây là bước quan trọng để biến TablePro thành multi-database client thực thụ.

## What Changes

- **Driver Interface**: DatabaseDriver interface với methods: Connect, Execute, GetSchema, GetTables, GetColumns, Close
- **PostgreSQL Driver**: pgx/v5-based driver với connection pooling, schema metadata
- **MySQL Driver**: go-sql-driver/mysql với charset config, ENUM/SET support
- **SQLite Driver**: mattn/go-sqlite3 cho file-based databases
- **DuckDB Driver**: go-duckdb cho analytical workloads
- **Additional Drivers**: MSSQL, ClickHouse, MongoDB, Redis drivers
- **Type Mapping**: Cross-database type conversion utilities
- **Schema Metadata**: Unified schema introspection API

## Capabilities

### New Capabilities
- `driver-interface`: DatabaseDriver interface với统一的 methods và types
- `postgresql-driver`: pgx/v5 driver với connection pooling và metadata
- `mysql-driver`: MySQL/MariaDB driver với proper type mapping
- `sqlite-driver`: SQLite driver cho file-based databases
- `duckdb-driver`: DuckDB driver cho analytical queries
- `mssql-driver`: SQL Server driver
- `clickhouse-driver`: ClickHouse columnar database driver
- `mongodb-driver`: MongoDB NoSQL driver
- `redis-driver`: Redis key-value store driver
- `type-mapping`: Cross-database type conversion utilities
- `schema-introspection`: Unified schema metadata API

### Modified Capabilities
- (None - đây là new drivers, không modify existing capabilities)

## Impact

- **Code**: Tạo packages: internal/driver/ với 8 driver implementations (~3000 LOC/driver)
- **Dependencies**: 
  - Go: pgx/v5, go-sql-driver/mysql, go-sqlite3, go-duckdb, go-mssqldb, clickhouse-go, mongo-go-driver, go-redis
  - Frontend: Schema metadata UI components
- **Systems**: Connection sessions, query execution pipeline
- **Timeline**: 4-5 weeks cho complete implementation
- **Downstream**: Phase 5 (Query Execution), Phase 7 (Data Grid) phụ thuộc vào drivers
