# Phase 7: Export/Import Services Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement data export to CSV/JSON/SQL/XLSX and SQL dump import with streaming support.

**Architecture:** ExportService writes results to files in chunks. ImportService parses SQL dumps with 64KB streaming. Both emit progress signals.

**Tech Stack:** C++20, Qt 6.6 (QFile, QTextStream), optional libxlsxwriter for Excel

---

## Task 1: Export Service

**Files:**
- Create: `src/services/export_service.hpp`
- Create: `src/services/export_service.cpp`

**Step 1: Create export_service.hpp**

```cpp
#pragma once

#include <QObject>
#include <QFuture>
#include "core/types.hpp"

namespace tablepro {

enum class ExportFormat {
    CSV,
    JSON,
    SQL,
    XLSX,
    Markdown
};

struct ExportOptions {
    bool includeHeaders = true;
    QString dateFormat = "yyyy-MM-dd";
    QString timestampFormat = "yyyy-MM-dd HH:mm:ss";
    QString nullValue = "NULL";
    QString csvSeparator = ",";
    bool jsonPretty = true;
    bool sqlIncludeCreate = false;
};

class ExportService : public QObject {
    Q_OBJECT

public:
    explicit ExportService(QObject* parent = nullptr);

    QFuture<bool> exportToFile(
        const QueryResult& result,
        const QString& filePath,
        ExportFormat format,
        const ExportOptions& options = ExportOptions()
    );

    QFuture<bool> exportTable(
        const QString& connectionId,
        const QString& schema,
        const QString& table,
        const QString& filePath,
        ExportFormat format,
        const ExportOptions& options = ExportOptions()
    );

signals:
    void progress(int percent);
    void completed(const QString& filePath);
    void error(const QString& message);

private:
    bool writeCSV(const QueryResult& result, const QString& path, const ExportOptions& options);
    bool writeJSON(const QueryResult& result, const QString& path, const ExportOptions& options);
    bool writeSQL(const QueryResult& result, const QString& path, const QString& tableName, const ExportOptions& options);
    bool writeMarkdown(const QueryResult& result, const QString& path, const ExportOptions& options);

    QString escapeCSV(const QString& value) const;
    QString formatValue(const QVariant& value, const ExportOptions& options) const;
};

} // namespace tablepro
```

**Step 2: Create export_service.cpp**

```cpp
#include "export_service.hpp"
#include "core/connection_manager.hpp"
#include <QFile>
#include <QTextStream>
#include <QJsonDocument>
#include <QJsonArray>
#include <QJsonObject>

namespace tablepro {

ExportService::ExportService(QObject* parent)
    : QObject(parent)
{
}

QFuture<bool> ExportService::exportToFile(
    const QueryResult& result,
    const QString& filePath,
    ExportFormat format,
    const ExportOptions& options
) {
    return QtConcurrent::run([this, result, filePath, format, options]() -> bool {
        bool success = false;

        switch (format) {
            case ExportFormat::CSV:
                success = writeCSV(result, filePath, options);
                break;
            case ExportFormat::JSON:
                success = writeJSON(result, filePath, options);
                break;
            case ExportFormat::SQL:
                success = writeSQL(result, filePath, "export_table", options);
                break;
            case ExportFormat::Markdown:
                success = writeMarkdown(result, filePath, options);
                break;
            case ExportFormat::XLSX:
                // TODO: Implement with libxlsxwriter
                emit error("XLSX export not yet implemented");
                return false;
        }

        if (success) {
            emit completed(filePath);
        }

        return success;
    });
}

QFuture<bool> ExportService::exportTable(
    const QString& connectionId,
    const QString& schema,
    const QString& table,
    const QString& filePath,
    ExportFormat format,
    const ExportOptions& options
) {
    return QtConcurrent::run([this, connectionId, schema, table, filePath, format, options]() -> bool {
        auto* driver = ConnectionManager::instance()->driver(connectionId);
        if (!driver) {
            emit error("Connection not found");
            return false;
        }

        // Get table data
        QString sql = QString("SELECT * FROM %1.%2")
            .arg(driver->identifierQuote() + schema + driver->identifierQuote())
            .arg(driver->identifierQuote() + table + driver->identifierQuote());

        auto future = driver->execute(sql);
        future.waitForFinished();

        auto result = future.result();
        if (!result.success) {
            emit error(result.error);
            return false;
        }

        return exportToFile(result, filePath, format, options).result();
    });
}

bool ExportService::writeCSV(const QueryResult& result, const QString& path, const ExportOptions& options) {
    QFile file(path);
    if (!file.open(QIODevice::WriteOnly | QIODevice::Text)) {
        emit error(QString("Cannot open file: %1").arg(path));
        return false;
    }

    QTextStream out(&file);
    out.setEncoding(QStringConverter::Utf8);

    // Headers
    if (options.includeHeaders) {
        QStringList headers;
        for (const auto& col : result.columnNames) {
            headers.append(escapeCSV(col));
        }
        out << headers.join(options.csvSeparator) << "\n";
    }

    // Rows
    int total = result.rows.size();
    for (int i = 0; i < total; ++i) {
        const auto& row = result.rows[i];

        QStringList values;
        for (const auto& colName : result.columnNames) {
            QVariant value = row.value(colName);
            values.append(escapeCSV(formatValue(value, options)));
        }

        out << values.join(options.csvSeparator) << "\n";

        if (i % 1000 == 0) {
            emit progress((i * 100) / total);
        }
    }

    file.close();
    return true;
}

bool ExportService::writeJSON(const QueryResult& result, const QString& path, const ExportOptions& options) {
    QFile file(path);
    if (!file.open(QIODevice::WriteOnly)) {
        emit error(QString("Cannot open file: %1").arg(path));
        return false;
    }

    QJsonArray rowsArray;

    int total = result.rows.size();
    for (int i = 0; i < total; ++i) {
        const auto& row = result.rows[i];

        QJsonObject rowObj;
        for (const auto& colName : result.columnNames) {
            QVariant value = row.value(colName);
            if (value.isNull()) {
                rowObj[colName] = QJsonValue::Null;
            } else {
                rowObj[colName] = QJsonValue::fromVariant(value);
            }
        }
        rowsArray.append(rowObj);

        if (i % 1000 == 0) {
            emit progress((i * 100) / total);
        }
    }

    QJsonDocument doc(rowsArray);
    file.write(options.jsonPretty ? doc.toJson() : doc.toJson(QJsonDocument::Compact));
    file.close();

    return true;
}

bool ExportService::writeSQL(const QueryResult& result, const QString& path, const QString& tableName, const ExportOptions& options) {
    QFile file(path);
    if (!file.open(QIODevice::WriteOnly | QIODevice::Text)) {
        emit error(QString("Cannot open file: %1").arg(path));
        return false;
    }

    QTextStream out(&file);
    out.setEncoding(QStringConverter::Utf8);

    // CREATE TABLE statement
    if (options.sqlIncludeCreate) {
        out << QString("CREATE TABLE %1 (\n").arg(tableName);

        QStringList colDefs;
        for (int i = 0; i < result.columns.size(); ++i) {
            const auto& col = result.columns[i];
            QString def = QString("  %1 %2")
                .arg(col.name)
                .arg(col.typeName.isEmpty() ? "TEXT" : col.typeName);

            if (!col.nullable) {
                def += " NOT NULL";
            }

            colDefs.append(def);
        }
        out << colDefs.join(",\n") << "\n);\n\n";
    }

    // INSERT statements
    int total = result.rows.size();
    for (int i = 0; i < total; ++i) {
        const auto& row = result.rows[i];

        QStringList columns, values;
        for (const auto& colName : result.columnNames) {
            columns.append(colName);

            QVariant value = row.value(colName);
            if (value.isNull()) {
                values.append("NULL");
            } else if (value.userType() == QMetaType::QString) {
                QString escaped = value.toString();
                escaped.replace("'", "''");
                values.append(QString("'%1'").arg(escaped));
            } else {
                values.append(value.toString());
            }
        }

        out << QString("INSERT INTO %1 (%2) VALUES (%3);\n")
            .arg(tableName)
            .arg(columns.join(", "))
            .arg(values.join(", "));

        if (i % 1000 == 0) {
            emit progress((i * 100) / total);
        }
    }

    file.close();
    return true;
}

bool ExportService::writeMarkdown(const QueryResult& result, const QString& path, const ExportOptions& options) {
    QFile file(path);
    if (!file.open(QIODevice::WriteOnly | QIODevice::Text)) {
        emit error(QString("Cannot open file: %1").arg(path));
        return false;
    }

    QTextStream out(&file);
    out.setEncoding(QStringConverter::Utf8);

    // Header row
    out << "| " << result.columnNames.join(" | ") << " |\n";

    // Separator
    QStringList separators;
    for (int i = 0; i < result.columnNames.size(); ++i) {
        separators.append("---");
    }
    out << "| " << separators.join(" | ") << " |\n";

    // Data rows
    for (const auto& row : result.rows) {
        QStringList values;
        for (const auto& colName : result.columnNames) {
            QString val = formatValue(row.value(colName), options);
            val.replace("|", "\\|");  // Escape pipe
            values.append(val);
        }
        out << "| " << values.join(" | ") << " |\n";
    }

    file.close();
    return true;
}

QString ExportService::escapeCSV(const QString& value) const {
    if (value.contains(',') || value.contains('"') || value.contains('\n')) {
        QString escaped = value;
        escaped.replace("\"", "\"\"");
        return QString("\"%1\"").arg(escaped);
    }
    return value;
}

QString ExportService::formatValue(const QVariant& value, const ExportOptions& options) const {
    if (value.isNull()) {
        return options.nullValue;
    }

    return value.toString();
}

} // namespace tablepro
```

**Step 3: Commit export service**

```bash
git add src/services/export_service.hpp src/services/export_service.cpp
git commit -m "feat: Add ExportService for CSV/JSON/SQL/Markdown"
```

---

## Task 2: Import Service

**Files:**
- Create: `src/services/import_service.hpp`
- Create: `src/services/import_service.cpp`

**Step 1: Create import_service.hpp**

```cpp
#pragma once

#include <QObject>
#include <QFuture>
#include <QIODevice>

namespace tablepro {

class ImportService : public QObject {
    Q_OBJECT

public:
    explicit ImportService(QObject* parent = nullptr);

    QFuture<bool> importSQLDump(
        const QString& connectionId,
        const QString& filePath,
        const QString& targetDatabase = QString()
    );

    void cancel();

signals:
    void progress(int percent, int statementsExecuted);
    void completed(int statementsExecuted, qint64 bytesProcessed);
    void error(const QString& message, int lineNumber);
    void statementExecuted(const QString& statement);

private:
    QStringList parseStatements(QIODevice* device);
    bool isGzipFile(const QString& path) const;
    std::atomic<bool> m_cancelled{false};
};

} // namespace tablepro
```

**Step 2: Create import_service.cpp**

```cpp
#include "import_service.hpp"
#include "core/connection_manager.hpp"
#include <QFile>
#include <QTextStream>
#include <QBuffer>

namespace tablepro {

ImportService::ImportService(QObject* parent)
    : QObject(parent)
{
}

void ImportService::cancel() {
    m_cancelled = true;
}

bool ImportService::isGzipFile(const QString& path) const {
    return path.endsWith(".gz", Qt::CaseInsensitive);
}

QStringList ImportService::parseStatements(QIODevice* device) {
    QStringList statements;
    QString currentStatement;
    bool inString = false;
    QChar stringChar;
    bool inComment = false;

    QByteArray chunk;
    while (!(chunk = device->read(65536)).isEmpty() && !m_cancelled) {
        QString text = QString::fromUtf8(chunk);

        for (int i = 0; i < text.length(); ++i) {
            QChar c = text[i];

            // Handle comments
            if (!inString && c == '-' && i + 1 < text.length() && text[i + 1] == '-') {
                inComment = true;
                continue;
            }
            if (inComment && c == '\n') {
                inComment = false;
                continue;
            }
            if (inComment) continue;

            // Handle strings
            if (!inString && (c == '\'' || c == '"')) {
                inString = true;
                stringChar = c;
                currentStatement += c;
                continue;
            }
            if (inString && c == stringChar) {
                // Check for escaped quote
                if (i + 1 < text.length() && text[i + 1] == stringChar) {
                    currentStatement += c;
                    continue;
                }
                inString = false;
                currentStatement += c;
                continue;
            }

            // Handle statement terminator
            if (!inString && c == ';') {
                QString stmt = currentStatement.trimmed();
                if (!stmt.isEmpty()) {
                    statements.append(stmt);
                }
                currentStatement.clear();
                continue;
            }

            currentStatement += c;
        }
    }

    // Add last statement if any
    QString lastStmt = currentStatement.trimmed();
    if (!lastStmt.isEmpty()) {
        statements.append(lastStmt);
    }

    return statements;
}

QFuture<bool> ImportService::importSQLDump(
    const QString& connectionId,
    const QString& filePath,
    const QString& targetDatabase
) {
    return QtConcurrent::run([this, connectionId, filePath, targetDatabase]() -> bool {
        m_cancelled = false;

        auto* driver = ConnectionManager::instance()->driver(connectionId);
        if (!driver) {
            emit error("Connection not found", 0);
            return false;
        }

        QFile file(filePath);
        if (!file.open(QIODevice::ReadOnly)) {
            emit error(QString("Cannot open file: %1").arg(filePath), 0);
            return false;
        }

        // Parse statements
        QStringList statements = parseStatements(&file);
        file.close();

        if (m_cancelled) {
            emit error("Import cancelled", 0);
            return false;
        }

        // Begin transaction
        if (!driver->beginTransaction()) {
            emit error("Failed to begin transaction", 0);
            return false;
        }

        // Execute statements
        int executed = 0;
        int total = statements.size();

        for (const auto& stmt : statements) {
            if (m_cancelled) {
                driver->rollbackTransaction();
                emit error("Import cancelled", executed);
                return false;
            }

            auto future = driver->execute(stmt);
            future.waitForFinished();

            auto result = future.result();
            if (!result.success) {
                driver->rollbackTransaction();
                emit error(result.error, executed);
                return false;
            }

            executed++;
            emit statementExecuted(stmt);

            if (executed % 10 == 0) {
                emit progress((executed * 100) / total, executed);
            }
        }

        // Commit
        if (!driver->commitTransaction()) {
            emit error("Failed to commit transaction", executed);
            return false;
        }

        emit completed(executed, file.size());
        return true;
    });
}

} // namespace tablepro
```

**Step 3: Commit import service**

```bash
git add src/services/import_service.hpp src/services/import_service.cpp
git commit -m "feat: Add ImportService for SQL dump import"
```

---

## Task 3: Export Dialog

**Files:**
- Create: `src/ui/dialogs/export_dialog.hpp`
- Create: `src/ui/dialogs/export_dialog.cpp`

**Step 1: Create export_dialog.hpp**

```cpp
#pragma once

#include <QDialog>
#include <QComboBox>
#include <QLineEdit>
#include <QCheckBox>
#include <QPushButton>
#include "services/export_service.hpp"
#include "core/types.hpp"

namespace tablepro {

class ExportDialog : public QDialog {
    Q_OBJECT

public:
    explicit ExportDialog(const QueryResult& result, QWidget* parent = nullptr);

    QString selectedFilePath() const;
    ExportFormat selectedFormat() const;
    ExportOptions exportOptions() const;

private slots:
    void onBrowseClicked();
    void onFormatChanged(int index);
    void onExportClicked();

private:
    void setupUI();
    void updateOptionsVisibility();

    QueryResult m_result;

    QLineEdit* m_filePathEdit;
    QComboBox* m_formatCombo;
    QPushButton* m_exportButton;
    QPushButton* m_cancelButton;

    // Options
    QCheckBox* m_includeHeadersCheck;
    QLineEdit* m_csvSeparatorEdit;
    QCheckBox* m_jsonPrettyCheck;
};

} // namespace tablepro
```

**Step 2: Create export_dialog.cpp**

```cpp
#include "export_dialog.hpp"
#include <QFileDialog>
#include <QFormLayout>
#include <QGroupBox>
#include <QHBoxLayout>

namespace tablepro {

ExportDialog::ExportDialog(const QueryResult& result, QWidget* parent)
    : QDialog(parent)
    , m_result(result)
    , m_filePathEdit(new QLineEdit(this))
    , m_formatCombo(new QComboBox(this))
    , m_exportButton(new QPushButton(tr("Export"), this))
    , m_cancelButton(new QPushButton(tr("Cancel"), this))
    , m_includeHeadersCheck(new QCheckBox(tr("Include headers"), this))
    , m_csvSeparatorEdit(new QLineEdit(",", this))
    , m_jsonPrettyCheck(new QCheckBox(tr("Pretty print"), this))
{
    setupUI();
}

void ExportDialog::setupUI() {
    setWindowTitle(tr("Export Data"));
    setMinimumWidth(500);

    auto* layout = new QVBoxLayout(this);

    // File selection
    auto* fileGroup = new QGroupBox(tr("File"), this);
    auto* fileLayout = new QHBoxLayout(fileGroup);

    fileLayout->addWidget(m_filePathEdit);
    auto* browseButton = new QPushButton(tr("Browse..."), this);
    fileLayout->addWidget(browseButton);

    layout->addWidget(fileGroup);

    // Format selection
    auto* formatGroup = new QGroupBox(tr("Format"), this);
    auto* formatLayout = new QFormLayout(formatGroup);

    m_formatCombo->addItem("CSV (Comma Separated)", static_cast<int>(ExportFormat::CSV));
    m_formatCombo->addItem("JSON", static_cast<int>(ExportFormat::JSON));
    m_formatCombo->addItem("SQL (INSERT statements)", static_cast<int>(ExportFormat::SQL));
    m_formatCombo->addItem("Markdown", static_cast<int>(ExportFormat::Markdown));
    m_formatCombo->addItem("Excel (XLSX)", static_cast<int>(ExportFormat::XLSX));

    formatLayout->addRow(tr("Format:"), m_formatCombo);

    layout->addWidget(formatGroup);

    // Options
    auto* optionsGroup = new QGroupBox(tr("Options"), this);
    auto* optionsLayout = new QFormLayout(optionsGroup);

    optionsLayout->addRow(m_includeHeadersCheck);
    optionsLayout->addRow(tr("CSV Separator:"), m_csvSeparatorEdit);
    optionsLayout->addRow(m_jsonPrettyCheck);

    layout->addWidget(optionsGroup);

    // Buttons
    auto* buttonLayout = new QHBoxLayout();
    buttonLayout->addStretch();
    buttonLayout->addWidget(m_exportButton);
    buttonLayout->addWidget(m_cancelButton);

    layout->addLayout(buttonLayout);

    // Connect
    connect(browseButton, &QPushButton::clicked, this, &ExportDialog::onBrowseClicked);
    connect(m_formatCombo, QOverload<int>::of(&QComboBox::currentIndexChanged),
            this, &ExportDialog::onFormatChanged);
    connect(m_exportButton, &QPushButton::clicked, this, &ExportDialog::onExportClicked);
    connect(m_cancelButton, &QPushButton::clicked, this, &QDialog::reject);

    // Initial state
    m_includeHeadersCheck->setChecked(true);
    m_jsonPrettyCheck->setChecked(true);
    updateOptionsVisibility();
}

void ExportDialog::onBrowseClicked() {
    QString filter;
    switch (selectedFormat()) {
        case ExportFormat::CSV: filter = "CSV Files (*.csv)"; break;
        case ExportFormat::JSON: filter = "JSON Files (*.json)"; break;
        case ExportFormat::SQL: filter = "SQL Files (*.sql)"; break;
        case ExportFormat::Markdown: filter = "Markdown Files (*.md)"; break;
        case ExportFormat::XLSX: filter = "Excel Files (*.xlsx)"; break;
    }

    QString path = QFileDialog::getSaveFileName(this, tr("Export File"), QString(), filter);
    if (!path.isEmpty()) {
        m_filePathEdit->setText(path);
    }
}

void ExportDialog::onFormatChanged(int index) {
    Q_UNUSED(index)
    updateOptionsVisibility();
}

void ExportDialog::updateOptionsVisibility() {
    ExportFormat format = selectedFormat();

    m_csvSeparatorEdit->setEnabled(format == ExportFormat::CSV);
    m_jsonPrettyCheck->setEnabled(format == ExportFormat::JSON);
    m_includeHeadersCheck->setEnabled(format == ExportFormat::CSV || format == ExportFormat::Markdown);
}

void ExportDialog::onExportClicked() {
    if (m_filePathEdit->text().isEmpty()) {
        m_filePathEdit->setFocus();
        return;
    }

    accept();
}

QString ExportDialog::selectedFilePath() const {
    return m_filePathEdit->text();
}

ExportFormat ExportDialog::selectedFormat() const {
    return static_cast<ExportFormat>(m_formatCombo->currentData().toInt());
}

ExportOptions ExportDialog::exportOptions() const {
    ExportOptions options;
    options.includeHeaders = m_includeHeadersCheck->isChecked();
    options.csvSeparator = m_csvSeparatorEdit->text();
    options.jsonPretty = m_jsonPrettyCheck->isChecked();
    return options;
}

} // namespace tablepro
```

**Step 3: Commit export dialog**

```bash
git add src/ui/dialogs/export_dialog.hpp src/ui/dialogs/export_dialog.cpp
git commit -m "feat: Add ExportDialog UI"
```

---

## Task 4: Update CMakeLists and Verify

**Step 1: Add to CMakeLists.txt**

```cmake
set(TABLEPRO_SOURCES
    # ... existing ...
    src/services/export_service.cpp
    src/services/import_service.cpp
    src/ui/dialogs/export_dialog.cpp
)
```

**Step 2: Build**

```bash
cmake --build build/debug -j$(nproc)
```

**Step 3: Commit**

```bash
git add CMakeLists.txt
git commit -m "build: Add export/import sources"
```

---

## Acceptance Criteria

- [ ] Export to CSV working
- [ ] Export to JSON working
- [ ] Export to SQL INSERT statements
- [ ] Export to Markdown
- [ ] Import SQL dump with streaming
- [ ] Progress reporting
- [ ] Cancel support
- [ ] Export dialog UI functional

---

**Phase 7 Complete.** Next: Phase 8 - History & Settings