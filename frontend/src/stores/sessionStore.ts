import { create } from 'zustand';
import type { Session, SessionState } from '../types/session';

/**
 * Toast notification for session events.
 */
export interface SessionToast {
  id: string;
  type: 'error' | 'reconnecting' | 'success';
  message: string;
  sessionId: string;
  connectionName?: string;
  isRecoverable?: boolean;
  retryCount?: number;
  nextRetryIn?: number;
  timestamp: number;
}

/**
 * Session store interface for managing active sessions and event state.
 */
interface SessionStore {
  // State
  sessions: Map<string, Session>;
  toasts: SessionToast[];
  reconnectingSessions: Map<string, { retryCount: number; nextRetryIn: number }>;

  // Session actions
  addSession: (session: Session) => void;
  updateSessionState: (sessionId: string, state: SessionState) => void;
  removeSession: (sessionId: string) => void;
  getSession: (sessionId: string) => Session | undefined;
  getSessionByConnectionId: (connectionId: string) => Session | undefined;

  // Toast actions
  addToast: (toast: SessionToast) => void;
  removeToast: (toastId: string) => void;
  clearToasts: () => void;

  // Reconnecting state actions
  setReconnecting: (sessionId: string, retryCount: number, nextRetryIn: number) => void;
  clearReconnecting: (sessionId: string) => void;
}

/**
 * Session store for managing active database sessions and event state.
 *
 * @example
 * ```typescript
 * // Add a new session
 * useSessionStore.getState().addSession({
 *   sessionId: 'session-1',
 *   connectionId: 'conn-1',
 *   state: 'active',
 *   connectionName: 'Production DB',
 *   databaseType: 'postgres'
 * });
 *
 * // Get session by connection ID
 * const session = useSessionStore.getState().getSessionByConnectionId('conn-1');
 *
 * // Subscribe to store changes
 * useSessionStore((state) => state.sessions);
 * ```
 */
export const useSessionStore = create<SessionStore>((set, get) => ({
  sessions: new Map(),
  toasts: [],
  reconnectingSessions: new Map(),

  // Session actions
  addSession: (session) => {
    const sessions = new Map(get().sessions);
    sessions.set(session.sessionId, session);
    set({ sessions });
  },

  updateSessionState: (sessionId, state) => {
    const sessions = new Map(get().sessions);
    const session = sessions.get(sessionId);
    if (session) {
      sessions.set(sessionId, { ...session, state });
      set({ sessions });
    }
  },

  removeSession: (sessionId) => {
    const sessions = new Map(get().sessions);
    sessions.delete(sessionId);
    set({ sessions });
  },

  getSession: (sessionId) => {
    return get().sessions.get(sessionId);
  },

  getSessionByConnectionId: (connectionId) => {
    for (const session of get().sessions.values()) {
      if (session.connectionId === connectionId) {
        return session;
      }
    }
    return undefined;
  },

  // Toast actions
  addToast: (toast) => {
    set((state) => ({
      toasts: [...state.toasts, toast],
    }));
  },

  removeToast: (toastId) => {
    set((state) => ({
      toasts: state.toasts.filter((t) => t.id !== toastId),
    }));
  },

  clearToasts: () => {
    set({ toasts: [] });
  },

  // Reconnecting state actions
  setReconnecting: (sessionId, retryCount, nextRetryIn) => {
    const reconnectingSessions = new Map(get().reconnectingSessions);
    reconnectingSessions.set(sessionId, { retryCount, nextRetryIn });
    set({ reconnectingSessions });
  },

  clearReconnecting: (sessionId) => {
    const reconnectingSessions = new Map(get().reconnectingSessions);
    reconnectingSessions.delete(sessionId);
    set({ reconnectingSessions });
  },
}));