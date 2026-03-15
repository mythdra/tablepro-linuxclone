import { create } from 'zustand';
import { nanoid } from 'nanoid';

export interface Tab {
  id: string;
  name: string;
  query: string;
  connectionId: string | null;
  isDirty: boolean;
  rowCount?: number;
  isExact?: boolean;
}

interface TabState {
  tabs: Tab[];
  activeTabId: string | null;
  addTab: (tab: Partial<Tab>) => void;
  removeTab: (tabId: string) => void;
  setActiveTab: (tabId: string) => void;
  updateTab: (tabId: string, updates: Partial<Tab>) => void;
}

export const useTabStore = create<TabState>((set, get) => ({
  tabs: [],
  activeTabId: null,

  addTab: (tab) => set((state) => {
    const newTab: Tab = {
      id: nanoid(),
      name: tab.name || 'Query',
      query: tab.query || '',
      connectionId: tab.connectionId || state.tabs[0]?.connectionId || null,
      isDirty: false,
      ...tab,
    };
    return {
      tabs: [...state.tabs, newTab],
      activeTabId: newTab.id,
    };
  }),

  removeTab: (tabId) => set((state) => {
    const newTabs = state.tabs.filter(t => t.id !== tabId);
    return {
      tabs: newTabs,
      activeTabId: state.activeTabId === tabId
        ? (newTabs.length > 0 ? newTabs[newTabs.length - 1].id : null)
        : state.activeTabId,
    };
  }),

  setActiveTab: (tabId) => set({ activeTabId: tabId }),

  updateTab: (tabId, updates) => set((state) => ({
    tabs: state.tabs.map(tab =>
      tab.id === tabId ? { ...tab, ...updates } : tab
    ),
  })),
}));
