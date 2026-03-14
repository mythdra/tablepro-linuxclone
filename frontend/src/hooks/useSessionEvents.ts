import { useEffect } from 'react';
import { EventsOn } from '../wailsjs/runtime';
import { useSessionStore, SessionToast } from '../stores/sessionStore';
import {
  SessionEvents,
  SessionCreatedPayload,
  SessionClosedPayload,
  SessionErrorPayload,
  SessionReconnectingPayload,
} from '../types/session';

/**
 * Callback functions for session event handlers.
 */
export interface SessionEventCallbacks {
  onSessionCreated?: (payload: SessionCreatedPayload) => void;
  onSessionClosed?: (payload: SessionClosedPayload) => void;
  onSessionError?: (payload: SessionErrorPayload) => void;
  onSessionReconnecting?: (payload: SessionReconnectingPayload) => void;
}

/**
 * Options for the useSessionEvents hook.
 */
export interface UseSessionEventsOptions {
  /** Custom callbacks for session events */
  callbacks?: SessionEventCallbacks;
  /** Whether to automatically show toasts for errors (default: true) */
  showErrorToasts?: boolean;
  /** Whether to automatically show toasts for reconnection attempts (default: true) */
  showReconnectingToasts?: boolean;
}

/**
 * Custom hook for subscribing to session lifecycle events from Wails backend.
 *
 * Subscribes to events:
 * - session:created - New session established
 * - session:closed - Session closed (intentionally or due to error)
 * - session:error - Session encountered an error
 * - session:reconnecting - Session attempting to reconnect
 *
 * Automatically cleans up all subscriptions on unmount.
 *
 * @param options - Configuration options for event handling
 *
 * @example
 * ```tsx
 * function AppLayout() {
 *   // Subscribe to session events with default behavior
 *   useSessionEvents();
 *
 *   return <div>...</div>;
 * }
 * ```
 *
 * @example
 * ```tsx
 * function CustomHandler() {
 *   useSessionEvents({
 *     callbacks: {
 *       onSessionCreated: (payload) => {
 *         console.log('Session created:', payload.connectionName);
 *       },
 *       onSessionError: (payload) => {
 *         if (payload.isRecoverable) {
 *           // Show custom recovery UI
 *         }
 *       },
 *     },
 *     showErrorToasts: false, // Disable default error toasts
 *   });
 *
 *   return <div>...</div>;
 * }
 * ```
 */
export function useSessionEvents(options: UseSessionEventsOptions = {}) {
  const { callbacks = {}, showErrorToasts = true, showReconnectingToasts = true } = options;

  const addSession = useSessionStore((state) => state.addSession);
  const removeSession = useSessionStore((state) => state.removeSession);
  const addToast = useSessionStore((state) => state.addToast);
  const removeToast = useSessionStore((state) => state.removeToast);
  const setReconnecting = useSessionStore((state) => state.setReconnecting);
  const clearReconnecting = useSessionStore((state) => state.clearReconnecting);

  useEffect(() => {
    // Subscribe to session:created event
    const unsubscribeCreated = EventsOn(SessionEvents.CREATED, (data: SessionCreatedPayload) => {
      // Clear any reconnecting state
      clearReconnecting(data.sessionId);

      // Add session to store
      addSession({
        sessionId: data.sessionId,
        connectionId: data.connectionId,
        state: 'active',
        connectionName: data.connectionName,
        databaseType: data.databaseType,
      });

      // Call custom callback if provided
      callbacks.onSessionCreated?.(data);
    });

    // Subscribe to session:closed event
    const unsubscribeClosed = EventsOn(SessionEvents.CLOSED, (data: SessionClosedPayload) => {
      // Clear reconnecting state if present
      clearReconnecting(data.sessionId);

      // Remove session from store
      removeSession(data.sessionId);

      // Call custom callback if provided
      callbacks.onSessionClosed?.(data);
    });

    // Subscribe to session:error event
    const unsubscribeError = EventsOn(SessionEvents.ERROR, (data: SessionErrorPayload) => {
      // Show error toast if enabled
      if (showErrorToasts) {
        const toast: SessionToast = {
          id: `error-${data.sessionId}-${Date.now()}`,
          type: 'error',
          message: data.error,
          sessionId: data.sessionId,
          isRecoverable: data.isRecoverable,
          timestamp: data.timestamp,
        };
        addToast(toast);

        // Auto-dismiss non-recoverable errors after 5 seconds
        if (!data.isRecoverable) {
          setTimeout(() => {
            removeToast(toast.id);
          }, 5000);
        }
      }

      // Call custom callback if provided
      callbacks.onSessionError?.(data);
    });

    // Subscribe to session:reconnecting event
    const unsubscribeReconnecting = EventsOn(
      SessionEvents.RECONNECTING,
      (data: SessionReconnectingPayload) => {
        // Update reconnecting state
        setReconnecting(data.sessionId, data.retryCount, data.nextRetryIn);

        // Show reconnecting toast if enabled
        if (showReconnectingToasts) {
          const toast: SessionToast = {
            id: `reconnecting-${data.sessionId}`,
            type: 'reconnecting',
            message: `Reconnecting... (attempt ${data.retryCount})`,
            sessionId: data.sessionId,
            retryCount: data.retryCount,
            nextRetryIn: data.nextRetryIn,
            timestamp: data.timestamp,
          };
          addToast(toast);
        }

        // Call custom callback if provided
        callbacks.onSessionReconnecting?.(data);
      }
    );

    // Cleanup on unmount
    return () => {
      unsubscribeCreated();
      unsubscribeClosed();
      unsubscribeError();
      unsubscribeReconnecting();
    };
  }, [
    addSession,
    removeSession,
    addToast,
    removeToast,
    setReconnecting,
    clearReconnecting,
    callbacks,
    showErrorToasts,
    showReconnectingToasts,
  ]);
}