import { create } from 'zustand';
import type { QueryResult, QueryHistoryEntry, ActiveQuery } from '../types';

/**
 * Query store interface for managing query execution state.
 * 
 * State:
 * - activeQueries: Map of currently executing queries by query ID
 * - history: List of executed query history entries
 * 
 * Actions:
 * - executeQuery: Execute a query and track its lifecycle
 * - cancelQuery: Cancel a running query
 * - addToHistory: Add completed query to history
 * - getHistory: Get filtered history for a connection
 * - clearHistory: Clear history for a connection or all
 * - clearActiveQueries: Clear all active query state
 */
interface QueryStore {
  // State
  activeQueries: Map<string, ActiveQuery>;
  history: QueryHistoryEntry[];

  // Actions
  executeQuery: (connId: string, query: string) => Promise<QueryResult>;
  cancelQuery: (queryId: string) => Promise<void>;
  addToHistory: (entry: QueryHistoryEntry) => void;
  getHistory: (connId: string) => QueryHistoryEntry[];
  clearHistory: (connId?: string) => void;
  clearActiveQueries: () => void;
}

/** Maximum number of history entries to retain */
const HISTORY_LIMIT = 50;

/**
 * Query store for managing query execution and history.
 * 
 * @example
 * ```typescript
 * // Execute a query
 * const result = await useQueryStore.getState().executeQuery(connId, 'SELECT * FROM users');
 * 
 * // Get history for connection
 * const history = useQueryStore.getState().getHistory(connId);
 * 
 * // Subscribe to store changes
 * useQueryStore((state) => state.activeQueries);
 * ```
 */
export const useQueryStore = create<QueryStore>((set, get) => ({
  activeQueries: new Map(),
  history: [],

  executeQuery: async (connId: string, query: string) => {
    const queryId = `query-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
    
    // Add to active queries
    const activeQuery: ActiveQuery = {
      queryId,
      connectionId: connId,
      query,
      status: 'executing',
      startTime: new Date().toISOString(),
    };
    
    const activeQueries = new Map(get().activeQueries);
    activeQueries.set(queryId, activeQuery);
    set({ activeQueries });

    try {
      // Call Wails binding (mock for now)
      const result = await mockExecuteQuery(connId, query);
      
      // Remove from active queries
      activeQueries.delete(queryId);
      set({ activeQueries });
      
      return result;
    } catch (error) {
      // Remove from active queries on error
      activeQueries.delete(queryId);
      set({ activeQueries });
      throw error;
    }
  },

  cancelQuery: async (queryId: string) => {
    const activeQueries = new Map(get().activeQueries);
    const query = activeQueries.get(queryId);
    
    if (query) {
      query.status = 'cancelling';
      activeQueries.set(queryId, query);
      set({ activeQueries });
      
      // Call Wails cancel (mock for now)
      await mockCancelQuery(queryId);
      
      // Remove from active queries
      activeQueries.delete(queryId);
      set({ activeQueries });
    }
  },

  addToHistory: (entry: QueryHistoryEntry) => {
    const history = [...get().history, entry];
    
    // Limit to last 50 entries
    if (history.length > HISTORY_LIMIT) {
      set({ history: history.slice(history.length - HISTORY_LIMIT) });
    } else {
      set({ history });
    }
  },

  getHistory: (connId: string) => {
    return get().history.filter((entry) => entry.connectionId === connId);
  },

  clearHistory: (connId?: string) => {
    if (connId) {
      set({ history: get().history.filter((entry) => entry.connectionId !== connId) });
    } else {
      set({ history: [] });
    }
  },

  clearActiveQueries: () => {
    set({ activeQueries: new Map() });
  },
}));

// Mock Wails bindings - replace with actual imports when available
async function mockExecuteQuery(_connId: string, _query: string): Promise<QueryResult> {
  // Placeholder - will be replaced with actual Wails binding
  return {
    queryId: `query-${Date.now()}`,
    connectionId: _connId,
    columns: [],
    rows: [],
    rowCount: 0,
    duration: 0,
    status: 'success',
  };
}

async function mockCancelQuery(_queryId: string): Promise<void> {
  // Placeholder - will be replaced with actual Wails binding
}
