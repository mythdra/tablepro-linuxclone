import { useEffect } from 'react';
import { X, AlertCircle, RefreshCw, CheckCircle } from 'lucide-react';
import { useSessionStore, SessionToast } from '../stores/sessionStore';

/**
 * Props for individual toast notification.
 */
interface ToastItemProps {
  toast: SessionToast;
  onDismiss: (id: string) => void;
  onReconnect?: (sessionId: string) => void;
  /** Auto-dismiss duration in milliseconds (0 = no auto-dismiss) */
  autoDismissAfter?: number;
}

/**
 * Individual toast notification item.
 */
function ToastItem({ toast, onDismiss, onReconnect, autoDismissAfter = 0 }: ToastItemProps) {
  const icons = {
    error: AlertCircle,
    reconnecting: RefreshCw,
    success: CheckCircle,
  };

  const colorClasses = {
    error: 'bg-red-900/90 border-red-700 text-red-100',
    reconnecting: 'bg-amber-900/90 border-amber-700 text-amber-100',
    success: 'bg-emerald-900/90 border-emerald-700 text-emerald-100',
  };

  const iconColors = {
    error: 'text-red-400',
    reconnecting: 'text-amber-400 animate-spin',
    success: 'text-emerald-400',
  };

  const Icon = icons[toast.type];

  useEffect(() => {
    if (autoDismissAfter > 0) {
      const timer = setTimeout(() => {
        onDismiss(toast.id);
      }, autoDismissAfter);
      return () => clearTimeout(timer);
    }
  }, [autoDismissAfter, toast.id, onDismiss]);

  return (
    <div
      className={`flex items-start gap-3 p-4 rounded-lg border shadow-lg min-w-[320px] max-w-[480px] ${colorClasses[toast.type]}`}
      role="alert"
      aria-live={toast.type === 'error' ? 'assertive' : 'polite'}
    >
      <Icon className={`w-5 h-5 flex-shrink-0 mt-0.5 ${iconColors[toast.type]}`} />

      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium">{toast.message}</p>

        {toast.type === 'reconnecting' && toast.retryCount !== undefined && (
          <p className="text-xs mt-1 opacity-80">
            Attempt {toast.retryCount}
            {toast.nextRetryIn !== undefined && ` · Next retry in ${toast.nextRetryIn}s`}
          </p>
        )}

        {toast.type === 'error' && toast.isRecoverable && onReconnect && (
          <button
            onClick={() => {
              onReconnect(toast.sessionId);
              onDismiss(toast.id);
            }}
            className="mt-2 px-3 py-1 text-xs font-medium bg-white/10 hover:bg-white/20 rounded transition-colors"
          >
            Reconnect
          </button>
        )}
      </div>

      <button
        onClick={() => onDismiss(toast.id)}
        className="flex-shrink-0 p-1 hover:bg-white/10 rounded transition-colors"
        aria-label="Dismiss notification"
      >
        <X className="w-4 h-4" />
      </button>
    </div>
  );
}

/**
 * Props for SessionStatusToast container.
 */
export interface SessionStatusToastProps {
  /** Position of the toast container */
  position?: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left';
  /** Maximum number of toasts to show at once */
  maxToasts?: number;
  /** Callback when user clicks reconnect button */
  onReconnect?: (sessionId: string) => void;
}

/**
 * Container component for displaying session-related toast notifications.
 *
 * Renders toast notifications from the session store for:
 * - Session errors (recoverable and non-recoverable)
 * - Reconnection attempts with retry count
 * - Success notifications
 *
 * @example
 * ```tsx
 * function App() {
 *   useSessionEvents(); // Populate toasts from events
 *
 *   return (
 *     <div>
 *       <SessionStatusToast
 *         position="bottom-right"
 *         onReconnect={(sessionId) => {
 *           // Handle reconnect
 *         }}
 *       />
 *       {/* ... rest of app *\/}
 *     </div>
 *   );
 * }
 * ```
 */
export function SessionStatusToast({
  position = 'bottom-right',
  maxToasts = 3,
  onReconnect,
}: SessionStatusToastProps) {
  const toasts = useSessionStore((state) => state.toasts);
  const removeToast = useSessionStore((state) => state.removeToast);

  // Only show the most recent toasts up to maxToasts
  const visibleToasts = toasts.slice(-maxToasts);

  const positionClasses = {
    'top-right': 'top-4 right-4',
    'top-left': 'top-4 left-4',
    'bottom-right': 'bottom-4 right-4',
    'bottom-left': 'bottom-4 left-4',
  };

  return (
    <div
      className={`fixed z-50 flex flex-col gap-2 ${positionClasses[position]}`}
      role="region"
      aria-label="Session notifications"
    >
      {visibleToasts.map((toast) => {
        const shouldAutoDismiss =
          toast.type === 'error' && toast.isRecoverable === false;

        return (
          <ToastItem
            key={toast.id}
            toast={toast}
            onDismiss={removeToast}
            onReconnect={onReconnect}
            autoDismissAfter={shouldAutoDismiss ? 5000 : 0}
          />
        );
      })}
    </div>
  );
}

export default SessionStatusToast;