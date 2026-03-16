import { useState, useEffect } from 'react';
import {
  History,
  AlignLeft,
  Copy,
  Check,
  Search,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useAppStore } from '@/stores/appStore';

interface QueryHistoryItem {
  id: string;
  sql: string;
  executedAt: string;
  duration: number;
  rowCount: number;
  connectionId: string;
}

export function RightPanel() {
  const rightPanelOpen = useAppStore((state) => state.rightPanelOpen);
  const rightPanelTab = useAppStore((state) => state.rightPanelTab);
  const setRightPanelTab = useAppStore((state) => state.setRightPanelTab);

  if (!rightPanelOpen) return null;

  return (
    <aside className="w-72 bg-card border-l flex flex-col">
      {/* Tabs */}
      <div className="flex border-b">
        <button
          onClick={() => setRightPanelTab('history')}
          className={`flex-1 flex items-center justify-center gap-2 px-3 py-2.5 text-sm font-medium transition-colors cursor-pointer ${
            rightPanelTab === 'history'
              ? 'text-primary border-b-2 border-primary bg-muted/30'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          <History className="w-4 h-4" />
          History
        </button>
        <button
          onClick={() => setRightPanelTab('formatter')}
          className={`flex-1 flex items-center justify-center gap-2 px-3 py-2.5 text-sm font-medium transition-colors cursor-pointer ${
            rightPanelTab === 'formatter'
              ? 'text-primary border-b-2 border-primary bg-muted/30'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          <AlignLeft className="w-4 h-4" />
          Formatter
        </button>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-hidden">
        {rightPanelTab === 'history' ? (
          <QueryHistory />
        ) : (
          <SqlFormatter />
        )}
      </div>
    </aside>
  );
}

function QueryHistory() {
  const [history, setHistory] = useState<QueryHistoryItem[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const activeConnectionId = useAppStore((state) => state.activeConnectionId);
  const addTab = useAppStore((state) => state.addTab);
  const setActiveTab = useAppStore((state) => state.setActiveTab);
  const updateTab = useAppStore((state) => state.updateTab);

  // Load history from localStorage
  useEffect(() => {
    const saved = localStorage.getItem('tablepro-query-history');
    if (saved) {
      try {
        setHistory(JSON.parse(saved));
      } catch (e) {
        console.error('Failed to parse history:', e);
      }
    }
  }, []);

  const filteredHistory = searchQuery
    ? history.filter(item => item.sql.toLowerCase().includes(searchQuery.toLowerCase()))
    : history;

  const handleCopy = async (sql: string, id: string) => {
    await navigator.clipboard.writeText(sql);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const handleRerun = (sql: string, connectionId?: string) => {
    // Use provided connectionId or fallback to active connection
    const connId = connectionId || activeConnectionId;
    if (!connId) {
      alert('No active connection. Please connect to a database first.');
      return;
    }

    // Create new tab
    const newTabId = crypto.randomUUID();
    const newTab = {
      id: newTabId,
      type: 'query' as const,
      title: `Query ${new Date().toLocaleTimeString()}`,
      connectionId: connId,
    };
    addTab(newTab);
    setActiveTab(newTabId);

    // Update the tab's SQL and trigger execute after a short delay
    setTimeout(() => {
      updateTab(newTabId, { sql } as any);
      window.dispatchEvent(new CustomEvent('tablepro:rerun-query', { detail: sql }));
    }, 100);
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const minutes = Math.floor(diff / 60000);

    if (minutes < 1) return 'Just now';
    if (minutes < 60) return `${minutes}m ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}h ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="flex flex-col h-full">
      {/* Search */}
      <div className="p-2 border-b">
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search queries..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-9 pr-3 py-1.5 text-sm bg-muted border rounded-md
                       focus:outline-none focus:ring-2 focus:ring-primary/50
                       placeholder:text-muted-foreground"
          />
        </div>
      </div>

      {/* List */}
      <div className="flex-1 overflow-auto">
        {filteredHistory.length === 0 ? (
          <div className="p-4 text-center text-muted-foreground text-sm">
            {searchQuery ? 'No matching queries' : 'No query history'}
          </div>
        ) : (
          <div className="p-2 space-y-1">
            {filteredHistory.map((item) => (
              <div
                key={item.id}
                className="group p-2 rounded-md hover:bg-accent cursor-pointer transition-colors"
                onClick={() => handleRerun(item.sql, item.connectionId)}
              >
                <div className="flex items-start gap-2">
                  <div className="flex-1 min-w-0">
                    <p className="text-xs font-mono text-foreground line-clamp-2 break-all">
                      {item.sql}
                    </p>
                    <div className="flex items-center gap-2 mt-1 text-xs text-muted-foreground">
                      <span>{formatDate(item.executedAt)}</span>
                      <span>·</span>
                      <span>{item.duration}ms</span>
                      <span>·</span>
                      <span>{item.rowCount} rows</span>
                    </div>
                  </div>
                  <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 w-6 p-0"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleCopy(item.sql, item.id);
                      }}
                      title="Copy"
                    >
                      {copiedId === item.id ? (
                        <Check className="w-3 h-3 text-success" />
                      ) : (
                        <Copy className="w-3 h-3" />
                      )}
                    </Button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function SqlFormatter() {
  const [input, setInput] = useState('');
  const [output, setOutput] = useState('');
  const [copied, setCopied] = useState(false);

  const handleFormat = () => {
    // Basic SQL formatter - capitalizes keywords
    const keywords = [
      'SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'JOIN', 'LEFT', 'RIGHT',
      'INNER', 'OUTER', 'ON', 'GROUP BY', 'ORDER BY', 'HAVING',
      'LIMIT', 'OFFSET', 'INSERT', 'INTO', 'VALUES', 'UPDATE', 'SET',
      'DELETE', 'CREATE', 'TABLE', 'ALTER', 'DROP', 'INDEX', 'VIEW',
      'AS', 'DISTINCT', 'COUNT', 'SUM', 'AVG', 'MAX', 'MIN', 'NULL',
      'NOT', 'IN', 'BETWEEN', 'LIKE', 'IS', 'ASC', 'DESC', 'UNION',
      'ALL', 'EXISTS', 'CASE', 'WHEN', 'THEN', 'ELSE', 'END'
    ];

    let formatted = input
      // Normalize whitespace
      .replace(/\s+/g, ' ')
      .trim();

    // Capitalize keywords
    keywords.forEach(keyword => {
      const regex = new RegExp(`\\b${keyword}\\b`, 'gi');
      formatted = formatted.replace(regex, keyword);
    });

    // Add newlines before major clauses
    const clauses = ['SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'JOIN',
                     'GROUP BY', 'ORDER BY', 'HAVING', 'LIMIT'];
    clauses.forEach(clause => {
      const regex = new RegExp(`\\b${clause}\\b`, 'gi');
      formatted = formatted.replace(regex, `\n${clause}`);
    });

    // Clean up extra whitespace
    formatted = formatted
      .replace(/\n\s+\n/g, '\n')
      .replace(/^\n/, '')
      .trim();

    setOutput(formatted);
  };

  const handleCopy = async () => {
    await navigator.clipboard.writeText(output);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleClear = () => {
    setInput('');
    setOutput('');
  };

  return (
    <div className="flex flex-col h-full p-2 gap-2">
      {/* Input */}
      <div className="flex-1 flex flex-col min-h-0">
        <label className="text-xs text-muted-foreground mb-1">Input</label>
        <textarea
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Enter SQL to format..."
          className="flex-1 min-h-[100px] p-2 text-xs font-mono bg-muted border rounded-md
                     focus:outline-none focus:ring-2 focus:ring-primary/50
                     placeholder:text-muted-foreground resize-none"
        />
      </div>

      {/* Actions */}
      <div className="flex gap-2">
        <Button size="sm" onClick={handleFormat} className="flex-1">
          Format
        </Button>
        <Button size="sm" variant="outline" onClick={handleClear}>
          Clear
        </Button>
      </div>

      {/* Output */}
      <div className="flex-1 flex flex-col min-h-0">
        <div className="flex items-center justify-between mb-1">
          <label className="text-xs text-muted-foreground">Output</label>
          {output && (
            <Button
              variant="ghost"
              size="sm"
              className="h-6 px-2"
              onClick={handleCopy}
            >
              {copied ? (
                <Check className="w-3 h-3 text-success mr-1" />
              ) : (
                <Copy className="w-3 h-3 mr-1" />
              )}
              {copied ? 'Copied' : 'Copy'}
            </Button>
          )}
        </div>
        <textarea
          value={output}
          readOnly
          placeholder="Formatted SQL will appear here..."
          className="flex-1 min-h-[100px] p-2 text-xs font-mono bg-muted border rounded-md
                     placeholder:text-muted-foreground resize-none"
        />
      </div>
    </div>
  );
}
