import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryEditor } from './QueryEditor';

const mockOnExecute = vi.fn();
const mockOnFormat = vi.fn((query: string) => query.toUpperCase());

vi.mock('@monaco-editor/react', () => ({
  default: ({ value, onChange }: { value: string; onChange?: (v: string) => void }) => (
    <textarea
      data-testid="monaco-editor"
      value={value}
      onChange={(e) => onChange?.(e.target.value)}
    />
  ),
}));

describe('QueryEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders editor with default tab', () => {
    render(
      <QueryEditor onExecute={mockOnExecute} />
    );

    expect(screen.getByTestId('monaco-editor')).toBeInTheDocument();
    expect(screen.getByText('Query 1')).toBeInTheDocument();
  });

  it('renders Run and Format buttons', () => {
    render(
      <QueryEditor onExecute={mockOnExecute} onFormat={mockOnFormat} />
    );

    expect(screen.getByRole('button', { name: 'Execute query' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Format query' })).toBeInTheDocument();
  });

  it('disables Run button when editor is empty', () => {
    render(
      <QueryEditor onExecute={mockOnExecute} />
    );

    const runButton = screen.getByRole('button', { name: 'Execute query' });
    expect(runButton).toBeDisabled();
  });

  it('enables Run button when editor has content', async () => {
    render(
      <QueryEditor onExecute={mockOnExecute} />
    );

    const editor = screen.getByTestId('monaco-editor');
    fireEvent.change(editor, { target: { value: 'SELECT * FROM users' } });

    await waitFor(() => {
      const runButton = screen.getByRole('button', { name: 'Execute query' });
      expect(runButton).not.toBeDisabled();
    });
  });

  it('creates new tab when + button is clicked', async () => {
    render(
      <QueryEditor onExecute={mockOnExecute} />
    );

    const newTabButton = screen.getByRole('button', { name: 'New query tab' });
    fireEvent.click(newTabButton);

    await waitFor(() => {
      expect(screen.getByText('Query 1')).toBeInTheDocument();
      expect(screen.getByText('Query 2')).toBeInTheDocument();
    });
  });

  it('switches between tabs', async () => {
    render(
      <QueryEditor onExecute={mockOnExecute} />
    );

    const newTabButton = screen.getByRole('button', { name: 'New query tab' });
    fireEvent.click(newTabButton);

    const tab2Button = screen.getByRole('tab', { name: /Query 2/ });
    fireEvent.click(tab2Button);

    await waitFor(() => {
      expect(tab2Button).toHaveAttribute('aria-selected', 'true');
    });
  });

  it('closes tab when close button is clicked', async () => {
    render(
      <QueryEditor onExecute={mockOnExecute} />
    );

    const newTabButton = screen.getByRole('button', { name: 'New query tab' });
    fireEvent.click(newTabButton);

    await waitFor(() => {
      expect(screen.getByText('Query 2')).toBeInTheDocument();
    });

    const closeButtons = screen.getAllByLabelText(/Close Query/);
    fireEvent.click(closeButtons[1]);

    await waitFor(() => {
      expect(screen.queryByText('Query 2')).not.toBeInTheDocument();
    });
  });

  it('does not close the last remaining tab', async () => {
    render(
      <QueryEditor onExecute={mockOnExecute} />
    );

    const closeButton = screen.getByLabelText('Close Query 1');
    fireEvent.click(closeButton);

    await waitFor(() => {
      expect(screen.getByText('Query 1')).toBeInTheDocument();
    });
  });

  it('shows dirty indicator when content changes', async () => {
    render(
      <QueryEditor onExecute={mockOnExecute} />
    );

    const editor = screen.getByTestId('monaco-editor');
    fireEvent.change(editor, { target: { value: 'SELECT 1' } });

    await waitFor(() => {
      const dirtyIndicators = screen.getAllByRole('button', { name: /Query 1/ });
      expect(dirtyIndicators[0]).toBeInTheDocument();
    });
  });

  it('renders with initial tabs', () => {
    const initialTabs = [
      {
        id: 'tab-1',
        name: 'Custom Query',
        content: 'SELECT * FROM products',
        isDirty: false,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
    ];

    render(
      <QueryEditor onExecute={mockOnExecute} initialTabs={initialTabs} />
    );

    expect(screen.getByText('Custom Query')).toBeInTheDocument();
    expect(screen.getByTestId('monaco-editor')).toHaveValue('SELECT * FROM products');
  });

  it('displays database type indicator when connection is active', () => {
    render(
      <QueryEditor
        onExecute={mockOnExecute}
        connectionId="conn-1"
        databaseType="mysql"
      />
    );

    expect(screen.getByText('MYSQL')).toBeInTheDocument();
  });
});

describe('TabBar', () => {
  const mockTabs = [
    { id: '1', name: 'Query 1', content: '', isDirty: false, createdAt: '', updatedAt: '' },
    { id: '2', name: 'Query 2', content: 'SELECT 1', isDirty: true, createdAt: '', updatedAt: '' },
  ];

  it('renders all tabs', () => {
    render(
      <QueryEditor onExecute={mockOnExecute} initialTabs={mockTabs} />
    );

    expect(screen.getByText('Query 1')).toBeInTheDocument();
    expect(screen.getByText('Query 2')).toBeInTheDocument();
  });

  it('highlights active tab', () => {
    render(
      <QueryEditor onExecute={mockOnExecute} initialTabs={mockTabs} />
    );

    const activeTab = screen.getByRole('tab', { name: /Query 1/ });
    expect(activeTab).toHaveAttribute('aria-selected', 'true');
  });

  it('calls onTabSelect when tab is clicked', async () => {
    render(
      <QueryEditor onExecute={mockOnExecute} initialTabs={mockTabs} />
    );

    const tab2 = screen.getByRole('tab', { name: /Query 2/ });
    fireEvent.click(tab2);

    await waitFor(() => {
      expect(tab2).toHaveAttribute('aria-selected', 'true');
    });
  });
});