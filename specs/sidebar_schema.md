# Sidebar Schema Internals (React + Go)

## Overview
The sidebar displays the database schema tree. Go fetches schema data lazily, caches it, and serves it to the React tree component.

## 1. Lazy Loading Pattern
```go
type SchemaCache struct {
    mu       sync.RWMutex
    schemas  map[string]*SchemaInfo  // cached per database
    tables   map[string][]TableInfo  // cached per schema
    expiry   time.Duration           // default 5 minutes
}

func (sc *SchemaCache) GetTables(connectionID uuid.UUID, schema string) ([]TableInfo, error) {
    sc.mu.RLock()
    if cached, ok := sc.tables[schema]; ok {
        sc.mu.RUnlock()
        return cached, nil
    }
    sc.mu.RUnlock()

    // Cache miss — fetch from driver
    tables, err := driver.FetchTables(schema)
    if err != nil {
        return nil, err
    }

    sc.mu.Lock()
    sc.tables[schema] = tables
    sc.mu.Unlock()

    return tables, nil
}
```

## 2. React Tree Component
```typescript
// Using a recursive tree with expand/collapse
function SchemaTree({ schemas }: { schemas: SchemaInfo[] }) {
  return schemas.map(schema => (
    <TreeNode key={schema.name} label={schema.name}>
      <LazyTreeSection
        label="Tables"
        icon={<TableIcon />}
        loadFn={() => GetTables(connectionId, schema.name)}
        renderItem={(table) => (
          <TableNode table={table} onDoubleClick={() => openTable(table)} />
        )}
      />
      <LazyTreeSection label="Views" icon={<ViewIcon />}
        loadFn={() => GetViews(connectionId, schema.name)} />
      <LazyTreeSection label="Routines" icon={<FunctionIcon />}
        loadFn={() => GetRoutines(connectionId, schema.name)} />
    </TreeNode>
  ));
}
```

## 3. Client-Side Search
```typescript
const [searchQuery, setSearchQuery] = useState('');
const filteredTables = useMemo(() =>
  allTables.filter(t =>
    t.name.toLowerCase().includes(searchQuery.toLowerCase())
  ),
  [allTables, searchQuery]
);
```
- Search input at top of sidebar
- Filters across all schemas in real-time
- Highlights matching text in results

## 4. Context Menu (Right-Click)
```typescript
const contextMenuItems = [
  { label: 'Open Table', action: () => openTable(table) },
  { label: 'Copy Name', action: () => clipboard.writeText(table.name) },
  { separator: true },
  { label: 'Truncate Table...', action: () => confirmTruncate(table), danger: true },
  { label: 'Drop Table...', action: () => confirmDrop(table), danger: true },
  { separator: true },
  { label: 'Show DDL', action: () => showDDL(table) },
];
```

## 5. Cache Invalidation
- After any DDL operation (DROP, CREATE, ALTER), Go calls `SchemaCache.Invalidate(schema)`
- Emits `runtime.EventsEmit(ctx, "schema:refresh")` → React re-fetches tree
- Manual refresh button in sidebar header

## 6. Batch Operations
- Multi-select tables (Cmd+Click / Shift+Click)
- Right-click → "Drop Selected" / "Truncate Selected"
- Confirmation dialog listing all selected tables
- Go executes operations sequentially in a transaction
