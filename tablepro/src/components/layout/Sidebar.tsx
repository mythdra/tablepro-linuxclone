import { useState } from 'react';
import { ChevronRight, ChevronDown, Database, Table, Eye } from 'lucide-react';
import { useSchemas, useTables } from '@/hooks/useSchema';
import { useAppStore } from '@/stores/appStore';
import type { TableInfo } from '@/types';

export function Sidebar() {
  const { activeConnectionId, sidebarCollapsed, addTab } = useAppStore();
  const [expandedSchemas, setExpandedSchemas] = useState<Set<string>>(new Set());

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

  if (sidebarCollapsed) {
    return (
      <div className="w-12 bg-muted border-r flex flex-col items-center py-2">
        <Database className="w-5 h-5 text-muted-foreground" />
      </div>
    );
  }

  return (
    <aside className="w-64 bg-muted/50 border-r flex flex-col">
      <div className="p-2 border-b flex items-center gap-2">
        <Database className="w-4 h-4" />
        <span className="text-sm font-medium">Schema</span>
      </div>
      <div className="flex-1 overflow-auto p-1">
        {isLoading ? (
          <div className="p-2 text-sm text-muted-foreground">Loading...</div>
        ) : (
          schemas?.map((schema) => (
            <SchemaNode
              key={schema}
              schema={schema}
              expanded={expandedSchemas.has(schema)}
              onToggle={() => toggleSchema(schema)}
              onTableClick={handleTableClick}
            />
          ))
        )}
      </div>
    </aside>
  );
}

function SchemaNode({
  schema,
  expanded,
  onToggle,
  onTableClick,
}: {
  schema: string;
  expanded: boolean;
  onToggle: () => void;
  onTableClick: (schema: string, table: TableInfo) => void;
}) {
  const { activeConnectionId } = useAppStore();
  const { data: tables, isLoading } = useTables(activeConnectionId, schema);

  return (
    <div>
      <button
        onClick={onToggle}
        className="w-full flex items-center gap-1 px-2 py-1 hover:bg-accent rounded text-sm"
      >
        {expanded ? (
          <ChevronDown className="w-3 h-3" />
        ) : (
          <ChevronRight className="w-3 h-3" />
        )}
        <span className="font-medium">{schema}</span>
      </button>
      {expanded && (
        <div className="ml-3">
          {isLoading ? (
            <div className="px-2 py-1 text-xs text-muted-foreground">Loading...</div>
          ) : (
            tables?.map((table) => (
              <button
                key={table.name}
                onClick={() => onTableClick(schema, table)}
                className="w-full flex items-center gap-2 px-2 py-0.5 hover:bg-accent rounded text-sm"
              >
                {table.type === 'view' ? (
                  <Eye className="w-3 h-3 text-blue-500" />
                ) : (
                  <Table className="w-3 h-3 text-green-500" />
                )}
                <span>{table.name}</span>
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
}
