import { invoke } from '@tauri-apps/api/core';
import type { ConnectionConfig, ConnectionInfo, QueryResult, TableInfo, ColumnInfo } from '@/types';

export const tauriApi = {
  connect: (
    connectionId: string,
    dbType: string,
    host: string,
    port: number,
    database: string,
    username: string,
    password: string
  ): Promise<ConnectionInfo> =>
    invoke('connect', { connectionId, dbType, host, port, database, username, password }),

  disconnect: (connectionId: string): Promise<void> =>
    invoke('disconnect', { connectionId }),

  testConnection: (
    connectionId: string,
    dbType: string,
    host: string,
    port: number,
    database: string,
    username: string,
    password: string
  ): Promise<boolean> =>
    invoke('test_connection', { connectionId, dbType, host, port, database, username, password }),

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
