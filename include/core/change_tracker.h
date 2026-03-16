#pragma once

#include <QObject>
#include <QHash>
#include <QList>
#include <QVector>
#include <QVariant>
#include <QDateTime>
#include <QUuid>
#include <QMutex>
#include "query_result.h"

namespace tablepro {

enum class ChangeType {
    Insert,
    Update,
    Delete
};

struct ChangeRecord {
    QUuid id;
    ChangeType type;
    QString tableName;
    QString primaryKey;  // Field name of primary key
    QVariant primaryKeyValue;  // Value of primary key
    QHash<QString, QVariant> oldValues;  // Old values for update/delete
    QHash<QString, QVariant> newValues;  // New values for insert/update
    QDateTime timestamp;
    QString userId;
    bool isDirty = true;
    QString connectionString;  // For multi-connection tracking

    ChangeRecord() : id(QUuid::createUuid()), timestamp(QDateTime::currentDateTime()) {}
};

class ChangeTracker : public QObject
{
    Q_OBJECT

public:
    explicit ChangeTracker(QObject *parent = nullptr);
    ~ChangeTracker();

    // Track changes
    bool recordInsert(const QString& tableName,
                    const QHash<QString, QVariant>& values,
                    const QString& primaryKeyField = "id");

    bool recordUpdate(const QString& tableName,
                     const QString& primaryKeyValue,
                     const QHash<QString, QVariant>& oldValues,
                     const QHash<QString, QVariant>& newValues,
                     const QString& primaryKeyField = "id");

    bool recordDelete(const QString& tableName,
                     const QString& primaryKeyValue,
                     const QHash<QString, QVariant>& oldValues,
                     const QString& primaryKeyField = "id");

    // Change management
    QList<ChangeRecord> getAllChanges() const;
    QList<ChangeRecord> getChangesForTable(const QString& tableName) const;
    QList<ChangeRecord> getPendingChanges() const;
    QList<ChangeRecord> getChangesSince(const QDateTime& dateTime) const;

    // Persistence
    bool persistChanges();
    bool applyChangesToDatabase();
    bool discardChanges();
    bool markChangeAsApplied(const QUuid& changeId);
    bool markAllChangesAsApplied();

    // Undo/Redo
    bool canUndo() const;
    bool canRedo() const;
    bool undoLastChange();
    bool redoLastChange();

    // Dirty state management
    bool hasPendingChanges() const;
    int pendingChangeCount() const;
    void clearDirtyFlag();
    bool isDirty() const;

    // Utilities
    QString generateSqlForChange(const ChangeRecord& change) const;
    QList<QString> generateSqlForAllChanges() const;

    // Configuration
    void setConnectionString(const QString& connectionString);
    void setUserId(const QString& userId);
    void setAutoApply(bool autoApply);

signals:
    void changeRecorded(const ChangeRecord& change);
    void changesApplied(int count);
    void changesDiscarded(int count);
    void changeUndone(const ChangeRecord& change);
    void changeRedone(const ChangeRecord& change);
    void dirtyStateChanged(bool isDirty);

private:
    mutable QMutex m_mutex;
    QList<ChangeRecord> m_changes;
    QHash<QUuid, int> m_changeIndex;  // For fast lookup
    bool m_isDirty;
    QString m_connectionString;
    QString m_userId;
    bool m_autoApply;

    void setDirty(bool dirty);
    void addToIndex(const ChangeRecord& record);
    void removeFromIndex(const QUuid& id);
};

} // namespace tablepro