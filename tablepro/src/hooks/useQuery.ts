import { useMutation } from '@tanstack/react-query';
import { tauriApi } from '@/lib/tauri';

interface QueryHistoryItem {
  id: string;
  sql: string;
  executedAt: string;
  duration: number;
  rowCount: number;
  connectionId: string;
}

export function useQuery() {
  const executeMutation = useMutation({
    mutationFn: async ({
      connectionId,
      sql,
      limit,
    }: {
      connectionId: string;
      sql: string;
      limit?: number;
    }) => {
      const startTime = Date.now();
      const result = await tauriApi.executeQuery(connectionId, sql, limit);
      const duration = Date.now() - startTime;

      // Save to history
      const historyItem: QueryHistoryItem = {
        id: crypto.randomUUID(),
        sql,
        executedAt: new Date().toISOString(),
        duration,
        rowCount: result.rowCount,
        connectionId,
      };

      // Get existing history
      const saved = localStorage.getItem('tablepro-query-history');
      let history: QueryHistoryItem[] = saved ? JSON.parse(saved) : [];

      // Add new item at the beginning
      history = [historyItem, ...history].slice(0, 100); // Keep last 100

      // Save
      localStorage.setItem('tablepro-query-history', JSON.stringify(history));

      return result;
    },
  });

  return {
    execute: executeMutation.mutateAsync,
    isExecuting: executeMutation.isPending,
    result: executeMutation.data,
    error: executeMutation.error,
  };
}
