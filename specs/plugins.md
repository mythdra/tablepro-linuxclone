# TablePro Plugin System & Database Drivers

## Overview
To keep the main application lightweight and modular, TablePro isolates all database driver implementations into dynamic plugins (`.tableplugin` bundles). They are loaded at runtime by `PluginManager` from the `PlugIns/` directory.

## TableProPluginKit
This is a shared static framework (`Plugins/TableProPluginKit/`) defining the contract between the main app and the individual driver bundles. 
Key protocols and structures:
1. `DriverPlugin`: The main entry point of a database plugin.
2. `PluginDatabaseDriver`: The core interface that every database driver must implement. Defines operations for querying, fetching schema (tables, columns, indexes, foreign keys, DDL), and transaction management.
3. `PluginQueryResult`, `PluginColumnInfo`, `PluginTableInfo`: Transfer data structures.

## Bridging (PluginDriverAdapter)
The main app defines a higher-level `DatabaseDriver` protocol in `TablePro/Core/Database/DatabaseDriver.swift`.
To communicate with plugins, `PluginDriverAdapter.swift` wraps the `PluginDatabaseDriver` provided by the plugin kit and adapts it to the app's internal interfaces.

## DatabaseManager (`DatabaseManager.swift`)
This is an `@Observable` Actor responsible for:
- Maintaining a list of active connection sessions (`ConnectionSession`).
- Handling SSH tunneling (using `SSHTunnelManager` and exponential backoff for tunnel death recovery).
- Connection health monitoring (background pinging: `SELECT 1`).
- Applying timeout configurations and startup commands.
- Sending UI-refresh notifications (`.databaseDidConnect`, `.refreshData`).

## C-Bridges
Because most database drivers (PostgreSQL, MySQL, SQL Server, MongoDB, Redis) are C libraries natively, each plugin contains an Objective-C/C bridge module to safely interface Swift with the C-API. 
Examples:
- MySQL/MariaDB: `CMariaDB`
- PostgreSQL: `CLibPQ`
- MSSQL: `CFreeTDS`
- MongoDB: `CLibMongoc`
- Redis: `CRedis`

*Notable Exceptions*: SQLite uses Foundation's built-in sqlite3. ClickHouse uses native URLSession HTTP. Oracle uses SPM library OracleNIO.

## Plugin Types
1. **Database Plugins**: MySQL, PostgreSQL, SQLite, ClickHouse, MSSQL, MongoDB, Redis, Oracle, DuckDB.
2. **Export/Import Plugins**: CSV, JSON, MQL, SQL, XLSX.

## Schema Switching Support
Certain drivers implement the `SchemaSwitchable` protocol (which aligns with `PluginDatabaseDriver`) to allow real-time switching of the search path (e.g., PostgreSQL `set search_path`, SQL Server `USE [db]`).
