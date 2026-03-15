# Data Models Specification (C++20 + Qt)

Data flows between C++ backend and Qt UI components via signals/slots and QVariant. All models use Qt's meta-object system for serialization.

## 1. Connection Models

### `ConnectionConfig` (C++ struct)
```cpp
struct ConnectionConfig {
    QUuid id;
    QString name;
    DatabaseType type;       // "postgres", "mysql", etc.
    QString group;
    QString colorTag;        // "red", "blue", "green"

    // Core
    QString host;
    int port{5432};
    QString database;
    QString username;
    QString localFilePath;   // SQLite/DuckDB

    // SSH
    SSHTunnelConfig ssh;
    // SSL
    SSLConfig ssl;

    // Advanced
    SafeModeLevel safeMode{SafeModeLevel::Off};
    QString startupCommand;
    QString preConnectScript;

    // Password is NEVER in this struct — stored in QKeychain
};

Q_DECLARE_METATYPE(ConnectionConfig)
```

### `ConnectionSession` (runtime only, not serialized)
```cpp
class ConnectionSession : public QObject {
    Q_OBJECT
    Q_PROPERTY(QUuid connectionId READ connectionId CONSTANT)
    Q_PROPERTY(ConnectionStatus status READ status NOTIFY statusChanged)
    Q_PROPERTY(QString activeDatabase READ activeDatabase NOTIFY databaseChanged)

public:
    QUuid connectionId() const { return m_connectionId; }
    ConnectionStatus status() const { return m_status; }
    QString activeDatabase() const { return m_activeDatabase; }

    // Driver is owned by the session
    std::unique_ptr<DatabaseDriver> driver() { return std::move(m_driver); }

signals:
    void statusChanged(ConnectionStatus newStatus);
    void databaseChanged(const QString& newDb);

private:
    QUuid m_connectionId;
    ConnectionStatus m_status{ConnectionStatus::Disconnected};
    QString m_activeDatabase;
    std::unique_ptr<DatabaseDriver> m_driver;
    std::unique_ptr<SSHTunnel> m_sshTunnel;
    QElapsedTimer m_lastPingTimer;
};
```

## 2. Editor & Tab Models

### `QueryTab` (C++ class with QProperties)
```cpp
class QueryTab : public QObject {
    Q_OBJECT
    Q_PROPERTY(QUuid id READ id CONSTANT)
    Q_PROPERTY(TabType type READ type WRITE setType NOTIFY typeChanged)
    Q_PROPERTY(QString title READ title WRITE setTitle NOTIFY titleChanged)
    Q_PROPERTY(QString query READ query WRITE setQuery NOTIFY queryChanged)

    // Table-specific
    Q_PROPERTY(QString schemaName READ schemaName NOTIFY schemaChanged)
    Q_PROPERTY(QString tableName READ tableName NOTIFY tableChanged)
    Q_PROPERTY(bool isView READ isView NOTIFY viewChanged)
    Q_PROPERTY(QString databaseName READ databaseName NOTIFY databaseChanged)

    // Results (populated after query execution)
    Q_PROPERTY(QVariantList columns READ columns NOTIFY resultsReady)
    Q_PROPERTY(QVariantList rows READ rows NOTIFY resultsReady)
    Q_PROPERTY(qint64 totalRowCount READ totalRowCount NOTIFY resultsReady)
    Q_PROPERTY(int offset READ offset WRITE setOffset NOTIFY paginationChanged)
    Q_PROPERTY(double executionTime READ executionTime NOTIFY resultsReady)
    Q_PROPERTY(QString errorMessage READ errorMessage NOTIFY errorOccurred)
    Q_PROPERTY(bool executing READ isExecuting NOTIFY executionStateChanged)

public:
    // Getters/setters...
    QVariantList columns() const;  // Serialized ColumnInfo list
    QVariantList rows() const;     // 2D array for grid: [[cell, cell], [cell, cell]]

signals:
    void typeChanged(TabType newType);
    void titleChanged(const QString& newTitle);
    void queryChanged(const QString& newQuery);
    void resultsReady();
    void errorOccurred(const QString& message);

private:
    QUuid m_id;
    TabType m_type{TabType::Query};
    QString m_title;
    QString m_query;
    QString m_schemaName;
    QString m_tableName;
    bool m_isView{false};
    QString m_databaseName;

    // Results
    QList<ColumnInfo> m_columns;
    QVariantList m_rows;  // QVariant for JSON-like flexibility
    qint64 m_totalRowCount{0};
    int m_offset{0};
    double m_executionTime{0.0};
    QString m_errorMessage;
    bool m_isExecuting{false};
};

Q_DECLARE_METATYPE(QueryTab*)
```

### `PersistedTab` (disk-only, no result data)
```cpp
struct PersistedTab {
    QUuid id;
    QString title;
    QString query;
    TabType type;
    QString tableName;
    bool isView{false};
    QString databaseName;

    // JSON serialization helpers
    QJsonObject toJson() const;
    static PersistedTab fromJson(const QJsonObject& json);
};
```

## 3. Schema Metadata Models

```cpp
struct DatabaseInfo {
    QString name;
    qint64 sizeBytes{0};

    QJsonObject toJson() const;
    static DatabaseInfo fromJson(const QJsonObject& json);
};

struct SchemaInfo {
    QString name;
    QList<TableInfo> tables;
    QList<TableInfo> views;
    QList<RoutineInfo> routines;

    QJsonObject toJson() const;
};

struct ColumnMetadata {
    QString name;
    QString type;
    bool isNullable{true};
    bool isAutoIncrement{false};
    bool isPrimaryKey{false};
    QVariant defaultValue;  // QVariant for type flexibility
    QString comment;

    QJsonObject toJson() const;
};

Q_DECLARE_METATYPE(DatabaseInfo)
Q_DECLARE_METATYPE(SchemaInfo)
Q_DECLARE_METATYPE(ColumnMetadata)
```

## 4. Settings & Preferences

```cpp
class AppSettings : public QObject {
    Q_OBJECT
    Q_PROPERTY(QString theme READ theme WRITE setTheme NOTIFY themeChanged)
    Q_PROPERTY(int fontSize READ fontSize WRITE setFontSize NOTIFY fontChanged)
    Q_PROPERTY(QString fontFamily READ fontFamily WRITE setFontFamily NOTIFY fontChanged)
    Q_PROPERTY(bool wrapLines READ wrapLines WRITE setWrapLines NOTIFY wrapChanged)
    Q_PROPERTY(bool showLineNumbers READ showLineNumbers WRITE setShowLineNumbers NOTIFY lineNumbersChanged)
    Q_PROPERTY(bool vimMode READ vimMode WRITE setVimMode NOTIFY vimModeChanged)
    Q_PROPERTY(bool autoCapitalize READ autoCapitalize WRITE setAutoCapitalize NOTIFY capitalizeChanged)
    Q_PROPERTY(bool autocomplete READ autocomplete WRITE setAutocomplete NOTIFY autocompleteChanged)
    Q_PROPERTY(int queryTimeout READ queryTimeout WRITE setQueryTimeout NOTIFY timeoutChanged)
    Q_PROPERTY(int rowsPerPage READ rowsPerPage WRITE setRowsPerPage NOTIFY paginationChanged)

public:
    // Getters/setters with change notifications...
    static AppSettings* instance();  // Singleton
    void load();  // From ~/.config/tablepro/settings.json
    void save();

signals:
    void themeChanged(const QString& theme);
    void fontChanged();
    void wrapChanged(bool enabled);
    void lineNumbersChanged(bool enabled);
    void vimModeChanged(bool enabled);
    void timeoutChanged(int seconds);
    void paginationChanged(int rows);

private:
    explicit AppSettings(QObject* parent = nullptr);
    QString m_theme{"system"};
    int m_fontSize{14};
    QString m_fontFamily{"JetBrains Mono"};
    bool m_wrapLines{false};
    bool m_showLineNumbers{true};
    bool m_vimMode{false};
    bool m_autoCapitalize{true};
    bool m_autocomplete{true};
    int m_queryTimeout{30};
    int m_rowsPerPage{500};
};

Q_DECLARE_METATYPE(AppSettings*)
```

## 5. Qt Meta-Object System

All custom types must be registered with Qt's meta-object system for signals/slots and QVariant support:

```cpp
// In main.cpp or initialization
qRegisterMetaType<ConnectionConfig>("ConnectionConfig");
qRegisterMetaType<QueryTab*>("QueryTab*");
qRegisterMetaType<DatabaseInfo>("DatabaseInfo");
qRegisterMetaType<SchemaInfo>("SchemaInfo");
qRegisterMetaType<ColumnMetadata>("ColumnMetadata");
qRegisterMetaType<ConnectionStatus>("ConnectionStatus");
qRegisterMetaType<TabType>("TabType");
```

## 6. JSON Serialization

Qt provides built-in JSON support via `QJsonDocument`, `QJsonObject`, `QJsonArray`:

```cpp
// Example: ConnectionConfig serialization
QJsonObject ConnectionConfig::toJson() const {
    QJsonObject obj;
    obj["id"] = id.toString();
    obj["name"] = name;
    obj["type"] = QString::fromStdString(databaseTypeToString(type));
    obj["host"] = host;
    obj["port"] = port;
    // ... other fields
    return obj;
}

ConnectionConfig ConnectionConfig::fromJson(const QJsonObject& json) {
    ConnectionConfig config;
    config.id = QUuid::fromString(json["id"].toString());
    config.name = json["name"].toString();
    config.type = stringToDatabaseType(json["type"].toString().toStdString());
    config.host = json["host"].toString();
    config.port = json["port"].toInt();
    // ... other fields
    return config;
}
```

## 7. Storage Locations

| Data Type | Storage Mechanism | Location |
|-----------|-------------------|----------|
| Passwords | QKeychain | macOS Keychain / Windows Credential Manager / libsecret |
| Preferences | QSettings / JSON | `~/.config/tablepro/settings.json` |
| Query History | SQLite with FTS5 | `~/.config/tablepro/history.db` |
| Tab State | JSON per connection | `~/.config/tablepro/tabs/{uuid}.json` |
| SSH Keys | QKeychain + file paths | OS Keychain + `~/.ssh/` |
