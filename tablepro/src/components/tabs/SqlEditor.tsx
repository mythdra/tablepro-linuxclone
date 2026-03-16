import { useCallback, useEffect } from 'react';
import CodeMirror from '@uiw/react-codemirror';
import { sql, PostgreSQL } from '@codemirror/lang-sql';
import { EditorView } from '@codemirror/view';

interface SqlEditorProps {
  value: string;
  onChange: (value: string) => void;
  onExecute?: () => void;
}

// Custom dark theme matching Catppuccin Mocha
const tableproTheme = EditorView.theme({
  '&': {
    backgroundColor: 'hsl(223 28% 11%)',
    height: '100%',
  },
  '.cm-content': {
    fontFamily: "'JetBrains Mono', monospace",
    fontSize: '14px',
    caretColor: 'hsl(217 91% 65%)',
  },
  '.cm-cursor': {
    borderLeftColor: 'hsl(217 91% 65%)',
  },
  '.cm-gutters': {
    backgroundColor: 'hsl(223 28% 14%)',
    color: 'hsl(232 15% 58%)',
    border: 'none',
  },
  '.cm-activeLineGutter': {
    backgroundColor: 'hsl(223 20% 19%)',
  },
  '.cm-activeLine': {
    backgroundColor: 'hsl(223 20% 16%)',
  },
  '.cm-selectionBackground': {
    backgroundColor: 'hsl(223 20% 32%) !important',
  },
  '&.cm-focused .cm-selectionBackground': {
    backgroundColor: 'hsl(223 20% 32%) !important',
  },
  '.cm-line': {
    padding: '0 4px',
  },
});

export function SqlEditor({ value, onChange, onExecute }: SqlEditorProps) {
  const handleChange = useCallback(
    (val: string) => {
      onChange(val);
    },
    [onChange]
  );

  // Handle Ctrl+Enter / Cmd+Enter to execute
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        e.preventDefault();
        onExecute?.();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [onExecute]);

  return (
    <div className="h-full w-full overflow-hidden rounded-md border border-border">
      <CodeMirror
        value={value}
        height="100%"
        theme={tableproTheme}
        extensions={[
          sql({ dialect: PostgreSQL }),
          EditorView.lineWrapping,
        ]}
        onChange={handleChange}
        className="h-full text-sm"
        basicSetup={{
          lineNumbers: true,
          highlightActiveLineGutter: true,
          highlightActiveLine: true,
          foldGutter: true,
          dropCursor: true,
          allowMultipleSelections: true,
          indentOnInput: true,
          bracketMatching: true,
          closeBrackets: true,
          autocompletion: true,
          rectangularSelection: true,
          crosshairCursor: false,
          highlightSelectionMatches: true,
        }}
      />
    </div>
  );
}
