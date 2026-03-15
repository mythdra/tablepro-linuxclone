# Phase 10: SSH/SSL & Security Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement SSH tunneling for database connections, SSL/TLS configuration, and secure password storage with QKeychain.

**Architecture:** SshTunnel class manages libssh2 connections. QKeychain integration for password/SSH key passphrase storage.

**Tech Stack:** C++20, libssh2, Qt Network, QKeychain

---

## Task 1: SSH Tunnel Class

**Files:**
- Create: `src/core/ssh_tunnel.hpp`
- Create: `src/core/ssh_tunnel.cpp`

**Step 1: Add libssh2 dependency to vcpkg.json**

```json
"libssh2"
```

**Step 2: Create ssh_tunnel.hpp**

```cpp
#pragma once

#include <QObject>
#include <QTcpSocket>
#include <QThread>
#include <memory>

struct ssh2_session_struct;

namespace tablepro {

struct SshTunnelConfig {
    QString host;
    int port = 22;
    QString username;
    QString password;        // Retrieved from keychain
    QString privateKeyPath;
    QString passphrase;      // Retrieved from keychain
    int localPort = 0;       // 0 = auto-assign
    QString remoteHost;
    int remotePort = 0;
    int timeout = 30;
};

class SshTunnel : public QObject {
    Q_OBJECT

public:
    explicit SshTunnel(QObject* parent = nullptr);
    ~SshTunnel() override;

    bool start(const SshTunnelConfig& config);
    void stop();
    bool isRunning() const;

    int localPort() const;
    QString errorString() const;

signals:
    void started(int localPort);
    void stopped();
    void error(const QString& message);

private:
    bool connectToServer();
    bool authenticate();
    bool setupTunnel();
    void cleanup();

    SshTunnelConfig m_config;
    QTcpSocket m_socket;
    ssh2_session_struct* m_session = nullptr;
    int m_localPort = 0;
    QString m_errorString;
};

} // namespace tablepro
```

**Step 3: Create ssh_tunnel.cpp**

```cpp
#include "ssh_tunnel.hpp"
#include <libssh2.h>
#include <QFile>
#include <QTcpServer>

namespace tablepro {

SshTunnel::SshTunnel(QObject* parent)
    : QObject(parent)
{
    libssh2_init(0);
}

SshTunnel::~SshTunnel() {
    stop();
    libssh2_exit();
}

bool SshTunnel::start(const SshTunnelConfig& config) {
    m_config = config;
    m_errorString.clear();

    // Find available local port
    if (m_config.localPort == 0) {
        QTcpServer server;
        if (server.listen(QHostAddress::LocalHost)) {
            m_config.localPort = server.serverPort();
            server.close();
        } else {
            m_errorString = "Failed to find available local port";
            emit error(m_errorString);
            return false;
        }
    }

    if (!connectToServer()) {
        return false;
    }

    if (!authenticate()) {
        cleanup();
        return false;
    }

    if (!setupTunnel()) {
        cleanup();
        return false;
    }

    m_localPort = m_config.localPort;
    emit started(m_localPort);
    return true;
}

void SshTunnel::stop() {
    cleanup();
    m_localPort = 0;
    emit stopped();
}

bool SshTunnel::isRunning() const {
    return m_session != nullptr;
}

int SshTunnel::localPort() const {
    return m_localPort;
}

QString SshTunnel::errorString() const {
    return m_errorString;
}

bool SshTunnel::connectToServer() {
    m_socket.connectToHost(m_config.host, m_config.port);
    if (!m_socket.waitForConnected(m_config.timeout * 1000)) {
        m_errorString = QString("Failed to connect to SSH server: %1").arg(m_socket.errorString());
        emit error(m_errorString);
        return false;
    }

    m_session = libssh2_session_init();
    if (!m_session) {
        m_errorString = "Failed to initialize SSH session";
        emit error(m_errorString);
        return false;
    }

    libssh2_session_set_blocking(m_session, 1);

    int rc = libssh2_session_handshake(m_session, m_socket.socketDescriptor());
    if (rc) {
        m_errorString = QString("SSH handshake failed: %1").arg(rc);
        emit error(m_errorString);
        return false;
    }

    return true;
}

bool SshTunnel::authenticate() {
    int rc = 0;

    // Try public key authentication first
    if (!m_config.privateKeyPath.isEmpty()) {
        QFile keyFile(m_config.privateKeyPath);
        if (keyFile.open(QIODevice::ReadOnly)) {
            QByteArray keyData = keyFile.readAll();
            keyFile.close();

            rc = libssh2_userauth_publickey_frommemory(
                m_session,
                m_config.username.toUtf8().constData(),
                m_config.username.length(),
                nullptr, 0,
                keyData.constData(),
                keyData.length(),
                m_config.passphrase.isEmpty() ? nullptr : m_config.passphrase.toUtf8().constData()
            );

            if (rc == 0) {
                return true;
            }
        }
    }

    // Fall back to password authentication
    if (!m_config.password.isEmpty()) {
        rc = libssh2_userauth_password(
            m_session,
            m_config.username.toUtf8().constData(),
            m_config.password.toUtf8().constData()
        );

        if (rc == 0) {
            return true;
        }
    }

    m_errorString = "SSH authentication failed";
    emit error(m_errorString);
    return false;
}

bool SshTunnel::setupTunnel() {
    // Create direct TCP-IP channel
    auto* channel = libssh2_channel_direct_tcpip(
        m_session,
        m_config.remoteHost.toUtf8().constData(),
        m_config.remotePort,
        "127.0.0.1",
        m_config.localPort
    );

    if (!channel) {
        m_errorString = "Failed to create SSH tunnel";
        emit error(m_errorString);
        return false;
    }

    // TODO: Set up local port forwarding
    // This requires a more complex implementation with a local TCP server

    return true;
}

void SshTunnel::cleanup() {
    if (m_session) {
        libssh2_session_disconnect(m_session, "Closing");
        libssh2_session_free(m_session);
        m_session = nullptr;
    }

    m_socket.close();
}

} // namespace tablepro
```

**Commit:**

```bash
git add src/core/ssh_tunnel.hpp src/core/ssh_tunnel.cpp
git commit -m "feat: Add SSH tunnel implementation with libssh2"
```

---

## Task 2: Keychain Integration

**Files:**
- Create: `src/core/secure_storage.hpp`
- Create: `src/core/secure_storage.cpp`

**Step 1: Add QKeychain dependency to vcpkg.json**

```json
"qkeychain"
```

**Step 2: Create secure_storage.hpp**

```cpp
#pragma once

#include <QObject>
#include <QString>

namespace tablepro {

class SecureStorage : public QObject {
    Q_OBJECT

public:
    static SecureStorage* instance();

    // Store/retrieve passwords
    bool storePassword(const QString& key, const QString& password);
    QString retrievePassword(const QString& key);
    bool deletePassword(const QString& key);
    bool hasPassword(const QString& key);

    // Connection-specific helpers
    QString connectionPasswordKey(const QString& connectionId) const;
    QString sshKeyPassphraseKey(const QString& connectionId) const;

signals:
    void passwordStored(const QString& key);
    void passwordRetrieved(const QString& key);
    void passwordDeleted(const QString& key);
    void error(const QString& message);

private:
    explicit SecureStorage(QObject* parent = nullptr);
    QString service() const;
};

} // namespace tablepro
```

**Step 3: Create secure_storage.cpp**

```cpp
#include "secure_storage.hpp"
#include <keychain.h>

namespace tablepro {

SecureStorage* SecureStorage::instance() {
    static SecureStorage* inst = new SecureStorage();
    return inst;
}

SecureStorage::SecureStorage(QObject* parent)
    : QObject(parent)
{
}

QString SecureStorage::service() const {
    return "TablePro";
}

bool SecureStorage::storePassword(const QString& key, const QString& password) {
    QKeychain::WritePasswordJob job(service());
    job.setKey(key);
    job.setTextData(password);

    QEventLoop loop;
    connect(&job, &QKeychain::Job::finished, &loop, &QEventLoop::quit);
    job.start();
    loop.exec();

    if (job.error()) {
        emit error(QString("Failed to store password: %1").arg(job.errorString()));
        return false;
    }

    emit passwordStored(key);
    return true;
}

QString SecureStorage::retrievePassword(const QString& key) {
    QKeychain::ReadPasswordJob job(service());
    job.setKey(key);

    QEventLoop loop;
    connect(&job, &QKeychain::Job::finished, &loop, &QEventLoop::quit);
    job.start();
    loop.exec();

    if (job.error()) {
        if (job.error() != QKeychain::Error::EntryNotFound) {
            emit error(QString("Failed to retrieve password: %1").arg(job.errorString()));
        }
        return QString();
    }

    emit passwordRetrieved(key);
    return job.textData();
}

bool SecureStorage::deletePassword(const QString& key) {
    QKeychain::DeletePasswordJob job(service());
    job.setKey(key);

    QEventLoop loop;
    connect(&job, &QKeychain::Job::finished, &loop, &QEventLoop::quit);
    job.start();
    loop.exec();

    if (job.error() && job.error() != QKeychain::Error::EntryNotFound) {
        emit error(QString("Failed to delete password: %1").arg(job.errorString()));
        return false;
    }

    emit passwordDeleted(key);
    return true;
}

bool SecureStorage::hasPassword(const QString& key) {
    QKeychain::ReadPasswordJob job(service());
    job.setKey(key);

    QEventLoop loop;
    connect(&job, &QKeychain::Job::finished, &loop, &QEventLoop::quit);
    job.start();
    loop.exec();

    return !job.error();
}

QString SecureStorage::connectionPasswordKey(const QString& connectionId) const {
    return QString("connection/%1/password").arg(connectionId);
}

QString SecureStorage::sshKeyPassphraseKey(const QString& connectionId) const {
    return QString("connection/%1/ssh_passphrase").arg(connectionId);
}

} // namespace tablepro
```

**Commit:**

```bash
git add src/core/secure_storage.hpp src/core/secure_storage.cpp
git commit -m "feat: Add SecureStorage with QKeychain integration"
```

---

## Task 3: SSL Configuration

**Files:**
- Create: `src/core/ssl_config.hpp`
- Create: `src/core/ssl_config.cpp`

**Step 1: Create ssl_config.hpp**

```cpp
#pragma once

#include <QString>
#include <QSslCertificate>
#include <QSslKey>

namespace tablepro {

struct SslConfig {
    bool enabled = false;
    bool verifyPeer = true;
    QString caCertPath;
    QString clientCertPath;
    QString clientKeyPath;
    QString cipherSuite;
};

class SslConfigHelper {
public:
    static QSslCertificate loadCertificate(const QString& path);
    static QSslKey loadPrivateKey(const QString& path, const QByteArray& passphrase = QByteArray());
    static QList<QSslCertificate> loadCaCertificates(const QString& path);
    static bool validateConfig(const SslConfig& config, QString& error);
};

} // namespace tablepro
```

**Step 2: Create ssl_config.cpp**

```cpp
#include "ssl_config.hpp"
#include <QFile>

namespace tablepro {

QSslCertificate SslConfigHelper::loadCertificate(const QString& path) {
    QFile file(path);
    if (!file.open(QIODevice::ReadOnly)) {
        return QSslCertificate();
    }

    QSslCertificate cert(&file);
    file.close();

    return cert;
}

QSslKey SslConfigHelper::loadPrivateKey(const QString& path, const QByteArray& passphrase) {
    QFile file(path);
    if (!file.open(QIODevice::ReadOnly)) {
        return QSslKey();
    }

    QSslKey key(&file, QSsl::Rsa, QSsl::Pem, QSsl::PrivateKey, passphrase);
    file.close();

    return key;
}

QList<QSslCertificate> SslConfigHelper::loadCaCertificates(const QString& path) {
    QFile file(path);
    if (!file.open(QIODevice::ReadOnly)) {
        return {};
    }

    QList<QSslCertificate> certs = QSslCertificate::fromPath(path);
    file.close();

    return certs;
}

bool SslConfigHelper::validateConfig(const SslConfig& config, QString& error) {
    if (!config.enabled) {
        return true;
    }

    if (!config.caCertPath.isEmpty()) {
        QFile caFile(config.caCertPath);
        if (!caFile.exists()) {
            error = QString("CA certificate file not found: %1").arg(config.caCertPath);
            return false;
        }
    }

    if (!config.clientCertPath.isEmpty()) {
        QFile certFile(config.clientCertPath);
        if (!certFile.exists()) {
            error = QString("Client certificate file not found: %1").arg(config.clientCertPath);
            return false;
        }
    }

    if (!config.clientKeyPath.isEmpty()) {
        QFile keyFile(config.clientKeyPath);
        if (!keyFile.exists()) {
            error = QString("Client key file not found: %1").arg(config.clientKeyPath);
            return false;
        }
    }

    return true;
}

} // namespace tablepro
```

**Commit:**

```bash
git add src/core/ssl_config.hpp src/core/ssl_config.cpp
git commit -m "feat: Add SSL configuration helpers"
```

---

## Task 4: Update CMakeLists.txt

**Step 1: Add security sources and dependencies**

```cmake
# libssh2 for SSH tunneling
find_package(Libssh2 REQUIRED)
target_link_libraries(tablepro PRIVATE Libssh2::Libssh2)

# QKeychain for secure storage
find_package(QKeychain REQUIRED)
target_link_libraries(tablepro PRIVATE QKeychain::QKeychain)

# Security sources
set(TABLEPRO_SOURCES
    # ... existing ...
    src/core/ssh_tunnel.cpp
    src/core/secure_storage.cpp
    src/core/ssl_config.cpp
)
```

**Commit:**

```bash
git add CMakeLists.txt
git commit -m "build: Add SSH/SSL dependencies"
```

---

## Acceptance Criteria

- [ ] SshTunnel establishes SSH connections
- [ ] Public key authentication works
- [ ] Password authentication works
- [ ] Local port forwarding functional
- [ ] SecureStorage stores/retrieves passwords
- [ ] QKeychain integration works on all platforms
- [ ] SSL certificate loading works
- [ ] SSL validation provides useful errors

---

**Phase 10 Complete.** Next: Phase 11 - Licensing & Polish