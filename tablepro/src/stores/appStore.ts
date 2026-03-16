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
  rightPanelOpen: boolean;
  rightPanelTab: 'history' | 'formatter';
  theme: 'light' | 'dark' | 'system';

  // Connection actions
  addConnection: (config: ConnectionConfig) => void;
  updateConnection: (id: string, config: Partial<ConnectionConfig>) => void;
  removeConnection: (id: string) => void;
  setActiveConnection: (id: string | null) => void;
  setConnectionInfo: (id: string, info: ConnectionInfo) => void;
  clearConnectionInfo: (id: string) => void;

  // Tab actions
  addTab: (tab: Tab) => void;
  closeTab: (id: string) => void;
  setActiveTab: (id: string) => void;
  updateTab: (id: string, updates: Partial<Tab>) => void;

  // UI actions
  toggleSidebar: () => void;
  setSidebarCollapsed: (collapsed: boolean) => void;
  toggleRightPanel: () => void;
  setRightPanelOpen: (open: boolean) => void;
  setRightPanelTab: (tab: 'history' | 'formatter') => void;
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
      rightPanelOpen: false,
      rightPanelTab: 'history',
      theme: 'system',

      // Connection actions
      addConnection: (config) =>
        set((state) => ({
          connections: [...state.connections, config],
        })),

      updateConnection: (id, updates) =>
        set((state) => ({
          connections: state.connections.map((c) =>
            c.id === id ? { ...c, ...updates } : c
          ),
        })),

      removeConnection: (id) =>
        set((state) => ({
          connections: state.connections.filter((c) => c.id !== id),
          connectionInfos: (() => {
            const newMap = new Map(state.connectionInfos);
            newMap.delete(id);
            return newMap;
          })(),
          activeConnectionId:
            state.activeConnectionId === id ? null : state.activeConnectionId,
        })),

      setActiveConnection: (id) =>
        set({ activeConnectionId: id }),

      setConnectionInfo: (id, info) =>
        set((state) => {
          const newMap = new Map(state.connectionInfos);
          newMap.set(id, info);
          return { connectionInfos: newMap };
        }),

      clearConnectionInfo: (id) =>
        set((state) => {
          const newMap = new Map(state.connectionInfos);
          newMap.delete(id);
          return { connectionInfos: newMap };
        }),

      // Tab actions
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

      // UI actions
      toggleSidebar: () =>
        set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),

      setSidebarCollapsed: (collapsed: boolean) =>
        set({ sidebarCollapsed: collapsed }),

      toggleRightPanel: () =>
        set((state) => ({ rightPanelOpen: !state.rightPanelOpen })),

      setRightPanelOpen: (open: boolean) =>
        set({ rightPanelOpen: open }),

      setRightPanelTab: (tab) =>
        set({ rightPanelTab: tab }),

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
