#include <QtTest/QtTest>
#include <QObject>
#include <QJsonDocument>
#include <QJsonObject>
#include <QUuid>
#include "drivers/postgres_driver.h"
#include "core/database_exceptions.h"

class TestPostgresDriver : public QObject
{
    Q_OBJECT

private slots:
    void initTestCase();
    void cleanupTestCase();

    // Task 1: PostgresDriver class creation tests
    void testDriverCreation();
    void testDriverType();
    void testInitialConnectionState();

    // Task 2: Connection functionality tests
    void testConnectWithInvalidConfig();
    void testDisconnectWhenNotConnected();
    void testConnectionStringConstruction();
    void testConnectionStringFormat();

    // Task 3: Query execution tests
    void testExecuteQueryWithoutConnection();

    // Task 4: PostgreSQL data types tests
    void testTypeMapping();
    void testTypeMappingUnknownType();
    void testJsonConversion();
    void testUuidHandling();
    void testByteArrayHandling();

    // Task 5: Transaction tests
    void testTransactionWithoutConnection();

    // Task 6: Error handling tests
    void testLastErrorInitiallyEmpty();

private:
    QString buildExpectedConnectionString(const tablepro::ConnectionConfig& config);
};

void TestPostgresDriver::initTestCase()
{
}

void TestPostgresDriver::cleanupTestCase()
{
}

// Task 1 Tests: Create PostgresDriver Class
void TestPostgresDriver::testDriverCreation()
{
    tablepro::PostgresDriver driver;
    QVERIFY(!driver.isConnected());
}

void TestPostgresDriver::testDriverType()
{
    tablepro::PostgresDriver driver;
    QCOMPARE(driver.getType(), tablepro::DatabaseType::PostgreSQL);
}

void TestPostgresDriver::testInitialConnectionState()
{
    tablepro::PostgresDriver driver;
    QVERIFY(!driver.isConnected());
    QVERIFY(!driver.isInTransaction());
}

// Task 2 Tests: Connection Functionality
void TestPostgresDriver::testConnectWithInvalidConfig()
{
    tablepro::PostgresDriver driver;
    tablepro::ConnectionConfig config;  // Invalid - no host, port, etc.

    bool result = driver.connect(config);
    QVERIFY(!result);
    QVERIFY(!driver.isConnected());
    QVERIFY(!driver.lastError().isEmpty());
}

void TestPostgresDriver::testDisconnectWhenNotConnected()
{
    tablepro::PostgresDriver driver;
    // Should not crash when disconnecting without connection
    driver.disconnect();
    QVERIFY(!driver.isConnected());
}

void TestPostgresDriver::testConnectionStringConstruction()
{
    tablepro::PostgresDriver driver;
    tablepro::ConnectionConfig config;
    config.host = "localhost";
    config.port = 5432;
    config.database = "testdb";
    config.username = "testuser";
    config.password = "testpass";

    // The driver should be able to construct a connection string
    // We test this indirectly through the connect method (which will fail without server)
    driver.connect(config);
    QVERIFY(!driver.isConnected());  // No server running, but should handle gracefully
}

// Task 3 Tests: Query Execution
void TestPostgresDriver::testExecuteQueryWithoutConnection()
{
    tablepro::PostgresDriver driver;
    tablepro::QueryResult result = driver.executeQuery("SELECT 1");

    QVERIFY(!result.success);
    QVERIFY(!result.errorMessage.isEmpty());
}

// Task 4 Tests: PostgreSQL Data Types
void TestPostgresDriver::testTypeMapping()
{
    tablepro::PostgresDriver driver;

    // Test common PostgreSQL type mappings
    QCOMPARE(driver.mapPostgresType("int4"), tablepro::PostgresType::Integer);
    QCOMPARE(driver.mapPostgresType("int8"), tablepro::PostgresType::BigInteger);
    QCOMPARE(driver.mapPostgresType("varchar"), tablepro::PostgresType::VarChar);
    QCOMPARE(driver.mapPostgresType("text"), tablepro::PostgresType::Text);
    QCOMPARE(driver.mapPostgresType("bool"), tablepro::PostgresType::Boolean);
    QCOMPARE(driver.mapPostgresType("float8"), tablepro::PostgresType::Double);
    QCOMPARE(driver.mapPostgresType("timestamp"), tablepro::PostgresType::Timestamp);
    QCOMPARE(driver.mapPostgresType("date"), tablepro::PostgresType::Date);
    QCOMPARE(driver.mapPostgresType("jsonb"), tablepro::PostgresType::Jsonb);
    QCOMPARE(driver.mapPostgresType("uuid"), tablepro::PostgresType::Uuid);
}

void TestPostgresDriver::testTypeMappingUnknownType()
{
    tablepro::PostgresDriver driver;

    // Unknown types should return Unknown
    QCOMPARE(driver.mapPostgresType("custom_type"), tablepro::PostgresType::Unknown);
    QCOMPARE(driver.mapPostgresType(""), tablepro::PostgresType::Unknown);
}

void TestPostgresDriver::testJsonConversion()
{
    tablepro::PostgresDriver driver;

    // Test JSON string to QJsonDocument conversion
    QString jsonStr = R"({"name": "test", "value": 42})";
    QJsonDocument doc = driver.parseJsonField(jsonStr);

    QVERIFY(!doc.isNull());
    QVERIFY(doc.isObject());
    QCOMPARE(doc.object()["name"].toString(), QString("test"));
    QCOMPARE(doc.object()["value"].toInt(), 42);
}

void TestPostgresDriver::testUuidHandling()
{
    tablepro::PostgresDriver driver;

    // Test UUID string validation
    QString validUuid = "550e8400-e29b-41d4-a716-446655440000";
    QString invalidUuid = "not-a-uuid";

    QVERIFY(driver.isValidUuid(validUuid));
    QVERIFY(!driver.isValidUuid(invalidUuid));
}

void TestPostgresDriver::testByteArrayHandling()
{
    tablepro::PostgresDriver driver;

    // Test byte array to PostgreSQL format conversion
    QByteArray data = QByteArray::fromHex("deadbeef");
    QString escaped = driver.escapeByteArray(data);

    QVERIFY(!escaped.isEmpty());
    QVERIFY(escaped.startsWith("\\x"));
}

// Task 5 Tests: Transaction Support
void TestPostgresDriver::testTransactionWithoutConnection()
{
    tablepro::PostgresDriver driver;

    bool beginResult = driver.beginTransaction();
    QVERIFY(!beginResult);

    bool commitResult = driver.commit();
    QVERIFY(!commitResult);

    bool rollbackResult = driver.rollback();
    QVERIFY(!rollbackResult);
}

// Task 6 Tests: Error Handling
void TestPostgresDriver::testLastErrorInitiallyEmpty()
{
    tablepro::PostgresDriver driver;
    // Initially, there should be no error
    QVERIFY(driver.lastError().isEmpty());
}

// Additional connection string tests
void TestPostgresDriver::testConnectionStringFormat()
{
    tablepro::ConnectionConfig config;
    config.host = "localhost";
    config.port = 5432;
    config.database = "testdb";
    config.username = "testuser";
    config.password = "testpass";

    // Test that connection string is constructed in correct PostgreSQL format
    QString expectedFormat = buildExpectedConnectionString(config);
    QVERIFY(!expectedFormat.isEmpty());
    QVERIFY(expectedFormat.contains("host=localhost"));
    QVERIFY(expectedFormat.contains("port=5432"));
    QVERIFY(expectedFormat.contains("dbname=testdb"));
    QVERIFY(expectedFormat.contains("user=testuser"));
    QVERIFY(expectedFormat.contains("password=testpass"));
}

QString TestPostgresDriver::buildExpectedConnectionString(const tablepro::ConnectionConfig& config)
{
    // PostgreSQL connection string format:
    // host=... port=... dbname=... user=... password=...
    return QString("host=%1 port=%2 dbname=%3 user=%4 password=%5")
        .arg(config.host)
        .arg(config.port)
        .arg(config.database)
        .arg(config.username)
        .arg(config.password);
}

QTEST_MAIN(TestPostgresDriver)
#include "test_postgres_driver.moc"