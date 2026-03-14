# Phase 4: Database Drivers Tasks

Implementation checklist for Phase 4 - Database Drivers (60 tasks)

---

## 1. Driver Interface

- [ ] 1.1 Define DatabaseDriver interface (Connect, Execute, GetSchema, GetTables, GetColumns, Close)
- [ ] 1.2 Define Row struct with Data map and ColumnNames slice
- [ ] 1.3 Define ColumnInfo struct (Name, Type, Nullable, Default, IsPrimaryKey)
- [ ] 1.4 Define SchemaInfo struct (Tables, Views, Procedures, Functions)
- [ ] 1.5 Define DriverCapabilities struct (Features, MaxConnections, MaxQueryTime)
- [ ] 1.6 Define TableInfo struct (Name, Schema, Type, Comment)
- [ ] 1.7 Define IndexInfo struct (Name, Columns, IsUnique, IsPrimary)
- [ ] 1.8 Define ForeignKeyInfo struct (Name, Columns, ReferencedTable, ReferencedColumns)
- [ ] 1.9 Create driver factory function NewDriver(type) DatabaseDriver

---

## 2. PostgreSQL Driver (pgx)

- [ ] 2.1 Add github.com/jackc/pgx/v5 dependency
- [ ] 2.2 Create PostgreSQLDriver struct with *pgx.Conn
- [ ] 2.3 Implement Connect(ctx, config) with 30s timeout
- [ ] 2.4 Implement Execute(ctx, query) returning []Row
- [ ] 2.5 Implement GetSchema() returning SchemaInfo
- [ ] 2.6 Implement GetTables(schema) returning []TableInfo
- [ ] 2.7 Implement GetColumns(table) returning []ColumnInfo
- [ ] 2.8 Implement GetIndexes(table) returning []IndexInfo
- [ ] 2.9 Implement GetForeignKeys(table) returning []ForeignKeyInfo
- [ ] 2.10 Handle PostgreSQL-specific types (ARRAY, JSONB, UUID, TIMESTAMPTZ)
- [ ] 2.11 Implement connection pooling with pgxpool
- [ ] 2.12 Implement Close() with proper cleanup
- [ ] 2.13 Write unit tests for PostgreSQL driver
- [ ] 2.14 Write integration tests with Docker PostgreSQL

---

## 3. MySQL Driver

- [ ] 3.1 Add github.com/go-sql-driver/mysql dependency
- [ ] 3.2 Create MySQLDriver struct with *sql.DB
- [ ] 3.3 Implement Connect(ctx, config) with charset utf8mb4
- [ ] 3.4 Implement Execute(ctx, query) returning []Row
- [ ] 3.5 Implement GetSchema() for MySQL
- [ ] 3.6 Implement GetTables(schema)
- [ ] 3.7 Implement GetColumns(table)
- [ ] 3.8 Handle MySQL-specific types (ENUM, SET, BLOB variants)
- [ ] 3.9 Support MariaDB variants (detect version)
- [ ] 3.10 Implement connection pooling
- [ ] 3.11 Implement Close()
- [ ] 3.12 Write unit tests
- [ ] 3.13 Write integration tests with Docker MySQL

---

## 4. SQLite Driver

- [ ] 4.1 Add github.com/mattn/go-sqlite3 dependency
- [ ] 4.2 Create SQLiteDriver struct with *sql.DB
- [ ] 4.3 Implement Connect(ctx, config) with file path
- [ ] 4.4 Support :memory: for in-memory databases
- [ ] 4.5 Implement Execute(ctx, query)
- [ ] 4.6 Implement GetSchema()
- [ ] 4.7 Implement GetTables(schema)
- [ ] 4.8 Implement GetColumns(table)
- [ ] 4.9 Enable foreign keys (PRAGMA foreign_keys = ON)
- [ ] 4.10 Handle SQLite-specific features (affinity types)
- [ ] 4.11 Implement Close()
- [ ] 4.12 Write unit tests

---

## 5. DuckDB Driver

- [ ] 5.1 Add github.com/marcboeker/go-duckdb dependency
- [ ] 5.2 Create DuckDBDriver struct with *sql.DB
- [ ] 5.3 Implement Connect(ctx, config) for analytical workloads
- [ ] 5.4 Implement Execute(ctx, query) with batch support
- [ ] 5.5 Implement GetSchema()
- [ ] 5.6 Implement GetTables(schema)
- [ ] 5.7 Implement GetColumns(table)
- [ ] 5.8 Support columnar storage format
- [ ] 5.9 Implement Close()
- [ ] 5.10 Write unit tests

---

## 6. MSSQL Driver

- [ ] 6.1 Add github.com/microsoft/go-mssqldb dependency
- [ ] 6.2 Create MSSQLDriver struct with *sql.DB
- [ ] 6.3 Implement Connect(ctx, config) with TDS protocol
- [ ] 6.4 Implement Execute(ctx, query)
- [ ] 6.5 Implement GetSchema()
- [ ] 6.6 Implement GetTables(schema)
- [ ] 6.7 Implement GetColumns(table)
- [ ] 6.8 Support Windows authentication
- [ ] 6.9 Handle SQL Server-specific types (DATETIME2, NVARCHAR, etc.)
- [ ] 6.10 Implement Close()
- [ ] 6.11 Write unit tests

---

## 7. ClickHouse Driver

- [ ] 7.1 Add github.com/ClickHouse/clickhouse-go/v2 dependency
- [ ] 7.2 Create ClickHouseDriver struct
- [ ] 7.3 Implement Connect(ctx, config) with native protocol
- [ ] 7.4 Implement Execute(ctx, query)
- [ ] 7.5 Handle Array types mapping to Go slices
- [ ] 7.6 Implement GetSchema()
- [ ] 7.7 Implement GetTables(schema)
- [ ] 7.8 Implement GetColumns(table)
- [ ] 7.9 Support columnar operations
- [ ] 7.10 Implement Close()
- [ ] 7.11 Write unit tests

---

## 8. MongoDB Driver

- [ ] 8.1 Add go.mongodb.org/mongo-driver dependency
- [ ] 8.2 Create MongoDBDriver struct with *mongo.Client
- [ ] 8.3 Implement Connect(ctx, config) with auth database
- [ ] 8.4 Implement Execute(ctx, query) for find() operations
- [ ] 8.5 Implement GetSchema() returning databases
- [ ] 8.6 Implement GetTables() returning collections
- [ ] 8.7 Implement GetColumns(collection) returning field schema
- [ ] 8.8 Handle document-based results as JSON
- [ ] 8.9 Implement Close()
- [ ] 8.10 Write unit tests

---

## 9. Redis Driver

- [ ] 9.1 Add github.com/redis/go-redis/v9 dependency
- [ ] 9.2 Create RedisDriver struct with *redis.Client
- [ ] 9.3 Implement Connect(ctx, config) with DB index
- [ ] 9.4 Implement Execute(ctx, command) for Redis commands
- [ ] 9.5 Implement GetSchema()
- [ ] 9.6 Implement GetTables() returning keys (with KEYS/SCAN)
- [ ] 9.7 Implement GetColumns(key) returning type and TTL
- [ ] 9.8 Handle key-value results
- [ ] 9.9 Implement Close()
- [ ] 9.10 Write unit tests

---

## 10. Type Mapping

- [ ] 10.1 Create TypeMapper struct
- [ ] 10.2 Implement MapDatabaseType(dbType, dbType) → GoType
- [ ] 10.3 Implement MapGoType(goType) → TypeScriptType
- [ ] 10.4 Handle PostgreSQL types (TIMESTAMPTZ, UUID, JSONB, ARRAY)
- [ ] 10.5 Handle MySQL types (ENUM, SET, BLOB, DATETIME)
- [ ] 10.6 Handle SQLite types (NULL, INTEGER, REAL, TEXT, BLOB)
- [ ] 10.7 Handle SQL Server types (DATETIME2, NVARCHAR, BIT)
- [ ] 10.8 Handle ClickHouse types (Array, Map, Tuple)
- [ ] 10.9 Handle MongoDB types (ObjectId, Date, Array, Object)
- [ ] 10.10 Handle Redis types (String, List, Set, Hash, ZSet)
- [ ] 10.11 Implement fallback for unknown types (interface{})
- [ ] 10.12 Write comprehensive type mapping tests

---

## 11. Schema Introspection

- [ ] 11.1 Create SchemaIntrospector interface
- [ ] 11.2 Implement for PostgreSQL (query pg_catalog)
- [ ] 11.3 Implement for MySQL (query information_schema)
- [ ] 11.4 Implement for SQLite (query sqlite_master)
- [ ] 11.5 Implement for DuckDB (query information_schema)
- [ ] 11.6 Implement for MSSQL (query sys.tables)
- [ ] 11.7 Implement for ClickHouse (query system.tables)
- [ ] 11.8 Implement for MongoDB (listCollections)
- [ ] 11.9 Implement for Redis (SCAN/TYPE commands)
- [ ] 11.10 Create unified SchemaInfo response format

---

## Verification Checklist

Run these commands to verify Phase 4 completion:

```bash
# Build app
go build ./...

# Test drivers
go test ./internal/driver/...

# Test type mapping
go test ./internal/driver/types/...

# Integration tests
docker-compose up -d postgres mysql
go test -tags=integration ./internal/driver/...
```

---

## Acceptance Criteria

- [ ] All 8 drivers implement DatabaseDriver interface
- [ ] Connection pooling working for SQL drivers
- [ ] Schema metadata retrieval working
- [ ] Type mapping correct for each database
- [ ] 80%+ test coverage per driver
- [ ] Integration tests passing

---

## Dependencies

← [Phase 3: Connection Management](../archive/2026-03-14-phase-3-connection-management/)  
→ [Phase 5: Query Execution](../phase-5-query-execution/)

---

## Notes

- Drivers should be implemented in priority order: PostgreSQL → MySQL → SQLite → DuckDB → MSSQL → ClickHouse → MongoDB → Redis
- Each driver needs both unit tests and integration tests
- Use Docker Compose for integration test databases
- Type mapping is critical for frontend compatibility
