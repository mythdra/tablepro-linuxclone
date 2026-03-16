import { Sidebar } from './Sidebar';
import { Toolbar } from './Toolbar';
import { TabBar } from '../tabs/TabBar';
import { QueryTab } from '../tabs/QueryTab';
import { RightPanel } from './RightPanel';
import { CommandPalette } from './CommandPalette';
import { useAppStore } from '@/stores/appStore';
import { Database, Terminal } from 'lucide-react';

export function MainLayout() {
  const { tabs, activeTabId, activeConnectionId } = useAppStore();
  const activeTab = tabs.find((t) => t.id === activeTabId);

  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className="flex flex-col flex-1 overflow-hidden">
        <Toolbar />
        <TabBar />
        <div className="flex flex-1 overflow-hidden">
          <main className="flex-1 overflow-hidden">
            {activeTab ? (
              <QueryTab tab={activeTab} />
            ) : (
              <div className="flex flex-col items-center justify-center h-full text-muted-foreground">
                <div className="flex flex-col items-center gap-4 p-8 max-w-md">
                  <div className="w-20 h-20 rounded-2xl bg-muted flex items-center justify-center">
                    {activeConnectionId ? (
                      <Terminal className="w-10 h-10 text-primary" />
                    ) : (
                      <Database className="w-10 h-10 text-primary" />
                    )}
                  </div>
                  <div className="text-center">
                    <h3 className="text-lg font-semibold text-foreground mb-2">
                      {activeConnectionId ? 'No query tab open' : 'No connection'}
                    </h3>
                    <p className="text-sm opacity-80 mb-4">
                      {activeConnectionId
                        ? 'Select a table from the sidebar or create a new query tab to get started.'
                        : 'Connect to a database to start exploring your data.'}
                    </p>
                  </div>
                </div>
              </div>
            )}
          </main>
          <RightPanel />
        </div>

        {/* Status Bar */}
        <StatusBar />
      </div>

      {/* Command Palette */}
      <CommandPalette />
    </div>
  );
}

function StatusBar() {
  const { activeConnectionId, connectionInfos } = useAppStore();

  const connection = activeConnectionId
    ? connectionInfos.get(activeConnectionId)
    : null;

  return (
    <div className="h-6 border-t bg-card flex items-center px-3 text-xs text-muted-foreground">
      <div className="flex items-center gap-4">
        {connection ? (
          <>
            <span className="flex items-center gap-1.5">
              <span className="w-2 h-2 rounded-full bg-success" />
              Connected
            </span>
            <span className="opacity-60">{connection.host}:{connection.port}</span>
            <span className="opacity-60">{connection.database}</span>
          </>
        ) : (
          <span className="flex items-center gap-1.5">
            <span className="w-2 h-2 rounded-full bg-muted-foreground/30" />
            Disconnected
          </span>
        )}
      </div>
      <div className="flex-1" />
      <div className="flex items-center gap-4">
        <span>TablePro v0.1.0</span>
      </div>
    </div>
  );
}
