import { v4 as uuidv4 } from 'uuid';
import type { DatabaseConnection, ConnectionStatus } from '../types';

const CONNECTIONS_KEY = 'tablepro_connections';

function generateId(): string {
  return uuidv4();
}

export async function loadConnections(): Promise<DatabaseConnection[]> {
  try {
    const stored = localStorage.getItem(CONNECTIONS_KEY);
    return stored ? JSON.parse(stored) : [];
  } catch {
    return [];
  }
}

export async function saveConnection(connection: DatabaseConnection): Promise<DatabaseConnection> {
  const connections = await loadConnections();
  const existingIndex = connections.findIndex((c) => c.id === connection.id);

  const toSave: DatabaseConnection = connection.id
    ? connection
    : { ...connection, id: generateId() };

  if (existingIndex >= 0) {
    connections[existingIndex] = toSave;
  } else {
    connections.push(toSave);
  }

  localStorage.setItem(CONNECTIONS_KEY, JSON.stringify(connections));
  return toSave;
}

export async function deleteConnection(id: string): Promise<void> {
  const connections = await loadConnections();
  const filtered = connections.filter((c) => c.id !== id);
  localStorage.setItem(CONNECTIONS_KEY, JSON.stringify(filtered));
}

export async function testConnection(_connection: DatabaseConnection): Promise<boolean> {
  await new Promise((resolve) => setTimeout(resolve, 1500));
  const success = Math.random() > 0.2;
  return success;
}

export async function connect(_id: string): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, 1000));
}

export async function disconnect(_id: string): Promise<void> {
  await new Promise((resolve) => setTimeout(resolve, 300));
}

export type { DatabaseConnection, ConnectionStatus };