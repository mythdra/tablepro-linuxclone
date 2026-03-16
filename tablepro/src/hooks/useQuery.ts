import { useMutation } from '@tanstack/react-query';
import { tauriApi } from '@/lib/tauri';

export function useQuery() {
  const executeMutation = useMutation({
    mutationFn: ({
      connectionId,
      sql,
      limit,
    }: {
      connectionId: string;
      sql: string;
      limit?: number;
    }) => tauriApi.executeQuery(connectionId, sql, limit),
  });

  return {
    execute: executeMutation.mutateAsync,
    isExecuting: executeMutation.isPending,
    result: executeMutation.data,
    error: executeMutation.error,
  };
}
