# SSH & SSL Internals (C++20 + Qt)

## SSH Tunneling

Qt provides SSH tunneling via **libssh2** integration with `QTcpSocket`. No external processes needed.

### Tunnel Setup Flow

```cpp
// src/core/SSHTunnel.hpp
#pragma once

#include <QObject>
#include <QTcpSocket>
#include <libssh2.h>
#include <libssh2_publickey.h>
#include <memory>

namespace tablepro {

struct SSHTunnelConfig {
    QString connectionId;
    QString sshHost;
    int sshPort{22};
    QString sshUser;
    QString sshPassword;  // Stored in QKeychain
    QString sshKeyPath;
    QString sshKeyPassphrase;  // Stored in QKeychain
    QString dbHost;  // Target database host (from tunnel perspective)
    int dbPort{5432};

    enum class AuthMethod {
        Password,
        KeyFile,
        Agent
    };
    AuthMethod authMethod{AuthMethod::Password};
};

class SSHTunnel : public QObject {
    Q_OBJECT

public:
    explicit SSHTunnel(QObject* parent = nullptr);
    ~SSHTunnel();

    // Start tunnel and return local port
    QFuture<int> start(const SSHTunnelConfig& config);

    // Close tunnel
    void close();

    // Health check
    bool isConnected() const;

    QString lastError() const { return m_error; }

signals:
    void tunnelStarted(int localPort);
    void tunnelError(const QString& error);
    void reconnecting();
    void reconnected();

private:
    struct Private;  // Pimpl pattern for libssh2 types
    std::unique_ptr<Private> d;

    QTcpSocket* m_socket{nullptr};
    LIBSSH2_SESSION* m_session{nullptr};
    LIBSSH2_CHANNEL* m_channel{nullptr};
    QTcpServer* m_localServer{nullptr};

    QString m_error;
    bool m_connected{false};
    QElapsedTimer m_keepaliveTimer;

    // Auth setup
    bool setupPasswordAuth(const SSHTunnelConfig& config);
    bool setupKeyAuth(const SSHTunnelConfig& config);
    bool setupAgentAuth(const SSHTunnelConfig& config);

    // Keepalive
    void startKeepalive();
    void stopKeepalive();

    // Forwarding
    void handleForwardConnection();
};

} // namespace tablepro
```

```cpp
// src/core/SSHTunnel.cpp
#include "SSHTunnel.hpp"
#include <QTcpServer>
#include <QKeychain>
#include <QtConcurrent>

QFuture<int> SSHTunnel::start(const SSHTunnelConfig& config) {
    return QtConcurrent::run([=]() {
        // Initialize libssh2
        libssh2_init(0);

        m_socket = new QTcpSocket(this);
        m_socket->connectToHost(config.sshHost, config.sshPort);

        if (!m_socket->waitForConnected(10000)) {
            m_error = tr("Failed to connect to SSH server: %1").arg(
                m_socket->errorString());
            emit tunnelError(m_error);
            return 0;
        }

        // Create SSH session
        m_session = libssh2_session_init();
        libssh2_session_set_blocking(m_session, 1);

        // Set socket for libssh2
        libssh2_session_set_socket(m_session, m_socket->socketDescriptor());

        // Perform SSH handshake
        if (libssh2_session_handshake(m_session, m_socket->socketDescriptor())) {
            m_error = tr("SSH handshake failed");
            return 0;
        }

        // Setup auth methods
        int rc = 0;
        switch (config.authMethod) {
            case SSHTunnelConfig::AuthMethod::Password:
                rc = libssh2_userauth_password(m_session,
                    config.sshUser.toUtf8().constData(),
                    config.sshPassword.toUtf8().constData());
                break;

            case SSHTunnelConfig::AuthMethod::KeyFile:
                rc = libssh2_userauth_publickey_fromfile_ex(
                    m_session,
                    config.sshUser.toUtf8().constData(),
                    config.sshKeyPath.toUtf8().constData(),
                    nullptr,  // Public key (can be nullptr)
                    config.sshKeyPassphrase.toUtf8().constData());
                break;

            case SSHTunnelConfig::AuthMethod::Agent:
                // SSH agent authentication
                rc = libssh2_userauth_agent(m_session,
                    config.sshUser.toUtf8().constData());
                break;
        }

        if (rc != 0) {
            m_error = tr("SSH authentication failed: %1").arg(rc);
            return 0;
        }

        // Start TCP forwarding
        // Listen on a random local port
        m_localServer = new QTcpServer(this);
        if (!m_localServer->listen(QHostAddress::LocalHost, 0)) {
            m_error = tr("Failed to start local listener");
            return 0;
        }

        int localPort = m_localServer->serverPort();

        // Accept connections and forward
        connect(m_localServer, &QTcpServer::newConnection,
                this, [this, config]() {
            handleForwardConnection();
        });

        m_connected = true;
        m_keepaliveTimer.start();
        startKeepalive();

        emit tunnelStarted(localPort);
        return localPort;
    });
}

void SSHTunnel::handleForwardConnection() {
    // Accept incoming local connection
    QTcpSocket* localClient = m_localServer->nextConnection();

    // Open direct TCP channel through SSH to database
    LIBSSH2_CHANNEL* channel = libssh2_channel_direct_tcpip_ex(
        m_session,
        config.dbHost.toUtf8().constData(),
        config.dbPort,
        "127.0.0.1",
        0);

    if (!channel) {
        localClient->close();
        localClient->deleteLater();
        return;
    }

    // Set up bidirectional forwarding between local client and SSH channel
    // This would use QSocketNotifier for async I/O
    // Implementation omitted for brevity
}

void SSHTunnel::startKeepalive() {
    // Send keepalive every 30 seconds
    auto* timer = new QTimer(this);
    connect(timer, &QTimer::timeout, this, [this]() {
        if (m_session) {
            int rc = libssh2_keepalive_send(m_session, nullptr);
            if (rc != 0) {
                // Tunnel may have died - attempt reconnection
                emit reconnecting();
            }
        }
    });
    timer->start(30000);
}

void SSHTunnel::close() {
    if (m_channel) {
        libssh2_channel_close(m_channel);
        m_channel = nullptr;
    }
    if (m_session) {
        libssh2_session_disconnect(m_session, "Client closing");
        libssh2_session_free(m_session);
        m_session = nullptr;
    }
    if (m_socket) {
        m_socket->close();
        m_socket->deleteLater();
    }
    if (m_localServer) {
        m_localServer->close();
        m_localServer->deleteLater();
    }
    m_connected = false;

    libssh2_exit();
}
```

### Tunnel Health & Recovery

```cpp
// src/core/SSHTunnelRecovery.hpp
#pragma once

#include <QObject>
#include <QTimer>
#include "SSHTunnel.hpp"

namespace tablepro {

class SSHTunnelRecovery : public QObject {
    Q_OBJECT

public:
    explicit SSHTunnelRecovery(SSHTunnel* tunnel, QObject* parent = nullptr);

    void startMonitoring(const SSHTunnelConfig& config);
    void stop();

signals:
    void tunnelRecovered();
    void tunnelPermanentlyFailed(const QString& error);

private slots:
    void checkHealth();
    void attemptReconnect();

private:
    SSHTunnel* m_tunnel;
    SSHTunnelConfig m_config;
    QTimer* m_healthCheckTimer;
    int m_reconnectAttempts{0};
    int m_backoffMs{1000};  // Start at 1s, double each time (max 30s)
    static constexpr int kMaxBackoffMs = 30000;
};

} // namespace tablepro
```

## SSL/TLS Configuration

Qt provides native TLS support via `QSslSocket` and `QSslConfiguration`.

```cpp
// src/core/SSLConfig.hpp
#pragma once

#include <QSslConfiguration>
#include <QSslCertificate>
#include <QSslKey>
#include <QString>

namespace tablepro {

struct SSLConfig {
    bool enabled{false};
    QString mode;  // "disable", "require", "verify-ca", "verify-full"
    QString caCertPath;
    QString clientCertPath;
    QString clientKeyPath;

    QSslConfiguration toSslConfiguration() const;
};

SSLConfig SSLConfig::toSslConfiguration() const {
    QSslConfiguration sslConfig;

    if (!caCertPath.isEmpty()) {
        QFile caFile(caCertPath);
        if (caFile.open(QIODevice::ReadOnly)) {
            QList<QSslCertificate> caCerts =
                QSslCertificate::fromData(caFile.readAll());
            sslConfig.setCaCertificates(caCerts);
        }
    }

    if (!clientCertPath.isEmpty() && !clientKeyPath.isEmpty()) {
        QFile certFile(clientCertPath);
        if (certFile.open(QIODevice::ReadOnly)) {
            QSslCertificate clientCert(certFile.readAll(),
                QSsl::Pem, QSsl::Opaque);
            sslConfig.setLocalCertificate(clientCert);
        }

        QFile keyFile(clientKeyPath);
        if (keyFile.open(QIODevice::ReadOnly)) {
            QSslKey clientKey(keyFile.readAll(),
                QSsl::Rsa, QSsl::Pem, QSsl::PrivateKey);
            sslConfig.setPrivateKey(clientKey);
        }
    }

    // Mode handling
    if (mode == "disable") {
        // No SSL
    } else if (mode == "require") {
        sslConfig.setPeerVerifyMode(QSslSocket::VerifyNone);
    } else if (mode == "verify-ca") {
        sslConfig.setPeerVerifyMode(QSslSocket::QueryPeer);
    } else if (mode == "verify-full") {
        sslConfig.setPeerVerifyMode(QSslSocket::VerifyPeer);
    }

    return sslConfig;
}

} // namespace tablepro
```

### PostgreSQL SSL (libpq)

```cpp
// In PostgresDriver::connect()
bool PostgresDriver::connect(const ConnectionConfig& config) {
    QString conninfo;

    // Build connection string
    conninfo += QString("host=%1 ").arg(config.host);
    conninfo += QString("port=%1 ").arg(config.port);
    conninfo += QString("dbname=%1 ").arg(config.database);
    conninfo += QString("user=%1 ").arg(config.username);

    // SSL parameters
    if (config.ssl.enabled) {
        conninfo += QString("sslmode=%1 ").arg(config.ssl.mode);

        if (!config.ssl.caCertPath.isEmpty()) {
            conninfo += QString("sslrootcert=%1 ").arg(config.ssl.caCertPath);
        }
        if (!config.ssl.clientCertPath.isEmpty()) {
            conninfo += QString("sslcert=%1 ").arg(config.ssl.clientCertPath);
        }
        if (!config.ssl.clientKeyPath.isEmpty()) {
            conninfo += QString("sslkey=%1 ").arg(config.ssl.clientKeyPath);
        }
    } else {
        conninfo += "sslmode=disable ";
    }

    m_connection.reset(PQconnectdb(conninfo.toUtf8().constData()));

    if (PQstatus(m_connection.get()) != CONNECTION_OK) {
        m_error = QString::fromUtf8(PQerrorMessage(m_connection.get()));
        return false;
    }

    return true;
}
```

### MySQL SSL

```cpp
// In MysqlDriver::connect()
bool MysqlDriver::connect(const ConnectionConfig& config) {
    m_connection.reset(mysql_init(nullptr));

    // SSL options
    if (config.ssl.enabled) {
        mysql_ssl_set(m_connection.get(),
            config.ssl.clientKeyPath.toUtf8().constData(),
            config.ssl.clientCertPath.toUtf8().constData(),
            config.ssl.caCertPath.toUtf8().constData(),
            nullptr,  // CA path
            nullptr   // Cipher
        );
    }

    MYSQL* conn = mysql_real_connect(m_connection.get(),
        config.host.toUtf8().constData(),
        config.username.toUtf8().constData(),
        config.password.toUtf8().constData(),
        config.database.toUtf8().constData(),
        config.port,
        nullptr,  // Unix socket
        CLIENT_SSL  // Client flags
    );

    if (!conn) {
        m_error = QString::fromUtf8(mysql_error(m_connection.get()));
        return false;
    }

    return true;
}
```

## QKeychain Integration

```cpp
// src/connection/KeychainStorage.hpp
#pragma once

#include <QObject>
#include <QKeychain::Job>
#include <QKeychain::ReadPasswordJob>
#include <QKeychain::WritePasswordJob>
#include <QKeychain::DeletePasswordJob>

namespace tablepro {

class KeychainStorage : public QObject {
    Q_OBJECT

public:
    explicit KeychainStorage(QObject* parent = nullptr);

    // Password operations
    QFuture<QString> readPassword(const QString& key);
    QFuture<bool> writePassword(const QString& key, const QString& password);
    QFuture<bool> deletePassword(const QString& key);

    // Connection-specific helpers
    QFuture<QString> readConnectionPassword(const QUuid& connectionId);
    QFuture<QString> readSshPassword(const QUuid& connectionId);
    QFuture<QString> readSshKeyPassphrase(const QUuid& connectionId);

signals:
    void passwordRead(const QString& key, const QString& password);
    void passwordWritten(const QString& key);
    void passwordDeleted(const QString& key);
    void error(QKeychain::Error error, const QString& errorString);

private:
    static constexpr const char* kServiceName = "tablepro";
};

} // namespace tablepro
```

## Advantages over Go Implementation

| Aspect | Go `golang.org/x/crypto/ssh` | C++ libssh2 + Qt |
|--------|------------------------------|------------------|
| Auth methods | Built-in | Via libssh2 API |
| SSH Agent | Direct Unix socket | Via libssh2 |
| Passphrase | Programmatic | Via QKeychain |
| Cross-platform | Pure Go | Requires libssh2 |
| Async I/O | Goroutines | Qt event loop |
| Binary size | Small | Larger (native lib) |
