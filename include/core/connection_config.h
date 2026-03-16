#pragma once

#include <QString>
#include "database_types.h"

namespace tablepro {

struct ConnectionConfig {
    DatabaseType type;
    QString host;
    int port;
    QString database;
    QString username;
    QString password;
    bool useSsl = false;
    QString sslCertPath;
    QString sshHost;
    int sshPort = 22;
    QString sshUsername;
    QString sshPrivateKeyPath;
    int timeout = 30;

    // Default constructor
    ConnectionConfig() : type(DatabaseType::PostgreSQL), port(0), sshPort(22), useSsl(false), timeout(30) {}

    // Helper methods
    QString toString() const {
        return QString("%1@%2:%3/%4").arg(username).arg(host).arg(port).arg(database);
    }

    bool isValid() const {
        return !host.isEmpty() && port > 0 && !database.isEmpty() && !username.isEmpty();
    }
};

} // namespace tablepro