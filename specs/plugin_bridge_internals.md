# Plugin Bridge Internals (C++20 + Qt 6)

## Overview
In the Swift version, the plugin bridge (`PluginBridgeService`) serialized method calls to JSON, sent them over stdin/stdout to plugin processes, and deserialized responses. The Go version eliminated this bridge entirely by compiling drivers into the same binary.

The C++ + Qt architecture **also eliminates the bridge layer**. All drivers are compiled into the same binary and communicate via direct C++ method calls and Qt signals/slots — no serialization, no IPC, no process boundaries.

## Why No Bridge Needed

| Concern | Swift Plugin Bridge | C++20 + Qt |
|---------|---------------------|------------|
| Method dispatch | JSON-RPC over stdin/stdout | Direct virtual method calls |
| Data serialization | Codable → JSON → Data | Direct C++ types |
| Error handling | NSError marshaling | C++ exceptions, Qt error signals |
| Plugin loading | `Bundle.load()` with sandbox | Compile-time linking |
| Crash isolation | Separate process = isolated crash | Same process (use try/catch) |
| Memory | ARC + bridging | RAII + smart pointers |
| Hot reload | Possible via `dlopen` | Requires rebuild |

## Driver Interface (Direct C++)

```cpp
// driver/driver.hpp
class DatabaseDriver : public QObject {
    Q_OBJECT

public:
    // No serialization - direct C++ types
    virtual QFuture<QueryResult> execute(const QString& sql) = 0;
    virtual QFuture<QList<TableInfo>> fetchTables(const QString& schema) = 0;

    // Direct Qt signal emission - no event marshaling
signals:
    void connected();
    void connectionError(const QString& error);
    void dataRefreshed();
};
```

## Comparison: Swift Bridge vs C++ Direct

### Swift Plugin Bridge (Removed)
```swift
// OLD: Swift → Plugin Bridge
class PluginBridgeService {
    func executeQuery(_ query: String, on session: PluginSession) async throws -> QueryResult {
        // 1. Serialize to JSON
        let request = PluginRequest(method: "execute", params: ["query": query])
        let jsonData = try JSONEncoder().encode(request)

        // 2. Send over stdin
        pluginProcess.standardInput?.write(jsonData)

        // 3. Wait for response on stdout
        let response = try await readResponse()

        // 4. Deserialize
        return try JSONDecoder().decode(QueryResult.self, from: response.data)
    }
}
```

### C++ Direct Call (Current)
```cpp
// C++: Direct method call - no bridge
auto result = co_await driver->execute("SELECT * FROM users");
// Result is directly accessible - no deserialization
```

## Qt Signal/Slot Communication

Qt's signal/slot system replaces both the Swift plugin event stream AND the Go channel-based event emission:

```cpp
// Driver emits signal directly
class PostgresDriver : public DatabaseDriver {
    Q_OBJECT

public:
    QFuture<QueryResult> execute(const QString& sql) override {
        return QtConcurrent::run([this, sql]() {
            // Execute query in thread pool
            PGresult* result = PQexec(m_conn, sql.toUtf8().constData());

            if (PQresultStatus(result) == PGRES_TUPLES_OK) {
                emit dataRefreshed();  // Direct signal emission
                return parseResult(result);
            } else {
                emit connectionError(QString::fromUtf8(PQerrorMessage(m_conn)));
                throw QueryExecutionError(PQerrorMessage(m_conn));
            }
        });
    }
};

// Frontend connects via lambda
connect(driver.get(), &DatabaseDriver::dataRefreshed,
        this, [this]() {
    m_statusBar->showMessage("Data refreshed", 3000);
    m_tableView->refresh();
});
```

## No JSON Overhead

| Operation | Swift Bridge | Go | C++ + Qt |
|-----------|--------------|-----|----------|
| Method call | JSON serialize → IPC → deserialize | Direct Go call | Direct C++ call |
| Data transfer | Codable encoding/decoding | Struct copy | QVariant/struct copy |
| Error propagation | NSError → JSON → NSError | error return | Exception/signal |
| Event stream | NotificationCenter → JSON | Go channel → runtime.EventsEmit | Qt signal → slot |

## Memory and Performance

```
┌─────────────────────────────────────────────────────────┐
│              DATA FLOW COMPARISON                       │
└─────────────────────────────────────────────────────────┘

Swift Plugin Architecture:
┌──────────┐    JSON encode   ┌──────────┐    decode    ┌──────────┐
│  Swift   │─────────────────▶│  Bridge  │─────────────▶│  Plugin  │
│  Frontend│◀─────────────────│  Service │◀─────────────│  Driver  │
└──────────┘    JSON decode   └──────────┐    encode    └──────────┘

C++ + Qt Architecture:
┌──────────┐    direct call   ┌──────────┐
│  Qt UI   │─────────────────▶│  Driver  │
│          │◀─────────────────│          │
└──────────┘   signal/slot    └──────────┘

Overhead eliminated:
• No JSON serialization (saves ~1-5ms per call)
• No IPC latency (saves ~0.1-1ms per call)
• No string parsing (saves CPU cycles)
• No encoding/decoding errors (more reliable)
```

## Exception Handling

Without a bridge, exceptions propagate directly:

```cpp
// Driver throws exception
class PostgresDriver : public DatabaseDriver {
    QFuture<QueryResult> execute(const QString& sql) override {
        try {
            // Direct C++ call - can throw
            if (PQstatus(m_conn) != CONNECTION_OK) {
                throw ConnectionLostException("Database connection lost");
            }

            PGresult* result = PQexec(m_conn, sql.toUtf8().constData());

            if (PQresultStatus(result) != PGRES_TUPLES_OK) {
                throw QueryExecutionError(PQerrorMessage(m_conn));
            }

            return parseResult(result);

        } catch (const std::exception& e) {
            emit connectionError(QString::fromUtf8(e.what()));
            throw;  // Re-throw for caller handling
        }
    }
};

// Caller catches
try {
    auto result = co_await driver->execute(sql);
    // Process result
} catch (const QueryExecutionError& e) {
    QMessageBox::critical(this, "Query Error", e.what());
}
```

## RAII Resource Management

C++ eliminates the need for bridge-based cleanup:

```cpp
// OLD Swift: Bridge must send "close" message to plugin
await bridge.closeSession(sessionId)

// C++: RAII automatic cleanup
class PostgresDriver : public DatabaseDriver {
    std::unique_ptr<PGconn, PqConnDeleter> m_conn;

    ~PostgresDriver() override {
        // Automatic cleanup via smart pointer
        // No explicit "disconnect" call needed
    }
};

// When driver goes out of scope, connection is automatically closed
auto driver = std::make_unique<PostgresDriver>();
// ... use driver ...
// driver.reset() or scope exit = automatic disconnect
```

## Optional: Qt Plugin Loader (For Extensions)

If third-party extensions are needed, Qt provides `QPluginLoader`:

```cpp
// core/PluginManager.hpp
class PluginManager : public QObject {
    Q_OBJECT

public:
    static PluginManager* instance();

    // Load a driver plugin from .dylib/.so/.dll
    QFuture<bool> loadDriverPlugin(const QString& pluginPath);
    DatabaseDriver* getDriver(const QString& pluginId);
    QStringList availablePlugins() const;

private:
    QMap<QString, QPluginLoader*> m_pluginLoaders;
    QMap<QString, DatabaseDriver*> m_pluginDrivers;
};
```

```cpp
// Example plugin structure
// mydb-driver/mydb_driver_plugin.cpp
#include <QtPlugins/QPluginLoader>
#include <QObject>

class MyDbDriverPlugin : public QObject, public DatabaseDriver {
    Q_OBJECT
    Q_PLUGIN_METADATA(IID "com.tablepro.driver" FILE "plugin.json")
    Q_INTERFACES(DatabaseDriver)

public:
    QObject* create(const QString& key) override {
        if (key == "database_driver") {
            return new MyDbDriver();
        }
        return nullptr;
    }
};
```

**However**, this is generally not recommended because:
- Compile-time linking provides better type safety
- No runtime plugin loading failures
- Easier distribution (single binary)
- Better performance (no dlopen overhead)

## Conclusion

The C++ architecture is **simpler** than both Swift and Go:
- **vs Swift**: No bridge, no JSON serialization, no IPC
- **vs Go**: Same direct-call model, but with Qt signals for events
- **Result**: Cleaner code, better performance, easier debugging
