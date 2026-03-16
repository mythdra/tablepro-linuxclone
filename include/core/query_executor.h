#pragma once

#include <QObject>
#include <QVector>
#include <QVariant>
#include <QSharedPointer>
#include <functional>
#include <memory>
#include "database_driver.h"

namespace tablepro {

class QueryExecutor : public QObject
{
    Q_OBJECT

public:
    explicit QueryExecutor(QSharedPointer<DatabaseDriver> driver, QObject *parent = nullptr);
    ~QueryExecutor();

    // Synchronous query execution
    QueryResult executeQuery(const QString& query, const QVector<QVariant>& params = {});
    QueryResult executeNonQuery(const QString& query, const QVector<QVariant>& params = {});

    // Prepared statements
    QueryResult prepareAndExecute(const QString& query, const QVector<QVariant>& params = {});

    // Asynchronous query execution
    void executeQueryAsync(const QString& query,
                          const QVector<QVariant>& params,
                          std::function<void(const QueryResult&)> callback);

    // Transaction management
    bool beginTransaction();
    bool commit();
    bool rollback();
    bool isInTransaction() const;

    // Utility methods
    void setDriver(QSharedPointer<DatabaseDriver> driver);
    QSharedPointer<DatabaseDriver> getDriver() const;

    // Query optimization hints
    void setCacheResults(bool cache);
    bool isCachingResults() const;

signals:
    void queryStarted(const QString& query);
    void queryFinished(const QString& query, const QueryResult& result);
    void queryError(const QString& query, const QString& error);
    void transactionStarted();
    void transactionCommitted();
    void transactionRolledBack();

private:
    QSharedPointer<DatabaseDriver> m_driver;
    bool m_cacheResults;
    bool m_transactionActive;
};

} // namespace tablepro