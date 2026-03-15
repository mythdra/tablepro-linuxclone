# UI Components & Navigation (Qt Widgets)

## Overview
The UI is a native Qt Widgets application with a QMainWindow-based architecture. Components use standard Qt widgets with custom styling via stylesheets. State is managed by C++ QObject stores that emit signals for UI updates.

## Component Architecture
```
MainWindow
├── QToolBar (MainToolbar)
├── QSplitter (MainSplitter) — Horizontal, 3-pane
│   ├── QDockWidget (Sidebar)
│   │   ├── SchemaTreeView (QTreeView)
│   │   ├── QLineEdit (SearchBar)
│   │   └── QMenu (ContextMenu)
│   ├── QTabWidget (MainWorkspace)
│   │   ├── TabBar (QTabBar)
│   │   ├── SqlEditor (QScintilla)
│   │   ├── QueryResultView (QTableView)
│   │   └── StatusBarWidget
│   └── QDockWidget (RightPanel)
│       ├── AIChatWidget
│       ├── QueryHistoryWidget
│       └── FormatterWidget
├── QStatusBar (MainWindow status bar)
└── QDialogs (Modals)
    ├── ConnectionDialog
    ├── ExportDialog
    ├── ImportDialog
    └── SettingsDialog
```

## State Management (Qt Signals/Slots)
Each major feature has its own QObject-based manager:
```cpp
// core/ConnectionManager.hpp
class ConnectionManager : public QObject {
    Q_OBJECT
    Q_PROPERTY(QList<ConnectionConfig> connections READ connections NOTIFY connectionsChanged)

public:
    explicit ConnectionManager(QObject* parent = nullptr);

    Q_INVOKABLE void loadConnections();
    Q_INVOKABLE bool testConnection(const ConnectionConfig& config);
    Q_INVOKABLE void saveConnection(const ConnectionConfig& config);

signals:
    void connectionsChanged();
    void connectionTested(bool success, const QString& message);
    void connectionStatusChanged(const QUuid& connectionId, ConnectionStatus status);

private:
    QList<ConnectionConfig> m_connections;
};
```

## Event-Driven Updates
Qt's signal/slot system replaces Wails events:
```cpp
// C++ backend emits
emit queryProgress(progressData);
emit connectionStatusChanged(connectionId, status);

// Frontend connects
connect(m_queryManager, &QueryManager::queryProgress,
        this, [this](const QueryProgress& progress) {
    m_progressBar->setValue(progress.percent());
});
```

## Main Content Coordinator
The Swift `MainContentCoordinator` monolith maps to focused Qt managers:
- `TabManager` — tab CRUD, selection, persistence
- `QueryManager` — query execution, results, pagination
- `DataChangeManager` — cell edits, pending changes, undo/redo
- `FilterManager` — column filters, sort state
- `SchemaManager` — schema tree, search, selection

## Widget Mapping

| React/Wails | Qt Equivalent |
|-------------|---------------|
| `div` with flexbox | `QWidget` with `QHBoxLayout`/`QVBoxLayout` |
| `react-resizable-panels` | `QSplitter` |
| AG Grid | `QTableView` + `QAbstractTableModel` |
| Monaco Editor | `QScintilla` (`QsciScintilla`) |
| Radix UI primitives | Native Qt widgets |
| Tailwind CSS | Qt Stylesheets (`.qss`) |
| Framer Motion | `QPropertyAnimation` |
| Lucide React | Custom SVG icons or FontAwesome |

## MainWindow Structure
```cpp
// src/ui/MainWindow/MainWindow.hpp
#pragma once

#include <QMainWindow>
#include <QUuid>
#include "core/ConnectionManager.hpp"
#include "core/TabManager.hpp"

namespace tablepro {

class MainWindow : public QMainWindow {
    Q_OBJECT

public:
    explicit MainWindow(QWidget* parent = nullptr);
    ~MainWindow();

    // Public API for menu actions
    void newQueryTab();
    void closeCurrentTab();
    void executeCurrentQuery();
    void executeAllQueries();

protected:
    void closeEvent(QCloseEvent* event) override;

private slots:
    void onConnectionSelected(const QUuid& connectionId);
    void onTableDoubleClicked(const TableInfo& table);
    void onViewDoubleClicked(const TableInfo& view);

private:
    void setupUi();
    void setupMenuBar();
    void setupToolBar();
    void setupStatusBar();
    void setupCentralWidget();
    void connectSignals();

    // Core services
    ConnectionManager* m_connectionManager;
    TabManager* m_tabManager;
    QueryManager* m_queryManager;
    SchemaManager* m_schemaManager;

    // UI widgets
    QSplitter* m_mainSplitter;
    SchemaTreeView* m_schemaTreeView;
    QTabWidget* m_mainWorkspace;
    SqlEditor* m_currentEditor;
    QueryResultView* m_currentGridView;
    QToolBar* m_mainToolBar;
    QStatusBar* m_statusBar;
};

} // namespace tablepro
```

## Stylesheet System
```cpp
// resources/styles/dark.qss
QMainWindow {
    background-color: #1E1E2E;
    color: #CDD6F4;
}

QTreeView {
    background-color: #181825;
    border: none;
    color: #CDD6F4;
}

QTreeView::item:selected {
    background-color: #45475A;
}

QTableView {
    background-color: #1E1E2E;
    gridline-color: #313244;
    selection-background-color: #45475A;
}

QTabBar::tab {
    background-color: #181825;
    padding: 8px 16px;
    border-top-left-radius: 4px;
    border-top-right-radius: 4px;
}

QTabBar::tab:selected {
    background-color: #1E1E2E;
}
```

## Keyboard Shortcuts (Qt)
```cpp
// In MainWindow::setupMenuBar()
auto* newTabAction = new QShortcut(QKeySequence::New, this);
connect(newTabAction, &QShortcut::activated,
        this, &MainWindow::newQueryTab);

auto* executeAction = new QShortcut(QKeySequence("Ctrl+R"), this);
connect(executeAction, &QShortcut::activated,
        this, [this]() { m_currentEditor->executeCurrentStatement(); });

auto* executeAllAction = new QShortcut(QKeySequence("Ctrl+Shift+R"), this);
connect(executeAllAction, &QShortcut::activated,
        this, [this]() { m_currentEditor->executeAllStatements(); });
```

## Modal Dialogs
```cpp
// src/ui/Dialogs/ConnectionDialog.hpp
class ConnectionDialog : public QDialog {
    Q_OBJECT

public:
    explicit ConnectionDialog(QWidget* parent = nullptr);
    ConnectionConfig config() const;
    void setConfig(const ConnectionConfig& config);

signals:
    void testConnectionRequested(const ConnectionConfig& config);

private slots:
    void onTestClicked();
    void onOkClicked();

private:
    void setupUi();
    void setupGeneralTab();
    void setupSSHTab();
    void setupSSLTab();
    void setupAdvancedTab();

    QTabWidget* m_tabWidget;
    QLineEdit* m_hostEdit;
    QLineEdit* m_portEdit;
    QLineEdit* m_usernameEdit;
    QLineEdit* m_passwordEdit;
    // ... more widgets
};
```
