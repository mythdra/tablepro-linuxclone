#pragma once

#include <QString>
#include <QHash>
#include <QVariant>
#include <QVector>
#include "database_types.h"

namespace tablepro {

class SqlGenerator
{
public:
    explicit SqlGenerator(DatabaseType dbType = DatabaseType::PostgreSQL);

    // CRUD Operations
    QString generateSelect(const QString& table,
                         const QStringList& columns = QStringList(),
                         const QString& whereClause = QString(),
                         const QString& orderBy = QString(),
                         int limit = -1,
                         int offset = 0) const;

    QString generateInsert(const QString& table,
                          const QHash<QString, QVariant>& values) const;

    QString generateUpdate(const QString& table,
                          const QHash<QString, QVariant>& values,
                          const QString& whereClause) const;

    QString generateDelete(const QString& table,
                          const QString& whereClause) const;

    // Schema operations
    QString generateCreateTable(const QString& table,
                              const QHash<QString, QString>& columns) const;

    QString generateDropTable(const QString& table) const;

    QString generateAddColumn(const QString& table,
                             const QString& columnName,
                             const QString& columnType) const;

    QString generateDropColumn(const QString& table,
                             const QString& columnName) const;

    // Parameter handling
    QStringList generateParameters(const QHash<QString, QVariant>& values) const;
    QString quoteIdentifier(const QString& identifier) const;
    QString escapeValue(const QVariant& value) const;

    // Database-specific adjustments
    void setDatabaseType(DatabaseType type);
    DatabaseType getDatabaseType() const { return m_dbType; }

    // SQL validation and sanitization
    bool validateQuery(const QString& query) const;
    QString sanitizeQuery(const QString& query) const;

    // Helpers
    QString buildWhereClause(const QHash<QString, QVariant>& conditions) const;
    QString buildOrderByClause(const QStringList& columns) const;

private:
    DatabaseType m_dbType;

    QString quoteIdentifierPostgreSQL(const QString& identifier) const;
    QString quoteIdentifierMySQL(const QString& identifier) const;
    QString quoteIdentifierSQLite(const QString& identifier) const;
    QString quoteIdentifierSQLServer(const QString& identifier) const;

    QString escapeValuePostgreSQL(const QVariant& value) const;
    QString escapeValueMySQL(const QVariant& value) const;
    QString escapeValueSQLite(const QVariant& value) const;
    QString escapeValueSQLServer(const QVariant& value) const;
};

} // namespace tablepro