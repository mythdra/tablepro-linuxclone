import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook } from '@testing-library/react';

vi.mock('../wailsjs/runtime', () => ({
  EventsOn: vi.fn(() => () => {}),
  EventsOff: vi.fn(),
}));

import { useQueryEvents } from './useQueryEvents';
import { useQueryStore } from '../stores/queryStore';

import { EventsOn } from '../wailsjs/runtime';

describe('useQueryEvents', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useQueryStore.getState().clearHistory();
    useQueryStore.getState().clearActiveQueries();
  });

  it('should subscribe to query:executing event', () => {
    renderHook(() => useQueryEvents());

    expect(EventsOn).toHaveBeenCalledWith('query:executing', expect.any(Function));
  });

  it('should subscribe to query:completed event', () => {
    renderHook(() => useQueryEvents());

    expect(EventsOn).toHaveBeenCalledWith('query:completed', expect.any(Function));
  });

  it('should subscribe to query:failed event', () => {
    renderHook(() => useQueryEvents());

    expect(EventsOn).toHaveBeenCalledWith('query:failed', expect.any(Function));
  });

  it('should subscribe to history:added event', () => {
    renderHook(() => useQueryEvents());

    expect(EventsOn).toHaveBeenCalledWith('history:added', expect.any(Function));
  });

  it('should unsubscribe from all events on unmount', async () => {
    const { unmount } = renderHook(() => useQueryEvents());

    unmount();

    expect(EventsOn).toHaveBeenCalledTimes(4);
  });
});
