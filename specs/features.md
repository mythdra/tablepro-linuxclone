# Features Specification (C++20 + Qt 6.6)

Exhaustive list of every discrete feature the new TablePro must implement.

## 1. Connection Management
- [ ] Create, edit, duplicate, delete database connections
- [ ] Connection groups and color tags
- [ ] Test connection before saving
- [ ] Import connection from URL string (including `scheme+ssh://` format)
- [ ] Deep linking: `tablepro://` URL scheme opens connections
- [ ] Secure password storage via QKeychain
- [ ] SSH tunnel setup (password, key file, SSH agent) using libssh2 + QtNetwork
- [ ] SSL/TLS configuration (CA cert, client cert, client key)
- [ ] Startup commands (executed after connect)
- [ ] Safe Mode levels (require WHERE clause for UPDATE/DELETE)
- [ ] `.pgpass` file detection and permission warning

## 2. Database Navigation (Sidebar)
- [ ] Database/schema tree with lazy loading (QTreeView)
- [ ] Tables, Views, Routines as collapsible groups
- [ ] Real-time search filtering
- [ ] Right-click context menu: Open, Copy Name, Drop, Truncate, Show DDL
- [ ] Batch operations: multi-select drop/truncate with confirmation
- [ ] Visual indicators for table types (icon differentiation)
- [ ] Database switcher dropdown in toolbar

## 3. SQL Editor
- [ ] QScintilla editor with SQL syntax highlighting
- [ ] Autocomplete: table names, column names, SQL keywords
- [ ] Multi-cursor editing (Ctrl+Click)
- [ ] Execute current statement (Ctrl+R / F5)
- [ ] Execute all statements (Ctrl+Shift+R)
- [ ] Execute selected text only
- [ ] Statement splitting (semicolon-aware, respecting strings/comments)
- [ ] Vim mode toggle (QScintilla Vim extension)
- [ ] SQL formatting / beautification
- [ ] Line numbers, minimap
- [ ] Auto-capitalize SQL keywords
- [ ] Find and Replace (Ctrl+F)

## 4. Data Grid (Query Results)
- [ ] Virtual scrolling via QTableView + custom QAbstractTableModel (millions of rows)
- [ ] Column resizing, reordering, hiding
- [ ] Click header to sort (re-executes with ORDER BY)
- [ ] Row numbering gutter
- [ ] Column type-aware rendering (dates, JSON, binary, null)
- [ ] Copy cell / row / column as text
- [ ] Inline cell editing (double-click → delegates)
- [ ] Add new row (+ button)
- [ ] Delete row (mark for deletion with visual strikethrough)
- [ ] Visual change indicators (yellow=edited, green=new, red=deleted)
- [ ] Commit changes (generates INSERT/UPDATE/DELETE SQL)
- [ ] Discard changes
- [ ] Undo/Redo for cell edits
- [ ] Pagination with offset/limit controls
- [ ] Row count display and execution time

## 5. Tab Management
- [ ] Horizontal tab bar with close buttons (QTabBar)
- [ ] New tab (Ctrl+T)
- [ ] Close tab (Ctrl+W)
- [ ] Switch tabs (Ctrl+1..9)
- [ ] Tab types: Query, Table, Structure
- [ ] Tab state persistence across app restarts (JSON files)
- [ ] LRU memory eviction for inactive tabs
- [ ] Lazy re-query when switching back to evicted tab
- [ ] Preview tabs (single-click on sidebar item)

## 6. Export
- [ ] Export to CSV, JSON, SQL, XLSX, Markdown
- [ ] Export current query results or entire table
- [ ] Streaming export for large datasets
- [ ] Configurable options per format (headers, separator, encoding)
- [ ] Native file save dialog via QFileDialog

## 7. Import
- [ ] Import from SQL dump files (.sql, .sql.gz)
- [ ] Streaming 64KB chunk parser (handles multi-GB files)
- [ ] Automatic gzip decompression
- [ ] Progress bar with cancellation
- [ ] Transaction wrapping with foreign key disable
- [ ] Error reporting with line number

## 8. Table Structure View
- [ ] Columns tab: name, type, nullable, auto-increment, default, comment
- [ ] Indexes tab: index name, columns, unique flag
- [ ] Foreign Keys tab: constraint name, column, referenced table/column
- [ ] DDL tab: raw CREATE TABLE statement

## 9. Query History
- [ ] Full-text search over executed queries (SQLite FTS5 via Qt SQL)
- [ ] Filter by connection, database, success/failure
- [ ] Click to re-execute a historical query
- [ ] Auto-cleanup of old history (configurable retention)

## 10. AI Integration
- [ ] Chat panel for query generation and explanation
- [ ] Support OpenAI, Anthropic, Ollama providers
- [ ] Streaming markdown response rendering
- [ ] Context-aware: inject current table schema into prompts
- [ ] API key stored securely in QKeychain

## 11. Quick Switcher (Command Palette)
- [ ] Ctrl+K opens floating search (QCompleter + QLineEdit overlay)
- [ ] Search tables, views, routines across all schemas
- [ ] Fuzzy matching
- [ ] Navigate to selected item

## 12. Settings
- [ ] Theme: system, light, dark
- [ ] Editor font family and size
- [ ] Line wrapping, line numbers
- [ ] Query timeout
- [ ] Rows per page (pagination default)
- [ ] Vim mode toggle
- [ ] Auto-capitalize keywords toggle
- [ ] Autocomplete toggle

## 13. Licensing
- [ ] Free vs Pro feature gating
- [ ] License key input and validation
- [ ] Ed25519 signature verification
- [ ] Lemon Squeezy or custom backend integration

## 14. Cross-Platform
- [ ] macOS native window controls
- [ ] Windows native title bar
- [ ] Linux desktop integration
- [ ] Consistent behavior across all platforms
- [ ] HiDPI/Retina support

## 15. Keyboard Shortcuts (Qt-native)
| Shortcut | Action |
|----------|--------|
| `Ctrl+T` | New Query Tab |
| `Ctrl+W` | Close current tab |
| `Ctrl+1..9` | Switch to nth tab |
| `Ctrl+R` / `F5` | Execute current statement |
| `Ctrl+Shift+R` | Execute all statements |
| `Ctrl+E` | Run EXPLAIN |
| `Ctrl+D` | Duplicate selected rows |
| `Delete` | Mark rows for deletion |
| `Ctrl+S` | Commit pending changes |
| `Ctrl+K` | Quick Switcher / Command Palette |
| `Ctrl+Z` | Undo |
| `Ctrl+Shift+Z` | Redo |
| `Ctrl+F` | Find in editor |
| `Ctrl+H` | Find and Replace |
