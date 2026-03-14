import { useState, useMemo, useCallback, useRef, useEffect } from 'react';
import { Search, Trash2, History, X, AlertTriangle } from 'lucide-react';
import type { QueryHistoryEntry } from '../types';
import { HistoryEntry } from './HistoryEntry';

/**
 * Props for the HistoryPanel component.
 */
interface HistoryPanelProps {
  /** Connection ID to filter history entries */
  connectionId: string;
  /** List of query history entries to display */
  entries: QueryHistoryEntry[];
  /** Callback when a history entry is selected to load its query */
  onLoadQuery: (query: string) => void;
  /** Callback to clear history for a connection */
  onClearHistory: (connectionId: string) => void;
}

/**
 * Debounce hook for delaying value updates.
 * Useful for search inputs to avoid excessive filtering.
 * 
 * @template T - Type of the value being debounced
 * @param value - Current value to debounce
 * @param delay - Delay in milliseconds
 * @returns Debounced value updated after delay
 */
function useDebouncedValue<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(timer);
    };
  }, [value, delay]);

  return debouncedValue;
}

/**
 * HistoryPanel displays query execution history in a sidebar panel.
 * 
 * Features:
 * - Real-time search with debouncing (300ms)
 * - Clear all history with confirmation dialog
 * - Click to load query into editor
 * - Keyboard shortcut (Escape) to cancel clear confirmation
 * - Empty state and no-results state displays
 * 
 * @example
 * ```tsx
 * <HistoryPanel
 *   connectionId="conn-123"
 *   entries={historyEntries}
 *   onLoadQuery={(query) => loadQueryInEditor(query)}
 *   onClearHistory={(id) => clearConnectionHistory(id)}
 * />
 * ```
 */
export function HistoryPanel({
  connectionId,
  entries,
  onLoadQuery,
  onClearHistory,
}: HistoryPanelProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [showClearConfirm, setShowClearConfirm] = useState(false);
  const searchInputRef = useRef<HTMLInputElement>(null);

  // Debounce search at 300ms
  const debouncedSearch = useDebouncedValue(searchQuery, 300);

  // Filter entries by search query
  const filteredEntries = useMemo(() => {
    if (!debouncedSearch.trim()) {
      return entries;
    }
    const searchLower = debouncedSearch.toLowerCase();
    return entries.filter((entry) =>
      entry.query.toLowerCase().includes(searchLower)
    );
  }, [entries, debouncedSearch]);

  // Handle entry click - load query in editor
  const handleEntryClick = useCallback(
    (entry: QueryHistoryEntry) => {
      onLoadQuery(entry.query);
    },
    [onLoadQuery]
  );

  // Handle clear history
  const handleClearHistory = useCallback(() => {
    onClearHistory(connectionId);
    setShowClearConfirm(false);
  }, [connectionId, onClearHistory]);

  // Clear search
  const handleClearSearch = useCallback(() => {
    setSearchQuery('');
    searchInputRef.current?.focus();
  }, []);

  // Keyboard shortcut: Escape to close clear confirm dialog
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && showClearConfirm) {
        setShowClearConfirm(false);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [showClearConfirm]);

  return (
    <div
      className="h-full flex flex-col bg-slate-900"
      role="region"
      aria-label="Query History"
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
        <h2 className="text-lg font-semibold text-white flex items-center gap-2">
          <History className="w-5 h-5 text-slate-400" />
          History
        </h2>
        {entries.length > 0 && (
          <button
            onClick={() => setShowClearConfirm(true)}
            disabled={showClearConfirm}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-slate-400 bg-slate-800 rounded-lg hover:text-red-400 hover:bg-slate-700 disabled:opacity-50 transition-colors"
            aria-label="Clear all history"
          >
            <Trash2 className="w-4 h-4" />
            Clear
          </button>
        )}
      </div>

      {/* Search input */}
      {entries.length > 0 && (
        <div className="px-4 py-3 border-b border-slate-700">
          <div className="relative">
            <Search
              className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500"
              aria-hidden="true"
            />
            <input
              ref={searchInputRef}
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search queries..."
              className="w-full pl-10 pr-8 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white placeholder-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50 focus:border-primary transition-colors"
              aria-label="Search query history"
            />
            {searchQuery && (
              <button
                onClick={handleClearSearch}
                className="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-slate-500 hover:text-white rounded transition-colors"
                aria-label="Clear search"
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </div>
        </div>
      )}

      {/* Clear confirmation dialog */}
      {showClearConfirm && (
        <div
          className="mx-4 my-3 p-4 bg-red-500/10 border border-red-500/30 rounded-lg"
          role="alertdialog"
          aria-labelledby="clear-confirm-title"
        >
          <div className="flex items-start gap-3">
            <AlertTriangle className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <h3
                id="clear-confirm-title"
                className="text-sm font-medium text-white"
              >
                Clear all history?
              </h3>
              <p className="mt-1 text-sm text-slate-400">
                This will remove all query history for this connection. This
                action cannot be undone.
              </p>
              <div className="flex items-center gap-2 mt-3">
                <button
                  onClick={handleClearHistory}
                  className="px-3 py-1.5 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 transition-colors"
                  aria-label="Confirm clear history"
                >
                  Clear History
                </button>
                <button
                  onClick={() => setShowClearConfirm(false)}
                  className="px-3 py-1.5 text-sm font-medium text-slate-300 bg-slate-700 rounded-lg hover:bg-slate-600 transition-colors"
                  aria-label="Cancel clear history"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* History list */}
      <div className="flex-1 overflow-y-auto">
        {entries.length === 0 ? (
          /* Empty state */
          <div
            className="flex flex-col items-center justify-center h-64 text-slate-500"
            role="status"
          >
            <History className="w-12 h-12 mb-3 opacity-50" />
            <p className="text-sm">No query history yet</p>
            <p className="text-xs text-slate-600 mt-1">
              Execute queries to see them here
            </p>
          </div>
        ) : filteredEntries.length === 0 ? (
          /* No search results */
          <div
            className="flex flex-col items-center justify-center h-64 text-slate-500"
            role="status"
          >
            <Search className="w-12 h-12 mb-3 opacity-50" />
            <p className="text-sm">No matching queries found</p>
            <button
              onClick={handleClearSearch}
              className="mt-2 text-primary hover:underline text-sm"
            >
              Clear search
            </button>
          </div>
        ) : (
          /* Entry list */
          <div role="list" aria-label="Query history entries">
            {filteredEntries.map((entry) => (
              <div key={entry.id} role="listitem">
                <HistoryEntry entry={entry} onClick={handleEntryClick} />
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Footer with count */}
      {entries.length > 0 && (
        <div className="px-4 py-2 border-t border-slate-700 text-xs text-slate-500">
          {filteredEntries.length === entries.length
            ? `${entries.length} ${entries.length === 1 ? 'query' : 'queries'}`
            : `${filteredEntries.length} of ${entries.length} queries`}
        </div>
      )}
    </div>
  );
}

export default HistoryPanel;