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
