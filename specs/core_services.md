# Core Services (C++20 + Qt)

## Overview
Business logic lives in C++ classes under `src/core/` and `src/services/`. Services are pure C++ classes (no Qt GUI dependencies) that communicate with UI via signals/slots.

## 1. Export & Import Services (`src/services/`)

### ExportService
```cpp
// src/services/ExportService.hpp
#pragma once

#include <QObject>
#include <QFuture>
#include <QString>
#include "core/QueryResult.hpp"

namespace tablepro {

enum class ExportFormat {
    CSV,
    JSON,
    SQL,
    XLSX,
    Markdown
};

class ExportService : public QObject {
    Q_OBJECT

public:
    explicit ExportService(QObject* parent = nullptr);

    Q_INVOKABLE QFuture<bool> exportTable(
        const QUuid& sessionId,
        const QString& tableName,
        ExportFormat format,
        const QString& outputPath,
        const QVariantMap& options = {});

    Q_INVOKABLE QFuture<bool> exportQuery(
        const QUuid& sessionId,
        const QString& sql,
        ExportFormat format,
        const QString& outputPath,
        const QVariantMap& options = {});

signals:
    void exportProgress(int percent);
    void exportFinished(const QString& outputPath);
    void exportError(const QString& message);

private:
    struct ExportContext {
        QUuid sessionId;
        ExportFormat format;
        QString outputPath;
        QVariantMap options;
        std::atomic<bool> cancelled{false};
    };

    // Format writers
    bool writeCSV(QTextStream& out, const QueryResult& result, const QVariantMap& options);
    bool writeJSON(QTextStream& out, const QueryResult& result, const QVariantMap& options);
    bool writeSQL(QTextStream& out, const QueryResult& result, const QString& tableName);
    bool writeXLSX(const QString& path, const QueryResult& result);  // Requires libxlsxwriter
    bool writeMarkdown(QTextStream& out, const QueryResult& result);
};

} // namespace tablepro
```

- **Streaming**: Large datasets written via `QFile` + `QTextStream` in chunks to avoid RAM spikes
- **XLSX**: Use `libxlsxwriter` C library for Excel file generation
- Progress emitted via `exportProgress(int)` signal
- Frontend connects to signals for progress bar updates

### ImportService
```cpp
// src/services/ImportService.hpp
#pragma once

#include <QObject>
#include <QFuture>

namespace tablepro {

class ImportService : public QObject {
    Q_OBJECT

public:
    explicit ImportService(QObject* parent = nullptr);

    Q_INVOKABLE QFuture<bool> importSQLDump(
        const QUuid& sessionId,
        const QString& inputPath,
        const QString& targetDatabase);

signals:
    void importProgress(int percent, int lineNumber);
    void importFinished();
    void importError(const QString& message, int lineNumber);

private:
    bool streamSQLFile(
        const QString& path,
        DatabaseDriver* driver,
        std::atomic<bool>& cancelled);
};

} // namespace tablepro
```

- **Streaming**: 64KB chunk parser for multi-GB SQL dump files
- **Automatic gzip**: Detect and decompress `.sql.gz` files
- **Transaction wrapping**: Group statements in transactions with foreign key disable
- **Progress**: Emit line number and percent complete

## 2. Formatting Services (`src/services/`)

### SQLFormatterService
```cpp
// src/services/SQLFormatterService.hpp
#pragma once

#include <QString>

namespace tablepro {

class SQLFormatterService {
public:
    // Simple regex-based formatter
    static QString format(const QString& sql, const QString& dialect = "postgres");

    // Or use external library if available
    // (e.g., libsqlformatter C bindings)
};

} // namespace tablepro
```

### DateFormatter
```cpp
// src/services/DateFormatter.hpp
#pragma once

#include <QString>
#include <QDateTime>
#include <QLocale>

namespace tablepro {

class DateFormatter {
public:
    static QString format(const QDateTime& dt, const QString& locale = "en_US");
    static QString formatRelative(const QDateTime& dt);  // "2 hours ago"
    static QString formatDatabaseTimestamp(const QString& dbString, DatabaseType type);
};

} // namespace tablepro
```

## 3. Infrastructure (`src/core/`)

### DeepLinkHandler
- Parses `tablepro://` URLs, extracts connection params
- On macOS: Register via `LSRegisterURLScheme()` or Info.plist
- On Linux: Register via `.desktop` file
- Queues deep links if app not fully loaded yet

```cpp
class DeepLinkHandler : public QObject {
    Q_OBJECT

public:
    void handleUrl(const QString& url);
    void setAppReady(bool ready);

private:
    void processUrl(const QString& url);
    void processQueuedUrls();

    bool m_appReady{false};
    QStringList m_queuedUrls;
};
```

### WindowManager
- Qt manages windows natively via `QMainWindow`
- Multi-window support via multiple `QMainWindow` instances
- Window state persistence via `QSettings`

```cpp
class WindowManager : public QObject {
    Q_OBJECT

public:
    Q_INVOKABLE QMainWindow* createWindow();
    Q_INVOKABLE void closeWindow(const QUuid& windowId);
    Q_INVOKABLE void restoreWindowState(QMainWindow* window);
    Q_INVOKABLE void saveWindowState(const QMainWindow* window);

private:
    QSettings m_settings;
};
```

### Updater (Optional)
- Self-update via custom HTTP check or third-party service
- Downloads new app image (macOS) or installer (Windows)
- Prompts user to install on next launch

## 4. Query Builders (`src/core/`)

### SqlDialect (see database_drivers.md)
Maps `DatabaseType` → quoting rules, param style, pagination syntax

### TableQueryBuilder
```cpp
// src/core/TableQueryBuilder.hpp
#pragma once

#include <QString>
#include "DatabaseDriver.hpp"

namespace tablepro {

class TableQueryBuilder {
public:
    static QString buildSelect(
        const QString& table,
        const QStringList& columns = {"*"},
        const QString& where = {},
        const QString& orderBy = {},
        const QString& orderDir = "ASC",
        int limit = -1,
        int offset = -1,
        DatabaseType type = DatabaseType::PostgreSQL);

    static QString buildInsert(
        const QString& table,
        const QStringList& columns,
        const QVariantList& values,
        DatabaseType type);

    static QString buildUpdate(
        const QString& table,
        const QMap<QString, QVariant>& values,
        const QString& where,
        DatabaseType type);

    static QString buildDelete(
        const QString& table,
        const QString& where,
        DatabaseType type);

private:
    static QString quote(const QString& identifier, DatabaseType type);
};

} // namespace tablepro
```

### RowParser
Converts raw `QVariantList` from drivers into typed frontend-consumable format

### SQLStatementGenerator
Converts `DataChangeManager` deltas into executable SQL

```cpp
// src/core/SQLStatementGenerator.hpp
#pragma once

#include "core/QueryResult.hpp"

namespace tablepro {

struct CellChange {
    int rowIndex;
    QString columnName;
    QVariant oldValue;
    QVariant newValue;
};

struct RowInsertion {
    int rowIndex;
    QString tableName;
    QMap<QString, QVariant> values;
};

struct RowDeletion {
    int rowIndex;
    QString tableName;
    QString primaryKeyColumn;
    QVariant primaryKeyValue;
};

class SQLStatementGenerator {
public:
    static QString generateUpdate(
        const QString& table,
        const QString& primaryKeyColumn,
        const QVariant& primaryKeyValue,
        const CellChange& change,
        DatabaseType type);

    static QString generateInsert(
        const QString& table,
        const RowInsertion& insertion,
        DatabaseType type);

    static QString generateDelete(
        const QString& table,
        const QString& primaryKeyColumn,
        const QVariant& primaryKeyValue,
        DatabaseType type);

    static QString buildBatch(const QList<QString>& statements);
};

} // namespace tablepro
```

## 5. Licensing (`src/services/`)

### LicenseManager
```cpp
// src/services/LicenseManager.hpp
#pragma once

#include <QObject>
#include <QString>

namespace tablepro {

enum class LicenseTier {
    Free,
    Pro,
    Enterprise
};

class LicenseManager : public QObject {
    Q_OBJECT
    Q_PROPERTY(LicenseTier tier READ tier NOTIFY tierChanged)
    Q_PROPERTY(bool isValid READ isValid NOTIFY validationChanged)

public:
    static LicenseManager* instance();

    LicenseTier tier() const { return m_tier; }
    bool isValid() const { return m_valid; }

    Q_INVOKABLE void activateLicense(const QString& licenseKey);
    Q_INVOKABLE void deactivateLicense();
    Q_INVOKABLE bool validateLicense(const QString& licenseKey);

signals:
    void tierChanged(LicenseTier newTier);
    void validationChanged(bool valid);
    void licenseExpiring(const QString& message);

private:
    explicit LicenseManager(QObject* parent = nullptr);

    bool verifySignature(const QByteArray& data, const QByteArray& signature);
    void storeLicenseKey(const QString& key);
    QString loadLicenseKey();

    LicenseTier m_tier{LicenseTier::Free};
    bool m_valid{false};
    QByteArray m_publicKey;  // Ed25519 public key for verification
};

} // namespace tablepro
```

- **License key storage**: QKeychain (secure keychain storage)
- **Signature verification**: Ed25519 via libsodium or `crypto_sign_verify_detached()`
- **Feature gating**: Check `tier()` before enabling Pro features
