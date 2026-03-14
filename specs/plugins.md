# TablePro Driver System (Go + Wails)

## Overview
In the Go rewrite, the plugin concept is replaced by a **Go interface** pattern. Each database driver implements the `DatabaseDriver` interface. Unlike the Swift version's runtime `.tableplugin` bundles, Go drivers are **compiled-in** at build time — Go does not have a production-grade dynamic plugin mechanism for cross-platform use.

## DatabaseDriver Interface
```go
type DatabaseDriver interface {
    // Connection
    Connect(config ConnectionConfig) error
    Disconnect() error
    TestConnection(config ConnectionConfig) error
    Ping() error

    // Query Execution
    Execute(query string) (*QueryResult, error)
    ExecuteWithParams(query string, params []any) (*QueryResult, error)

    // Schema Inspection
    FetchDatabases() ([]DatabaseInfo, error)
    FetchSchemas(database string) ([]SchemaInfo, error)
    FetchTables(schema string) ([]TableInfo, error)
    FetchViews(schema string) ([]TableInfo, error)
    FetchRoutines(schema string) ([]RoutineInfo, error)
    FetchColumns(schema, table string) ([]ColumnMetadata, error)
    FetchIndexes(schema, table string) ([]IndexMetadata, error)
    FetchForeignKeys(schema, table string) ([]ForeignKeyMetadata, error)
    FetchTableDDL(schema, table string) (string, error)
    FetchTableRowCount(schema, table string) (int64, error)

    // Transactions
    BeginTransaction() error
    CommitTransaction() error
    RollbackTransaction() error

    // Dialect
    DialectInfo() DialectInfo
    ForeignKeyDisableStatements() []string
    ForeignKeyEnableStatements() []string
}
```

## Driver Registration
```go
// internal/driver/registry.go
var drivers = map[DatabaseType]func() DatabaseDriver{
    Postgres:   func() DatabaseDriver { return &PostgresDriver{} },
    MySQL:      func() DatabaseDriver { return &MySQLDriver{} },
    SQLite:     func() DatabaseDriver { return &SQLiteDriver{} },
    DuckDB:     func() DatabaseDriver { return &DuckDBDriver{} },
    MSSQL:      func() DatabaseDriver { return &MSSQLDriver{} },
    ClickHouse: func() DatabaseDriver { return &ClickHouseDriver{} },
    MongoDB:    func() DatabaseDriver { return &MongoDriver{} },
    Redis:      func() DatabaseDriver { return &RedisDriver{} },
}

func NewDriver(dbType DatabaseType) (DatabaseDriver, error) {
    factory, ok := drivers[dbType]
    if !ok {
        return nil, fmt.Errorf("unsupported database type: %s", dbType)
    }
    return factory(), nil
}
```

## DatabaseManager (Go Service)
Replaces Swift's `DatabaseManager` actor. Bound to Wails frontend.
```go
type DatabaseManager struct {
    ctx      context.Context
    sessions map[uuid.UUID]*ConnectionSession
    mu       sync.RWMutex
}
```
Responsibilities:
- Maintain active `ConnectionSession` map (driver + SSH tunnel + health)
- Background goroutine pings every 30s (`driver.Ping()`)
- SSH tunnel lifecycle via `golang.org/x/crypto/ssh`
- Emit Wails events: `connection:status`, `connection:error`, `data:refresh`

## Export/Import Format Handlers
Instead of dynamic plugins, export/import formats use the same interface pattern:
```go
type ExportFormat interface {
    ID() string
    Name() string
    Extension() string
    Export(source DataSource, writer io.Writer, options ExportOptions) error
}

type ImportFormat interface {
    ID() string
    Name() string
    Extensions() []string
    Import(reader io.Reader, sink DataSink, progress *ImportProgress) (*ImportResult, error)
}
```
Built-in formats: CSV, JSON, SQL, XLSX, Markdown.

## Advantages over Swift Plugin System
- **No runtime loading failures** — all drivers compile-checked
- **Single binary deployment** — no `.tableplugin` bundles to distribute
- **Cross-platform** — same binary runs on macOS, Windows, Linux
- **Build tags** for optional drivers: `go build -tags "oracle"` to include heavy drivers
