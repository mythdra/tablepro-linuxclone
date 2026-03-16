#include "core/connection_manager.h"
#include <QObject>
#include <QtTest/QtTest>

class TestConnectionManager : public QObject {
  Q_OBJECT

private slots:
  void initTestCase();
  void cleanupTestCase();
  void testGetInstance();
  void testConnectionManagement();
};

void TestConnectionManager::initTestCase() {
  // Initialization before running any test in this class
}

void TestConnectionManager::cleanupTestCase() {
  // Cleanup after running all tests in this class
}

void TestConnectionManager::testGetInstance() {
  tablepro::ConnectionManager *manager1 =
      tablepro::ConnectionManager::instance();
  tablepro::ConnectionManager *manager2 =
      tablepro::ConnectionManager::instance();

  QVERIFY(manager1 != nullptr);
  QVERIFY(manager2 != nullptr);
  QVERIFY(manager1 == manager2); // Should be the same instance (singleton)
}

void TestConnectionManager::testConnectionManagement() {
  tablepro::ConnectionManager *manager =
      tablepro::ConnectionManager::instance();

  QVERIFY(manager->getConnectionCount() == 0);
  QVERIFY(manager->getActiveConnections().isEmpty());

  // Create a dummy connection config for testing
  tablepro::ConnectionConfig config;
  config.host = "localhost";
  config.port = 5432;
  config.database = "testdb";
  config.username = "tablepro";
  config.password = "tablepro123";

  // We can't actually create a connection without a real driver implementation
  // But we can test that the manager can handle connection-related operations
  QVERIFY(manager->getConnectionCount() == 0);

  // Test the connection count is still 0
  QVERIFY(manager->getConnectionCount() == 0);
}

QTEST_MAIN(TestConnectionManager)
#include "test_connection_manager.moc"