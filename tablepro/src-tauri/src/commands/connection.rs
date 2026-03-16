use tauri::State;
use crate::db::{ConnectionPool, ConnectionConfig, ConnectionInfo};
use sqlx::postgres::PgPool;

#[tauri::command]
pub async fn connect(
    config: ConnectionConfig,
    pool: State<'_, ConnectionPool>,
) -> Result<ConnectionInfo, String> {
    pool.connect(&config)
        .await
        .map_err(|e| e.to_string())?;

    let pg_pool: PgPool = pool.get(&config.id)
        .await
        .map_err(|e| e.to_string())?;

    let version: String = sqlx::query_scalar("SELECT version()")
        .fetch_one(&pg_pool)
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
    let options = sqlx::postgres::PgConnectOptions::new()
        .host(&config.host)
        .port(config.port)
        .database(&config.database)
        .username(&config.username)
        .password(config.password.as_deref().unwrap_or(""));

    let pool = sqlx::postgres::PgPoolOptions::new()
        .max_connections(1)
        .connect_with(options)
        .await
        .map_err(|e| e.to_string())?;

    sqlx::query("SELECT 1")
        .fetch_one(&pool)
        .await
        .map_err(|e| e.to_string())?;

    pool.close().await;
    Ok(true)
}
