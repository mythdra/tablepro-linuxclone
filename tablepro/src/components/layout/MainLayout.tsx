import { Sidebar } from './Sidebar';
import { Toolbar } from './Toolbar';
import { TabBar } from '../tabs/TabBar';
import { QueryTab } from '../tabs/QueryTab';
import { useAppStore } from '@/stores/appStore';

export function MainLayout() {
  const { tabs, activeTabId } = useAppStore();
  const activeTab = tabs.find((t) => t.id === activeTabId);

  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className="flex flex-col flex-1 overflow-hidden">
        <Toolbar />
        <TabBar />
        <main className="flex-1 overflow-hidden">
          {activeTab ? (
            <QueryTab tab={activeTab} />
          ) : (
            <div className="flex items-center justify-center h-full text-muted-foreground">
              No tab open. Press Ctrl+T to create a new query.
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
