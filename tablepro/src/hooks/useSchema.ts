import { useQuery } from '@tanstack/react-query';
import { tauriApi } from '@/lib/tauri';

export function useSchemas(connectionId: string | null) {
  return useQuery({
    queryKey: ['schemas', connectionId],
    queryFn: async () => {
      if (!connectionId) throw new Error('No connection');
      return tauriApi.getSchemas(connectionId);
    },
    enabled: connectionId !== null,
  });
}

export function useTables(connectionId: string | null, schema: string) {
  return useQuery({
    queryKey: ['tables', connectionId, schema],
    queryFn: async () => {
      if (!connectionId) throw new Error('No connection');
      return tauriApi.getTables(connectionId, schema);
    },
    enabled: connectionId !== null && schema !== '',
  });
}

export function useColumns(
  connectionId: string | null,
  schema: string,
  table: string
) {
  return useQuery({
    queryKey: ['columns', connectionId, schema, table],
    queryFn: async () => {
      if (!connectionId) throw new Error('No connection');
      return tauriApi.getColumns(connectionId, schema, table);
    },
    enabled: connectionId !== null && schema !== '' && table !== '',
  });
}
