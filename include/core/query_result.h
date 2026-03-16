#pragma once

#include <QVector>
#include <QVariant>
#include <QString>

namespace tablepro {

// Extracting QueryResult definition to separate file to avoid circular dependencies
struct QueryResult {
    QVector<QString> columnNames;
    QVector<QVector<QVariant>> rows;
    int affectedRows = 0;
    bool success = false;
    QString errorMessage;

    // Convenience method to get data
    QVariant getValue(int row, int col) const {
        if (row >= 0 && row < rows.size() && col >= 0 && col < rows[row].size()) {
            return rows[row][col];
        }
        return QVariant();
    }

    int rowCount() const { return rows.size(); }
    int columnCount() const {
        return columnNames.size();
    }
};

} // namespace tablepro