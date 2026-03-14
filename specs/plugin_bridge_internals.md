# Driver Integration (Go — replaces Plugin Bridge)

## Overview
In the Swift version, `PluginDriverAdapter` bridged the `PluginDatabaseDriver` (from dynamic `.tableplugin` bundles) to the app's internal `DatabaseDriver` protocol. In the Go rewrite, this bridge layer is **eliminated** — all drivers implement the same Go `DatabaseDriver` interface directly.

## Why No Bridge Needed
- Swift needed bridging because plugins were compiled separately as dynamic bundles
- Go drivers are compiled into the same binary — they share the same type system
- No serialization/deserialization overhead between plugin ↔ app
- No `dlopen`/`QPluginLoader` equivalent needed

## Driver Registration
```go
// internal/driver/registry.go
type DriverFactory func() DatabaseDriver

var registry = map[DatabaseType]DriverFactory{
    "postgres":   func() DatabaseDriver { return NewPostgresDriver() },
    "mysql":      func() DatabaseDriver { return NewMySQLDriver() },
    "sqlite":     func() DatabaseDriver { return NewSQLiteDriver() },
    "duckdb":     func() DatabaseDriver { return NewDuckDBDriver() },
    "mssql":      func() DatabaseDriver { return NewMSSQLDriver() },
    "clickhouse": func() DatabaseDriver { return NewClickHouseDriver() },
    "mongodb":    func() DatabaseDriver { return NewMongoDriver() },
    "redis":      func() DatabaseDriver { return NewRedisDriver() },
}
```

## Build Tags for Optional Drivers
Heavy drivers (Oracle, MongoDB) can be excluded from builds:
```go
//go:build oracle
// +build oracle

package driver

func init() {
    registry["oracle"] = func() DatabaseDriver { return NewOracleDriver() }
}
```
Build with: `go build -tags "oracle"` to include Oracle support.

## Driver Lifecycle
```go
// Creating a session
driver := NewDriver("postgres")
driver.Connect(config)
defer driver.Disconnect()

// All operations go directly through the interface
result, err := driver.Execute("SELECT * FROM users LIMIT 10")
tables, err := driver.FetchTables("public")
```

## Comparison with Swift Plugin System
| Aspect | Swift (Plugins) | Go (Interface) |
|---|---|---|
| Loading | Runtime `Bundle.load()` | Compile-time linking |
| Serialization | JSON/Codable across process boundary | Direct Go types |
| Error handling | Plugin crashes isolated | Same process (handle with `recover()`) |
| Distribution | Separate `.tableplugin` files | Single binary |
| Hot-reload | Possible (reload bundle) | Requires rebuild |
| Cross-platform | macOS only | macOS + Windows + Linux |
