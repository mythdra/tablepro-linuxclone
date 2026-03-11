"""Tests for database drivers and the ConnectionManager.

Requires local PostgreSQL and MySQL instances with a ``tablefree_test``
database.  See the project README or implementation_plan.md for setup
instructions.
"""

import pytest

from tablefree.db.config import ConnectionConfig, DriverType
from tablefree.db.driver import ColumnInfo, IndexInfo
from tablefree.db.manager import ConnectionManager
from tablefree.db.mysql_driver import MySQLDriver
from tablefree.db.postgres_driver import PostgreSQLDriver

# ────────────────────────────────────────────────────────────
# Helpers
# ────────────────────────────────────────────────────────────

_PG_TEMP_TABLE = """
    CREATE TABLE IF NOT EXISTS _tf_test (
        id   SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        age  INTEGER DEFAULT 0
    )
"""

_MYSQL_TEMP_TABLE = """
    CREATE TABLE IF NOT EXISTS _tf_test (
        id   INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        age  INT DEFAULT 0
    )
"""

_DROP_TABLE = "DROP TABLE IF EXISTS _tf_test"


# ════════════════════════════════════════════════════════════
# PostgreSQL Driver
# ════════════════════════════════════════════════════════════


class TestPostgreSQLDriver:
    """Integration tests for PostgreSQLDriver."""

    def test_connect_disconnect(self, pg_config: ConnectionConfig) -> None:
        driver = PostgreSQLDriver(pg_config)
        assert not driver.is_connected

        driver.connect()
        assert driver.is_connected

        driver.disconnect()
        assert not driver.is_connected

    def test_select_one(self, pg_config: ConnectionConfig) -> None:
        with PostgreSQLDriver(pg_config) as driver:
            result = driver.execute("SELECT 1 AS n")
            assert result == [{"n": 1}]

    def test_test_connection(self, pg_config: ConnectionConfig) -> None:
        driver = PostgreSQLDriver(pg_config)
        assert driver.test_connection() is True

    def test_get_schemas(self, pg_config: ConnectionConfig) -> None:
        with PostgreSQLDriver(pg_config) as driver:
            schemas = driver.get_schemas()
            assert "public" in schemas
            # System schemas should be excluded
            assert "pg_catalog" not in schemas
            assert "information_schema" not in schemas

    def test_get_tables(self, pg_config: ConnectionConfig) -> None:
        with PostgreSQLDriver(pg_config) as driver:
            driver.execute(_PG_TEMP_TABLE)
            try:
                tables = driver.get_tables()
                assert "_tf_test" in tables
            finally:
                driver.execute(_DROP_TABLE)

    def test_get_columns(self, pg_config: ConnectionConfig) -> None:
        with PostgreSQLDriver(pg_config) as driver:
            driver.execute(_PG_TEMP_TABLE)
            try:
                columns = driver.get_columns("_tf_test")
                assert len(columns) == 3

                col_names = [c.name for c in columns]
                assert col_names == ["id", "name", "age"]

                # Check types
                name_col = columns[1]
                assert isinstance(name_col, ColumnInfo)
                assert name_col.data_type == "character varying"
                assert name_col.is_nullable is False
            finally:
                driver.execute(_DROP_TABLE)

    def test_get_indexes(self, pg_config: ConnectionConfig) -> None:
        with PostgreSQLDriver(pg_config) as driver:
            driver.execute(_PG_TEMP_TABLE)
            try:
                indexes = driver.get_indexes("_tf_test")
                assert len(indexes) >= 1

                # Find the primary key index
                pk = [i for i in indexes if i.is_primary]
                assert len(pk) == 1
                assert isinstance(pk[0], IndexInfo)
                assert "id" in pk[0].columns
                assert pk[0].is_unique is True
            finally:
                driver.execute(_DROP_TABLE)

    def test_context_manager(self, pg_config: ConnectionConfig) -> None:
        driver = PostgreSQLDriver(pg_config)
        with driver:
            assert driver.is_connected
            result = driver.execute("SELECT 42 AS answer")
            assert result == [{"answer": 42}]
        assert not driver.is_connected


# ════════════════════════════════════════════════════════════
# MySQL Driver
# ════════════════════════════════════════════════════════════


class TestMySQLDriver:
    """Integration tests for MySQLDriver."""

    def test_connect_disconnect(self, mysql_config: ConnectionConfig) -> None:
        driver = MySQLDriver(mysql_config)
        assert not driver.is_connected

        driver.connect()
        assert driver.is_connected

        driver.disconnect()
        assert not driver.is_connected

    def test_select_one(self, mysql_config: ConnectionConfig) -> None:
        with MySQLDriver(mysql_config) as driver:
            result = driver.execute("SELECT 1 AS n")
            assert result == [{"n": 1}]

    def test_test_connection(self, mysql_config: ConnectionConfig) -> None:
        driver = MySQLDriver(mysql_config)
        assert driver.test_connection() is True

    def test_get_schemas(self, mysql_config: ConnectionConfig) -> None:
        with MySQLDriver(mysql_config) as driver:
            schemas = driver.get_schemas()
            assert "tablefree_test" in schemas
            # System databases should be excluded
            assert "information_schema" not in schemas
            assert "mysql" not in schemas

    def test_get_tables(self, mysql_config: ConnectionConfig) -> None:
        with MySQLDriver(mysql_config) as driver:
            driver.execute(_MYSQL_TEMP_TABLE)
            try:
                tables = driver.get_tables()
                assert "_tf_test" in tables
            finally:
                driver.execute(_DROP_TABLE)

    def test_get_columns(self, mysql_config: ConnectionConfig) -> None:
        with MySQLDriver(mysql_config) as driver:
            driver.execute(_MYSQL_TEMP_TABLE)
            try:
                columns = driver.get_columns("_tf_test")
                assert len(columns) == 3

                col_names = [c.name for c in columns]
                assert col_names == ["id", "name", "age"]

                name_col = columns[1]
                assert isinstance(name_col, ColumnInfo)
                assert name_col.data_type == "varchar"
                assert name_col.is_nullable is False
            finally:
                driver.execute(_DROP_TABLE)

    def test_get_indexes(self, mysql_config: ConnectionConfig) -> None:
        with MySQLDriver(mysql_config) as driver:
            driver.execute(_MYSQL_TEMP_TABLE)
            try:
                indexes = driver.get_indexes("_tf_test")
                assert len(indexes) >= 1

                pk = [i for i in indexes if i.is_primary]
                assert len(pk) == 1
                assert isinstance(pk[0], IndexInfo)
                assert "id" in pk[0].columns
                assert pk[0].is_unique is True
            finally:
                driver.execute(_DROP_TABLE)

    def test_context_manager(self, mysql_config: ConnectionConfig) -> None:
        driver = MySQLDriver(mysql_config)
        with driver:
            assert driver.is_connected
            result = driver.execute("SELECT 42 AS answer")
            assert result == [{"answer": 42}]
        assert not driver.is_connected


# ════════════════════════════════════════════════════════════
# ConnectionManager
# ════════════════════════════════════════════════════════════


class TestConnectionManager:
    """Integration tests for ConnectionManager."""

    def test_create_connection_pg(self, pg_config: ConnectionConfig) -> None:
        mgr = ConnectionManager()
        try:
            driver = mgr.create_connection("pg-1", pg_config)
            assert driver.is_connected
            assert "pg-1" in mgr.active_connections
        finally:
            mgr.close_all()

    def test_get_connection(self, pg_config: ConnectionConfig) -> None:
        mgr = ConnectionManager()
        try:
            mgr.create_connection("pg-1", pg_config)
            driver = mgr.get_connection("pg-1")
            assert driver.is_connected
            result = driver.execute("SELECT 1 AS n")
            assert result == [{"n": 1}]
        finally:
            mgr.close_all()

    def test_close_connection(self, pg_config: ConnectionConfig) -> None:
        mgr = ConnectionManager()
        driver = mgr.create_connection("pg-1", pg_config)
        mgr.close_connection("pg-1")
        assert not driver.is_connected
        assert "pg-1" not in mgr.active_connections

    def test_close_all(
        self,
        pg_config: ConnectionConfig,
        mysql_config: ConnectionConfig,
    ) -> None:
        mgr = ConnectionManager()
        pg_drv = mgr.create_connection("pg-1", pg_config)
        mysql_drv = mgr.create_connection("mysql-1", mysql_config)

        mgr.close_all()

        assert not pg_drv.is_connected
        assert not mysql_drv.is_connected
        assert len(mgr.active_connections) == 0

    def test_duplicate_id_raises(self, pg_config: ConnectionConfig) -> None:
        mgr = ConnectionManager()
        try:
            mgr.create_connection("dup", pg_config)
            with pytest.raises(ValueError, match="already exists"):
                mgr.create_connection("dup", pg_config)
        finally:
            mgr.close_all()
