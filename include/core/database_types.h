#pragma once

namespace tablepro {

enum class DatabaseType {
    PostgreSQL,
    MySQL,
    SQLite,
    DuckDB,
    SQLServer,
    ClickHouse,
    MongoDB,
    Redis
};

} // namespace tablepro