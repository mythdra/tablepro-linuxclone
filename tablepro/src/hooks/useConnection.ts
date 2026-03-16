import { useMutation } from '@tanstack/react-query';
import { tauriApi } from '@/lib/tauri';
import { useAppStore } from '@/stores/appStore';

interface ConnectVariables {
  connectionId: string;
  dbType: string;
  host: string;
  port: number;
  database: string;
  username: string;
  password: string;
}

export function useConnection() {
  const { setConnectionInfo, setActiveConnection } = useAppStore();

  const connectMutation = useMutation({
    mutationFn: async (vars: ConnectVariables) => {
      return tauriApi.connect(
        vars.connectionId,
        vars.dbType,
        vars.host,
        vars.port,
        vars.database,
        vars.username,
        vars.password
      );
    },
    onSuccess: (info, variables) => {
      setConnectionInfo(variables.connectionId, info);
      setActiveConnection(variables.connectionId);
    },
  });

  const disconnectMutation = useMutation({
    mutationFn: (connectionId: string) => tauriApi.disconnect(connectionId),
    onSuccess: () => {
      setActiveConnection(null);
    },
  });

  const testMutation = useMutation({
    mutationFn: async (vars: { dbType: string; host: string; port: number; database: string; username: string; password: string }) => {
      return tauriApi.testConnection(
        vars.dbType,
        vars.host,
        vars.port,
        vars.database,
        vars.username,
        vars.password
      );
    },
  });

  return {
    connect: connectMutation.mutateAsync,
    disconnect: disconnectMutation.mutateAsync,
    test: testMutation.mutateAsync,
    isConnecting: connectMutation.isPending,
    isDisconnecting: disconnectMutation.isPending,
    isTesting: testMutation.isPending,
    error: connectMutation.error || disconnectMutation.error,
  };
}
