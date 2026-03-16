import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { useQuery } from '@/hooks/useQuery';
import { Play, Loader2, Table2 } from 'lucide-react';
import type { Tab } from '@/types';
import { SqlEditor } from './SqlEditor';
import { useAppStore } from '@/stores/appStore';

export function QueryTab({ tab }: { tab: Tab }) {
  const [sql, setSql] = useState('');
  const { setRightPanelTab, setRightPanelOpen } = useAppStore();

  const { execute, isExecuting, result, error } = useQuery();

  const handleExecute = async () => {
    if (sql.trim() && tab.connectionId) {
      await execute({ connectionId: tab.connectionId, sql });
    }
  };

  // Listen for events from CommandPalette and History
  useEffect(() => {
    const handleRunQuery = () => {
      handleExecute();
    };

    const handleRerun = (e: CustomEvent) => {
      setSql(e.detail);
      setTimeout(() => handleExecute(), 0);
    };

    const handleFormatSql = () => {
      setRightPanelOpen(true);
      setRightPanelTab('formatter');
    };

    window.addEventListener('tablepro:run-query', handleRunQuery);
    window.addEventListener('tablepro:rerun-query', handleRerun as EventListener);
    window.addEventListener('tablepro:format-sql', handleFormatSql);

    return () => {
      window.removeEventListener('tablepro:run-query', handleRunQuery);
      window.removeEventListener('tablepro:rerun-query', handleRerun as EventListener);
      window.removeEventListener('tablepro:format-sql', handleFormatSql);
    };
  }, [sql, tab.connectionId]);

  return (
    <div className="flex flex-col h-full">
      {/* Toolbar */}
      <div className="flex items-center gap-2 p-2 border-b bg-card">
        <Button
          size="sm"
          onClick={handleExecute}
          disabled={!sql.trim() || isExecuting}
          className="gap-1.5"
        >
          {isExecuting ? (
            <Loader2 className="w-4 h-4 animate-spin" />
          ) : (
            <Play className="w-4 h-4" />
          )}
          {isExecuting ? 'Running...' : 'Execute'}
        </Button>
        <span className="text-xs text-muted-foreground ml-2">
          Ctrl+Enter to run
        </span>
      </div>

      {/* Editor + Results split */}
      <div className="flex-1 flex flex-col min-h-0">
        {/* SQL Editor */}
        <div className="flex-1 min-h-0 p-2">
          <SqlEditor
            value={sql}
            onChange={setSql}
            onExecute={handleExecute}
          />
        </div>

        {/* Results Panel */}
        <div className="h-[45%] border-t flex flex-col min-h-0">
          {isExecuting ? (
            <div className="flex-1 flex items-center justify-center">
              <Loader2 className="w-8 h-8 animate-spin text-primary" />
              <span className="ml-2 text-muted-foreground">Executing query...</span>
            </div>
          ) : error ? (
            <div className="flex-1 p-4">
              <div className="bg-destructive/10 border border-destructive/30 rounded-lg p-4">
                <h4 className="font-semibold text-destructive mb-1">Query Error</h4>
                <p className="text-sm text-destructive/80 font-mono">{String(error)}</p>
              </div>
            </div>
          ) : result ? (
            <div className="flex-1 flex flex-col min-h-0 overflow-hidden">
              {/* Results Header */}
              <div className="flex items-center justify-between px-4 py-2 bg-muted/30 border-b">
                <div className="flex items-center gap-2">
                  <Table2 className="w-4 h-4 text-muted-foreground" />
                  <span className="text-sm font-medium">
                    {result.rowCount} rows
                  </span>
                </div>
                <div className="flex items-center gap-3 text-xs text-muted-foreground">
                  <span className="flex items-center gap-1">
                    <span className="w-2 h-2 rounded-full bg-success" />
                    {result.executionTimeMs}ms
                  </span>
                </div>
              </div>

              {/* Results Table */}
              <div className="flex-1 overflow-auto">
                <table className="w-full text-sm">
                  <thead className="sticky top-0 bg-muted">
                    <tr>
                      <th className="w-10 px-2 py-1.5 text-left text-muted-foreground font-medium text-xs">
                        #
                      </th>
                      {result.columns.map((col) => (
                        <th
                          key={col.name}
                          className="px-3 py-1.5 text-left text-muted-foreground font-medium text-xs border-b border-border"
                        >
                          <div className="flex flex-col">
                            <span className="font-semibold">{col.name}</span>
                            <span className="text-[10px] opacity-70 font-normal">
                              {col.type || 'unknown'}
                            </span>
                          </div>
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {result.rows.map((row, i) => (
                      <tr
                        key={i}
                        className="hover:bg-muted/30 transition-colors"
                      >
                        <td className="px-2 py-1.5 text-muted-foreground text-xs font-mono">
                          {i + 1}
                        </td>
                        {result.columns.map((col) => (
                          <td
                            key={col.name}
                            className="px-3 py-1.5 font-mono text-xs border-b border-border/50"
                          >
                            {row[col.name] === null ? (
                              <span className="text-muted-foreground italic">NULL</span>
                            ) : (
                              String(row[col.name])
                            )}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          ) : (
            <div className="flex-1 flex flex-col items-center justify-center text-muted-foreground">
              <Table2 className="w-12 h-12 mb-3 opacity-30" />
              <p className="text-sm">Execute a query to see results</p>
              <p className="text-xs mt-1 opacity-70">
                Press Ctrl+Enter or click Execute
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
