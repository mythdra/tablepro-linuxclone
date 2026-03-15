# Tab Management & Session Restoration (C++20 + Qt)

## 1. Tab Model

```cpp
// src/core/TabInfo.hpp
#pragma once

#include <QUuid>
#include <QString>
#include <QJsonObject>

namespace tablepro {

enum class TabType {
    Query,
    Table,
    Structure
};

struct TabInfo {
    QUuid id;
    TabType type{TabType::Query};
    QString title;
    QString query;

    // Table-specific
    QString tableName;
    QString schemaName;
    bool isView{false};
    QString databaseName;

    bool isExecuting{false};

    // Serialization
    QJsonObject toJson() const;
    static TabInfo fromJson(const QJsonObject& json);
};

struct PersistedTabInfo {
    QUuid id;
    TabType type;
    QString title;
    QString query;  // Truncated if > 500KB
    QString tableName;
    bool isView{false};
    QString databaseName;

    QJsonObject toJson() const;
    static PersistedTabInfo fromJson(const QJsonObject& json);
};

} // namespace tablepro

Q_DECLARE_METATYPE(tablepro::TabInfo)
Q_DECLARE_METATYPE(tablepro::TabType)
```

## 2. Tab Manager Class

```cpp
// src/core/TabManager.hpp
#pragma once

#include <QObject>
#include <QMap>
#include <QUuid>
#include <QDir>
#include "TabInfo.hpp"

namespace tablepro {

class TabManager : public QObject {
    Q_OBJECT

public:
    explicit TabManager(QObject* parent = nullptr);

    // Tab CRUD
    Q_INVOKABLE TabInfo createTab(const QUuid& connectionId, TabType type);
    Q_INVOKABLE void closeTab(const QUuid& connectionId, const QUuid& tabId);
    Q_INVOKABLE void switchTab(const QUuid& connectionId, const QUuid& tabId);
    Q_INVOKABLE void updateTab(const QUuid& connectionId, const TabInfo& tab);

    // Persistence
    Q_INVOKABLE void saveTabs(const QUuid& connectionId);
    Q_INVOKABLE QList<PersistedTabInfo> restoreTabs(const QUuid& connectionId);
    Q_INVOKABLE void saveTabsSync(const QUuid& connectionId);

    // State access
    Q_INVOKABLE QList<TabInfo> getTabs(const QUuid& connectionId) const;
    Q_INVOKABLE QUuid getActiveTab(const QUuid& connectionId) const;
    Q_INVOKABLE TabInfo getTab(const QUuid& connectionId, const QUuid& tabId) const;

    // Memory management
    void evictOldResults(const QUuid& connectionId, int maxCached);

signals:
    void tabsChanged(const QUuid& connectionId);
    void activeTabChanged(const QUuid& connectionId, const QUuid& newActiveTabId);

private:
    struct ConnectionTabState {
        QList<TabInfo> tabs;
        QUuid activeTabId;
        QMap<QUuid, QueryResult> cachedResults;  // For LRU eviction
        QElapsedTimer lastAccessTime;
    };

    QMap<QUuid, ConnectionTabState> m_tabStates;
    QString m_baseDir;  // ~/.config/tablepro/tabs/

    void ensureDirectoryExists();
    QString tabFilePath(const QUuid& connectionId) const;
    void truncateLargeQueries(QList<PersistedTabInfo>& tabs, qint64 maxSize = 500 * 1024);
};

} // namespace tablepro
```

```cpp
// src/core/TabManager.cpp
#include "TabManager.hpp"
#include <QJsonDocument>
#include <QFile>
#include <QStandardPaths>

TabManager::TabManager(QObject* parent)
    : QObject(parent)
{
    m_baseDir = QDir::cleanPath(
        QStandardPaths::writableLocation(QStandardPaths::AppConfigLocation) +
        "/tabs"
    );
    ensureDirectoryExists();
}

TabInfo TabManager::createTab(const QUuid& connectionId, TabType type) {
    TabInfo tab;
    tab.id = QUuid::createUuid();
    tab.type = type;
    tab.title = type == TabType::Query ? tr("New Query") : tr("New Table");
    tab.databaseName = "";

    auto& state = m_tabStates[connectionId];
    state.tabs.append(tab);
    state.activeTabId = tab.id;
    state.lastAccessTime.start();

    saveTabs(connectionId);
    emit tabsChanged(connectionId);
    emit activeTabChanged(connectionId, tab.id);

    return tab;
}

void TabManager::closeTab(const QUuid& connectionId, const QUuid& tabId) {
    auto& state = m_tabStates[connectionId];

    // Remove from cache
    state.cachedResults.remove(tabId);

    // Remove from list
    auto it = std::find_if(state.tabs.begin(), state.tabs.end(),
        [&tabId](const TabInfo& tab) { return tab.id == tabId; });

    if (it != state.tabs.end()) {
        int removedIndex = it - state.tabs.begin();
        state.tabs.erase(it);

        // Update active tab if needed
        if (state.activeTabId == tabId) {
            if (!state.tabs.isEmpty()) {
                // Select next tab, or previous if last
                int newIndex = qMin(removedIndex, state.tabs.count() - 1);
                state.activeTabId = state.tabs[newIndex].id;
            } else {
                state.activeTabId = QUuid();
            }
        }

        saveTabs(connectionId);
        emit tabsChanged(connectionId);
        if (state.activeTabId != tabId) {
            emit activeTabChanged(connectionId, state.activeTabId);
        }
    }
}

void TabManager::saveTabs(const QUuid& connectionId) {
    auto& state = m_tabStates[connectionId];
    state.lastAccessTime.start();

    // Convert to persisted format (truncate large queries)
    QList<PersistedTabInfo> persisted;
    for (const auto& tab : state.tabs) {
        persisted.append(PersistedTabInfo{
            .id = tab.id,
            .type = tab.type,
            .title = tab.title,
            .query = tab.query,  // Will be truncated if needed
            .tableName = tab.tableName,
            .isView = tab.isView,
            .databaseName = tab.databaseName
        });
    }

    truncateLargeQueries(persisted);

    // Serialize to JSON
    QJsonArray tabsJson;
    for (const auto& tab : persisted) {
        tabsJson.append(tab.toJson());
    }

    QJsonObject doc;
    doc["tabs"] = tabsJson;
    doc["activeTabId"] = state.activeTabId.toString();
    doc["lastSaved"] = QDateTime::currentDateTime().toString(Qt::ISODate);

    // Write to file
    QString path = tabFilePath(connectionId);
    QFile file(path);
    if (file.open(QIODevice::WriteOnly)) {
        file.write(QJsonDocument(doc).toJson(QJsonDocument::Compact));
    }
}

void TabManager::saveTabsSync(const QUuid& connectionId) {
    // Synchronous save - used during app quit
    // Same as saveTabs but blocks until complete
    saveTabs(connectionId);
}

QList<PersistedTabInfo> TabManager::restoreTabs(const QUuid& connectionId) {
    QString path = tabFilePath(connectionId);
    QFile file(path);

    if (!file.exists()) {
        return {};  // No saved state
    }

    if (!file.open(QIODevice::ReadOnly)) {
        return {};
    }

    QJsonParseError error;
    QJsonDocument doc = QJsonDocument::fromJson(file.readAll(), &error);

    if (error.error != QJsonParseError::NoError) {
        return {};
    }

    QJsonObject root = doc.object();
    QJsonArray tabsJson = root["tabs"].toArray();

    QList<PersistedTabInfo> tabs;
    for (const auto& value : tabsJson) {
        tabs.append(PersistedTabInfo::fromJson(value.toObject()));
    }

    // Restore active tab
    auto& state = m_tabStates[connectionId];
    state.activeTabId = QUuid::fromString(root["activeTabId"].toString());
    state.lastAccessTime.start();

    return tabs;
}

void TabManager::evictOldResults(const QUuid& connectionId, int maxCached) {
    auto& state = m_tabStates[connectionId];

    if (state.cachedResults.count() <= maxCached) {
        return;  // No eviction needed
    }

    // Sort tabs by last access time (oldest first)
    // This is simplified - real implementation would track per-tab access
    int toEvict = state.cachedResults.count() - maxCached;

    auto it = state.cachedResults.begin();
    while (toEvict > 0 && it != state.cachedResults.end()) {
        it = state.cachedResults.erase(it);
        --toEvict;
    }
}
```

## 3. App Quit — Synchronous Save

```cpp
// In main.cpp or MainWindow closeEvent
void MainWindow::closeEvent(QCloseEvent* event) {
    // Save all open tabs synchronously before quit
    for (auto it = m_tabManager->m_tabStates.begin();
         it != m_tabManager->m_tabStates.end(); ++it) {
        m_tabManager->saveTabsSync(it.key());
    }

    // Save window state
    saveWindowState();

    event->accept();
}

// Or via session manager
void SessionManager::prepareForQuit() {
    QElapsedTimer timer;
    timer.start();

    // Save all tab states
    for (const auto& connectionId : m_tabManager->m_tabStates.keys()) {
        m_tabManager->saveTabsSync(connectionId);
    }

    // Save settings
    m_settings->save();

    // Save query history (flush any pending writes)
    m_historyManager->flush();

    qDebug() << "Session saved in" << timer.elapsed() << "ms";
}
```

## 4. Memory Management

Qt widgets handle memory differently than React:

- **QTableView virtual scrolling**: Only visible rows rendered via model
- **Result data**: Stored in C++ backend, not UI
- **Tab switching**: No LRU eviction needed for UI state
- **Backend eviction**: `TabManager::evictOldResults()` removes cached query results

```cpp
// When switching tabs, check if results are cached
void MainWindow::switchTab(const QUuid& tabId) {
    auto result = m_tabManager->getCachedResult(m_currentConnectionId, tabId);

    if (result.isNull()) {
        // Results were evicted - re-execute the saved query
        auto tab = m_tabManager->getTab(m_currentConnectionId, tabId);
        if (!tab.query.isEmpty()) {
            // Re-execute query silently
            m_queryManager->execute(m_currentConnectionId, tab.query)
                .then([=](const QueryResult& result) {
                    m_tabManager->cacheResult(m_currentConnectionId, tabId, result);
                    m_dataGridView->setModelResult(result);
                });
        }
    } else {
        // Use cached results
        m_dataGridView->setModelResult(result);
    }

    m_tabManager->switchTab(m_currentConnectionId, tabId);
}
```

## 5. Tab Bar Widget

```cpp
// src/ui/TabBar/TabBarWidget.hpp
#pragma once

#include <QTabBar>
#include <QMenu>
#include "core/TabManager.hpp"

namespace tablepro {

class TabBarWidget : public QTabBar {
    Q_OBJECT

public:
    explicit TabBarWidget(QWidget* parent = nullptr);

    void setTabManager(TabManager* manager);
    void setConnection(const QUuid& connectionId);

signals:
    void newTabRequested();
    void tabCloseRequested(const QUuid& tabId);
    void tabSelected(const QUuid& tabId);
    void tabMoved(const QUuid& tabId, int newIndex);

protected:
    void tabRemoved(int index) override;
    void tabInserted(int index) override;
    void mousePressEvent(QMouseEvent* event) override;
    void contextMenuEvent(QContextMenuEvent* event) override;

private:
    TabManager* m_tabManager{nullptr};
    QUuid m_connectionId;
    QMenu* m_contextMenu;

    QUuid tabIdAt(int index) const;
    void updateTabTitles();
};

} // namespace tablepro
```

## 6. Tab Persistence Format

```json
// ~/.config/tablepro/tabs/{connectionId}.json
{
  "tabs": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "type": "query",
      "title": "User Analysis",
      "query": "SELECT * FROM users WHERE status = 'active'...",
      "databaseName": "production"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "type": "table",
      "title": "public.users",
      "query": "SELECT * FROM users LIMIT 500",
      "tableName": "users",
      "databaseName": "production"
    }
  ],
  "activeTabId": "550e8400-e29b-41d4-a716-446655440002",
  "lastSaved": "2026-03-15T10:30:00Z"
}
```
