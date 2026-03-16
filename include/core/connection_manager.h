#pragma once

#include <QObject>
#include <QHash>
#include <QSharedPointer>
#include <QMutex>
#include <QTimer>
#include <QDateTime>
#include <memory>
#include "database_driver.h"
#include "connection_config.h"

namespace tablepro {

struct ConnectionHandle {
    QString id;
    QSharedPointer<DatabaseDriver> driver;
    ConnectionConfig config;
    qint64 timestamp;

    ConnectionHandle() : timestamp(0) {}
    ConnectionHandle(const QString& connId, QSharedPointer<DatabaseDriver> drv,
                     const ConnectionConfig& cfg)
        : id(connId), driver(drv), config(cfg), timestamp(QDateTime::currentMSecsSinceEpoch()) {}
};

class ConnectionManager : public QObject
{
    Q_OBJECT

public:
    static ConnectionManager* instance();

    // Connection lifecycle
    QString createConnection(const ConnectionConfig& config);
    bool testConnection(const QString& connectionId);
    bool closeConnection(const QString& connectionId);
    void closeAllConnections();

    // Connection access
    QSharedPointer<DatabaseDriver> getConnection(const QString& connectionId);
    QList<QString> getActiveConnections() const;
    bool isConnected(const QString& connectionId) const;

    // Connection management
    int getConnectionCount() const;
    ConnectionConfig getConnectionConfig(const QString& connectionId) const;

    // Health checks
    void startHealthCheck();
    void stopHealthCheck();
    void checkConnectionHealth();

    // Pooling interface (for future expansion)
    void setPoolSize(const QString& connectionId, int size);
    int getPoolSize(const QString& connectionId) const;

signals:
    void connectionOpened(const QString& connectionId);
    void connectionClosed(const QString& connectionId);
    void connectionError(const QString& connectionId, const QString& error);
    void healthCheckCompleted(const QString& connectionId, bool healthy);

private:
    explicit ConnectionManager(QObject *parent = nullptr);
    ~ConnectionManager();

    QString generateConnectionId();
    void cleanupDisconnectedConnections();

    static QMutex m_instanceMutex;
    static ConnectionManager* m_instance;

    QHash<QString, ConnectionHandle> m_connections;
    mutable QMutex m_connectionsMutex;

    QTimer* m_healthCheckTimer;
    int m_healthCheckInterval; // in seconds
};

} // namespace tablepro