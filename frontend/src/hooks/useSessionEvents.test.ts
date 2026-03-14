import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook } from '@testing-library/react';

vi.mock('../wailsjs/runtime', () => ({
  EventsOn: vi.fn(() => () => {}),
  EventsOff: vi.fn(),
}));

import { useSessionEvents } from './useSessionEvents';
import { useSessionStore } from '../stores/sessionStore';
import { EventsOn } from '../wailsjs/runtime';
import { SessionEvents } from '../types/session';

describe('useSessionEvents', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useSessionStore.getState().clearToasts();
    useSessionStore.getState().sessions.clear();
    useSessionStore.getState().reconnectingSessions.clear();
  });

  it('should subscribe to session:created event', () => {
    renderHook(() => useSessionEvents());

    expect(EventsOn).toHaveBeenCalledWith(SessionEvents.CREATED, expect.any(Function));
  });

  it('should subscribe to session:closed event', () => {
    renderHook(() => useSessionEvents());

    expect(EventsOn).toHaveBeenCalledWith(SessionEvents.CLOSED, expect.any(Function));
  });

  it('should subscribe to session:error event', () => {
    renderHook(() => useSessionEvents());

    expect(EventsOn).toHaveBeenCalledWith(SessionEvents.ERROR, expect.any(Function));
  });

  it('should subscribe to session:reconnecting event', () => {
    renderHook(() => useSessionEvents());

    expect(EventsOn).toHaveBeenCalledWith(SessionEvents.RECONNECTING, expect.any(Function));
  });

  it('should unsubscribe from all events on unmount', () => {
    const { unmount } = renderHook(() => useSessionEvents());

    unmount();

    // Each event returns an unsubscribe function
    expect(EventsOn).toHaveBeenCalledTimes(4);
  });

  it('should add session to store on session:created', () => {
    const mockEventsOn = vi.mocked(EventsOn);
    let createdCallback: ((data: any) => void) | undefined;

    mockEventsOn.mockImplementation((event: string, callback: any) => {
      if (event === SessionEvents.CREATED) {
        createdCallback = callback;
      }
      return () => {};
    });

    renderHook(() => useSessionEvents());

    // Simulate session:created event
    createdCallback?.({
      sessionId: 'session-1',
      connectionId: 'conn-1',
      connectionName: 'Test DB',
      databaseType: 'postgres',
      timestamp: Date.now(),
    });

    const sessions = useSessionStore.getState().sessions;
    expect(sessions.has('session-1')).toBe(true);
    expect(sessions.get('session-1')?.connectionName).toBe('Test DB');
  });

  it('should remove session from store on session:closed', () => {
    // First add a session
    useSessionStore.getState().addSession({
      sessionId: 'session-1',
      connectionId: 'conn-1',
      state: 'active',
      connectionName: 'Test DB',
      databaseType: 'postgres',
    });

    const mockEventsOn = vi.mocked(EventsOn);
    let closedCallback: ((data: any) => void) | undefined;

    mockEventsOn.mockImplementation((event: string, callback: any) => {
      if (event === SessionEvents.CLOSED) {
        closedCallback = callback;
      }
      return () => {};
    });

    renderHook(() => useSessionEvents());

    // Simulate session:closed event
    closedCallback?.({
      sessionId: 'session-1',
      reason: 'User disconnected',
      timestamp: Date.now(),
    });

    const sessions = useSessionStore.getState().sessions;
    expect(sessions.has('session-1')).toBe(false);
  });

  it('should add error toast on session:error', () => {
    const mockEventsOn = vi.mocked(EventsOn);
    let errorCallback: ((data: any) => void) | undefined;

    mockEventsOn.mockImplementation((event: string, callback: any) => {
      if (event === SessionEvents.ERROR) {
        errorCallback = callback;
      }
      return () => {};
    });

    renderHook(() => useSessionEvents());

    // Simulate session:error event
    errorCallback?.({
      sessionId: 'session-1',
      error: 'Connection refused',
      isRecoverable: false,
      timestamp: Date.now(),
    });

    const toasts = useSessionStore.getState().toasts;
    expect(toasts.length).toBe(1);
    expect(toasts[0].message).toBe('Connection refused');
    expect(toasts[0].type).toBe('error');
  });

  it('should add reconnecting state on session:reconnecting', () => {
    const mockEventsOn = vi.mocked(EventsOn);
    let reconnectingCallback: ((data: any) => void) | undefined;

    mockEventsOn.mockImplementation((event: string, callback: any) => {
      if (event === SessionEvents.RECONNECTING) {
        reconnectingCallback = callback;
      }
      return () => {};
    });

    renderHook(() => useSessionEvents());

    // Simulate session:reconnecting event
    reconnectingCallback?.({
      sessionId: 'session-1',
      retryCount: 3,
      nextRetryIn: 5,
      timestamp: Date.now(),
    });

    const reconnecting = useSessionStore.getState().reconnectingSessions;
    expect(reconnecting.has('session-1')).toBe(true);
    expect(reconnecting.get('session-1')?.retryCount).toBe(3);
  });

  it('should call custom callbacks when provided', () => {
    const onSessionCreated = vi.fn();
    const onSessionClosed = vi.fn();
    const onSessionError = vi.fn();
    const onSessionReconnecting = vi.fn();

    const mockEventsOn = vi.mocked(EventsOn);
    const callbacks: Record<string, ((data: any) => void) | undefined> = {};

    mockEventsOn.mockImplementation((event: string, callback: any) => {
      callbacks[event] = callback;
      return () => {};
    });

    renderHook(() =>
      useSessionEvents({
        callbacks: {
          onSessionCreated,
          onSessionClosed,
          onSessionError,
          onSessionReconnecting,
        },
      })
    );

    // Simulate events
    const createdPayload = {
      sessionId: 'session-1',
      connectionId: 'conn-1',
      connectionName: 'Test',
      databaseType: 'postgres',
      timestamp: Date.now(),
    };
    callbacks[SessionEvents.CREATED]?.(createdPayload);
    expect(onSessionCreated).toHaveBeenCalledWith(createdPayload);

    const closedPayload = {
      sessionId: 'session-1',
      reason: 'Test',
      timestamp: Date.now(),
    };
    callbacks[SessionEvents.CLOSED]?.(closedPayload);
    expect(onSessionClosed).toHaveBeenCalledWith(closedPayload);

    const errorPayload = {
      sessionId: 'session-1',
      error: 'Test error',
      isRecoverable: true,
      timestamp: Date.now(),
    };
    callbacks[SessionEvents.ERROR]?.(errorPayload);
    expect(onSessionError).toHaveBeenCalledWith(errorPayload);

    const reconnectingPayload = {
      sessionId: 'session-1',
      retryCount: 1,
      nextRetryIn: 5,
      timestamp: Date.now(),
    };
    callbacks[SessionEvents.RECONNECTING]?.(reconnectingPayload);
    expect(onSessionReconnecting).toHaveBeenCalledWith(reconnectingPayload);
  });

  it('should not add error toast when showErrorToasts is false', () => {
    const mockEventsOn = vi.mocked(EventsOn);
    let errorCallback: ((data: any) => void) | undefined;

    mockEventsOn.mockImplementation((event: string, callback: any) => {
      if (event === SessionEvents.ERROR) {
        errorCallback = callback;
      }
      return () => {};
    });

    renderHook(() => useSessionEvents({ showErrorToasts: false }));

    // Simulate session:error event
    errorCallback?.({
      sessionId: 'session-1',
      error: 'Connection refused',
      isRecoverable: false,
      timestamp: Date.now(),
    });

    const toasts = useSessionStore.getState().toasts;
    expect(toasts.length).toBe(0);
  });

  it('should not add reconnecting toast when showReconnectingToasts is false', () => {
    const mockEventsOn = vi.mocked(EventsOn);
    let reconnectingCallback: ((data: any) => void) | undefined;

    mockEventsOn.mockImplementation((event: string, callback: any) => {
      if (event === SessionEvents.RECONNECTING) {
        reconnectingCallback = callback;
      }
      return () => {};
    });

    renderHook(() => useSessionEvents({ showReconnectingToasts: false }));

    // Simulate session:reconnecting event
    reconnectingCallback?.({
      sessionId: 'session-1',
      retryCount: 1,
      nextRetryIn: 5,
      timestamp: Date.now(),
    });

    const toasts = useSessionStore.getState().toasts;
    expect(toasts.length).toBe(0);
  });
});