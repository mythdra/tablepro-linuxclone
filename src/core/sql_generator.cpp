#include "sql_generator.h"
#include <QStringList>
#include <QRegularExpression>
#include <QDateTime>
#include <QDebug>

namespace tablepro {

SqlGenerator::SqlGenerator(DatabaseType dbType)
    : m_dbType(dbType)
{
}

QString SqlGenerator::generateSelect(const QString& table,
                                  const QStringList& columns,
                                  const QString& whereClause,
                                  const QString& orderBy,
                                  int limit,
                                  int offset) const
{
    QString selectColumns = "*";
    if (!columns.isEmpty()) {
        selectColumns = "\"" + columns.join("\", \"") + "\"";
    }

    QString sql = QString("SELECT %1 FROM \"%2\"").arg(selectColumns).arg(table);

    if (!whereClause.isEmpty()) {
        sql += " WHERE " + whereClause;
    }

    if (!orderBy.isEmpty()) {
        sql += " ORDER BY " + orderBy;
    }

    if (limit > 0) {
        sql += QString(" LIMIT %1").arg(limit);
    }

    if (offset > 0) {
        sql += QString(" OFFSET %1").arg(offset);
    }

    return sql;
}

QString SqlGenerator::generateInsert(const QString& table,
                                  const QHash<QString, QVariant>& values) const
{
    if (values.isEmpty()) {
        return QString();
    }

    QStringList columns;
    QStringList placeholders;

    for (auto it = values.constBegin(); it != values.constEnd(); ++it) {
        columns << quoteIdentifier(it.key());
        placeholders << "?";
    }

    QString sql = QString("INSERT INTO %1 (%2) VALUES (%3)")
                      .arg(quoteIdentifier(table))
                      .arg(columns.join(", "))
                      .arg(placeholders.join(", "));

    return sql;
}

QString SqlGenerator::generateUpdate(const QString& table,
                                  const QHash<QString, QVariant>& values,
                                  const QString& whereClause) const
{
    if (values.isEmpty()) {
        return QString();
    }

    QStringList setParts;

    for (auto it = values.constBegin(); it != values.constEnd(); ++it) {
        setParts << QString("%1 = ?").arg(quoteIdentifier(it.key()));
    }

    QString sql = QString("UPDATE %1 SET %2")
                      .arg(quoteIdentifier(table))
                      .arg(setParts.join(", "));

    if (!whereClause.isEmpty()) {
        sql += " WHERE " + whereClause;
    }

    return sql;
}

QString SqlGenerator::generateDelete(const QString& table,
                                  const QString& whereClause) const
{
    QString sql = QString("DELETE FROM %1").arg(quoteIdentifier(table));

    if (!whereClause.isEmpty()) {
        sql += " WHERE " + whereClause;
    } else {
        // Force user to specify WHERE clause to prevent accidental deletion
        qWarning() << "Warning: DELETE without WHERE clause will delete all records!";
    }

    return sql;
}

QString SqlGenerator::generateCreateTable(const QString& table,
                                       const QHash<QString, QString>& columns) const
{
    if (columns.isEmpty()) {
        return QString();
    }

    QStringList columnDefs;

    for (auto it = columns.constBegin(); it != columns.constEnd(); ++it) {
        columnDefs << QString("%1 %2").arg(quoteIdentifier(it.key())).arg(it.value());
    }

    QString sql = QString("CREATE TABLE %1 (%2)")
                      .arg(quoteIdentifier(table))
                      .arg(columnDefs.join(", "));

    return sql;
}

QString SqlGenerator::generateDropTable(const QString& table) const
{
    return QString("DROP TABLE %1").arg(quoteIdentifier(table));
}

QString SqlGenerator::generateAddColumn(const QString& table,
                                     const QString& columnName,
                                     const QString& columnType) const
{
    return QString("ALTER TABLE %1 ADD COLUMN %2 %3")
               .arg(quoteIdentifier(table))
               .arg(quoteIdentifier(columnName))
               .arg(columnType);
}

QString SqlGenerator::generateDropColumn(const QString& table,
                                      const QString& columnName) const
{
    return QString("ALTER TABLE %1 DROP COLUMN %2")
               .arg(quoteIdentifier(table))
               .arg(quoteIdentifier(columnName));
}

QStringList SqlGenerator::generateParameters(const QHash<QString, QVariant>& values) const
{
    QStringList params;
    for (auto it = values.constBegin(); it != values.constEnd(); ++it) {
        params << escapeValue(it.value());
    }
    return params;
}

QString SqlGenerator::quoteIdentifier(const QString& identifier) const
{
    switch (m_dbType) {
        case DatabaseType::PostgreSQL:
            return quoteIdentifierPostgreSQL(identifier);
        case DatabaseType::MySQL:
            return quoteIdentifierMySQL(identifier);
        case DatabaseType::SQLite:
            return quoteIdentifierSQLite(identifier);
        case DatabaseType::SQLServer:
            return quoteIdentifierSQLServer(identifier);
        default:
            return quoteIdentifierPostgreSQL(identifier);  // Default to PostgreSQL
    }
}

QString SqlGenerator::escapeValue(const QVariant& value) const
{
    switch (m_dbType) {
        case DatabaseType::PostgreSQL:
            return escapeValuePostgreSQL(value);
        case DatabaseType::MySQL:
            return escapeValueMySQL(value);
        case DatabaseType::SQLite:
            return escapeValueSQLite(value);
        case DatabaseType::SQLServer:
            return escapeValueSQLServer(value);
        default:
            return escapeValuePostgreSQL(value);  // Default to PostgreSQL
    }
}

void SqlGenerator::setDatabaseType(DatabaseType type)
{
    m_dbType = type;
}

bool SqlGenerator::validateQuery(const QString& query) const
{
    // Basic SQL injection prevention
    QString upperQuery = query.toUpper();

    // Check for dangerous keywords in inappropriate contexts
    QRegularExpression dangerousPattern(R"(DROP\s+TABLE|DROP\s+DATABASE|SHUTDOWN|EXEC\s*\(|EXECUTE\s*\(|SP_|XP_)",
                                        QRegularExpression::CaseInsensitiveOption);

    if (dangerousPattern.match(query).hasMatch()) {
        return false;
    }

    // Check for comment sequences
    if (query.contains("--") || query.contains("/*")) {
        return false;
    }

    // Basic structural check - query should have content
    return !query.trimmed().isEmpty();
}

QString SqlGenerator::sanitizeQuery(const QString& query) const
{
    QString sanitized = query;

    // Remove potential comment sequences
    sanitized.replace("--", "");
    sanitized.replace("/*", "");
    sanitized.replace("*/", "");

    // Remove potential script tags (though not SQL, defense in depth)
    sanitized.replace("<script", "");
    sanitized.replace("</script>", "");

    // Remove multiple whitespace sequences
    QRegularExpression wsRegex("\\s+");
    sanitized = sanitized.trimmed();
    sanitized.replace(wsRegex, " ");

    return sanitized.trimmed();
}

QString SqlGenerator::buildWhereClause(const QHash<QString, QVariant>& conditions) const
{
    if (conditions.isEmpty()) {
        return QString();
    }

    QStringList clauses;
    for (auto it = conditions.constBegin(); it != conditions.constEnd(); ++it) {
        clauses << QString("%1 = ?").arg(quoteIdentifier(it.key()));
    }

    return clauses.join(" AND ");
}

QString SqlGenerator::buildOrderByClause(const QStringList& columns) const
{
    if (columns.isEmpty()) {
        return QString();
    }

    QStringList quotedColumns;
    for (const QString& col : columns) {
        // Handle ASC/DESC suffixes
        if (col.toUpper().endsWith(" ASC") || col.toUpper().endsWith(" DESC")) {
            QStringList parts = col.split(' ');
            if (parts.size() >= 2) {
                QString column = parts[0];
                QString direction = parts[1];
                quotedColumns << QString("%1 %2").arg(quoteIdentifier(column)).arg(direction);
            }
        } else {
            quotedColumns << quoteIdentifier(col);
        }
    }

    return quotedColumns.join(", ");
}

// PostgreSQL-specific implementations
QString SqlGenerator::quoteIdentifierPostgreSQL(const QString& identifier) const
{
    return "\"" + identifier + "\"";
}

QString SqlGenerator::escapeValuePostgreSQL(const QVariant& value) const
{
    if (!value.isValid()) {
        return "NULL";
    }

    switch (value.typeId()) {
        case QMetaType::Int:
        case QMetaType::UInt:
        case QMetaType::LongLong:
        case QMetaType::ULongLong:
        case QMetaType::Double:
        case QMetaType::Float:
            return value.toString();
        case QMetaType::Bool:
            return value.toBool() ? "TRUE" : "FALSE";
        case QMetaType::QDateTime:
            return QString("'%1'").arg(value.toDateTime().toString(Qt::ISODate));
        case QMetaType::QDate:
            return QString("'%1'").arg(value.toDate().toString(Qt::ISODate));
        case QMetaType::QTime:
            return QString("'%1'").arg(value.toTime().toString(Qt::ISODate));
        default:
            // Escape single quotes and wrap in quotes
            return "'" + value.toString().replace("'", "''") + "'";
    }
}

// MySQL-specific implementations
QString SqlGenerator::quoteIdentifierMySQL(const QString& identifier) const
{
    return "`" + identifier + "`";
}

QString SqlGenerator::escapeValueMySQL(const QVariant& value) const
{
    if (!value.isValid()) {
        return "NULL";
    }

    switch (value.typeId()) {
        case QMetaType::Int:
        case QMetaType::UInt:
        case QMetaType::LongLong:
        case QMetaType::ULongLong:
        case QMetaType::Double:
        case QMetaType::Float:
            return value.toString();
        case QMetaType::Bool:
            return value.toBool() ? "1" : "0";
        case QMetaType::QDateTime:
            return QString("'%1'").arg(value.toDateTime().toString("yyyy-MM-dd hh:mm:ss"));
        case QMetaType::QDate:
            return QString("'%1'").arg(value.toDate().toString("yyyy-MM-dd"));
        case QMetaType::QTime:
            return QString("'%1'").arg(value.toTime().toString("hh:mm:ss"));
        default:
            // MySQL-style escaping
            return "'" + value.toString().replace("'", "\\'") + "'";
    }
}

// SQLite-specific implementations
QString SqlGenerator::quoteIdentifierSQLite(const QString& identifier) const
{
    return "\"" + identifier + "\"";
}

QString SqlGenerator::escapeValueSQLite(const QVariant& value) const
{
    if (!value.isValid()) {
        return "NULL";
    }

    switch (value.typeId()) {
        case QMetaType::Int:
        case QMetaType::UInt:
        case QMetaType::LongLong:
        case QMetaType::ULongLong:
        case QMetaType::Double:
        case QMetaType::Float:
            return value.toString();
        case QMetaType::Bool:
            return value.toBool() ? "1" : "0";
        case QMetaType::QDateTime:
            return QString("'%1'").arg(value.toDateTime().toString(Qt::ISODate));
        case QMetaType::QDate:
            return QString("'%1'").arg(value.toDate().toString(Qt::ISODate));
        case QMetaType::QTime:
            return QString("'%1'").arg(value.toTime().toString(Qt::ISODate));
        default:
            // SQLite-style escaping
            return "'" + value.toString().replace("'", "''") + "'";
    }
}

// SQL Server-specific implementations
QString SqlGenerator::quoteIdentifierSQLServer(const QString& identifier) const
{
    return "[" + identifier + "]";
}

QString SqlGenerator::escapeValueSQLServer(const QVariant& value) const
{
    if (!value.isValid()) {
        return "NULL";
    }

    switch (value.typeId()) {
        case QMetaType::Int:
        case QMetaType::UInt:
        case QMetaType::LongLong:
        case QMetaType::ULongLong:
        case QMetaType::Double:
        case QMetaType::Float:
            return value.toString();
        case QMetaType::Bool:
            return value.toBool() ? "1" : "0";
        case QMetaType::QDateTime:
            return QString("'%1'").arg(value.toDateTime().toString("yyyy-MM-dd hh:mm:ss.zzz"));
        case QMetaType::QDate:
            return QString("'%1'").arg(value.toDate().toString("yyyy-MM-dd"));
        case QMetaType::QTime:
            return QString("'%1'").arg(value.toTime().toString("hh:mm:ss.zzz"));
        default:
            // SQL Server-style escaping
            return "'" + value.toString().replace("'", "''") + "'";
    }
}

} // namespace tablepro