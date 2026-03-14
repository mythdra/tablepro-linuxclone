import { X, Plus, Code2 } from 'lucide-react';
import type { QueryTab } from '../types';

interface TabBarProps {
  tabs: QueryTab[];
  activeTabId: string | null;
  onTabSelect: (tabId: string) => void;
  onTabClose: (tabId: string) => void;
  onNewTab: () => void;
}

export function TabBar({ tabs, activeTabId, onTabSelect, onTabClose, onNewTab }: TabBarProps) {
  return (
    <div className="flex items-center bg-slate-900 border-b border-slate-700 overflow-x-auto">
      <div className="flex items-center flex-1 min-w-0">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => onTabSelect(tab.id)}
            className={`group flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-r border-slate-700 min-w-0 max-w-48 transition-colors ${
              activeTabId === tab.id
                ? 'bg-slate-800 text-white border-b-2 border-b-primary'
                : 'text-slate-400 hover:text-white hover:bg-slate-800/50'
            }`}
            aria-selected={activeTabId === tab.id}
            role="tab"
          >
            <Code2 className="w-4 h-4 flex-shrink-0" />
            <span className="truncate">{tab.name}</span>
            {tab.isDirty && (
              <span className="w-1.5 h-1.5 rounded-full bg-amber-500 flex-shrink-0" />
            )}
            <button
              onClick={(e) => {
                e.stopPropagation();
                onTabClose(tab.id);
              }}
              className="ml-1 p-0.5 rounded hover:bg-slate-600 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0"
              aria-label={`Close ${tab.name}`}
            >
              <X className="w-3.5 h-3.5" />
            </button>
          </button>
        ))}
      </div>
      <button
        onClick={onNewTab}
        className="flex items-center justify-center w-10 h-10 text-slate-400 hover:text-white hover:bg-slate-800 transition-colors flex-shrink-0"
        aria-label="New query tab"
      >
        <Plus className="w-5 h-5" />
      </button>
    </div>
  );
}

export default TabBar;