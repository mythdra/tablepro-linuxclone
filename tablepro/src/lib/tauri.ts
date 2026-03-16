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
