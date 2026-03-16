#pragma once

#include <QObject>
#include <QString>
#include <QVector>
#include <QVariant>
#include <QJsonDocument>
#include <QByteArray>
#include "core/database_driver.h"

// Forward declaration of PostgreSQL connection struct
struct pg_conn;
typedef struct pg_conn PGconn;

namespace tablepro {

// PostgreSQL type enumeration for type mapping
enum class PostgresType {
    Unknown = 0,
    Integer,
    BigInteger,
    SmallInteger,
    VarChar,
    Text,
    Boolean,
    Double,
    Float,
    Numeric,
    Timestamp,
    TimestampTz,
    Date,
    Time,
    Json,
    Jsonb,
    Uuid,
    ByteArray,
    Array
};

// Connection pool configuration
struct ConnectionPoolConfig {
    int minConnections = 1;
    int maxConnections = 10;
    int connectionTimeout = 30;
    int idleTimeout = 300;
    bool validateOnReturn = true;
};

class PostgresDriver : public DatabaseDriver
{
    Q_OBJECT

public:
    explicit PostgresDriver(QObject *parent = nullptr);
    ~PostgresDriver() override;

    // Connection methods
    bool connect(const ConnectionConfig& config) override;
    void disconnect() override;
    bool isConnected() const override;

    // Query execution
    QueryResult executeQuery(const QString& query,
                            const QVector<QVariant>& params = {}) override;
    QueryResult executeNonQuery(const QString& query,
                               const QVector<QVariant>& params = {}) override;

    // Prepared statements
    QueryResult prepareAndExecute(const QString& query,
                                 const QVector<QVariant>& params = {}) override;

    // Transaction support
    bool beginTransaction() override;
    bool commit() override;
    bool rollback() override;
    bool isInTransaction() const override;

    // Savepoint support
    bool createSavepoint(const QString& name);
    bool releaseSavepoint(const QString& name);
    bool rollbackToSavepoint(const QString& name);

    // Schema introspection
    QueryResult getDatabases() override;
    QueryResult getTables(const QString& database = "") override;
    QueryResult getColumns(const QString& table, const QString& schema = "") override;

    // Utility methods
    QString getServerVersion() override;
    DatabaseType getType() const override;

    // Error handling
    QString lastError() const override;

    // Async execution
    void executeQueryAsync(const QString& query,
                          const QVector<QVariant>& params,
                          std::function<void(const QueryResult&)> callback) override;

    // Type mapping methods
    PostgresType mapPostgresType(const QString& typeName) const;
    QString postgresTypeToQString(PostgresType type) const;

    // JSON handling
    QJsonDocument parseJsonField(const QString& jsonStr) const;
    QString formatJsonField(const QJsonDocument& doc) const;

    // UUID handling
    bool isValidUuid(const QString& uuid) const;
    QString formatUuid(const QString& uuid) const;

    // Binary data handling
    QString escapeByteArray(const QByteArray& data) const;
    QByteArray unescapeByteArray(const QString& escaped) const;

    // Connection pooling
    void setPoolConfig(const ConnectionPoolConfig& config);
    ConnectionPoolConfig poolConfig() const { return m_poolConfig; }
    bool isPoolEnabled() const { return m_poolEnabled; }
    void enablePooling(bool enable);
    int activeConnectionCount() const;
    int idleConnectionCount() const;

private:
    PGconn* m_pgconn = nullptr;
    QString m_lastError;
    bool m_inTransaction = false;

    // Connection pooling
    ConnectionPoolConfig m_poolConfig;
    bool m_poolEnabled = false;
    QVector<PGconn*> m_idleConnections;
    QVector<PGconn*> m_activeConnections;

    // Helper methods
    QString buildConnectionString(const ConnectionConfig& config) const;
    void updateErrorFromConnection();
    QueryResult processResult(void* result);  // PGresult*

    // Connection recovery
    bool tryReconnect();

    // Logging helpers
    void logError(const QString& context, const QString& message) const;
    void logInfo(const QString& message) const;

    // Pool helpers
    PGconn* acquireFromPool();
    void releaseToPool(PGconn* conn);
    void initializePool();
    void shutdownPool();
};

} // namespace tablepro