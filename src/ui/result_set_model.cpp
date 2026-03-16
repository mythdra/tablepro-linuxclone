#include "result_set_model.h"

#include <QTextStream>
#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QPair>

namespace tablepro {

ResultSetModel::ResultSetModel(QObject* parent)
    : QAbstractTableModel(parent)
{
}

void ResultSetModel::setQueryResult(const QueryResult& result)
{
    beginResetModel();

    m_rows = result.rows;
    m_columnNames = result.columnNames;

    // Initialize column metadata
    m_columnMeta.clear();
    m_columnMeta.reserve(m_columnNames.size());
    for (const QString& name : m_columnNames) {
        ColumnMeta meta;
        meta.name = name;
        meta.type = "text";  // Default type
        meta.nullable = true;
        meta.editable = true;
        meta.width = kDefaultColumnWidth;
        m_columnMeta.append(meta);
    }

    // Clear change tracking
    m_changes.clear();
    m_changeIndex.clear();
    m_undoStack.clear();
    m_redoStack.clear();

    endResetModel();
}

void ResultSetModel::clear()
{
    beginResetModel();
    m_rows.clear();
    m_columnNames.clear();
    m_columnMeta.clear();
    m_changes.clear();
    m_changeIndex.clear();
    m_undoStack.clear();
    m_redoStack.clear();
    endResetModel();
}

int ResultSetModel::rowCount(const QModelIndex& parent) const
{
    // Table models should return 0 for valid parent indices (Qt convention)
    if (parent.isValid()) {
        return 0;
    }
    return m_rows.size();
}

int ResultSetModel::columnCount(const QModelIndex& parent) const
{
    if (parent.isValid()) {
        return 0;
    }
    return m_columnNames.size();
}

QVariant ResultSetModel::data(const QModelIndex& index, int role) const
{
    if (!index.isValid()) {
        return QVariant();
    }

    const int row = index.row();
    const int col = index.column();

    // Bounds checking (ES.46 - avoid narrowing)
    if (row < 0 || row >= m_rows.size() || col < 0 || col >= m_columnNames.size()) {
        return QVariant();
    }

    switch (role) {
    case Qt::DisplayRole:
    case Qt::EditRole: {
        const QVariant& value = m_rows[row][col];
        if (value.isNull()) {
            return "NULL";
        }
        return value;
    }

    case Qt::ForegroundRole: {
        // Show NULL values in gray
        if (m_rows[row][col].isNull()) {
            return QColor(Qt::gray);
        }
        return QVariant();
    }

    case Qt::BackgroundRole: {
        // Highlight changed cells
        if (isCellChanged(row, col)) {
            return QColor(255, 255, 200);  // Light yellow
        }
        return QVariant();
    }

    case Qt::ToolTipRole: {
        const QVariant& value = m_rows[row][col];
        if (value.isNull()) {
            return tr("NULL value");
        }
        // Show full content for long values
        const QString text = value.toString();
        if (text.length() > 50) {
            return text;
        }
        return QVariant();
    }

    case Qt::TextAlignmentRole: {
        // Right-align numbers
        const QVariant& value = m_rows[row][col];
        if (value.typeId() == QMetaType::Type::Int ||
            value.typeId() == QMetaType::Type::Double ||
            value.typeId() == QMetaType::Type::LongLong) {
            return Qt::AlignRight | Qt::AlignVCenter;
        }
        return Qt::AlignLeft | Qt::AlignVCenter;
    }

    default:
        return QVariant();
    }
}

QVariant ResultSetModel::headerData(int section, Qt::Orientation orientation, int role) const
{
    if (role != Qt::DisplayRole) {
        return QVariant();
    }

    if (orientation == Qt::Horizontal) {
        if (section >= 0 && section < m_columnNames.size()) {
            return m_columnNames[section];
        }
    }

    // Row numbers for vertical header
    if (orientation == Qt::Vertical) {
        return section + 1;  // 1-based row numbers
    }

    return QVariant();
}

Qt::ItemFlags ResultSetModel::flags(const QModelIndex& index) const
{
    if (!index.isValid()) {
        return Qt::NoItemFlags;
    }

    Qt::ItemFlags flags = QAbstractTableModel::flags(index);

    // Add editable flag if column is editable
    const int col = index.column();
    if (col >= 0 && col < m_columnMeta.size() && m_columnMeta[col].editable) {
        flags |= Qt::ItemIsEditable;
    }

    return flags;
}

bool ResultSetModel::setData(const QModelIndex& index, const QVariant& value, int role)
{
    if (!index.isValid() || role != Qt::EditRole) {
        return false;
    }

    const int row = index.row();
    const int col = index.column();

    if (row < 0 || row >= m_rows.size() || col < 0 || col >= m_columnNames.size()) {
        return false;
    }

    if (!m_columnMeta[col].editable) {
        return false;
    }

    const QVariant oldValue = m_rows[row][col];

    // No actual change
    if (oldValue == value) {
        return false;
    }

    // Record the change
    CellChange change;
    change.row = row;
    change.column = col;
    change.oldValue = oldValue;
    change.newValue = value;

    // Check if this cell already has a change
    const auto key = qMakePair(row, col);
    if (m_changeIndex.contains(key)) {
        // Update existing change - don't record intermediate states
        const int idx = m_changeIndex[key];
        m_changes[idx].newValue = value;
    } else {
        // New change
        m_changeIndex[key] = m_changes.size();
        m_changes.append(change);
    }

    // Clear redo stack (new change invalidates redo)
    m_redoStack.clear();

    // Add to undo stack
    m_undoStack.append(change);
    if (m_undoStack.size() > kMaxUndoSteps) {
        m_undoStack.removeFirst();
    }

    // Update data
    m_rows[row][col] = value;

    emit dataChanged(index, index, {Qt::DisplayRole, Qt::EditRole, Qt::BackgroundRole});
    return true;
}

void ResultSetModel::setColumnMeta(int column, const ColumnMeta& meta)
{
    if (column >= 0 && column < m_columnMeta.size()) {
        m_columnMeta[column] = meta;
        emit headerDataChanged(Qt::Horizontal, column, column);
    }
}

ColumnMeta ResultSetModel::columnMeta(int column) const
{
    if (column >= 0 && column < m_columnMeta.size()) {
        return m_columnMeta[column];
    }
    return ColumnMeta();
}

void ResultSetModel::setColumnEditable(int column, bool editable)
{
    if (column >= 0 && column < m_columnMeta.size()) {
        m_columnMeta[column].editable = editable;
    }
}

void ResultSetModel::setColumnWidth(int column, int width)
{
    if (column >= 0 && column < m_columnMeta.size()) {
        m_columnMeta[column].width = width;
    }
}

void ResultSetModel::clearChanges()
{
    m_changes.clear();
    m_changeIndex.clear();
    m_undoStack.clear();
    m_redoStack.clear();

    // Refresh all cells to remove highlight
    if (!m_rows.isEmpty() && !m_columnNames.isEmpty()) {
        emit dataChanged(index(0, 0), index(m_rows.size() - 1, m_columnNames.size() - 1),
                        {Qt::BackgroundRole});
    }
}

bool ResultSetModel::isCellChanged(int row, int column) const
{
    return m_changeIndex.contains(qMakePair(row, column));
}

bool ResultSetModel::canUndo() const
{
    return !m_undoStack.isEmpty();
}

bool ResultSetModel::canRedo() const
{
    return !m_redoStack.isEmpty();
}

void ResultSetModel::undo()
{
    if (m_undoStack.isEmpty()) {
        return;
    }

    const CellChange change = m_undoStack.takeLast();

    // Restore old value
    m_rows[change.row][change.column] = change.oldValue;

    // Move to redo stack
    m_redoStack.append(change);

    // Update change tracking
    const auto key = qMakePair(change.row, change.column);
    if (change.oldValue == m_rows[change.row][change.column]) {
        // If restored to original, remove from changes
        if (m_changeIndex.contains(key)) {
            const int idx = m_changeIndex[key];
            m_changes.remove(idx);
            m_changeIndex.remove(key);
            // Rebuild index
            m_changeIndex.clear();
            for (int i = 0; i < m_changes.size(); ++i) {
                m_changeIndex[qMakePair(m_changes[i].row, m_changes[i].column)] = i;
            }
        }
    }

    const QModelIndex idx = index(change.row, change.column);
    emit dataChanged(idx, idx, {Qt::DisplayRole, Qt::EditRole, Qt::BackgroundRole});
}

void ResultSetModel::redo()
{
    if (m_redoStack.isEmpty()) {
        return;
    }

    const CellChange change = m_redoStack.takeLast();

    // Restore new value
    m_rows[change.row][change.column] = change.newValue;

    // Move back to undo stack
    m_undoStack.append(change);

    // Update change tracking
    const auto key = qMakePair(change.row, change.column);
    if (!m_changeIndex.contains(key)) {
        m_changeIndex[key] = m_changes.size();
        m_changes.append(change);
    }

    const QModelIndex idx = index(change.row, change.column);
    emit dataChanged(idx, idx, {Qt::DisplayRole, Qt::EditRole, Qt::BackgroundRole});
}

QString ResultSetModel::toCsv() const
{
    QString output;
    QTextStream stream(&output);

    // Header row
    for (int col = 0; col < m_columnNames.size(); ++col) {
        if (col > 0) stream << ",";
        // Quote and escape
        QString name = m_columnNames[col];
        name.replace("\"", "\"\"");
        stream << "\"" << name << "\"";
    }
    stream << "\n";

    // Data rows
    for (const auto& row : m_rows) {
        for (int col = 0; col < row.size(); ++col) {
            if (col > 0) stream << ",";

            const QVariant& value = row[col];
            if (value.isNull()) {
                // NULL representation
                stream << "NULL";
            } else {
                QString text = value.toString();
                text.replace("\"", "\"\"");
                stream << "\"" << text << "\"";
            }
        }
        stream << "\n";
    }

    stream.flush();
    return output;
}

QString ResultSetModel::toJson() const
{
    QJsonArray rowsArray;

    for (const auto& row : m_rows) {
        QJsonObject rowObj;
        for (int col = 0; col < m_columnNames.size() && col < row.size(); ++col) {
            const QVariant& value = row[col];
            const QString& key = m_columnNames[col];

            if (value.isNull()) {
                rowObj[key] = QJsonValue::Null;
            } else {
                // Try to convert to appropriate JSON type
                switch (value.typeId()) {
                case QMetaType::Type::Bool:
                    rowObj[key] = value.toBool();
                    break;
                case QMetaType::Type::Int:
                case QMetaType::Type::LongLong:
                    rowObj[key] = value.toLongLong();
                    break;
                case QMetaType::Type::Double:
                case QMetaType::Type::Float:
                    rowObj[key] = value.toDouble();
                    break;
                default:
                    rowObj[key] = value.toString();
                    break;
                }
            }
        }
        rowsArray.append(rowObj);
    }

    QJsonDocument doc(rowsArray);
    return QString::fromUtf8(doc.toJson(QJsonDocument::Indented));
}

} // namespace tablepro