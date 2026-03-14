# UI/UX Specification (React + Wails)

The frontend is built with **React + TypeScript**, styled with **Tailwind CSS**, and rendered inside Wails' native WebView. The design targets a modern IDE-like experience similar to TablePlus, DataGrip, or VS Code.

## 1. Global Layout Structure
Uses `react-resizable-panels` for a collapsible multi-pane layout.

```
┌──────────────────────────────────────────────────┐
│ Toolbar (Connection status, Run, Cancel, Toggle) │
├─────────┬──────────────────────────┬─────────────┤
│ Sidebar │ Main Workspace           │ Right Panel │
│ (Schema │ ┌──────────────────────┐ │ (AI Chat /  │
│  Tree)  │ │ Tab Bar              │ │  History /  │
│         │ ├──────────────────────┤ │  Format)    │
│         │ │ SQL Editor (Monaco)  │ │             │
│         │ ├──────────────────────┤ │             │
│         │ │ Data Grid (AG Grid)  │ │             │
│         │ │ + Status Bar         │ │             │
│         │ └──────────────────────┘ │             │
└─────────┴──────────────────────────┴─────────────┘
```

### Toolbar
- Connection status indicator (colored dot)
- Database/Schema dropdown selector
- Run Query (`Cmd+R` / `F5`)
- Cancel Query (stop icon)
- Toggle Right Panel

### Sidebar (Left)
- Schema tree with lazy-loaded nodes
- Search bar for real-time table filtering
- Context menus: Open, Copy Name, Drop, Truncate, Show DDL
- Collapsible folders: Tables, Views, Routines

### Main Workspace
- **Tab Bar**: Horizontal tabs with close buttons, drag-to-reorder
- **Editor Area**: Monaco Editor instance for SQL / Redis commands
- **Results Area**: AG Grid for query results / table browsing
- **Status Bar**: Row count, execution time, pagination, Save/Discard buttons

### Right Panel (Collapsible)
- AI Chat tab (streaming markdown)
- Query History tab (searchable)
- SQL Formatter tab

## 2. Core Components

### 2.1 Connection Form (`ConnectionForm.tsx`)
A modal dialog with tabs:
- **General**: DB type dropdown, host, port, username, password, database
- **SSH**: Enable toggle, host, port, user, auth method (password/key/agent)
- **SSL**: Enable toggle, mode picker, file pickers for certs
- **Advanced**: Safe mode, default schema, startup commands
- **Footer**: Test Connection, Connect, Save buttons

### 2.2 SQL Editor
- **Engine**: `@monaco-editor/react` (VS Code engine)
- **Features**: Syntax highlighting, autocomplete, multi-cursor, minimap
- **Vim Mode**: Monaco Vim extension
- **Query Splitting**: Parse multiple statements, execute selected or all
- **Theme**: Custom dark/light theme matching app design

### 2.3 Data Grid
- **Engine**: `AG Grid Community` with React wrapper
- **Features**:
  - Virtual scrolling for millions of rows
  - Resizable, reorderable columns
  - Click header to sort (→ `ORDER BY` re-query)
  - Row numbering gutter
  - **Inline Editing**: Double-click cell → editable
  - **Visual Deltas**: Yellow=edited, Green=new row, Red strikethrough=deleted
- **Pagination**: Offset/limit controls in status bar

### 2.4 Structure / DDL Tab
- Sub-tabs: Columns | Indexes | Foreign Keys | DDL
- Columns grid showing type, nullable, auto-increment, default

### 2.5 Export / Import Dialogs
- Format picker, options toggles, file picker via Wails native dialog
- Import: target table selector, column mapping UI, progress bar

## 3. Keyboard Shortcuts
| Shortcut | Action |
|---|---|
| `Cmd+T` | New Query Tab |
| `Cmd+W` | Close current tab |
| `Cmd+1..9` | Switch to nth tab |
| `Cmd+R` / `F5` | Execute current statement |
| `Cmd+Shift+R` | Execute all statements |
| `Cmd+E` | Run EXPLAIN |
| `Cmd+D` | Duplicate selected rows |
| `Delete` | Mark rows for deletion |
| `Cmd+S` | Commit pending changes |
| `Cmd+K` | Quick Switcher / Command Palette |

## 4. Design System
- **Colors**: Tailwind CSS with custom palette (dark mode first)
- **Typography**: Inter / JetBrains Mono for code
- **Animations**: Framer Motion for panel transitions, toasts
- **Icons**: Lucide React
- **OS Integration**: Wails' native title bar, file dialogs, system tray
