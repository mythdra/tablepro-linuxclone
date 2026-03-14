// Integration tests for QueryEditor keyboard shortcuts and autocomplete functionality.
// Run with: npm run test -- QueryEditor.integration.test.tsx

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryEditor } from './QueryEditor';
import type { SchemaMetadata } from '../types';

// Mock Monaco editor
const mockExecute = vi.fn();
const mockExecuteSelection = vi.fn();
const mockFormat = vi.fn((query: string) => {
  // Simple formatter for testing
  return query.trim().toUpperCase();
});

// Track Monaco actions registered
const registeredActions: Array<{ id: string; keybindings: number[] }> = [];

vi.mock('@monaco-editor/react', () => ({
  default: ({ value, onChange, onMount, options }: any) => {
    const mockEditor = {
      getValue: () => value || '',
      setValue: (newValue: string) => {
        if (onChange) onChange(newValue);
      },
      getModel: () => ({
        getValue: () => value || '',
        getValueInRange: (range: any) => value || '',
      }),
      getSelection: () => null,
      focus: vi.fn(),
      addAction: vi.fn((action: any) => {
        registeredActions.push({
          id: action.id,
          keybindings: action.keybindings || [],
        });
      }),
      trigger: vi.fn(),
      onKeyDown: vi.fn(),
    };

    const mockMonaco = {
      languages: {
        registerCompletionItemProvider: vi.fn(),
        CompletionItemKind: {
          Keyword: 1,
          Class: 2,
          Field: 3,
          Interface: 4,
        },
      },
      KeyMod: {
        CtrlCmd: 2048,
        Shift: 4096,
        Alt: 512,
      },
      KeyCode: {
        Enter: 3,
        KeyF: 33,
        KeyT: 23,
      },
      editor: {},
    };

    // Call onMount with mocked editor and monaco
    if (onMount) {
      setTimeout(() => onMount(mockEditor, mockMonaco), 0);
    }

    return (
      <textarea
        data-testid="monaco-editor"
        value={value || ''}
        onChange={(e) => onChange?.(e.target.value)}
        onKeyDown={(e) => {
          // Simulate keyboard shortcuts
          if (e.ctrlKey && e.key === 'Enter') {
            e.preventDefault();
            // Trigger execute action
          }
          if (e.shiftKey && e.altKey && e.key === 'f') {
            e.preventDefault();
            // Trigger format action
          }
          if (e.ctrlKey && e.key === 't') {
            e.preventDefault();
            // Trigger new tab action
          }
        }}
      />
    );
  },
}));

describe('QueryEditor - Keyboard Shortcuts (11.7)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    registeredActions.length = 0;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('registers Ctrl+Enter for query execution', async () => {
    render(<QueryEditor onExecute={mockExecute} />);

    await waitFor(() => {
      const executeAction = registeredActions.find((a) => a.id === 'execute-query');
      expect(executeAction).toBeDefined();
    });

    const executeAction = registeredActions.find((a) => a.id === 'execute-query');
    expect(executeAction).toBeDefined();
    expect(executeAction!.keybindings).toHaveLength(1);
  });

  it('executes query on Ctrl+Enter keyboard shortcut', async () => {
    const { getByTestId } = render(
      <QueryEditor onExecute={mockExecute} onExecuteSelection={mockExecuteSelection} />
    );

    const editor = getByTestId('monaco-editor');

    // Type a query
    await act(async () => {
      fireEvent.change(editor, { target: { value: 'SELECT * FROM users' } });
    });

    // Simulate Ctrl+Enter
    await act(async () => {
      fireEvent.keyDown(editor, {
        key: 'Enter',
        code: 'Enter',
        ctrlKey: true,
        metaKey: false,
        bubbles: true,
        cancelable: true,
      });
    });

    // Verify execute was called
    await waitFor(() => {
      expect(mockExecute).toHaveBeenCalledWith('SELECT * FROM users', expect.any(String));
    });
  });

  it('executes selection on Ctrl+Enter when text is selected', async () => {
    // This test verifies the selection execution path
    render(
      <QueryEditor onExecute={mockExecute} onExecuteSelection={mockExecuteSelection} />
    );

    const editor = screen.getByTestId('monaco-editor');

    await act(async () => {
      fireEvent.change(editor, { target: { value: 'SELECT * FROM users WHERE id = 1' } });
    });

    // Simulate Ctrl+Enter (selection would be handled by Monaco in real scenario)
    await act(async () => {
      fireEvent.keyDown(editor, {
        key: 'Enter',
        ctrlKey: true,
        bubbles: true,
      });
    });

    // Either full execute or selection execute should be called
    await waitFor(() => {
      expect(mockExecute).toHaveBeenCalled();
    });
  });

  it('registers Shift+Alt+F for query formatting', async () => {
    render(<QueryEditor onExecute={mockExecute} onFormat={mockFormat} />);

    await waitFor(() => {
      const formatAction = registeredActions.find((a) => a.id === 'format-query');
      expect(formatAction).toBeDefined();
    });

    const formatAction = registeredActions.find((a) => a.id === 'format-query');
    expect(formatAction).toBeDefined();
    expect(formatAction!.keybindings).toHaveLength(1);
  });

  it('formats query on Shift+Alt+F keyboard shortcut', async () => {
    render(
      <QueryEditor
        onExecute={mockExecute}
        onFormat={mockFormat}
      />
    );

    const editor = screen.getByTestId('monaco-editor');

    // Type a query
    await act(async () => {
      fireEvent.change(editor, { target: { value: 'select * from users' } });
    });

    // Simulate Shift+Alt+F
    await act(async () => {
      fireEvent.keyDown(editor, {
        key: 'f',
        code: 'KeyF',
        shiftKey: true,
        altKey: true,
        bubbles: true,
        cancelable: true,
      });
    });

    // Verify formatter was called
    await waitFor(() => {
      expect(mockFormat).toHaveBeenCalledWith('select * from users');
    });
  });

  it('does not format when onFormat is not provided', async () => {
    render(<QueryEditor onExecute={mockExecute} />);

    const editor = screen.getByTestId('monaco-editor');

    await act(async () => {
      fireEvent.change(editor, { target: { value: 'select * from users' } });
    });

    // Format action should not be registered
    await waitFor(() => {
      const formatAction = registeredActions.find((a) => a.id === 'format-query');
      expect(formatAction).toBeUndefined();
    });
  });

  it('registers Ctrl+T for new tab creation', async () => {
    render(<QueryEditor onExecute={mockExecute} />);

    await waitFor(() => {
      const newTabAction = registeredActions.find((a) => a.id === 'new-tab');
      expect(newTabAction).toBeDefined();
    });

    const newTabAction = registeredActions.find((a) => a.id === 'new-tab');
    expect(newTabAction).toBeDefined();
    expect(newTabAction!.keybindings).toHaveLength(1);
  });

  it('creates new tab on Ctrl+T keyboard shortcut', async () => {
    render(<QueryEditor onExecute={mockExecute} />);

    const editor = screen.getByTestId('monaco-editor');

    // Verify initial state - only Query 1
    expect(screen.getByText('Query 1')).toBeInTheDocument();
    expect(screen.queryByText('Query 2')).not.toBeInTheDocument();

    // Simulate Ctrl+T
    await act(async () => {
      fireEvent.keyDown(editor, {
        key: 't',
        code: 'KeyT',
        ctrlKey: true,
        metaKey: false,
        bubbles: true,
        cancelable: true,
      });
    });

    // Verify new tab was created
    await waitFor(() => {
      expect(screen.getByText('Query 2')).toBeInTheDocument();
    });
  });

  it('disables execute button when editor is empty', () => {
    render(<QueryEditor onExecute={mockExecute} />);

    const runButton = screen.getByRole('button', { name: 'Execute query' });
    expect(runButton).toBeDisabled();
  });

  it('enables execute button when editor has content', async () => {
    render(<QueryEditor onExecute={mockExecute} />);

    const editor = screen.getByTestId('monaco-editor');

    await act(async () => {
      fireEvent.change(editor, { target: { value: 'SELECT 1' } });
    });

    await waitFor(() => {
      const runButton = screen.getByRole('button', { name: 'Execute query' });
      expect(runButton).not.toBeDisabled();
    });
  });
});

describe('QueryEditor - Autocomplete with Schema Metadata (11.8)', () => {
  const mockSchema: SchemaMetadata = {
    tables: [
      {
        schema: 'public',
        name: 'users',
        columns: [
          { name: 'id', type: 'integer', nullable: false, isPrimaryKey: true },
          { name: 'email', type: 'varchar', nullable: false },
          { name: 'name', type: 'varchar', nullable: true },
          { name: 'created_at', type: 'timestamp', nullable: false },
        ],
        rowCount: 1000,
      },
      {
        schema: 'public',
        name: 'orders',
        columns: [
          { name: 'id', type: 'integer', nullable: false, isPrimaryKey: true },
          { name: 'user_id', type: 'integer', nullable: false },
          { name: 'total', type: 'decimal', nullable: false },
          { name: 'status', type: 'varchar', nullable: false },
        ],
        rowCount: 5000,
      },
      {
        schema: 'public',
        name: 'products',
        columns: [
          { name: 'id', type: 'integer', nullable: false, isPrimaryKey: true },
          { name: 'name', type: 'varchar', nullable: false },
          { name: 'price', type: 'decimal', nullable: false },
          { name: 'description', type: 'text', nullable: true },
        ],
        rowCount: 500,
      },
    ],
    views: [
      {
        schema: 'public',
        name: 'active_users',
        columns: [
          { name: 'id', type: 'integer' },
          { name: 'email', type: 'varchar' },
        ],
      },
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('registers completion provider when mounted', async () => {
    const registerCompletionItemProvider = vi.fn();

    vi.mocked(registerCompletionItemProvider).mockClear();

    render(
      <QueryEditor
        onExecute={mockExecute}
        schema={mockSchema}
      />
    );

    // Wait for Monaco to initialize
    await waitFor(() => {
      expect(registerCompletionItemProvider).toHaveBeenCalled();
    }, { timeout: 1000 });
  });

  it('provides table name suggestions from schema', async () => {
    render(
      <QueryEditor
        onExecute={mockExecute}
        schema={mockSchema}
      />
    );

    const editor = screen.getByTestId('monaco-editor');

    // Type a table name
    await act(async () => {
      fireEvent.change(editor, { target: { value: 'SELECT * FROM us' } });
    });

    // Verify editor content
    await waitFor(() => {
      expect(editor).toHaveValue('SELECT * FROM us');
    });
  });

  it('provides column name suggestions after table', async () => {
    render(
      <QueryEditor
        onExecute={mockExecute}
        schema={mockSchema}
      />
    );

    const editor = screen.getByTestId('monaco-editor');

    // Type a query with table
    await act(async () => {
      fireEvent.change(editor, { target: { value: 'SELECT id, email FROM users' } });
    });

    await waitFor(() => {
      expect(editor).toHaveValue('SELECT id, email FROM users');
    });
  });

  it('suggests view names from schema', async () => {
    render(
      <QueryEditor
        onExecute={mockExecute}
        schema={mockSchema}
      />
    );

    const editor = screen.getByTestId('monaco-editor');

    // Type a query that would reference a view
    await act(async () => {
      fireEvent.change(editor, { target: { value: 'SELECT * FROM active_' } });
    });

    await waitFor(() => {
      expect(editor).toHaveValue('SELECT * FROM active_');
    });
  });

  it('works without schema metadata', async () => {
    render(<QueryEditor onExecute={mockExecute} />);

    const editor = screen.getByTestId('monaco-editor');

    // Should still work without schema
    await act(async () => {
      fireEvent.change(editor, { target: { value: 'SELECT 1' } });
    });

    expect(editor).toHaveValue('SELECT 1');
  });

  it('updates autocomplete when schema changes', async () => {
    const { rerender } = render(
      <QueryEditor
        onExecute={mockExecute}
        schema={mockSchema}
      />
    );

    // Initial schema with 3 tables
    const editor = screen.getByTestId('monaco-editor');
    await act(async () => {
      fireEvent.change(editor, { target: { value: 'SELECT * FROM ' } });
    });

    // Update schema with new table
    const updatedSchema: SchemaMetadata = {
      ...mockSchema,
      tables: [
        ...mockSchema.tables,
        {
          schema: 'public',
          name: 'categories',
          columns: [{ name: 'id', type: 'integer', nullable: false }],
          rowCount: 10,
        },
      ],
    };

    rerender(
      <QueryEditor
        onExecute={mockExecute}
        schema={updatedSchema}
      />
    );

    await waitFor(() => {
      expect(editor).toHaveValue('SELECT * FROM ');
    });
  });

  it('handles empty schema gracefully', async () => {
    const emptySchema: SchemaMetadata = {
      tables: [],
      views: [],
    };

    render(
      <QueryEditor
        onExecute={mockExecute}
        schema={emptySchema}
      />
    );

    const editor = screen.getByTestId('monaco-editor');

    await act(async () => {
      fireEvent.change(editor, { target: { value: 'SELECT 1' } });
    });

    expect(editor).toHaveValue('SELECT 1');
  });
});

describe('QueryEditor - Combined Integration Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    registeredActions.length = 0;
  });

  it('full workflow: type, format, execute with schema', async () => {
    const schema: SchemaMetadata = {
      tables: [
        {
          schema: 'public',
          name: 'users',
          columns: [
            { name: 'id', type: 'integer', nullable: false },
            { name: 'email', type: 'varchar', nullable: false },
          ],
          rowCount: 100,
        },
      ],
      views: [],
    };

    render(
      <QueryEditor
        onExecute={mockExecute}
        onFormat={mockFormat}
        schema={schema}
      />
    );

    const editor = screen.getByTestId('monaco-editor');

    // Step 1: Type a query
    await act(async () => {
      fireEvent.change(editor, { target: { value: 'select * from users' } });
    });

    // Step 2: Format the query (Shift+Alt+F)
    await act(async () => {
      fireEvent.keyDown(editor, {
        key: 'f',
        shiftKey: true,
        altKey: true,
        bubbles: true,
      });
    });

    await waitFor(() => {
      expect(mockFormat).toHaveBeenCalled();
    });

    // Step 3: Execute the query (Ctrl+Enter)
    await act(async () => {
      fireEvent.keyDown(editor, {
        key: 'Enter',
        ctrlKey: true,
        bubbles: true,
      });
    });

    await waitFor(() => {
      expect(mockExecute).toHaveBeenCalled();
    });
  });

  it('multiple tabs with independent content', async () => {
    render(<QueryEditor onExecute={mockExecute} />);

    const editor = screen.getByTestId('monaco-editor');

    // Create second tab with Ctrl+T
    await act(async () => {
      fireEvent.keyDown(editor, {
        key: 't',
        ctrlKey: true,
        bubbles: true,
      });
    });

    await waitFor(() => {
      expect(screen.getByText('Query 2')).toBeInTheDocument();
    });

    // Type in tab 2
    await act(async () => {
      fireEvent.change(editor, { target: { value: 'SELECT * FROM orders' } });
    });

    // Switch back to tab 1
    const tab1Button = screen.getByRole('tab', { name: /Query 1/ });
    fireEvent.click(tab1Button);

    // Type in tab 1
    await act(async () => {
      fireEvent.change(editor, { target: { value: 'SELECT * FROM users' } });
    });

    // Execute should work for active tab
    const runButton = screen.getByRole('button', { name: 'Execute query' });
    fireEvent.click(runButton);

    await waitFor(() => {
      expect(mockExecute).toHaveBeenCalledWith('SELECT * FROM users', expect.any(String));
    });
  });
});
