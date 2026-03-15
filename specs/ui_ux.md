# UI/UX Specification (Qt Widgets)

## Overview
The application is a native Qt Widgets desktop application styled with custom stylesheets. The design targets a modern IDE-like experience similar to TablePlus, DataGrip, or Qt Creator.

## 1. Global Layout Structure
Uses `QSplitter` for a collapsible multi-pane layout.

```
┌──────────────────────────────────────────────────┐
│ QMenuBar (File, Edit, View, Tools, Help)         │
├──────────────────────────────────────────────────┤
│ QToolBar (Connection, Run, Cancel, Toggle)       │
├─────────┬──────────────────────────┬─────────────┤
│ QDockWidget │ QTabWidget         │ QDockWidget   │
│ (Sidebar)   │ (Main Workspace)   │ (Right Panel) │
│  QTreeView  │ ┌────────────────┐ │  AI Chat      │
│  + Search   │ │ QTabBar        │ │  History      │
│             │ ├────────────────┤ │  Formatter    │
│             │ │ QScintilla     │ │               │
│             │ ├────────────────┤ │               │
│             │ │ QTableView     │ │               │
│             │ │ + Status Bar   │ │               │
│             │ └────────────────┘ │               │
└─────────┴──────────────────────────┴─────────────┘
└──────────────────────────────────────────────────┘
│              QStatusBar                          │
└──────────────────────────────────────────────────┘
```

### QMenuBar
- **File**: New Connection, Open Recent, Save, Export, Import, Exit
- **Edit**: Undo, Redo, Cut, Copy, Paste, Find, Select All
- **View**: Zoom In/Out, Toggle Sidebar, Toggle Right Panel, Full Screen
- **Query**: Execute, Execute All, Explain, Cancel, History
- **Tools**: SQL Formatter, Data Compare, Database Diff
- **Help**: Documentation, Keyboard Shortcuts, About

### QToolBar
- Connection status indicator (QLabel with colored pixmap)
- Database/Schema selector (QComboBox)
- Run Query (`Ctrl+R` / `F5`) — `QToolButton` with green icon
- Cancel Query — `QToolButton` with stop icon
- Toggle Right Panel — `QAction` with toggle slot

### Sidebar (QDockWidget)
- Schema tree with lazy-loaded nodes (`QTreeView` + custom model)
- Search bar (`QLineEdit`) for real-time table filtering
- Context menus (`QMenu`): Open, Copy Name, Drop, Truncate, Show DDL
- Collapsible folders: Tables, Views, Routines (via tree items)

### Main Workspace (QTabWidget)
- **Tab Bar**: `QTabBar` with close buttons (`QTabClosable`), movable tabs
- **Editor Area**: `QScintilla` instance for SQL / Redis commands
- **Results Area**: `QTableView` for query results / table browsing
- **Status Bar**: `QWidget` showing row count, execution time, pagination controls

### Right Panel (QDockWidget)
- AI Chat tab (streaming `QTextBrowser`)
- Query History tab (`QListView` with search)
- SQL Formatter tab (`QPlainTextEdit` + format button)

## 2. Core Widgets

### 2.1 Connection Dialog (`ConnectionDialog`)
A `QDialog` with `QTabWidget`:
- **General**: DB type `QComboBox`, host/port edits, username/password edits, database edit
- **SSH**: Enable `QCheckBox`, host/port/user edits, auth method `QComboBox` (password/key/agent), file pickers
- **SSL**: Enable `QCheckBox`, mode `QComboBox`, file pickers for certs (`.pem`, `.crt`)
- **Advanced**: Safe mode `QComboBox`, default schema edit, startup commands `QPlainTextEdit`
- **Footer**: Test Connection (`QPushButton`), Connect, Save buttons with `QDialogButtonBox`

### 2.2 SQL Editor (`SqlEditor`)
- **Engine**: `QScintilla` (`QsciScintilla`) — same as Notepad++
- **Features**: Syntax highlighting (`QsciLexerSQL`), autocomplete (`QsciAPIs`), line numbers, code folding
- **Vim Mode**: Custom `eventFilter()` with basic h/j/k/l navigation
- **Query Splitting**: `splitStatements()` with state machine, execute selected or all
- **Theme**: Custom dark/light lexer colors matching app design

### 2.3 Data Grid (`QueryResultView`)
- **Engine**: `QTableView` + `QueryResultModel` (`QAbstractTableModel` subclass)
- **Features**:
  - Virtual scrolling via model (renders only visible rows)
  - Resizable columns (`QHeaderView::ResizeToContents`)
  - Click header to sort (emits `sortRequested` → re-execute with `ORDER BY`)
  - Row numbering via `headerData()` override
  - **Inline Editing**: Double-click cell → `QItemDelegate` editor → `setData()`
  - **Visual Deltas**: Yellow=edited, Green=new row, Red strikethrough=deleted (via `data(BackgroundRole)`)
- **Pagination**: `GridStatusBar` widget with offset/limit controls

### 2.4 Structure / DDL Tab
- Sub-tabs: `QTabWidget` with Columns | Indexes | Foreign Keys | DDL
- Columns grid (`QTableView`) showing type, nullable, auto-increment, default
- DDL display (`QPlainTextEdit` with readonly flag)

### 2.5 Export / Import Dialogs
- **ExportDialog**: `QDialog` with format `QComboBox`, options `QCheckBox`es, file picker via `QFileDialog::getSaveFileName()`
- **ImportDialog**: `QDialog` with target table selector, column mapping `QTableWidget`, progress `QProgressBar`

## 3. Keyboard Shortcuts
| Shortcut | Action | Qt Equivalent |
|---|---|---|
| `Ctrl+N` | New Query Tab | `QShortcut(QKeySequence::New)` |
| `Ctrl+W` | Close current tab | `QShortcut(QKeySequence::Close)` |
| `Ctrl+1..9` | Switch to nth tab | Custom `QShortcut` |
| `Ctrl+R` / `F5` | Execute current statement | `QShortcut("Ctrl+R")` |
| `Ctrl+Shift+R` | Execute all statements | `QShortcut("Ctrl+Shift+R")` |
| `Ctrl+E` | Run EXPLAIN | Custom `QShortcut` |
| `Ctrl+D` | Duplicate selected rows | Custom `QShortcut` |
| `Delete` | Mark rows for deletion | `QAction(Delete)` |
| `Ctrl+S` | Commit pending changes | `QShortcut(QKeySequence::Save)` |
| `Ctrl+K` | Quick Switcher / Command Palette | Custom `QShortcut` |
| `Ctrl+F` | Find | `QShortcut(QKeySequence::Find)` |
| `Ctrl+H` | Replace | `QShortcut(QKeySequence::Replace)` |
| `Ctrl++` | Zoom in | `QShortcut(QKeySequence::ZoomIn)` |
| `Ctrl+-` | Zoom out | `QShortcut(QKeySequence::ZoomOut)` |

## 4. Design System

### Color Palette (Dark Theme)
```css
/* Base colors — Catppuccin Mocha inspired */
--bg-primary: #1E1E2E;      /* Main background */
--bg-secondary: #181825;    /* Sidebar, tabs */
--bg-tertiary: #313244;     /* Borders, dividers */
--fg-primary: #CDD6F4;      /* Primary text */
--fg-secondary: #A6ADC8;    /* Secondary text */
--accent: #89B4FA;          /* Blue accent */
--success: #A6E3A1;         /* Green */
--warning: #F9E2AF;         /* Yellow */
--danger: #F38BA8;          /* Red */
```

### Typography
- **UI Font**: System default (San Francisco on macOS, Segoe UI on Windows)
- **Code Font**: JetBrains Mono 14pt (or Consolas fallback)
- Applied via stylesheets:
```css
QsciScintilla, QPlainTextEdit, QLineEdit {
    font-family: "JetBrains Mono", "Consolas", monospace;
    font-size: 14pt;
}

QTreeView, QTableView, QListView {
    font-size: 13pt;
}
```

### Icons
- Built-in Qt icons where available: `QStyle::SP_*`
- Custom icons: SVG files loaded via `QIcon::fromTheme()` or direct `QIcon("path/to/icon.svg")`
- Icon size: 16x16 for toolbar, 24x24 for dialogs

### Animations
- Panel transitions: `QPropertyAnimation` on splitter sizes
- Toast notifications: Custom `QWidget` with `QPropertyAnimation` for fade in/out
- Loading spinners: `QProgressBar` with `QProgressBar::chunk` styling

### OS Integration
- Native menu bar on macOS: `QMenuBar` automatically integrates
- Native file dialogs: `QFileDialog` uses system dialogs
- System tray: `QSystemTrayIcon` for background operation
- Native notifications: `QSystemTrayIcon::showMessage()`

## 5. Responsive Behavior

### Window States
- **Minimum window size**: 1024x768 (enforced in `MainWindow::resizeEvent()`)
- **Sidebar**: Collapsible via `QDockWidget::toggleViewAction()`
- **Right panel**: Collapsible via `QDockWidget::toggleViewAction()`
- **Full screen**: `showFullScreen()` toggled via `F11`

### High DPI Support
- Qt automatically scales on HiDPI displays (`Qt::AA_EnableHighDpiScaling`)
- Icon sizes scale with `devicePixelRatio()`
- Font sizes use point system (automatic scaling)

### Layout Behavior
- `QSplitter` maintains proportions on resize
- `QScrollArea` for overflow content
- `sizePolicy` set appropriately:
  - Editors: `QSizePolicy::Expanding`
  - Buttons: `QSizePolicy::Fixed` or `QSizePolicy::Preferred`
  - Stretch factors in layouts for proportional sizing

## 6. Accessibility
- Tab order: `setTabOrder()` for logical navigation
- Keyboard navigation: All actions accessible via shortcuts
- Screen reader: Qt's accessibility integration
- High contrast: Custom high-contrast stylesheet option
- Font scaling: `Ctrl++`/`Ctrl+-` for zoom
