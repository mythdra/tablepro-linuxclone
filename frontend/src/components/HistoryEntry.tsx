import { CheckCircle, XCircle, Clock, Hash } from 'lucide-react';
import type { QueryHistoryEntry } from '../types';

interface HistoryEntryProps {
  entry: QueryHistoryEntry;
  onClick: (entry: QueryHistoryEntry) => void;
}

/**
 * Format a duration in milliseconds to a human-readable string
 */
function formatDuration(ms: number): string {
  if (ms < 1000) {
    return `${ms}ms`;
  }
  if (ms < 60000) {
    return `${(ms / 1000).toFixed(1)}s`;
  }
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.round((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}

/**
 * Format a timestamp to a relative time string
 */
function formatRelativeTime(timestamp: string): string {
  const now = new Date();
  const date = new Date(timestamp);
  const diffMs = now.getTime() - date.getTime();
  const diffSec = Math.floor(diffMs / 1000);
  const diffMin = Math.floor(diffSec / 60);
  const diffHour = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHour / 24);

  if (diffSec < 60) {
    return 'just now';
  }
  if (diffMin < 60) {
    return `${diffMin} min ago`;
  }
  if (diffHour < 24) {
    return `${diffHour}h ago`;
  }
  if (diffDay === 1) {
    return 'yesterday';
  }
  if (diffDay < 7) {
    return `${diffDay} days ago`;
  }
  // Fall back to absolute date
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

/**
 * Format timestamp to absolute format: "2026-03-14 10:30:00"
 */
function formatAbsoluteTime(timestamp: string): string {
  const date = new Date(timestamp);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hour = String(date.getHours()).padStart(2, '0');
  const minute = String(date.getMinutes()).padStart(2, '0');
  const second = String(date.getSeconds()).padStart(2, '0');
  return `${year}-${month}-${day} ${hour}:${minute}:${second}`;
}

/**
 * Truncate query text for display
 */
function truncateQuery(query: string, maxLength: number = 80): string {
  const normalized = query.trim().replace(/\s+/g, ' ');
  if (normalized.length <= maxLength) {
    return normalized;
  }
  return normalized.slice(0, maxLength - 3) + '...';
}

export function HistoryEntry({ entry, onClick }: HistoryEntryProps) {
  const isSuccess = entry.status === 'success';
  const isError = entry.status === 'error';

  return (
    <button
      onClick={() => onClick(entry)}
      className="w-full text-left px-4 py-3 hover:bg-slate-800/50 transition-colors group focus:outline-none focus:ring-2 focus:ring-primary/50 focus:ring-inset"
      aria-label={`Query: ${truncateQuery(entry.query, 40)}. Click to load.`}
    >
      <div className="flex items-start gap-3">
        {/* Status icon */}
        <div className="mt-0.5 flex-shrink-0">
          {isSuccess && (
            <CheckCircle
              className="w-4 h-4 text-emerald-500"
              aria-label="Success"
            />
          )}
          {isError && (
            <XCircle
              className="w-4 h-4 text-red-500"
              aria-label="Error"
            />
          )}
          {entry.status === 'cancelled' && (
            <div
              className="w-4 h-4 rounded-full bg-amber-500 flex items-center justify-center"
              aria-label="Cancelled"
            >
              <div className="w-2 h-0.5 bg-white rounded" />
            </div>
          )}
        </div>

        {/* Query text and metadata */}
        <div className="flex-1 min-w-0">
          <p className="text-sm text-slate-200 font-mono truncate group-hover:text-white transition-colors">
            {truncateQuery(entry.query)}
          </p>

          <div className="flex items-center gap-3 mt-1.5 text-xs text-slate-500">
            {/* Timestamp */}
            <span
              className="flex items-center gap-1"
              title={formatAbsoluteTime(entry.executedAt)}
            >
              <Clock className="w-3 h-3" />
              {formatRelativeTime(entry.executedAt)}
            </span>

            {/* Duration */}
            <span className="flex items-center gap-1">
              <div className="w-3 h-3 flex items-center justify-center">
                •
              </div>
              {formatDuration(entry.duration)}
            </span>

            {/* Row count (only for successful queries) */}
            {isSuccess && entry.rowCount > 0 && (
              <span className="flex items-center gap-1">
                <Hash className="w-3 h-3" />
                {entry.rowCount.toLocaleString()} rows
              </span>
            )}
          </div>

          {/* Error message */}
          {isError && entry.error && (
            <p className="mt-1.5 text-xs text-red-400 truncate">
              {entry.error}
            </p>
          )}
        </div>
      </div>
    </button>
  );
}

export default HistoryEntry;