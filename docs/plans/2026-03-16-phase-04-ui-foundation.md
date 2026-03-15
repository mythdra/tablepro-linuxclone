# Phase 4: UI Foundation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build main UI layout with sidebar (schema tree), tab management, toolbar, and status bar.

**Architecture:** QMainWindow with QSplitter dividing sidebar and main content. QTreeView for schema browser. QTabBar for query/table tabs. Qt signals/slots for UI-state communication.

**Tech Stack:** C++20, Qt 6.6 Widgets (QMainWindow, QSplitter, QTreeView, QTabBar)

---

## Task 1: Schema Tree Widget

**Files:**
- Create: `src/ui/widgets/schema_tree.hpp`
- Create: `src/ui/widgets/schema_tree.cpp`

**Step 1: Create schema_tree.hpp**

```cpp
#pragma once

#include <QTreeView>
#include <QStandardItemModel>
#include <QMenu>
#include "core/types.hpp"

namespace tablepro {

class SchemaTree : public QTreeView {
    Q_OBJECT

public:
    explicit SchemaTree(QWidget* parent = nullptr);

    void setSchema(const SchemaInfo& schema);
    void setTables(const QList<TableInfo>& tables);
    void setViews(const QList<TableInfo>& views);
    void clear();

signals:
    void tableSelected(const QString& schema, const QString& table);
    void tableDoubleClicked(const QString& schema, const QString& table);
    void viewSelected(const QString& schema, const QString& view);
    void refreshRequested();

private slots:
    void onItemDoubleClicked(const QModelIndex& index);
    void onCustomContextMenu(const QPoint& point);

private:
    void setupModel();
    void setupContextMenu();
    QStandardItem* findOrCreateCategory(const QString& name);

    QStandardItemModel* m_model;
    QMenu* m_contextMenu;
    QString m_currentSchema;
};

} // namespace tablepro
```

**Step 2: Create schema_tree.cpp**

```cpp
#include "schema_tree.hpp"
#include <QHeaderView>
#include <QApplication>

namespace tablepro {

SchemaTree::SchemaTree(QWidget* parent)
    : QTreeView(parent)
    , m_model(new QStandardItemModel(this))
    , m_contextMenu(new QMenu(this))
{
    setupModel();
    setupContextMenu();
}

void SchemaTree::setupModel() {
    m_model->setHorizontalHeaderLabels({tr("Schema")});
    setModel(m_model);

    // Appearance
    setAnimated(true);
    setExpandsOnDoubleClick(true);
    setSelectionMode(QAbstractItemView::SingleSelection);
    setSelectionBehavior(QAbstractItemView::SelectItems);
    setContextMenuPolicy(Qt::CustomContextMenu);

    // Hide header
    header()->hide();

    // Connect signals
    connect(this, &QTreeView::doubleClicked, this, &SchemaTree::onItemDoubleClicked);
    connect(this, &QTreeView::customContextMenuRequested, this, &SchemaTree::onCustomContextMenu);
}

void SchemaTree::setupContextMenu() {
    m_contextMenu->addAction(tr("Refresh"), this, &SchemaTree::refreshRequested);
    m_contextMenu->addSeparator();
    m_contextMenu->addAction(tr("Copy Name"), [this]() {
        auto idx = currentIndex();
        if (idx.isValid()) {
            QApplication::clipboard()->setText(m_model->itemFromIndex(idx)->text());
        }
    });
}

void SchemaTree::setSchema(const SchemaInfo& schema) {
    clear();
    m_currentSchema = schema.currentSchema;

    // Create root item for database
    auto* dbItem = new QStandardItem(schema.databaseName);
    dbItem->setIcon(QIcon::fromTheme("database"));
    m_model->appendRow(dbItem);

    // Add schemas as children
    for (const auto& schemaName : schema.schemas) {
        auto* schemaItem = new QStandardItem(schemaName);
        schemaItem->setData(schemaName, Qt::UserRole + 1); // Store schema name
        schemaItem->setData("schema", Qt::UserRole + 2);   // Store type
        dbItem->appendRow(schemaItem);
    }

    expand(dbItem->index());
}

void SchemaTree::setTables(const QList<TableInfo>& tables) {
    auto* tablesCategory = findOrCreateCategory(tr("Tables"));
    tablesCategory->removeRows(0, tablesCategory->rowCount());

    for (const auto& table : tables) {
        if (table.type != "table") continue;

        auto* tableItem = new QStandardItem(table.name);
        tableItem->setIcon(QIcon::fromTheme("table"));
        tableItem->setData(table.schema, Qt::UserRole + 1);
        tableItem->setData(table.name, Qt::UserRole + 2);
        tableItem->setData("table", Qt::UserRole + 3);
        tablesCategory->appendRow(tableItem);
    }
}

void SchemaTree::setViews(const QList<TableInfo>& views) {
    auto* viewsCategory = findOrCreateCategory(tr("Views"));
    viewsCategory->removeRows(0, viewsCategory->rowCount());

    for (const auto& view : views) {
        if (view.type != "view") continue;

        auto* viewItem = new QStandardItem(view.name);
        viewItem->setIcon(QIcon::fromTheme("view"));
        viewItem->setData(view.schema, Qt::UserRole + 1);
        viewItem->setData(view.name, Qt::UserRole + 2);
        viewItem->setData("view", Qt::UserRole + 3);
        viewsCategory->appendRow(viewItem);
    }
}

QStandardItem* SchemaTree::findOrCreateCategory(const QString& name) {
    auto root = m_model->invisibleRootItem();

    for (int i = 0; i < root->rowCount(); ++i) {
        auto* item = root->child(i);
        if (item && item->text() == name) {
            return item;
        }
    }

    auto* category = new QStandardItem(name);
    category->setIcon(QIcon::fromTheme("folder"));
    m_model->appendRow(category);
    return category;
}

void SchemaTree::clear() {
    m_model->clear();
    m_model->setHorizontalHeaderLabels({tr("Schema")});
}

void SchemaTree::onItemDoubleClicked(const QModelIndex& index) {
    auto* item = m_model->itemFromIndex(index);
    if (!item) return;

    QString type = item->data(Qt::UserRole + 3).toString();
    QString schema = item->data(Qt::UserRole + 1).toString();
    QString name = item->data(Qt::UserRole + 2).toString();

    if (type == "table") {
        emit tableDoubleClicked(schema, name);
    } else if (type == "view") {
        emit viewSelected(schema, name);
    }
}

void SchemaTree::onCustomContextMenu(const QPoint& point) {
    QModelIndex index = indexAt(point);
    if (index.isValid()) {
        m_contextMenu->exec(viewport()->mapToGlobal(point));
    }
}

} // namespace tablepro
```

**Step 3: Commit schema tree**

```bash
git add src/ui/widgets/schema_tree.hpp src/ui/widgets/schema_tree.cpp
git commit -m "feat: Add SchemaTree widget for database browser"
```

---

## Task 2: Tab Manager

**Files:**
- Create: `src/ui/widgets/tab_manager.hpp`
- Create: `src/ui/widgets/tab_manager.cpp`

**Step 1: Create tab_manager.hpp**

```cpp
#pragma once

#include <QObject>
#include <QTabWidget>
#include <QTabBar>
#include <QHash>
#include <QUuid>
#include <memory>

namespace tablepro {

class TabContent;

enum class TabType {
    Query,
    TableData,
    TableStructure
};

struct TabInfo {
    QString id;
    QString title;
    TabType type;
    QString connectionId;
    QString schema;
    QString tableName;
    bool isModified = false;
};

class TabManager : public QTabWidget {
    Q_OBJECT

public:
    explicit TabManager(QWidget* parent = nullptr);

    // Tab creation
    QString createQueryTab(const QString& connectionId, const QString& title = QString());
    QString createTableDataTab(const QString& connectionId, const QString& schema, const QString& table);
    QString createTableStructureTab(const QString& connectionId, const QString& schema, const QString& table);

    // Tab management
    void closeTab(const QString& tabId);
    void closeCurrentTab();
    void closeAllTabs();
    void closeOtherTabs(const QString& tabId);

    // Tab queries
    TabInfo tabInfo(const QString& tabId) const;
    QString currentTabId() const;
    QList<TabInfo> allTabs() const;
    int tabCount() const;
    TabContent* tabContent(const QString& tabId) const;

    // Tab state
    void setTabModified(const QString& tabId, bool modified);
    void setTabTitle(const QString& tabId, const QString& title);

    // Persistence
    void saveTabState();
    void restoreTabState();

signals:
    void tabCreated(const QString& tabId);
    void tabClosed(const QString& tabId);
    void tabActivated(const QString& tabId);
    void tabModifiedChanged(const QString& tabId, bool modified);
    void allTabsClosed();

private slots:
    void onTabCloseRequested(int index);
    void onCurrentChanged(int index);

private:
    QString generateTabId() const;
    void updateTabTitle(int index, const QString& title, bool modified);
    QString tabTitleWithModifier(const QString& title, bool modified) const;

    QHash<QString, TabInfo> m_tabs;
    QHash<QString, TabContent*> m_contents;
};

} // namespace tablepro
```

**Step 2: Create tab_manager.cpp**

```cpp
#include "tab_manager.hpp"
#include <QTabBar>
#include <QMessageBox>
#include <QJsonDocument>
#include <QJsonArray>
#include <QFile>

namespace tablepro {

TabManager::TabManager(QWidget* parent)
    : QTabWidget(parent)
{
    setTabsClosable(true);
    setMovable(true);
    setDocumentMode(true);

    // Connect signals
    connect(tabBar(), &QTabBar::tabCloseRequested, this, &TabManager::onTabCloseRequested);
    connect(this, &QTabWidget::currentChanged, this, &TabManager::onCurrentChanged);

    // Style
    setStyleSheet(R"(
        QTabWidget::pane {
            border: none;
            background-color: #1E1E2E;
        }
        QTabBar::tab {
            background-color: #181825;
            color: #CDD6F4;
            padding: 8px 16px;
            margin-right: 2px;
            border-top-left-radius: 4px;
            border-top-right-radius: 4px;
        }
        QTabBar::tab:selected {
            background-color: #1E1E2E;
            border-bottom: 2px solid #89B4FA;
        }
        QTabBar::tab:hover:!selected {
            background-color: #313244;
        }
        QTabBar::close-button {
            image: none;
            subcontrol-position: right;
            margin-right: 4px;
        }
    )");
}

QString TabManager::generateTabId() const {
    return QUuid::createUuid().toString(QUuid::WithoutBraces);
}

QString TabManager::createQueryTab(const QString& connectionId, const QString& title) {
    QString tabId = generateTabId();

    TabInfo info;
    info.id = tabId;
    info.title = title.isEmpty() ? tr("Query %1").arg(m_tabs.size() + 1) : title;
    info.type = TabType::Query;
    info.connectionId = connectionId;

    m_tabs.insert(tabId, info);

    // Create placeholder widget (will be replaced with QueryEditor)
    auto* placeholder = new QWidget(this);
    int index = addTab(placeholder, info.title);
    setTabToolTip(index, info.title);

    emit tabCreated(tabId);

    return tabId;
}

QString TabManager::createTableDataTab(const QString& connectionId, const QString& schema, const QString& table) {
    QString tabId = generateTabId();

    TabInfo info;
    info.id = tabId;
    info.title = table;
    info.type = TabType::TableData;
    info.connectionId = connectionId;
    info.schema = schema;
    info.tableName = table;

    m_tabs.insert(tabId, info);

    // Create placeholder widget (will be replaced with DataGrid)
    auto* placeholder = new QWidget(this);
    int index = addTab(placeholder, info.title);
    setTabToolTip(index, QString("%1.%2").arg(schema, table));

    emit tabCreated(tabId);

    return tabId;
}

QString TabManager::createTableStructureTab(const QString& connectionId, const QString& schema, const QString& table) {
    QString tabId = generateTabId();

    TabInfo info;
    info.id = tabId;
    info.title = QString("%1 (Structure)").arg(table);
    info.type = TabType::TableStructure;
    info.connectionId = connectionId;
    info.schema = schema;
    info.tableName = table;

    m_tabs.insert(tabId, info);

    auto* placeholder = new QWidget(this);
    int index = addTab(placeholder, info.title);

    emit tabCreated(tabId);

    return tabId;
}

void TabManager::closeTab(const QString& tabId) {
    if (!m_tabs.contains(tabId)) return;

    int index = -1;
    for (int i = 0; i < count(); ++i) {
        auto* widget = this->widget(i);
        if (m_tabs.value(tabId).connectionId == "placeholder") {
            // Find by tabId stored in widget property
        }
    }

    // Find tab index by iterating
    auto tabIds = m_tabs.keys();
    for (int i = 0; i < qMin(count(), tabIds.size()); ++i) {
        if (i < tabIds.size() && m_tabs.contains(tabIds[i])) {
            removeTab(i);
            m_tabs.remove(tabIds[i]);
            emit tabClosed(tabIds[i]);
            break;
        }
    }
}

void TabManager::closeCurrentTab() {
    int index = currentIndex();
    if (index >= 0) {
        onTabCloseRequested(index);
    }
}

void TabManager::closeAllTabs() {
    while (count() > 0) {
        removeTab(0);
    }
    m_tabs.clear();
    emit allTabsClosed();
}

void TabManager::closeOtherTabs(const QString& tabId) {
    // Close all tabs except the specified one
    QList<QString> toRemove;
    for (auto it = m_tabs.begin(); it != m_tabs.end(); ++it) {
        if (it.key() != tabId) {
            toRemove.append(it.key());
        }
    }

    for (const auto& id : toRemove) {
        closeTab(id);
    }
}

TabInfo TabManager::tabInfo(const QString& tabId) const {
    return m_tabs.value(tabId);
}

QString TabManager::currentTabId() const {
    int index = currentIndex();
    if (index < 0 || index >= m_tabs.size()) return QString();

    auto keys = m_tabs.keys();
    if (index < keys.size()) {
        return keys[index];
    }
    return QString();
}

QList<TabInfo> TabManager::allTabs() const {
    return m_tabs.values();
}

int TabManager::tabCount() const {
    return count();
}

TabContent* TabManager::tabContent(const QString& tabId) const {
    return m_contents.value(tabId);
}

void TabManager::setTabModified(const QString& tabId, bool modified) {
    if (!m_tabs.contains(tabId)) return;

    m_tabs[tabId].isModified = modified;

    int index = -1;
    auto keys = m_tabs.keys();
    for (int i = 0; i < keys.size(); ++i) {
        if (keys[i] == tabId) {
            index = i;
            break;
        }
    }

    if (index >= 0) {
        updateTabTitle(index, m_tabs[tabId].title, modified);
        emit tabModifiedChanged(tabId, modified);
    }
}

void TabManager::setTabTitle(const QString& tabId, const QString& title) {
    if (!m_tabs.contains(tabId)) return;

    m_tabs[tabId].title = title;

    int index = -1;
    auto keys = m_tabs.keys();
    for (int i = 0; i < keys.size(); ++i) {
        if (keys[i] == tabId) {
            index = i;
            break;
        }
    }

    if (index >= 0) {
        updateTabTitle(index, title, m_tabs[tabId].isModified);
    }
}

void TabManager::updateTabTitle(int index, const QString& title, bool modified) {
    QString displayTitle = tabTitleWithModifier(title, modified);
    setTabText(index, displayTitle);
}

QString TabManager::tabTitleWithModifier(const QString& title, bool modified) const {
    return modified ? QString("* %1").arg(title) : title;
}

void TabManager::onTabCloseRequested(int index) {
    // TODO: Check for unsaved changes

    QString tabId;
    auto keys = m_tabs.keys();
    if (index < keys.size()) {
        tabId = keys[index];
    }

    removeTab(index);

    if (!tabId.isEmpty()) {
        m_tabs.remove(tabId);
        emit tabClosed(tabId);
    }
}

void TabManager::onCurrentChanged(int index) {
    QString tabId;
    auto keys = m_tabs.keys();
    if (index >= 0 && index < keys.size()) {
        tabId = keys[index];
        emit tabActivated(tabId);
    }
}

void TabManager::saveTabState() {
    // TODO: Implement persistence
}

void TabManager::restoreTabState() {
    // TODO: Implement persistence
}

} // namespace tablepro
```

**Step 3: Create TabContent placeholder**

```cpp
// src/ui/widgets/tab_content.hpp
#pragma once

#include <QWidget>

namespace tablepro {

class TabContent : public QWidget {
    Q_OBJECT

public:
    explicit TabContent(QWidget* parent = nullptr) : QWidget(parent) {}
};

} // namespace tablepro
```

**Step 4: Commit tab manager**

```bash
git add src/ui/widgets/tab_manager.hpp src/ui/widgets/tab_manager.cpp src/ui/widgets/tab_content.hpp
git commit -m "feat: Add TabManager for query/table tabs"
```

---

## Task 3: Connection Selector

**Files:**
- Create: `src/ui/widgets/connection_selector.hpp`
- Create: `src/ui/widgets/connection_selector.cpp`

**Step 1: Create connection_selector.hpp**

```cpp
#pragma once

#include <QWidget>
#include <QComboBox>
#include <QPushButton>
#include <QHBoxLayout>
#include "core/types.hpp"

namespace tablepro {

class ConnectionSelector : public QWidget {
    Q_OBJECT

public:
    explicit ConnectionSelector(QWidget* parent = nullptr);

    void setConnections(const QList<ConnectionInfo>& connections);
    void setCurrentConnection(const QString& connectionId);
    QString currentConnectionId() const;

signals:
    void connectionSelected(const QString& connectionId);
    void newConnectionRequested();
    void manageConnectionsRequested();

private slots:
    void onSelectionChanged(int index);

private:
    void setupUI();
    void refreshComboBox();

    QComboBox* m_combo;
    QPushButton* m_newButton;
    QPushButton* m_manageButton;
    QList<ConnectionInfo> m_connections;
};

} // namespace tablepro
```

**Step 2: Create connection_selector.cpp**

```cpp
#include "connection_selector.hpp"

namespace tablepro {

ConnectionSelector::ConnectionSelector(QWidget* parent)
    : QWidget(parent)
{
    setupUI();
}

void ConnectionSelector::setupUI() {
    auto* layout = new QHBoxLayout(this);
    layout->setContentsMargins(0, 0, 0, 0);
    layout->setSpacing(8);

    m_combo = new QComboBox(this);
    m_combo->setMinimumWidth(200);
    m_combo->setPlaceholderText(tr("Select connection..."));

    m_newButton = new QPushButton(tr("New"), this);
    m_newButton->setFixedWidth(60);

    m_manageButton = new QPushButton(tr("Manage"), this);
    m_manageButton->setFixedWidth(70);

    layout->addWidget(m_combo);
    layout->addWidget(m_newButton);
    layout->addWidget(m_manageButton);
    layout->addStretch();

    connect(m_combo, QOverload<int>::of(&QComboBox::currentIndexChanged),
            this, &ConnectionSelector::onSelectionChanged);
    connect(m_newButton, &QPushButton::clicked, this, &ConnectionSelector::newConnectionRequested);
    connect(m_manageButton, &QPushButton::clicked, this, &ConnectionSelector::manageConnectionsRequested);
}

void ConnectionSelector::setConnections(const QList<ConnectionInfo>& connections) {
    m_connections = connections;
    refreshComboBox();
}

void ConnectionSelector::refreshComboBox() {
    m_combo->clear();

    for (const auto& conn : m_connections) {
        QString display = QString("%1 (%2)").arg(conn.name, conn.database);
        m_combo->addItem(display, conn.id);
    }
}

void ConnectionSelector::setCurrentConnection(const QString& connectionId) {
    int index = m_combo->findData(connectionId);
    if (index >= 0) {
        m_combo->setCurrentIndex(index);
    }
}

QString ConnectionSelector::currentConnectionId() const {
    return m_combo->currentData().toString();
}

void ConnectionSelector::onSelectionChanged(int index) {
    if (index >= 0) {
        QString connectionId = m_combo->itemData(index).toString();
        emit connectionSelected(connectionId);
    }
}

} // namespace tablepro
```

**Step 3: Commit connection selector**

```bash
git add src/ui/widgets/connection_selector.hpp src/ui/widgets/connection_selector.cpp
git commit -m "feat: Add ConnectionSelector widget"
```

---

## Task 4: Update MainWindow

**Files:**
- Modify: `src/ui/MainWindow.hpp`
- Modify: `src/ui/MainWindow.cpp`

**Step 1: Update MainWindow.hpp**

```cpp
#pragma once

#include <QMainWindow>
#include <QMenuBar>
#include <QToolBar>
#include <QStatusBar>
#include <QSplitter>

namespace tablepro {

class SchemaTree;
class TabManager;
class ConnectionSelector;

class MainWindow : public QMainWindow {
    Q_OBJECT

public:
    explicit MainWindow(QWidget* parent = nullptr);
    ~MainWindow() override;

private slots:
    void onNewConnection();
    void onConnectionSelected(const QString& connectionId);
    void onTableDoubleClicked(const QString& schema, const QString& table);
    void onNewQueryTab();
    void onExecuteQuery();

private:
    void setupMenuBar();
    void setupToolBar();
    void setupStatusBar();
    void setupCentralWidget();
    void applyStyleSheet();
    void setupConnections();

    QSplitter* m_mainSplitter;
    QWidget* m_sidebarWidget;
    QWidget* m_mainContent;

    SchemaTree* m_schemaTree;
    TabManager* m_tabManager;
    ConnectionSelector* m_connectionSelector;

    QString m_currentConnectionId;
};

} // namespace tablepro
```

**Step 2: Update MainWindow.cpp**

```cpp
#include "MainWindow.hpp"
#include "widgets/schema_tree.hpp"
#include "widgets/tab_manager.hpp"
#include "widgets/connection_selector.hpp"
#include "core/connection_manager.hpp"
#include <QFile>
#include <QMessageBox>

namespace tablepro {

MainWindow::MainWindow(QWidget* parent)
    : QMainWindow(parent)
    , m_mainSplitter(new QSplitter(Qt::Horizontal, this))
    , m_sidebarWidget(new QWidget(this))
    , m_mainContent(new QWidget(this))
    , m_schemaTree(new SchemaTree(this))
    , m_tabManager(new TabManager(this))
    , m_connectionSelector(new ConnectionSelector(this))
{
    setWindowTitle(tr("TablePro"));
    setMinimumSize(1280, 720);
    resize(1400, 900);

    applyStyleSheet();
    setupMenuBar();
    setupToolBar();
    setupCentralWidget();
    setupStatusBar();
    setupConnections();
}

MainWindow::~MainWindow() = default;

void MainWindow::setupMenuBar() {
    // File menu
    auto* fileMenu = menuBar()->addMenu(tr("&File"));

    auto* newConnAction = fileMenu->addAction(tr("&New Connection..."));
    newConnAction->setShortcut(QKeySequence::New);
    connect(newConnAction, &QAction::triggered, this, &MainWindow::onNewConnection);

    fileMenu->addSeparator();

    auto* quitAction = fileMenu->addAction(tr("&Quit"));
    quitAction->setShortcut(QKeySequence::Quit);
    connect(quitAction, &QAction::triggered, qApp, &QApplication::quit);

    // Edit menu
    auto* editMenu = menuBar()->addMenu(tr("&Edit"));
    editMenu->addAction(tr("&Preferences..."));

    // View menu
    auto* viewMenu = menuBar()->addMenu(tr("&View"));
    viewMenu->addAction(tr("&Refresh"))->setShortcut(QKeySequence::Refresh);
    viewMenu->addAction(tr("&Toggle Sidebar"));

    // Help menu
    auto* helpMenu = menuBar()->addMenu(tr("&Help"));
    helpMenu->addAction(tr("&About TablePro"));
}

void MainWindow::setupToolBar() {
    auto* toolbar = addToolBar(tr("Main Toolbar"));
    toolbar->setMovable(false);
    toolbar->setIconSize(QSize(24, 24));

    // Connection selector in toolbar
    toolbar->addWidget(m_connectionSelector);

    toolbar->addSeparator();

    auto* newQueryAction = toolbar->addAction(tr("New Query"));
    connect(newQueryAction, &QAction::triggered, this, &MainWindow::onNewQueryTab);

    auto* executeAction = toolbar->addAction(tr("Execute"));
    connect(executeAction, &QAction::triggered, this, &MainWindow::onExecuteQuery);
}

void MainWindow::setupStatusBar() {
    statusBar()->showMessage(tr("Ready"));
}

void MainWindow::setupCentralWidget() {
    // Sidebar layout
    auto* sidebarLayout = new QVBoxLayout(m_sidebarWidget);
    sidebarLayout->setContentsMargins(0, 0, 0, 0);
    sidebarLayout->addWidget(m_schemaTree);

    // Main splitter
    m_mainSplitter->addWidget(m_sidebarWidget);
    m_mainSplitter->addWidget(m_tabManager);
    m_mainSplitter->setSizes({250, 1150});

    setCentralWidget(m_mainSplitter);
}

void MainWindow::setupConnections() {
    connect(m_connectionSelector, &ConnectionSelector::connectionSelected,
            this, &MainWindow::onConnectionSelected);
    connect(m_connectionSelector, &ConnectionSelector::newConnectionRequested,
            this, &MainWindow::onNewConnection);

    connect(m_schemaTree, &SchemaTree::tableDoubleClicked,
            this, &MainWindow::onTableDoubleClicked);
}

void MainWindow::applyStyleSheet() {
    QFile styleFile(":/styles/dark.qss");
    if (styleFile.open(QIODevice::ReadOnly | QIODevice::Text)) {
        setStyleSheet(QString::fromUtf8(styleFile.readAll()));
    }
}

void MainWindow::onNewConnection() {
    // TODO: Show connection dialog
    QMessageBox::information(this, tr("New Connection"), tr("Connection dialog not yet implemented"));
}

void MainWindow::onConnectionSelected(const QString& connectionId) {
    m_currentConnectionId = connectionId;

    auto* driver = ConnectionManager::instance()->driver(connectionId);
    if (driver) {
        auto schemaFuture = driver->fetchSchema();
        schemaFuture.waitForFinished();

        auto schema = schemaFuture.result();
        m_schemaTree->setSchema(schema);

        auto tablesFuture = driver->fetchTables();
        tablesFuture.waitForFinished();

        auto tables = tablesFuture.result();
        m_schemaTree->setTables(tables);

        statusBar()->showMessage(tr("Connected to %1").arg(schema.databaseName));
    }
}

void MainWindow::onTableDoubleClicked(const QString& schema, const QString& table) {
    if (m_currentConnectionId.isEmpty()) {
        QMessageBox::warning(this, tr("No Connection"), tr("Please select a connection first"));
        return;
    }

    QString tabId = m_tabManager->createTableDataTab(m_currentConnectionId, schema, table);
    statusBar()->showMessage(tr("Opened table: %1.%2").arg(schema, table));
}

void MainWindow::onNewQueryTab() {
    if (m_currentConnectionId.isEmpty()) {
        QMessageBox::warning(this, tr("No Connection"), tr("Please select a connection first"));
        return;
    }

    m_tabManager->createQueryTab(m_currentConnectionId);
}

void MainWindow::onExecuteQuery() {
    // TODO: Execute current query
}

} // namespace tablepro
```

**Step 3: Commit MainWindow update**

```bash
git add src/ui/MainWindow.hpp src/ui/MainWindow.cpp
git commit -m "feat: Integrate SchemaTree and TabManager into MainWindow"
```

---

## Task 5: Update CMakeLists.txt

**Files:**
- Modify: `CMakeLists.txt`

**Step 1: Add UI widgets sources**

```cmake
set(TABLEPRO_SOURCES
    src/main.cpp
    src/ui/MainWindow.cpp
    src/ui/widgets/schema_tree.cpp
    src/ui/widgets/tab_manager.cpp
    src/ui/widgets/tab_content.hpp
    src/ui/widgets/connection_selector.cpp
    # ... existing core sources ...
)
```

**Step 2: Commit CMakeLists update**

```bash
git add CMakeLists.txt
git commit -m "build: Add UI widget sources to CMakeLists"
```

---

## Task 6: Verify Build

**Step 1: Build**

```bash
cmake --build build/debug -j$(nproc)
```

Expected: Build succeeds

**Step 2: Run**

```bash
./build/debug/tablepro
```

Expected: Window shows sidebar and tab area

---

## Acceptance Criteria

- [ ] SchemaTree displays database hierarchy
- [ ] TabManager creates and manages tabs
- [ ] ConnectionSelector shows active connections
- [ ] MainWindow integrates all widgets
- [ ] Dark theme applied consistently
- [ ] Toolbar functional
- [ ] Status bar updates on actions

---

**Phase 4 Complete.** Next: Phase 5 - Data Grid & Mutation