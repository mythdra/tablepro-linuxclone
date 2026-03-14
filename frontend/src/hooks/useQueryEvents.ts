import { useEffect } from 'react';
import { EventsOn } from '../wailsjs/runtime';
import { useQueryStore } from '../stores/queryStore';
import type { ActiveQuery, QueryHistoryEntry } from '../types';

/**
 * Custom hook for subscribing to query lifecycle events from Wails backend.
 * 
 * Subscribes to events:
 * - query:executing - Query started, adds to active queries
 * - query:completed - Query finished, removes from active and adds to history
 * - query:failed - Query failed, removes from active queries
 * - history:added - History entry added, syncs with store
 * 
 * Automatically cleans up all subscriptions on unmount.
 * 
 * @example
 * ```tsx
 * function QueryPanel() {
 *   // Subscribe to query events
 *   useQueryEvents();
 *   
 *   return <div>...</div>;
 * }
 * ```
 */
export function useQueryEvents() {
  const { activeQueries, addToHistory } = useQueryStore.getState();

  useEffect(() => {
    // Subscribe to query lifecycle events
    const unsubscribeExecuting = EventsOn('query:executing', (data: any) => {
      const activeQuery: ActiveQuery = {
        queryId: data.queryId,
        connectionId: data.connectionId,
        query: data.query,
        status: 'executing',
        startTime: new Date().toISOString(),
      };

      const updated = new Map(activeQueries);
      updated.set(data.queryId, activeQuery);
      useQueryStore.setState({ activeQueries: updated });
    });

    const unsubscribeCompleted = EventsOn('query:completed', (data: any) => {
      // Remove from active queries
      const updated = new Map(activeQueries);
      updated.delete(data.queryId);
      useQueryStore.setState({ activeQueries: updated });

      // Add to history
      const historyEntry: QueryHistoryEntry = {
        id: data.queryId,
        connectionId: data.connectionId,
        query: data.query || '',
        executedAt: new Date().toISOString(),
        duration: data.duration || 0,
        rowCount: data.rowCount || 0,
        status: data.status || 'success',
      };
      addToHistory(historyEntry);
    });

    const unsubscribeFailed = EventsOn('query:failed', (data: any) => {
      // Remove from active queries
      const updated = new Map(activeQueries);
      updated.delete(data.queryId);
      useQueryStore.setState({ activeQueries: updated });
    });

    const unsubscribeHistory = EventsOn('history:added', (data: any) => {
      addToHistory(data as QueryHistoryEntry);
    });

    // Cleanup on unmount
    return () => {
      unsubscribeExecuting();
      unsubscribeCompleted();
      unsubscribeFailed();
      unsubscribeHistory();
    };
  }, [activeQueries, addToHistory]);
}
