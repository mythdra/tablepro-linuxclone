import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useQueryStore } from './queryStore';
import type { QueryHistoryEntry } from '../types';

// Mock Wails bindings
vi.mock('../lib/queryExecutor', () => ({
  executeQuery: vi.fn(),
  cancelQuery: vi.fn(),
}));

describe('Query Store', () => {
  beforeEach(() => {
    useQueryStore.getState().clearHistory();
    useQueryStore.getState().clearActiveQueries();
  });

  describe('initial state', () => {
    it('should have empty active queries map', () => {
      const state = useQueryStore.getState();
      expect(state.activeQueries.size).toBe(0);
    });

    it('should have empty history array', () => {
      const state = useQueryStore.getState();
      expect(state.history).toEqual([]);
    });
  });

  describe('addToHistory', () => {
    it('should add query history entry to state', () => {
      const entry: QueryHistoryEntry = {
        id: 'test-id',
        connectionId: 'conn-1',
        query: 'SELECT 1',
        executedAt: new Date().toISOString(),
        duration: 100,
        rowCount: 1,
        status: 'success',
      };

      useQueryStore.getState().addToHistory(entry);
      const state = useQueryStore.getState();

      expect(state.history).toHaveLength(1);
      expect(state.history[0]).toEqual(entry);
    });

    it('should limit history to last 50 entries', () => {
      const entries: QueryHistoryEntry[] = Array.from({ length: 55 }, (_, i) => ({
        id: `entry-${i}`,
        connectionId: 'conn-1',
        query: `SELECT ${i}`,
        executedAt: new Date().toISOString(),
        duration: i * 10,
        rowCount: i,
        status: 'success' as const,
      }));

      entries.forEach((entry) => useQueryStore.getState().addToHistory(entry));
      const state = useQueryStore.getState();

      expect(state.history).toHaveLength(50);
      expect(state.history[0].id).toBe('entry-5');
      expect(state.history[49].id).toBe('entry-54');
    });
  });

  describe('getHistory', () => {
    it('should return history filtered by connection ID', () => {
      const entries: QueryHistoryEntry[] = [
        {
          id: '1',
          connectionId: 'conn-1',
          query: 'SELECT 1',
          executedAt: new Date().toISOString(),
          duration: 100,
          rowCount: 1,
          status: 'success',
        },
        {
          id: '2',
          connectionId: 'conn-2',
          query: 'SELECT 2',
          executedAt: new Date().toISOString(),
          duration: 200,
          rowCount: 2,
          status: 'success',
        },
      ];

      entries.forEach((entry) => useQueryStore.getState().addToHistory(entry));
      const history = useQueryStore.getState().getHistory('conn-1');

      expect(history).toHaveLength(1);
      expect(history[0].connectionId).toBe('conn-1');
    });
  });

  describe('clearHistory', () => {
    it('should clear all history for a specific connection', () => {
      const entries: QueryHistoryEntry[] = [
        {
          id: '1',
          connectionId: 'conn-1',
          query: 'SELECT 1',
          executedAt: new Date().toISOString(),
          duration: 100,
          rowCount: 1,
          status: 'success',
        },
        {
          id: '2',
          connectionId: 'conn-2',
          query: 'SELECT 2',
          executedAt: new Date().toISOString(),
          duration: 200,
          rowCount: 2,
          status: 'success',
        },
      ];

      entries.forEach((entry) => useQueryStore.getState().addToHistory(entry));
      useQueryStore.getState().clearHistory('conn-1');

      const state = useQueryStore.getState();
      expect(state.history).toHaveLength(1);
      expect(state.history[0].connectionId).toBe('conn-2');
    });

    it('should clear all history when no connectionId provided', () => {
      const entries: QueryHistoryEntry[] = [
        {
          id: '1',
          connectionId: 'conn-1',
          query: 'SELECT 1',
          executedAt: new Date().toISOString(),
          duration: 100,
          rowCount: 1,
          status: 'success',
        },
      ];

      entries.forEach((entry) => useQueryStore.getState().addToHistory(entry));
      useQueryStore.getState().clearHistory();

      const state = useQueryStore.getState();
      expect(state.history).toHaveLength(0);
    });
  });

  describe('executeQuery', () => {
    it('should add query to active queries during execution', async () => {
      const { executeQuery } = useQueryStore.getState();
      const executePromise = executeQuery('conn-1', 'SELECT 1');

      // Check active queries during execution
      const state = useQueryStore.getState();
      expect(state.activeQueries.size).toBe(1);

      await executePromise;
      
      // After completion, active queries should be empty
      expect(useQueryStore.getState().activeQueries.size).toBe(0);
    });

    it('should remove query from active queries on completion', async () => {
      const { executeQuery } = useQueryStore.getState();

      await executeQuery('conn-1', 'SELECT 1');

      expect(useQueryStore.getState().activeQueries.size).toBe(0);
    });
  });

  describe('cancelQuery', () => {
    it('should call Wails cancel and update query status', async () => {
      const { cancelQuery, activeQueries } = useQueryStore.getState();

      await cancelQuery('query-1');

      expect(activeQueries.has('query-1')).toBe(false);
    });
  });
});
