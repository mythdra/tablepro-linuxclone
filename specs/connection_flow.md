# Connection Management Flow (C++20 + Qt)

## 1. Connection Storage
```cpp
class ConnectionManager : public QObject {
    Q_OBJECT

public:
    explicit ConnectionManager(QObject* parent = nullptr);

    // CRUD operations
    Q_INVOKABLE bool save(const ConnectionConfig& config);
    Q_INVOKABLE bool remove(const QUuid& id);
    Q_INVOKABLE ConnectionConfig load(const QUuid& id);
    Q_INVOKABLE QList<ConnectionConfig> listAll();

    // Password management (QKeychain)
    Q_INVOKABLE bool savePassword(const QUuid& id, const QString& password);
    Q_INVOKABLE QString loadPassword(const QUuid& id);
    Q_INVOKABLE bool deletePassword(const QUuid& id);

    // Test connection
    Q_INVOKABLE QFuture<bool> testConnection(const ConnectionConfig& config);

private:
    QList<ConnectionConfig> m_connections;
    QString m_configPath;  // ~/.config/tablepro/connections.json
    QKeychain::Job* m_currentKeychainJob{nullptr};

    void loadFromDisk();
    void saveToDisk();
};
```

- **Metadata**: Connections serialized as JSON array to `~/.config/tablepro/connections.json`
- **Passwords**: Stored in OS Keychain via `QKeychain`
  - Key: `tablepro:password:{UUID}`, `tablepro:ssh-password:{UUID}`, `tablepro:ssh-passphrase:{UUID}`
- **Duplicate**: Generate new UUID, suffix name with " (Copy)", copy Keychain entries to new keys

## 2. Connection Dialog (Qt Widgets)
- `QDialog` with tabs: General, SSH, SSL, Advanced
- All form state managed in dialog class
- On Save: calls `ConnectionManager::save()` + `ConnectionManager::savePassword()`
- On Test: calls `ConnectionManager::testConnection()` — returns success/error via signal
- Database type dropdown shows relevant fields (e.g., MongoDB shows Auth Source)

```cpp
class ConnectionDialog : public QDialog {
    Q_OBJECT

public:
    explicit ConnectionDialog(QWidget* parent = nullptr);
    void setConnection(const ConnectionConfig& config);

signals:
    void connectionSaved(const ConnectionConfig& config);

private slots:
    void onTestClicked();
    void onSaveClicked();
    void onDatabaseTypeChanged(int index);

private:
    Ui::ConnectionDialog* ui;
    ConnectionManager* m_manager;
    QStackedWidget* m_databaseTypeStack;  // Show different fields per DB type
};
```

## 3. URL Parser (C++)
```cpp
// src/core/ConnectionUrlParser.hpp
#pragma once

#include <QString>
#include <QUrl>
#include "DatabaseDriver.hpp"

struct ParsedConnectionUrl {
    DatabaseType type{DatabaseType::Unknown};
    QString host;
    int port{0};
    QString database;
    QString username;
    QString password;

    // SSH
    QString sshHost;
    int sshPort{22};
    QString sshUser;
    QString sshPassword;
    QString sshKeyPath;

    // SSL
    QString sslMode;
    QString sslCAPath;
    QString sslCertPath;
    QString sslKeyPath;

    // UI metadata
    QString statusColor;
    QString environment;

    // Deep link actions
    QString schema;
    QString tableName;
    QString filterColumn;
    QString filterOp;
    QString filterValue;

    bool hasSSH() const { return !sshHost.isEmpty(); }
    bool hasSSL() const { return !sslMode.isEmpty(); }
};

class ConnectionUrlParser {
public:
    static ParsedConnectionUrl parse(const QString& urlString);
    static QString toConnectionUrl(const ConnectionConfig& config);

private:
    static DatabaseType typeFromScheme(const QString& scheme);
    static void parseDualAtUrl(const QString& url, ParsedConnectionUrl& result);
    static void parseQueryParams(const QString& query, ParsedConnectionUrl& result);
};
```

```cpp
// src/core/ConnectionUrlParser.cpp
#include "ConnectionUrlParser.hpp"

ParsedConnectionUrl ConnectionUrlParser::parse(const QString& urlString) {
    ParsedConnectionUrl result;

    // Handle schemes: postgres://, mysql://, mongodb://, redis://
    // Handle SSH schemes: postgres+ssh://sshuser@bastion:22/dbuser:pass@dbhost:5432/mydb
    // Handle query params: ?sslmode=require&statusColor=red&env=Production

    QUrl url(urlString);
    QString scheme = url.scheme().toLower();

    // Step 1: Check for "+ssh" in scheme
    if (scheme.contains("+ssh")) {
        scheme = scheme.replace("+ssh", "");
        // SSH will be parsed from the URL authority
    }

    result.type = typeFromScheme(scheme);

    // Step 2: Parse dual @ URLs (SSH + DB credentials)
    if (urlString.contains('@')) {
        parseDualAtUrl(urlString, result);
    } else {
        result.host = url.host();
        result.port = url.port();
        result.username = url.userName();
        result.password = url.password();
    }

    result.database = url.path().mid(1);  // Remove leading /

    // Step 3: Parse query parameters
    parseQueryParams(url.query(), result);

    return result;
}

void ConnectionUrlParser::parseDualAtUrl(
    const QString& url,
    ParsedConnectionUrl& result)
{
    // Format: scheme://sshuser[:sshpass]@sshhost[:sshport]/dbuser[:dbpass]@dbhost[:dbport]/database

    // Split by @ to find the boundary
    int lastAt = url.lastIndexOf('@');
    QString sshPart = url.mid(url.indexOf("://") + 3, lastAt - url.indexOf("://") - 3);
    QString dbPart = url.mid(lastAt + 1);

    // Parse SSH part: sshuser[:sshpass]@sshhost[:sshport]
    if (sshPart.contains('@')) {
        auto [sshUser, sshPass, sshHost, sshPort] = parseUserHost(sshPart);
        result.sshUser = sshUser;
        result.sshPassword = sshPass;
        result.sshHost = sshHost;
        result.sshPort = sshPort;
    }

    // Parse DB part: dbuser[:dbpass]@dbhost[:dbport]/database
    auto [dbUser, dbPass, dbHost, dbPort] = parseUserHost(dbPart);
    result.username = dbUser;
    result.password = dbPass;
    result.host = dbHost;
    result.port = dbPort;
}

void ConnectionUrlParser::parseQueryParams(
    const QString& query,
    ParsedConnectionUrl& result)
{
    QUrlQuery urlQuery(query);

    result.sslMode = urlQuery.queryItemValue("sslmode");
    result.statusColor = urlQuery.queryItemValue("statusColor");
    result.environment = urlQuery.queryItemValue("env");
    result.schema = urlQuery.queryItemValue("schema");
    result.tableName = urlQuery.queryItemValue("table");
    result.filterColumn = urlQuery.queryItemValue("filterColumn");
    result.filterOp = urlQuery.queryItemValue("filterOp");
    result.filterValue = urlQuery.queryItemValue("filterValue");
}
```

## 4. Deep Linking (Qt)
- Register `tablepro://` URL scheme in application bundle (macOS) or desktop file (Linux)
- On URL received: parse with `ConnectionUrlParser::parse()`
- Search existing connections for match (host + port + database + username)
- If match found → open that connection
- If no match → create transient in-memory connection
- Queue URLs if app not fully loaded (buffer until main window ready)
- Post-connect actions: switch schema, open table, apply filter

```cpp
// src/ui/DeepLinkHandler.hpp
#pragma once

#include <QObject>
#include <QStringList>

class ConnectionManager;
class MainWindow;

class DeepLinkHandler : public QObject {
    Q_OBJECT

public:
    explicit DeepLinkHandler(
        ConnectionManager* connMgr,
        MainWindow* mainWindow,
        QObject* parent = nullptr);

    // Called from main() with argc/argv
    void handleCommandLine(const QStringList& arguments);

    // Called from macOS event handler
    void handleUrlEvent(const QString& url);

private:
    void processUrl(const QString& url);
    void queueUrl(const QString& url);
    void processQueuedUrls();

    ConnectionManager* m_connectionManager;
    MainWindow* m_mainWindow;
    QStringList m_queuedUrls;
    bool m_appReady{false};
};
```

## 5. Test Connection Flow
```cpp
QFuture<bool> ConnectionManager::testConnection(const ConnectionConfig& config) {
    return QtConcurrent::run([=]() {
        // 1. Create driver for the database type
        auto driver = DriverFactory::create(config.type);
        if (!driver) {
            emit testFailed(tr("Unknown database type"));
            return false;
        }

        // 2. If SSH enabled, start temporary tunnel
        std::optional<int> localPort;
        std::unique_ptr<SSHTunnel> tunnel;

        if (config.ssh.enabled) {
            tunnel = std::make_unique<SSHTunnel>();
            localPort = tunnel->start(config.ssh);
            if (!localPort) {
                emit testFailed(tr("SSH tunnel failed: %1").arg(tunnel->errorString()));
                return false;
            }
            // Temporarily modify config to use tunnel
            // (config is captured by value, so safe to modify)
        }

        // 3. Attempt connection with timeout
        QElapsedTimer timer;
        timer.start();

        bool success = driver->connect(config);

        if (success) {
            driver->disconnect();
            emit testSucceeded(timer.elapsed());
            return true;
        } else {
            emit testFailed(driver->lastError());
            return false;
        }
    });
}
```

## 6. Pgpass Detection (PostgreSQL only)
```cpp
// src/core/PgpassChecker.hpp
#pragma once

#include <QString>
#include <QFile>
#include <QFileInfo>
#include <QStandardPaths>

#ifdef Q_OS_UNIX
#include <sys/stat.h>
#endif

struct PgpassWarning {
    bool hasWarning{false};
    QString message;
    QString fixCommand;
};

PgpassWarning checkPgpass() {
#ifdef Q_OS_UNIX
    QString pgpassPath = QDir::home().filePath(".pgpass");
    QFileInfo info(pgpassPath);

    if (!info.exists()) {
        return {};  // No file, no warning
    }

    struct stat st;
    if (stat(pgpassPath.toUtf8().constData(), &st) != 0) {
        return {};  // Can't stat, skip
    }

    // Check permissions (should be 0600)
    mode_t mode = st.st_mode & 0777;
    if (mode & 0077) {  // More permissive than 0600
        return {
            .hasWarning = true,
            .message = QObject::tr("~/.pgpass has insecure permissions. "
                                   "Others can read this file."),
            .fixCommand = "chmod 600 ~/.pgpass"
        };
    }
#endif

    return {};
}
```

## 7. Connection Metadata File Format

```json
// ~/.config/tablepro/connections.json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Production PostgreSQL",
    "type": "postgres",
    "group": "Production",
    "colorTag": "red",
    "host": "db.example.com",
    "port": 5432,
    "database": "production",
    "username": "app_user",
    "ssh": {
      "enabled": true,
      "host": "bastion.example.com",
      "port": 22,
      "user": "deploy",
      "authMethod": "key",
      "keyPath": "~/.ssh/id_ed25519"
    },
    "ssl": {
      "enabled": true,
      "mode": "require",
      "caPath": "",
      "certPath": "",
      "keyPath": ""
    },
    "safeMode": "require_where",
    "startupCommand": "SET search_path TO app_schema",
    "lastConnected": "2026-03-15T10:30:00Z"
  }
]
```
