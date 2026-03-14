// Database connection types
export type DatabaseType =
  | 'postgres'
  | 'mysql'
  | 'sqlite'
  | 'duckdb'
  | 'mssql'
  | 'clickhouse'
  | 'mongodb'
  | 'redis';

export type ConnectionStatus =
  | 'disconnected'
  | 'connecting'
  | 'connected'
  | 'error';

export type SafeModeLevel = 'off' | 'safe' | 'very_safe';

export interface SSHTunnelConfig {
  enabled: boolean;
  host: string;
  port: number;
  username: string;
  authMethod: 'password' | 'key' | 'agent';
}

export interface SSLConfig {
  enabled: boolean;
  mode: 'disable' | 'require' | 'verify-ca' | 'verify-full';
  caCert: string;
  clientCert: string;
}

export interface DatabaseConnection {
  id: string;
  name: string;
  type: DatabaseType;
  group: string;
  colorTag: string;
  host: string;
  port: number;
  database: string;
  username: string;
  localFilePath: string;
  ssh: SSHTunnelConfig;
  ssl: SSLConfig;
  safeMode: SafeModeLevel;
  startupCommand: string;
  preConnectScript: string;
}

export interface ConnectionSession {
  connectionId: string;
  status: ConnectionStatus;
  activeDb: string;
  lastPingAt: string;
}

// Query execution types
export interface QueryResult {
  queryId: string;
  connectionId: string;
  columns: ColumnInfo[];
  rows: any[];
  rowCount: number;
  duration: number;
  status: 'success' | 'error' | 'cancelled';
  error?: string;
}

export interface ColumnInfo {
  name: string;
  type: string;
  nullable?: boolean;
}

export interface ActiveQuery {
  queryId: string;
  connectionId: string;
  query: string;
  status: 'executing' | 'cancelling';
  startTime: string;
}

export interface QueryHistoryEntry {
  id: string;
  connectionId: string;
  query: string;
  executedAt: string;
  duration: number;
  rowCount: number;
  status: 'success' | 'error' | 'cancelled';
  error?: string;
}

// Query Editor types
export interface QueryTab {
  id: string;
  name: string;
  content: string;
  isDirty: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface SchemaMetadata {
  tables: TableMetadata[];
  views: ViewMetadata[];
}

export interface TableMetadata {
  name: string;
  schema: string;
  columns: ColumnMetadata[];
}

export interface ViewMetadata {
  name: string;
  schema: string;
  columns: ColumnMetadata[];
}

export interface ColumnMetadata {
  name: string;
  type: string;
  nullable: boolean;
  defaultValue?: string;
  isPrimaryKey?: boolean;
  isForeignKey?: boolean;
}

export interface PaginationContext {
  page: number;
  pageSize: number;
  totalCount: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
  offset: number;
  isExact: boolean;
}

export interface PaginatedQueryResult {
  queryId: string;
  connectionId: string;
  columns: ColumnInfo[];
  rows: unknown[][];
  rowCount: number;
  duration: number;
  status: 'success' | 'error' | 'cancelled';
  error?: string;
  pagination: PaginationContext;
}

export interface CountResult {
  count: number;
  isExact: boolean;
}
