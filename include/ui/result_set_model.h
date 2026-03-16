#pragma once

#include <QAbstractTableModel>
#include <QVector>
#include <QVariant>
#include <QString>
#include <QStringList>
#include <QHash>
#include <QColor>
#include <memory>
#include "core/query_result.h"

namespace tablepro {

// Column metadata for result set
struct ColumnMeta {
    QString name;
    QString type;
    bool nullable = true;
    bool editable = true;
    int width = -1;  // -1 = auto
};

// Cell change record for tracking edits
struct CellChange {
    int row;
    int column;
    QVariant oldValue;
    QVariant newValue;
};

/**
 * Custom model for displaying database query results.
 * Follows C++ Core Guidelines for memory management and type safety.
 */
class ResultSetModel : public QAbstractTableModel
{
    Q_OBJECT

public:
    explicit ResultSetModel(QObject* parent = nullptr);
    ~ResultSetModel() override = default;

    // Disable copy (C.67 - polymorphic class)
    ResultSetModel(const ResultSetModel&) = delete;
    ResultSetModel& operator=(const ResultSetModel&) = delete;

    // Allow move
    ResultSetModel(ResultSetModel&&) noexcept = default;
    ResultSetModel& operator=(ResultSetModel&&) noexcept = default;

    // QAbstractTableModel interface
    int rowCount(const QModelIndex& parent = QModelIndex()) const override;
    int columnCount(const QModelIndex& parent = QModelIndex()) const override;
    QVariant data(const QModelIndex& index, int role = Qt::DisplayRole) const override;
    QVariant headerData(int section, Qt::Orientation orientation, int role = Qt::DisplayRole) const override;
    Qt::ItemFlags flags(const QModelIndex& index) const override;
    bool setData(const QModelIndex& index, const QVariant& value, int role = Qt::EditRole) override;

    // Data management
    void setQueryResult(const QueryResult& result);
    void clear();

    // Column management
    void setColumnMeta(int column, const ColumnMeta& meta);
    ColumnMeta columnMeta(int column) const;
    void setColumnEditable(int column, bool editable);
    void setColumnWidth(int column, int width);

    // Change tracking
    bool hasChanges() const { return !m_changes.isEmpty(); }
    QVector<CellChange> changes() const { return m_changes; }
    void clearChanges();
    bool isCellChanged(int row, int column) const;

    // Undo/Redo
    bool canUndo() const;
    bool canRedo() const;
    void undo();
    void redo();

    // Export helpers
    QString toCsv() const;
    QString toJson() const;

    // Direct data access (const only - Con.3)
    const QVector<QVector<QVariant>>& rows() const { return m_rows; }
    const QStringList& columnNames() const { return m_columnNames; }

private:
    // Data storage (R.1 - RAII via Qt containers)
    QVector<QVector<QVariant>> m_rows;
    QStringList m_columnNames;
    QVector<ColumnMeta> m_columnMeta;

    // Change tracking
    QVector<CellChange> m_changes;
    QHash<QPair<int, int>, int> m_changeIndex;  // For fast lookup

    // Undo/Redo stacks
    QVector<CellChange> m_undoStack;
    QVector<CellChange> m_redoStack;

    // Constants (Con.5)
    static constexpr int kDefaultColumnWidth = 100;
    static constexpr int kMaxUndoSteps = 100;
};

} // namespace tablepro