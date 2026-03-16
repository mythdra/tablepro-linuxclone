import { useState, useEffect, useRef, useMemo } from 'react';
import {
  Search,
  Plug,
  Plus,
  Play,
  AlignLeft,
  Settings,
  PanelLeftClose,
  PanelRightClose,
  CornerDownLeft,
} from 'lucide-react';
import { useAppStore } from '@/stores/appStore';
import { cn } from '@/lib/utils';

interface Command {
  id: string;
  label: string;
  description?: string;
  icon: React.ReactNode;
  shortcut?: string;
  action: () => void;
  category: 'connection' | 'query' | 'view' | 'settings';
}

export function CommandPalette() {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  const {
    setSidebarCollapsed,
    sidebarCollapsed,
    setRightPanelOpen,
    rightPanelOpen,
    setActiveConnection,
    connections,
  } = useAppStore();

  const commands: Command[] = useMemo(() => [
    // Connection commands
    {
      id: 'new-connection',
      label: 'New Connection',
      description: 'Connect to a new database',
      icon: <Plug className="w-4 h-4" />,
      action: () => {
        setOpen(false);
        window.dispatchEvent(new CustomEvent('tablepro:open-connection-dialog'));
      },
      category: 'connection',
    },
    ...connections.map((conn) => ({
      id: `connect-${conn.id}`,
      label: `Connect to ${conn.name || conn.host}`,
      description: `${conn.host}:${conn.port}/${conn.database}`,
      icon: <Plug className="w-4 h-4" />,
      action: () => {
        setActiveConnection(conn.id);
        setOpen(false);
      },
      category: 'connection' as const,
    })),

    // Query commands
    {
      id: 'new-tab',
      label: 'New Query Tab',
      description: 'Create a new query tab',
      icon: <Plus className="w-4 h-4" />,
      shortcut: 'Ctrl+T',
      action: () => {
        setOpen(false);
        window.dispatchEvent(new CustomEvent('tablepro:new-tab'));
      },
      category: 'query',
    },
    {
      id: 'run-query',
      label: 'Run Current Query',
      description: 'Execute the current query',
      icon: <Play className="w-4 h-4" />,
      shortcut: 'Ctrl+Enter',
      action: () => {
        setOpen(false);
        window.dispatchEvent(new CustomEvent('tablepro:run-query'));
      },
      category: 'query',
    },
    {
      id: 'format-sql',
      label: 'Format SQL',
      description: 'Format current SQL query',
      icon: <AlignLeft className="w-4 h-4" />,
      shortcut: 'Ctrl+Shift+F',
      action: () => {
        setOpen(false);
        setRightPanelOpen(true);
        window.dispatchEvent(new CustomEvent('tablepro:format-sql'));
      },
      category: 'query',
    },

    // View commands
    {
      id: 'toggle-sidebar',
      label: sidebarCollapsed ? 'Show Sidebar' : 'Hide Sidebar',
      description: 'Toggle the sidebar panel',
      icon: <PanelLeftClose className="w-4 h-4" />,
      shortcut: 'Ctrl+B',
      action: () => {
        setSidebarCollapsed(!sidebarCollapsed);
        setOpen(false);
      },
      category: 'view',
    },
    {
      id: 'toggle-right-panel',
      label: rightPanelOpen ? 'Hide Right Panel' : 'Show Right Panel',
      description: 'Toggle the right panel (History/Formatter)',
      icon: <PanelRightClose className="w-4 h-4" />,
      shortcut: 'Ctrl+Shift+B',
      action: () => {
        setRightPanelOpen(!rightPanelOpen);
        setOpen(false);
      },
      category: 'view',
    },

    // Settings
    {
      id: 'settings',
      label: 'Open Settings',
      description: 'Open application settings',
      icon: <Settings className="w-4 h-4" />,
      action: () => {
        setOpen(false);
        window.dispatchEvent(new CustomEvent('tablepro:open-settings'));
      },
      category: 'settings',
    },
  ], [connections, sidebarCollapsed, rightPanelOpen, setSidebarCollapsed, setRightPanelOpen, setActiveConnection]);

  // Filter commands based on query
  const filteredCommands = useMemo(() => {
    if (!query.trim()) return commands;
    const lowerQuery = query.toLowerCase();
    return commands.filter(
      (cmd) =>
        cmd.label.toLowerCase().includes(lowerQuery) ||
        cmd.description?.toLowerCase().includes(lowerQuery)
    );
  }, [commands, query]);

  // Keyboard shortcut to open
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        setOpen(true);
        setQuery('');
        setSelectedIndex(0);
      }
      if (e.key === 'Escape' && open) {
        setOpen(false);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [open]);

  // Handle keyboard navigation
  useEffect(() => {
    if (!open) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          setSelectedIndex((i) => (i + 1) % filteredCommands.length);
          break;
        case 'ArrowUp':
          e.preventDefault();
          setSelectedIndex((i) => (i - 1 + filteredCommands.length) % filteredCommands.length);
          break;
        case 'Enter':
          e.preventDefault();
          filteredCommands[selectedIndex]?.action();
          break;
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [open, filteredCommands, selectedIndex]);

  // Scroll selected item into view
  useEffect(() => {
    if (listRef.current) {
      const selected = listRef.current.querySelector(`[data-index="${selectedIndex}"]`);
      selected?.scrollIntoView({ block: 'nearest' });
    }
  }, [selectedIndex]);

  // Focus input when opened
  useEffect(() => {
    if (open) {
      inputRef.current?.focus();
    }
  }, [open]);

  if (!open) return null;

  const categories = ['connection', 'query', 'view', 'settings'] as const;
  const categoryLabels: Record<string, string> = {
    connection: 'Connections',
    query: 'Query',
    view: 'View',
    settings: 'Settings',
  };

  return (
    <div className="fixed inset-0 z-50">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={() => setOpen(false)}
      />

      {/* Dialog */}
      <div className="absolute top-[20%] left-1/2 -translate-x-1/2 w-full max-w-lg">
        <div className="bg-card border rounded-lg shadow-2xl overflow-hidden">
          {/* Search input */}
          <div className="flex items-center gap-2 px-3 py-3 border-b">
            <Search className="w-5 h-5 text-muted-foreground" />
            <input
              ref={inputRef}
              type="text"
              value={query}
              onChange={(e) => {
                setQuery(e.target.value);
                setSelectedIndex(0);
              }}
              placeholder="Type a command..."
              className="flex-1 bg-transparent outline-none text-sm placeholder:text-muted-foreground"
            />
            <kbd className="hidden sm:inline-flex h-5 items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] text-muted-foreground">
              ESC
            </kbd>
          </div>

          {/* Commands list */}
          <div ref={listRef} className="max-h-80 overflow-auto p-2">
            {filteredCommands.length === 0 ? (
              <div className="py-8 text-center text-muted-foreground text-sm">
                No commands found
              </div>
            ) : (
              categories.map((category) => {
                const categoryCommands = filteredCommands.filter(
                  (cmd) => cmd.category === category
                );
                if (categoryCommands.length === 0) return null;

                return (
                  <div key={category} className="mb-2">
                    <div className="px-2 py-1 text-xs font-medium text-muted-foreground">
                      {categoryLabels[category]}
                    </div>
                    {categoryCommands.map((cmd) => {
                      const index = filteredCommands.indexOf(cmd);
                      return (
                        <button
                          key={cmd.id}
                          data-index={index}
                          onClick={() => cmd.action()}
                          onMouseEnter={() => setSelectedIndex(index)}
                          className={cn(
                            'w-full flex items-center gap-3 px-2 py-2 rounded-md text-left transition-colors',
                            selectedIndex === index
                              ? 'bg-primary/10 text-primary'
                              : 'hover:bg-accent'
                          )}
                        >
                          <div className="flex-shrink-0 text-muted-foreground">
                            {cmd.icon}
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="text-sm font-medium">{cmd.label}</div>
                            {cmd.description && (
                              <div className="text-xs text-muted-foreground truncate">
                                {cmd.description}
                              </div>
                            )}
                          </div>
                          {cmd.shortcut && (
                            <kbd className="flex-shrink-0 hidden sm:inline-flex h-5 items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] text-muted-foreground">
                              {cmd.shortcut}
                            </kbd>
                          )}
                        </button>
                      );
                    })}
                  </div>
                );
              })
            )}
          </div>

          {/* Footer */}
          <div className="px-3 py-2 border-t flex items-center gap-4 text-xs text-muted-foreground">
            <span className="flex items-center gap-1">
              <CornerDownLeft className="w-3 h-3" />
              to select
            </span>
            <span className="flex items-center gap-1">
              <kbd className="rounded border bg-muted px-1">↑</kbd>
              <kbd className="rounded border bg-muted px-1">↓</kbd>
              to navigate
            </span>
            <span className="flex items-center gap-1">
              <kbd className="rounded border bg-muted px-1">↵</kbd>
              to execute
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
