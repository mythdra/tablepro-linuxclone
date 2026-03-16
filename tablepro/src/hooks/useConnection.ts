import { useMutation } from '@tanstack/react-query';
import { tauriApi } from '@/lib/tauri';
import { useAppStore } from '@/stores/appStore';

export function useConnection() {
  const { setConnectionInfo, setActiveConnection } = useAppStore();

  const connectMutation = useMutation({
    mutationFn: tauriApi.connect,
    onSuccess: (info, variables) => {
      setConnectionInfo(variables.id, info);
      setActiveConnection(variables.id);
    },
  });

  const disconnectMutation = useMutation({
    mutationFn: (connectionId: string) => tauriApi.disconnect(connectionId),
    onSuccess: (_, connectionId) => {
      setActiveConnection(null);
    },
  });

  const testMutation = useMutation({
    mutationFn: tauriApi.testConnection,
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
