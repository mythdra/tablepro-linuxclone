pub mod commands;
pub mod db;
pub mod error;

pub use db::ConnectionPool;
pub use error::AppError;

// Learn more about Tauri commands at https://tauri.app/develop/calling-rust/
#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}! You've been greeted from Rust!", name)
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .manage(ConnectionPool::new())
        .plugin(tauri_plugin_shell::init())
        .invoke_handler(tauri::generate_handler![
            greet,
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
