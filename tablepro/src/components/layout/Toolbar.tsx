import { Database, Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useAppStore } from '@/stores/appStore';
import { ConnectionDialog } from '@/components/connection/ConnectionDialog';
import { useState } from 'react';

export function Toolbar() {
  const { activeConnectionId, connectionInfos } = useAppStore();
  const [connectionDialogOpen, setConnectionDialogOpen] = useState(false);

  const activeConnection = activeConnectionId
    ? connectionInfos.get(activeConnectionId)
    : null;

  return (
    <div className="h-10 border-b flex items-center px-2 gap-2">
      <Button
        variant="outline"
        size="sm"
        onClick={() => setConnectionDialogOpen(true)}
      >
        <Plus className="w-4 h-4 mr-1" />
        New Connection
      </Button>

      {activeConnection && (
        <div className="flex items-center gap-2 ml-auto">
          <Database className="w-4 h-4 text-green-500" />
          <span className="text-sm">{activeConnection.name}</span>
          <span className="text-xs text-muted-foreground">
            ({activeConnection.database})
          </span>
        </div>
      )}

      <ConnectionDialog
        open={connectionDialogOpen}
        onClose={() => setConnectionDialogOpen(false)}
      />
    </div>
  );
}
