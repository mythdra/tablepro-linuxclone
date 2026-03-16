# TablePro Tauri Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Set up Tauri 2.0 + React + PostgreSQL foundation with connection management and basic query execution

**Architecture:** Tauri 2.0 with Rust backend using sqlx for PostgreSQL, React 18 with TypeScript frontend using shadcn/ui, AG-Grid for data display, Monaco for SQL editor

**Tech Stack:** Tauri 2.0, Rust, sqlx, tokio, React 18, TypeScript, shadcn/ui, Tailwind CSS, AG-Grid, Monaco Editor, Zustand, React Query

---

## Phase 1: Project Setup

### Task 1: Create Tauri Project with React Template

**Files:**
- Create: `tablepro/` (new directory)
- Test: Verify project builds

**Step 1: Initialize Tauri project**

```bash
# Install Rust if not already installed
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source ~/.cargo/env

# Install Tauri CLI
cargo install tauri-cli

# Create Tauri project with React TypeScript template
npm create tauri-app@latest tablepro -- --template react-ts --manager npm
cd tablepro
```

**Step 2: Verify dev server runs**

```bash
cd tablepro
npm install
npm run tauri dev
```

Expected: App opens with default Tauri template

**Step 3: Commit**

```bash
git init
git add .
git commit -m "feat: initialize Tauri project with React TypeScript template"
```

---

### Task 2: Configure Tailwind CSS and shadcn/ui

**Files:**
- Modify: `tablepro/package.json`
- Create: `tablepro/tailwind.config.js`
- Create: `tablepro/postcss.config.js`
- Create: `tablepro/src/lib/utils.ts`
- Create: `tablepro/src/components/ui/button.tsx`
- Create: `tablepro/src/components/ui/input.tsx`
- Create: `tablepro/src/components/ui/dialog.tsx`
- Create: `tablepro/src/components/ui/label.tsx`
- Create: `tablepro/src/components/ui/select.tsx`
- Create: `tablepro/src/components/ui/tabs.tsx`
- Create: `tablepro/src/components/ui/card.tsx`

**Step 1: Install Tailwind CSS dependencies**

```bash
cd tablepro
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

**Step 2: Configure tailwind.config.js**

```javascript
/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
    },
  },
  plugins: [],
}
```

**Step 3: Update src/index.css**

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --card: 0 0% 100%;
    --card-foreground: 222.2 84% 4.9%;
    --popover: 0 0% 100%;
    --popover-foreground: 222.2 84% 4.9%;
    --primary: 221.2 83.2% 53.3%;
    --primary-foreground: 210 40% 98%;
    --secondary: 210 40% 96.1%;
    --secondary-foreground: 222.2 47.4% 11.2%;
    --muted: 210 40% 96.1%;
    --muted-foreground: 215.4 16.3% 46.9%;
    --accent: 210 40% 96.1%;
    --accent-foreground: 222.2 47.4% 11.2%;
    --destructive: 0 84.2% 60.2%;
    --destructive-foreground: 210 40% 98%;
    --border: 214.3 31.8% 91.4%;
    --input: 214.3 31.8% 91.4%;
    --ring: 221.2 83.2% 53.3%;
    --radius: 0.5rem;
  }

  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    --card: 222.2 84% 4.9%;
    --card-foreground: 210 40% 98%;
    --popover: 222.2 84% 4.9%;
    --popover-foreground: 210 40% 98%;
    --primary: 217.2 91.2% 59.8%;
    --primary-foreground: 222.2 47.4% 11.2%;
    --secondary: 217.2 32.6% 17.5%;
    --secondary-foreground: 210 40% 98%;
    --muted: 217.2 32.6% 17.5%;
    --muted-foreground: 215 20.2% 65.1%;
    --accent: 217.2 32.6% 17.5%;
    --accent-foreground: 210 40% 98%;
    --destructive: 0 62.8% 30.6%;
    --destructive-foreground: 210 40% 98%;
    --border: 217.2 32.6% 17.5%;
    --input: 217.2 32.6% 17.5%;
    --ring: 224.3 76.3% 48%;
  }
}

@layer base {
  * {
    @apply border-border;
  }
  body {
    @apply bg-background text-foreground;
  }
}
```

**Step 4: Create lib/utils.ts**

```typescript
import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

**Step 5: Install shadcn/ui dependencies**

```bash
cd tablepro
npm install clsx tailwind-merge lucide-react
```

**Step 6: Create Button component**

```typescript
import * as React from "react"
import { cn } from "@/lib/utils"

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "default" | "destructive" | "outline" | "secondary" | "ghost" | "link"
  size?: "default" | "sm" | "lg" | "icon"
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", size = "default", ...props }, ref) => {
    return (
      <button
        className={cn(
          "inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
          {
            "bg-primary text-primary-foreground hover:bg-primary/90": variant === "default",
            "bg-destructive text-destructive-foreground hover:bg-destructive/90": variant === "destructive",
            "border border-input bg-background hover:bg-accent hover:text-accent-foreground": variant === "outline",
            "bg-secondary text-secondary-foreground hover:bg-secondary/80": variant === "secondary",
            "hover:bg-accent hover:text-accent-foreground": variant === "ghost",
            "text-primary underline-offset-4 hover:underline": variant === "link",
            "h-10 px-4 py-2": size === "default",
            "h-9 rounded-md px-3": size === "sm",
            "h-11 rounded-md px-8": size === "lg",
            "h-10 w-10": size === "icon",
          },
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"

export { Button }
```

**Step 7: Create other UI components (Input, Dialog, Label, Select, Tabs, Card)**

Create each component following the same pattern - see shadcn/ui repository for complete code

**Step 8: Commit**

```bash
git add .
git commit -m "feat: configure Tailwind CSS and shadcn/ui components"
```

---

### Task 3: Set Up Rust Backend Dependencies

**Files:**
- Modify: `tablepro/src-tauri/Cargo.toml`
- Modify: `tablepro/src-tauri/tauri.conf.json`

**Step 1: Update Cargo.toml**

```toml
[package]
name = "tablepro"
version = "0.1.0"
edition = "2021"

[lib]
name = "tablepro_lib"
crate-type = ["staticlib", "cdylib", "rlib"]

[dependencies]
# Tauri
tauri = { version = "2", features = ["devtools"] }
tauri-plugin-shell = "2"

# Database
sqlx = { version = "0.8", features = ["runtime-tokio", "postgres", "uuid", "chrono", "json"] }

# Async runtime
tokio = { version = "1", features = ["full"] }

# Error handling
thiserror = "2"

# Serialization
serde = { version = "1", features = ["derive"] }
serde_json = "1"

# Logging
tracing = "0.1"
tracing-subscriber = { version = "0.3", features = ["env-filter"] }

# Time
chrono = { version = "0.4", features = ["serde"] }

# UUID
uuid = { version = "1", features = ["v4", "serde"] }

# Async trait
async-trait = "0.1"

[build-dependencies]
tauri-build = { version = "2", features = [] }

[profile.release]
panic = "abort"
codegen-units = 1
lto = true
opt-level = "s"
strip = true
```

**Step 2: Update tauri.conf.json**

```json
{
  "$schema": "https://schema.tauri.app/config/2",
  "productName": "TablePro",
  "version": "0.1.0",
  "identifier": "com.tablepro.app",
  "build": {
    "beforeBuildCommand": "npm run build",
    "beforeDevCommand": "npm run dev",
    "devUrl": "http://localhost:1420",
    "frontendDist": "../dist",
    "devtools": true
  },
  "app": {
    "withGlobalTauri": true,
    "windows": [
      {
        "title": "TablePro",
        "width": 1280,
        "height": 800,
        "minWidth": 800,
        "minHeight": 600,
        "resizable": true,
        "fullscreen": false,
        "center": true
      }
    ],
    "security": {
      "csp": null
    }
  },
  "bundle": {
    "active": true,
    "targets": "all"
  }
}
```

**Step 3: Verify Rust compiles**

```bash
cd tablepro/src-tauri
cargo check
```

Expected: No errors

**Step 4: Commit**

```bash
git add .
git commit -m "feat: configure Rust dependencies for PostgreSQL"
```

---

## Phase 2: Rust Backend Core

### Task 4: Create Error Types

**Files:**
- Create: `tablepro/src-tauri/src/error.rs`

**Step 1: Write the failing test**

```rust
// src-tauri/src/lib.rs
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_error_serialization() {
        let err = AppError::NotConnected("test-id".to_string());
        let json = serde_json::to_string(&err).unwrap();
        assert!(json.contains("test-id"));
    }
}
```

**Step 2: Run test to verify it fails**

```bash
cd tablepro/src-tauri
cargo test
```

Expected: FAIL - AppError not defined

**Step 3: Write error.rs**

```rust
use thiserror::Error;

#[derive(Debug, Error)]
pub enum AppError {
    #[error("Database error: {0}")]
    Database(#[from] sqlx::Error),

    #[error("Connection not found: {0}")]
    NotConnected(String),

    #[error("Invalid configuration: {0}")]
    InvalidConfig(String),

    #[error("Query execution failed: {0}")]
    QueryFailed(String),
}

impl serde::Serialize for AppError {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        serializer.serialize_str(&self.to_string())
    }
}
```

**Step 4: Update lib.rs**

```rust
pub mod error;
pub use error::AppError;
```

**Step 5: Run test to verify it passes**

```bash
cd tablepro/src-tauri
cargo test
```

Expected: PASS

**Step 6: Commit**

```bash
git add .
git commit -m "feat: add error types"
```

---

### Task 5: Create Database Types

**Files:**
- Create: `tablepro/src-tauri/src/db/types.rs`

**Step 1: Write types.rs**

```rust
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConnectionConfig {
    pub id: String,
    pub name: String,
    pub host: String,
    pub port: u16,
    pub database: String,
    pub username: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub password: Option<String>,
    pub ssl_mode: SslMode,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub ssh_config: Option<SshConfig>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SslMode {
    Disable,
    Require,
    VerifyFull,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SshConfig {
    pub host: String,
    pub port: u16,
    pub username: String,
    pub auth_type: SshAuthType,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub private_key: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub passphrase: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SshAuthType {
    Password,
    Key,
    Agent,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConnectionInfo {
    pub id: String,
    pub name: String,
    pub database: String,
    pub server_version: String,
    pub connected_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct QueryResult {
    pub columns: Vec<ColumnMeta>,
    pub rows: Vec<serde_json::Value>,
    pub row_count: usize,
    pub execution_time_ms: u64,
    pub truncated: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ColumnMeta {
    pub name: String,
    pub type_: String,
    pub nullable: bool,
    pub primary_key: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TableInfo {
    pub name: String,
    pub schema: String,
    pub r#type: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ColumnInfo {
    pub name: String,
    pub type_: String,
    pub nullable: bool,
    pub default_value: Option<String>,
    pub is_primary_key: bool,
    pub is_foreign_key: bool,
}
```

**Step 2: Commit**

```bash
git add .
git commit -m "feat: add database types"
```

---

### Task 6: Create Connection Pool Manager

**Files:**
- Create: `tablepro/src-tauri/src/db/pool.rs`
- Create: `tablepro/src-tauri/src/db/mod.rs`

**Step 1: Write pool.rs**

```rust
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use sqlx::postgres::{PgPool, PgPoolOptions};
use crate::db::types::ConnectionConfig;
use crate::error::AppError;

pub struct ConnectionPool {
    pools: Arc<RwLock<HashMap<String, PgPool>>>,
}

impl ConnectionPool {
    pub fn new() -> Self {
        Self {
            pools: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub async fn connect(&self, config: &ConnectionConfig) -> Result<(), AppError> {
        let connect_string = format!(
            "postgres://{}@{}:{}/{}",
            config.username,
            config.host,
            config.port,
            config.database
        );

        let pool = PgPoolOptions::new()
            .max_connections(5)
            .connect(&connect_string)
            .await?;

        let mut pools = self.pools.write().await;
        pools.insert(config.id.clone(), pool);
        Ok(())
    }

    pub async fn disconnect(&self, connection_id: &str) -> Result<(), AppError> {
        let mut pools = self.pools.write().await;
        if let Some(pool) = pools.remove(connection_id) {
            pool.close().await;
        }
        Ok(())
    }

    pub async fn get(&self, connection_id: &str) -> Result<PgPool, AppError> {
        let pools = self.pools.read().await;
        pools.get(connection_id)
            .cloned()
            .ok_or_else(|| AppError::NotConnected(connection_id.to_string()))
    }

    pub async fn is_connected(&self, connection_id: &str) -> bool {
        let pools = self.pools.read().await;
        pools.contains_key(connection_id)
    }
}

impl Default for ConnectionPool {
    fn default() -> Self {
        Self::new()
    }
}
```

**Step 2: Write db/mod.rs**

```rust
pub mod types;
pub mod pool;

pub use pool::ConnectionPool;
pub use types::*;
```

**Step 3: Update lib.rs**

```rust
pub mod db;
pub use db::ConnectionPool;
```

**Step 4: Verify compilation**

```bash
cd tablepro/src-tauri
cargo check
```

Expected: No errors

**Step 5: Commit**

```bash
git add .
git commit -m "feat: add connection pool manager"
```

---

### Task 7: Create Query Commands

**Files:**
- Create: `tablepro/src-tauri/src/commands/mod.rs`
- Create: `tablepro/src-tauri/src/commands/connection.rs`
- Create: `tablepro/src-tauri/src/commands/query.rs`
- Create: `tablepro/src-tauri/src/commands/schema.rs`

**Step 1: Write connection.rs**

```rust
use tauri::State;
use crate::db::{ConnectionPool, ConnectionConfig, ConnectionInfo};
use crate::error::AppError;

#[tauri::command]
pub async fn connect(
    config: ConnectionConfig,
    pool: State<'_, ConnectionPool>,
) -> Result<ConnectionInfo, String> {
    pool.connect(&config)
        .await
        .map_err(|e| e.to_string())?;

    let pg_pool = pool.get(&config.id)
        .await
        .map_err(|e| e.to_string())?;

    let version: String = sqlx::query_scalar("SELECT version()")
        .fetch_one(&*pg_pool)
        .await
        .map_err(|e| e.to_string())?;

    Ok(ConnectionInfo {
        id: config.id,
        name: config.name,
        database: config.database,
        server_version: version,
        connected_at: chrono::Utc::now().to_rfc3339(),
    })
}

#[tauri::command]
pub async fn disconnect(
    connection_id: String,
    pool: State<'_, ConnectionPool>,
) -> Result<(), String> {
    pool.disconnect(&connection_id)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn test_connection(config: ConnectionConfig) -> Result<bool, String> {
    let connect_string = format!(
        "postgres://{}@{}:{}/{}",
        config.username,
        config.host,
        config.port,
        config.database
    );

    let pool = PgPool::connect(&connect_string)
        .await
        .map_err(|e| e.to_string())?;

    sqlx::query("SELECT 1")
        .fetch_one(&pool)
        .await
        .map_err(|e| e.to_string())?;

    pool.close().await;
    Ok(true)
}
```

**Step 2: Write query.rs**

```rust
use tauri::State;
use crate::db::{ConnectionPool, QueryResult, ColumnMeta};
use sqlx::Row;

#[tauri::command]
pub async fn execute_query(
    connection_id: String,
    sql: String,
    limit: Option<usize>,
    pool: State<'_, ConnectionPool>,
) -> Result<QueryResult, String> {
    let pg_pool = pool.get(&connection_id)
        .await
        .map_err(|e| e.to_string())?;

    let start = std::time::Instant::now();

    let query = if let Some(limit) = limit {
        format!("{} LIMIT {}", sql.trim_end_matches(';'), limit)
    } else {
        sql
    };

    let rows = sqlx::query(&query)
        .fetch_all(&*pg_pool)
        .await
        .map_err(|e| e.to_string())?;

    let columns: Vec<ColumnMeta> = if !rows.is_empty() {
        rows[0].columns().iter()
            .map(|col| ColumnMeta {
                name: col.name().to_string(),
                type_: format!("{:?}", col.type_info()),
                nullable: true,
                primary_key: false,
            })
            .collect()
    } else {
        vec![]
    };

    let row_data: Vec<serde_json::Value> = rows.iter()
        .map(|row| {
            let mut map = serde_json::Map::new();
            for (i, col) in row.columns().iter().enumerate() {
                let value: serde_json::Value = match row.try_get::<Option<String>, _>(i) {
                    Ok(Some(v)) => serde_json::Value::String(v),
                    Ok(None) => serde_json::Value::Null,
                    Err(_) => serde_json::Value::String(format!("{:?}", row[i])),
                };
                map.insert(col.name().to_string(), value);
            }
            serde_json::Value::Object(map)
        })
        .collect();

    Ok(QueryResult {
        columns,
        rows: row_data,
        row_count: rows.len(),
        execution_time_ms: start.elapsed().as_millis() as u64,
        truncated: false,
    })
}
```

**Step 3: Write schema.rs**

```rust
use tauri::State;
use crate::db::{ConnectionPool, TableInfo, ColumnInfo};

#[tauri::command]
pub async fn get_schemas(
    connection_id: String,
    pool: State<'_, ConnectionPool>,
) -> Result<Vec<String>, String> {
    let pg_pool = pool.get(&connection_id)
        .await
        .map_err(|e| e.to_string())?;

    let schemas: Vec<String> = sqlx::query_scalar(
        "SELECT schema_name FROM information_schema.schemata
         WHERE schema_name NOT IN ('pg_catalog', 'information_schema')
         ORDER BY schema_name"
    )
    .fetch_all(&*pg_pool)
    .await
    .map_err(|e| e.to_string())?;

    Ok(schemas)
}

#[tauri::command]
pub async fn get_tables(
    connection_id: String,
    schema: String,
    pool: State<'_, ConnectionPool>,
) -> Result<Vec<TableInfo>, String> {
    let pg_pool = pool.get(&connection_id)
        .await
        .map_err(|e| e.to_string())?;

    let rows = sqlx::query(
        "SELECT table_name, table_schema, table_type
         FROM information_schema.tables
         WHERE table_schema = $1
         ORDER BY table_name"
    )
    .bind(&schema)
    .fetch_all(&*pg_pool)
    .await
    .map_err(|e| e.to_string())?;

    let tables: Vec<TableInfo> = rows.iter()
        .map(|row| {
            let table_type: String = row.get("table_type");
            TableInfo {
                name: row.get("table_name"),
                schema: row.get("table_schema"),
                r#type: match table_type.as_str() {
                    "BASE TABLE" => "table".to_string(),
                    "VIEW" => "view".to_string(),
                    _ => table_type,
                },
            }
        })
        .collect();

    Ok(tables)
}

#[tauri::command]
pub async fn get_columns(
    connection_id: String,
    schema: String,
    table: String,
    pool: State<'_, ConnectionPool>,
) -> Result<Vec<ColumnInfo>, String> {
    let pg_pool = pool.get(&connection_id)
        .await
        .map_err(|e| e.to_string())?;

    let rows = sqlx::query(
        "SELECT
            c.column_name,
            c.data_type,
            c.is_nullable,
            c.column_default,
            CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary_key,
            CASE WHEN fk.column_name IS NOT NULL THEN true ELSE false END as is_foreign_key
        FROM information_schema.columns c
        LEFT JOIN (
            SELECT ku.column_name
            FROM information_schema.table_constraints tc
            JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
            WHERE tc.constraint_type = 'PRIMARY KEY'
        ) pk ON c.column_name = pk.column_name AND c.table_name = pk.table_name
        LEFT JOIN (
            SELECT ku.column_name
            FROM information_schema.table_constraints tc
            JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
            WHERE tc.constraint_type = 'FOREIGN KEY'
        ) fk ON c.column_name = fk.column_name AND c.table_name = fk.table_name
        WHERE c.table_schema = $1 AND c.table_name = $2
        ORDER BY c.ordinal_position"
    )
    .bind(&schema)
    .bind(&table)
    .fetch_all(&*pg_pool)
    .map_err(|e| e.to_string())?;

    let columns: Vec<ColumnInfo> = rows.iter()
        .map(|row| {
            ColumnInfo {
                name: row.get("column_name"),
                type_: row.get("data_type"),
                nullable: row.get::<String, _>("is_nullable") == "YES",
                default_value: row.try_get("column_default").ok(),
                is_primary_key: row.get("is_primary_key"),
                is_foreign_key: row.get("is_foreign_key"),
            }
        })
        .collect();

    Ok(columns)
}
```

**Step 4: Write commands/mod.rs**

```rust
pub mod connection;
pub mod query;
pub mod schema;

pub use connection::*;
pub use query::*;
pub use schema::*;
```

**Step 5: Update main.rs**

```rust
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod commands;
mod db;
mod error;

use db::ConnectionPool;

fn main() {
    tauri::Builder::default()
        .manage(ConnectionPool::new())
        .invoke_handler(tauri::generate_handler![
            commands::connect,
            commands::disconnect,
            commands::test_connection,
            commands::execute_query,
            commands::get_schemas,
            commands::get_tables,
            commands::get_columns,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
```

**Step 6: Verify compilation**

```bash
cd tablepro/src-tauri
cargo check
```

Expected: No errors

**Step 7: Commit**

```bash
git add .
git commit -m "feat: add Rust commands for connection, query, and schema"
```

---

## Phase 3: React Frontend

### Task 8: Set Up TypeScript Types

**Files:**
- Create: `tablepro/src/types/index.ts`

**Step 1: Write types**

```typescript
export interface ConnectionConfig {
  id: string;
  name: string;
  host: string;
  port: number;
  database: string;
  username: string;
  password?: string;
  sslMode: SslMode;
  sshConfig?: SshConfig;
}

export type SslMode = 'disable' | 'require' | 'verify-full';

export interface SshConfig {
  host: string;
  port: number;
  username: string;
  authType: 'password' | 'key' | 'agent';
  privateKey?: string;
  passphrase?: string;
}

export interface ConnectionInfo {
  id: string;
  name: string;
  database: string;
  serverVersion: string;
  connectedAt: string;
}

export interface QueryResult {
  columns: ColumnMeta[];
  rows: Record<string, unknown>[];
  rowCount: number;
  executionTimeMs: number;
  truncated: boolean;
}

export interface ColumnMeta {
  name: string;
  type: string;
  nullable: boolean;
  primaryKey: boolean;
}

export interface TableInfo {
  name: string;
  schema: string;
  type: 'table' | 'view' | 'materialized-view';
}

export interface ColumnInfo {
  name: string;
  type: string;
  nullable: boolean;
  defaultValue: string | null;
  isPrimaryKey: boolean;
  isForeignKey: boolean;
}

export interface Tab {
  id: string;
  type: 'query' | 'table' | 'structure';
  title: string;
  connectionId: string;
}

export interface QueryTab extends Tab {
  type: 'query';
  sql: string;
  result?: QueryResult;
}
```

**Step 2: Commit**

```bash
git add .
git commit -m "feat: add TypeScript types"
```

---

### Task 9: Create Tauri API Wrapper

**Files:**
- Create: `tablepro/src/lib/tauri.ts`

**Step 1: Write tauri.ts**

```typescript
import { invoke } from '@tauri-apps/api/tauri';
import type { ConnectionConfig, ConnectionInfo, QueryResult, TableInfo, ColumnInfo } from '@/types';

export const tauriApi = {
  connect: (config: ConnectionConfig): Promise<ConnectionInfo> =>
    invoke('connect', { config }),

  disconnect: (connectionId: string): Promise<void> =>
    invoke('disconnect', { connectionId }),

  testConnection: (config: ConnectionConfig): Promise<boolean> =>
    invoke('test_connection', { config }),

  executeQuery: (
    connectionId: string,
    sql: string,
    limit?: number
  ): Promise<QueryResult> =>
    invoke('execute_query', { connectionId, sql, limit }),

  getSchemas: (connectionId: string): Promise<string[]> =>
    invoke('get_schemas', { connectionId }),

  getTables: (connectionId: string, schema: string): Promise<TableInfo[]> =>
    invoke('get_tables', { connectionId, schema }),

  getColumns: (
    connectionId: string,
    schema: string,
    table: string
  ): Promise<ColumnInfo[]> =>
    invoke('get_columns', { connectionId, schema, table }),
};
```

**Step 2: Commit**

```bash
git add .
git commit -m "feat: add Tauri API wrapper"
```

---

### Task 10: Create Zustand Store

**Files:**
- Create: `tablepro/src/stores/appStore.ts`

**Step 1: Write appStore.ts**

```typescript
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { ConnectionConfig, ConnectionInfo, Tab } from '@/types';

interface AppState {
  connections: ConnectionConfig[];
  activeConnectionId: string | null;
  connectionInfos: Map<string, ConnectionInfo>;
  tabs: Tab[];
  activeTabId: string | null;
  sidebarCollapsed: boolean;
  theme: 'light' | 'dark' | 'system';

  addConnection: (config: ConnectionConfig) => void;
  removeConnection: (id: string) => void;
  setActiveConnection: (id: string | null) => void;
  setConnectionInfo: (id: string, info: ConnectionInfo) => void;

  addTab: (tab: Tab) => void;
  closeTab: (id: string) => void;
  setActiveTab: (id: string) => void;
  updateTab: (id: string, updates: Partial<Tab>) => void;

  toggleSidebar: () => void;
  setTheme: (theme: 'light' | 'dark' | 'system') => void;
}

export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      connections: [],
      activeConnectionId: null,
      connectionInfos: new Map(),
      tabs: [],
      activeTabId: null,
      sidebarCollapsed: false,
      theme: 'system',

      addConnection: (config) =>
        set((state) => ({
          connections: [...state.connections, config],
        })),

      removeConnection: (id) =>
        set((state) => ({
          connections: state.connections.filter((c) => c.id !== id),
          connectionInfos: (() => {
            const newMap = new Map(state.connectionInfos);
            newMap.delete(id);
            return newMap;
          })(),
        })),

      setActiveConnection: (id) =>
        set({ activeConnectionId: id }),

      setConnectionInfo: (id, info) =>
        set((state) => {
          const newMap = new Map(state.connectionInfos);
          newMap.set(id, info);
          return { connectionInfos: newMap };
        }),

      addTab: (tab) =>
        set((state) => ({
          tabs: [...state.tabs, tab],
          activeTabId: tab.id,
        })),

      closeTab: (id) =>
        set((state) => ({
          tabs: state.tabs.filter((t) => t.id !== id),
          activeTabId:
            state.activeTabId === id
              ? state.tabs[state.tabs.length - 2]?.id ?? null
              : state.activeTabId,
        })),

      setActiveTab: (id) =>
        set({ activeTabId: id }),

      updateTab: (id, updates) =>
        set((state) => ({
          tabs: state.tabs.map((t) =>
            t.id === id ? { ...t, ...updates } : t
          ),
        })),

      toggleSidebar: () =>
        set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),

      setTheme: (theme) => set({ theme }),
    }),
    {
      name: 'tablepro-storage',
      partialize: (state) => ({
        connections: state.connections,
        sidebarCollapsed: state.sidebarCollapsed,
        theme: state.theme,
      }),
    }
  )
);
```

**Step 2: Commit**

```bash
git add .
git commit -m "feat: add Zustand store"
```

---

### Task 11: Create Custom Hooks

**Files:**
- Create: `tablepro/src/hooks/useConnection.ts`
- Create: `tablepro/src/hooks/useQuery.ts`
- Create: `tablepro/src/hooks/useSchema.ts`

**Step 1: Write useConnection.ts**

```typescript
import { useMutation } from '@tanstack/react-query';
import { tauriApi } from '@/lib/tauri';
import { useAppStore } from '@/stores/appStore';

export function useConnection() {
  const { setConnectionInfo, setActiveConnection } = useAppStore();

  const connectMutation = useMutation({
    mutationFn: tauriApi.connect,
    onSuccess: (info, variables) => {
      setConnectionInfo(variables.id, info);
      setActiveConnection(variables.id);
    },
  });

  const disconnectMutation = useMutation({
    mutationFn: tauriApi.disconnect,
    onSuccess: () => {
      setActiveConnection(null);
    },
  });

  const testMutation = useMutation({
    mutationFn: tauriApi.testConnection,
  });

  return {
    connect: connectMutation.mutateAsync,
    disconnect: disconnectMutation.mutateAsync,
    test: testMutation.mutateAsync,
    isConnecting: connectMutation.isPending,
    isDisconnecting: disconnectMutation.isPending,
    isTesting: testMutation.isPending,
    error: connectMutation.error || disconnectMutation.error,
  };
}
```

**Step 2: Write useQuery.ts**

```typescript
import { useMutation } from '@tanstack/react-query';
import { tauriApi } from '@/lib/tauri';

export function useQuery() {
  const executeMutation = useMutation({
    mutationFn: ({
      connectionId,
      sql,
      limit,
    }: {
      connectionId: string;
      sql: string;
      limit?: number;
    }) => tauriApi.executeQuery(connectionId, sql, limit),
  });

  return {
    execute: executeMutation.mutateAsync,
    isExecuting: executeMutation.isPending,
    result: executeMutation.data,
    error: executeMutation.error,
  };
}
```

**Step 3: Write useSchema.ts**

```typescript
import { useQuery } from '@tanstack/react-query';
import { tauriApi } from '@/lib/tauri';

export function useSchemas(connectionId: string | null) {
  return useQuery({
    queryKey: ['schemas', connectionId],
    queryFn: () => tauriApi.getSchemas(connectionId!),
    enabled: connectionId !== null,
  });
}

export function useTables(connectionId: string | null, schema: string) {
  return useQuery({
    queryKey: ['tables', connectionId, schema],
    queryFn: () => tauriApi.getTables(connectionId!, schema),
    enabled: connectionId !== null && schema !== '',
  });
}

export function useColumns(
  connectionId: string | null,
  schema: string,
  table: string
) {
  return useQuery({
    queryKey: ['columns', connectionId, schema, table],
    queryFn: () => tauriApi.getColumns(connectionId!, schema, table),
    enabled: connectionId !== null && schema !== '' && table !== '',
  });
}
```

**Step 4: Commit**

```bash
git add .
git commit -m "feat: add React hooks for connection, query, and schema"
```

---

### Task 12: Create Main Layout Components

**Files:**
- Create: `tablepro/src/components/layout/MainLayout.tsx`
- Create: `tablepro/src/components/layout/Sidebar.tsx`
- Create: `tablepro/src/components/layout/Toolbar.tsx`

**Step 1: Write MainLayout.tsx**

```typescript
import { Sidebar } from './Sidebar';
import { Toolbar } from './Toolbar';
import { TabBar } from '../tabs/TabBar';
import { QueryTab } from '../tabs/QueryTab';
import { useAppStore } from '@/stores/appStore';

export function MainLayout() {
  const { tabs, activeTabId } = useAppStore();
  const activeTab = tabs.find((t) => t.id === activeTabId);

  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className="flex flex-col flex-1 overflow-hidden">
        <Toolbar />
        <TabBar />
        <main className="flex-1 overflow-hidden">
          {activeTab ? (
            <QueryTab tab={activeTab} />
          ) : (
            <div className="flex items-center justify-center h-full text-muted-foreground">
              No tab open. Press Ctrl+T to create a new query.
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
```

**Step 2: Write Sidebar.tsx**

```typescript
import { useState } from 'react';
import { ChevronRight, ChevronDown, Database, Table, Eye } from 'lucide-react';
import { useSchemas, useTables } from '@/hooks/useSchema';
import { useAppStore } from '@/stores/appStore';

export function Sidebar() {
  const { activeConnectionId, sidebarCollapsed, addTab } = useAppStore();
  const [expandedSchemas, setExpandedSchemas] = useState<Set<string>>(new Set());

  const { data: schemas, isLoading } = useSchemas(activeConnectionId);

  const toggleSchema = (schema: string) => {
    setExpandedSchemas((prev) => {
      const next = new Set(prev);
      if (next.has(schema)) {
        next.delete(schema);
      } else {
        next.add(schema);
      }
      return next;
    });
  };

  const handleTableClick = (schema: string, table: TableInfo) => {
    addTab({
      id: crypto.randomUUID(),
      type: 'query',
      title: table.name,
      connectionId: activeConnectionId!,
    });
  };

  if (sidebarCollapsed) {
    return (
      <div className="w-12 bg-muted border-r flex flex-col items-center py-2">
        <Database className="w-5 h-5 text-muted-foreground" />
      </div>
    );
  }

  return (
    <aside className="w-64 bg-muted/50 border-r flex flex-col">
      <div className="p-2 border-b flex items-center gap-2">
        <Database className="w-4 h-4" />
        <span className="text-sm font-medium">Schema</span>
      </div>
      <div className="flex-1 overflow-auto p-1">
        {isLoading ? (
          <div className="p-2 text-sm text-muted-foreground">Loading...</div>
        ) : (
          schemas?.map((schema) => (
            <SchemaNode
              key={schema}
              schema={schema}
              expanded={expandedSchemas.has(schema)}
              onToggle={() => toggleSchema(schema)}
              onTableClick={handleTableClick}
            />
          ))
        )}
      </div>
    </aside>
  );
}

function SchemaNode({
  schema,
  expanded,
  onToggle,
  onTableClick,
}: {
  schema: string;
  expanded: boolean;
  onToggle: () => void;
  onTableClick: (schema: string, table: TableInfo) => void;
}) {
  const { activeConnectionId } = useAppStore();
  const { data: tables, isLoading } = useTables(activeConnectionId, schema);

  return (
    <div>
      <button
        onClick={onToggle}
        className="w-full flex items-center gap-1 px-2 py-1 hover:bg-accent rounded text-sm"
      >
        {expanded ? (
          <ChevronDown className="w-3 h-3" />
        ) : (
          <ChevronRight className="w-3 h-3" />
        )}
        <span className="font-medium">{schema}</span>
      </button>
      {expanded && (
        <div className="ml-3">
          {isLoading ? (
            <div className="px-2 py-1 text-xs text-muted-foreground">Loading...</div>
          ) : (
            tables?.map((table) => (
              <button
                key={table.name}
                onClick={() => onTableClick(schema, table)}
                className="w-full flex items-center gap-2 px-2 py-0.5 hover:bg-accent rounded text-sm"
              >
                {table.type === 'view' ? (
                  <Eye className="w-3 h-3 text-blue-500" />
                ) : (
                  <Table className="w-3 h-3 text-green-500" />
                )}
                <span>{table.name}</span>
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
}
```

**Step 3: Write Toolbar.tsx**

```typescript
import { Database, Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useAppStore } from '@/stores/appStore';
import { ConnectionDialog } from '@/components/connection/ConnectionDialog';
import { useState } from 'react';

export function Toolbar() {
  const { activeConnectionId, connectionInfos } = useAppStore();
  const [connectionDialogOpen, setConnectionDialogOpen] = useState(false);

  const activeConnection = activeConnectionId
    ? connectionInfos.get(activeConnectionId)
    : null;

  return (
    <div className="h-10 border-b flex items-center px-2 gap-2">
      <Button
        variant="outline"
        size="sm"
        onClick={() => setConnectionDialogOpen(true)}
      >
        <Plus className="w-4 h-4 mr-1" />
        New Connection
      </Button>

      {activeConnection && (
        <div className="flex items-center gap-2 ml-auto">
          <Database className="w-4 h-4 text-green-500" />
          <span className="text-sm">{activeConnection.name}</span>
          <span className="text-xs text-muted-foreground">
            ({activeConnection.database})
          </span>
        </div>
      )}

      <ConnectionDialog
        open={connectionDialogOpen}
        onClose={() => setConnectionDialogOpen(false)}
      />
    </div>
  );
}
```

**Step 4: Commit**

```bash
git add .
git commit -m "feat: add main layout components"
```

---

### Task 13: Create Connection Dialog

**Files:**
- Create: `tablepro/src/components/connection/ConnectionDialog.tsx`

**Step 1: Write ConnectionDialog.tsx**

```typescript
import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useConnection } from '@/hooks/useConnection';
import { useAppStore } from '@/stores/appStore';
import type { ConnectionConfig } from '@/types';

interface ConnectionDialogProps {
  open: boolean;
  onClose: () => void;
}

export function ConnectionDialog({ open, onClose }: ConnectionDialogProps) {
  const [config, setConfig] = useState<Partial<ConnectionConfig>>({
    name: '',
    host: 'localhost',
    port: 5432,
    database: '',
    username: 'postgres',
    sslMode: 'disable',
  });

  const { connect, test, isConnecting, isTesting, error, addConnection } = useConnection();
  const { setActiveConnection } = useAppStore();

  const handleTest = async () => {
    try {
      await test(config as ConnectionConfig);
      alert('Connection successful!');
    } catch (e) {
      alert(`Connection failed: ${(e as Error).message}`);
    }
  };

  const handleSave = async () => {
    try {
      const fullConfig = {
        ...config,
        id: crypto.randomUUID(),
      } as ConnectionConfig;

      addConnection(fullConfig);
      await connect(fullConfig);
      onClose();
    } catch (e) {
      alert(`Failed to connect: ${(e as Error).message}`);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>New Connection</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Connection Name</Label>
            <Input
              id="name"
              value={config.name ?? ''}
              onChange={(e) => setConfig({ ...config, name: e.target.value })}
              placeholder="My Database"
            />
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-2 space-y-2">
              <Label htmlFor="host">Host</Label>
              <Input
                id="host"
                value={config.host ?? ''}
                onChange={(e) => setConfig({ ...config, host: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="port">Port</Label>
              <Input
                id="port"
                type="number"
                value={config.port ?? 5432}
                onChange={(e) =>
                  setConfig({ ...config, port: parseInt(e.target.value) })
                }
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="database">Database</Label>
            <Input
              id="database"
              value={config.database ?? ''}
              onChange={(e) => setConfig({ ...config, database: e.target.value })}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="username">Username</Label>
            <Input
              id="username"
              value={config.username ?? ''}
              onChange={(e) => setConfig({ ...config, username: e.target.value })}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              type="password"
              value={config.password ?? ''}
              onChange={(e) => setConfig({ ...config, password: e.target.value })}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleTest} disabled={isTesting}>
            {isTesting ? 'Testing...' : 'Test Connection'}
          </Button>
          <Button onClick={handleSave} disabled={isConnecting}>
            {isConnecting ? 'Connecting...' : 'Save & Connect'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
```

**Step 2: Commit**

```bash
git add .
git commit -m "feat: add connection dialog"
```

---

### Task 14: Create Query Tab with SQL Editor and Data Grid

**Files:**
- Create: `tablepro/src/components/tabs/TabBar.tsx`
- Create: `tablepro/src/components/tabs/QueryTab.tsx`
- Create: `tablepro/src/components/editor/SqlEditor.tsx`
- Create: `tablepro/src/components/grid/DataGrid.tsx`

**Step 1: Install dependencies**

```bash
cd tablepro
npm install @monaco-editor/react ag-grid-community ag-grid-react
```

**Step 2: Write TabBar.tsx**

```typescript
import { X } from 'lucide-react';
import { useAppStore } from '@/stores/appStore';

export function TabBar() {
  const { tabs, activeTabId, setActiveTab, closeTab } = useAppStore();

  return (
    <div className="flex border-b bg-muted/30 overflow-x-auto">
      {tabs.map((tab) => (
        <div
          key={tab.id}
          className={`flex items-center gap-2 px-3 py-2 text-sm border-r cursor-pointer ${
            activeTabId === tab.id
              ? 'bg-background border-t border-t-primary'
              : 'hover:bg-muted'
          }`}
          onClick={() => setActiveTab(tab.id)}
        >
          <span>{tab.title}</span>
          <button
            onClick={(e) => {
              e.stopPropagation();
              closeTab(tab.id);
            }}
            className="hover:bg-accent rounded p-0.5"
          >
            <X className="w-3 h-3" />
          </button>
        </div>
      ))}
    </div>
  );
}
```

**Step 3: Write SqlEditor.tsx**

```typescript
import Editor from '@monaco-editor/react';
import type { editor } from 'monaco-editor';

interface SqlEditorProps {
  value: string;
  onChange: (value: string) => void;
  onExecute: (sql: string) => void;
}

export function SqlEditor({ value, onChange, onExecute }: SqlEditorProps) {
  const handleEditorMount = (editor: editor.IStandaloneCodeEditor) => {
    editor.addCommand(
      // Ctrl+Enter or Cmd+Enter
      2048 | 3, // KeyMod.CtrlCmd | KeyCode.Enter
      () => {
        onExecute(editor.getValue());
      }
    );
  };

  return (
    <Editor
      height="100%"
      defaultLanguage="sql"
      value={value}
      onChange={(v) => onChange(v ?? '')}
      onMount={handleEditorMount}
      theme="vs-dark"
      options={{
        fontSize: 14,
        minimap: { enabled: false },
        lineNumbers: 'on',
        wordWrap: 'on',
        automaticLayout: true,
        scrollBeyondLastLine: false,
        tabSize: 2,
      }}
    />
  );
}
```

**Step 4: Write DataGrid.tsx**

```typescript
import { useMemo } from 'react';
import { AgGridReact } from 'ag-grid-react';
import { AllCommunityModule, ModuleRegistry } from 'ag-grid-community';
import type { ColDef } from 'ag-grid-community';
import 'ag-grid-community/styles/ag-grid.css';
import 'ag-grid-community/styles/ag-theme-alpine.css';

ModuleRegistry.registerModules([AllCommunityModule]);

interface DataGridProps {
  columns: { name: string; type: string }[];
  rows: Record<string, unknown>[];
}

export function DataGrid({ columns, rows }: DataGridProps) {
  const columnDefs = useMemo<ColDef[]>(
    () =>
      columns.map((col) => ({
        field: col.name,
        headerName: col.name,
        sortable: true,
        filter: true,
        resizable: true,
        cellRenderer: (params: { value: unknown }) => {
          if (params.value === null) {
            return <span className="text-muted-foreground italic">NULL</span>;
          }
          if (typeof params.value === 'object') {
            return JSON.stringify(params.value);
          }
          return String(params.value);
        },
      })),
    [columns]
  );

  const defaultColDef = useMemo<ColDef>(
    () => ({
      minWidth: 100,
      flex: 1,
    }),
    []
  );

  return (
    <div className="ag-theme-alpine h-full w-full">
      <AgGridReact
        columnDefs={columnDefs}
        rowData={rows}
        defaultColDef={defaultColDef}
        animateRows={true}
        enableCellChangeFlash={true}
      />
    </div>
  );
}
```

**Step 5: Write QueryTab.tsx**

```typescript
import { useState } from 'react';
import { Play, Clock, Rows } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { SqlEditor } from '@/components/editor/SqlEditor';
import { DataGrid } from '@/components/grid/DataGrid';
import { useQuery } from '@/hooks/useQuery';
import { useAppStore } from '@/stores/appStore';
import type { QueryTab, QueryResult } from '@/types';

interface QueryTabProps {
  tab: QueryTab;
}

export function QueryTab({ tab }: QueryTabProps) {
  const [sql, setSql] = useState(tab.sql ?? 'SELECT 1;');
  const [result, setResult] = useState<QueryResult | undefined>(tab.result);
  const { activeConnectionId, updateTab } = useAppStore();
  const { execute, isExecuting } = useQuery();

  const handleExecute = async (query: string) => {
    if (!activeConnectionId) return;

    try {
      const res = await execute({
        connectionId: activeConnectionId,
        sql: query,
        limit: 1000,
      });
      setResult(res);

      updateTab(tab.id, {
        sql: query,
        result: res,
      });
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 p-2 border-b bg-muted/50">
        <Button
          size="sm"
          onClick={() => handleExecute(sql)}
          disabled={isExecuting || !activeConnectionId}
        >
          <Play className="w-4 h-4 mr-1" />
          {isExecuting ? 'Executing...' : 'Execute'}
        </Button>
        <span className="text-xs text-muted-foreground">
          Ctrl+Enter to execute
        </span>
      </div>

      <div className="flex-1 flex flex-col overflow-hidden">
        <div className="h-48 border-b">
          <SqlEditor
            value={sql}
            onChange={setSql}
            onExecute={handleExecute}
          />
        </div>

        <div className="flex-1 flex flex-col">
          {result && (
            <div className="flex items-center gap-4 px-3 py-1 border-b text-xs text-muted-foreground">
              <span className="flex items-center gap-1">
                <Rows className="w-3 h-3" />
                {result.rowCount} rows
              </span>
              <span className="flex items-center gap-1">
                <Clock className="w-3 h-3" />
                {result.executionTimeMs}ms
              </span>
            </div>
          )}

          <div className="flex-1">
            {result ? (
              <DataGrid columns={result.columns} rows={result.rows} />
            ) : (
              <div className="flex items-center justify-center h-full text-muted-foreground">
                Execute a query to see results
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
```

**Step 6: Commit**

```bash
git add .
git commit -m "feat: add query tab with SQL editor and data grid"
```

---

### Task 15: Update App.tsx

**Files:**
- Modify: `tablepro/src/App.tsx`

**Step 1: Write App.tsx**

```typescript
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MainLayout } from './components/layout/MainLayout';

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <MainLayout />
    </QueryClientProvider>
  );
}

export default App;
```

**Step 2: Commit**

```bash
git add .
git commit -m "feat: update App.tsx with providers"
```

---

### Task 16: Verify Build

**Files:**
- Test: Verify project builds successfully

**Step 1: Build the project**

```bash
cd tablepro
npm run build
```

Expected: Build succeeds with no errors

**Step 2: Test Tauri dev**

```bash
npm run tauri dev
```

Expected: App opens with sidebar, toolbar, and query tab

**Step 3: Commit**

```bash
git add .
git commit -m "feat: verify build and dev server"
```

---

## Summary

After completing all tasks, you will have:

1. **Working Tauri app** with React frontend
2. **PostgreSQL connection** via Rust sqlx
3. **Schema browser** in sidebar
4. **SQL Editor** with Monaco
5. **Data Grid** with AG-Grid
6. **Query execution** with results display
7. **Connection management** dialog

**Next phases would add:**
- SQLite/MySQL drivers
- Data editing
- Export/Import
- SSH tunneling
- Query history
- AI integration
