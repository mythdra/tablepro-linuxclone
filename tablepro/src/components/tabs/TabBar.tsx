import { TabsList } from '@/components/ui/tabs';
import { X } from 'lucide-react';
import { useAppStore } from '@/stores/appStore';

export function TabBar() {
  const { tabs, activeTabId, setActiveTab, closeTab } = useAppStore();

  if (tabs.length === 0) {
    return null;
  }

  return (
    <div className="border-b flex items-center">
      <TabsList className="rounded-none bg-transparent h-9 p-0">
        {tabs.map((tab) => (
          <div
            key={tab.id}
            className={`flex items-center gap-2 px-3 py-1.5 text-sm border-r cursor-pointer ${
              activeTabId === tab.id
                ? 'bg-background text-foreground'
                : 'text-muted-foreground hover:text-foreground'
            }`}
            onClick={() => setActiveTab(tab.id)}
          >
            <span>{tab.title}</span>
            <button
              onClick={(e) => {
                e.stopPropagation();
                closeTab(tab.id);
              }}
              className="hover:bg-accent rounded p-0.5"
            >
              <X className="w-3 h-3" />
            </button>
          </div>
        ))}
      </TabsList>
    </div>
  );
}
