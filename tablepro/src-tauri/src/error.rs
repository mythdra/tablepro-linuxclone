use serde::Serialize;
use thiserror::Error;

#[derive(Debug, Error, Serialize)]
#[serde(tag = "code", content = "message")]
pub enum AppError {
    #[error("Database error: {0}")]
    Database(String),

    #[error("Connection not found: {0}")]
    NotConnected(String),

    #[error("Invalid configuration: {0}")]
    InvalidConfig(String),

    #[error("Query execution failed: {0}")]
    QueryFailed(String),
}
