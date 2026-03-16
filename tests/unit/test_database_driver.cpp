#include <QtTest/QtTest>
#include <QObject>
#include "core/database_driver.h"
#include "core/database_exceptions.h"

class TestDatabaseDriver : public QObject
{
    Q_OBJECT

private slots:
    void initTestCase();
    void cleanupTestCase();
    void testExceptionCreation();
    void testConnectionConfig();
    void testQueryResult();
};

void TestDatabaseDriver::initTestCase()
{
    // Initialization before running any test in this class
}

void TestDatabaseDriver::cleanupTestCase()
{
    // Cleanup after running all tests in this class
}

void TestDatabaseDriver::testExceptionCreation()
{
    try {
        throw tablepro::DatabaseException("Test exception");
        QVERIFY2(false, "Exception was not thrown");
    } catch (const tablepro::DatabaseException& ex) {
        QVERIFY(ex.message() == "Test exception");
    }

    try {
        throw tablepro::ConnectionException("Connection test");
        QVERIFY2(false, "ConnectionException was not thrown");
    } catch (const tablepro::ConnectionException& ex) {
        QVERIFY(ex.message() == "Connection test");
    }
}

void TestDatabaseDriver::testConnectionConfig()
{
    tablepro::ConnectionConfig config;
    QVERIFY(config.port == 0);
    QVERIFY(config.sshPort == 22);
    QVERIFY(config.timeout == 30);

    config.port = 5432;
    config.database = "testdb";
    config.username = "testuser";
    config.host = "localhost";

    QVERIFY(config.port == 5432);
    QVERIFY(config.database == "testdb");
    QVERIFY(config.username == "testuser");
    QVERIFY(config.host == "localhost");

    QVERIFY(config.isValid());  // Should be valid with host, port, database, and username
}

void TestDatabaseDriver::testQueryResult()
{
    tablepro::QueryResult result;
    QVERIFY(result.columnNames.isEmpty());
    QVERIFY(result.rows.isEmpty());
    QVERIFY(result.affectedRows == 0);
    QVERIFY(!result.success);
    QVERIFY(result.errorMessage.isEmpty());

    result.success = true;
    result.affectedRows = 5;
    result.errorMessage = "No error";

    QVERIFY(result.success);
    QVERIFY(result.affectedRows == 5);
    QVERIFY(result.errorMessage == "No error");

    // Add some fake data to test getValue method
    result.columnNames = {"id", "name", "age"};
    result.rows = {{"1", "Alice", "25"}, {"2", "Bob", "30"}};

    QVERIFY(result.getValue(0, 0).toString() == "1");
    QVERIFY(result.getValue(0, 1).toString() == "Alice");
    QVERIFY(result.getValue(1, 2).toString() == "30");

    QVERIFY(result.rowCount() == 2);
    QVERIFY(result.columnCount() == 3);
}

QTEST_MAIN(TestDatabaseDriver)
#include "test_database_driver.moc"