#include "query_executor.h"
#include <QThread>
#include <QThreadPool>
#include <QRunnable>
#include <QTimer>
#include <QMetaObject>
#include <QDebug>

namespace tablepro {

QueryExecutor::QueryExecutor(QSharedPointer<DatabaseDriver> driver, QObject *parent)
    : QObject(parent)
    , m_driver(driver)
    , m_cacheResults(false)
    , m_transactionActive(false)
{
}

QueryExecutor::~QueryExecutor()
{
}

QueryResult QueryExecutor::executeQuery(const QString& query, const QVector<QVariant>& params)
{
    if (!m_driver) {
        QueryResult result;
        result.success = false;
        result.errorMessage = "No database driver available";
        return result;
    }

    emit queryStarted(query);

    QueryResult result = m_driver->executeQuery(query, params);

    if (result.success) {
        emit queryFinished(query, result);
    } else {
        emit queryError(query, result.errorMessage);
    }

    return result;
}

QueryResult QueryExecutor::executeNonQuery(const QString& query, const QVector<QVariant>& params)
{
    if (!m_driver) {
        QueryResult result;
        result.success = false;
        result.errorMessage = "No database driver available";
        return result;
    }

    emit queryStarted(query);

    QueryResult result = m_driver->executeNonQuery(query, params);

    if (result.success) {
        emit queryFinished(query, result);
    } else {
        emit queryError(query, result.errorMessage);
    }

    return result;
}

QueryResult QueryExecutor::prepareAndExecute(const QString& query, const QVector<QVariant>& params)
{
    if (!m_driver) {
        QueryResult result;
        result.success = false;
        result.errorMessage = "No database driver available";
        return result;
    }

    emit queryStarted(query);

    QueryResult result = m_driver->prepareAndExecute(query, params);

    if (result.success) {
        emit queryFinished(query, result);
    } else {
        emit queryError(query, result.errorMessage);
    }

    return result;
}

void QueryExecutor::executeQueryAsync(const QString& query,
                                    const QVector<QVariant>& params,
                                    std::function<void(const QueryResult&)> callback)
{
    if (!m_driver) {
        QueryResult result;
        result.success = false;
        result.errorMessage = "No database driver available";

        if (callback) {
            callback(result);
        }
        return;
    }

    // Emit the signal before the query starts
    emit queryStarted(query);

    // Use Qt's meta-object system to call the async operation in the correct thread
    auto *timer = new QTimer(this);
    timer->setSingleShot(true);

    connect(timer, &QTimer::timeout, this, [=]() {
        QueryResult result = m_driver->executeQuery(query, params);

        if (result.success) {
            emit queryFinished(query, result);
        } else {
            emit queryError(query, result.errorMessage);
        }

        if (callback) {
            callback(result);
        }

        timer->deleteLater();
    });

    timer->start(0); // Execute as soon as possible in the event loop
}

bool QueryExecutor::beginTransaction()
{
    if (!m_driver) {
        return false;
    }

    bool result = m_driver->beginTransaction();
    if (result) {
        m_transactionActive = true;
        emit transactionStarted();
    }
    return result;
}

bool QueryExecutor::commit()
{
    if (!m_driver) {
        return false;
    }

    bool result = m_driver->commit();
    if (result) {
        m_transactionActive = false;
        emit transactionCommitted();
    }
    return result;
}

bool QueryExecutor::rollback()
{
    if (!m_driver) {
        return false;
    }

    bool result = m_driver->rollback();
    if (result) {
        m_transactionActive = false;
        emit transactionRolledBack();
    }
    return result;
}

bool QueryExecutor::isInTransaction() const
{
    return m_transactionActive && m_driver && m_driver->isInTransaction();
}

void QueryExecutor::setDriver(QSharedPointer<DatabaseDriver> driver)
{
    m_driver = driver;
}

QSharedPointer<DatabaseDriver> QueryExecutor::getDriver() const
{
    return m_driver;
}

void QueryExecutor::setCacheResults(bool cache)
{
    m_cacheResults = cache;
}

bool QueryExecutor::isCachingResults() const
{
    return m_cacheResults;
}

} // namespace tablepro