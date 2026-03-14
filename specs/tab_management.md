# Tab Management & Session Restoration (Go + React)

## 1. Tab State in React
```typescript
interface Tab {
  id: string;
  type: 'table' | 'query' | 'structure';
  title: string;
  query: string;
  tableName?: string;
  schemaName?: string;
  isView?: boolean;
  databaseName: string;
  isExecuting: boolean;
  // Results are fetched from Go, not stored in tab state
}

const useTabStore = create<TabStore>((set, get) => ({
  tabs: [],
  activeTabId: null,

  addTab: (tab) => {
    set(s => ({ tabs: [...s.tabs, tab], activeTabId: tab.id }));
    TabManager.SaveTabs(get().tabs, tab.id); // Persist to Go
  },

  closeTab: (tabId) => {
    set(s => {
      const newTabs = s.tabs.filter(t => t.id !== tabId);
      const newActive = s.activeTabId === tabId ? newTabs[0]?.id : s.activeTabId;
      TabManager.SaveTabs(newTabs, newActive);
      return { tabs: newTabs, activeTabId: newActive };
    });
  },

  switchTab: (tabId) => {
    set({ activeTabId: tabId });
    TabManager.SaveTabs(get().tabs, tabId);
  },
}));
```

## 2. Tab Persistence (Go Backend)
```go
type TabManager struct {
    ctx     context.Context
    baseDir string // ~/.config/tablepro/tabs/
}

func (tm *TabManager) SaveTabs(connectionID string, tabs []PersistedTab, activeTabID string) error {
    // Filter out preview tabs
    // Truncate queries > 500KB
    state := TabDiskState{Tabs: tabs, SelectedTabID: activeTabID}
    data, _ := json.Marshal(state)
    path := filepath.Join(tm.baseDir, connectionID+".json")
    return os.WriteFile(path, data, 0644)
}

func (tm *TabManager) RestoreTabs(connectionID string) (*TabDiskState, error) {
    path := filepath.Join(tm.baseDir, connectionID+".json")
    data, err := os.ReadFile(path)
    if err != nil { return nil, nil } // No saved state
    var state TabDiskState
    json.Unmarshal(data, &state)
    return &state, nil
}
```

## 3. App Quit — Synchronous Save
```go
// In main.go, Wails OnBeforeClose hook
func (a *App) beforeClose(ctx context.Context) bool {
    // Save all open tabs synchronously before quit
    for connID, tabs := range a.tabManager.GetAllOpenTabs() {
        a.tabManager.SaveTabsSync(connID, tabs)
    }
    return false // Allow close
}
```
- Uses Wails' `OnBeforeClose` hook — guarantees save before process exit
- Synchronous write (no goroutines) since the event loop is shutting down

## 4. Memory Management (React-side)
Unlike the Swift version's LRU eviction of `NSTableView` data, the React version handles memory differently:

- **AG Grid handles its own virtualization** — only visible rows are in DOM
- **Result data** is stored in Go, not React. React only holds the current page (500 rows)
- **Tab switching** doesn't need LRU eviction because result data stays in Go's memory
- **Go-side eviction**: If too many connections are open, Go can evict cached query results:
```go
func (qm *QueryManager) EvictOldResults(maxCached int) {
    // Sort tabs by lastExecutedAt
    // Keep only maxCached most recent results
    // Evicted tabs re-execute query on next switch
}
```

## 5. Lazy Re-Query on Tab Switch
When user switches to a tab whose results were evicted:
1. React sends `QueryManager.GetResults(tabID)`
2. Go checks if results are cached → if not, re-executes the saved query
3. Results streamed back to React via return value
4. AG Grid populates seamlessly — user doesn't notice the re-fetch
