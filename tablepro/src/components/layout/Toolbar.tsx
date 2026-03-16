import {
  Plus,
  Play,
  Square,
  Settings,
  PanelLeftClose,
  PanelLeft,
  PanelRightClose,
  PanelRight,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useAppStore } from '@/stores/appStore';
import { ConnectionDialog } from '@/components/connection/ConnectionDialog';
import { useState, useEffect } from 'react';

export function Toolbar() {
  const {
    activeConnectionId,
    connectionInfos,
    sidebarCollapsed,
    rightPanelOpen,
    toggleSidebar,
    toggleRightPanel,
  } = useAppStore();
  const [connectionDialogOpen, setConnectionDialogOpen] = useState(false);

  const activeConnection = activeConnectionId
    ? connectionInfos.get(activeConnectionId)
    : null;

  // Keyboard shortcut for new tab + CommandPalette events
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 't') {
        e.preventDefault();
        setConnectionDialogOpen(true);
      }
    };

    const handleOpenConnection = () => setConnectionDialogOpen(true);
    const handleNewTab = () => setConnectionDialogOpen(true);

    document.addEventListener('keydown', handleKeyDown);
    window.addEventListener('tablepro:open-connection-dialog', handleOpenConnection);
    window.addEventListener('tablepro:new-tab', handleNewTab);

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('tablepro:open-connection-dialog', handleOpenConnection);
      window.removeEventListener('tablepro:new-tab', handleNewTab);
    };
  }, []);

  return (
    <div className="h-11 border-b bg-card flex items-center px-2 gap-1.5">
      {/* Sidebar toggle */}
      <Button
        variant="ghost"
        size="sm"
        onClick={toggleSidebar}
        className="gap-1.5"
        title={sidebarCollapsed ? 'Show sidebar' : 'Hide sidebar'}
      >
        {sidebarCollapsed ? (
          <PanelLeft className="w-4 h-4" />
        ) : (
          <PanelLeftClose className="w-4 h-4" />
        )}
      </Button>

      <div className="w-px h-5 bg-border mx-1" />

      {/* New connection */}
      <Button
        variant="ghost"
        size="sm"
        onClick={() => setConnectionDialogOpen(true)}
        className="gap-1.5"
      >
        <Plus className="w-4 h-4" />
        <span className="hidden sm:inline">Connect</span>
      </Button>

      {/* Connection status */}
      {activeConnection && (
        <>
          <div className="w-px h-5 bg-border mx-1" />
          <div className="flex items-center gap-2 px-2 py-1 bg-muted rounded-md">
            <div className="w-2 h-2 rounded-full bg-success animate-pulse" />
            <div className="flex flex-col">
              <span className="text-sm font-medium leading-none">
                {activeConnection.name}
              </span>
              <span className="text-xs text-muted-foreground leading-none mt-0.5">
                {activeConnection.database} @ {activeConnection.host}
              </span>
            </div>
          </div>
        </>
      )}

      {/* Spacer */}
      <div className="flex-1" />

      {/* Right side actions */}
      <div className="flex items-center gap-1">
        {activeConnection && (
          <>
            <Button variant="ghost" size="sm" className="gap-1.5" title="Run (Ctrl+Enter)">
              <Play className="w-4 h-4 text-success" />
              <span className="hidden sm:inline">Run</span>
            </Button>
            <Button variant="ghost" size="sm" className="gap-1.5" title="Stop (Escape)">
              <Square className="w-4 h-4 text-destructive" />
            </Button>
            <div className="w-px h-5 bg-border mx-1" />
          </>
        )}
        <Button
          variant="ghost"
          size="sm"
          onClick={toggleRightPanel}
          title={rightPanelOpen ? 'Hide panel' : 'Show panel'}
        >
          {rightPanelOpen ? (
            <PanelRightClose className="w-4 h-4" />
          ) : (
            <PanelRight className="w-4 h-4" />
          )}
        </Button>
        <Button variant="ghost" size="sm" title="Settings">
          <Settings className="w-4 h-4" />
        </Button>
      </div>

      <ConnectionDialog
        open={connectionDialogOpen}
        onClose={() => setConnectionDialogOpen(false)}
      />
    </div>
  );
}
