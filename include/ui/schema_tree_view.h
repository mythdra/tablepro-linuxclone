#pragma once

#include <QTreeView>
#include <QAbstractItemModel>
#include <QModelIndex>
#include <QVariant>
#include <QVector>
#include <QIcon>
#include <memory>
#include "core/database_types.h"

namespace tablepro {

// Forward declarations
class DatabaseDriver;
struct ConnectionConfig;

// Schema item types
enum class SchemaItemType {
    Root,
    Database,
    Schema,
    Table,
    View,
    Column,
    Index,
    Trigger,
    Function
};

// Schema item data structure
struct SchemaItem {
    SchemaItemType type;
    QString name;
    QString schema;  // Schema name for tables/views
    QString parent;  // Parent item name
    bool loaded = false;  // For lazy loading
    bool hasChildren = true;

    // Children (using raw pointers for QVector compatibility)
    QVector<SchemaItem*> children;
    SchemaItem* parentItem = nullptr;

    // Additional data
    QVariant data;  // Column type, function args, etc.

    ~SchemaItem() {
        qDeleteAll(children);
    }
};

/**
 * Custom model for database schema tree view.
 * Supports lazy loading of schema elements.
 */
class SchemaTreeModel : public QAbstractItemModel
{
    Q_OBJECT

public:
    explicit SchemaTreeModel(QObject* parent = nullptr);
    ~SchemaTreeModel() override;

    // QAbstractItemModel interface
    QModelIndex index(int row, int column, const QModelIndex& parent = QModelIndex()) const override;
    QModelIndex parent(const QModelIndex& child) const override;
    int rowCount(const QModelIndex& parent = QModelIndex()) const override;
    int columnCount(const QModelIndex& parent = QModelIndex()) const override;
    QVariant data(const QModelIndex& index, int role = Qt::DisplayRole) const override;
    QVariant headerData(int section, Qt::Orientation orientation, int role = Qt::DisplayRole) const override;
    Qt::ItemFlags flags(const QModelIndex& index) const override;
    bool hasChildren(const QModelIndex& parent = QModelIndex()) const override;
    bool canFetchMore(const QModelIndex& parent) const override;
    void fetchMore(const QModelIndex& parent) override;

    // Public API
    void setDriver(DatabaseDriver* driver);
    void loadDatabase(const QString& databaseName);
    void refreshSchema(const QString& databaseName);
    void clear();

    SchemaItem* itemFromIndex(const QModelIndex& index) const;
    SchemaItemType typeFromIndex(const QModelIndex& index) const;
    QString nameFromIndex(const QModelIndex& index) const;

signals:
    void schemaLoaded(const QString& databaseName);
    void loadError(const QString& error);

private:
    void loadDatabases();
    void loadTables(SchemaItem* schemaItem);
    void loadColumns(SchemaItem* tableItem);

    QIcon iconForType(SchemaItemType type) const;

    std::unique_ptr<SchemaItem> m_rootItem;
    DatabaseDriver* m_driver = nullptr;
    QString m_currentDatabase;
};

/**
 * Tree view widget for displaying database schema.
 */
class SchemaTreeView : public QTreeView
{
    Q_OBJECT

public:
    explicit SchemaTreeView(QWidget* parent = nullptr);
    ~SchemaTreeView() override;

    void setDriver(DatabaseDriver* driver);
    void loadDatabase(const QString& databaseName);
    void refreshCurrentSchema();
    void clearSchema();

signals:
    void tableSelected(const QString& schema, const QString& table);
    void viewSelected(const QString& schema, const QString& view);
    void columnSelected(const QString& table, const QString& column);
    void refreshRequested();

protected:
    void contextMenuEvent(QContextMenuEvent* event) override;

private slots:
    void onItemActivated(const QModelIndex& index);

private:
    void setupContextMenu();
    void showContextMenu(const QPoint& pos);

    SchemaTreeModel* m_model;
    DatabaseDriver* m_driver = nullptr;

    // Context menu actions
    QAction* m_refreshAction;
    QAction* m_selectAllAction;
    QAction* m_dropTableAction;
    QAction* m_alterTableAction;
};

} // namespace tablepro