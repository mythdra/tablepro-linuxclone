import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { HistoryPanel } from './HistoryPanel';
import { HistoryEntry } from './HistoryEntry';
import type { QueryHistoryEntry } from '../types';

// Helper to create mock history entries
function createMockEntry(overrides: Partial<QueryHistoryEntry> = {}): QueryHistoryEntry {
  return {
    id: `entry-${Math.random().toString(36).slice(2)}`,
    connectionId: 'conn-1',
    query: 'SELECT * FROM users',
    executedAt: new Date().toISOString(),
    duration: 150,
    rowCount: 100,
    status: 'success',
    ...overrides,
  };
}

describe('HistoryPanel', () => {
  const mockOnLoadQuery = vi.fn();
  const mockOnClearHistory = vi.fn();
  const connectionId = 'conn-1';

  beforeEach(() => {
    mockOnLoadQuery.mockClear();
    mockOnClearHistory.mockClear();
  });

  describe('empty state', () => {
    it('should render empty state when no entries', () => {
      render(
        <HistoryPanel
          connectionId={connectionId}
          entries={[]}
          onLoadQuery={mockOnLoadQuery}
          onClearHistory={mockOnClearHistory}
        />
      );

      expect(screen.getByText('No query history yet')).toBeInTheDocument();
      expect(screen.getByText('Execute queries to see them here')).toBeInTheDocument();
    });

    it('should not show clear button when empty', () => {
      render(
        <HistoryPanel
          connectionId={connectionId}
          entries={[]}
          onLoadQuery={mockOnLoadQuery}
          onClearHistory={mockOnClearHistory}
        />
      );

      expect(screen.queryByRole('button', { name: /clear all history/i })).not.toBeInTheDocument();
    });
  });

  describe('with entries', () => {
    const mockEntries: QueryHistoryEntry[] = [
      createMockEntry({ id: 'entry-1', query: 'SELECT * FROM users', duration: 150 }),
      createMockEntry({ id: 'entry-2', query: 'SELECT * FROM products', duration: 200 }),
      createMockEntry({ id: 'entry-3', query: 'SELECT * FROM orders', duration: 50 }),
    ];

    it('should render all entries', () => {
      render(
        <HistoryPanel
          connectionId={connectionId}
          entries={mockEntries}
          onLoadQuery={mockOnLoadQuery}
          onClearHistory={mockOnClearHistory}
        />
      );

      expect(screen.getByText(/SELECT \* FROM users/)).toBeInTheDocument();
      expect(screen.getByText(/SELECT \* FROM products/)).toBeInTheDocument();
      expect(screen.getByText(/SELECT \* FROM orders/)).toBeInTheDocument();
    });

    it('should show query count in footer', () => {
      render(
        <HistoryPanel
          connectionId={connectionId}
          entries={mockEntries}
          onLoadQuery={mockOnLoadQuery}
          onClearHistory={mockOnClearHistory}
        />
      );

      expect(screen.getByText('3 queries')).toBeInTheDocument();
    });

    it('should call onLoadQuery when entry is clicked', async () => {
      render(
        <HistoryPanel
          connectionId={connectionId}
          entries={mockEntries}
          onLoadQuery={mockOnLoadQuery}
          onClearHistory={mockOnClearHistory}
        />
      );

      const firstEntry = screen.getByText(/SELECT \* FROM users/);
      fireEvent.click(firstEntry);

      expect(mockOnLoadQuery).toHaveBeenCalledWith('SELECT * FROM users');
    });
  });

  describe('search functionality', () => {
    const mockEntries: QueryHistoryEntry[] = [
      createMockEntry({ id: 'entry-1', query: 'SELECT * FROM users WHERE id = 1' }),
      createMockEntry({ id: 'entry-2', query: 'SELECT * FROM products' }),
      createMockEntry({ id: 'entry-3', query: 'UPDATE users SET name = "test"' }),
    ];

    it('should filter entries by search query', async () => {
      render(
        <HistoryPanel
          connectionId={connectionId}
          entries={mockEntries}
          onLoadQuery={mockOnLoadQuery}
          onClearHistory={mockOnClearHistory}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search queries...');
      fireEvent.change(searchInput, { target: { value: 'users' } });

      // Wait for debounce (300ms)
      await waitFor(() => {
        expect(screen.getByText(/SELECT \* FROM users WHERE id = 1/)).toBeInTheDocument();
        expect(screen.getByText(/UPDATE users SET name/)).toBeInTheDocument();
        expect(screen.queryByText(/SELECT \* FROM products/)).not.toBeInTheDocument();
      }, { timeout: 500 });
    });

    it('should show "no matching queries" when search has no results', async () => {
      render(
        <HistoryPanel
          connectionId={connectionId}
          entries={mockEntries}
          onLoadQuery={mockOnLoadQuery}
          onClearHistory={mockOnClearHistory}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search queries...');
      fireEvent.change(searchInput, { target: { value: 'nonexistent' } });

      await waitFor(() => {
        expect(screen.getByText('No matching queries found')).toBeInTheDocument();
      }, { timeout: 500 });
    });

    it('should clear search when clear button is clicked', async () => {
      render(
        <HistoryPanel
          connectionId={connectionId}
          entries={mockEntries}
          onLoadQuery={mockOnLoadQuery}
          onClearHistory={mockOnClearHistory}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search queries...') as HTMLInputElement;
      fireEvent.change(searchInput, { target: { value: 'users' } });

      // Wait for debounce
      await waitFor(() => {
        expect(searchInput.value).toBe('users');
      });

      // Click clear button
      const clearButton = screen.getByRole('button', { name: /clear search/i });
      fireEvent.click(clearButton);

      expect(searchInput.value).toBe('');
    });

    it('should show filtered count in footer', async () => {
      render(
        <HistoryPanel
          connectionId={connectionId}
          entries={mockEntries}
          onLoadQuery={mockOnLoadQuery}
          onClearHistory={mockOnClearHistory}
        />
      );

      const searchInput = screen.getByPlaceholderText('Search queries...');
      fireEvent.change(searchInput, { target: { value: 'users' } });

      await waitFor(() => {
        expect(screen.getByText('2 of 3 queries')).toBeInTheDocument();
      }, { timeout: 500 });
    });
  });

  describe('clear history', () => {
    const mockEntries: QueryHistoryEntry[] = [
      createMockEntry({ id: 'entry-1' }),
      createMockEntry({ id: 'entry-2' }),
    ];

    it('should show confirmation dialog when clear is clicked', () => {
      render(
        <HistoryPanel
          connectionId={connectionId}
          entries={mockEntries}
          onLoadQuery={mockOnLoadQuery}
          onClearHistory={mockOnClearHistory}
        />
      );

      const clearButton = screen.getByRole('button', { name: /clear all history/i });
      fireEvent.click(clearButton);

      expect(screen.getByText('Clear all history?')).toBeInTheDocument();
      expect(screen.getByText(/This will remove all query history/)).toBeInTheDocument();
    });

    it('should call onClearHistory when confirmed', () => {
      render(
        <HistoryPanel
          connectionId={connectionId}
          entries={mockEntries}
          onLoadQuery={mockOnLoadQuery}
          onClearHistory={mockOnClearHistory}
        />
      );

      const clearButton = screen.getByRole('button', { name: /clear all history/i });
      fireEvent.click(clearButton);

      const confirmButton = screen.getByRole('button', { name: /confirm clear history/i });
      fireEvent.click(confirmButton);

      expect(mockOnClearHistory).toHaveBeenCalledWith(connectionId);
    });

    it('should cancel clear when cancel is clicked', () => {
      render(
        <HistoryPanel
          connectionId={connectionId}
          entries={mockEntries}
          onLoadQuery={mockOnLoadQuery}
          onClearHistory={mockOnClearHistory}
        />
      );

      const clearButton = screen.getByRole('button', { name: /clear all history/i });
      fireEvent.click(clearButton);

      const cancelButton = screen.getByRole('button', { name: /cancel clear history/i });
      fireEvent.click(cancelButton);

      expect(mockOnClearHistory).not.toHaveBeenCalled();
      expect(screen.queryByText('Clear all history?')).not.toBeInTheDocument();
    });

    it('should close confirmation dialog on Escape key', () => {
      render(
        <HistoryPanel
          connectionId={connectionId}
          entries={mockEntries}
          onLoadQuery={mockOnLoadQuery}
          onClearHistory={mockOnClearHistory}
        />
      );

      const clearButton = screen.getByRole('button', { name: /clear all history/i });
      fireEvent.click(clearButton);

      expect(screen.getByText('Clear all history?')).toBeInTheDocument();

      fireEvent.keyDown(window, { key: 'Escape' });

      expect(screen.queryByText('Clear all history?')).not.toBeInTheDocument();
    });
  });

  describe('accessibility', () => {
    const mockEntries: QueryHistoryEntry[] = [createMockEntry()];

    it('should have proper ARIA labels', () => {
      render(
        <HistoryPanel
          connectionId={connectionId}
          entries={mockEntries}
          onLoadQuery={mockOnLoadQuery}
          onClearHistory={mockOnClearHistory}
        />
      );

      expect(screen.getByRole('region', { name: 'Query History' })).toBeInTheDocument();
      expect(screen.getByRole('list', { name: 'Query history entries' })).toBeInTheDocument();
      expect(screen.getByLabelText('Search query history')).toBeInTheDocument();
    });
  });
});

describe('HistoryEntry', () => {
  const mockOnClick = vi.fn();

  beforeEach(() => {
    mockOnClick.mockClear();
  });

  describe('success entry', () => {
    it('should render success icon for successful queries', () => {
      const entry = createMockEntry({ status: 'success' });
      render(<HistoryEntry entry={entry} onClick={mockOnClick} />);

      expect(screen.getByLabelText('Success')).toBeInTheDocument();
    });

    it('should show row count for successful queries', () => {
      const entry = createMockEntry({ status: 'success', rowCount: 500 });
      render(<HistoryEntry entry={entry} onClick={mockOnClick} />);

      expect(screen.getByText('500 rows')).toBeInTheDocument();
    });
  });

  describe('error entry', () => {
    it('should render error icon for failed queries', () => {
      const entry = createMockEntry({ status: 'error' });
      render(<HistoryEntry entry={entry} onClick={mockOnClick} />);

      expect(screen.getByLabelText('Error')).toBeInTheDocument();
    });

    it('should show error message for failed queries', () => {
      const entry = createMockEntry({
        status: 'error',
        error: 'Relation "users" does not exist'
      });
      render(<HistoryEntry entry={entry} onClick={mockOnClick} />);

      expect(screen.getByText(/Relation "users" does not exist/)).toBeInTheDocument();
    });
  });

  describe('cancelled entry', () => {
    it('should render cancelled icon for cancelled queries', () => {
      const entry = createMockEntry({ status: 'cancelled' });
      render(<HistoryEntry entry={entry} onClick={mockOnClick} />);

      expect(screen.getByLabelText('Cancelled')).toBeInTheDocument();
    });
  });

  describe('duration formatting', () => {
    it('should format milliseconds correctly', () => {
      const entry = createMockEntry({ duration: 150 });
      render(<HistoryEntry entry={entry} onClick={mockOnClick} />);

      expect(screen.getByText('150ms')).toBeInTheDocument();
    });

    it('should format seconds correctly', () => {
      const entry = createMockEntry({ duration: 1500 });
      render(<HistoryEntry entry={entry} onClick={mockOnClick} />);

      expect(screen.getByText('1.5s')).toBeInTheDocument();
    });

    it('should format minutes correctly', () => {
      const entry = createMockEntry({ duration: 90000 }); // 1m 30s
      render(<HistoryEntry entry={entry} onClick={mockOnClick} />);

      expect(screen.getByText('1m 30s')).toBeInTheDocument();
    });
  });

  describe('timestamp formatting', () => {
    it('should show "just now" for recent entries', () => {
      const entry = createMockEntry({ executedAt: new Date().toISOString() });
      render(<HistoryEntry entry={entry} onClick={mockOnClick} />);

      expect(screen.getByText('just now')).toBeInTheDocument();
    });

    it('should show "X min ago" for entries within an hour', () => {
      const fiveMinutesAgo = new Date(Date.now() - 5 * 60 * 1000).toISOString();
      const entry = createMockEntry({ executedAt: fiveMinutesAgo });
      render(<HistoryEntry entry={entry} onClick={mockOnClick} />);

      expect(screen.getByText('5 min ago')).toBeInTheDocument();
    });
  });

  describe('query truncation', () => {
    it('should truncate long queries', () => {
      const longQuery = 'SELECT * FROM very_long_table_name WHERE very_long_column_name = "some_value" AND another_column = "another_value"';
      const entry = createMockEntry({ query: longQuery });
      render(<HistoryEntry entry={entry} onClick={mockOnClick} />);

      // The query should be truncated with "..."
      expect(screen.getByText(/\.\.\./)).toBeInTheDocument();
    });
  });

  describe('click behavior', () => {
    it('should call onClick with entry when clicked', () => {
      const entry = createMockEntry({ query: 'SELECT 1' });
      render(<HistoryEntry entry={entry} onClick={mockOnClick} />);

      fireEvent.click(screen.getByRole('button'));

      expect(mockOnClick).toHaveBeenCalledWith(entry);
    });
  });
});