"""ConnectionManager — central registry for active database connections."""

from tablefree.db.config import ConnectionConfig, DriverType
from tablefree.db.driver import DatabaseDriver
from tablefree.db.mysql_driver import MySQLDriver
from tablefree.db.postgres_driver import PostgreSQLDriver


class ConnectionManager:
    """Manages active database connections.

    Each connection is identified by a unique string *connection_id*.
    The manager maps driver types to concrete driver classes and handles
    the full lifecycle: create → use → close.
    """

    _DRIVER_MAP: dict[DriverType, type[DatabaseDriver]] = {
        DriverType.POSTGRESQL: PostgreSQLDriver,
        DriverType.MYSQL: MySQLDriver,
    }

    def __init__(self) -> None:
        self._connections: dict[str, DatabaseDriver] = {}

    # ── Public API ───────────────────────────────────────────

    def create_connection(
        self, connection_id: str, config: ConnectionConfig
    ) -> DatabaseDriver:
        """Create, connect, and register a new driver instance.

        Raises:
            ValueError: If *connection_id* is already in use or the
                driver type is not supported.
        """
        if connection_id in self._connections:
            raise ValueError(f"Connection '{connection_id}' already exists")

        driver_cls = self._DRIVER_MAP.get(config.driver_type)
        if driver_cls is None:
            raise ValueError(f"Unsupported driver type: {config.driver_type}")

        driver = driver_cls(config)
        driver.connect()
        self._connections[connection_id] = driver
        return driver

    def get_connection(self, connection_id: str) -> DatabaseDriver:
        """Retrieve an active driver by its *connection_id*.

        Raises:
            KeyError: If no connection exists with that ID.
        """
        if connection_id not in self._connections:
            raise KeyError(f"No connection with ID '{connection_id}'")
        return self._connections[connection_id]

    def close_connection(self, connection_id: str) -> None:
        """Disconnect and remove a connection by its *connection_id*."""
        driver = self._connections.pop(connection_id, None)
        if driver is not None:
            driver.disconnect()

    def close_all(self) -> None:
        """Disconnect every active connection."""
        for conn_id in list(self._connections):
            self.close_connection(conn_id)

    # ── Properties ───────────────────────────────────────────

    @property
    def active_connections(self) -> dict[str, DatabaseDriver]:
        """Return a shallow copy of the active connections dict."""
        return dict(self._connections)
