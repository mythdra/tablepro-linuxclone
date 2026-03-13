# Editor Internals & Autocomplete

This document outlines the hidden complexities of the main SQL editor embedded within the query tabs. When migrating this to Qt (e.g., using `QQuickTextDocument`, `QScintilla`, or KTextEditor frameworks), these behaviors must be replicated identically.

## 1. Syntax Highlighting Strategy
The current application uses `CodeEditSourceEditor`, which relies on `Tree-sitter` for parsing. 

- **Performance Requirement**: Syntax highlighting *must* be asynchronous or heavily optimized. If a user pastes a 1-megabyte single-line SQL dump, naive regex highlighters will freeze the UI thread.
- **Line Wrapping**: Hard wraps must be disabled by default (horizontal scrolling enabled), but users can toggle soft-wrapping.

## 2. Autocomplete Engine (`CompletionEngine`)
Autocomplete suggestions are populated dynamically based on the connected database schema and dialect.

### Triggers & Debouncing
- Trigger Characters: `.` (dot) and ` ` (space).
- **Debounce**: 50ms. If the user types very fast, it abandons the previous AST parse request and only handles the latest.
- **Suppression Rules**: 
  - Never trigger immediately after a semicolon `;`.
  - Never trigger immediately after a newline `\n`.

### The 3-Tier Matching Algorithm
When the completion window is open and the user continues to type, the UI filters the existing items rather than re-querying the database.
Priority matching:
1. `hasPrefix`: Exact prefix match (`SEL` matches `SELECT`).
2. `contains`: Substring match (`ECT` matches `SELECT`).
3. `fuzzyMatch`: Non-contiguous character matching (e.g., `SCT` matches `SELECT`).

### Cursor Placement after Apply
When a user hits Enter/Tab to accept a suggestion:
- If the suggestion text ends with `()`, the cursor is automatically calculated to be placed *between* the parentheses (e.g., `COUNT(|)`).
- Otherwise, the cursor is placed at the very end of the inserted string.

## 3. Inline AI Suggestions (`InlineSuggestionManager`)
This feature mimics GitHub Copilot's ghost text.

1. **Trigger**: Triggered strictly on `textViewDidChangeSelection` or `textViewDidChangeText` (debounced by 250ms or similar).
2. **Context Gathering**: Sends the SQL schema (Tables, Columns) + the *entire* editor text before the cursor to the AI prompt.
3. **Display**: The returned string is rendered inline with a faded gray color. It is strictly visual (a separate layout layer or NSAttributedString specific attribute) and is *not* actually part of the text buffer.
4. **Acceptance**: Hitting `TAB` inserts the ghost text physically into the buffer and moves the cursor to the end.

## 4. Multi-Query Execution (`SQLStatementScanner`)
Users rarely select text before clicking "Run". Usually, they just put their cursor inside a block of text.
1. The engine scans the entire text buffer and tokenizes it by the `;` delimiter (ignoring semicolons inside string literals like `'John;Doe'`).
2. It calculates the byte ranges for every isolated statement.
3. It checks which statement range intersects with the current `CursorPosition`.
4. Only that isolated statement string is dispatched to the database execution engine.
