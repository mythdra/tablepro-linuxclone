# Phase 8: History & Settings Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement query history with full-text search and application settings with persistence.

**Architecture:** HistoryService uses SQLite with FTS5 for searchable history. SettingsManager uses JSON file with QSettings fallback.

**Tech Stack:** C++20, Qt 6.6 SQL (SQLite), QSettings, QJsonDocument

---

## Task 1: History Service

**Files:**
- Create: `src/services/history_service.hpp`
- Create: `src/services/history_service.cpp`

**Step 1: Create history_service.hpp**

```cpp
#pragma once

#include <QObject>
#include <QSqlDatabase>
#include <QDateTime>

namespace tablepro {

struct HistoryItem {
    qint64 id;
    QString connectionId;
    QString connectionName;
    QString database;
    QString sql;
    qint64 executionTimeMs;
    bool success;
    QString error;
    QDateTime timestamp;
};

class HistoryService : public QObject {
    Q_OBJECT

public:
    static HistoryService* instance();

    void addEntry(const HistoryItem& item);
    QList<HistoryItem> recentEntries(int limit = 100) const;
    QList<HistoryItem> search(const QString& query, int limit = 100) const;
    QList<HistoryItem> byConnection(const QString& connectionId, int limit = 100) const;
    QList<HistoryItem> byDatabase(const QString& database, int limit = 100) const;

    void clear();
    void clearOlderThan(int days);
    qint64 count() const;

signals:
    void entryAdded(const HistoryItem& item);
    void historyCleared();

private:
    explicit HistoryService(QObject* parent = nullptr);

    void initDatabase();
    QString databasePath() const;
    HistoryItem itemFromQuery(const QSqlQuery& query) const;

    QSqlDatabase m_db;
    mutable QMutex m_mutex;
};

} // namespace tablepro
```

**Step 2: Create history_service.cpp**

```cpp
#include "history_service.hpp"
#include <QSqlQuery>
#include <QSqlError>
#include <QStandardPaths>
#include <QDir>
#include <QMutexLocker>

namespace tablepro {

HistoryService* HistoryService::instance() {
    static HistoryService* inst = new HistoryService();
    return inst;
}

HistoryService::HistoryService(QObject* parent)
    : QObject(parent)
{
    initDatabase();
}

QString HistoryService::databasePath() const {
    QString dataPath = QStandardPaths::writableLocation(QStandardPaths::AppDataLocation);
    QDir dir(dataPath);
    if (!dir.exists()) {
        dir.mkpath(".");
    }
    return dataPath + "/history.db";
}

void HistoryService::initDatabase() {
    m_db = QSqlDatabase::addDatabase("QSQLITE", "history");
    m_db.setDatabaseName(databasePath());

    if (!m_db.open()) {
        qWarning() << "Failed to open history database:" << m_db.lastError();
        return;
    }

    QSqlQuery query(m_db);

    // Create main table
    query.exec(R"(
        CREATE TABLE IF NOT EXISTS query_history (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            connection_id TEXT,
            connection_name TEXT,
            database TEXT,
            sql TEXT NOT NULL,
            execution_time_ms INTEGER,
            success INTEGER,
            error TEXT,
            timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    )");

    // Create FTS5 virtual table for search
    query.exec(R"(
        CREATE VIRTUAL TABLE IF NOT EXISTS query_history_fts USING fts5(
            sql,
            content=query_history,
            content_rowid=id
        )
    )");

    // Create triggers to keep FTS in sync
    query.exec(R"(
        CREATE TRIGGER IF NOT EXISTS query_history_ai AFTER INSERT ON query_history BEGIN
            INSERT INTO query_history_fts(rowid, sql) VALUES (new.id, new.sql);
        END
    )");

    query.exec(R"(
        CREATE TRIGGER IF NOT EXISTS query_history_ad AFTER DELETE ON query_history BEGIN
            INSERT INTO query_history_fts(query_history_fts, rowid, sql)
            VALUES('delete', old.id, old.sql);
        END
    )");

    query.exec(R"(
        CREATE TRIGGER IF NOT EXISTS query_history_au AFTER UPDATE ON query_history BEGIN
            INSERT INTO query_history_fts(query_history_fts, rowid, sql)
            VALUES('delete', old.id, old.sql);
            INSERT INTO query_history_fts(rowid, sql) VALUES (new.id, new.sql);
        END
    )");

    // Create index
    query.exec("CREATE INDEX IF NOT EXISTS idx_timestamp ON query_history(timestamp DESC)");
    query.exec("CREATE INDEX IF NOT EXISTS idx_connection ON query_history(connection_id)");
}

void HistoryService::addEntry(const HistoryItem& item) {
    QMutexLocker locker(&m_mutex);

    QSqlQuery query(m_db);
    query.prepare(R"(
        INSERT INTO query_history
        (connection_id, connection_name, database, sql, execution_time_ms, success, error)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    )");

    query.addBindValue(item.connectionId);
    query.addBindValue(item.connectionName);
    query.addBindValue(item.database);
    query.addBindValue(item.sql);
    query.addBindValue(item.executionTimeMs);
    query.addBindValue(item.success ? 1 : 0);
    query.addBindValue(item.error);

    if (!query.exec()) {
        qWarning() << "Failed to add history entry:" << query.lastError();
        return;
    }

    HistoryItem inserted = item;
    inserted.id = query.lastInsertId().toLongLong();

    emit entryAdded(inserted);
}

QList<HistoryItem> HistoryService::recentEntries(int limit) const {
    QMutexLocker locker(&m_mutex);

    QList<HistoryItem> items;

    QSqlQuery query(m_db);
    query.prepare(QString(R"(
        SELECT id, connection_id, connection_name, database, sql,
               execution_time_ms, success, error, timestamp
        FROM query_history
        ORDER BY timestamp DESC
        LIMIT %1
    )").arg(limit));

    if (query.exec()) {
        while (query.next()) {
            items.append(itemFromQuery(query));
        }
    }

    return items;
}

QList<HistoryItem> HistoryService::search(const QString& queryString, int limit) const {
    QMutexLocker locker(&m_mutex);

    QList<HistoryItem> items;

    QSqlQuery query(m_db);
    query.prepare(QString(R"(
        SELECT h.id, h.connection_id, h.connection_name, h.database, h.sql,
               h.execution_time_ms, h.success, h.error, h.timestamp
        FROM query_history h
        JOIN query_history_fts fts ON h.id = fts.rowid
        WHERE query_history_fts MATCH ?
        ORDER BY h.timestamp DESC
        LIMIT %1
    )").arg(limit));

    query.addBindValue(queryString);

    if (query.exec()) {
        while (query.next()) {
            items.append(itemFromQuery(query));
        }
    }

    return items;
}

QList<HistoryItem> HistoryService::byConnection(const QString& connectionId, int limit) const {
    QMutexLocker locker(&m_mutex);

    QList<HistoryItem> items;

    QSqlQuery query(m_db);
    query.prepare(QString(R"(
        SELECT id, connection_id, connection_name, database, sql,
               execution_time_ms, success, error, timestamp
        FROM query_history
        WHERE connection_id = ?
        ORDER BY timestamp DESC
        LIMIT %1
    )").arg(limit));

    query.addBindValue(connectionId);

    if (query.exec()) {
        while (query.next()) {
            items.append(itemFromQuery(query));
        }
    }

    return items;
}

QList<HistoryItem> HistoryService::byDatabase(const QString& database, int limit) const {
    QMutexLocker locker(&m_mutex);

    QList<HistoryItem> items;

    QSqlQuery query(m_db);
    query.prepare(QString(R"(
        SELECT id, connection_id, connection_name, database, sql,
               execution_time_ms, success, error, timestamp
        FROM query_history
        WHERE database = ?
        ORDER BY timestamp DESC
        LIMIT %1
    )").arg(limit));

    query.addBindValue(database);

    if (query.exec()) {
        while (query.next()) {
            items.append(itemFromQuery(query));
        }
    }

    return items;
}

void HistoryService::clear() {
    QMutexLocker locker(&m_mutex);

    QSqlQuery query(m_db);
    query.exec("DELETE FROM query_history");

    emit historyCleared();
}

void HistoryService::clearOlderThan(int days) {
    QMutexLocker locker(&m_mutex);

    QSqlQuery query(m_db);
    query.prepare("DELETE FROM query_history WHERE timestamp < datetime('now', ?)");
    query.addBindValue(QString("-%1 days").arg(days));
    query.exec();
}

qint64 HistoryService::count() const {
    QMutexLocker locker(&m_mutex);

    QSqlQuery query(m_db);
    query.exec("SELECT COUNT(*) FROM query_history");

    if (query.next()) {
        return query.value(0).toLongLong();
    }

    return 0;
}

HistoryItem HistoryService::itemFromQuery(const QSqlQuery& query) const {
    HistoryItem item;
    item.id = query.value(0).toLongLong();
    item.connectionId = query.value(1).toString();
    item.connectionName = query.value(2).toString();
    item.database = query.value(3).toString();
    item.sql = query.value(4).toString();
    item.executionTimeMs = query.value(5).toLongLong();
    item.success = query.value(6).toBool();
    item.error = query.value(7).toString();
    item.timestamp = query.value(8).toDateTime();

    return item;
}

} // namespace tablepro
```

**Step 3: Commit history service**

```bash
git add src/services/history_service.hpp src/services/history_service.cpp
git commit -m "feat: Add HistoryService with FTS5 search"
```

---

## Task 2: Settings Manager

**Files:**
- Create: `src/core/settings_manager.hpp`
- Create: `src/core/settings_manager.cpp`

**Step 1: Create settings_manager.hpp**

```cpp
#pragma once

#include <QObject>
#include <QSettings>
#include <QFont>
#include <QColor>

namespace tablepro {

struct EditorSettings {
    QFont font;
    int fontSize = 12;
    bool showLineNumbers = true;
    bool wordWrap = false;
    int tabWidth = 4;
    bool useTabs = false;
    bool autoComplete = true;
    bool autoIndent = true;
    int autoSaveInterval = 60; // seconds
};

struct QuerySettings {
    int queryTimeout = 30; // seconds
    int defaultLimit = 500;
    bool autoCommit = false;
    bool safeMode = true;
};

struct UISettings {
    QString theme = "dark";
    int sidebarWidth = 250;
    bool showStatusBar = true;
    bool rememberWindowPosition = true;
};

struct ApplicationSettings {
    EditorSettings editor;
    QuerySettings query;
    UISettings ui;
    int historyRetentionDays = 30;
    bool checkUpdatesOnStartup = true;
};

class SettingsManager : public QObject {
    Q_OBJECT

public:
    static SettingsManager* instance();

    ApplicationSettings settings() const;
    void setSettings(const ApplicationSettings& settings);

    // Individual getters/setters
    EditorSettings editorSettings() const;
    void setEditorSettings(const EditorSettings& settings);

    QuerySettings querySettings() const;
    void setQuerySettings(const QuerySettings& settings);

    UISettings uiSettings() const;
    void setUISettings(const UISettings& settings);

    // Convenience
    void setValue(const QString& key, const QVariant& value);
    QVariant value(const QString& key, const QVariant& defaultValue = QVariant()) const;

signals:
    void settingsChanged();
    void editorSettingsChanged(const EditorSettings& settings);
    void themeChanged(const QString& theme);

private:
    explicit SettingsManager(QObject* parent = nullptr);

    void loadSettings();
    void saveSettings();
    QString settingsFilePath() const;

    ApplicationSettings m_settings;
    QSettings m_qsettings;
    mutable QMutex m_mutex;
};

} // namespace tablepro
```

**Step 2: Create settings_manager.cpp**

```cpp
#include "settings_manager.hpp"
#include <QStandardPaths>
#include <QDir>
#include <QJsonDocument>
#include <QJsonObject>
#include <QFile>
#include <QMutexLocker>

namespace tablepro {

SettingsManager* SettingsManager::instance() {
    static SettingsManager* inst = new SettingsManager();
    return inst;
}

SettingsManager::SettingsManager(QObject* parent)
    : QObject(parent)
{
    loadSettings();
}

QString SettingsManager::settingsFilePath() const {
    QString configPath = QStandardPaths::writableLocation(QStandardPaths::AppConfigLocation);
    QDir dir(configPath);
    if (!dir.exists()) {
        dir.mkpath(".");
    }
    return configPath + "/settings.json";
}

void SettingsManager::loadSettings() {
    QFile file(settingsFilePath());

    if (file.open(QIODevice::ReadOnly)) {
        QJsonDocument doc = QJsonDocument::fromJson(file.readAll());
        file.close();

        if (doc.isObject()) {
            QJsonObject json = doc.object();

            // Editor settings
            QJsonObject editor = json["editor"].toObject();
            m_settings.editor.font.setFamily(editor["fontFamily"].toString("JetBrains Mono"));
            m_settings.editor.fontSize = editor["fontSize"].toInt(12);
            m_settings.editor.showLineNumbers = editor["showLineNumbers"].toBool(true);
            m_settings.editor.wordWrap = editor["wordWrap"].toBool(false);
            m_settings.editor.tabWidth = editor["tabWidth"].toInt(4);
            m_settings.editor.useTabs = editor["useTabs"].toBool(false);
            m_settings.editor.autoComplete = editor["autoComplete"].toBool(true);
            m_settings.editor.autoIndent = editor["autoIndent"].toBool(true);
            m_settings.editor.autoSaveInterval = editor["autoSaveInterval"].toInt(60);

            // Query settings
            QJsonObject query = json["query"].toObject();
            m_settings.query.queryTimeout = query["queryTimeout"].toInt(30);
            m_settings.query.defaultLimit = query["defaultLimit"].toInt(500);
            m_settings.query.autoCommit = query["autoCommit"].toBool(false);
            m_settings.query.safeMode = query["safeMode"].toBool(true);

            // UI settings
            QJsonObject ui = json["ui"].toObject();
            m_settings.ui.theme = ui["theme"].toString("dark");
            m_settings.ui.sidebarWidth = ui["sidebarWidth"].toInt(250);
            m_settings.ui.showStatusBar = ui["showStatusBar"].toBool(true);
            m_settings.ui.rememberWindowPosition = ui["rememberWindowPosition"].toBool(true);

            // Application
            m_settings.historyRetentionDays = json["historyRetentionDays"].toInt(30);
            m_settings.checkUpdatesOnStartup = json["checkUpdatesOnStartup"].toBool(true);
        }
    }

    // Apply font size
    m_settings.editor.font.setPointSize(m_settings.editor.fontSize);
}

void SettingsManager::saveSettings() {
    QJsonObject json;

    // Editor settings
    QJsonObject editor;
    editor["fontFamily"] = m_settings.editor.font.family();
    editor["fontSize"] = m_settings.editor.fontSize;
    editor["showLineNumbers"] = m_settings.editor.showLineNumbers;
    editor["wordWrap"] = m_settings.editor.wordWrap;
    editor["tabWidth"] = m_settings.editor.tabWidth;
    editor["useTabs"] = m_settings.editor.useTabs;
    editor["autoComplete"] = m_settings.editor.autoComplete;
    editor["autoIndent"] = m_settings.editor.autoIndent;
    editor["autoSaveInterval"] = m_settings.editor.autoSaveInterval;
    json["editor"] = editor;

    // Query settings
    QJsonObject query;
    query["queryTimeout"] = m_settings.query.queryTimeout;
    query["defaultLimit"] = m_settings.query.defaultLimit;
    query["autoCommit"] = m_settings.query.autoCommit;
    query["safeMode"] = m_settings.query.safeMode;
    json["query"] = query;

    // UI settings
    QJsonObject ui;
    ui["theme"] = m_settings.ui.theme;
    ui["sidebarWidth"] = m_settings.ui.sidebarWidth;
    ui["showStatusBar"] = m_settings.ui.showStatusBar;
    ui["rememberWindowPosition"] = m_settings.ui.rememberWindowPosition;
    json["ui"] = ui;

    // Application
    json["historyRetentionDays"] = m_settings.historyRetentionDays;
    json["checkUpdatesOnStartup"] = m_settings.checkUpdatesOnStartup;

    QFile file(settingsFilePath());
    if (file.open(QIODevice::WriteOnly)) {
        file.write(QJsonDocument(json).toJson());
        file.close();
    }
}

ApplicationSettings SettingsManager::settings() const {
    QMutexLocker locker(&m_mutex);
    return m_settings;
}

void SettingsManager::setSettings(const ApplicationSettings& settings) {
    QMutexLocker locker(&m_mutex);
    m_settings = settings;
    saveSettings();
    emit settingsChanged();
}

EditorSettings SettingsManager::editorSettings() const {
    QMutexLocker locker(&m_mutex);
    return m_settings.editor;
}

void SettingsManager::setEditorSettings(const EditorSettings& settings) {
    QMutexLocker locker(&m_mutex);
    m_settings.editor = settings;
    saveSettings();
    emit editorSettingsChanged(settings);
}

QuerySettings SettingsManager::querySettings() const {
    QMutexLocker locker(&m_mutex);
    return m_settings.query;
}

void SettingsManager::setQuerySettings(const QuerySettings& settings) {
    QMutexLocker locker(&m_mutex);
    m_settings.query = settings;
    saveSettings();
}

UISettings SettingsManager::uiSettings() const {
    QMutexLocker locker(&m_mutex);
    return m_settings.ui;
}

void SettingsManager::setUISettings(const UISettings& settings) {
    QMutexLocker locker(&m_mutex);

    QString oldTheme = m_settings.ui.theme;
    m_settings.ui = settings;
    saveSettings();

    if (oldTheme != settings.theme) {
        emit themeChanged(settings.theme);
    }
}

void SettingsManager::setValue(const QString& key, const QVariant& value) {
    QMutexLocker locker(&m_mutex);
    m_qsettings.setValue(key, value);
    m_qsettings.sync();
}

QVariant SettingsManager::value(const QString& key, const QVariant& defaultValue) const {
    QMutexLocker locker(&m_mutex);
    return m_qsettings.value(key, defaultValue);
}

} // namespace tablepro
```

**Step 3: Commit settings manager**

```bash
git add src/core/settings_manager.hpp src/core/settings_manager.cpp
git commit -m "feat: Add SettingsManager with JSON persistence"
```

---

## Task 3: Settings Dialog

**Files:**
- Create: `src/ui/dialogs/settings_dialog.hpp`
- Create: `src/ui/dialogs/settings_dialog.cpp`

**Step 1: Create settings_dialog.hpp**

```cpp
#pragma once

#include <QDialog>
#include <QTabWidget>
#include <QSpinBox>
#include <QCheckBox>
#include <QFontComboBox>
#include <QComboBox>

namespace tablepro {

class SettingsDialog : public QDialog {
    Q_OBJECT

public:
    explicit SettingsDialog(QWidget* parent = nullptr);

private slots:
    void onAccept();
    void onReset();

private:
    void setupUI();
    void loadSettings();
    void saveSettings();

    QWidget* createEditorTab();
    QWidget* createQueryTab();
    QWidget* createUITab();

    // Editor tab
    QFontComboBox* m_fontCombo;
    QSpinBox* m_fontSizeSpin;
    QCheckBox* m_lineNumbersCheck;
    QCheckBox* m_wordWrapCheck;
    QSpinBox* m_tabWidthSpin;
    QCheckBox* m_autoCompleteCheck;

    // Query tab
    QSpinBox* m_timeoutSpin;
    QSpinBox* m_defaultLimitSpin;
    QCheckBox* m_safeModeCheck;

    // UI tab
    QComboBox* m_themeCombo;
    QSpinBox* m_historyRetentionSpin;
};

} // namespace tablepro
```

**Step 2: Create settings_dialog.cpp**

```cpp
#include "settings_dialog.hpp"
#include "core/settings_manager.hpp"
#include <QVBoxLayout>
#include <QFormLayout>
#include <QGroupBox>
#include <QDialogButtonBox>

namespace tablepro {

SettingsDialog::SettingsDialog(QWidget* parent)
    : QDialog(parent)
    , m_fontCombo(new QFontComboBox(this))
    , m_fontSizeSpin(new QSpinBox(this))
    , m_lineNumbersCheck(new QCheckBox(tr("Show line numbers"), this))
    , m_wordWrapCheck(new QCheckBox(tr("Word wrap"), this))
    , m_tabWidthSpin(new QSpinBox(this))
    , m_autoCompleteCheck(new QCheckBox(tr("Enable autocomplete"), this))
    , m_timeoutSpin(new QSpinBox(this))
    , m_defaultLimitSpin(new QSpinBox(this))
    , m_safeModeCheck(new QCheckBox(tr("Safe mode (require WHERE for UPDATE/DELETE)"), this))
    , m_themeCombo(new QComboBox(this))
    , m_historyRetentionSpin(new QSpinBox(this))
{
    setupUI();
    loadSettings();
}

void SettingsDialog::setupUI() {
    setWindowTitle(tr("Settings"));
    setMinimumSize(500, 400);

    auto* layout = new QVBoxLayout(this);

    // Tabs
    auto* tabs = new QTabWidget(this);
    tabs->addTab(createEditorTab(), tr("Editor"));
    tabs->addTab(createQueryTab(), tr("Query"));
    tabs->addTab(createUITab(), tr("Interface"));

    layout->addWidget(tabs);

    // Buttons
    auto* buttonBox = new QDialogButtonBox(
        QDialogButtonBox::Ok | QDialogButtonBox::Cancel | QDialogButtonBox::Reset,
        this
    );

    connect(buttonBox, &QDialogButtonBox::accepted, this, &SettingsDialog::onAccept);
    connect(buttonBox, &QDialogButtonBox::rejected, this, &QDialog::reject);
    connect(buttonBox->button(QDialogButtonBox::Reset), &QPushButton::clicked, this, &SettingsDialog::onReset);

    layout->addWidget(buttonBox);
}

QWidget* SettingsDialog::createEditorTab() {
    auto* widget = new QWidget(this);
    auto* layout = new QFormLayout(widget);

    layout->addRow(tr("Font:"), m_fontCombo);

    m_fontSizeSpin->setRange(8, 32);
    layout->addRow(tr("Font size:"), m_fontSizeSpin);

    m_tabWidthSpin->setRange(2, 8);
    layout->addRow(tr("Tab width:"), m_tabWidthSpin);

    layout->addRow(m_lineNumbersCheck);
    layout->addRow(m_wordWrapCheck);
    layout->addRow(m_autoCompleteCheck);

    return widget;
}

QWidget* SettingsDialog::createQueryTab() {
    auto* widget = new QWidget(this);
    auto* layout = new QFormLayout(widget);

    m_timeoutSpin->setRange(1, 300);
    m_timeoutSpin->setSuffix(tr(" seconds"));
    layout->addRow(tr("Query timeout:"), m_timeoutSpin);

    m_defaultLimitSpin->setRange(100, 100000);
    m_defaultLimitSpin->setSingleStep(100);
    layout->addRow(tr("Default row limit:"), m_defaultLimitSpin);

    layout->addRow(m_safeModeCheck);

    return widget;
}

QWidget* SettingsDialog::createUITab() {
    auto* widget = new QWidget(this);
    auto* layout = new QFormLayout(widget);

    m_themeCombo->addItem(tr("Dark"), "dark");
    m_themeCombo->addItem(tr("Light"), "light");
    m_themeCombo->addItem(tr("System"), "system");
    layout->addRow(tr("Theme:"), m_themeCombo);

    m_historyRetentionSpin->setRange(1, 365);
    m_historyRetentionSpin->setSuffix(tr(" days"));
    layout->addRow(tr("Keep history for:"), m_historyRetentionSpin);

    return widget;
}

void SettingsDialog::loadSettings() {
    auto settings = SettingsManager::instance()->settings();

    // Editor
    m_fontCombo->setCurrentFont(settings.editor.font);
    m_fontSizeSpin->setValue(settings.editor.fontSize);
    m_tabWidthSpin->setValue(settings.editor.tabWidth);
    m_lineNumbersCheck->setChecked(settings.editor.showLineNumbers);
    m_wordWrapCheck->setChecked(settings.editor.wordWrap);
    m_autoCompleteCheck->setChecked(settings.editor.autoComplete);

    // Query
    m_timeoutSpin->setValue(settings.query.queryTimeout);
    m_defaultLimitSpin->setValue(settings.query.defaultLimit);
    m_safeModeCheck->setChecked(settings.query.safeMode);

    // UI
    int themeIndex = m_themeCombo->findData(settings.ui.theme);
    if (themeIndex >= 0) {
        m_themeCombo->setCurrentIndex(themeIndex);
    }
    m_historyRetentionSpin->setValue(settings.historyRetentionDays);
}

void SettingsDialog::saveSettings() {
    auto settings = SettingsManager::instance()->settings();

    // Editor
    settings.editor.font = m_fontCombo->currentFont();
    settings.editor.fontSize = m_fontSizeSpin->value();
    settings.editor.tabWidth = m_tabWidthSpin->value();
    settings.editor.showLineNumbers = m_lineNumbersCheck->isChecked();
    settings.editor.wordWrap = m_wordWrapCheck->isChecked();
    settings.editor.autoComplete = m_autoCompleteCheck->isChecked();

    // Query
    settings.query.queryTimeout = m_timeoutSpin->value();
    settings.query.defaultLimit = m_defaultLimitSpin->value();
    settings.query.safeMode = m_safeModeCheck->isChecked();

    // UI
    settings.ui.theme = m_themeCombo->currentData().toString();
    settings.historyRetentionDays = m_historyRetentionSpin->value();

    SettingsManager::instance()->setSettings(settings);
}

void SettingsDialog::onAccept() {
    saveSettings();
    accept();
}

void SettingsDialog::onReset() {
    // Reset to defaults
    m_fontCombo->setCurrentFont(QFont("JetBrains Mono"));
    m_fontSizeSpin->setValue(12);
    m_tabWidthSpin->setValue(4);
    m_lineNumbersCheck->setChecked(true);
    m_wordWrapCheck->setChecked(false);
    m_autoCompleteCheck->setChecked(true);
    m_timeoutSpin->setValue(30);
    m_defaultLimitSpin->setValue(500);
    m_safeModeCheck->setChecked(true);
    m_themeCombo->setCurrentIndex(0);
    m_historyRetentionSpin->setValue(30);
}

} // namespace tablepro
```

**Step 3: Commit settings dialog**

```bash
git add src/ui/dialogs/settings_dialog.hpp src/ui/dialogs/settings_dialog.cpp
git commit -m "feat: Add SettingsDialog UI"
```

---

## Task 4: Update CMakeLists and Verify

**Step 1: Add to CMakeLists.txt**

```cmake
set(TABLEPRO_SOURCES
    # ... existing ...
    src/services/history_service.cpp
    src/core/settings_manager.cpp
    src/ui/dialogs/settings_dialog.cpp
)
```

**Step 2: Build**

```bash
cmake --build build/debug -j$(nproc)
```

**Step 3: Commit**

```bash
git add CMakeLists.txt
git commit -m "build: Add history/settings sources"
```

---

## Acceptance Criteria

- [ ] HistoryService stores queries in SQLite
- [ ] FTS5 full-text search works
- [ ] History can be filtered by connection/database
- [ ] SettingsManager persists to JSON
- [ ] SettingsDialog allows editing all settings
- [ ] Theme switching works
- [ ] History retention can be configured

---

**Phase 8 Complete.** Next: Phase 9 - Additional Drivers