import { useState } from 'react';
import {
  Database,
  Server,
  MoreVertical,
  Edit2,
  Trash2,
  Copy,
  Plug,
  PlugZap,
  Loader2,
  Plus,
  FolderOpen,
} from 'lucide-react';
import { DatabaseIcon } from './DatabaseIcon';
import type { DatabaseConnection, ConnectionStatus } from '../types';

// Database icons now provided by shared DatabaseIcon component

const statusColors: Record<ConnectionStatus, string> = {
  disconnected: 'bg-slate-500',
  connecting: 'bg-amber-500 animate-pulse',
  connected: 'bg-emerald-500',
  error: 'bg-red-500',
};

const statusLabels: Record<ConnectionStatus, string> = {
  disconnected: 'Disconnected',
  connecting: 'Connecting...',
  connected: 'Connected',
  error: 'Error',
};

interface ConnectionListProps {
  connections: DatabaseConnection[];
  sessions: Map<string, { status: ConnectionStatus }>;
  onSelect: (connection: DatabaseConnection) => void;
  onEdit: (connection: DatabaseConnection) => void;
  onDelete: (connection: DatabaseConnection) => void;
  onDuplicate: (connection: DatabaseConnection) => void;
  onConnect: (connection: DatabaseConnection) => void;
  onDisconnect: (connection: DatabaseConnection) => void;
  onNewConnection: () => void;
}

interface ConnectionGroup {
  name: string;
  connections: DatabaseConnection[];
}

export function ConnectionList({
  connections,
  sessions,
  onSelect,
  onEdit,
  onDelete,
  onDuplicate,
  onConnect,
  onDisconnect,
  onNewConnection,
}: ConnectionListProps) {
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set(['default']));
  const [activeMenu, setActiveMenu] = useState<string | null>(null);

  const groupedConnections = connections.reduce<ConnectionGroup[]>((acc, conn) => {
    const groupName = conn.group || 'default';
    const existing = acc.find((g) => g.name === groupName);
    if (existing) {
      existing.connections.push(conn);
    } else {
      acc.push({ name: groupName, connections: [conn] });
    }
    return acc;
  }, []);

  const toggleGroup = (groupName: string) => {
    setExpandedGroups((prev) => {
      const next = new Set(prev);
      if (next.has(groupName)) {
        next.delete(groupName);
      } else {
        next.add(groupName);
      }
      return next;
    });
  };

  const getStatus = (connectionId: string): ConnectionStatus => {
    return sessions.get(connectionId)?.status || 'disconnected';
  };

  return (
    <div className="h-full flex flex-col bg-slate-900">
      <div className="flex items-center justify-between px-4 py-3 border-b border-slate-700">
        <h2 className="text-lg font-semibold text-white">Connections</h2>
        <button
          onClick={onNewConnection}
          className="flex items-center gap-2 px-3 py-1.5 text-sm font-medium text-white bg-primary rounded-lg hover:bg-blue-600 transition-colors"
        >
          <Plus className="w-4 h-4" />
          New
        </button>
      </div>

      <div className="flex-1 overflow-y-auto">
        {groupedConnections.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 text-slate-500">
            <Database className="w-12 h-12 mb-3" />
            <p className="text-sm">No connections yet</p>
            <button
              onClick={onNewConnection}
              className="mt-3 text-primary hover:underline text-sm"
            >
              Create your first connection
            </button>
          </div>
        ) : (
          <div className="py-2">
            {groupedConnections.map((group) => (
              <div key={group.name}>
                <button
                  onClick={() => toggleGroup(group.name)}
                  className="w-full flex items-center gap-2 px-4 py-2 text-sm font-medium text-slate-400 hover:text-white hover:bg-slate-800/50 transition-colors"
                >
                  <FolderOpen
                    className={`w-4 h-4 transition-transform ${
                      expandedGroups.has(group.name) ? 'rotate-0' : '-rotate-90'
                    }`}
                  />
                  {group.name}
                  <span className="ml-auto text-xs text-slate-500">
                    {group.connections.length}
                  </span>
                </button>

                {expandedGroups.has(group.name) && (
                  <div className="ml-4">
                    {group.connections.map((connection) => {
                      const status = getStatus(connection.id);
                      const isConnecting = status === 'connecting';
                      const isConnected = status === 'connected';

                      return (
                        <div
                          key={connection.id}
                          className="group relative flex items-center gap-3 px-4 py-2.5 hover:bg-slate-800/50 cursor-pointer transition-colors"
                          onClick={() => onSelect(connection)}
                        >
                          <div
                            className={`w-2 h-2 rounded-full ${statusColors[status]}`}
                            title={statusLabels[status]}
                          />

                          <DatabaseIcon type={connection.type} size="md" />

                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2">
                              <span className="text-sm font-medium text-white truncate">
                                {connection.name}
                              </span>
                              {connection.colorTag && (
                                <span
                                  className="w-2 h-2 rounded-full"
                                  style={{
                                    backgroundColor: {
                                      red: '#ef4444',
                                      orange: '#f97316',
                                      yellow: '#eab308',
                                      green: '#22c55e',
                                      blue: '#3b82f6',
                                      purple: '#a855f7',
                                    }[connection.colorTag] || connection.colorTag,
                                  }}
                                />
                              )}
                            </div>
                            <div className="flex items-center gap-1 text-xs text-slate-500">
                              <Server className="w-3 h-3" />
                              <span className="truncate">
                                {connection.host}:{connection.port}
                              </span>
                            </div>
                          </div>

                          <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                            {isConnecting ? (
                              <Loader2 className="w-4 h-4 text-amber-500 animate-spin" />
                            ) : isConnected ? (
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  onDisconnect(connection);
                                }}
                                className="p-1.5 text-slate-400 hover:text-white rounded hover:bg-slate-700 transition-colors"
                                title="Disconnect"
                              >
                                <PlugZap className="w-4 h-4" />
                              </button>
                            ) : (
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  onConnect(connection);
                                }}
                                className="p-1.5 text-slate-400 hover:text-white rounded hover:bg-slate-700 transition-colors"
                                title="Connect"
                              >
                                <Plug className="w-4 h-4" />
                              </button>
                            )}

                            <div className="relative">
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  setActiveMenu(activeMenu === connection.id ? null : connection.id);
                                }}
                                className="p-1.5 text-slate-400 hover:text-white rounded hover:bg-slate-700 transition-colors"
                              >
                                <MoreVertical className="w-4 h-4" />
                              </button>

                              {activeMenu === connection.id && (
                                <div className="absolute right-0 top-full mt-1 w-40 bg-slate-800 border border-slate-700 rounded-lg shadow-lg z-10">
                                  <button
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      onEdit(connection);
                                      setActiveMenu(null);
                                    }}
                                    className="w-full flex items-center gap-2 px-3 py-2 text-sm text-slate-300 hover:bg-slate-700 rounded-t-lg transition-colors"
                                  >
                                    <Edit2 className="w-4 h-4" />
                                    Edit
                                  </button>
                                  <button
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      onDuplicate(connection);
                                      setActiveMenu(null);
                                    }}
                                    className="w-full flex items-center gap-2 px-3 py-2 text-sm text-slate-300 hover:bg-slate-700 transition-colors"
                                  >
                                    <Copy className="w-4 h-4" />
                                    Duplicate
                                  </button>
                                  <button
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      onDelete(connection);
                                      setActiveMenu(null);
                                    }}
                                    className="w-full flex items-center gap-2 px-3 py-2 text-sm text-red-400 hover:bg-slate-700 rounded-b-lg transition-colors"
                                  >
                                    <Trash2 className="w-4 h-4" />
                                    Delete
                                  </button>
                                </div>
                              )}
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {activeMenu && (
        <div
          className="fixed inset-0 z-0"
          onClick={() => setActiveMenu(null)}
        />
      )}
    </div>
  );
}

export default ConnectionList;