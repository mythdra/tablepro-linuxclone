#pragma once

#include <QObject>
#include <QVector>
#include <QVariant>
#include <QJsonObject>
#include <QJsonArray>
#include <QSharedPointer>
#include <QScopedPointer>
#include <functional>
#include <memory>
#include "database_types.h"
#include "connection_config.h"
#include "query_result.h"

namespace tablepro {

// Forward declarations for error handling
enum class ErrorCode;
class DatabaseException;
class ConnectionException;
class QueryExecutionException;
class TransactionException;
class AuthenticationException;

class DatabaseDriver : public QObject
{
    Q_OBJECT

public:
    explicit DatabaseDriver(QObject *parent = nullptr);
    virtual ~DatabaseDriver();

    // Connection methods
    virtual bool connect(const ConnectionConfig& config) = 0;
    virtual void disconnect() = 0;
    virtual bool isConnected() const = 0;

    // Query execution
    virtual QueryResult executeQuery(const QString& query,
                                   const QVector<QVariant>& params = {}) = 0;
    virtual QueryResult executeNonQuery(const QString& query,
                                      const QVector<QVariant>& params = {}) = 0;

    // Prepared statements
    virtual QueryResult prepareAndExecute(const QString& query,
                                        const QVector<QVariant>& params = {}) = 0;

    // Transaction support
    virtual bool beginTransaction() = 0;
    virtual bool commit() = 0;
    virtual bool rollback() = 0;
    virtual bool isInTransaction() const = 0;

    // Schema introspection
    virtual QueryResult getDatabases() = 0;
    virtual QueryResult getTables(const QString& database = "") = 0;
    virtual QueryResult getColumns(const QString& table, const QString& schema = "") = 0;

    // Utility methods
    virtual QString getServerVersion() = 0;
    virtual DatabaseType getType() const = 0;

    // Error handling
    virtual QString lastError() const = 0;

    // Async execution (optional)
    virtual void executeQueryAsync(const QString& query,
                                 const QVector<QVariant>& params,
                                 std::function<void(const QueryResult&)> callback) = 0;

signals:
    void connected();
    void disconnected();
    void errorOccurred(const QString& error);
    void queryExecuted(const QString& query, const QueryResult& result);

protected:
    ConnectionConfig m_config;
    bool m_connected = false;
};

} // namespace tablepro