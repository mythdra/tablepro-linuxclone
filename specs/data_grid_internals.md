# Data Grid Internals (QTableView + Qt Model/View)

## Overview
The Data Grid uses Qt's **Model/View architecture** with `QTableView` and a custom `QAbstractTableModel` subclass. This provides native virtual scrolling and can handle millions of rows efficiently.

## 1. Virtual Scrolling

```cpp
// src/ui/DataGrid/QueryResultModel.hpp
#pragma once

#include <QAbstractTableModel>
#include <QVariantList>
#include "core/QueryResult.hpp"

namespace tablepro {

class QueryResultModel : public QAbstractTableModel {
    Q_OBJECT

public:
    explicit QueryResultModel(QObject* parent = nullptr);

    // QAbstractItemModel interface
    int rowCount(const QModelIndex& parent = QModelIndex()) const override;
    int columnCount(const QModelIndex& parent = QModelIndex()) const override;
    QVariant data(const QModelIndex& index, int role) const override;
    QVariant headerData(int section, Qt::Orientation orientation, int role) const override;

    // Editing support
    Qt::ItemFlags flags(const QModelIndex& index) const override;
    bool setData(const QModelIndex& index, const QVariant& value, int role) override;

    // For large datasets (pagination/virtual loading)
    bool canFetchMore(const QModelIndex& parent) const override;
    void fetchMore(const QModelIndex& parent) override;

    // Public API
    void setQueryResult(const QueryResult& result);
    void clear();

    // Change tracking
    void markCellEdited(int row, int column);
    void markRowInserted(int row);
    void markRowDeleted(int row);
    QSet<int> editedRows() const { return m_editedRows; }
    QSet<int> insertedRows() const { return m_insertedRows; }
    QSet<int> deletedRows() const { return m_deletedRows; }

signals:
    void dataRequested(int offset, int limit);  // For server-side pagination

private:
    QueryResult m_result;
    int m_fetchedRows{0};
    bool m_hasMoreData{false};

    // Change tracking
    QSet<int> m_editedRows;
    QSet<int> m_insertedRows;
    QSet<int> m_deletedRows;
    QMap<QPair<int, int>, QVariant> m_originalValues;  // For undo
};

} // namespace tablepro
```

```cpp
// src/ui/DataGrid/QueryResultModel.cpp
#include "QueryResultModel.hpp"

int QueryResultModel::rowCount(const QModelIndex& parent) const {
    return parent.isValid() ? 0 : m_result.rows.count();
}

int QueryResultModel::columnCount(const QModelIndex& parent) const {
    return parent.isValid() ? 0 : m_result.columns.count();
}

QVariant QueryResultModel::data(const QModelIndex& index, int role) const {
    if (!index.isValid() || m_result.rows.isEmpty())
        return {};

    int row = index.row();
    int col = index.column();

    switch (role) {
        case Qt::DisplayRole:
        case Qt::EditRole: {
            QVariant value = m_result.rows.at(row).toList().at(col);

            // NULL handling
            if (value.isNull() || !value.isValid()) {
                return role == Qt::DisplayRole ? QStringLiteral("NULL") : QVariant();
            }

            return value;
        }

        case Qt::BackgroundRole:
            // Visual deltas
            if (m_deletedRows.contains(row))
                return QColor(254, 226, 226);  // Light red
            if (m_insertedRows.contains(row))
                return QColor(209, 250, 229);  // Light green
            if (m_editedRows.contains(row))
                return QColor(254, 243, 199);  // Light yellow
            return {};

        case Qt::ForegroundRole:
            if (m_deletedRows.contains(row))
                return QColor(156, 163, 175);  // Gray
            if (m_result.rows.at(row).toList().at(col).isNull())
                return QColor(107, 114, 128);  // Gray for NULL
            return {};

        case Qt::FontRole:
            if (m_deletedRows.contains(row)) {
                QFont font;
                font.setStrikeOut(true);
                return font;
            }
            return {};

        case Qt::TextAlignmentRole:
            // Right-align numbers
            if (col < m_result.columns.count()) {
                QString type = m_result.columns.at(col).type.toLower();
                if (type.contains("int") || type.contains("real") ||
                    type.contains("float") || type.contains("decimal")) {
                    return Qt::AlignRight | Qt::AlignVCenter;
                }
            }
            return Qt::AlignLeft | Qt::AlignVCenter;

        default:
            return {};
    }
}

QVariant QueryResultModel::headerData(
    int section,
    Qt::Orientation orientation,
    int role) const
{
    if (orientation == Qt::Horizontal && role == Qt::DisplayRole) {
        if (section < m_result.columns.count())
            return m_result.columns.at(section).name;
    }
    return {};
}

bool QueryResultModel::setData(
    const QModelIndex& index,
    const QVariant& value,
    int role)
{
    if (role != Qt::EditRole || !index.isValid())
        return false;

    // Store original value for undo
    if (!m_originalValues.contains({index.row(), index.column()})) {
        m_originalValues[{index.row(), index.column()}] =
            data(index, Qt::EditRole);
    }

    markCellEdited(index.row(), index.column());

    emit dataChanged(index, index, {Qt::DisplayRole, Qt::BackgroundRole});
    return true;
}
```

## 2. Column Definition

```cpp
// Column headers are handled by QTableView header
// Customization via header delegate:

class GridHeaderView : public QHeaderView {
    Q_OBJECT

public:
    explicit GridHeaderView(Qt::Orientation orientation, QWidget* parent = nullptr);

signals:
    void sortRequested(int logicalIndex, Qt::SortOrder order);

protected:
    void mousePressEvent(QMouseEvent* event) override;

private:
    int m_sortSection{-1};
    Qt::SortOrder m_sortOrder{Qt::AscendingOrder};
};

// Usage:
auto* header = new GridHeaderView(Qt::Horizontal, tableView);
connect(header, &GridHeaderView::sortRequested,
        this, [=](int col, Qt::SortOrder order) {
    // Re-execute query with ORDER BY
    emit sortRequested(m_result.columns[col].name,
                       order == Qt::AscendingOrder ? "ASC" : "DESC");
});
```

## 3. Sorting

- Click column header → `sortRequested` signal emitted
- C++ rebuilds query with `ORDER BY` and re-executes
- Grid refreshes with new data (not client-side sort)

```cpp
void QueryResultModel::setSortOrder(const QString& column, const QString& direction) {
    // This triggers a full model refresh from the backend
    emit sortRequested(column, direction);
}
```

## 4. Inline Editing

- Double-click cell → QTableView activates item delegate editor
- On cell value change: `setData()` called → `DataChangeManager` tracks delta
- Cell background updates to yellow indicating "pending change"

```cpp
// Enable editing for specific columns
Qt::ItemFlags QueryResultModel::flags(const QModelIndex& index) const {
    Qt::ItemFlags defaultFlags = QAbstractTableModel::flags(index);

    if (!index.isValid())
        return defaultFlags;

    // Make editable unless it's a primary key or deleted row
    if (!m_result.columns.at(index.column()).isPrimaryKey &&
        !m_deletedRows.contains(index.row())) {
        return defaultFlags | Qt::ItemIsEditable;
    }

    return defaultFlags;
}
```

## 5. Visual Deltas

```cpp
// Handled in data() method - see above
// Colors:
// - Yellow (#FEF3C7): Edited cell
// - Green (#D1FAE5): New row
// - Red (#FEE2E2) + strikethrough: Deleted row
```

## 6. Row Operations

| Action | Trigger | Backend Call |
|--------|---------|--------------|
| Add Row | "+" button in status bar | `DataChangeManager::insertRow(tabID)` |
| Delete Row | Select + Delete key | `DataChangeManager::deleteRow(tabID, rowIndices)` |
| Duplicate Row | Ctrl+D | `DataChangeManager::duplicateRow(tabID, rowIndex)` |
| Commit | Ctrl+S | `DataChangeManager::commit(tabID)` → SQL execution |
| Discard | Discard button | `DataChangeManager::discard(tabID)` → restore originals |

## 7. Pagination

```cpp
// Status bar widget shows: Rows 1-500 of 12,345 | 0.045s
class GridStatusBar : public QWidget {
    Q_OBJECT

public:
    void setRowCount(qint64 total, int offset, int limit);
    void setExecutionTime(double seconds);

signals:
    void firstPage();
    void previousPage();
    void nextPage();
    void lastPage();
    void goToPage(int page);
    void changePageSize(int newSize);

private:
    QLabel* m_rowCountLabel;
    QLabel* m_executionTimeLabel;
    QPushButton* m_firstBtn;
    QPushButton* m_prevBtn;
    QPushButton* m_nextBtn;
    QPushButton* m_lastBtn;
    QSpinBox* m_pageSpinBox;
    QComboBox* m_pageSizeCombo;
};
```

- Next/Previous page buttons call `QueryManager::executeWithOffset(tabID, newOffset)`
- C++ modifies query with `LIMIT {pageSize} OFFSET {newOffset}`
- Model replaced with new data via `setQueryResult()`

## 8. Copy Operations

```cpp
// src/ui/DataGrid/GridActions.hpp
#pragma once

#include <QApplication>
#include <QClipboard>
#include <QModelIndexList>

namespace tablepro {

class GridCopyActions {
public:
    static void copyCell(QTableView* view, const QModelIndex& index) {
        QVariant value = view->model()->data(index, Qt::DisplayRole);
        QApplication::clipboard()->setText(value.toString());
    }

    static void copyRow(QTableView* view, int row) {
        QStringList values;
        for (int col = 0; col < view->model()->columnCount(); ++col) {
            values << view->model()->data(view->model()->index(row, col),
                                          Qt::DisplayRole).toString();
        }
        QApplication::clipboard()->setText(values.join('\t'));
    }

    static void copyColumn(QTableView* view, int column) {
        QStringList values;
        for (int row = 0; row < view->model()->rowCount(); ++row) {
            values << view->model()->data(view->model()->index(row, column),
                                          Qt::DisplayRole).toString();
        }
        QApplication::clipboard()->setText(values.join('\n'));
    }

    static void copyAsInsert(QTableView* view, const QModelIndexList& selection,
                             const QString& tableName) {
        // Generate INSERT statements for selected rows
        QString sql = generateInsertSQL(view, selection, tableName);
        QApplication::clipboard()->setText(sql);
    }
};

} // namespace tablepro
```

## 9. NULL Handling

```cpp
// Custom item delegate for NULL rendering
class NullAwareItemDelegate : public QStyledItemDelegate {
public:
    void paint(QPainter* painter,
               const QStyleOptionViewItem& option,
               const QModelIndex& index) const override {
        QVariant value = index.data(Qt::DisplayRole);

        if (value.isNull() || !value.isValid()) {
            // Draw distinctive NULL style
            painter->save();
            painter->setPen(QColor(128, 128, 128));
            painter->setFont(QFont("Helvetica", 10, QFont::Italic));
            painter->drawText(option.rect, Qt::AlignCenter, "NULL");
            painter->restore();
        } else {
            QStyledItemDelegate::paint(painter, option, index);
        }
    }
};

// Usage:
tableView->setItemDelegate(new NullAwareItemDelegate(tableView));
```

- NULL values rendered with distinctive style: italic gray "NULL" text
- Editing a NULL cell: empty string saves as empty string, special "Set NULL" checkbox restores NULL

## Performance

| Aspect | Performance |
|--------|-------------|
| 10K rows | Instant (virtual scroll renders ~50 DOM rows) |
| 100K rows | Smooth with pagination |
| 1M+ rows | Server-side row model (fetch on demand) |
| Memory | Only visible rows in memory |
