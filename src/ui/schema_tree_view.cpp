#include "schema_tree_view.h"
#include "core/database_driver.h"
#include "core/query_result.h"

#include <QContextMenuEvent>
#include <QMenu>
#include <QAction>
#include <QHeaderView>
#include <QApplication>
#include <QDebug>

namespace tablepro {

// ==================== SchemaTreeModel Implementation ====================

SchemaTreeModel::SchemaTreeModel(QObject* parent)
    : QAbstractItemModel(parent)
    , m_rootItem(std::make_unique<SchemaItem>())
{
    m_rootItem->type = SchemaItemType::Root;
    m_rootItem->name = "Root";
}

SchemaTreeModel::~SchemaTreeModel()
{
    // Root item destructor will delete all children
}

QModelIndex SchemaTreeModel::index(int row, int column, const QModelIndex& parent) const
{
    if (!hasIndex(row, column, parent)) {
        return QModelIndex();
    }

    SchemaItem* parentItem = parent.isValid()
        ? static_cast<SchemaItem*>(parent.internalPointer())
        : m_rootItem.get();

    if (row >= 0 && row < parentItem->children.size()) {
        return createIndex(row, column, parentItem->children[row]);
    }

    return QModelIndex();
}

QModelIndex SchemaTreeModel::parent(const QModelIndex& child) const
{
    if (!child.isValid()) {
        return QModelIndex();
    }

    SchemaItem* childItem = static_cast<SchemaItem*>(child.internalPointer());
    SchemaItem* parentItem = childItem->parentItem;

    if (parentItem == m_rootItem.get() || parentItem == nullptr) {
        return QModelIndex();
    }

    // Find parent's row
    if (parentItem->parentItem) {
        for (int i = 0; i < parentItem->parentItem->children.size(); ++i) {
            if (parentItem->parentItem->children[i] == parentItem) {
                return createIndex(i, 0, parentItem);
            }
        }
    }

    return QModelIndex();
}

int SchemaTreeModel::rowCount(const QModelIndex& parent) const
{
    SchemaItem* parentItem = parent.isValid()
        ? static_cast<SchemaItem*>(parent.internalPointer())
        : m_rootItem.get();

    return parentItem->children.size();
}

int SchemaTreeModel::columnCount(const QModelIndex& parent) const
{
    Q_UNUSED(parent)
    return 1;  // Single column tree view
}

QVariant SchemaTreeModel::data(const QModelIndex& index, int role) const
{
    if (!index.isValid()) {
        return QVariant();
    }

    SchemaItem* item = static_cast<SchemaItem*>(index.internalPointer());

    switch (role) {
    case Qt::DisplayRole:
        return item->name;

    case Qt::DecorationRole:
        return iconForType(item->type);

    case Qt::ToolTipRole:
        switch (item->type) {
        case SchemaItemType::Table:
            return QString("Table: %1.%2").arg(item->schema, item->name);
        case SchemaItemType::Column:
            return QString("Column: %1 (%2)").arg(item->name, item->data.toString());
        default:
            return item->name;
        }

    default:
        return QVariant();
    }
}

QVariant SchemaTreeModel::headerData(int section, Qt::Orientation orientation, int role) const
{
    if (orientation == Qt::Horizontal && role == Qt::DisplayRole && section == 0) {
        return tr("Database Schema");
    }
    return QVariant();
}

Qt::ItemFlags SchemaTreeModel::flags(const QModelIndex& index) const
{
    if (!index.isValid()) {
        return Qt::NoItemFlags;
    }

    Qt::ItemFlags flags = QAbstractItemModel::flags(index);

    SchemaItem* item = static_cast<SchemaItem*>(index.internalPointer());

    // Allow selection and interaction
    flags |= Qt::ItemIsSelectable | Qt::ItemIsEnabled;

    // Tables and views can be dragged
    if (item->type == SchemaItemType::Table || item->type == SchemaItemType::View) {
        flags |= Qt::ItemIsDragEnabled;
    }

    return flags;
}

bool SchemaTreeModel::hasChildren(const QModelIndex& parent) const
{
    if (!parent.isValid()) {
        return true;  // Root always has children
    }

    SchemaItem* item = static_cast<SchemaItem*>(parent.internalPointer());
    return item->hasChildren;
}

bool SchemaTreeModel::canFetchMore(const QModelIndex& parent) const
{
    if (!parent.isValid()) {
        return false;
    }

    SchemaItem* item = static_cast<SchemaItem*>(parent.internalPointer());

    // Can fetch more if has children and not loaded yet
    switch (item->type) {
    case SchemaItemType::Database:
    case SchemaItemType::Schema:
    case SchemaItemType::Table:
        return !item->loaded && item->hasChildren;
    default:
        return false;
    }
}

void SchemaTreeModel::fetchMore(const QModelIndex& parent)
{
    if (!parent.isValid() || !m_driver) {
        return;
    }

    SchemaItem* item = static_cast<SchemaItem*>(parent.internalPointer());

    if (item->loaded) {
        return;
    }

    switch (item->type) {
    case SchemaItemType::Database:
        loadTables(item);
        break;
    case SchemaItemType::Table:
        loadColumns(item);
        break;
    default:
        break;
    }

    item->loaded = true;
}

void SchemaTreeModel::setDriver(DatabaseDriver* driver)
{
    m_driver = driver;
}

void SchemaTreeModel::loadDatabase(const QString& databaseName)
{
    beginResetModel();
    m_rootItem = std::make_unique<SchemaItem>();
    m_rootItem->type = SchemaItemType::Root;

    if (m_driver && m_driver->isConnected()) {
        m_currentDatabase = databaseName;

        // Create database item
        auto* dbItem = new SchemaItem();
        dbItem->type = SchemaItemType::Database;
        dbItem->name = databaseName;
        dbItem->parentItem = m_rootItem.get();
        dbItem->hasChildren = true;

        m_rootItem->children.append(dbItem);
    }

    endResetModel();
}

void SchemaTreeModel::refreshSchema(const QString& databaseName)
{
    loadDatabase(databaseName);
}

void SchemaTreeModel::clear()
{
    beginResetModel();
    m_rootItem = std::make_unique<SchemaItem>();
    m_rootItem->type = SchemaItemType::Root;
    endResetModel();
}

SchemaItem* SchemaTreeModel::itemFromIndex(const QModelIndex& index) const
{
    if (!index.isValid()) {
        return nullptr;
    }
    return static_cast<SchemaItem*>(index.internalPointer());
}

SchemaItemType SchemaTreeModel::typeFromIndex(const QModelIndex& index) const
{
    SchemaItem* item = itemFromIndex(index);
    return item ? item->type : SchemaItemType::Root;
}

QString SchemaTreeModel::nameFromIndex(const QModelIndex& index) const
{
    SchemaItem* item = itemFromIndex(index);
    return item ? item->name : QString();
}

void SchemaTreeModel::loadDatabases()
{
    if (!m_driver || !m_driver->isConnected()) {
        return;
    }

    QueryResult result = m_driver->getDatabases();

    if (!result.success) {
        emit loadError(result.errorMessage);
        return;
    }

    beginResetModel();
    m_rootItem = std::make_unique<SchemaItem>();
    m_rootItem->type = SchemaItemType::Root;

    for (int i = 0; i < result.rowCount(); ++i) {
        auto* dbItem = new SchemaItem();
        dbItem->type = SchemaItemType::Database;
        dbItem->name = result.getValue(i, 0).toString();
        dbItem->parentItem = m_rootItem.get();
        dbItem->hasChildren = true;

        m_rootItem->children.append(dbItem);
    }

    endResetModel();
}

void SchemaTreeModel::loadTables(SchemaItem* dbItem)
{
    if (!m_driver || !m_driver->isConnected() || !dbItem) {
        return;
    }

    QueryResult result = m_driver->getTables(dbItem->name);

    if (!result.success) {
        emit loadError(result.errorMessage);
        return;
    }

    // Group by schema
    QMap<QString, QVector<QPair<QString, QString>>> schemaTables;

    for (int i = 0; i < result.rowCount(); ++i) {
        QString schema = result.getValue(i, 0).toString();
        QString table = result.getValue(i, 1).toString();
        schemaTables[schema].append({table, "table"});
    }

    // Add schema nodes
    int insertPos = dbItem->children.size();
    beginInsertRows(createIndex(0, 0, dbItem), insertPos, insertPos + schemaTables.size() - 1);

    for (auto it = schemaTables.begin(); it != schemaTables.end(); ++it) {
        auto* schemaItem = new SchemaItem();
        schemaItem->type = SchemaItemType::Schema;
        schemaItem->name = it.key();
        schemaItem->parentItem = dbItem;
        schemaItem->hasChildren = true;

        // Add tables under schema
        for (const auto& tableInfo : it.value()) {
            auto* tableItem = new SchemaItem();
            tableItem->type = SchemaItemType::Table;
            tableItem->name = tableInfo.first;
            tableItem->schema = it.key();
            tableItem->parentItem = schemaItem;
            tableItem->hasChildren = true;

            schemaItem->children.append(tableItem);
        }

        dbItem->children.append(schemaItem);
    }

    endInsertRows();
}

void SchemaTreeModel::loadColumns(SchemaItem* tableItem)
{
    if (!m_driver || !m_driver->isConnected() || !tableItem) {
        return;
    }

    QueryResult result = m_driver->getColumns(tableItem->name, tableItem->schema);

    if (!result.success) {
        emit loadError(result.errorMessage);
        return;
    }

    int insertPos = tableItem->children.size();
    beginInsertRows(createIndex(0, 0, tableItem), insertPos, insertPos + result.rowCount() - 1);

    for (int i = 0; i < result.rowCount(); ++i) {
        auto* columnItem = new SchemaItem();
        columnItem->type = SchemaItemType::Column;
        columnItem->name = result.getValue(i, 0).toString();
        columnItem->data = result.getValue(i, 1).toString();  // Data type
        columnItem->parentItem = tableItem;
        columnItem->hasChildren = false;

        tableItem->children.append(columnItem);
    }

    endInsertRows();
}

QIcon SchemaTreeModel::iconForType(SchemaItemType type) const
{
    // Use standard icons or custom icons
    QStyle* style = QApplication::style();

    switch (type) {
    case SchemaItemType::Database:
        return style->standardIcon(QStyle::SP_DirIcon);
    case SchemaItemType::Schema:
        return style->standardIcon(QStyle::SP_DirLinkIcon);
    case SchemaItemType::Table:
        return style->standardIcon(QStyle::SP_FileIcon);
    case SchemaItemType::View:
        return style->standardIcon(QStyle::SP_FileLinkIcon);
    case SchemaItemType::Column:
        return style->standardIcon(QStyle::SP_ArrowRight);
    default:
        return QIcon();
    }
}

// ==================== SchemaTreeView Implementation ====================

SchemaTreeView::SchemaTreeView(QWidget* parent)
    : QTreeView(parent)
    , m_model(new SchemaTreeModel(this))
{
    setModel(m_model);
    setHeaderHidden(false);
    setAnimated(true);
    setExpandsOnDoubleClick(true);
    setSortingEnabled(false);
    setContextMenuPolicy(Qt::DefaultContextMenu);

    // Connect signals
    connect(this, &QTreeView::activated, this, &SchemaTreeView::onItemActivated);

    // Create context menu actions
    m_refreshAction = new QAction(tr("Refresh"), this);
    m_refreshAction->setShortcut(QKeySequence::Refresh);
    connect(m_refreshAction, &QAction::triggered, this, &SchemaTreeView::refreshRequested);

    m_selectAllAction = new QAction(tr("Select All"), this);
    connect(m_selectAllAction, &QAction::triggered, this, &QTreeView::selectAll);

    m_dropTableAction = new QAction(tr("Drop Table..."), this);
    // Connect to drop table handler

    m_alterTableAction = new QAction(tr("Alter Table..."), this);
    // Connect to alter table handler

    header()->setStretchLastSection(true);
}

SchemaTreeView::~SchemaTreeView() = default;

void SchemaTreeView::setDriver(DatabaseDriver* driver)
{
    m_driver = driver;
    m_model->setDriver(driver);
}

void SchemaTreeView::loadDatabase(const QString& databaseName)
{
    m_model->loadDatabase(databaseName);
    expandAll();
}

void SchemaTreeView::refreshCurrentSchema()
{
    m_model->refreshSchema(m_model->nameFromIndex(rootIndex()));
}

void SchemaTreeView::clearSchema()
{
    m_model->clear();
}

void SchemaTreeView::contextMenuEvent(QContextMenuEvent* event)
{
    QModelIndex index = indexAt(event->pos());

    if (!index.isValid()) {
        return;
    }

    showContextMenu(event->globalPos());
}

void SchemaTreeView::showContextMenu(const QPoint& pos)
{
    QMenu menu(this);

    QModelIndex index = currentIndex();
    SchemaItemType type = m_model->typeFromIndex(index);

    menu.addAction(m_refreshAction);
    menu.addSeparator();

    if (type == SchemaItemType::Table) {
        menu.addAction(m_alterTableAction);
        menu.addAction(m_dropTableAction);
    }

    menu.addSeparator();
    menu.addAction(m_selectAllAction);

    menu.exec(pos);
}

void SchemaTreeView::onItemActivated(const QModelIndex& index)
{
    if (!index.isValid()) {
        return;
    }

    SchemaItem* item = m_model->itemFromIndex(index);
    if (!item) {
        return;
    }

    switch (item->type) {
    case SchemaItemType::Table:
        emit tableSelected(item->schema, item->name);
        break;
    case SchemaItemType::View:
        emit viewSelected(item->schema, item->name);
        break;
    case SchemaItemType::Column:
        if (item->parentItem) {
            emit columnSelected(item->parentItem->name, item->name);
        }
        break;
    default:
        break;
    }
}

} // namespace tablepro