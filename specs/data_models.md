# Data Models Specification

To ensure state consistency, the C++ rewrite needs to emulate the core Swift Data Models. Swift structs map nicely to standard C++ structs, while Swift classes (`@Observable`) map to `QObject` classes utilizing Signals and Slots (`Q_PROPERTY`).

## 1. Connection Models

### `DatabaseConnection`
Represents the persistent user connection profile.
```cpp
// Stored in QtKeychain & QSettings via Serialization
struct DatabaseConnection {
    QUuid id;
    QString name;
    DatabaseType type; // ENUM: Postgres, MySQL, etc.
    QString group;     // Optional grouping name
    QString colorTag;  // Enum string: "red", "blue"
    
    // Core Params
    QString host;
    int port;
    QString database;
    QString username;
    bool hasPasswordInKeychain; // Actual password stored securely
    QString localFilePath; // For SQLite/DuckDB

    // SSH Tunnel Config
    SSHTunnelConfig ssh;
    // SSL Config
    SSLConfig ssl;
    
    // Advanced
    SafeModeLevel safeModeLevel;
    QString startupCommand;
    QString preConnectScript;
    // ...
};
```

### `ConnectionSession`
Represents the _active_ runtime state of a connection. Should be a `QObject`.
```cpp
class ConnectionSession : public QObject {
    Q_OBJECT
    Q_PROPERTY(ConnectionStatus status READ status NOTIFY statusChanged)
public:
    QUuid connectionId;
    ConnectionStatus status; // disconnected, connecting, connected, error
    QString currentDatabase; // The currently selected DB (e.g., MySQL "USE db")
    QSharedPointer<DatabaseDriver> driver; // Pointer to the active C++ driver
    // Tracks current queries running, error messages, connection timeline
};
```

## 2. Editor & Tabs Models

### `QueryTab`
Represents a single open tab inside a Connection Window.
```cpp
struct QueryTab {
    QUuid id;
    TabType type; // "table", "query", "structure", "info"
    QString title;

    // For Table Type
    QString schemaName;
    QString tableName;
    
    // For Query Type
    QString query;
    bool hasUserInteraction;
    
    // Results
    QVector<TableRow> resultRows; // Mapped dynamically
    QVector<TableColumn> resultColumns;
    QString errorMessage;
    bool isExecuting;
    float executionTime;
    
    // Pagination / Scrolling state
    int currentOffset;
    
    // Unsaved modifications
    DataChangeTracker pendingChanges; 
};
```

### `CursorPosition`
For tracking where the user is focused inside the Code Editor.
```cpp
struct CursorPosition {
    int start;
    int length;
};
```

## 3. Schema Metadata Models

### `DatabaseMetadata`
```cpp
struct DatabaseMetadata {
    QString name;
    qint64 sizeBytes;
    QVector<SchemaInfo> schemas;
};

struct SchemaInfo {
    QString name;
    QVector<TableInfo> tables;
    QVector<TableInfo> views;
    QVector<RoutineInfo> routines;
};
```

### `TableMetadata`
Defines exactly what a specific table looks like.
```cpp
struct TableMetadata {
    QString name;
    QString schema;
    QString description;
    qint64 rowsEstimate;
    qint64 tableSize;
    QVector<ColumnMetadata> columns;
    QVector<IndexMetadata> indexes;
    QVector<ForeignKeyMetadata> foreignKeys;
};

struct ColumnMetadata {
    QString name;
    QString type;
    bool isNullable;
    bool isAutoIncrement;
    bool isPrimaryKey;
    QVariant defaultValue;
    QString comment;
};
```

## 4. Setting & Preferences Models

### `AppTheme`
System, Light, Dark. Maps directly to Qt palettes and qss.

### `EditorSettings`
```cpp
struct EditorSettings {
    int fontSize;
    QString fontFamily; // Default: Hack/Menlo
    bool wrapLines;
    bool showLineNumbers;
    bool vimModeEnabled;
    bool autoCapitalizeKeywords;
    bool autocompleteEnabled;
};
```

### `AISettings`
```cpp
struct AISettings {
    AIProvider provider; // OpenAI, Anthropic, Ollama, etc.
    QString apiKey;      // Securely stored
    QString customBaseUrl;
    QString modelIdentifier;
    double temperature;
};
```
