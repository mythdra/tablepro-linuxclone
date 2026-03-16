#include <QtTest/QtTest>
#include <QObject>
#include "drivers/postgres_driver.h"

/**
 * Integration tests for PostgresDriver with actual PostgreSQL database.
 *
 * Prerequisites:
 * - PostgreSQL container running (docker-compose up postgres)
 * - Connection: host=localhost, port=5432, user=tablepro, password=tablepro123, db=testdb
 */
class TestPostgresDriverIntegration : public QObject
{
    Q_OBJECT

private slots:
    void initTestCase();
    void cleanupTestCase();
    void cleanup();

    // Connection tests
    void testConnect();
    void testConnectWithWrongCredentials();
    void testDisconnect();

    // Query execution tests
    void testExecuteSelectQuery();
    void testExecuteInsertQuery();
    void testExecuteUpdateQuery();
    void testExecuteDeleteQuery();
    void testExecuteQueryWithParams();

    // Transaction tests
    void testTransactionCommit();
    void testTransactionRollback();
    void testSavepoint();

    // Data type tests
    void testJsonType();
    void testUuidType();
    void testByteArrayType();

    // Schema introspection tests
    void testGetDatabases();
    void testGetTables();
    void testGetColumns();
    void testGetServerVersion();

private:
    tablepro::PostgresDriver* m_driver = nullptr;
    tablepro::ConnectionConfig m_config;
    bool m_connected = false;

    void createTestTable();
    void dropTestTable();
};

void TestPostgresDriverIntegration::initTestCase()
{
    m_driver = new tablepro::PostgresDriver(this);

    // Configure connection from docker-compose
    m_config.host = "localhost";
    m_config.port = 5432;
    m_config.database = "testdb";
    m_config.username = "tablepro";
    m_config.password = "tablepro123";
    m_config.timeout = 10;
}

void TestPostgresDriverIntegration::cleanupTestCase()
{
    if (m_driver) {
        m_driver->disconnect();
        delete m_driver;
        m_driver = nullptr;
    }
}

void TestPostgresDriverIntegration::cleanup()
{
    if (m_connected) {
        m_driver->executeQuery("DROP TABLE IF EXISTS test_table");
    }
}

void TestPostgresDriverIntegration::createTestTable()
{
    m_driver->executeQuery(
        "CREATE TABLE IF NOT EXISTS test_table ("
        "id SERIAL PRIMARY KEY, "
        "name VARCHAR(100), "
        "value INTEGER, "
        "data JSONB, "
        "uuid_col UUID"
        ")"
    );
}

void TestPostgresDriverIntegration::dropTestTable()
{
    m_driver->executeQuery("DROP TABLE IF EXISTS test_table");
}

void TestPostgresDriverIntegration::testConnect()
{
    QVERIFY(m_driver != nullptr);

    bool result = m_driver->connect(m_config);
    QVERIFY2(result, qPrintable(m_driver->lastError()));
    QVERIFY(m_driver->isConnected());
    m_connected = true;
}

void TestPostgresDriverIntegration::testConnectWithWrongCredentials()
{
    tablepro::PostgresDriver driver;

    tablepro::ConnectionConfig wrongConfig;
    wrongConfig.host = "localhost";
    wrongConfig.port = 5432;
    wrongConfig.database = "nonexistent";
    wrongConfig.username = "wronguser";
    wrongConfig.password = "wrongpass";

    bool result = driver.connect(wrongConfig);
    QVERIFY(!result);
    QVERIFY(!driver.isConnected());
    QVERIFY(!driver.lastError().isEmpty());
}

void TestPostgresDriverIntegration::testDisconnect()
{
    if (!m_connected) {
        QSKIP("Not connected");
    }

    m_driver->disconnect();
    QVERIFY(!m_driver->isConnected());
    m_connected = false;

    bool result = m_driver->connect(m_config);
    QVERIFY(result);
    m_connected = true;
}

void TestPostgresDriverIntegration::testExecuteSelectQuery()
{
    if (!m_connected) QSKIP("Not connected");

    tablepro::QueryResult result = m_driver->executeQuery("SELECT 1 AS value, 'test' AS name");

    QVERIFY(result.success);
    QCOMPARE(result.rowCount(), 1);
    QCOMPARE(result.columnCount(), 2);
    QCOMPARE(result.columnNames[0], QString("value"));
    QCOMPARE(result.columnNames[1], QString("name"));
    QCOMPARE(result.getValue(0, 0).toInt(), 1);
    QCOMPARE(result.getValue(0, 1).toString(), QString("test"));
}

void TestPostgresDriverIntegration::testExecuteInsertQuery()
{
    if (!m_connected) QSKIP("Not connected");

    createTestTable();

    tablepro::QueryResult result = m_driver->executeNonQuery(
        "INSERT INTO test_table (name, value) VALUES ('test_name', 42)"
    );

    QVERIFY(result.success);
    QCOMPARE(result.affectedRows, 1);
}

void TestPostgresDriverIntegration::testExecuteUpdateQuery()
{
    if (!m_connected) QSKIP("Not connected");

    createTestTable();
    m_driver->executeNonQuery("INSERT INTO test_table (name, value) VALUES ('update_test', 10)");

    tablepro::QueryResult result = m_driver->executeNonQuery(
        "UPDATE test_table SET value = 20 WHERE name = 'update_test'"
    );

    QVERIFY(result.success);
    QCOMPARE(result.affectedRows, 1);
}

void TestPostgresDriverIntegration::testExecuteDeleteQuery()
{
    if (!m_connected) QSKIP("Not connected");

    createTestTable();
    m_driver->executeNonQuery("INSERT INTO test_table (name, value) VALUES ('delete_test', 99)");

    tablepro::QueryResult result = m_driver->executeNonQuery(
        "DELETE FROM test_table WHERE name = 'delete_test'"
    );

    QVERIFY(result.success);
    QCOMPARE(result.affectedRows, 1);
}

void TestPostgresDriverIntegration::testExecuteQueryWithParams()
{
    if (!m_connected) QSKIP("Not connected");

    createTestTable();

    QVector<QVariant> params;
    params << QVariant(QString("param_test")) << QVariant(123);

    tablepro::QueryResult result = m_driver->executeQuery(
        "INSERT INTO test_table (name, value) VALUES ($1, $2)",
        params
    );

    QVERIFY(result.success);

    params.clear();
    params << QVariant(QString("param_test"));

    result = m_driver->executeQuery(
        "SELECT name, value FROM test_table WHERE name = $1",
        params
    );

    QVERIFY(result.success);
    QCOMPARE(result.rowCount(), 1);
    QCOMPARE(result.getValue(0, 0).toString(), QString("param_test"));
    QCOMPARE(result.getValue(0, 1).toInt(), 123);
}

void TestPostgresDriverIntegration::testTransactionCommit()
{
    if (!m_connected) QSKIP("Not connected");

    createTestTable();

    QVERIFY(m_driver->beginTransaction());
    QVERIFY(m_driver->isInTransaction());

    m_driver->executeNonQuery("INSERT INTO test_table (name, value) VALUES ('tx_commit', 1)");

    QVERIFY(m_driver->commit());
    QVERIFY(!m_driver->isInTransaction());

    tablepro::QueryResult result = m_driver->executeQuery(
        "SELECT COUNT(*) FROM test_table WHERE name = 'tx_commit'"
    );
    QVERIFY(result.success);
    QCOMPARE(result.getValue(0, 0).toInt(), 1);
}

void TestPostgresDriverIntegration::testTransactionRollback()
{
    if (!m_connected) QSKIP("Not connected");

    createTestTable();

    QVERIFY(m_driver->beginTransaction());

    m_driver->executeNonQuery("INSERT INTO test_table (name, value) VALUES ('tx_rollback', 2)");

    QVERIFY(m_driver->rollback());
    QVERIFY(!m_driver->isInTransaction());

    tablepro::QueryResult result = m_driver->executeQuery(
        "SELECT COUNT(*) FROM test_table WHERE name = 'tx_rollback'"
    );
    QVERIFY(result.success);
    QCOMPARE(result.getValue(0, 0).toInt(), 0);
}

void TestPostgresDriverIntegration::testSavepoint()
{
    if (!m_connected) QSKIP("Not connected");

    createTestTable();

    QVERIFY(m_driver->beginTransaction());

    m_driver->executeNonQuery("INSERT INTO test_table (name, value) VALUES ('savepoint_test', 3)");

    QVERIFY(m_driver->createSavepoint("sp1"));

    m_driver->executeNonQuery("INSERT INTO test_table (name, value) VALUES ('savepoint_rollback', 4)");

    QVERIFY(m_driver->rollbackToSavepoint("sp1"));

    QVERIFY(m_driver->commit());

    tablepro::QueryResult result = m_driver->executeQuery(
        "SELECT COUNT(*) FROM test_table WHERE name = 'savepoint_rollback'"
    );
    QVERIFY(result.success);
    QCOMPARE(result.getValue(0, 0).toInt(), 0);
}

void TestPostgresDriverIntegration::testJsonType()
{
    if (!m_connected) QSKIP("Not connected");

    createTestTable();

    m_driver->executeNonQuery(
        "INSERT INTO test_table (name, data) VALUES ('json_test', '{\"key\": \"value\", \"number\": 42}'::jsonb)"
    );

    tablepro::QueryResult result = m_driver->executeQuery(
        "SELECT data FROM test_table WHERE name = 'json_test'"
    );

    QVERIFY(result.success);
    QCOMPARE(result.rowCount(), 1);

    QString jsonStr = result.getValue(0, 0).toString();
    QVERIFY(jsonStr.contains("key"));
    QVERIFY(jsonStr.contains("value"));
}

void TestPostgresDriverIntegration::testUuidType()
{
    if (!m_connected) QSKIP("Not connected");

    createTestTable();

    QString testUuid = "550e8400-e29b-41d4-a716-446655440000";

    m_driver->executeNonQuery(
        QString("INSERT INTO test_table (name, uuid_col) VALUES ('uuid_test', '%1'::uuid)").arg(testUuid)
    );

    tablepro::QueryResult result = m_driver->executeQuery(
        "SELECT uuid_col FROM test_table WHERE name = 'uuid_test'"
    );

    QVERIFY(result.success);
    QCOMPARE(result.rowCount(), 1);

    QString returnedUuid = result.getValue(0, 0).toString();
    QCOMPARE(returnedUuid, testUuid);
}

void TestPostgresDriverIntegration::testByteArrayType()
{
    if (!m_connected) QSKIP("Not connected");

    m_driver->executeQuery("DROP TABLE IF EXISTS bytea_test");
    m_driver->executeQuery("CREATE TABLE bytea_test (id SERIAL PRIMARY KEY, data BYTEA)");

    QByteArray testData = QByteArray::fromHex("deadbeef1234");

    tablepro::QueryResult result = m_driver->executeQuery(
        QString("INSERT INTO bytea_test (data) VALUES ('%1'::bytea)").arg(m_driver->escapeByteArray(testData))
    );
    QVERIFY(result.success);

    result = m_driver->executeQuery("SELECT data FROM bytea_test");
    QVERIFY(result.success);

    m_driver->executeQuery("DROP TABLE bytea_test");
}

void TestPostgresDriverIntegration::testGetDatabases()
{
    if (!m_connected) QSKIP("Not connected");

    tablepro::QueryResult result = m_driver->getDatabases();

    QVERIFY(result.success);
    QVERIFY(result.rowCount() > 0);

    bool foundTestDb = false;
    for (int i = 0; i < result.rowCount(); ++i) {
        if (result.getValue(i, 0).toString() == "testdb") {
            foundTestDb = true;
            break;
        }
    }
    QVERIFY(foundTestDb);
}

void TestPostgresDriverIntegration::testGetTables()
{
    if (!m_connected) QSKIP("Not connected");

    createTestTable();

    tablepro::QueryResult result = m_driver->getTables();

    QVERIFY(result.success);
    QVERIFY(result.rowCount() > 0);
}

void TestPostgresDriverIntegration::testGetColumns()
{
    if (!m_connected) QSKIP("Not connected");

    createTestTable();

    tablepro::QueryResult result = m_driver->getColumns("test_table", "public");

    QVERIFY(result.success);
    QVERIFY(result.rowCount() >= 5);
}

void TestPostgresDriverIntegration::testGetServerVersion()
{
    if (!m_connected) QSKIP("Not connected");

    QString version = m_driver->getServerVersion();

    QVERIFY(!version.isEmpty());
    QVERIFY(version.contains("PostgreSQL"));
    qDebug() << "Server version:" << version;
}

QTEST_MAIN(TestPostgresDriverIntegration)
#include "test_postgres_driver_integration.moc"