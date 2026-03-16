import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { useQuery } from '@/hooks/useQuery';
import { Play } from 'lucide-react';
import type { Tab } from '@/types';

export function QueryTab({ tab }: { tab: Tab }) {
  const [sql, setSql] = useState('');

  const { execute, isExecuting, result, error } = useQuery();

  const handleExecute = async () => {
    if (sql.trim() && tab.connectionId) {
      await execute({ connectionId: tab.connectionId, sql });
    }
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2 p-2 border-b">
        <Button
          size="sm"
          onClick={handleExecute}
          disabled={!sql.trim() || isExecuting}
        >
          <Play className="w-4 h-4 mr-1" />
          Execute
        </Button>
        <span className="text-xs text-muted-foreground">
          Ctrl+Enter to execute
        </span>
      </div>
      <div className="flex-1 flex">
        <div className="flex-1 p-4">
          <textarea
            value={sql}
            onChange={(e) => setSql(e.target.value)}
            className="w-full h-full resize-none border rounded p-2 font-mono text-sm"
            placeholder="Enter SQL query..."
          />
        </div>
        <div className="w-1/2 p-4 border-l overflow-auto">
          {isExecuting ? (
            <div className="text-muted-foreground">Loading...</div>
          ) : error ? (
            <div className="text-red-500">{String(error)}</div>
          ) : result ? (
            <div>
              <div className="text-sm text-muted-foreground mb-2">
                {result.rowCount} rows ({result.executionTimeMs}ms)
              </div>
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    {result.columns.map((col) => (
                      <th key={col.name} className="px-2 py-1 text-left">
                        {col.name}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {result.rows.map((row, i) => (
                    <tr key={i} className="border-b">
                      {result.columns.map((col) => (
                        <td key={col.name} className="px-2 py-1">
                          {String(row[col.name] ?? '')}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-muted-foreground">
              Execute a query to see results
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
