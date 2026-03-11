"""Connection configuration dataclass."""

from dataclasses import dataclass, field
from enum import Enum


class DriverType(str, Enum):
    """Supported database driver types."""

    POSTGRESQL = "postgresql"
    MYSQL = "mysql"


@dataclass(frozen=True)
class ConnectionConfig:
    """Immutable configuration for a database connection.

    Attributes:
        host: Database server hostname or IP.
        port: Database server port.
        database: Name of the database to connect to.
        username: Authentication username.
        password: Authentication password.
        driver_type: Which database driver to use.
        name: Human-friendly label (e.g. "prod-users-db").
        ssl: Whether to use SSL for the connection.
        options: Driver-specific extra parameters.
    """

    host: str
    port: int
    database: str
    username: str
    password: str
    driver_type: DriverType
    name: str = ""
    ssl: bool = False
    options: dict = field(default_factory=dict)
