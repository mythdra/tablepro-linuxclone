## Context

Phase 3 đã hoàn thành với Connection Management, ConnectionManager, và TestConnection capability. Hiện tại đã có internal/connection package với data models và internal/ssh cho tunneling. Phase 4 xây dựng Database Drivers để thực thi queries và lấy schema metadata.

**Ràng buộc**:
- Must use database/sql interface cho SQL drivers
- PostgreSQL driver phải dùng pgx/v5 (không dùng lib/pq)
- Context timeout cho tất cả operations
- Connection pooling required cho production drivers
- Timeline: 4-5 weeks

## Goals / Non-Goals

**Goals:**
- DatabaseDriver interface với Connect, Execute, GetSchema, GetTables, GetColumns, Close
- 8 drivers: PostgreSQL, MySQL, SQLite, DuckDB, MSSQL, ClickHouse, MongoDB, Redis
- Type mapping utilities cho cross-database compatibility
- Schema introspection API unified
- Connection pooling với configurable limits
- Unit tests với 80%+ coverage

**Non-Goals:**
- Query builder (Phase 5)
- Result pagination (Phase 5)
- Data mutation/UPDATE/INSERT (Phase 7)
- Driver-specific advanced features (chỉ basic CRUD)

## Decisions

### 1. Interface: database/sql cho SQL, Native cho NoSQL
**Rationale**: Standard library interface, consistent API
**Alternatives**: 
- Custom interface: Reinventing the wheel
- sqlx: Nice nhưng thêm dependency

### 2. PostgreSQL: pgx/v5 over lib/pq
**Rationale**: Better performance, active maintenance, v5 có improved type support
**Alternatives**: lib/pq - deprecated, slower

### 3. MySQL: go-sql-driver/mysql
**Rationale**: Official driver, most popular, well-maintained
**Alternatives**: go-mysql-driver - ít features hơn

### 4. SQLite: mattn/go-sqlite3
**Rationale**: CGO-based, full SQLite features, standard choice
**Alternatives**: modernc.org/sqlite - pure Go nhưng slower

### 5. DuckDB: go-duckdb
**Rationale**: Official driver, analytical workloads support
**Alternatives**: None (only official driver)

### 6. NoSQL: Native APIs
**Rationale**: mongo-go-driver và go-redis có APIs tốt, không cần sql wrapper
**Alternatives**: sql-like wrappers - mất tính năng native

### 7. Schema Metadata: Unified Interface
**Rationale**: Frontend không cần biết database type
**Pattern**: Each driver implements GetSchema() trả về standardized SchemaInfo struct

### 8. Type Mapping: Common Type System
**Rationale**: Frontend cần consistent types
**Pattern**: Map tất cả database types → Go types → TypeScript types

## Risks / Trade-offs

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Driver bugs in edge cases | High | Medium | Comprehensive unit tests, integration tests với Docker |
| Connection leaks | High | Medium | Context timeouts, proper Close() in defer |
| Type mapping errors | Medium | Medium | Test matrix với sample data |
| Performance issues | Medium | Low | Benchmark tests, connection pooling |
| Dependency conflicts | Low | Low | Pin versions, regular go mod tidy |
| NoSQL driver API changes | Low | Low | Wrapper layer, version pinning |

## Migration Plan

Not applicable - greenfield development.

## Open Questions

1. **Driver priority**: Implement drivers theo thứ tự nào? (recommend: PostgreSQL → MySQL → SQLite → DuckDB → others)
2. **Connection pooling**: Pool size config per connection hay global? (recommend: per connection type)
3. **Error messages**: Translate database-specific errors to user-friendly messages? (recommend: yes, Phase 5)
