use tauri::State;
use crate::db::{ConnectionPool, QueryResult, ColumnMeta};
use sqlx::{Column, Row, postgres::PgPool};

#[tauri::command]
pub async fn execute_query(
    connection_id: String,
    sql: String,
    limit: Option<usize>,
    pool: State<'_, ConnectionPool>,
) -> Result<QueryResult, String> {
    let pg_pool: PgPool = pool.get(&connection_id)
        .await
        .map_err(|e| e.to_string())?;

    let start = std::time::Instant::now();

    let query = if let Some(limit) = limit {
        format!("{} LIMIT {}", sql.trim_end_matches(';'), limit)
    } else {
        sql
    };

    let rows: Vec<sqlx::postgres::PgRow> = sqlx::query(&query)
        .fetch_all(&pg_pool)
        .await
        .map_err(|e| e.to_string())?;

    let columns: Vec<ColumnMeta> = if !rows.is_empty() {
        rows[0].columns().iter()
            .map(|col: &sqlx::postgres::PgColumn| ColumnMeta {
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
        .map(|row: &sqlx::postgres::PgRow| {
            let mut map = serde_json::Map::new();
            for (i, col) in row.columns().iter().enumerate() {
                let value: serde_json::Value = match row.try_get::<Option<String>, _>(i) {
                    Ok(Some(v)) => serde_json::Value::String(v),
                    Ok(None) => serde_json::Value::Null,
                    Err(_) => serde_json::Value::Null,
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
        truncated: limit.is_some() && rows.len() >= limit.unwrap(),
    })
}
