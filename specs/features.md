# Exhaustive Feature Specification: TablePro

This document lists every discrete feature, configuration option, and UI capability that must be built in the Qt/C++ rewrite to achieve parity with the current TablePro macOS application.

## 1. Connection Management
- **Connection Groups**: Grouping connections visually in the sidebar.
- **Color Tags**: 8 preset colors to visually distinct connections (red, orange, yellow, green, blue, purple, pink, gray).
- **Environment Tags**: Tag connections for easy filtering (e.g., Prod, Dev, Staging).
- **Database Types Supported**: 
  - Relational: MySQL, MariaDB, PostgreSQL, SQLite, MS SQL Server, Oracle, ClickHouse, DuckDB, Redshift.
  - NoSQL/K-V: MongoDB, Redis.
  *Downloadable Plugins System*: Certain drivers (Oracle, ClickHouse) are optionally downloaded to reduce binary size.

### Connection Configuration Fields
- **Basic:** Host, Port, Database, Username, Password.
- **File-based (SQLite/DuckDB):** File path picker to local `.db`/`.sqlite` files.
- **SSL Configurations:**
  - Modes: Disabled, Preferred, Required, Verify CA, Verify Identity.
  - Custom Paths: CA Certificate, Client Certificate, Client Key.
- **SSH Tunneling:**
  - Standard Tunnel: Host, Port, Username, Auth Method (Password, Private Key, SSH Agent).
  - Jump Hosts (Bastions): Ability to configure multiple chained jump hosts with their own Auth methods (Agent/Key).
  - Agent Sock: System Default, 1Password (`~/.1password/agent.sock`), or Custom Path.
  - Passphrase support for private keys.
- **Advanced / Specific:**
  - Startup Commands: Run custom SQL statements immediately after connection (e.g., `SET time_zone = '+00:00'`).
  - Pre-Connect Script: Run a shell script *before* the connection is made; abort if exit code != 0.
  - Pgpass Support: Auto-detect and parse `~/.pgpass` for PostgreSQL credentials.
  - Safe Mode Levels: Restrict query execution.
    - Silent: No restrictions.
    - Read-Only: Blocks all UPDATE/INSERT/DELETE/DROP statements entirely.
    - Confirm Prompt: Pops a dialog to confirm execution of any write query, or any query altogether depending on level.
  - AI Policy: Override global AI settings on a per-connection basis.
  - MongoDB specifics: Auth Source, Read Preference, Write Concern.
  - Custom Connection String parser (`postgresql://user:pass@host:5432/db`).

## 2. Main Windows & Tabs
- **Native Tabs Strategy**: Uses standard macOS-like Tabs (or custom Qt Tabs) per connection window.
- **Session Restoration**: Application remembers opened tabs and their SQL contents across restarts (serializes to JSON).
- **Tab Types**:
  - `Table Tab`: Browsing tabular data.
  - `Query Tab`: Blank SQL editor.
  - `Structure Tab`: Viewing/editing table DDL, indexes, and foreign keys.

## 3. SQL Editor
- **Tree-sitter Highlighting**: Real-time syntax highlighting resilient to massive single-line SQL dumps (capped at 10k chars per line for perf).
- **Autocomplete (IntelliSense)**: Context-aware suggestions pulling from the cached database schema (Tables, Columns, Keywords).
- **Multi-Cursor & Vim Mode**: Natively toggleable Vim keybindings and block-cursor mode.
- **AI Integrations**:
  - Inline Context Menu: Right-click selected SQL to "Explain with AI" or "Optimize with AI".
  - AI Chat Panel: A collapsable right sidebar for conversational generation of SQL queries using the current DB schema as context.
- **Execution Logic**:
  - "Run Current": Automatically extracts the statement under the cursor (split by `;`) and executes only that.
  - "Run Selection": Executes only the dragged text.
  - Auto-Pagination: Appends `LIMIT 10000` under the hood if the user's SELECT query lacks a limit clause.
- **Explain Plan**: Dedicated button to run `EXPLAIN` (or `EXPLAIN QUERY PLAN`, `EXPLAIN ANALYZE` depending on dialect) and view the AST/Plan output in a specialized tree or text view.

## 4. Data Grid & Results
- **Pagination**: Grid isn't infinitely rendered. Uses explicit offset/limit pages or Infinite Scroll mechanisms fetching batches.
- **Sorting**: Server-side sorting by clicking column headers. Emits precise `ORDER BY {col} {ASC/DESC}`.
- **Filtering UI**:
  - Global Search: Filters the entire table.
  - Column Filters: Dedicated UI input per column (e.g., `= Val`, `> Val`, `CONTAINS`, `IS NULL`). Maps to SQL `WHERE` clauses via `FilterSQLGenerator`.
- **Inline Editing (Change Tracking)**:
  - Double-click a cell to modify its content.
  - UI strictly visualizes changes (delta colors: Green for New Row, Yellow for modified cell, Red for deleted row).
  - Undo/Redo stack for un-saved cell modifications.
  - "Save Changes" button generates the safe dialect-specific SQL (using Primary Key or hidden ROWID to ensure exact row mutation).
- **Row Operations**: Duplicate Row, Delete selected rows, Copy Row as SQL/JSON.

## 5. Schema Viewing & Structure
- **Table List (Sidebar)**: Real-time filtering, schema grouping (e.g., `public`, `pg_catalog`).
- **Structure View**:
  - Columns: Name, Type, Nullable, Default, Auto Increment.
  - Indexes: Name, Columns, Unique/Primary.
  - Foreign Keys: Source, Target Table, Constraints (Cascade/Restrict).
  - DDL Preview: Raw `CREATE TABLE ...` statement generation.
- **DB Operations**: Create generic Database/Schema, Drop Table, Truncate Table.

## 6. Import & Export
- **Exporting**: Export Table or Query Result to CSV, JSON, Markdown (MQL), SQL inserts, or XLSX native.
- **Importing**: Import SQL dumps or CSV files directly into a specific table. Runs chunks sequentially avoiding memory blowups.

## 7. Global Settings & Preferences
- **Themes**: System Sync, Light, Dark.
- **Query History**: Automatically logs every successfully executed query. Stored in local SQLite. UI provides full-text search over history.
- **Font & Layout**: Configurable Editor Font, Results Grid Font, and Row Height.
- **AI Setup**: Input API Keys for OpenAI, Anthropic, or configure Custom OpenAI-compatible endpoints (Local LLM via Ollama).
