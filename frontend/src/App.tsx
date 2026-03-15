import { useState, useEffect, useCallback } from 'react';
import {
  Search,
  ChevronDown,
  Database,
  Table2,
  Columns,
  History,
  Settings,
  PanelLeftClose,
  PanelLeftOpen,
  PanelRightClose,
  PanelRightOpen,
  X,
  AlertCircle,
  FileCode2,
  Layers,
  Plus,
  Play,
  Undo2,
  Redo2,
  Check,
  Clock,
  Rows,
  HardDrive,
  Loader2,
  Key,
  Maximize2,
} from 'lucide-react';
import { useConnectionStore } from './stores/connectionStore';
import { useChangeStore } from './stores/changeStore';
import ConnectionForm from './components/ConnectionForm';
import ConnectionList from './components/ConnectionList';
import QueryEditor from './components/QueryEditor';
import DataGrid from './components/DataGrid/DataGrid';
import PaginationBar from './components/PaginationBar';
import ChangePanel from './components/ChangePanel';
import HistoryPanel from './components/HistoryPanel';
import SessionStatusToast from './components/SessionStatusToast';
import { ConnectionStatusIndicator } from './components/ConnectionStatusIndicator';
import type { DatabaseConnection } from './types';
import type { ConnectionFormData } from './lib/connectionSchema';

function KeyboardShortcut({ keys, label }: { keys: string[]; label: string }) {
  return (
    <div className="flex items-center gap-2 text-xs text-slate-400">
      <span>{label}</span>
      <div className="flex items-center gap-1">
        {keys.map((key, i) => (
          <span key={i}>
            <kbd className="px-1.5 py-0.5 bg-slate-700 border border-slate-600 rounded text-[10px] font-mono text-slate-300">
              {key}
            </kbd>
            {i < keys.length - 1 && <span className="mx-0.5 text-slate-500">+</span>}
          </span>
        ))}
      </div>
    </div>
  );
}

function LoadingSkeleton({ className = '' }: { className?: string }) {
  return (
    <div className={`animate-pulse ${className}`}>
      <div className="h-full bg-slate-700/50 rounded" />
    </div>
  );
}

interface ToastProps {
  message: string;
  type?: 'error' | 'warning' | 'info' | 'success';
  onClose: () => void;
}

function Toast({ message, type = 'error', onClose }: ToastProps) {
  const typeStyles = {
    error: 'bg-red-900/90 border-red-700 text-red-100',
    warning: 'bg-amber-900/90 border-amber-700 text-amber-100',
    info: 'bg-blue-900/90 border-blue-700 text-blue-100',
    success: 'bg-green-900/90 border-green-700 text-green-100',
  };

  const icons = {
    error: <AlertCircle className="w-4 h-4" />,
    warning: <AlertCircle className="w-4 h-4" />,
    info: <AlertCircle className="w-4 h-4" />,
    success: <Check className="w-4 h-4" />,
  };

  return (
    <div className={`flex items-start gap-3 p-4 rounded-lg border ${typeStyles[type]} backdrop-blur-sm shadow-lg`}>
      {icons[type]}
      <p className="text-sm flex-1">{message}</p>
      <button onClick={onClose} className="text-current hover:opacity-70 transition-opacity cursor-pointer">
        <X className="w-4 h-4" />
      </button>
    </div>
  );
}

// Main App Component
export default function App() {
  const {
    connections,
    sessions,
    loadConnections,
    saveConnection,
  } = useConnectionStore();

  const {
    hasChanges,
    discardAllChanges,
  } = useChangeStore();

  const [showConnectionPanel, setShowConnectionPanel] = useState(false);
  const [showHistory, setShowHistory] = useState(false);
  const [activeConnectionId, setActiveConnectionId] = useState<string | null>(null);
  const [activeTabId, setActiveTabId] = useState<string | null>(null);
  const [tabs, setTabs] = useState<Array<{id: string; name: string; query: string; isDirty?: boolean}>>([]);
  
  // UI State
  const [leftSidebarCollapsed, setLeftSidebarCollapsed] = useState(false);
  const [rightSidebarCollapsed, setRightSidebarCollapsed] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [toasts, setToasts] = useState<Array<{id: string; message: string; type: 'error' | 'warning' | 'info' | 'success'}>>([]);
  
  // Change tracking state (derived from store)

  // Panel resizing
  const [editorHeight, setEditorHeight] = useState(40); // percentage
  const [isResizing, setIsResizing] = useState(false);

  const addToast = useCallback((message: string, type: 'error' | 'warning' | 'info' | 'success' = 'info') => {
    const id = Date.now().toString();
    setToasts(prev => [...prev, { id, message, type }]);
    setTimeout(() => {
      setToasts(prev => prev.filter(t => t.id !== id));
    }, 5000);
  }, []);

  const removeToast = useCallback((id: string) => {
    setToasts(prev => prev.filter(t => t.id !== id));
  }, []);

  // Load connections on mount
  useEffect(() => {
    loadConnections();
  }, [loadConnections]);

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ctrl/Cmd+Enter: Execute
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        e.preventDefault();
        handleExecute();
      }
      // Ctrl/Cmd+Z: Undo
      if ((e.ctrlKey || e.metaKey) && e.key === 'z' && !e.shiftKey) {
        e.preventDefault();
        handleUndo();
      }
      // Ctrl/Cmd+Y or Ctrl/Cmd+Shift+Z: Redo
      if ((e.ctrlKey || e.metaKey) && (e.key === 'y' || (e.shiftKey && e.key === 'Z'))) {
        e.preventDefault();
        handleRedo();
      }
      // Ctrl/Cmd+S: Save/Commit
      if ((e.ctrlKey || e.metaKey) && e.key === 's') {
        e.preventDefault();
        handleCommit();
      }
      // Escape: Clear error
      if (e.key === 'Escape' && error) {
        setError(null);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [error]);

  const handleExecute = useCallback(() => {
    setIsLoading(true);
    setError(null);
    // Simulate query execution - replace with actual ExecutePaginated call
    setTimeout(() => {
      setIsLoading(false);
      addToast('Query executed successfully', 'success');
    }, 1000);
  }, [addToast]);

  const handleUndo = useCallback(() => {
    // Undo logic - replace with actual undo call
    addToast('Undo performed', 'info');
  }, [addToast]);

  const handleRedo = useCallback(() => {
    // Redo logic - replace with actual redo call
    addToast('Redo performed', 'info');
  }, [addToast]);

  const handleDiscard = useCallback(() => {
    discardAllChanges(activeTabId || '');
    addToast('All changes discarded', 'warning');
  }, [activeTabId, discardAllChanges, addToast]);

  const handleCommit = useCallback(async () => {
    // Commit changes - replace with actual commit call
    addToast('Changes committed successfully', 'success');
  }, [addToast]);

  // Panel resizer handlers
  const startResizing = useCallback(() => {
    setIsResizing(true);
  }, []);

  const stopResizing = useCallback(() => {
    setIsResizing(false);
  }, []);

  const resize = useCallback((e: MouseEvent) => {
    if (isResizing) {
      const container = document.getElementById('main-container');
      if (container) {
        const rect = container.getBoundingClientRect();
        const newHeight = ((e.clientY - rect.top) / rect.height) * 100;
        if (newHeight > 20 && newHeight < 80) {
          setEditorHeight(newHeight);
        }
      }
    }
  }, [isResizing]);

  useEffect(() => {
    window.addEventListener('mousemove', resize);
    window.addEventListener('mouseup', stopResizing);
    return () => {
      window.removeEventListener('mousemove', resize);
      window.removeEventListener('mouseup', stopResizing);
    };
  }, [resize, stopResizing]);

  const activeResults: { columns: any[]; rows: any[] } | null = null;

  return (
    <div className="flex flex-col h-screen bg-[#0F172A] text-slate-100 overflow-hidden">
      {/* Top Banner */}
      <header className="h-14 bg-[#1E293B] border-b border-slate-700 flex items-center justify-between px-4 shrink-0">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Database className="w-6 h-6 text-green-500" />
            <h1 className="text-lg font-bold text-white">TablePro</h1>
          </div>
          <ConnectionStatusIndicator connectionId={activeConnectionId || undefined} compact />
        </div>
        
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-3">
            <KeyboardShortcut keys={['⌘', '↵']} label="" />
            <KeyboardShortcut keys={['⌘', 'Z']} label="" />
            <KeyboardShortcut keys={['⌘', 'Y']} label="" />
            <KeyboardShortcut keys={['⌘', 'S']} label="" />
          </div>
          <div className="w-px h-6 bg-slate-600" />
          <button className="p-2 hover:bg-slate-700 rounded transition-colors cursor-pointer" title="Settings">
            <Settings className="w-5 h-5" />
          </button>
          <button className="p-2 hover:bg-slate-700 rounded transition-colors cursor-pointer" title="Fullscreen">
            <Maximize2 className="w-5 h-5" />
          </button>
        </div>
      </header>

      {/* Main Content */}
      <div id="main-container" className="flex-1 flex overflow-hidden">
        {/* Left Sidebar - Connections */}
        {!leftSidebarCollapsed && (
          <aside className="w-72 bg-[#1E293B] border-r border-slate-700 flex flex-col shrink-0">
            <div className="p-4 border-b border-slate-700 flex items-center justify-between">
              <h2 className="text-sm font-semibold text-slate-300 flex items-center gap-2">
                <Database className="w-4 h-4" />
                Connections
              </h2>
              <button 
                onClick={() => setShowConnectionPanel(true)}
                className="p-1.5 hover:bg-slate-700 rounded transition-colors cursor-pointer"
                title="New Connection"
              >
                <Plus className="w-4 h-4" />
              </button>
            </div>
            
            <div className="flex-1 overflow-y-auto p-2">
              {connections.length === 0 ? (
                <div className="text-center py-8">
                  <Database className="w-12 h-12 mx-auto text-slate-600 mb-3" />
                  <p className="text-sm text-slate-400 mb-4">No connections yet</p>
                  <button 
                    onClick={() => setShowConnectionPanel(true)}
                    className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white text-sm rounded transition-colors cursor-pointer"
                  >
                    Create your first connection
                  </button>
                </div>
              ) : (
                <ConnectionList
                  connections={connections}
                  sessions={sessions}
                  onSelect={(conn) => setActiveConnectionId(conn.id)}
                  onEdit={() => {}}
                  onDelete={() => {}}
                  onDuplicate={() => {}}
                  onConnect={() => {}}
                  onDisconnect={() => {}}
                  onNewConnection={() => setShowConnectionPanel(true)}
                />
              )}
            </div>
          </aside>
        )}

        {/* Left Sidebar Toggle */}
        <button
          onClick={() => setLeftSidebarCollapsed(!leftSidebarCollapsed)}
          className="w-6 bg-[#1E293B] border-r border-slate-700 flex items-center justify-center hover:bg-slate-700 transition-colors cursor-pointer"
          title={leftSidebarCollapsed ? 'Show sidebar' : 'Hide sidebar'}
        >
          {leftSidebarCollapsed ? <PanelLeftOpen className="w-4 h-4" /> : <PanelLeftClose className="w-4 h-4" />}
        </button>

        {/* Main Editor + Grid Area */}
        <main className="flex-1 flex flex-col overflow-hidden bg-[#0F172A]">
          {/* Tab Bar */}
          <div className="h-10 bg-[#1E293B] border-b border-slate-700 flex items-center px-2 shrink-0">
            <div className="flex items-center gap-1 overflow-x-auto flex-1">
              {tabs.length === 0 ? (
                <div className="text-sm text-slate-400 px-3">No open tabs</div>
              ) : (
                tabs.map(tab => (
                  <div
                    key={tab.id}
                    className={`flex items-center gap-2 px-3 py-1.5 rounded-t text-sm cursor-pointer transition-colors ${
                      activeTabId === tab.id
                        ? 'bg-[#0F172A] text-white border-t-2 border-green-500'
                        : 'text-slate-400 hover:bg-slate-700'
                    }`}
                    onClick={() => setActiveTabId(tab.id)}
                  >
                    <FileCode2 className="w-4 h-4 shrink-0" />
                    <span className="truncate max-w-[150px]">{tab.name}</span>
                    {tab.isDirty && <div className="w-2 h-2 rounded-full bg-amber-500" />}
                    <button 
                      className="hover:bg-slate-600 rounded p-0.5 cursor-pointer"
                      onClick={(e) => { e.stopPropagation(); }}
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </div>
                ))
              )}
            </div>
            <button 
              className="p-1.5 hover:bg-slate-700 rounded transition-colors cursor-pointer ml-2"
              title="New query tab"
              onClick={() => setTabs([...tabs, { id: Date.now().toString(), name: `Query ${tabs.length + 1}`, query: '' }])}
            >
              <Plus className="w-4 h-4" />
            </button>
          </div>

          {/* Toolbar */}
          <div className="h-12 bg-[#1E293B] border-b border-slate-700 flex items-center justify-between px-4 shrink-0">
            <div className="flex items-center gap-2">
              <button
                onClick={handleExecute}
                disabled={isLoading || !activeTabId}
                className="flex items-center gap-2 px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-slate-600 disabled:cursor-not-allowed text-white text-sm rounded transition-colors cursor-pointer"
              >
                <Play className="w-4 h-4" />
                Run
              </button>
              <button className="p-2 hover:bg-slate-700 rounded transition-colors cursor-pointer" title="Format query">
                <FileCode2 className="w-4 h-4" />
              </button>
              <div className="w-px h-6 bg-slate-600 mx-2" />
              <button
                onClick={handleUndo}
                className="p-2 hover:bg-slate-700 rounded transition-colors cursor-pointer"
                title="Undo"
              >
                <Undo2 className="w-4 h-4" />
              </button>
              <button
                onClick={handleRedo}
                className="p-2 hover:bg-slate-700 rounded transition-colors cursor-pointer"
                title="Redo"
              >
                <Redo2 className="w-4 h-4" />
              </button>
            </div>
            
            {isLoading && (
              <div className="flex items-center gap-2 text-sm text-slate-400">
                <Loader2 className="w-4 h-4 animate-spin" />
                Executing...
              </div>
            )}
          </div>

          {/* Query Editor */}
          <div style={{ height: `${editorHeight}%` }} className="border-b border-slate-700 overflow-hidden shrink-0">
            <QueryEditor
              connectionId={activeConnectionId || undefined}
              onExecute={() => handleExecute()}
            />
          </div>

          {/* Resizer Handle */}
          <div
            onMouseDown={startResizing}
            className="h-1 bg-slate-600 hover:bg-green-500 cursor-row-resize transition-colors shrink-0"
          />

          {/* Data Grid Results */}
          <div className="flex-1 overflow-hidden flex flex-col">
            {isLoading ? (
              <LoadingSkeleton className="m-4" />
            ) : activeResults?.rows ? (
              <>
                <div className="flex-1 overflow-auto">
                  <DataGrid
                    columns={activeResults.columns || []}
                    rowData={activeResults.rows || []}
                  />
                </div>
                <PaginationBar
                  page={1}
                  pageSize={100}
                  totalCount={1000}
                  isExact={false}
                  onPageChange={() => {}}
                  onPageSizeChange={() => {}}
                />
              </>
            ) : (
              <div className="flex-1 flex items-center justify-center">
                <div className="text-center text-slate-400">
                  <Table2 className="w-16 h-16 mx-auto mb-4 opacity-50" />
                  <p className="text-lg mb-2">No results to display</p>
                  <p className="text-sm">Execute a query to see results</p>
                </div>
              </div>
            )}
          </div>

          {/* Change Tracking Panel */}
          {activeTabId && hasChanges(activeTabId) && (
            <div className="h-48 bg-[#1E293B] border-t-2 border-amber-500 shrink-0">
              <ChangePanel
                tabId={activeTabId}
                columns={[]}
                tableName=""
                databaseType="postgres"
                onCommit={handleCommit}
                onDiscard={handleDiscard}
              />
            </div>
          )}
        </main>

        {/* Right Sidebar - Schema Browser */}
        {!rightSidebarCollapsed && (
          <aside className="w-80 bg-[#1E293B] border-l border-slate-700 flex flex-col shrink-0">
            <div className="p-4 border-b border-slate-700 flex items-center justify-between">
              <h2 className="text-sm font-semibold text-slate-300 flex items-center gap-2">
                <Layers className="w-4 h-4" />
                Schema
              </h2>
              <button 
                onClick={() => setRightSidebarCollapsed(true)}
                className="p-1.5 hover:bg-slate-700 rounded transition-colors cursor-pointer"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
            
            <div className="p-2 border-b border-slate-700">
              <div className="relative">
                <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                <input
                  type="text"
                  placeholder="Search tables..."
                  className="w-full pl-9 pr-3 py-2 bg-slate-800 border border-slate-700 rounded text-sm text-slate-200 placeholder-slate-500 focus:outline-none focus:border-green-600"
                />
              </div>
            </div>
            
            <div className="flex-1 overflow-y-auto p-2">
              {activeConnectionId ? (
                <div className="space-y-1">
                  <div className="px-2 py-1.5 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                    Tables
                  </div>
                  <div className="space-y-1">
                    <div className="px-2 py-1.5 text-sm text-slate-300 hover:bg-slate-700 rounded cursor-pointer flex items-center gap-2">
                      <ChevronDown className="w-4 h-4" />
                      <Table2 className="w-4 h-4 text-blue-400" />
                      <span>users</span>
                      <span className="text-xs text-slate-500 ml-auto">(4)</span>
                    </div>
                    <div className="pl-8 space-y-1">
                      <div className="px-2 py-1 text-xs text-slate-400 hover:bg-slate-700 rounded cursor-pointer flex items-center gap-2">
                        <Key className="w-3 h-3 text-amber-400" />
                        <span>id</span>
                        <span className="text-xs text-slate-600 ml-auto">integer</span>
                      </div>
                      <div className="px-2 py-1 text-xs text-slate-400 hover:bg-slate-700 rounded cursor-pointer flex items-center gap-2">
                        <Columns className="w-3 h-3 text-green-400" />
                        <span>email</span>
                        <span className="text-xs text-slate-600 ml-auto">varchar(255)</span>
                      </div>
                      <div className="px-2 py-1 text-xs text-slate-400 hover:bg-slate-700 rounded cursor-pointer flex items-center gap-2">
                        <Columns className="w-3 h-3 text-green-400" />
                        <span>name</span>
                        <span className="text-xs text-slate-600 ml-auto">varchar(100)</span>
                      </div>
                      <div className="px-2 py-1 text-xs text-slate-400 hover:bg-slate-700 rounded cursor-pointer flex items-center gap-2">
                        <Columns className="w-3 h-3 text-green-400" />
                        <span>created_at</span>
                        <span className="text-xs text-slate-600 ml-auto">timestamp</span>
                      </div>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="text-center py-8 text-slate-400">
                  <Database className="w-12 h-12 mx-auto mb-3 opacity-50" />
                  <p className="text-sm">Connect to see schema</p>
                </div>
              )}
            </div>
          </aside>
        )}

        {/* Right Sidebar Toggle */}
        <button
          onClick={() => setRightSidebarCollapsed(!rightSidebarCollapsed)}
          className="w-6 bg-[#1E293B] border-l border-slate-700 flex items-center justify-center hover:bg-slate-700 transition-colors cursor-pointer"
          title={rightSidebarCollapsed ? 'Show schema' : 'Hide schema'}
        >
          {rightSidebarCollapsed ? <PanelRightOpen className="w-4 h-4" /> : <PanelRightClose className="w-4 h-4" />}
        </button>
      </div>

      {/* Status Bar */}
      <footer className="h-6 bg-[#1E293B] border-t border-slate-700 flex items-center justify-between px-4 text-xs text-slate-400 shrink-0">
        <div className="flex items-center gap-4">
          <span className="flex items-center gap-1">
            <Clock className="w-3 h-3" />
            Ready
          </span>
          {activeResults?.rows && (
            <>
              <span className="flex items-center gap-1">
                <Rows className="w-3 h-3" />
                {activeResults.rows.length} rows
              </span>
              <span className="flex items-center gap-1">
                <HardDrive className="w-3 h-3" />
                ~1,000 total
              </span>
            </>
          )}
        </div>
        <div className="flex items-center gap-2">
          <span>Fira Code</span>
          <span>UTF-8</span>
          <span>SQL</span>
        </div>
      </footer>

      {/* Toast Notifications */}
      <div className="fixed bottom-8 right-4 z-50 space-y-2">
        {toasts.map(toast => (
          <Toast key={toast.id} message={toast.message} type={toast.type} onClose={() => removeToast(toast.id)} />
        ))}
      </div>

      {/* Error Display */}
      {error && (
        <div className="fixed bottom-8 left-1/2 -translate-x-1/2 z-50 max-w-2xl">
          <div className="bg-red-900/90 border border-red-700 text-red-100 px-4 py-3 rounded-lg backdrop-blur-sm shadow-lg flex items-start gap-3">
            <AlertCircle className="w-5 h-5 shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="font-semibold mb-1">Query Error</p>
              <p className="text-sm">{error}</p>
            </div>
            <button onClick={() => setError(null)} className="text-red-300 hover:text-red-100 transition-colors cursor-pointer">
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>
      )}

      {/* Connection Form Modal */}
      {showConnectionPanel && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center">
          <div className="bg-[#1E293B] rounded-lg shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-auto border border-slate-700">
            <div className="p-6 border-b border-slate-700 flex items-center justify-between sticky top-0 bg-[#1E293B] z-10">
              <h2 className="text-xl font-bold text-white">New Connection</h2>
              <button 
                onClick={() => setShowConnectionPanel(false)}
                className="p-2 hover:bg-slate-700 rounded transition-colors cursor-pointer"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <ConnectionForm
              onSave={async (data: ConnectionFormData) => {
                try {
                  const connection: DatabaseConnection = {
                    ...data,
                    id: data.id || '',
                  };
                  await saveConnection(connection);
                  setShowConnectionPanel(false);
                  addToast('Connection saved', 'success');
                } catch (err) {
                  addToast('Failed to save connection', 'error');
                }
              }}
              onCancel={() => setShowConnectionPanel(false)}
            />
          </div>
        </div>
      )}

      {/* History Panel */}
      {showHistory && (
        <div className="fixed inset-y-0 right-0 w-96 bg-[#1E293B] shadow-2xl z-50 border-l border-slate-700">
          <div className="h-full flex flex-col">
            <div className="p-4 border-b border-slate-700 flex items-center justify-between">
              <h2 className="text-lg font-bold text-white flex items-center gap-2">
                <History className="w-5 h-5" />
                Query History
              </h2>
              <button 
                onClick={() => setShowHistory(false)}
                className="p-2 hover:bg-slate-700 rounded transition-colors cursor-pointer"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <HistoryPanel
              connectionId={activeConnectionId}
              entries={[]}
              onLoadQuery={() => {}}
              onClearHistory={() => {}}
            />
          </div>
        </div>
      )}

      <SessionStatusToast />
    </div>
  );
}
