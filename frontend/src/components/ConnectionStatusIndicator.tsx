import { Loader2, Wifi, WifiOff } from 'lucide-react';
import { useSessionStore } from '../stores/sessionStore';

/**
 * Display status for the connection indicator.
 */
type DisplayStatus = 'connected' | 'disconnected' | 'reconnecting';

/**
 * Props for ConnectionStatusIndicator component.
 */
export interface ConnectionStatusIndicatorProps {
  /** Connection ID to show status for */
  connectionId?: string;
  /** Show connection name alongside status */
  showConnectionName?: boolean;
  /** Show database type icon */
  showDatabaseType?: boolean;
  /** Compact mode (smaller text, no labels) */
  compact?: boolean;
  /** Additional CSS classes */
  className?: string;
}

/**
 * Status indicator configuration.
 */
const statusConfig: Record<
  DisplayStatus,
  {
    label: string;
    dotColor: string;
    textColor: string;
    icon?: typeof Wifi;
  }
> = {
  connected: {
    label: 'Connected',
    dotColor: 'bg-emerald-500',
    textColor: 'text-emerald-400',
    icon: Wifi,
  },
  disconnected: {
    label: 'Disconnected',
    dotColor: 'bg-slate-500',
    textColor: 'text-slate-400',
    icon: WifiOff,
  },
  reconnecting: {
    label: 'Reconnecting...',
    dotColor: 'bg-amber-500',
    textColor: 'text-amber-400',
  },
};

/**
 * Database type display names and icons.
 */
const databaseDisplay: Record<string, { name: string; icon: string }> = {
  postgres: { name: 'PostgreSQL', icon: '🐘' },
  mysql: { name: 'MySQL', icon: '🐬' },
  sqlite: { name: 'SQLite', icon: '📦' },
  duckdb: { name: 'DuckDB', icon: '🦆' },
  mssql: { name: 'SQL Server', icon: '🏢' },
  clickhouse: { name: 'ClickHouse', icon: '🏠' },
  mongodb: { name: 'MongoDB', icon: '🍃' },
  redis: { name: 'Redis', icon: '🔴' },
};

/**
 * Component that displays real-time connection status based on session events.
 *
 * Shows:
 * - Connected (green) when session is active
 * - Disconnected (gray) when no active session
 * - Reconnecting (amber spinner) during reconnection attempts
 *
 * @example
 * ```tsx
 * function QueryToolbar() {
 *   return (
 *     <div className="flex items-center gap-4">
 *       <ConnectionStatusIndicator
 *         connectionId="conn-123"
 *         showConnectionName
 *         showDatabaseType
 *       />
 *     </div>
 *   );
 * }
 * ```
 *
 * @example
 * ```tsx
 * // Compact mode for sidebar
 * <ConnectionStatusIndicator
 *   connectionId="conn-123"
 *   compact
 * />
 * ```
 */
export function ConnectionStatusIndicator({
  connectionId,
  showConnectionName = true,
  showDatabaseType = true,
  compact = false,
  className = '',
}: ConnectionStatusIndicatorProps) {
  const sessions = useSessionStore((state) => state.sessions);
  const reconnectingSessions = useSessionStore((state) => state.reconnectingSessions);

  // Find session for this connection
  const session = connectionId
    ? sessions.size > 0
      ? Array.from(sessions.values()).find((s) => s.connectionId === connectionId)
      : undefined
    : sessions.size > 0
      ? Array.from(sessions.values())[0]
      : undefined;

  // Check if reconnecting
  const isReconnecting = session ? reconnectingSessions.has(session.sessionId) : false;

  // Determine display status
  const getDisplayStatus = (): DisplayStatus => {
    if (!session) return 'disconnected';
    if (isReconnecting) return 'reconnecting';
    if (session.state === 'active' || session.state === 'idle') return 'connected';
    return 'disconnected';
  };

  const status = getDisplayStatus();
  const config = statusConfig[status];

  // Get database display info
  const dbInfo = session?.databaseType ? databaseDisplay[session.databaseType] : null;

  if (compact) {
    return (
      <div className={`flex items-center gap-1.5 ${className}`}>
        <div
          className={`w-2 h-2 rounded-full ${config.dotColor} ${status === 'reconnecting' ? 'animate-pulse' : ''}`}
          title={config.label}
        />
        {session && showConnectionName && (
          <span className="text-xs text-slate-400 truncate max-w-[100px]">
            {session.connectionName}
          </span>
        )}
      </div>
    );
  }

  return (
    <div className={`flex items-center gap-2 ${className}`}>
      {/* Status indicator */}
      <div className="flex items-center gap-2">
        {status === 'reconnecting' ? (
          <Loader2 className={`w-4 h-4 animate-spin ${config.textColor}`} />
        ) : config.icon ? (
          <config.icon className={`w-4 h-4 ${config.textColor}`} />
        ) : (
          <div
            className={`w-2.5 h-2.5 rounded-full ${config.dotColor}`}
          />
        )}
        <span className={`text-sm font-medium ${config.textColor}`}>{config.label}</span>
      </div>

      {/* Connection name */}
      {session && showConnectionName && (
        <span className="text-sm text-slate-400 truncate max-w-[200px]">
          {session.connectionName}
        </span>
      )}

      {/* Database type */}
      {session && showDatabaseType && dbInfo && (
        <div className="flex items-center gap-1 px-2 py-0.5 bg-slate-800 rounded text-xs text-slate-400">
          <span>{dbInfo.icon}</span>
          <span>{dbInfo.name}</span>
        </div>
      )}
    </div>
  );
}

export default ConnectionStatusIndicator;