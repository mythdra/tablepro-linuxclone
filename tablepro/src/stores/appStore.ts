import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { ConnectionConfig, ConnectionInfo, Tab } from '@/types';

interface AppState {
  connections: ConnectionConfig[];
  activeConnectionId: string | null;
  connectionInfos: Map<string, ConnectionInfo>;
  tabs: Tab[];
  activeTabId: string | null;
  sidebarCollapsed: boolean;
  theme: 'light' | 'dark' | 'system';

  addConnection: (config: ConnectionConfig) => void;
  removeConnection: (id: string) => void;
  setActiveConnection: (id: string | null) => void;
  setConnectionInfo: (id: string, info: ConnectionInfo) => void;

  addTab: (tab: Tab) => void;
  closeTab: (id: string) => void;
  setActiveTab: (id: string) => void;
  updateTab: (id: string, updates: Partial<Tab>) => void;

  toggleSidebar: () => void;
  setTheme: (theme: 'light' | 'dark' | 'system') => void;
}

export const useAppStore = create<AppState>()(
  persist(
    (set) => ({
      connections: [],
      activeConnectionId: null,
      connectionInfos: new Map(),
      tabs: [],
      activeTabId: null,
      sidebarCollapsed: false,
      theme: 'system',

      addConnection: (config) =>
        set((state) => ({
          connections: [...state.connections, config],
        })),

      removeConnection: (id) =>
        set((state) => ({
          connections: state.connections.filter((c) => c.id !== id),
          connectionInfos: (() => {
            const newMap = new Map(state.connectionInfos);
            newMap.delete(id);
            return newMap;
          })(),
        })),

      setActiveConnection: (id) =>
        set({ activeConnectionId: id }),

      setConnectionInfo: (id, info) =>
        set((state) => {
          const newMap = new Map(state.connectionInfos);
          newMap.set(id, info);
          return { connectionInfos: newMap };
        }),

      addTab: (tab) =>
        set((state) => ({
          tabs: [...state.tabs, tab],
          activeTabId: tab.id,
        })),

      closeTab: (id) =>
        set((state) => ({
          tabs: state.tabs.filter((t) => t.id !== id),
          activeTabId:
            state.activeTabId === id
              ? state.tabs[state.tabs.length - 2]?.id ?? null
              : state.activeTabId,
        })),

      setActiveTab: (id) =>
        set({ activeTabId: id }),

      updateTab: (id, updates) =>
        set((state) => ({
          tabs: state.tabs.map((t) =>
            t.id === id ? { ...t, ...updates } : t
          ),
        })),

      toggleSidebar: () =>
        set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),

      setTheme: (theme) => set({ theme }),
    }),
    {
      name: 'tablepro-storage',
      partialize: (state) => ({
        connections: state.connections,
        sidebarCollapsed: state.sidebarCollapsed,
        theme: state.theme,
      }),
    }
  )
);
