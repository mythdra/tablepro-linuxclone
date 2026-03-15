# Phase 5: Data Grid & Mutation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build data grid with QTableView for displaying query results, inline editing, change tracking, and SQL generation for commits.

**Architecture:** Custom QAbstractTableModel with virtual scrolling. DataChangeManager tracks edits/new rows/deletes. SqlGenerator produces UPDATE/INSERT/DELETE statements.

**Tech Stack:** C++20, Qt 6.6 Widgets (QTableView, QAbstractTableModel), Qt Concurrent

---

## Task 1: ResultSet Model

**Files:**
- Create: `src/ui/grid/result_set_model.hpp`
- Create: `src/ui/grid/result_set_model.cpp`

**Step 1: Create result_set_model.hpp**

```cpp
#pragma once

#include <QAbstractTableModel>
#include <QList>
#include <QVariantMap>
#include "core/types.hpp"

namespace tablepro {

class ResultSetModel : public QAbstractTableModel {
    Q_OBJECT

public:
    explicit ResultSetModel(QObject* parent = nullptr);

    void setQueryResult(const QueryResult& result);
    const QueryResult& queryResult() const { return m_result; }

    // QAbstractTableModel interface
    int rowCount(const QModelIndex& parent = QModelIndex()) const override;
    int columnCount(const QModelIndex& parent = QModelIndex()) const override;
    QVariant data(const QModelIndex& index, int role = Qt::DisplayRole) const override;
    bool setData(const QModelIndex& index, const QVariant& value, int role = Qt::EditRole) override;
    QVariant headerData(int section, Qt::Orientation orientation, int role = Qt::DisplayRole) const override;
    Qt::ItemFlags flags(const QModelIndex& index) const override;

    // Editing
    void markRowAsNew(int row);
    void markRowAsDeleted(int row);
    bool isRowNew(int row) const;
    bool isRowDeleted(int row) const;

signals:
    void dataEdited(int row, int column, const QVariant& oldValue, const QVariant& newValue);

private:
    QString formatValue(const QVariant& value, const QString& typeName) const;

    QueryResult m_result;
    QSet<int> m_newRows;
    QSet<int> m_deletedRows;
};

} // namespace tablepro
```

**Step 2: Create result_set_model.cpp**

```cpp
#include "result_set_model.hpp"
#include <QFont>
#include <QBrush>
#include <QDateTime>

namespace tablepro {

ResultSetModel::ResultSetModel(QObject* parent)
    : QAbstractTableModel(parent)
{
}

void ResultSetModel::setQueryResult(const QueryResult& result) {
    beginResetModel();
    m_result = result;
    m_newRows.clear();
    m_deletedRows.clear();
    endResetModel();
}

int ResultSetModel::rowCount(const QModelIndex& parent) const {
    Q_UNUSED(parent)
    return m_result.rows.size();
}

int ResultSetModel::columnCount(const QModelIndex& parent) const {
    Q_UNUSED(parent)
    return m_result.columnNames.size();
}

QVariant ResultSetModel::data(const QModelIndex& index, int role) const {
    if (!index.isValid() || index.row() >= m_result.rows.size()) {
        return QVariant();
    }

    const auto& row = m_result.rows[index.row()];
    const QString& columnName = m_result.columnNames[index.column()];
    const QVariant& value = row.value(columnName);

    if (role == Qt::DisplayRole || role == Qt::EditRole) {
        if (value.isNull()) {
            return QString("NULL");
        }

        QString typeName;
        if (index.column() < m_result.columns.size()) {
            typeName = m_result.columns[index.column()].typeName;
        }

        return formatValue(value, typeName);
    }

    if (role == Qt::ForegroundRole) {
        if (value.isNull()) {
            return QColor("#929292");  // Gray for NULL
        }
        if (isRowDeleted(index.row())) {
            return QColor("#F38BA8");  // Red for deleted
        }
    }

    if (role == Qt::FontRole) {
        QFont font;
        if (value.isNull()) {
            font.setItalic(true);
        }
        if (isRowDeleted(index.row())) {
            font.setStrikeOut(true);
        }
        return font;
    }

    if (role == Qt::BackgroundRole) {
        if (isRowNew(index.row())) {
            return QColor("#A6E3A1");  // Green for new rows
        }
        // Yellow for edited cells handled by delegate
    }

    if (role == Qt::TextAlignmentRole) {
        // Right-align numbers
        if (index.column() < m_result.columns.size()) {
            QString typeName = m_result.columns[index.column()].typeName.toLower();
            if (typeName.contains("int") || typeName.contains("float") ||
                typeName.contains("double") || typeName.contains("numeric") ||
                typeName.contains("decimal")) {
                return QVariant(Qt::AlignRight | Qt::AlignVCenter);
            }
        }
        return QVariant(Qt::AlignLeft | Qt::AlignVCenter);
    }

    return QVariant();
}

bool ResultSetModel::setData(const QModelIndex& index, const QVariant& value, int role) {
    if (role != Qt::EditRole || !index.isValid()) {
        return false;
    }

    const QString& columnName = m_result.columnNames[index.column()];
    QVariant oldValue = m_result.rows[index.row()].value(columnName);

    m_result.rows[index.row()].insert(columnName, value);

    emit dataEdited(index.row(), index.column(), oldValue, value);
    emit dataChanged(index, index, {Qt::DisplayRole, Qt::EditRole});

    return true;
}

QVariant ResultSetModel::headerData(int section, Qt::Orientation orientation, int role) const {
    if (role != Qt::DisplayRole) {
        return QVariant();
    }

    if (orientation == Qt::Horizontal) {
        if (section < m_result.columnNames.size()) {
            return m_result.columnNames[section];
        }
    } else {
        return QString::number(section + 1);
    }

    return QVariant();
}

Qt::ItemFlags ResultSetModel::flags(const QModelIndex& index) const {
    Qt::ItemFlags defaultFlags = QAbstractTableModel::flags(index);

    if (index.isValid() && !isRowDeleted(index.row())) {
        return defaultFlags | Qt::ItemIsEditable;
    }

    return defaultFlags;
}

void ResultSetModel::markRowAsNew(int row) {
    m_newRows.insert(row);
    emit dataChanged(index(row, 0), index(row, columnCount() - 1));
}

void ResultSetModel::markRowAsDeleted(int row) {
    m_deletedRows.insert(row);
    emit dataChanged(index(row, 0), index(row, columnCount() - 1));
}

bool ResultSetModel::isRowNew(int row) const {
    return m_newRows.contains(row);
}

bool ResultSetModel::isRowDeleted(int row) const {
    return m_deletedRows.contains(row);
}

QString ResultSetModel::formatValue(const QVariant& value, const QString& typeName) const {
    if (value.isNull()) {
        return QString("NULL");
    }

    QString type = typeName.toLower();

    if (type.contains("json")) {
        return value.toString().left(100) + (value.toString().length() > 100 ? "..." : "");
    }

    if (type.contains("timestamp") || type.contains("date")) {
        return value.toString();
    }

    return value.toString();
}

} // namespace tablepro
```

**Step 3: Commit model**

```bash
git add src/ui/grid/result_set_model.hpp src/ui/grid/result_set_model.cpp
git commit -m "feat: Add ResultSetModel for QTableView"
```

---

## Task 2: DataGrid Widget

**Files:**
- Create: `src/ui/grid/data_grid.hpp`
- Create: `src/ui/grid/data_grid.cpp`

**Step 1: Create data_grid.hpp**

```cpp
#pragma once

#include <QTableView>
#include <QUndoStack>
#include "result_set_model.hpp"
#include "core/types.hpp"

namespace tablepro {

class DataChangeManager;

class DataGrid : public QTableView {
    Q_OBJECT

public:
    explicit DataGrid(QWidget* parent = nullptr);
    ~DataGrid() override;

    void setQueryResult(const QueryResult& result);
    QueryResult queryResult() const;

    void setTableName(const QString& schema, const QString& table);
    QString schema() const { return m_schema; }
    QString tableName() const { return m_tableName; }

    void setPrimaryKeyColumns(const QStringList& columns);

    // Change tracking
    bool hasChanges() const;
    void discardChanges();
    PendingChanges getPendingChanges() const;

    // Row operations
    void addNewRow();
    void deleteSelectedRows();

    // Commit
    QFuture<bool> commitChanges(const QString& connectionId);

signals:
    void changesChanged(bool hasChanges);
    void commitCompleted(bool success);
    void commitFailed(const QString& error);

protected:
    void keyPressEvent(QKeyEvent* event) override;

private:
    void setupAppearance();
    void setupConnections();

    ResultSetModel* m_model;
    DataChangeManager* m_changeManager;
    QUndoStack* m_undoStack;

    QString m_schema;
    QString m_tableName;
    QStringList m_primaryKeyColumns;
};

} // namespace tablepro
```

**Step 2: Create data_grid.cpp**

```cpp
#include "data_grid.hpp"
#include "../core/change_manager.hpp"
#include <QHeaderView>
#include <QKeyEvent>
#include <QMessageBox>
#include <QtConcurrent>

namespace tablepro {

DataGrid::DataGrid(QWidget* parent)
    : QTableView(parent)
    , m_model(new ResultSetModel(this))
    , m_changeManager(new DataChangeManager(this))
    , m_undoStack(new QUndoStack(this))
{
    setModel(m_model);
    setupAppearance();
    setupConnections();
}

DataGrid::~DataGrid() = default;

void DataGrid::setupAppearance() {
    // Enable virtual scrolling
    setUniformItemSizes(true);
    setWordWrap(false);

    // Alternating row colors
    setAlternatingRowColors(true);

    // Selection
    setSelectionBehavior(QAbstractItemView::SelectItems);
    setSelectionMode(QAbstractItemView::ExtendedSelection);

    // Editing
    setEditTriggers(QAbstractItemView::DoubleClicked | QAbstractItemView::EditKeyPressed);

    // Grid
    setShowGrid(true);
    setGridStyle(Qt::SolidLine);

    // Headers
    horizontalHeader()->setStretchLastSection(true);
    horizontalHeader()->setHighlightSections(false);
    horizontalHeader()->setSectionsClickable(true);

    verticalHeader()->setVisible(true);
    verticalHeader()->setDefaultSectionSize(28);

    // Style
    setStyleSheet(R"(
        QTableView {
            background-color: #1E1E2E;
            alternate-background-color: #181825;
            color: #CDD6F4;
            gridline-color: #313244;
            border: none;
            selection-background-color: #45475A;
        }
        QTableView::item {
            padding: 4px 8px;
        }
        QTableView::item:selected {
            background-color: #45475A;
        }
        QHeaderView::section {
            background-color: #181825;
            color: #CDD6F4;
            padding: 8px;
            border: none;
            border-bottom: 1px solid #313244;
            font-weight: bold;
        }
        QHeaderView::section:hover {
            background-color: #313244;
        }
    )");
}

void DataGrid::setupConnections() {
    connect(m_model, &ResultSetModel::dataEdited,
            m_changeManager, &DataChangeManager::trackEdit);

    connect(m_changeManager, &DataChangeManager::changesChanged,
            this, &DataGrid::changesChanged);
}

void DataGrid::setQueryResult(const QueryResult& result) {
    m_model->setQueryResult(result);
    m_changeManager->clear();
}

QueryResult DataGrid::queryResult() const {
    return m_model->queryResult();
}

void DataGrid::setTableName(const QString& schema, const QString& table) {
    m_schema = schema;
    m_tableName = table;
}

void DataGrid::setPrimaryKeyColumns(const QStringList& columns) {
    m_primaryKeyColumns = columns;
    m_changeManager->setPrimaryKeyColumns(columns);
}

bool DataGrid::hasChanges() const {
    return m_changeManager->hasChanges();
}

void DataGrid::discardChanges() {
    m_changeManager->discardChanges();
    m_undoStack->clear();

    // Revert model
    // TODO: Restore original values
}

PendingChanges DataGrid::getPendingChanges() const {
    return m_changeManager->getPendingChanges();
}

void DataGrid::addNewRow() {
    int newRow = m_model->rowCount();

    // Create empty row with column names
    QVariantMap emptyRow;
    for (const auto& colName : m_model->queryResult().columnNames) {
        emptyRow[colName] = QVariant();
    }

    // TODO: Insert into model

    m_model->markRowAsNew(newRow);
    m_changeManager->trackNewRow(emptyRow);

    emit changesChanged(true);
}

void DataGrid::deleteSelectedRows() {
    auto selected = selectionModel()->selectedRows();

    for (const auto& index : selected) {
        int row = index.row();

        if (!m_model->isRowNew(row)) {
            m_model->markRowAsDeleted(row);

            QVariantMap rowData = m_model->queryResult().rows[row];
            m_changeManager->trackDeletedRow(rowData);
        }
    }

    emit changesChanged(hasChanges());
}

QFuture<bool> DataGrid::commitChanges(const QString& connectionId) {
    return QtConcurrent::run([this, connectionId]() -> bool {
        // TODO: Get driver and execute generated SQL
        return true;
    });
}

void DataGrid::keyPressEvent(QKeyEvent* event) {
    if (event->matches(QKeySequence::Undo)) {
        if (m_undoStack->canUndo()) {
            m_undoStack->undo();
        }
        return;
    }

    if (event->matches(QKeySequence::Redo)) {
        if (m_undoStack->canRedo()) {
            m_undoStack->redo();
        }
        return;
    }

    if (event->key() == Qt::Key_Delete) {
        deleteSelectedRows();
        return;
    }

    QTableView::keyPressEvent(event);
}

} // namespace tablepro
```

**Step 3: Commit data grid**

```bash
git add src/ui/grid/data_grid.hpp src/ui/grid/data_grid.cpp
git commit -m "feat: Add DataGrid widget with editing support"
```

---

## Task 3: Change Manager

**Files:**
- Create: `src/core/change_manager.hpp`
- Create: `src/core/change_manager.cpp`

**Step 1: Create change_manager.hpp**

```cpp
#pragma once

#include <QObject>
#include <QMap>
#include <QList>
#include <QVariantMap>
#include <QStringList>

namespace tablepro {

struct CellEdit {
    int row;
    QString columnName;
    QVariant oldValue;
    QVariant newValue;
};

struct PendingChanges {
    QList<CellEdit> editedCells;
    QList<QVariantMap> newRows;
    QList<QVariantMap> deletedRows;

    bool isEmpty() const {
        return editedCells.isEmpty() && newRows.isEmpty() && deletedRows.isEmpty();
    }

    int totalCount() const {
        return editedCells.size() + newRows.size() + deletedRows.size();
    }
};

class DataChangeManager : public QObject {
    Q_OBJECT

public:
    explicit DataChangeManager(QObject* parent = nullptr);

    void trackEdit(int row, int column, const QVariant& oldValue, const QVariant& newValue);
    void trackNewRow(const QVariantMap& rowData);
    void trackDeletedRow(const QVariantMap& rowData);

    PendingChanges getPendingChanges() const;
    bool hasChanges() const;
    int changeCount() const;

    void discardChanges();
    void clear();

    void setPrimaryKeyColumns(const QStringList& columns);
    QStringList primaryKeyColumns() const { return m_primaryKeyColumns; }

signals:
    void changesChanged(bool hasChanges);
    void editTracked(int row, const QString& column);
    void newRowTracked(const QVariantMap& rowData);
    void deleteTracked(const QVariantMap& rowData);

private:
    QMap<QString, CellEdit> m_edits;  // Key: "row:column"
    QList<QVariantMap> m_newRows;
    QList<QVariantMap> m_deletedRows;
    QStringList m_primaryKeyColumns;

    QString editKey(int row, const QString& column) const;
};

} // namespace tablepro
```

**Step 2: Create change_manager.cpp**

```cpp
#include "change_manager.hpp"

namespace tablepro {

DataChangeManager::DataChangeManager(QObject* parent)
    : QObject(parent)
{
}

QString DataChangeManager::editKey(int row, const QString& column) const {
    return QString("%1:%2").arg(row).arg(column);
}

void DataChangeManager::trackEdit(int row, int column, const QVariant& oldValue, const QVariant& newValue) {
    Q_UNUSED(column)

    if (oldValue == newValue) {
        return;
    }

    // Find column name from model would be needed
    // For now, use index as placeholder
    QString columnName = QString("col%1").arg(column);

    QString key = editKey(row, columnName);

    CellEdit edit;
    edit.row = row;
    edit.columnName = columnName;
    edit.oldValue = oldValue;
    edit.newValue = newValue;

    m_edits.insert(key, edit);

    emit editTracked(row, columnName);
    emit changesChanged(hasChanges());
}

void DataChangeManager::trackNewRow(const QVariantMap& rowData) {
    m_newRows.append(rowData);
    emit newRowTracked(rowData);
    emit changesChanged(hasChanges());
}

void DataChangeManager::trackDeletedRow(const QVariantMap& rowData) {
    m_deletedRows.append(rowData);
    emit deleteTracked(rowData);
    emit changesChanged(hasChanges());
}

PendingChanges DataChangeManager::getPendingChanges() const {
    PendingChanges changes;
    changes.editedCells = m_edits.values();
    changes.newRows = m_newRows;
    changes.deletedRows = m_deletedRows;
    return changes;
}

bool DataChangeManager::hasChanges() const {
    return !m_edits.isEmpty() || !m_newRows.isEmpty() || !m_deletedRows.isEmpty();
}

int DataChangeManager::changeCount() const {
    return m_edits.size() + m_newRows.size() + m_deletedRows.size();
}

void DataChangeManager::discardChanges() {
    m_edits.clear();
    m_newRows.clear();
    m_deletedRows.clear();
    emit changesChanged(false);
}

void DataChangeManager::clear() {
    discardChanges();
}

void DataChangeManager::setPrimaryKeyColumns(const QStringList& columns) {
    m_primaryKeyColumns = columns;
}

} // namespace tablepro
```

**Step 3: Commit change manager**

```bash
git add src/core/change_manager.hpp src/core/change_manager.cpp
git commit -m "feat: Add DataChangeManager for tracking edits"
```

---

## Task 4: SQL Generator

**Files:**
- Create: `src/core/sql_generator.hpp`
- Create: `src/core/sql_generator.cpp`

**Step 1: Create sql_generator.hpp**

```cpp
#pragma once

#include <QString>
#include <QStringList>
#include <QVariantMap>
#include "change_manager.hpp"
#include "types.hpp"

namespace tablepro {

class SqlGenerator {
public:
    explicit SqlGenerator(const QString& dialect = "postgresql");

    // Set schema and table
    void setTable(const QString& schema, const QString& table);

    // Generate individual statements
    QString generateUpdate(const CellEdit& edit, const QVariantMap& primaryKeyValues) const;
    QString generateInsert(const QVariantMap& rowData) const;
    QString generateDelete(const QVariantMap& primaryKeyValues) const;

    // Generate batch from pending changes
    QStringList generateFromChanges(
        const PendingChanges& changes,
        const QStringList& allColumns,
        const QList<QVariantMap>& primaryKeyValues
    ) const;

    // Utility
    QString quoteIdentifier(const QString& identifier) const;
    QString escapeValue(const QVariant& value) const;

private:
    QString m_dialect;
    QString m_schema;
    QString m_table;

    QString identifierQuote() const;
};

} // namespace tablepro
```

**Step 2: Create sql_generator.cpp**

```cpp
#include "sql_generator.hpp"

namespace tablepro {

SqlGenerator::SqlGenerator(const QString& dialect)
    : m_dialect(dialect)
{
}

void SqlGenerator::setTable(const QString& schema, const QString& table) {
    m_schema = schema;
    m_table = table;
}

QString SqlGenerator::identifierQuote() const {
    if (m_dialect == "mysql" || m_dialect == "mariadb") {
        return "`";
    }
    return "\"";
}

QString SqlGenerator::quoteIdentifier(const QString& identifier) const {
    QString quote = identifierQuote();
    return quote + identifier + quote;
}

QString SqlGenerator::escapeValue(const QVariant& value) const {
    if (value.isNull()) {
        return "NULL";
    }

    switch (value.userType()) {
        case QMetaType::Int:
        case QMetaType::LongLong:
        case QMetaType::UInt:
        case QMetaType::ULongLong:
            return value.toString();

        case QMetaType::Double:
        case QMetaType::Float:
            return value.toString();

        case QMetaType::Bool:
            return value.toBool() ? "TRUE" : "FALSE";

        default: {
            // String - escape single quotes
            QString str = value.toString();
            str.replace("'", "''");
            return QString("'%1'").arg(str);
        }
    }
}

QString SqlGenerator::generateUpdate(const CellEdit& edit, const QVariantMap& primaryKeyValues) const {
    QStringList setClauses;
    QStringList whereClauses;

    // SET clause
    setClauses.append(QString("%1 = %2")
        .arg(quoteIdentifier(edit.columnName))
        .arg(escapeValue(edit.newValue)));

    // WHERE clause from primary key
    for (auto it = primaryKeyValues.begin(); it != primaryKeyValues.end(); ++it) {
        whereClauses.append(QString("%1 = %2")
            .arg(quoteIdentifier(it.key()))
            .arg(escapeValue(it.value())));
    }

    QString schemaPrefix = m_schema.isEmpty() ? "" : quoteIdentifier(m_schema) + ".";

    return QString("UPDATE %1%2 SET %3 WHERE %4;")
        .arg(schemaPrefix)
        .arg(quoteIdentifier(m_table))
        .arg(setClauses.join(", "))
        .arg(whereClauses.join(" AND "));
}

QString SqlGenerator::generateInsert(const QVariantMap& rowData) const {
    QStringList columns;
    QStringList values;

    for (auto it = rowData.begin(); it != rowData.end(); ++it) {
        if (!it.value().isNull()) {
            columns.append(quoteIdentifier(it.key()));
            values.append(escapeValue(it.value()));
        }
    }

    QString schemaPrefix = m_schema.isEmpty() ? "" : quoteIdentifier(m_schema) + ".";

    return QString("INSERT INTO %1%2 (%3) VALUES (%4);")
        .arg(schemaPrefix)
        .arg(quoteIdentifier(m_table))
        .arg(columns.join(", "))
        .arg(values.join(", "));
}

QString SqlGenerator::generateDelete(const QVariantMap& primaryKeyValues) const {
    QStringList whereClauses;

    for (auto it = primaryKeyValues.begin(); it != primaryKeyValues.end(); ++it) {
        whereClauses.append(QString("%1 = %2")
            .arg(quoteIdentifier(it.key()))
            .arg(escapeValue(it.value())));
    }

    QString schemaPrefix = m_schema.isEmpty() ? "" : quoteIdentifier(m_schema) + ".";

    return QString("DELETE FROM %1%2 WHERE %3;")
        .arg(schemaPrefix)
        .arg(quoteIdentifier(m_table))
        .arg(whereClauses.join(" AND "));
}

QStringList SqlGenerator::generateFromChanges(
    const PendingChanges& changes,
    const QStringList& allColumns,
    const QList<QVariantMap>& primaryKeyValues
) const {
    QStringList statements;

    // Order: DELETEs first, then UPDATEs, then INSERTs
    // This avoids FK conflicts

    // DELETEs
    for (const auto& rowData : changes.deletedRows) {
        QVariantMap pkValues;
        for (const auto& pkCol : m_schema.isEmpty() ? QStringList() : QStringList()) {
            if (rowData.contains(pkCol)) {
                pkValues[pkCol] = rowData[pkCol];
            }
        }
        statements.append(generateDelete(pkValues));
    }

    // UPDATEs
    for (const auto& edit : changes.editedCells) {
        if (edit.row < primaryKeyValues.size()) {
            statements.append(generateUpdate(edit, primaryKeyValues[edit.row]));
        }
    }

    // INSERTs
    for (const auto& rowData : changes.newRows) {
        statements.append(generateInsert(rowData));
    }

    return statements;
}

} // namespace tablepro
```

**Step 3: Commit SQL generator**

```bash
git add src/core/sql_generator.hpp src/core/sql_generator.cpp
git commit -m "feat: Add SqlGenerator for UPDATE/INSERT/DELETE generation"
```

---

## Task 5: Update CMakeLists and Verify

**Files:**
- Modify: `CMakeLists.txt`

**Step 1: Add sources**

```cmake
set(TABLEPRO_SOURCES
    # ... existing ...
    src/ui/grid/result_set_model.cpp
    src/ui/grid/data_grid.cpp
    src/core/change_manager.cpp
    src/core/sql_generator.cpp
)
```

**Step 2: Build and verify**

```bash
cmake --build build/debug -j$(nproc)
```

**Step 3: Commit**

```bash
git add CMakeLists.txt
git commit -m "build: Add data grid sources"
```

---

## Acceptance Criteria

- [ ] ResultSetModel displays query results
- [ ] DataGrid supports inline editing
- [ ] ChangeManager tracks edits/new rows/deletes
- [ ] SqlGenerator produces valid SQL statements
- [ ] Visual indicators for changes
- [ ] Keyboard shortcuts work (Delete, Undo)

---

**Phase 5 Complete.** Next: Phase 6 - SQL Editor