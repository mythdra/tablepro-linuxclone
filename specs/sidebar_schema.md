# Sidebar Schema Internals (C++20 + Qt)

## Overview
The sidebar displays the database schema tree using `QTreeView`. C++ backend fetches schema data lazily, caches it, and serves it to the Qt model.

## 1. Lazy Loading Pattern

```cpp
// src/core/SchemaCache.hpp
#pragma once

#include <QObject>
#include <QMap>
#include <QMutex>
#include <QTimer>
#include "core/DatabaseDriver.hpp"

namespace tablepro {

struct SchemaCacheEntry {
    SchemaInfo schema;
    QList<TableInfo> tables;
    QList<TableInfo> views;
    QList<RoutineInfo> routines;
    QElapsedTimer lastFetchTime;
    bool isValid() const {
        return lastFetchTime.elapsed() < 5 * 60 * 1000;  // 5 minutes
    }
};

class SchemaCache : public QObject {
    Q_OBJECT

public:
    explicit SchemaCache(QObject* parent = nullptr);

    // Async fetch with cache
    QFuture<QList<TableInfo>> getTables(
        const QUuid& connectionId,
        const QString& schema);

    QFuture<QList<TableInfo>> getViews(
        const QUuid& connectionId,
        const QString& schema);

    QFuture<QList<RoutineInfo>> getRoutines(
        const QUuid& connectionId,
        const QString& schema);

    // Invalidate cache (called after DDL operations)
    void invalidate(const QUuid& connectionId, const QString& schema);
    void invalidateAll(const QUuid& connectionId);

signals:
    void schemaRefreshed(const QUuid& connectionId);
    void cacheInvalidated(const QUuid& connectionId, const QString& schema);

private:
    struct ConnectionCache {
        QMap<QString, SchemaCacheEntry> schemas;
    };

    QMap<QUuid, ConnectionCache> m_cache;
    mutable QMutex m_mutex;
    QMap<QUuid, DatabaseDriver*> m_drivers;

    SchemaCacheEntry& getOrCreateEntry(
        const QUuid& connectionId,
        const QString& schema);
};

} // namespace tablepro
```

```cpp
// src/core/SchemaCache.cpp
#include "SchemaCache.hpp"
#include <QtConcurrent>

QFuture<QList<TableInfo>> SchemaCache::getTables(
    const QUuid& connectionId,
    const QString& schema)
{
    return QtConcurrent::run([=]() {
        QMutexLocker locker(&m_mutex);

        auto& connCache = m_cache[connectionId];
        auto& entry = connCache.schemas[schema];

        // Cache hit and still valid
        if (entry.isValid() && !entry.tables.isEmpty()) {
            return entry.tables;
        }

        // Cache miss - fetch from driver
        auto* driver = m_drivers.value(connectionId);
        if (!driver) {
            return QList<TableInfo>{};
        }

        SchemaInfo schemaInfo = driver->introspectSchema(schema);
        entry.tables = schemaInfo.tables;
        entry.views = schemaInfo.views;
        entry.lastFetchTime.start();

        return entry.tables;
    });
}
```

## 2. Qt Tree Model

```cpp
// src/ui/Sidebar/SchemaTreeModel.hpp
#pragma once

#include <QAbstractItemModel>
#include <QFutureWatcher>
#include "core/SchemaCache.hpp"

namespace tablepro {

class SchemaTreeItem {
public:
    enum class Type {
        Root,
        Schema,
        Folder,      // "Tables", "Views", "Routines"
        Table,
        View,
        Routine
    };

    SchemaTreeItem(Type type, const QString& name, SchemaTreeItem* parent = nullptr);

    Type type() const { return m_type; }
    QString name() const { return m_name; }
    SchemaTreeItem* parent() const { return m_parent; }
    QList<SchemaTreeItem*> children() const { return m_children; }
    SchemaTreeItem* child(int row) { return m_children.value(row); }
    int childCount() const { return m_children.count(); }
    int row() const;

    void appendChild(SchemaTreeItem* child);
    void clearChildren();

    // Data
    QVariant data(int column, int role) const;
    void setTableInfo(const TableInfo& table) { m_table = table; }
    TableInfo tableInfo() const { return m_table; }

private:
    Type m_type;
    QString m_name;
    SchemaTreeItem* m_parent;
    QList<SchemaTreeItem*> m_children;
    TableInfo m_table;  // For Table/View items
};

class SchemaTreeModel : public QAbstractItemModel {
    Q_OBJECT

public:
    explicit SchemaTreeModel(QObject* parent = nullptr);

    // QAbstractItemModel interface
    QModelIndex index(int row, int column,
                      const QModelIndex& parent = QModelIndex()) const override;
    QModelIndex parent(const QModelIndex& index) const override;
    int rowCount(const QModelIndex& parent = QModelIndex()) const override;
    int columnCount(const QModelIndex& parent = QModelIndex()) const override;
    QVariant data(const QModelIndex& index, int role) const override;
    Qt::ItemFlags flags(const QModelIndex& index) const override;

    // Public API
    void setConnection(const QUuid& connectionId);
    void refresh();

signals:
    void tableDoubleClicked(const TableInfo& table);
    void viewDoubleClicked(const TableInfo& view);

private slots:
    void onTablesFetched(const QList<TableInfo>& tables);
    void onViewsFetched(const QList<TableInfo>& views);

private:
    SchemaTreeItem* m_rootItem;
    QUuid m_connectionId;
    SchemaCache* m_schemaCache;

    SchemaTreeItem* itemFromIndex(const QModelIndex& index) const;
    void loadSchemaChildren(SchemaTreeItem* schemaItem);
    void loadFolderChildren(SchemaTreeItem* folderItem);
};

} // namespace tablepro
```

## 3. Tree View Widget

```cpp
// src/ui/Sidebar/SchemaTreeView.hpp
#pragma once

#include <QTreeView>
#include <QMenu>
#include "SchemaTreeModel.hpp"

namespace tablepro {

class SchemaTreeView : public QTreeView {
    Q_OBJECT

public:
    explicit SchemaTreeView(QWidget* parent = nullptr);

    void setConnection(const QUuid& connectionId);

signals:
    void tableOpened(const TableInfo& table);
    void viewOpened(const TableInfo& view);
    void copyNameRequested(const QString& name);
    void truncateRequested(const TableInfo& table);
    void dropRequested(const TableInfo& table);
    void showDdlRequested(const TableInfo& table);
    void refreshRequested();

private slots:
    void onItemDoubleClicked(const QModelIndex& index);
    void showContextMenu(const QPoint& pos);

private:
    SchemaTreeModel* m_model;
    QMenu* m_contextMenu;
    QUuid m_connectionId;
};

} // namespace tablepro
```

```cpp
// src/ui/Sidebar/SchemaTreeView.cpp
#include "SchemaTreeView.hpp"

void SchemaTreeView::showContextMenu(const QPoint& pos) {
    QModelIndex index = indexAt(pos);
    if (!index.isValid())
        return;

    auto* item = m_model->itemFromIndex(index);
    if (!item || item->type() != SchemaTreeItem::Type::Table)
        return;

    TableInfo table = item->tableInfo();

    m_contextMenu->clear();
    m_contextMenu->addAction(tr("Open Table"), [=]() {
        emit tableOpened(table);
    });
    m_contextMenu->addAction(tr("Copy Name"), [=]() {
        emit copyNameRequested(table.name);
    });
    m_contextMenu->addSeparator();
    m_contextMenu->addAction(tr("Truncate Table..."), [=]() {
        emit truncateRequested(table);
    })->setEnabled(false);  // TODO: Implement
    m_contextMenu->addAction(tr("Drop Table..."), [=]() {
        emit dropRequested(table);
    })->setEnabled(false);  // TODO: Implement
    m_contextMenu->addSeparator();
    m_contextMenu->addAction(tr("Show DDL"), [=]() {
        emit showDdlRequested(table);
    });

    m_contextMenu->exec(viewport()->mapToGlobal(pos));
}
```

## 4. Client-Side Search

```cpp
// src/ui/Sidebar/SearchFilterProxy.hpp
#pragma once

#include <QSortFilterProxyModel>

namespace tablepro {

class SearchFilterProxy : public QSortFilterProxyModel {
    Q_OBJECT
    Q_PROPERTY(QString filterText READ filterText WRITE setFilterText NOTIFY filterTextChanged)

public:
    explicit SearchFilterProxy(QObject* parent = nullptr);

    QString filterText() const { return m_filterText; }
    void setFilterText(const QString& text);

signals:
    void filterTextChanged(const QString& text);

protected:
    bool filterAcceptsRow(int sourceRow,
                          const QModelIndex& sourceParent) const override;

private:
    QString m_filterText;
};

} // namespace tablepro

// src/ui/Sidebar/SearchFilterProxy.cpp
bool SearchFilterProxy::filterAcceptsRow(
    int sourceRow,
    const QModelIndex& sourceParent) const
{
    if (m_filterText.isEmpty())
        return true;

    QModelIndex index = sourceModel()->index(sourceRow, 0, sourceParent);
    if (!index.isValid())
        return true;

    // Check if item name contains filter text (case-insensitive)
    QString name = sourceModel()->data(index, Qt::DisplayRole).toString();
    return name.contains(m_filterText, Qt::CaseInsensitive);
}
```

## 5. Cache Invalidation

```cpp
// After DDL operations, invalidate cache
void ConnectionManager::executeDDL(const QUuid& sessionId, const QString& sql) {
    // Execute the DDL statement
    auto result = m_queryManager->execute(sessionId, sql);

    // Parse DDL type to know which schema to invalidate
    if (sql.trimmed().startsWith("DROP") ||
        sql.trimmed().startsWith("CREATE") ||
        sql.trimmed().startsWith("ALTER")) {

        // Invalidate the affected schema
        m_schemaCache->invalidateAll(sessionId);

        // Notify UI to refresh
        emit schemaChanged(sessionId);
    }
}
```

## 6. Batch Operations

```cpp
// src/ui/Sidebar/BatchOperations.hpp
#pragma once

#include <QObject>
#include <QFuture>

namespace tablepro {

class BatchOperations : public QObject {
    Q_OBJECT

public:
    explicit BatchOperations(QObject* parent = nullptr);

    QFuture<bool> dropTables(
        const QUuid& sessionId,
        const QList<TableInfo>& tables,
        bool confirm = true);

    QFuture<bool> truncateTables(
        const QUuid& sessionId,
        const QList<TableInfo>& tables,
        bool confirm = true);

signals:
    void batchProgress(int current, int total);
    void batchFinished(int successCount, int failureCount);
    void batchError(const QString& message);

private:
    bool confirmBatchDrop(const QList<TableInfo>& tables, QWidget* parent);
};

} // namespace tablepro
```

## 7. Multi-Select Support

```cpp
// In SchemaTreeView
void SchemaTreeView::setSelection(const QRect& rect,
                                  QItemSelectionModel::SelectionFlags flags) {
    // Enable multi-select with Ctrl+Click and Shift+Click
    QTreeView::setSelection(rect, flags);
}

QList<TableInfo> SchemaTreeView::selectedTables() const {
    QList<TableInfo> tables;
    QModelIndexList selected = selectionModel()->selectedRows();

    for (const QModelIndex& index : selected) {
        auto* item = m_model->itemFromIndex(index);
        if (item && item->type() == SchemaTreeItem::Type::Table) {
            tables.append(item->tableInfo());
        }
    }

    return tables;
}
```
