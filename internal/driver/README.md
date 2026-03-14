# Database Drivers

This package provides database driver implementations for TablePro, supporting 8 different database systems.

## Supported Databases

| Database | Type | Driver Package | Connection Pool |
|----------|------|----------------|-----------------|
| PostgreSQL | SQL | `github.com/jackc/pgx/v5` | pgxpool |
| MySQL | SQL | `github.com/go-sql-driver/mysql` | database/sql |
| SQLite | SQL | `github.com/mattn/go-sqlite3` | database/sql |
| DuckDB | SQL | `github.com/marcboeker/go-duckdb` | database/sql |
| MSSQL | SQL | `github.com/microsoft/go-mssqldb` | database/sql |
| ClickHouse | SQL | `github.com/ClickHouse/clickhouse-go/v2` | custom |
| MongoDB | NoSQL | `go.mongodb.org/mongo-driver` | mongo.Client |
| Redis | NoSQL | `github.com/redis/go-redis/v9` | redis.Client |

## DatabaseDriver Interface

All drivers implement the `DatabaseDriver` interface defined in `driver.go`:

```go
type DatabaseDriver interface {
    Connect(ctx context.Context, config *ConnectionConfig) error
    Execute(ctx context.Context, query string, params ...any) (*Result, error)
    Query(ctx context.Context, query string, params ...any) (*Row, error)
    QueryContext(ctx context.Context, timeout time.Duration, query string, params ...any) (*Row, error)
    GetSchema(ctx context.Context) (*SchemaInfo, error)
    GetTables(ctx context.Context, schemaName string) ([]TableInfo, error)
    GetColumns(ctx context.Context, schemaName, tableName string) ([]ColumnInfo, error)
    GetIndexes(ctx context.Context, schemaName, tableName string) ([]IndexInfo, error)
    GetForeignKeys(ctx context.Context, schemaName, tableName string) ([]ForeignKeyInfo, error)
    Ping(ctx context.Context) error
    Close() error
    GetCapabilities() *DriverCapabilities
    GetDB() *sql.DB
}
```

## Driver-Specific Features

### PostgreSQL (`postgres/`)
- **Connection Pool**: pgxpool with configurable min/max connections
- **SSL**: Supports all SSL modes (disable, require, verify-ca, verify-full)
- **Schema Support**: Full schema support with namespaces
- **Advanced Types**: JSON, JSONB, ARRAY, UUID, inet, cidr, macaddr
- **Full Text Search**: Native FTS support
- **Window Functions**: Full support

### MySQL (`mysql/`)
- **Connection Pool**: Standard database/sql pool
- **SSL**: TLS/SSL connection support
- **Charset**: UTF8MB4 by default
- **Version Detection**: Auto-detects MySQL vs MariaDB
- **Stored Procedures**: Full support
- **Views**: Full support

### SQLite (`sqlite/`)
- **Embedded**: No external server required
- **In-Memory**: Supports in-memory databases
- **File-based**: Local file database
- **Types**: Dynamic typing with affinity

### DuckDB (`duckdb/`)
- **Embedded OLAP**: Analytical database
- **Parquet**: Native Parquet file support
- **Arrays**: Native array type
- **Structs**: Native struct type
- **NULL**: Full NULL handling

### MSSQL (`mssql/`)
- **Windows Auth**: NTLM/Kerberos support
- **SSL**: TLS encryption
- **Stored Procedures**: Full T-SQL support
- **Schemas**: Multiple schemas (dbo, etc.)

### ClickHouse (`clickhouse/`)
- **Columnar**: Column-oriented database
- **Distributed**: Distributed table support
- **Arrays**: Native array type
- **JSON**: Native JSON type
- **TTL**: Table TTL support

### MongoDB (`mongodb/`)
- **Document**: Document-based NoSQL
- **Aggregation**: Pipeline aggregation
- **Transactions**: Multi-document transactions
- **Sharding**: Sharded cluster support

### Redis (`redis/`)
- **Key-Value**: In-memory data store
- **Data Structures**: String, Hash, List, Set, Sorted Set
- **Pub/Sub**: Publish/Subscribe
- **Streams**: Redis Streams support

## Usage Examples

### Creating a PostgreSQL Driver

```go
import (
    "context"
    "time"
    
    "tablepro/internal/driver"
    "tablepro/internal/driver/postgres"
)

func main() {
    // Create driver instance
    pgDriver := postgres.NewPostgreSQLDriver()
    
    // Configure connection
    config := &driver.ConnectionConfig{
        Host:     "localhost",
        Port:     5432,
        Database: "testdb",
        Username: "user",
        Password: "password",
        SSLMode:  "disable",
        MaxOpenConnections: 25,
        MaxIdleConnections: 5,
        MaxConnectionLife:  5 * time.Minute,
    }
    
    // Connect
    ctx := context.Background()
    if err := pgDriver.Connect(ctx, config); err != nil {
        panic(err)
    }
    defer pgDriver.Close()
    
    // Execute query
    result, err := pgDriver.Execute(ctx, "SELECT * FROM users")
    // ...
}
```

### Creating a MySQL Driver

```go
import (
    "tablepro/internal/driver/mysql"
)

func main() {
    driver := mysql.New()
    
    config := &connection.DatabaseConnection{
        Host:     "localhost",
        Port:     3306,
        Database: "testdb",
        Username: "user",
    }
    
    if err := driver.Connect(ctx, config, "password"); err != nil {
        panic(err)
    }
}
```

### Creating a MongoDB Driver

```go
import (
    "tablepro/internal/driver/mongodb"
)

func main() {
    driver := mongodb.NewMongoDBDriver()
    
    config := &driver.ConnectionConfig{
        Host:     "localhost",
        Port:     27017,
        Database: "testdb",
        Username: "user",
        Password: "password",
    }
    
    if err := driver.Connect(ctx, config); err != nil {
        panic(err)
    }
}
```

## Type Mappings

All drivers support type mapping from database types to Go types. See `types.go` for the complete mapping table.

```go
import "tablepro/internal/driver"

// Get type mapping
mapping := driver.GetDataTypeMapping(driver.DatabaseTypePostgreSQL, "varchar")
if mapping != nil {
    fmt.Printf("Go type: %s\n", mapping.GoType)
    fmt.Printf("Is string: %v\n", mapping.IsString)
}

// Check if type is numeric
isNum := driver.IsNumericType(driver.DatabaseTypePostgreSQL, "int4")
```

## Driver Factory

Use the factory to create drivers dynamically:

```go
import "tablepro/internal/driver"

func createDriver(dbType driver.DatabaseType) (driver.DatabaseDriver, error) {
    if !driver.IsSupported(dbType) {
        return nil, fmt.Errorf("unsupported database: %s", dbType)
    }
    return driver.NewDriver(dbType)
}
```

## Verification Checklist

- [x] PostgreSQL - Connection pool, schema, type mapping
- [x] MySQL - Connection pool, schema, type mapping
- [x] SQLite - Connection pool, schema, type mapping
- [x] DuckDB - Connection pool, schema, type mapping
- [x] MSSQL - Connection pool, schema, type mapping
- [x] ClickHouse - Connection pool, schema, type mapping
- [x] MongoDB - Connection, schema (collections), type mapping
- [x] Redis - Connection, schema (keys), type mapping

## Testing

Run driver tests:

```bash
# Run all driver tests
go test ./internal/driver/...

# Run specific driver tests
go test ./internal/driver/postgres/...
go test ./internal/driver/mysql/...

# Run integration tests (requires Docker)
go test -tags=integration ./internal/driver/...
```

## Dependencies

All driver dependencies are in `go.mod`:

- `github.com/jackc/pgx/v5` - PostgreSQL
- `github.com/go-sql-driver/mysql` - MySQL
- `github.com/mattn/go-sqlite3` - SQLite
- `github.com/marcboeker/go-duckdb` - DuckDB
- `github.com/microsoft/go-mssqldb` - MSSQL
- `github.com/ClickHouse/clickhouse-go/v2` - ClickHouse
- `go.mongodb.org/mongo-driver` - MongoDB
- `github.com/redis/go-redis/v9` - Redis

## Notes

- All drivers use context.Context for cancellation and timeouts
- Connection pooling is configured via ConnectionConfig
- Passwords should be stored in OS Keychain, not in config
- All drivers implement proper error wrapping