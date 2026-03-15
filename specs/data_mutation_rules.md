# Data Mutation Rules (C++20 + Qt 6)

## Overview
When users edit cells in the `QTableView`, changes are tracked by the C++ `DataChangeManager` and compiled into dialect-specific SQL statements on commit.

## Change Tracking Flow
```
Qt Frontend (QTableView)        C++ Backend
─────────────────────           ───────────
User edits cell ──────────→  DataChangeManager::updateCell(tabId, rowIdx, colName, newVal)
User adds row   ──────────→  DataChangeManager::insertRow(tabId, rowData)
User deletes row ─────────→  DataChangeManager::deleteRow(tabId, rowIdentity)
User presses Ctrl+S ──────→  DataChangeManager::commit(tabId)
                                ↓
                            SqlDialect::generateSql(changes, dialect)
                                ↓
                            DatabaseDriver::execute(statements) in transaction
                                ↓
                            emit dataSaved(tabId) → QTableView refreshes
```

## Cell Change Tracking
```cpp
// core/DataChangeManager.hpp
class DataChangeManager : public QObject {
    Q_OBJECT

public:
    struct CellChange {
        int rowIndex;
        QString column;
        QVariant originalValue;
        QVariant newValue;
        QVariantMap primaryKey;  // Identity for WHERE clause
    };

    struct RowInsertion {
        QVariantMap rowData;
        bool isPending = true;
    };

    struct RowDeletion {
        int rowIndex;
        QVariantMap primaryKey;
    };

    // Public API
    Q_INVOKABLE void updateCell(const QUuid& tabId, int row, const QString& column, const QVariant& value);
    Q_INVOKABLE void insertRow(const QUuid& tabId, const QVariantMap& rowData);
    Q_INVOKABLE void deleteRow(const QUuid& tabId, int row, const QVariantMap& primaryKey);
    Q_INVOKABLE QFuture<bool> commit(const QUuid& tabId);
    Q_INVOKABLE void discard(const QUuid& tabId);
    Q_INVOKABLE void undo(const QUuid& tabId);
    Q_INVOKABLE void redo(const QUuid& tabId);

    // State inspection
    Q_INVOKABLE bool hasPendingChanges(const QUuid& tabId) const;
    Q_INVOKABLE int pendingChangeCount(const QUuid& tabId) const;
    Q_INVOKABLE QList<CellChange> getCellChanges(const QUuid& tabId) const;

signals:
    void changesCommitted(const QUuid& tabId, int successCount, int failureCount);
    void changesDiscarded(const QUuid& tabId);
    void undoRedoStateChanged(const QUuid& tabId);
    void cellUpdated(const QUuid& tabId, int row, const QString& column);

private:
    QMutex m_mutex;
    QMap<QUuid, QList<CellChange>> m_changes;
    QMap<QUuid, QList<RowInsertion>> m_insertedRows;
    QMap<QUuid, QList<RowDeletion>> m_deletedRows;
    QMap<QUuid, QVector<ChangeAction>> m_undoStack;
    QMap<QUuid, QVector<ChangeAction>> m_redoStack;

    // Helper methods
    QSqlField generateUpdateStatement(const CellChange& change, const QString& tableName);
    QSqlField generateInsertStatement(const RowInsertion& insertion, const QString& tableName);
    QSqlField generateDeleteStatement(const RowDeletion& deletion, const QString& tableName);
};
```

```cpp
// core/DataChangeManager.cpp
void DataChangeManager::updateCell(
    const QUuid& tabId,
    int row,
    const QString& column,
    const QVariant& value)
{
    QMutexLocker locker(&m_mutex);

    auto& changes = m_changes[tabId];

    // Check if this cell was already changed
    auto it = std::find_if(changes.begin(), changes.end(),
        [row, column](const CellChange& c) {
            return c.rowIndex == row && c.column == column;
        });

    if (it == changes.end()) {
        // New change - need to fetch original value from model
        CellChange change;
        change.rowIndex = row;
        change.column = column;
        change.originalValue = m_models[tabId]->data(row, column);  // From cached model
        change.newValue = value;
        change.primaryKey = m_models[tabId]->getPrimaryKey(row);
        changes.append(change);
    } else {
        // Update existing change
        it->newValue = value;

        // If reverted to original, remove the change
        if (it->originalValue == value) {
            changes.erase(it);
        }
    }

    emit cellUpdated(tabId, row, column);
}

void DataChangeManager::insertRow(const QUuid& tabId, const QVariantMap& rowData) {
    QMutexLocker locker(&m_mutex);
    m_insertedRows[tabId].append({rowData, true});
}

void DataChangeManager::deleteRow(const QUuid& tabId, int row, const QVariantMap& primaryKey) {
    QMutexLocker locker(&m_mutex);
    m_deletedRows[tabId].append({row, primaryKey});
}
```

## SQL Generation (SqlDialect)

### UPDATE Statements
```cpp
// core/SqlDialect.cpp
QString SqlDialect::generateUpdateSql(const CellChange& change, const QString& tableName) {
    QString sql;
    QVariantList bindings;

    switch (m_dialectType) {
        case DialectType::PostgreSQL:
            // UPDATE "schema"."table" SET "column" = $1 WHERE "pk_col" = $2
            sql = QString("UPDATE %1 SET %2 = %3 WHERE %4")
                .arg(quoteIdentifier(tableName))
                .arg(quoteIdentifier(change.column))
                .arg(nextPlaceholder())
                .arg(buildWhereClause(change.primaryKey));
            bindings << change.newValue << change.primaryKey.values();
            break;

        case DialectType::MySQL:
            // UPDATE `schema`.`table` SET `column` = ? WHERE `pk_col` = ?
            sql = QString("UPDATE %1 SET %1 = %2 WHERE %3")
                .arg(quoteIdentifier(tableName))
                .arg("?")
                .arg(buildWhereClause(change.primaryKey));
            bindings << change.newValue << change.primaryKey.values();
            break;

        case DialectType::SQLite:
            // Same as PostgreSQL but with ? placeholders
            sql = QString("UPDATE %1 SET %2 = ? WHERE %3")
                .arg(quoteIdentifier(tableName))
                .arg(quoteIdentifier(change.column))
                .arg(buildWhereClause(change.primaryKey));
            bindings << change.newValue << change.primaryKey.values();
            break;

        case DialectType::SqlServer:
            // UPDATE [schema].[table] SET [column] = @p1 WHERE [pk_col] = @p2
            sql = QString("UPDATE %1 SET %2 = %3 WHERE %4")
                .arg(quoteIdentifier(tableName))
                .arg(quoteIdentifier(change.column))
                .arg(nextPlaceholder(PlaceholderStyle::AtSymbol))
                .arg(buildWhereClause(change.primaryKey));
            bindings << change.newValue << change.primaryKey.values();
            break;
    }

    return sql;
}
```

### INSERT Statements
```cpp
QString SqlDialect::generateInsertSql(const RowInsertion& insertion, const QString& tableName, const QStringList& columns) {
    QString sql;
    QVariantList bindings;

    switch (m_dialectType) {
        case DialectType::PostgreSQL:
            // INSERT INTO "table" ("col1", "col2") VALUES ($1, $2) RETURNING *
            sql = QString("INSERT INTO %1 (%2) VALUES (%3) RETURNING *")
                .arg(quoteIdentifier(tableName))
                .arg(columns.join(", "))
                .arg(generatePlaceholders(columns.size()));
            for (const auto& col : columns) {
                bindings << insertion.rowData[col];
            }
            break;

        case DialectType::MySQL:
            // INSERT INTO `table` (`col1`, `col2`) VALUES (?, ?)
            sql = QString("INSERT INTO %1 (%2) VALUES (%3)")
                .arg(quoteIdentifier(tableName))
                .arg(columns.join(", "))
                .arg(generatePlaceholders(columns.size(), PlaceholderStyle::QuestionMark));
            for (const auto& col : columns) {
                bindings << insertion.rowData[col];
            }
            break;

        case DialectType::SQLite:
            // Same as MySQL
            sql = QString("INSERT INTO %1 (%2) VALUES (%3)")
                .arg(quoteIdentifier(tableName))
                .arg(columns.join(", "))
                .arg(generatePlaceholders(columns.size(), PlaceholderStyle::QuestionMark));
            break;

        case DialectType::SqlServer:
            // INSERT INTO [table] ([col1], [col2]) VALUES (@p1, @p2)
            sql = QString("INSERT INTO %1 (%2) VALUES (%3)")
                .arg(quoteIdentifier(tableName))
                .arg(columns.join(", "))
                .arg(generatePlaceholders(columns.size(), PlaceholderStyle::AtSymbol));
            for (const auto& col : columns) {
                bindings << insertion.rowData[col];
            }
            break;
    }

    return sql;
}
```

### DELETE Statements
```cpp
QString SqlDialect::generateDeleteSql(const RowDeletion& deletion, const QString& tableName) {
    QString sql;
    QVariantList bindings;

    // DELETE FROM "table" WHERE "pk_col" = $1
    sql = QString("DELETE FROM %1 WHERE %2")
        .arg(quoteIdentifier(tableName))
        .arg(buildWhereClause(deletion.primaryKey));

    for (const auto& value : deletion.primaryKey.values()) {
        bindings << value;
    }

    return sql;
}
```

## Commit Execution
```cpp
QFuture<bool> DataChangeManager::commit(const QUuid& tabId) {
    return QtConcurrent::run([=]() {
        QMutexLocker locker(&m_mutex);

        auto* driver = ConnectionManager::instance()->getDriver(m_tabSessions[tabId].connectionId);
        if (!driver) {
            return false;
        }

        // Begin transaction
        if (!driver->beginTransaction()) {
            return false;
        }

        int successCount = 0;
        int failureCount = 0;

        // Process cell changes
        const auto& changes = m_changes[tabId];
        for (const auto& change : changes) {
            QString sql = generateUpdateSql(change, change.tableName);
            if (driver->execute(sql, change.bindings())) {
                successCount++;
            } else {
                failureCount++;
                qWarning() << "Failed to update cell:" << driver->lastError();
            }
        }

        // Process inserted rows
        const auto& insertions = m_insertedRows[tabId];
        for (const auto& insertion : insertions) {
            QString sql = generateInsertSql(insertion, insertion.tableName, insertion.rowData.keys());
            if (driver->execute(sql, insertion.bindings())) {
                successCount++;
            } else {
                failureCount++;
            }
        }

        // Process deleted rows
        const auto& deletions = m_deletedRows[tabId];
        for (const auto& deletion : deletions) {
            QString sql = generateDeleteSql(deletion, deletion.tableName);
            if (driver->execute(sql, deletion.bindings())) {
                successCount++;
            } else {
                failureCount++;
            }
        }

        // Commit or rollback
        if (failureCount == 0) {
            if (driver->commitTransaction()) {
                // Clear changes on success
                m_changes[tabId].clear();
                m_insertedRows[tabId].clear();
                m_deletedRows[tabId].clear();
                m_undoStack[tabId].clear();
                emit changesCommitted(tabId, successCount, failureCount);
                return true;
            }
        }

        driver->rollbackTransaction();
        emit changesCommitted(tabId, successCount, failureCount);
        return false;
    });
}
```

## Safe Mode Enforcement
```cpp
// core/SafeModeChecker.hpp
class SafeModeChecker : public QObject {
    Q_OBJECT
    Q_GADGET

public:
    enum class Level {
        Off = 0,          // No restrictions
        RequireWhere = 1, // Block UPDATE/DELETE without WHERE
        ReadOnly = 2      // Block all mutations
    };
    Q_ENUM(Level)

    static bool canExecuteMutation(Level level, const QString& sql);
    static bool hasWhereClause(const QString& sql);
    static QStringList extractTablesFromSql(const QString& sql);
};

// core/SafeModeChecker.cpp
bool SafeModeChecker::canExecuteMutation(Level level, const QString& sql) {
    if (level == Level::ReadOnly) {
        return false;  // Block all mutations
    }

    if (level == Level::RequireWhere) {
        QString trimmed = sql.trimmed().toUpper();
        if (trimmed.startsWith("UPDATE") || trimmed.startsWith("DELETE")) {
            return hasWhereClause(sql);
        }
    }

    return true;  // SafeMode::Off allows everything
}

bool SafeModeChecker::hasWhereClause(const QString& sql) {
    // Simple check - production would use proper SQL parser
    return sql.contains("WHERE", Qt::CaseInsensitive);
}
```

## Undo/Redo
```cpp
// core/DataChangeManager.cpp
void DataChangeManager::undo(const QUuid& tabId) {
    QMutexLocker locker(&m_mutex);

    auto& undoStack = m_undoStack[tabId];
    auto& redoStack = m_redoStack[tabId];

    if (undoStack.isEmpty()) {
        return;
    }

    ChangeAction action = undoStack.takeLast();
    action.undo();  // Execute reverse operation
    redoStack.append(action);

    emit undoRedoStateChanged(tabId);
}

void DataChangeManager::redo(const QUuid& tabId) {
    QMutexLocker locker(&m_mutex);

    auto& undoStack = m_undoStack[tabId];
    auto& redoStack = m_redoStack[tabId];

    if (redoStack.isEmpty()) {
        return;
    }

    ChangeAction action = redoStack.takeLast();
    action.redo();  // Re-execute original operation
    undoStack.append(action);

    emit undoRedoStateChanged(tabId);
}

// ChangeAction records both forward and reverse operations
struct ChangeAction {
    enum class Type { Update, Insert, Delete };
    Type type;
    QString tableName;
    QVariantMap primaryKey;
    QString column;
    QVariant oldValue;
    QVariant newValue;

    void undo() {
        // Execute reverse SQL
        switch (type) {
            case Type::Update:
                // UPDATE table SET column = oldValue WHERE pk = ...
                break;
            case Type::Insert:
                // DELETE FROM table WHERE pk = ...
                break;
            case Type::Delete:
                // INSERT INTO table VALUES (oldValue)
                break;
        }
    }

    void redo() {
        // Re-execute original SQL
        switch (type) {
            case Type::Update:
                // UPDATE table SET column = newValue WHERE pk = ...
                break;
            case Type::Insert:
                // INSERT INTO table VALUES (newValue)
                break;
            case Type::Delete:
                // DELETE FROM table WHERE pk = ...
                break;
        }
    }
};
```

## Model Integration
```cpp
// ui/DataGrid/QueryResultModel.cpp
bool QueryResultModel::setData(const QModelIndex& index, const QVariant& value, int role) {
    if (role != Qt::EditRole || !index.isValid()) {
        return false;
    }

    int row = index.row();
    int col = index.column();
    QString columnName = m_result.columns.at(col).name;

    // Store original value for undo
    if (!m_originalValues.contains({row, col})) {
        m_originalValues[{row, col}] = data(index, Qt::EditRole);
    }

    // Mark as edited
    m_editedRows.insert(row);

    // Notify DataChangeManager
    emit cellDataChanged(row, columnName, value);

    // Notify model of data change (triggers cell repaint)
    emit dataChanged(index, index, {Qt::DisplayRole, Qt::BackgroundRole});

    return true;
}

// Connect model to DataChangeManager
connect(m_model, &QueryResultModel::cellDataChanged,
        m_dataChangeManager, &DataChangeManager::updateCell);
```
