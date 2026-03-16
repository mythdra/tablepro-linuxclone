use tauri::State;
use crate::db::{ConnectionPool, TableInfo, ColumnInfo};
use sqlx::{Row, postgres::PgPool};

#[tauri::command]
pub async fn get_schemas(
    connection_id: String,
    pool: State<'_, ConnectionPool>,
) -> Result<Vec<String>, String> {
    let pg_pool: PgPool = pool.get(&connection_id)
        .await
        .map_err(|e| e.to_string())?;

    let schemas: Vec<String> = sqlx::query_scalar(
        "SELECT schema_name FROM information_schema.schemata
         WHERE schema_name NOT IN ('pg_catalog', 'information_schema')
         ORDER BY schema_name"
    )
    .fetch_all(&pg_pool)
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
    let pg_pool: PgPool = pool.get(&connection_id)
        .await
        .map_err(|e| e.to_string())?;

    let rows: Vec<sqlx::postgres::PgRow> = sqlx::query(
        "SELECT table_name, table_schema, table_type
         FROM information_schema.tables
         WHERE table_schema = $1
         ORDER BY table_name"
    )
    .bind(&schema)
    .fetch_all(&pg_pool)
    .await
    .map_err(|e| e.to_string())?;

    let tables: Vec<TableInfo> = rows.iter()
        .map(|row: &sqlx::postgres::PgRow| {
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
    let pg_pool: PgPool = pool.get(&connection_id)
        .await
        .map_err(|e| e.to_string())?;

    let rows: Vec<sqlx::postgres::PgRow> = sqlx::query(
        "SELECT
            c.column_name,
            c.data_type,
            c.is_nullable,
            c.column_default,
            CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary_key,
            CASE WHEN fk.column_name IS NOT NULL THEN true ELSE false END as is_foreign_key
        FROM information_schema.columns c
        LEFT JOIN (
            SELECT ku.column_name, ku.table_name
            FROM information_schema.table_constraints tc
            JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
            WHERE tc.constraint_type = 'PRIMARY KEY'
        ) pk ON c.column_name = pk.column_name AND c.table_name = pk.table_name
        LEFT JOIN (
            SELECT ku.column_name, ku.table_name
            FROM information_schema.table_constraints tc
            JOIN information_schema.key_column_usage ku ON tc.constraint_name = ku.constraint_name
            WHERE tc.constraint_type = 'FOREIGN KEY'
        ) fk ON c.column_name = fk.column_name AND c.table_name = fk.table_name
        WHERE c.table_schema = $1 AND c.table_name = $2
        ORDER BY c.ordinal_position"
    )
    .bind(&schema)
    .bind(&table)
    .fetch_all(&pg_pool)
    .await
    .map_err(|e| e.to_string())?;

    let columns: Vec<ColumnInfo> = rows.iter()
        .map(|row: &sqlx::postgres::PgRow| {
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
