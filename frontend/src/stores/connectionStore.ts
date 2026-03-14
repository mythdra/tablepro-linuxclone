import { create } from 'zustand';
import type { DatabaseConnection, ConnectionStatus } from '../types';
import * as connectionManager from '../lib/connectionManager';

interface SessionInfo {
  status: ConnectionStatus;
  activeDb: string;
  lastPingAt: string;
}

interface ConnectionState {
  connections: DatabaseConnection[];
  sessions: Map<string, SessionInfo>;
  isLoading: boolean;
  error: string | null;

  loadConnections: () => Promise<void>;
  saveConnection: (connection: DatabaseConnection) => Promise<DatabaseConnection>;
  deleteConnection: (id: string) => Promise<void>;
  duplicateConnection: (id: string) => Promise<DatabaseConnection | null>;
  testConnection: (connection: DatabaseConnection) => Promise<boolean>;
  connect: (id: string) => Promise<void>;
  disconnect: (id: string) => Promise<void>;
  setConnectionStatus: (id: string, status: ConnectionStatus) => void;
}

export const useConnectionStore = create<ConnectionState>((set, get) => ({
  connections: [],
  sessions: new Map(),
  isLoading: false,
  error: null,

  loadConnections: async () => {
    set({ isLoading: true, error: null });
    try {
      const connections = await connectionManager.loadConnections();
      set({ connections, isLoading: false });
    } catch (error) {
      set({ error: (error as Error).message, isLoading: false });
    }
  },

  saveConnection: async (connection) => {
    set({ isLoading: true, error: null });
    try {
      const saved = await connectionManager.saveConnection(connection);
      const connections = get().connections;
      const existingIndex = connections.findIndex((c) => c.id === saved.id);

      if (existingIndex >= 0) {
        const updated = [...connections];
        updated[existingIndex] = saved;
        set({ connections: updated, isLoading: false });
      } else {
        set({ connections: [...connections, saved], isLoading: false });
      }

      return saved;
    } catch (error) {
      set({ error: (error as Error).message, isLoading: false });
      throw error;
    }
  },

  deleteConnection: async (id) => {
    set({ isLoading: true, error: null });
    try {
      await connectionManager.deleteConnection(id);
      const connections = get().connections.filter((c) => c.id !== id);
      const sessions = new Map(get().sessions);
      sessions.delete(id);
      set({ connections, sessions, isLoading: false });
    } catch (error) {
      set({ error: (error as Error).message, isLoading: false });
    }
  },

  duplicateConnection: async (id) => {
    const connection = get().connections.find((c) => c.id === id);
    if (!connection) return null;

    const duplicated: DatabaseConnection = {
      ...connection,
      id: '',
      name: `${connection.name} (copy)`,
    };

    return get().saveConnection(duplicated);
  },

  testConnection: async (connection) => {
    return connectionManager.testConnection(connection);
  },

  connect: async (id) => {
    const sessions = new Map(get().sessions);
    sessions.set(id, { status: 'connecting', activeDb: '', lastPingAt: '' });
    set({ sessions });

    try {
      await connectionManager.connect(id);
      sessions.set(id, {
        status: 'connected',
        activeDb: '',
        lastPingAt: new Date().toISOString(),
      });
      set({ sessions });
    } catch (error) {
      sessions.set(id, { status: 'error', activeDb: '', lastPingAt: '' });
      set({ sessions, error: (error as Error).message });
    }
  },

  disconnect: async (id) => {
    try {
      await connectionManager.disconnect(id);
      const sessions = new Map(get().sessions);
      sessions.set(id, { status: 'disconnected', activeDb: '', lastPingAt: '' });
      set({ sessions });
    } catch (error) {
      set({ error: (error as Error).message });
    }
  },

  setConnectionStatus: (id, status) => {
    const sessions = new Map(get().sessions);
    const existing = sessions.get(id);
    sessions.set(id, {
      status,
      activeDb: existing?.activeDb || '',
      lastPingAt: existing?.lastPingAt || '',
    });
    set({ sessions });
  },
}));