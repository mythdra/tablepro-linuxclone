import { v4 as uuidv4 } from 'uuid';
import type { DatabaseConnection, ConnectionStatus } from '../types';
import type { connection } from '../wailsjs/go/models';

function generateId(): string {
  return uuidv4();
}

// Get the Wails Go bindings from window
function getGoBindings() {
  return (window as any).go?.main?.App;
}

// Convert frontend DatabaseConnection to backend model
function toBackendConnection(conn: DatabaseConnection): connection.DatabaseConnection {
  return {
    id: conn.id,
    name: conn.name,
    type: conn.type,
    group: conn.group || '',
    colorTag: conn.colorTag || '',
    host: conn.host || '',
    port: conn.port || 0,
    database: conn.database || '',
    username: conn.username || '',
    localFilePath: conn.localFilePath || '',
    ssh: {
      enabled: conn.ssh?.enabled || false,
      host: conn.ssh?.host || '',
      port: conn.ssh?.port || 22,
      username: conn.ssh?.username || '',
      authMethod: conn.ssh?.authMethod || 'password',
    },
    ssl: {
      enabled: conn.ssl?.enabled || false,
      mode: conn.ssl?.mode || 'disable',
      caCert: conn.ssl?.caCert || '',
      clientCert: conn.ssl?.clientCert || '',
      serverName: '',
    },
    safeMode: conn.safeMode || 'safe',
    startupCommand: conn.startupCommand || '',
    preConnectScript: conn.preConnectScript || '',
  };
}

// Convert backend model to frontend DatabaseConnection
function fromBackendConnection(conn: connection.DatabaseConnection): DatabaseConnection {
  return {
    id: conn.id,
    name: conn.name,
    type: conn.type as DatabaseConnection['type'],
    group: conn.group,
    colorTag: conn.colorTag,
    host: conn.host,
    port: conn.port,
    database: conn.database,
    username: conn.username,
    localFilePath: conn.localFilePath,
    ssh: {
      enabled: conn.ssh?.enabled || false,
      host: conn.ssh?.host || '',
      port: conn.ssh?.port || 22,
      username: conn.ssh?.username || '',
      authMethod: (conn.ssh?.authMethod as 'password' | 'key' | 'agent') || 'password',
    },
    ssl: {
      enabled: conn.ssl?.enabled || false,
      mode: (conn.ssl?.mode as 'disable' | 'require' | 'verify-ca' | 'verify-full') || 'disable',
      caCert: conn.ssl?.caCert || '',
      clientCert: conn.ssl?.clientCert || '',
    },
    safeMode: (conn.safeMode as 'off' | 'safe' | 'very_safe') || 'safe',
    startupCommand: conn.startupCommand || '',
    preConnectScript: conn.preConnectScript || '',
  };
}

export async function loadConnections(): Promise<DatabaseConnection[]> {
  try {
    const go = getGoBindings();
    if (!go?.ListConnections) {
      console.warn('Wails bindings not available, returning empty array');
      return [];
    }
    const result = await go.ListConnections();
    return (result || []).map(fromBackendConnection);
  } catch (error) {
    console.error('Failed to load connections:', error);
    return [];
  }
}

export async function saveConnection(connection: DatabaseConnection): Promise<DatabaseConnection> {
  const toSave: DatabaseConnection = connection.id
    ? connection
    : { ...connection, id: generateId() };

  const go = getGoBindings();
  if (!go?.SaveConnection) {
    throw new Error('Wails bindings not available');
  }
  await go.SaveConnection('', toBackendConnection(toSave));
  return toSave;
}

export async function deleteConnection(id: string): Promise<void> {
  const go = getGoBindings();
  if (!go?.DeleteConnection) {
    throw new Error('Wails bindings not available');
  }
  await go.DeleteConnection('', id);
}

export async function testConnection(connection: DatabaseConnection): Promise<boolean> {
  try {
    const go = getGoBindings();
    if (!go?.TestConnection) {
      console.warn('Wails bindings not available, returning false');
      return false;
    }
    const result = await go.TestConnection('', toBackendConnection(connection));
    return result.success;
  } catch (error) {
    console.error('Connection test failed:', error);
    return false;
  }
}

export async function connect(_id: string): Promise<void> {
  // TODO: Implement connect via Wails
  await new Promise((resolve) => setTimeout(resolve, 1000));
}

export async function disconnect(_id: string): Promise<void> {
  // TODO: Implement disconnect via Wails
  await new Promise((resolve) => setTimeout(resolve, 300));
}

export type { DatabaseConnection, ConnectionStatus };