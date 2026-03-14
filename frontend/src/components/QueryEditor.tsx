import { useState, useCallback, useRef, useEffect } from 'react';
import Editor, { OnMount, OnChange } from '@monaco-editor/react';
import type { editor, languages, Position, IRange } from 'monaco-editor';
import { Play, AlignLeft, Database } from 'lucide-react';
import { TabBar } from './TabBar';
import type { QueryTab, SchemaMetadata, DatabaseType } from '../types';
import { v4 as uuidv4 } from 'uuid';

/**
 * Props for the QueryEditor component.
 */
interface QueryEditorProps {
  /** Optional connection ID for database context */
  connectionId?: string;
  /** Database type for SQL dialect selection (default: 'postgres') */
  databaseType?: DatabaseType;
  /** Schema metadata for autocomplete suggestions */
  schema?: SchemaMetadata;
  /** Callback when full query execution is requested */
  onExecute: (query: string, tabId: string) => void;
  /** Callback when selected text execution is requested */
  onExecuteSelection?: (query: string, tabId: string) => void;
  /** Optional formatter function for SQL beautification */
  onFormat?: (query: string) => string;
  /** Initial tabs to load (creates new tab if empty) */
  initialTabs?: QueryTab[];
}

/** Common SQL keywords for autocomplete suggestions */
const SQL_KEYWORDS = [
  'SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'NOT', 'IN', 'IS', 'NULL', 'AS',
  'JOIN', 'LEFT', 'RIGHT', 'INNER', 'OUTER', 'ON', 'GROUP', 'BY', 'HAVING',
  'ORDER', 'ASC', 'DESC', 'LIMIT', 'OFFSET', 'INSERT', 'INTO', 'VALUES',
  'UPDATE', 'SET', 'DELETE', 'CREATE', 'TABLE', 'DROP', 'ALTER', 'ADD',
  'COLUMN', 'INDEX', 'VIEW', 'DISTINCT', 'COUNT', 'SUM', 'AVG', 'MIN', 'MAX',
  'CASE', 'WHEN', 'THEN', 'ELSE', 'END', 'UNION', 'ALL', 'EXISTS', 'BETWEEN',
  'LIKE', 'CAST', 'COALESCE', 'NULLIF', 'EXTRACT', 'CURRENT_DATE', 'CURRENT_TIME',
  'CURRENT_TIMESTAMP', 'PRIMARY', 'KEY', 'FOREIGN', 'REFERENCES', 'CONSTRAINT',
  'UNIQUE', 'DEFAULT', 'CHECK', 'GRANT', 'REVOKE', 'COMMIT', 'ROLLBACK',
  'TRANSACTION', 'BEGIN', 'DECLARE', 'IF', 'WHILE', 'FOR', 'LOOP', 'RETURN',
  'FUNCTION', 'PROCEDURE', 'TRIGGER', 'CASCADE', 'RESTRICT', 'TRUNCATE',
];

/**
 * Creates a new query tab with default values.
 * @param index - Tab index for naming
 * @returns New query tab object
 */
function createNewTab(index: number): QueryTab {
  return {
    id: uuidv4(),
    name: `Query ${index}`,
    content: '',
    isDirty: false,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  };
}

/**
 * QueryEditor component provides a Monaco-based SQL editor with tab management.
 * 
 * Features:
 * - Multiple query tabs with dirty state tracking
 * - SQL autocomplete with schema-aware suggestions
 * - Execute full query or selected text (Ctrl/Cmd+Enter)
 * - Query formatting support (Shift+Alt+F)
 * - Keyboard shortcuts for common actions
 * 
 * @example
 * ```tsx
 * <QueryEditor
 *   connectionId="conn-123"
 *   databaseType="postgres"
 *   schema={schemaMetadata}
 *   onExecute={handleExecute}
 *   onFormat={formatSQL}
 * />
 * ```
 */
export function QueryEditor({
  connectionId,
  databaseType = 'postgres',
  schema,
  onExecute,
  onExecuteSelection,
  onFormat,
  initialTabs,
}: QueryEditorProps) {
  const [tabs, setTabs] = useState<QueryTab[]>(() => {
    return initialTabs && initialTabs.length > 0 ? initialTabs : [createNewTab(1)];
  });
  const [activeTabId, setActiveTabId] = useState<string>(() => tabs[0].id);
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null);
  const monacoRef = useRef<typeof import('monaco-editor') | null>(null);

  const activeTab = tabs.find((t) => t.id === activeTabId) ?? tabs[0];

  const getTabLanguage = useCallback(() => {
    return databaseType === 'mysql' ? 'mysql' : 'pgsql';
  }, [databaseType]);

  const handleEditorMount: OnMount = (editor, monaco) => {
    editorRef.current = editor;
    monacoRef.current = monaco;

    monaco.languages.registerCompletionItemProvider('sql', {
      provideCompletionItems: (model: editor.ITextModel, position: Position) => {
        const word = model.getWordUntilPosition(position);
        const range: IRange = {
          startLineNumber: position.lineNumber,
          endLineNumber: position.lineNumber,
          startColumn: word.startColumn,
          endColumn: word.endColumn,
        };

        const suggestions: languages.CompletionItem[] = [];

        SQL_KEYWORDS.forEach((keyword) => {
          suggestions.push({
            label: keyword,
            kind: monaco.languages.CompletionItemKind.Keyword,
            insertText: keyword,
            range,
          });
        });

        if (schema) {
          schema.tables.forEach((table) => {
            suggestions.push({
              label: table.name,
              kind: monaco.languages.CompletionItemKind.Class,
              insertText: table.name,
              range,
              detail: `Table: ${table.schema}.${table.name}`,
            });

            table.columns.forEach((column) => {
              suggestions.push({
                label: column.name,
                kind: monaco.languages.CompletionItemKind.Field,
                insertText: column.name,
                range,
                detail: `Column: ${column.type}`,
                documentation: `${table.name}.${column.name}`,
              });
            });
          });

          schema.views.forEach((view) => {
            suggestions.push({
              label: view.name,
              kind: monaco.languages.CompletionItemKind.Interface,
              insertText: view.name,
              range,
              detail: `View: ${view.schema}.${view.name}`,
            });
          });
        }

        return { suggestions };
      },
    });

    editor.addAction({
      id: 'execute-query',
      label: 'Execute Query',
      keybindings: [
        monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter,
      ],
      run: () => {
        const selection = editor.getSelection();
        const model = editor.getModel();
        if (!model) return;

        let query: string;
        if (selection && !selection.isEmpty()) {
          query = model.getValueInRange(selection);
          onExecuteSelection?.(query, activeTabId);
        } else {
          query = model.getValue();
          onExecute(query, activeTabId);
        }
      },
    });

    editor.addAction({
      id: 'format-query',
      label: 'Format Query',
      keybindings: [
        monaco.KeyMod.Shift | monaco.KeyMod.Alt | monaco.KeyCode.KeyF,
      ],
      run: () => {
        if (onFormat) {
          const model = editor.getModel();
          if (!model) return;
          const formatted = onFormat(model.getValue());
          editor.setValue(formatted);
        }
      },
    });

    editor.addAction({
      id: 'new-tab',
      label: 'New Query Tab',
      keybindings: [
        monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyT,
      ],
      run: () => {
        handleNewTab();
      },
    });

    editor.focus();
  };

  const handleEditorChange: OnChange = (value) => {
    setTabs((prev) =>
      prev.map((tab) =>
        tab.id === activeTabId
          ? {
              ...tab,
              content: value ?? '',
              isDirty: true,
              updatedAt: new Date().toISOString(),
            }
          : tab
      )
    );
  };

  const handleTabSelect = useCallback((tabId: string) => {
    setActiveTabId(tabId);
  }, []);

  const handleTabClose = useCallback((tabId: string) => {
    setTabs((prev) => {
      if (prev.length <= 1) return prev;

      const newTabs = prev.filter((t) => t.id !== tabId);

      if (tabId === activeTabId) {
        const closedIndex = prev.findIndex((t) => t.id === tabId);
        const newActiveIndex = Math.min(closedIndex, newTabs.length - 1);
        setActiveTabId(newTabs[newActiveIndex].id);
      }

      return newTabs;
    });
  }, [activeTabId]);

  const handleNewTab = useCallback(() => {
    const newTab = createNewTab(tabs.length + 1);
    setTabs((prev) => [...prev, newTab]);
    setActiveTabId(newTab.id);
  }, [tabs.length]);

  const handleExecuteClick = useCallback(() => {
    const editor = editorRef.current;
    if (!editor) return;

    const model = editor.getModel();
    if (!model) return;

    const selection = editor.getSelection();
    let query: string;

    if (selection && !selection.isEmpty()) {
      query = model.getValueInRange(selection);
      onExecuteSelection?.(query, activeTabId);
    } else {
      query = model.getValue();
      onExecute(query, activeTabId);
    }
  }, [activeTabId, onExecute, onExecuteSelection]);

  const handleFormatClick = useCallback(() => {
    if (!onFormat) return;
    const editor = editorRef.current;
    if (!editor) return;

    const model = editor.getModel();
    if (!model) return;

    const formatted = onFormat(model.getValue());
    editor.setValue(formatted);
  }, [onFormat]);

  useEffect(() => {
    if (editorRef.current) {
      editorRef.current.focus();
    }
  }, [activeTabId]);

  return (
    <div className="h-full flex flex-col bg-slate-900">
      <TabBar
        tabs={tabs}
        activeTabId={activeTabId}
        onTabSelect={handleTabSelect}
        onTabClose={handleTabClose}
        onNewTab={handleNewTab}
      />

      <div className="flex items-center gap-2 px-4 py-2 bg-slate-800 border-b border-slate-700">
        <button
          onClick={handleExecuteClick}
          disabled={!activeTab.content.trim()}
          className="flex items-center gap-2 px-4 py-1.5 text-sm font-medium text-white bg-primary rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          aria-label="Execute query"
        >
          <Play className="w-4 h-4" />
          Run
        </button>
        <button
          onClick={handleFormatClick}
          disabled={!onFormat || !activeTab.content.trim()}
          className="flex items-center gap-2 px-3 py-1.5 text-sm font-medium text-slate-300 bg-slate-700 rounded-lg hover:bg-slate-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          aria-label="Format query"
        >
          <AlignLeft className="w-4 h-4" />
          Format
        </button>
        <div className="flex-1" />
        {connectionId && (
          <div className="flex items-center gap-2 text-xs text-slate-400">
            <Database className="w-4 h-4" />
            <span>{databaseType.toUpperCase()}</span>
          </div>
        )}
      </div>

      <div className="flex-1 min-h-0">
        <Editor
          height="100%"
          language={getTabLanguage()}
          value={activeTab.content}
          onChange={handleEditorChange}
          onMount={handleEditorMount}
          theme="vs-dark"
          options={{
            fontSize: 14,
            fontFamily: "'Fira Code', 'Cascadia Code', Consolas, monospace",
            fontLigatures: true,
            minimap: { enabled: false },
            scrollBeyondLastLine: false,
            automaticLayout: true,
            tabSize: 2,
            wordWrap: 'on',
            lineNumbers: 'on',
            renderLineHighlight: 'all',
            cursorBlinking: 'smooth',
            cursorSmoothCaretAnimation: 'on',
            padding: { top: 8, bottom: 8 },
            suggest: {
              showKeywords: true,
              showSnippets: true,
            },
          }}
        />
      </div>
    </div>
  );
}

export default QueryEditor;