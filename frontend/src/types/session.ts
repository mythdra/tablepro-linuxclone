/**
 * Session state types for managing active database connections.
 */

/**
 * Possible states for a session.
 */
export type SessionState = 'active' | 'idle' | 'closed';

/**
 * Represents an active database session.
 */
export interface Session {
  sessionId: string;
  connectionId: string;
  state: SessionState;
  connectionName: string;
  databaseType: string;
}

/**
 * Base payload for all session events.
 */
export interface SessionEventPayload {
  sessionId: string;
  timestamp: number;
}

/**
 * Payload for session:created event.
 * Emitted when a new session is successfully established.
 */
export interface SessionCreatedPayload extends SessionEventPayload {
  connectionId: string;
  connectionName: string;
  databaseType: string;
}

/**
 * Payload for session:closed event.
 * Emitted when a session is closed (intentionally or due to error).
 */
export interface SessionClosedPayload extends SessionEventPayload {
  reason: string;
}

/**
 * Payload for session:error event.
 * Emitted when a session encounters an error.
 */
export interface SessionErrorPayload extends SessionEventPayload {
  error: string;
  isRecoverable: boolean;
}

/**
 * Payload for session:reconnecting event.
 * Emitted when a session is attempting to reconnect after a failure.
 */
export interface SessionReconnectingPayload extends SessionEventPayload {
  retryCount: number;
  nextRetryIn: number;
}

/**
 * Union type for all session event payloads.
 */
export type AnySessionEventPayload =
  | SessionCreatedPayload
  | SessionClosedPayload
  | SessionErrorPayload
  | SessionReconnectingPayload;

/**
 * Event names for session lifecycle events.
 */
export const SessionEvents = {
  CREATED: 'session:created',
  CLOSED: 'session:closed',
  ERROR: 'session:error',
  RECONNECTING: 'session:reconnecting',
} as const;

export type SessionEventName = (typeof SessionEvents)[keyof typeof SessionEvents];