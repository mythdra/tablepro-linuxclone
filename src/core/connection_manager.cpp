#include "connection_manager.h"
#include <QUuid>
#include <QDateTime>
#include <QThread>
#include <QMutexLocker>
#include <QDebug>

namespace tablepro {

// Static members
ConnectionManager* ConnectionManager::m_instance = nullptr;
QMutex ConnectionManager::m_instanceMutex;

ConnectionManager* ConnectionManager::instance()
{
    QMutexLocker locker(&m_instanceMutex);
    if (!m_instance) {
        m_instance = new ConnectionManager();
    }
    return m_instance;
}

ConnectionManager::ConnectionManager(QObject *parent)
    : QObject(parent)
    , m_healthCheckTimer(new QTimer(this))
    , m_healthCheckInterval(30)  // Check every 30 seconds
{
    connect(m_healthCheckTimer, &QTimer::timeout, this, &ConnectionManager::checkConnectionHealth);
}

ConnectionManager::~ConnectionManager()
{
    closeAllConnections();
}

QString ConnectionManager::generateConnectionId()
{
    QUuid uuid = QUuid::createUuid();
    return QString("conn_%1").arg(uuid.toString());
}

QString ConnectionManager::createConnection(const ConnectionConfig& config)
{
    QMutexLocker locker(&m_connectionsMutex);

    QString connectionId = generateConnectionId();

    // Create appropriate driver based on database type
    QSharedPointer<DatabaseDriver> driver;

    // For now, we'll use a placeholder - this will be expanded when we implement actual drivers
    Q_UNUSED(config)

    // Add the connection to our registry
    ConnectionHandle handle(connectionId, driver, config);
    m_connections.insert(connectionId, handle);

    emit connectionOpened(connectionId);

    return connectionId;
}

bool ConnectionManager::testConnection(const QString& connectionId)
{
    QMutexLocker locker(&m_connectionsMutex);

    if (!m_connections.contains(connectionId)) {
        return false;
    }

    ConnectionHandle& handle = m_connections[connectionId];
    if (!handle.driver) {
        return false;
    }

    bool wasConnected = handle.driver->isConnected();
    if (!wasConnected) {
        bool connected = handle.driver->connect(handle.config);
        if (connected) {
            // If we successfully connected, disconnect if we weren't connected before
            if (!wasConnected) {
                handle.driver->disconnect();
            }
            return true;
        }
        return false;
    }

    // If already connected, execute a simple query to test
    auto result = handle.driver->executeQuery("SELECT 1");
    return result.success;
}

bool ConnectionManager::closeConnection(const QString& connectionId)
{
    QMutexLocker locker(&m_connectionsMutex);

    if (!m_connections.contains(connectionId)) {
        return false;
    }

    ConnectionHandle& handle = m_connections[connectionId];

    if (handle.driver && handle.driver->isConnected()) {
        handle.driver->disconnect();
    }

    m_connections.remove(connectionId);

    emit connectionClosed(connectionId);

    return true;
}

void ConnectionManager::closeAllConnections()
{
    QMutexLocker locker(&m_connectionsMutex);

    auto it = m_connections.begin();
    while (it != m_connections.end()) {
        if (it.value().driver && it.value().driver->isConnected()) {
            it.value().driver->disconnect();
        }
        emit connectionClosed(it.key());
        it++;
    }

    m_connections.clear();
}

QSharedPointer<DatabaseDriver> ConnectionManager::getConnection(const QString& connectionId)
{
    QMutexLocker locker(&m_connectionsMutex);

    if (m_connections.contains(connectionId)) {
        return m_connections[connectionId].driver;
    }

    return QSharedPointer<DatabaseDriver>();
}

QList<QString> ConnectionManager::getActiveConnections() const
{
    QMutexLocker locker(&m_connectionsMutex);
    return m_connections.keys();
}

bool ConnectionManager::isConnected(const QString& connectionId) const
{
    QMutexLocker locker(&m_connectionsMutex);

    if (!m_connections.contains(connectionId)) {
        return false;
    }

    const ConnectionHandle& handle = m_connections[connectionId];
    return handle.driver && handle.driver->isConnected();
}

int ConnectionManager::getConnectionCount() const
{
    QMutexLocker locker(&m_connectionsMutex);
    return m_connections.size();
}

ConnectionConfig ConnectionManager::getConnectionConfig(const QString& connectionId) const
{
    QMutexLocker locker(&m_connectionsMutex);

    if (m_connections.contains(connectionId)) {
        return m_connections[connectionId].config;
    }

    return ConnectionConfig();
}

void ConnectionManager::startHealthCheck()
{
    m_healthCheckTimer->start(m_healthCheckInterval * 1000);  // Convert to milliseconds
}

void ConnectionManager::stopHealthCheck()
{
    m_healthCheckTimer->stop();
}

void ConnectionManager::checkConnectionHealth()
{
    QMutexLocker locker(&m_connectionsMutex);

    auto it = m_connections.begin();
    while (it != m_connections.end()) {
        const ConnectionHandle& handle = it.value();
        if (handle.driver) {
            bool isHealthy = handle.driver->isConnected();
            emit healthCheckCompleted(handle.id, isHealthy);

            // Optionally handle unhealthy connections here
            if (!isHealthy) {
                // Could implement auto-reconnection logic here
            }
        }
        ++it;
    }
}

void ConnectionManager::setPoolSize(const QString& connectionId, int size)
{
    Q_UNUSED(connectionId)
    Q_UNUSED(size)
    // For future expansion - connection pooling functionality
}

int ConnectionManager::getPoolSize(const QString& connectionId) const
{
    Q_UNUSED(connectionId)
    // For future expansion - connection pooling functionality
    return 1; // Default to single connection
}

} // namespace tablepro