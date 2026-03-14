# UI Components & Navigation (React + Wails)

## Overview
The UI is a single-page React application rendered inside Wails' WebView. Component library uses **Radix UI** primitives for accessibility, styled with **Tailwind CSS**. State is managed by **Zustand** stores that call Wails-bound Go methods.

## Component Architecture
```
App.tsx
├── Toolbar.tsx
├── PanelLayout.tsx (react-resizable-panels)
│   ├── Sidebar/
│   │   ├── SchemaTree.tsx
│   │   ├── TableSearchBar.tsx
│   │   └── ContextMenu.tsx
│   ├── MainWorkspace/
│   │   ├── TabBar.tsx
│   │   ├── QueryEditor.tsx (Monaco)
│   │   ├── DataGrid.tsx (AG Grid)
│   │   └── StatusBar.tsx
│   └── RightPanel/
│       ├── AIChatPanel.tsx
│       ├── QueryHistoryPanel.tsx
│       └── FormatterPanel.tsx
├── Modals/
│   ├── ConnectionForm.tsx
│   ├── ExportDialog.tsx
│   ├── ImportDialog.tsx
│   └── SettingsDialog.tsx
└── Providers/
    ├── ThemeProvider.tsx
    └── KeyboardShortcutProvider.tsx
```

## State Management (Zustand)
Each major feature has its own store, calling Go backend methods:
```typescript
// stores/connectionStore.ts
import { GetConnections, SaveConnection, TestConnection } from '../wailsjs/go/main/ConnectionManager';

export const useConnectionStore = create((set) => ({
  connections: [],
  loadConnections: async () => {
    const conns = await GetConnections();
    set({ connections: conns });
  },
  testConnection: async (config) => {
    return await TestConnection(config);
  },
}));
```

## Event-Driven Updates
Go pushes real-time updates to React:
```go
// Go backend
runtime.EventsEmit(ctx, "query:progress", progressData)
runtime.EventsEmit(ctx, "connection:status", statusData)
```
```typescript
// React frontend
import { EventsOn } from '../wailsjs/runtime/runtime';
useEffect(() => {
  EventsOn("query:progress", (data) => { /* update progress bar */ });
}, []);
```

## MainContentCoordinator → React Equivalent
The Swift `MainContentCoordinator` monolith maps to multiple focused Zustand stores:
- `useTabStore` — tab CRUD, selection, persistence
- `useQueryStore` — query execution, results, pagination
- `useChangeStore` — cell edits, pending changes, undo/redo
- `useFilterStore` — column filters, sort state
- `useSidebarStore` — schema tree, search, selection
