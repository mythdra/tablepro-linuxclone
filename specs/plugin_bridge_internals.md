# Plugin Bridge & Data Serialization

TablePro uses a dynamic plugin architecture to keep the core binary small and licensing clean (some databases require GPL drivers, which must be dynamically linked separate plugins). The Qt/C++ rewrite will achieve this using `QPluginLoader`.

## 1. The Interface Boundary

### `PluginDatabaseDriver`
The core application only knows about an abstract interface `PluginDatabaseDriver`. Every database driver (e.g., `PostgreSQLDriverPlugin`) implements this interface and gets compiled into a `.dylib`, `.so`, or `.dll`.

```cpp
// Example Qt Interface
class DatabaseDriverPlugin {
public:
    virtual ~DatabaseDriverPlugin() = default;
    virtual void connect(const QVariantMap& config) = 0;
    virtual void disconnect() = 0;
    virtual QVariantMap fetchMetadata(const QString& table) = 0;
    virtual QList<QVariantMap> executeQuery(const QString& sql) = 0;
};

#define DatabaseDriverPlugin_iid "com.tablepro.DatabaseDriverPlugin"
Q_DECLARE_INTERFACE(DatabaseDriverPlugin, DatabaseDriverPlugin_iid)
```

## 2. Parameter Passing (Config -> Plugin)
When establishing a connection, the core application cannot safely pass its internal memory models (`DatabaseConnection` structs) across the dynamic library boundary due to potential ABI incompatibilities or version mismatches.

Instead, the connection model must be serialized into a flat key-value map (`QVariantMap`), passed across the C-boundary, and unpacked by the plugin.

Keys include:
- `"host"` -> String
- `"port"` -> Integer
- `"database"` -> String
- `"username"` -> String
- `"password"` -> String
- `"ssl_mode"` -> String Enum ("require", "prefer", "disable")

## 3. Data Returning (Plugin -> Results Grid)
When executing `SELECT * FROM massive_table`, the plugin must return the result set back to the core UI efficiently. 

### Current Swift Approach
The C driver returns an array of dictionaries `[[String: Any]]`. The `Any` value handles type erasure (Int, String, Double, Blob, Dates). 

### Qt Rewrite Optimization Strategy
Passing thousands of small `QVariantMap` dictionaries across the boundary allocates massive amounts of RAM and fragments the heap. 
- **Better Approach**: The plugin should return data using a Columnar Matrix format (`QVector<QVariantList>` where each list represents an entire column of data, or `QAbstractTableModel` passing raw pointers). 
- If sticking to rows, use `QList<QVariantList>` where the outer index is the row, inner index is column. A separate `QList<QString>` passes the column names once to save memory. 

## 4. Generating SQL within the Plugin
The current `DataChangeManager` delegates dialect-specific statement generation to the core app for standard SQL (Postgres, MySQL, SQLite). 

However, for NoSQL or specialized engines (e.g., Redis, MongoDB), the driver plugin is entirely responsible for mutating data. If the user edits a cell in the Mongo grid, the core app passes the `RowChange` delta to the plugin's `generateStatements` method, which generates `db.collection.updateOne(...)` syntax and immediately executes it.
