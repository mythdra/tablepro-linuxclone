import { z } from 'zod';

// Database type enum
export const databaseTypes = [
  'postgres',
  'mysql',
  'sqlite',
  'duckdb',
  'mssql',
  'clickhouse',
  'mongodb',
  'redis',
] as const;

// SSH auth methods
export const sshAuthMethods = ['password', 'key', 'agent'] as const;

// SSL modes
export const sslModes = ['disable', 'require', 'verify-ca', 'verify-full'] as const;

// Safe mode levels
export const safeModeLevels = ['off', 'safe', 'very_safe'] as const;

// Default ports for each database type
export const defaultPorts: Record<string, number> = {
  postgres: 5432,
  mysql: 3306,
  sqlite: 0,
  duckdb: 0,
  mssql: 1433,
  clickhouse: 9000,
  mongodb: 27017,
  redis: 6379,
};

// SSH Tunnel Schema
export const sshTunnelSchema = z.object({
  enabled: z.boolean().default(false),
  host: z.string().min(1, 'Host is required').default(''),
  port: z.number().min(1).max(65535).default(22),
  username: z.string().min(1, 'Username is required').default(''),
  authMethod: z.enum(sshAuthMethods).default('password'),
});

// SSL Config Schema
export const sslConfigSchema = z.object({
  enabled: z.boolean().default(false),
  mode: z.enum(sslModes).default('disable'),
  caCert: z.string().default(''),
  clientCert: z.string().default(''),
});

// Main Connection Schema
export const connectionFormSchema = z.object({
  // Basic info
  id: z.string().optional(),
  name: z.string().min(1, 'Connection name is required').max(100),
  type: z.enum(databaseTypes),
  group: z.string().default(''),
  colorTag: z.string().default(''),

  // Connection details
  host: z.string().min(1, 'Host is required').default('localhost'),
  port: z.number().min(1).max(65535),
  database: z.string().min(1, 'Database name is required'),
  username: z.string().min(1, 'Username is required').default(''),
  localFilePath: z.string().default(''),

  // SSH and SSL
  ssh: sshTunnelSchema,
  ssl: sslConfigSchema,

  // Advanced
  safeMode: z.enum(safeModeLevels).default('safe'),
  startupCommand: z.string().default(''),
  preConnectScript: z.string().default(''),
});

export type ConnectionFormData = z.infer<typeof connectionFormSchema>;

// Form default values
export const getDefaultFormValues = (type: string = 'postgres'): ConnectionFormData => ({
  name: '',
  type: type as any,
  group: '',
  colorTag: '',
  host: 'localhost',
  port: defaultPorts[type] || 5432,
  database: '',
  username: '',
  localFilePath: '',
  ssh: {
    enabled: false,
    host: '',
    port: 22,
    username: '',
    authMethod: 'password',
  },
  ssl: {
    enabled: false,
    mode: 'disable',
    caCert: '',
    clientCert: '',
  },
  safeMode: 'safe',
  startupCommand: '',
  preConnectScript: '',
});