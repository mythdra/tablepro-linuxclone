use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use tokio::time::{timeout, Duration};
use sqlx::postgres::{PgPool, PgPoolOptions, PgConnectOptions};
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
        let options = PgConnectOptions::new()
            .host(&config.host)
            .port(config.port)
            .database(&config.database)
            .username(&config.username)
            .password(config.password.as_deref().unwrap_or(""));

        let pool = PgPoolOptions::new()
            .max_connections(5)
            .connect_with(options)
            .await
            .map_err(|e| AppError::Database(e.to_string()))?;

        let mut pools = self.pools.write().await;
        pools.insert(config.id.clone(), pool);
        Ok(())
    }

    pub async fn disconnect(&self, connection_id: &str) -> Result<(), AppError> {
        let mut pools = self.pools.write().await;
        if let Some(pool) = pools.remove(connection_id) {
            let _ = timeout(Duration::from_secs(5), pool.close()).await;
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
        if let Some(pool) = pools.get(connection_id) {
            !pool.is_closed()
        } else {
            false
        }
    }
}

impl Default for ConnectionPool {
    fn default() -> Self {
        Self::new()
    }
}
