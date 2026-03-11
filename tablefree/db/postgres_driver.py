"""PostgreSQL driver implementation using psycopg2."""

from typing import Any

import psycopg2
import psycopg2.extras

from tablefree.db.config import ConnectionConfig
from tablefree.db.driver import ColumnInfo, DatabaseDriver, IndexInfo


class PostgreSQLDriver(DatabaseDriver):
    """PostgreSQL driver backed by psycopg2."""

    def connect(self) -> None:
        if self._connection is not None:
            return
        self._connection = psycopg2.connect(
            host=self._config.host,
            port=self._config.port,
            dbname=self._config.database,
            user=self._config.username,
            password=self._config.password,
            **self._config.options,
        )
        # Enable autocommit so metadata queries and SELECT don't need
        # explicit transaction handling.
        self._connection.autocommit = True

    def disconnect(self) -> None:
        if self._connection is not None:
            self._connection.close()
            self._connection = None

    def execute(
        self, query: str, params: tuple | None = None
    ) -> list[dict[str, Any]]:
        if self._connection is None:
            raise RuntimeError("Not connected — call connect() first")
        with self._connection.cursor(
            cursor_factory=psycopg2.extras.RealDictCursor,
        ) as cur:
            cur.execute(query, params)
            # Non-SELECT statements (DDL / DML) produce no rows.
            if cur.description is None:
                return []
            return [dict(row) for row in cur.fetchall()]

    # ── Metadata introspection ───────────────────────────────

    def get_schemas(self) -> list[str]:
        rows = self.execute(
            """
            SELECT schema_name
              FROM information_schema.schemata
             WHERE schema_name NOT LIKE 'pg_%%'
               AND schema_name <> 'information_schema'
             ORDER BY schema_name
            """
        )
        return [r["schema_name"] for r in rows]

    def get_tables(self, schema: str | None = None) -> list[str]:
        schema = schema or "public"
        rows = self.execute(
            """
            SELECT table_name
              FROM information_schema.tables
             WHERE table_schema = %s
               AND table_type = 'BASE TABLE'
             ORDER BY table_name
            """,
            (schema,),
        )
        return [r["table_name"] for r in rows]

    def get_columns(
        self, table: str, schema: str | None = None
    ) -> list[ColumnInfo]:
        schema = schema or "public"
        rows = self.execute(
            """
            SELECT column_name,
                   data_type,
                   is_nullable,
                   column_default,
                   ordinal_position
              FROM information_schema.columns
             WHERE table_schema = %s
               AND table_name   = %s
             ORDER BY ordinal_position
            """,
            (schema, table),
        )
        return [
            ColumnInfo(
                name=r["column_name"],
                data_type=r["data_type"],
                is_nullable=r["is_nullable"] == "YES",
                column_default=r["column_default"],
                ordinal_position=r["ordinal_position"],
            )
            for r in rows
        ]

    def get_indexes(
        self, table: str, schema: str | None = None
    ) -> list[IndexInfo]:
        schema = schema or "public"
        rows = self.execute(
            """
            SELECT i.relname            AS index_name,
                   ix.indisunique       AS is_unique,
                   ix.indisprimary      AS is_primary,
                   array_agg(a.attname ORDER BY k.n) AS columns
              FROM pg_class t
              JOIN pg_index ix     ON t.oid = ix.indrelid
              JOIN pg_class i      ON i.oid = ix.indexrelid
              JOIN pg_namespace ns ON ns.oid = t.relnamespace
              JOIN LATERAL unnest(ix.indkey) WITH ORDINALITY AS k(attnum, n)
                   ON TRUE
              JOIN pg_attribute a   ON a.attrelid = t.oid
                                   AND a.attnum   = k.attnum
             WHERE t.relname  = %s
               AND ns.nspname = %s
             GROUP BY i.relname, ix.indisunique, ix.indisprimary
             ORDER BY i.relname
            """,
            (table, schema),
        )
        return [
            IndexInfo(
                name=r["index_name"],
                columns=list(r["columns"]),
                is_unique=r["is_unique"],
                is_primary=r["is_primary"],
            )
            for r in rows
        ]
