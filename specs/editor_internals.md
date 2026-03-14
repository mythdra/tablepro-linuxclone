# Editor Internals (Monaco Editor + React)

## Overview
The SQL Editor uses **Monaco Editor** (`@monaco-editor/react`) — the same engine powering VS Code. This provides enterprise-grade editing capabilities out of the box.

## Key Capabilities
| Feature | Implementation |
|---|---|
| Syntax Highlighting | Monaco's built-in SQL language mode |
| Autocomplete | Custom `CompletionItemProvider` injecting table/column names from Go backend |
| Multi-cursor | Built-in (Cmd+Click / Alt+Click) |
| Vim Mode | `monaco-vim` npm package |
| Minimap | Built-in toggle |
| Find & Replace | Built-in (Cmd+F / Cmd+H) |
| Line Numbers | Built-in |
| Word Wrap | Configurable via settings |

## SQL Autocomplete Provider
```typescript
monaco.languages.registerCompletionItemProvider('sql', {
  provideCompletionItems: async (model, position) => {
    // Fetch schema from Go backend
    const tables = await GetTableNames(connectionId);
    const columns = await GetColumnNames(connectionId, currentTable);

    return {
      suggestions: [
        ...tables.map(t => ({
          label: t, kind: monaco.languages.CompletionItemKind.Class,
          insertText: t,
        })),
        ...columns.map(c => ({
          label: c, kind: monaco.languages.CompletionItemKind.Field,
          insertText: c,
        })),
        ...SQL_KEYWORDS.map(kw => ({
          label: kw, kind: monaco.languages.CompletionItemKind.Keyword,
          insertText: kw,
        })),
      ],
    };
  },
});
```

## Statement Execution Logic
1. User presses `Cmd+R`
2. React checks if there's a text selection → execute selected text only
3. If no selection, find the statement under cursor by parsing semicolons (respecting strings/comments)
4. Send statement to Go: `QueryManager.Execute(connectionID, tabID, sql)`
5. Go executes via driver, returns `QueryResult` with rows + columns
6. React updates AG Grid with result data

## Statement Splitting
Go backend provides `SplitStatements(sql string) []Statement` that:
- Splits on `;` outside of string literals, comments
- Returns each statement with its line number offset
- Used for "Execute All" (`Cmd+Shift+R`)

## Theme Customization
```typescript
monaco.editor.defineTheme('tablepro-dark', {
  base: 'vs-dark',
  inherit: true,
  rules: [
    { token: 'keyword', foreground: '#569CD6', fontStyle: 'bold' },
    { token: 'string', foreground: '#CE9178' },
    { token: 'comment', foreground: '#6A9955', fontStyle: 'italic' },
    { token: 'number', foreground: '#B5CEA8' },
  ],
  colors: {
    'editor.background': '#1E1E2E',
    'editor.foreground': '#CDD6F4',
  },
});
```
