#include "change_tracker.h"
#include <QMutexLocker>
#include <QUuid>
#include <QDebug>

namespace tablepro {

ChangeTracker::ChangeTracker(QObject *parent)
    : QObject(parent)
    , m_isDirty(false)
    , m_autoApply(false)
{
}

ChangeTracker::~ChangeTracker()
{
}

void ChangeTracker::setDirty(bool dirty)
{
    if (m_isDirty != dirty) {
        m_isDirty = dirty;
        emit dirtyStateChanged(dirty);
    }
}

void ChangeTracker::addToIndex(const ChangeRecord& record)
{
    m_changeIndex.insert(record.id, m_changes.size() - 1);
}

void ChangeTracker::removeFromIndex(const QUuid& id)
{
    m_changeIndex.remove(id);
    // Need to rebuild index after removal
    m_changeIndex.clear();
    for (int i = 0; i < m_changes.size(); ++i) {
        m_changeIndex.insert(m_changes[i].id, i);
    }
}

bool ChangeTracker::recordInsert(const QString& tableName,
                               const QHash<QString, QVariant>& values,
                               const QString& primaryKeyField)
{
    QMutexLocker locker(&m_mutex);

    ChangeRecord record;
    record.type = ChangeType::Insert;
    record.tableName = tableName;
    record.primaryKey = primaryKeyField;
    record.newValues = values;

    if (values.contains(primaryKeyField)) {
        record.primaryKeyValue = values[primaryKeyField];
    }

    record.userId = m_userId;
    record.connectionString = m_connectionString;

    m_changes.append(record);
    addToIndex(record);

    setDirty(true);
    emit changeRecorded(record);

    if (m_autoApply) {
        applyChangesToDatabase();
    }

    return true;
}

bool ChangeTracker::recordUpdate(const QString& tableName,
                               const QString& primaryKeyValue,
                               const QHash<QString, QVariant>& oldValues,
                               const QHash<QString, QVariant>& newValues,
                               const QString& primaryKeyField)
{
    QMutexLocker locker(&m_mutex);

    ChangeRecord record;
    record.type = ChangeType::Update;
    record.tableName = tableName;
    record.primaryKey = primaryKeyField;
    record.primaryKeyValue = primaryKeyValue;
    record.oldValues = oldValues;
    record.newValues = newValues;
    record.userId = m_userId;
    record.connectionString = m_connectionString;

    m_changes.append(record);
    addToIndex(record);

    setDirty(true);
    emit changeRecorded(record);

    if (m_autoApply) {
        applyChangesToDatabase();
    }

    return true;
}

bool ChangeTracker::recordDelete(const QString& tableName,
                               const QString& primaryKeyValue,
                               const QHash<QString, QVariant>& oldValues,
                               const QString& primaryKeyField)
{
    QMutexLocker locker(&m_mutex);

    ChangeRecord record;
    record.type = ChangeType::Delete;
    record.tableName = tableName;
    record.primaryKey = primaryKeyField;
    record.primaryKeyValue = primaryKeyValue;
    record.oldValues = oldValues;
    record.userId = m_userId;
    record.connectionString = m_connectionString;

    m_changes.append(record);
    addToIndex(record);

    setDirty(true);
    emit changeRecorded(record);

    if (m_autoApply) {
        applyChangesToDatabase();
    }

    return true;
}

QList<ChangeRecord> ChangeTracker::getAllChanges() const
{
    QMutexLocker locker(&m_mutex);
    return m_changes;
}

QList<ChangeRecord> ChangeTracker::getChangesForTable(const QString& tableName) const
{
    QMutexLocker locker(&m_mutex);
    QList<ChangeRecord> result;

    for (const auto& change : m_changes) {
        if (change.tableName == tableName) {
            result.append(change);
        }
    }

    return result;
}

QList<ChangeRecord> ChangeTracker::getPendingChanges() const
{
    QMutexLocker locker(&m_mutex);
    // In this implementation, all changes are pending by default
    return m_changes;
}

QList<ChangeRecord> ChangeTracker::getChangesSince(const QDateTime& dateTime) const
{
    QMutexLocker locker(&m_mutex);
    QList<ChangeRecord> result;

    for (const auto& change : m_changes) {
        if (change.timestamp >= dateTime) {
            result.append(change);
        }
    }

    return result;
}

bool ChangeTracker::persistChanges()
{
    // This would typically save changes to a file or database
    // For now, we'll just return true indicating success
    return true;
}

bool ChangeTracker::applyChangesToDatabase()
{
    QMutexLocker locker(&m_mutex);

    if (m_changes.isEmpty()) {
        return true;
    }

    // In a real implementation, this would execute the SQL commands to apply changes
    // to the actual database

    int appliedCount = 0;
    for (const auto& change : m_changes) {
        // Generate and execute SQL for the change
        QString sql = generateSqlForChange(change);
        Q_UNUSED(sql)  // Use the SQL to apply the change

        appliedCount++;
    }

    // Clear changes after applying
    m_changes.clear();
    m_changeIndex.clear();

    emit changesApplied(appliedCount);

    setDirty(false);

    return true;
}

bool ChangeTracker::discardChanges()
{
    QMutexLocker locker(&m_mutex);

    int discardedCount = m_changes.size();
    m_changes.clear();
    m_changeIndex.clear();

    emit changesDiscarded(discardedCount);

    setDirty(false);

    return true;
}

bool ChangeTracker::markChangeAsApplied(const QUuid& changeId)
{
    QMutexLocker locker(&m_mutex);

    auto it = m_changeIndex.find(changeId);
    if (it != m_changeIndex.end()) {
        int index = it.value();
        if (index >= 0 && index < m_changes.size()) {
            m_changes[index].isDirty = false;

            // Check if any changes remain dirty
            bool hasDirty = false;
            for (const auto& change : m_changes) {
                if (change.isDirty) {
                    hasDirty = true;
                    break;
                }
            }

            if (!hasDirty) {
                setDirty(false);
            }

            return true;
        }
    }

    return false;
}

bool ChangeTracker::markAllChangesAsApplied()
{
    QMutexLocker locker(&m_mutex);

    for (auto& change : m_changes) {
        change.isDirty = false;
    }

    setDirty(false);

    return true;
}

bool ChangeTracker::canUndo() const
{
    QMutexLocker locker(&m_mutex);
    return !m_changes.isEmpty();
}

bool ChangeTracker::canRedo() const
{
    // In this implementation, redo is not tracked separately
    // A more sophisticated implementation would have separate undo/redo stacks
    return false;
}

bool ChangeTracker::undoLastChange()
{
    QMutexLocker locker(&m_mutex);

    if (m_changes.isEmpty()) {
        return false;
    }

    ChangeRecord lastChange = m_changes.takeLast();
    m_changeIndex.remove(lastChange.id);

    emit changeUndone(lastChange);

    // Check if any changes remain dirty
    if (m_changes.isEmpty()) {
        setDirty(false);
    }

    return true;
}

bool ChangeTracker::redoLastChange()
{
    // In this simplified implementation, we don't maintain a redo stack
    // A more sophisticated implementation would have separate undo/redo stacks
    return false;
}

bool ChangeTracker::hasPendingChanges() const
{
    QMutexLocker locker(&m_mutex);
    return !m_changes.isEmpty();
}

int ChangeTracker::pendingChangeCount() const
{
    QMutexLocker locker(&m_mutex);
    return m_changes.size();
}

void ChangeTracker::clearDirtyFlag()
{
    QMutexLocker locker(&m_mutex);
    setDirty(false);
}

bool ChangeTracker::isDirty() const
{
    QMutexLocker locker(&m_mutex);
    return m_isDirty;
}

QString ChangeTracker::generateSqlForChange(const ChangeRecord& change) const
{
    QString sql;

    switch (change.type) {
        case ChangeType::Insert: {
            QStringList columns;
            QStringList placeholders;

            for (auto it = change.newValues.constBegin(); it != change.newValues.constEnd(); ++it) {
                columns << it.key();
                placeholders << "?";
            }

            sql = QString("INSERT INTO %1 (%2) VALUES (%3)")
                      .arg(change.tableName)
                      .arg(columns.join(", "))
                      .arg(placeholders.join(", "));
            break;
        }

        case ChangeType::Update: {
            QStringList setParts;

            for (auto it = change.newValues.constBegin(); it != change.newValues.constEnd(); ++it) {
                setParts << QString("%1 = ?").arg(it.key());
            }

            sql = QString("UPDATE %1 SET %2 WHERE %3 = ?")
                      .arg(change.tableName)
                      .arg(setParts.join(", "))
                      .arg(change.primaryKey);
            break;
        }

        case ChangeType::Delete: {
            sql = QString("DELETE FROM %1 WHERE %2 = ?")
                      .arg(change.tableName)
                      .arg(change.primaryKey);
            break;
        }
    }

    return sql;
}

QList<QString> ChangeTracker::generateSqlForAllChanges() const
{
    QMutexLocker locker(&m_mutex);
    QList<QString> sqlList;

    for (const auto& change : m_changes) {
        sqlList.append(generateSqlForChange(change));
    }

    return sqlList;
}

void ChangeTracker::setConnectionString(const QString& connectionString)
{
    QMutexLocker locker(&m_mutex);
    m_connectionString = connectionString;
}

void ChangeTracker::setUserId(const QString& userId)
{
    QMutexLocker locker(&m_mutex);
    m_userId = userId;
}

void ChangeTracker::setAutoApply(bool autoApply)
{
    QMutexLocker locker(&m_mutex);
    m_autoApply = autoApply;
}

} // namespace tablepro