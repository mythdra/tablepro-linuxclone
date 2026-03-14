import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { SessionStatusToast } from '../SessionStatusToast';
import { useSessionStore } from '../../stores/sessionStore';

describe('SessionStatusToast', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useSessionStore.getState().clearToasts();
  });

  describe('Error toasts', () => {
    it('renders error toast on session:error', () => {
      // Add an error toast to the store
      useSessionStore.getState().addToast({
        id: 'error-1',
        type: 'error',
        message: 'Connection failed',
        sessionId: 'session-1',
        isRecoverable: false,
        timestamp: Date.now(),
      });

      render(<SessionStatusToast />);

      expect(screen.getByText('Connection failed')).toBeInTheDocument();
      expect(screen.getByRole('alert')).toBeInTheDocument();
    });

    it('shows reconnect button for recoverable errors', () => {
      useSessionStore.getState().addToast({
        id: 'error-2',
        type: 'error',
        message: 'Connection lost',
        sessionId: 'session-1',
        isRecoverable: true,
        timestamp: Date.now(),
      });

      render(<SessionStatusToast onReconnect={() => {}} />);

      expect(screen.getByRole('button', { name: 'Reconnect' })).toBeInTheDocument();
    });

    it('does not show reconnect button for non-recoverable errors', () => {
      useSessionStore.getState().addToast({
        id: 'error-3',
        type: 'error',
        message: 'Authentication failed',
        sessionId: 'session-1',
        isRecoverable: false,
        timestamp: Date.now(),
      });

      render(<SessionStatusToast />);

      expect(screen.queryByRole('button', { name: 'Reconnect' })).not.toBeInTheDocument();
    });

    it('calls onReconnect when reconnect button is clicked', () => {
      const onReconnect = vi.fn();

      useSessionStore.getState().addToast({
        id: 'error-4',
        type: 'error',
        message: 'Connection lost',
        sessionId: 'session-1',
        isRecoverable: true,
        timestamp: Date.now(),
      });

      render(<SessionStatusToast onReconnect={onReconnect} />);

      const reconnectButton = screen.getByRole('button', { name: 'Reconnect' });
      fireEvent.click(reconnectButton);

      expect(onReconnect).toHaveBeenCalledWith('session-1');
    });
  });

  describe('Reconnecting toasts', () => {
    it('shows retry count on reconnecting', () => {
      useSessionStore.getState().addToast({
        id: 'reconnecting-1',
        type: 'reconnecting',
        message: 'Reconnecting...',
        sessionId: 'session-1',
        retryCount: 3,
        nextRetryIn: 5,
        timestamp: Date.now(),
      });

      render(<SessionStatusToast />);

      expect(screen.getByText('Reconnecting...')).toBeInTheDocument();
      expect(screen.getByText(/Attempt 3/)).toBeInTheDocument();
      expect(screen.getByText(/Next retry in 5s/)).toBeInTheDocument();
    });

    it('shows spinner animation for reconnecting state', () => {
      useSessionStore.getState().addToast({
        id: 'reconnecting-2',
        type: 'reconnecting',
        message: 'Reconnecting...',
        sessionId: 'session-1',
        retryCount: 1,
        timestamp: Date.now(),
      });

      render(<SessionStatusToast />);

      // The spinner should have animate-spin class
      const spinner = document.querySelector('.animate-spin');
      expect(spinner).toBeInTheDocument();
    });
  });

  describe('Auto-dismiss', () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it('auto-dismisses non-recoverable errors after 5 seconds', async () => {
      useSessionStore.getState().addToast({
        id: 'error-5',
        type: 'error',
        message: 'Temporary error',
        sessionId: 'session-1',
        isRecoverable: false,
        timestamp: Date.now(),
      });

      render(<SessionStatusToast />);

      expect(screen.getByText('Temporary error')).toBeInTheDocument();

      await act(async () => {
        vi.advanceTimersByTime(5000);
      });

      expect(screen.queryByText('Temporary error')).not.toBeInTheDocument();
    });

    it('does not auto-dismiss recoverable errors', async () => {
      useSessionStore.getState().addToast({
        id: 'error-6',
        type: 'error',
        message: 'Recoverable error',
        sessionId: 'session-1',
        isRecoverable: true,
        timestamp: Date.now(),
      });

      render(<SessionStatusToast />);

      await act(async () => {
        vi.advanceTimersByTime(10000);
      });

      expect(screen.getByText('Recoverable error')).toBeInTheDocument();
    });
  });

  describe('Manual dismiss', () => {
    it('dismisses toast when X button is clicked', () => {
      useSessionStore.getState().addToast({
        id: 'error-7',
        type: 'error',
        message: 'Dismissible error',
        sessionId: 'session-1',
        isRecoverable: true,
        timestamp: Date.now(),
      });

      render(<SessionStatusToast />);

      const dismissButton = screen.getByRole('button', { name: 'Dismiss notification' });
      fireEvent.click(dismissButton);

      expect(screen.queryByText('Dismissible error')).not.toBeInTheDocument();
    });
  });

  describe('Multiple toasts', () => {
    it('limits visible toasts to maxToasts prop', () => {
      // Add 5 toasts
      for (let i = 1; i <= 5; i++) {
        useSessionStore.getState().addToast({
          id: `error-${i}`,
          type: 'error',
          message: `Error ${i}`,
          sessionId: 'session-1',
          isRecoverable: false,
          timestamp: Date.now(),
        });
      }

      render(<SessionStatusToast maxToasts={3} />);

      // Should only show last 3 toasts
      expect(screen.queryByText('Error 1')).not.toBeInTheDocument();
      expect(screen.queryByText('Error 2')).not.toBeInTheDocument();
      expect(screen.getByText('Error 3')).toBeInTheDocument();
      expect(screen.getByText('Error 4')).toBeInTheDocument();
      expect(screen.getByText('Error 5')).toBeInTheDocument();
    });

    it('renders all toast types together', () => {
      useSessionStore.getState().addToast({
        id: 'error-1',
        type: 'error',
        message: 'Error message',
        sessionId: 'session-1',
        isRecoverable: false,
        timestamp: Date.now(),
      });

      useSessionStore.getState().addToast({
        id: 'reconnecting-1',
        type: 'reconnecting',
        message: 'Reconnecting message',
        sessionId: 'session-2',
        retryCount: 2,
        timestamp: Date.now(),
      });

      render(<SessionStatusToast />);

      expect(screen.getByText('Error message')).toBeInTheDocument();
      expect(screen.getByText('Reconnecting message')).toBeInTheDocument();
    });
  });

  describe('Positioning', () => {
    it('renders at bottom-right by default', () => {
      useSessionStore.getState().addToast({
        id: 'error-1',
        type: 'error',
        message: 'Test',
        sessionId: 'session-1',
        timestamp: Date.now(),
      });

      render(<SessionStatusToast />);

      const container = screen.getByRole('region', { name: 'Session notifications' });
      expect(container.className).toContain('bottom-4');
      expect(container.className).toContain('right-4');
    });

    it('renders at specified position', () => {
      useSessionStore.getState().addToast({
        id: 'error-1',
        type: 'error',
        message: 'Test',
        sessionId: 'session-1',
        timestamp: Date.now(),
      });

      render(<SessionStatusToast position="top-left" />);

      const container = screen.getByRole('region', { name: 'Session notifications' });
      expect(container.className).toContain('top-4');
      expect(container.className).toContain('left-4');
    });
  });
});