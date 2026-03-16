import { useState, useMemo } from 'react';
import { Database, Table, Eye, Search, Loader2, Plug, Settings2 } from 'lucide-react';
import { useSchemas, useTables } from '@/hooks/useSchema';
import { useAppStore } from '@/stores/appStore';
import type { TableInfo } from '@/types';
import { cn } from '@/lib/utils';
import { tauriApi } from '@/lib/tauri';

type SidebarTab = 'connections' | 'schema';

export function Sidebar() {
  const { sidebarCollapsed, setSidebarCollapsed } = useAppStore();
  const [activeTab, setActiveTab] = useState<SidebarTab>('connections');

  if (sidebarCollapsed) {
    return (
      <aside className="w-12 bg-card border-r flex flex-col items-center py-3 gap-2">
        <button
          onClick={() => setSidebarCollapsed(false)}
          className="p-2 rounded-md hover:bg-accent transition-colors cursor-pointer"
          title="Expand sidebar"
        >
          <Database className="w-5 h-5 text-primary" />
        </button>
        <div className="flex-1" />
        <button
          onClick={() => setSidebarCollapsed(false)}
          className="p-2 rounded-md hover:bg-accent transition-colors cursor-pointer"
        >
          <Settings2 className="w-4 h-4 text-muted-foreground" />
        </button>
      </aside>
    );
  }

  return (
    <aside className="w-64 bg-card border-r flex flex-col">
      {/* Tabs */}
      <div className="flex border-b">
        <button
          onClick={() => setActiveTab('connections')}
          className={cn(
            "flex-1 flex items-center justify-center gap-2 px-3 py-2.5 text-sm font-medium transition-colors cursor-pointer",
            activeTab === 'connections'
              ? "text-primary border-b-2 border-primary bg-muted/30"
              : "text-muted-foreground hover:text-foreground"
          )}
        >
          <Plug className="w-4 h-4" />
          <span>Connections</span>
        </button>
        <button
          onClick={() => setActiveTab('schema')}
          className={cn(
            "flex-1 flex items-center justify-center gap-2 px-3 py-2.5 text-sm font-medium transition-colors cursor-pointer",
            activeTab === 'schema'
              ? "text-primary border-b-2 border-primary bg-muted/30"
              : "text-muted-foreground hover:text-foreground"
          )}
        >
          <Database className="w-4 h-4" />
          <span>Schema</span>
        </button>
        <button
          onClick={() => setSidebarCollapsed(true)}
          className="px-2 hover:bg-accent transition-colors cursor-pointer"
          title="Collapse"
        >
          <Settings2 className="w-4 h-4 text-muted-foreground" />
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-hidden">
        {activeTab === 'connections' ? (
          <ConnectionsList />
        ) : (
          <SchemaTree />
        )}
      </div>
    </aside>
  );
}

function ConnectionsList() {
  const {
    connections,
    activeConnectionId,
    setActiveConnection,
    setConnectionInfo,
  } = useAppStore();

  const handleConnect = async (connectionId: string) => {
    const conn = connections.find(c => c.id === connectionId);
    if (!conn) return;

    try {
      const info = await tauriApi.connect(
        connectionId,
        'postgresql',
        conn.host,
        conn.port,
        conn.database,
        conn.username,
        ''
      );
      // Update connection info with host/port from config
      setConnectionInfo(connectionId, {
        ...info,
        host: conn.host,
        port: conn.port,
      });
      setActiveConnection(connectionId);
    } catch (error) {
      console.error('Failed to connect:', error);
      alert(`Failed to connect: ${error}`);
    }
  };

  if (connections.length === 0) {
    return (
      <div className="p-4 text-center">
        <Plug className="w-8 h-8 text-muted-foreground/50 mx-auto mb-2" />
        <p className="text-sm text-muted-foreground">No connections</p>
        <p className="text-xs text-muted-foreground/70 mt-1">Click + to add</p>
      </div>
    );
  }

  return (
    <div className="p-2 space-y-1">
      {connections.map((conn) => {
        const isActive = conn.id === activeConnectionId;
        return (
          <button
            key={conn.id}
            onClick={() => handleConnect(conn.id)}
            className={cn(
              "w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-left transition-colors cursor-pointer",
              isActive
                ? "bg-primary/10 text-primary"
                : "hover:bg-accent text-foreground"
            )}
          >
            {isActive ? (
              <span className="w-2 h-2 rounded-full bg-success animate-pulse" />
            ) : (
              <span className="w-2 h-2 rounded-full bg-muted-foreground/30" />
            )}
            <div className="flex-1 min-w-0">
              <div className="text-sm font-medium truncate">{conn.name}</div>
              <div className="text-xs text-muted-foreground truncate">
                {conn.host}:{conn.port}/{conn.database}
              </div>
            </div>
          </button>
        );
      })}
    </div>
  );
}

function SchemaTree() {
  const { activeConnectionId, addTab } = useAppStore();
  const [expandedSchemas, setExpandedSchemas] = useState<Set<string>>(new Set());
  const [searchQuery, setSearchQuery] = useState('');

  const { data: schemas, isLoading } = useSchemas(activeConnectionId);

  const toggleSchema = (schema: string) => {
    setExpandedSchemas((prev) => {
      const next = new Set(prev);
      if (next.has(schema)) {
        next.delete(schema);
      } else {
        next.add(schema);
      }
      return next;
    });
  };

  const handleTableClick = (_schema: string, table: TableInfo) => {
    addTab({
      id: crypto.randomUUID(),
      type: 'query',
      title: table.name,
      connectionId: activeConnectionId!,
    });
  };

  return (
    <div className="flex flex-col h-full">
      {/* Search */}
      <div className="p-2 border-b">
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Filter tables..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-9 pr-3 py-1.5 text-sm bg-muted border rounded-md
                       focus:outline-none focus:ring-2 focus:ring-primary/50
                       placeholder:text-muted-foreground"
          />
        </div>
      </div>

      {/* Tree */}
      <div className="flex-1 overflow-auto p-1">
        {!activeConnectionId ? (
          <div className="p-4 text-center">
            <p className="text-sm text-muted-foreground">No connection</p>
            <p className="text-xs text-muted-foreground mt-1">Connect to view schema</p>
          </div>
        ) : isLoading ? (
          <div className="flex items-center justify-center p-4">
            <Loader2 className="w-5 h-5 animate-spin text-primary" />
          </div>
        ) : (
          schemas?.map((schema) => (
            <SchemaNode
              key={schema}
              schema={schema}
              expanded={expandedSchemas.has(schema)}
              onToggle={() => toggleSchema(schema)}
              onTableClick={handleTableClick}
              searchQuery={searchQuery}
            />
          ))
        )}
      </div>
    </div>
  );
}

function SchemaNode({
  schema,
  expanded,
  onToggle,
  onTableClick,
  searchQuery = '',
}: {
  schema: string;
  expanded: boolean;
  onToggle: () => void;
  onTableClick: (schema: string, table: TableInfo) => void;
  searchQuery?: string;
}) {
  const { activeConnectionId } = useAppStore();
  const { data: tables, isLoading } = useTables(activeConnectionId, schema);

  const filteredTables = useMemo(() => {
    if (!tables) return [];
    if (!searchQuery.trim()) return tables;
    const query = searchQuery.toLowerCase();
    return tables.filter((table) =>
      table.name.toLowerCase().includes(query)
    );
  }, [tables, searchQuery]);

  return (
    <div className="mb-0.5">
      <button
        onClick={onToggle}
        className="w-full flex items-center gap-1.5 px-2 py-1.5 hover:bg-accent rounded-md text-sm transition-colors cursor-pointer"
      >
        {expanded ? (
          <Search className="w-3.5 h-3.5 text-muted-foreground rotate-90" />
        ) : (
          <Search className="w-3.5 h-3.5 text-muted-foreground" />
        )}
        <Database className="w-4 h-4 text-primary" />
        <span className="font-medium truncate">{schema}</span>
        {tables && (
          <span className="ml-auto text-xs text-muted-foreground">
            {tables.length}
          </span>
        )}
      </button>
      {expanded && (
        <div className="ml-2 border-l border-border pl-1.5 mt-0.5">
          {isLoading ? (
            <div className="flex items-center gap-2 px-2 py-1.5 text-xs text-muted-foreground">
              <Loader2 className="w-3 h-3 animate-spin" />
              Loading...
            </div>
          ) : filteredTables.length === 0 ? (
            searchQuery ? (
              <div className="px-2 py-1.5 text-xs text-muted-foreground italic">
                No matches
              </div>
            ) : null
          ) : (
            filteredTables.map((table) => (
              <button
                key={table.name}
                onClick={() => onTableClick(schema, table)}
                className="w-full flex items-center gap-2 px-2 py-1 hover:bg-accent rounded-md text-sm transition-colors cursor-pointer group"
              >
                {table.type === 'view' ? (
                  <Eye className="w-3.5 h-3.5 text-info flex-shrink-0" />
                ) : (
                  <Table className="w-3.5 h-3.5 text-success flex-shrink-0" />
                )}
                <span className="truncate">{table.name}</span>
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
}
