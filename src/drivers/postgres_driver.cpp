#include "postgres_driver.h"
#include "core/database_exceptions.h"
#include <libpq-fe.h>
#include <QByteArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QRegularExpression>
#include <QDebug>

namespace tablepro {

PostgresDriver::PostgresDriver(QObject *parent)
    : DatabaseDriver(parent)
    , m_pgconn(nullptr)
    , m_inTransaction(false)
{
}

PostgresDriver::~PostgresDriver()
{
    disconnect();
}

DatabaseType PostgresDriver::getType() const
{
    return DatabaseType::PostgreSQL;
}

bool PostgresDriver::isConnected() const
{
    return m_connected && m_pgconn != nullptr;
}

bool PostgresDriver::isInTransaction() const
{
    return m_inTransaction;
}

QString PostgresDriver::lastError() const
{
    return m_lastError;
}

QString PostgresDriver::buildConnectionString(const ConnectionConfig& config) const
{
    // Build PostgreSQL connection string
    // Format: host=... port=... dbname=... user=... password=...
    QString connStr;
    connStr += QString("host=%1 ").arg(config.host);
    connStr += QString("port=%1 ").arg(config.port);
    connStr += QString("dbname=%1 ").arg(config.database);
    connStr += QString("user=%1 ").arg(config.username);

    if (!config.password.isEmpty()) {
        connStr += QString("password=%1 ").arg(config.password);
    }

    if (config.useSsl) {
        connStr += "sslmode=require ";
        if (!config.sslCertPath.isEmpty()) {
            connStr += QString("sslcert=%1 ").arg(config.sslCertPath);
        }
    }

    connStr += QString("connect_timeout=%1").arg(config.timeout);

    return connStr.trimmed();
}

void PostgresDriver::updateErrorFromConnection()
{
    if (m_pgconn) {
        m_lastError = QString::fromUtf8(PQerrorMessage(m_pgconn));
    }
}

bool PostgresDriver::connect(const ConnectionConfig& config)
{
    // Validate configuration
    if (!config.isValid()) {
        m_lastError = "Invalid connection configuration: missing required fields";
        emit errorOccurred(m_lastError);
        return false;
    }

    // Disconnect any existing connection
    disconnect();

    // Build connection string and connect
    QString connStr = buildConnectionString(config);
    QByteArray connStrUtf8 = connStr.toUtf8();

    m_pgconn = PQconnectdb(connStrUtf8.constData());

    // Check connection status
    ConnStatusType status = PQstatus(m_pgconn);
    if (status != CONNECTION_OK) {
        updateErrorFromConnection();
        PQfinish(m_pgconn);
        m_pgconn = nullptr;
        emit errorOccurred(m_lastError);
        return false;
    }

    // Store config and update state
    m_config = config;
    m_connected = true;
    m_lastError.clear();

    emit connected();
    return true;
}

void PostgresDriver::disconnect()
{
    if (m_pgconn) {
        PQfinish(m_pgconn);
        m_pgconn = nullptr;
    }
    m_connected = false;
    m_inTransaction = false;
    emit disconnected();
}

QueryResult PostgresDriver::processResult(void* result)
{
    QueryResult queryResult;
    PGresult* pgResult = static_cast<PGresult*>(result);

    if (!pgResult) {
        queryResult.success = false;
        queryResult.errorMessage = "No result from query";
        return queryResult;
    }

    ExecStatusType status = PQresultStatus(pgResult);

    if (status == PGRES_TUPLES_OK || status == PGRES_SINGLE_TUPLE) {
        // Query returned data
        int nFields = PQnfields(pgResult);
        int nRows = PQntuples(pgResult);

        // Get column names
        queryResult.columnNames.reserve(nFields);
        for (int i = 0; i < nFields; ++i) {
            queryResult.columnNames.append(QString::fromUtf8(PQfname(pgResult, i)));
        }

        // Get rows
        queryResult.rows.reserve(nRows);
        for (int row = 0; row < nRows; ++row) {
            QVector<QVariant> rowData;
            rowData.reserve(nFields);
            for (int col = 0; col < nFields; ++col) {
                if (PQgetisnull(pgResult, row, col)) {
                    rowData.append(QVariant());
                } else {
                    char* value = PQgetvalue(pgResult, row, col);
                    rowData.append(QString::fromUtf8(value));
                }
            }
            queryResult.rows.append(rowData);
        }

        queryResult.success = true;
        queryResult.affectedRows = nRows;
    } else if (status == PGRES_COMMAND_OK) {
        // Non-SELECT query (INSERT, UPDATE, DELETE, etc.)
        queryResult.success = true;
        char* affectedRows = PQcmdTuples(pgResult);
        queryResult.affectedRows = affectedRows ? QString::fromUtf8(affectedRows).toInt() : 0;
    } else {
        // Error
        queryResult.success = false;
        queryResult.errorMessage = QString::fromUtf8(PQresultErrorMessage(pgResult));
        m_lastError = queryResult.errorMessage;
    }

    PQclear(pgResult);
    return queryResult;
}

QueryResult PostgresDriver::executeQuery(const QString& query,
                                         const QVector<QVariant>& params)
{
    QueryResult result;

    if (!isConnected()) {
        result.success = false;
        result.errorMessage = "Not connected to database";
        m_lastError = result.errorMessage;
        return result;
    }

    QByteArray queryUtf8 = query.toUtf8();

    PGresult* pgResult;
    if (params.isEmpty()) {
        pgResult = PQexec(m_pgconn, queryUtf8.constData());
    } else {
        // Convert parameters to char* array
        QVector<const char*> paramValues;
        QVector<int> paramLengths;
        QVector<int> paramFormats;
        QVector<QByteArray> paramData;

        paramValues.reserve(params.size());
        paramLengths.reserve(params.size());
        paramFormats.reserve(params.size());
        paramData.reserve(params.size());

        for (const QVariant& param : params) {
            QByteArray data = param.toString().toUtf8();
            paramData.append(data);
            paramValues.append(data.constData());
            paramLengths.append(data.size());
            paramFormats.append(0);  // 0 = text format
        }

        pgResult = PQexecParams(
            m_pgconn,
            queryUtf8.constData(),
            params.size(),
            nullptr,  // Let libpq infer types
            paramValues.constData(),
            paramLengths.constData(),
            paramFormats.constData(),
            0  // Result format: 0 = text
        );
    }

    result = processResult(pgResult);
    emit queryExecuted(query, result);
    return result;
}

QueryResult PostgresDriver::executeNonQuery(const QString& query,
                                            const QVector<QVariant>& params)
{
    return executeQuery(query, params);
}

QueryResult PostgresDriver::prepareAndExecute(const QString& query,
                                              const QVector<QVariant>& params)
{
    return executeQuery(query, params);
}

bool PostgresDriver::beginTransaction()
{
    if (!isConnected()) {
        m_lastError = "Not connected to database";
        return false;
    }

    QueryResult result = executeQuery("BEGIN");
    if (result.success) {
        m_inTransaction = true;
        return true;
    }
    return false;
}

bool PostgresDriver::commit()
{
    if (!isConnected()) {
        m_lastError = "Not connected to database";
        return false;
    }

    QueryResult result = executeQuery("COMMIT");
    if (result.success) {
        m_inTransaction = false;
        return true;
    }
    return false;
}

bool PostgresDriver::rollback()
{
    if (!isConnected()) {
        m_lastError = "Not connected to database";
        return false;
    }

    QueryResult result = executeQuery("ROLLBACK");
    if (result.success) {
        m_inTransaction = false;
        return true;
    }
    return false;
}

bool PostgresDriver::createSavepoint(const QString& name)
{
    if (!isConnected()) {
        m_lastError = "Not connected to database";
        return false;
    }

    if (!m_inTransaction) {
        m_lastError = "Not in a transaction";
        return false;
    }

    // Sanitize savepoint name (only allow alphanumeric and underscore)
    QString safeName = name;
    safeName.remove(QRegularExpression("[^a-zA-Z0-9_]"));

    if (safeName.isEmpty()) {
        m_lastError = "Invalid savepoint name";
        return false;
    }

    QueryResult result = executeQuery(QString("SAVEPOINT %1").arg(safeName));
    return result.success;
}

bool PostgresDriver::releaseSavepoint(const QString& name)
{
    if (!isConnected()) {
        m_lastError = "Not connected to database";
        return false;
    }

    QString safeName = name;
    safeName.remove(QRegularExpression("[^a-zA-Z0-9_]"));

    if (safeName.isEmpty()) {
        m_lastError = "Invalid savepoint name";
        return false;
    }

    QueryResult result = executeQuery(QString("RELEASE SAVEPOINT %1").arg(safeName));
    return result.success;
}

bool PostgresDriver::rollbackToSavepoint(const QString& name)
{
    if (!isConnected()) {
        m_lastError = "Not connected to database";
        return false;
    }

    QString safeName = name;
    safeName.remove(QRegularExpression("[^a-zA-Z0-9_]"));

    if (safeName.isEmpty()) {
        m_lastError = "Invalid savepoint name";
        return false;
    }

    QueryResult result = executeQuery(QString("ROLLBACK TO SAVEPOINT %1").arg(safeName));
    return result.success;
}

QueryResult PostgresDriver::getDatabases()
{
    return executeQuery("SELECT datname FROM pg_database WHERE datistemplate = false");
}

QueryResult PostgresDriver::getTables(const QString& database)
{
    Q_UNUSED(database)
    return executeQuery(
        "SELECT table_schema, table_name FROM information_schema.tables "
        "WHERE table_schema NOT IN ('pg_catalog', 'information_schema') "
        "ORDER BY table_schema, table_name"
    );
}

QueryResult PostgresDriver::getColumns(const QString& table, const QString& schema)
{
    QString query = QString(
        "SELECT column_name, data_type, is_nullable, column_default "
        "FROM information_schema.columns "
        "WHERE table_name = '%1'"
    ).arg(table);

    if (!schema.isEmpty()) {
        query += QString(" AND table_schema = '%1'").arg(schema);
    }

    query += " ORDER BY ordinal_position";
    return executeQuery(query);
}

QString PostgresDriver::getServerVersion()
{
    if (!isConnected()) {
        return QString();
    }

    QueryResult result = executeQuery("SELECT version()");
    if (result.success && result.rowCount() > 0) {
        return result.getValue(0, 0).toString();
    }
    return QString();
}

void PostgresDriver::executeQueryAsync(const QString& query,
                                       const QVector<QVariant>& params,
                                       std::function<void(const QueryResult&)> callback)
{
    QueryResult result = executeQuery(query, params);
    if (callback) {
        callback(result);
    }
}

// Type mapping implementation
PostgresType PostgresDriver::mapPostgresType(const QString& typeName) const
{
    QString type = typeName.toLower().trimmed();

    // Integer types
    if (type == "int4" || type == "integer" || type == "int") {
        return PostgresType::Integer;
    }
    if (type == "int8" || type == "bigint") {
        return PostgresType::BigInteger;
    }
    if (type == "int2" || type == "smallint") {
        return PostgresType::SmallInteger;
    }

    // String types
    if (type == "varchar" || type == "character varying") {
        return PostgresType::VarChar;
    }
    if (type == "text") {
        return PostgresType::Text;
    }
    if (type == "char" || type == "character") {
        return PostgresType::VarChar;
    }

    // Boolean
    if (type == "bool" || type == "boolean") {
        return PostgresType::Boolean;
    }

    // Floating point
    if (type == "float8" || type == "double precision") {
        return PostgresType::Double;
    }
    if (type == "float4" || type == "real") {
        return PostgresType::Float;
    }
    if (type == "numeric" || type == "decimal") {
        return PostgresType::Numeric;
    }

    // Date/Time
    if (type == "timestamp" || type == "timestamp without time zone") {
        return PostgresType::Timestamp;
    }
    if (type == "timestamptz" || type == "timestamp with time zone") {
        return PostgresType::TimestampTz;
    }
    if (type == "date") {
        return PostgresType::Date;
    }
    if (type == "time" || type == "time without time zone") {
        return PostgresType::Time;
    }

    // JSON
    if (type == "json") {
        return PostgresType::Json;
    }
    if (type == "jsonb") {
        return PostgresType::Jsonb;
    }

    // UUID
    if (type == "uuid") {
        return PostgresType::Uuid;
    }

    // Binary
    if (type == "bytea") {
        return PostgresType::ByteArray;
    }

    // Array (prefixed with _)
    if (type.startsWith("_") || type.endsWith("[]")) {
        return PostgresType::Array;
    }

    return PostgresType::Unknown;
}

QString PostgresDriver::postgresTypeToQString(PostgresType type) const
{
    switch (type) {
        case PostgresType::Integer: return "int4";
        case PostgresType::BigInteger: return "int8";
        case PostgresType::SmallInteger: return "int2";
        case PostgresType::VarChar: return "varchar";
        case PostgresType::Text: return "text";
        case PostgresType::Boolean: return "bool";
        case PostgresType::Double: return "float8";
        case PostgresType::Float: return "float4";
        case PostgresType::Numeric: return "numeric";
        case PostgresType::Timestamp: return "timestamp";
        case PostgresType::TimestampTz: return "timestamptz";
        case PostgresType::Date: return "date";
        case PostgresType::Time: return "time";
        case PostgresType::Json: return "json";
        case PostgresType::Jsonb: return "jsonb";
        case PostgresType::Uuid: return "uuid";
        case PostgresType::ByteArray: return "bytea";
        case PostgresType::Array: return "array";
        default: return "unknown";
    }
}

QJsonDocument PostgresDriver::parseJsonField(const QString& jsonStr) const
{
    if (jsonStr.isEmpty()) {
        return QJsonDocument();
    }

    QJsonParseError error;
    QJsonDocument doc = QJsonDocument::fromJson(jsonStr.toUtf8(), &error);

    if (error.error != QJsonParseError::NoError) {
        return QJsonDocument();
    }

    return doc;
}

QString PostgresDriver::formatJsonField(const QJsonDocument& doc) const
{
    return QString::fromUtf8(doc.toJson(QJsonDocument::Compact));
}

bool PostgresDriver::isValidUuid(const QString& uuid) const
{
    if (uuid.isEmpty()) {
        return false;
    }

    // Standard UUID format: 8-4-4-4-12 hex digits
    QRegularExpression uuidRegex(
        "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"
    );

    return uuidRegex.match(uuid).hasMatch();
}

QString PostgresDriver::formatUuid(const QString& uuid) const
{
    // Remove any non-hex characters and format as UUID
    QString clean = uuid.toLower().remove(QRegularExpression("[^0-9a-fA-F]"));

    if (clean.length() != 32) {
        return uuid;  // Return as-is if not valid
    }

    return QString("%1-%2-%3-%4-%5")
        .arg(clean.left(8))
        .arg(clean.mid(8, 4))
        .arg(clean.mid(12, 4))
        .arg(clean.mid(16, 4))
        .arg(clean.right(12));
}

QString PostgresDriver::escapeByteArray(const QByteArray& data) const
{
    // PostgreSQL hex format: \x followed by hex digits
    return "\\x" + data.toHex();
}

QByteArray PostgresDriver::unescapeByteArray(const QString& escaped) const
{
    if (!escaped.startsWith("\\x")) {
        return QByteArray();
    }

    QString hexPart = escaped.mid(2);
    return QByteArray::fromHex(hexPart.toUtf8());
}

bool PostgresDriver::tryReconnect()
{
    if (!m_config.isValid()) {
        logError("reconnect", "Cannot reconnect: invalid configuration");
        return false;
    }

    logInfo("Attempting to reconnect to database...");

    // Disconnect existing connection
    disconnect();

    // Attempt to reconnect with stored config
    const bool success = connect(m_config);

    if (success) {
        logInfo("Successfully reconnected to database");
    } else {
        logError("reconnect", QString("Reconnection failed: %1").arg(m_lastError));
    }

    return success;
}

void PostgresDriver::logError(const QString& context, const QString& message) const
{
    // Qt logging with context
    qWarning() << "[PostgresDriver]" << context << ":" << message;
}

void PostgresDriver::logInfo(const QString& message) const
{
    qDebug() << "[PostgresDriver]" << message;
}

// Connection Pooling Implementation
void PostgresDriver::setPoolConfig(const ConnectionPoolConfig& config)
{
    m_poolConfig = config;
}

void PostgresDriver::enablePooling(bool enable)
{
    if (m_poolEnabled == enable) {
        return;
    }

    if (enable) {
        initializePool();
    } else {
        shutdownPool();
    }

    m_poolEnabled = enable;
}

int PostgresDriver::activeConnectionCount() const
{
    return static_cast<int>(m_activeConnections.size());
}

int PostgresDriver::idleConnectionCount() const
{
    return static_cast<int>(m_idleConnections.size());
}

void PostgresDriver::initializePool()
{
    if (!m_config.isValid()) {
        logError("pool", "Cannot initialize pool: invalid configuration");
        return;
    }

    logInfo(QString("Initializing connection pool with %1 min connections").arg(m_poolConfig.minConnections));

    for (int i = 0; i < m_poolConfig.minConnections; ++i) {
        QString connStr = buildConnectionString(m_config);
        QByteArray connStrUtf8 = connStr.toUtf8();
        PGconn* conn = PQconnectdb(connStrUtf8.constData());

        if (PQstatus(conn) == CONNECTION_OK) {
            m_idleConnections.append(conn);
        } else {
            logError("pool", QString("Failed to create pool connection: %1")
                .arg(QString::fromUtf8(PQerrorMessage(conn))));
            PQfinish(conn);
        }
    }

    logInfo(QString("Pool initialized with %1 connections").arg(m_idleConnections.size()));
}

void PostgresDriver::shutdownPool()
{
    logInfo("Shutting down connection pool...");

    for (PGconn* conn : m_idleConnections) {
        if (conn) {
            PQfinish(conn);
        }
    }
    m_idleConnections.clear();

    for (PGconn* conn : m_activeConnections) {
        if (conn) {
            PQfinish(conn);
        }
    }
    m_activeConnections.clear();

    logInfo("Connection pool shut down");
}

PGconn* PostgresDriver::acquireFromPool()
{
    if (!m_poolEnabled || m_idleConnections.isEmpty()) {
        return nullptr;
    }

    PGconn* conn = m_idleConnections.takeLast();
    if (m_poolConfig.validateOnReturn) {
        if (PQstatus(conn) != CONNECTION_OK) {
            logError("pool", "Acquired connection is invalid, discarding");
            PQfinish(conn);
            return nullptr;
        }
    }

    m_activeConnections.append(conn);
    return conn;
}

void PostgresDriver::releaseToPool(PGconn* conn)
{
    if (!conn || !m_poolEnabled) {
        return;
    }

    const int index = m_activeConnections.indexOf(conn);
    if (index >= 0) {
        m_activeConnections.remove(index);
    }

    if (m_idleConnections.size() < m_poolConfig.maxConnections) {
        if (m_poolConfig.validateOnReturn && PQstatus(conn) != CONNECTION_OK) {
            logError("pool", "Released connection is invalid, discarding");
            PQfinish(conn);
            return;
        }
        m_idleConnections.append(conn);
    } else {
        PQfinish(conn);
    }
}

} // namespace tablepro