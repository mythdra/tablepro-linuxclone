#include "mainwindow.h"
#include "schema_tree_view.h"
#include "connection_dialog.h"
#include "theme_manager.h"
#include "core/database_driver.h"
#include "core/connection_config.h"
#include "drivers/postgres_driver.h"

#include <QMenuBar>
#include <QMenu>
#include <QToolBar>
#include <QAction>
#include <QSplitter>
#include <QTabWidget>
#include <QMessageBox>
#include <QFileDialog>
#include <QStatusBar>
#include <QLabel>
#include <QApplication>
#include <QCloseEvent>
#include <QSettings>
#include <QDir>
#include <QKeyEvent>
#include <QTextEdit>
#include <QVBoxLayout>
#include <QProgressBar>

namespace tablepro {

MainWindow::MainWindow(QWidget *parent)
    : QMainWindow(parent)
    , m_centralSplitter(nullptr)
    , m_schemaTree(nullptr)
    , m_tabWidget(nullptr)
    , m_settings(new QSettings(this))
    , m_statusConnectionLabel(new QLabel(this))
    , m_statusMessageLabel(new QLabel(this))
{
    // Apply default theme
    ThemeManager::instance()->applyTheme(ThemeType::Light);

    createMenus();
    createToolBars();
    createStatusBar();
    createCentralWidget();
    createDockWidgets();
    setupConnections();

    setWindowTitle("TablePro - Database Client");
    readSettings();
}

MainWindow::~MainWindow()
{
}

void MainWindow::closeEvent(QCloseEvent *event)
{
    saveTabState();
    writeSettings();
    event->accept();
}

void MainWindow::keyPressEvent(QKeyEvent *event)
{
    // Ctrl+W - close current tab
    if (event->modifiers() == Qt::ControlModifier && event->key() == Qt::Key_W) {
        closeCurrentTab();
        return;
    }

    // Ctrl+Tab - next tab
    if (event->modifiers() == Qt::ControlModifier && event->key() == Qt::Key_Tab) {
        nextTab();
        return;
    }

    // Ctrl+Shift+Tab - previous tab
    if (event->modifiers() == (Qt::ControlModifier | Qt::ShiftModifier) && event->key() == Qt::Key_Backtab) {
        previousTab();
        return;
    }

    QMainWindow::keyPressEvent(event);
}

void MainWindow::createMenus()
{
    // File menu
    QMenu *fileMenu = menuBar()->addMenu(tr("&File"));

    m_newConnectionAct = new QAction(tr("&New Connection"), this);
    m_newConnectionAct->setShortcuts(QKeySequence::New);
    m_newConnectionAct->setStatusTip(tr("Create a new database connection"));
    connect(m_newConnectionAct, &QAction::triggered, this, &MainWindow::newConnection);
    fileMenu->addAction(m_newConnectionAct);

    m_openAct = new QAction(tr("&Open"), this);
    m_openAct->setShortcuts(QKeySequence::Open);
    m_openAct->setStatusTip(tr("Open a database connection"));
    connect(m_openAct, &QAction::triggered, this, &MainWindow::openConnection);
    fileMenu->addAction(m_openAct);

    m_saveAct = new QAction(tr("&Save"), this);
    m_saveAct->setShortcuts(QKeySequence::Save);
    m_saveAct->setStatusTip(tr("Save the current connection"));
    connect(m_saveAct, &QAction::triggered, this, &MainWindow::saveConnection);
    fileMenu->addAction(m_saveAct);

    fileMenu->addSeparator();

    m_exitAct = new QAction(tr("E&xit"), this);
    m_exitAct->setShortcuts(QKeySequence::Quit);
    m_exitAct->setStatusTip(tr("Exit the application"));
    connect(m_exitAct, &QAction::triggered, this, &MainWindow::close);
    fileMenu->addAction(m_exitAct);

    // Edit menu
    QMenu *editMenu = menuBar()->addMenu(tr("&Edit"));
    QAction *undoAct = new QAction(tr("&Undo"), this);
    undoAct->setShortcuts(QKeySequence::Undo);
    editMenu->addAction(undoAct);

    QAction *redoAct = new QAction(tr("&Redo"), this);
    redoAct->setShortcuts(QKeySequence::Redo);
    editMenu->addAction(redoAct);

    editMenu->addSeparator();

    QAction *cutAct = new QAction(tr("Cu&t"), this);
    cutAct->setShortcuts(QKeySequence::Cut);
    editMenu->addAction(cutAct);

    QAction *copyAct = new QAction(tr("&Copy"), this);
    copyAct->setShortcuts(QKeySequence::Copy);
    editMenu->addAction(copyAct);

    QAction *pasteAct = new QAction(tr("&Paste"), this);
    pasteAct->setShortcuts(QKeySequence::Paste);
    editMenu->addAction(pasteAct);

    // View menu
    QMenu *viewMenu = menuBar()->addMenu(tr("&View"));

    m_fullScreenAct = new QAction(tr("&Full Screen"), this);
    m_fullScreenAct->setShortcut(Qt::Key_F11);
    connect(m_fullScreenAct, &QAction::triggered, this, &MainWindow::toggleFullScreen);
    viewMenu->addAction(m_fullScreenAct);

    m_toggleThemeAct = new QAction(tr("Toggle &Theme"), this);
    m_toggleThemeAct->setShortcut(QKeySequence(Qt::CTRL | Qt::Key_T));
    connect(m_toggleThemeAct, &QAction::triggered, this, &MainWindow::toggleTheme);
    viewMenu->addAction(m_toggleThemeAct);

    // Help menu
    QMenu *helpMenu = menuBar()->addMenu(tr("&Help"));

    m_aboutAct = new QAction(tr("&About"), this);
    m_aboutAct->setStatusTip(tr("Show the application's About box"));
    connect(m_aboutAct, &QAction::triggered, this, &MainWindow::showAboutDialog);
    helpMenu->addAction(m_aboutAct);

    QAction *aboutQtAct = new QAction(tr("About &Qt"), this);
    aboutQtAct->setStatusTip(tr("Show the Qt library's About box"));
    connect(aboutQtAct, &QAction::triggered, qApp, &QApplication::aboutQt);
    helpMenu->addAction(aboutQtAct);
}

void MainWindow::createToolBars()
{
    QToolBar *toolbar = addToolBar(tr("Main"));
    toolbar->addAction(m_newConnectionAct);
    toolbar->addAction(m_openAct);
    toolbar->addAction(m_saveAct);
    toolbar->addSeparator();
    toolbar->addAction(m_toggleThemeAct);
}

void MainWindow::createStatusBar()
{
    m_statusConnectionLabel->setText(tr("Not Connected"));
    statusBar()->addWidget(m_statusConnectionLabel);

    statusBar()->addPermanentWidget(m_statusMessageLabel);
    m_statusMessageLabel->setText(tr("Ready"));
}

void MainWindow::createCentralWidget()
{
    m_centralSplitter = new QSplitter(Qt::Horizontal, this);

    // Schema tree (left panel)
    m_schemaTree = new SchemaTreeView(m_centralSplitter);
    m_schemaTree->setMinimumWidth(200);
    m_schemaTree->setMaximumWidth(400);

    connect(m_schemaTree, &SchemaTreeView::tableSelected, this, &MainWindow::onTableSelected);

    // Tab widget for editors/results (right panel)
    m_tabWidget = new QTabWidget(m_centralSplitter);
    m_tabWidget->setTabsClosable(true);
    m_tabWidget->setMovable(true);
    m_tabWidget->setDocumentMode(true);

    connect(m_tabWidget, &QTabWidget::tabCloseRequested, this, [this](int index) {
        QWidget* tab = m_tabWidget->widget(index);
        m_tabWidget->removeTab(index);
        delete tab;
    });

    // Add welcome tab
    QWidget *welcomeTab = new QWidget();
    auto* welcomeLayout = new QVBoxLayout(welcomeTab);
    QLabel* welcomeLabel = new QLabel(
        tr("<h2>Welcome to TablePro</h2>"
           "<p>Connect to a database to get started.</p>"
           "<p>Use <b>File → New Connection</b> to create a new database connection.</p>")
    );
    welcomeLabel->setAlignment(Qt::AlignCenter);
    welcomeLayout->addWidget(welcomeLabel);
    m_tabWidget->addTab(welcomeTab, tr("Welcome"));

    setCentralWidget(m_centralSplitter);
}

void MainWindow::createDockWidgets()
{
    setCorner(Qt::TopLeftCorner, Qt::LeftDockWidgetArea);
    setCorner(Qt::BottomLeftCorner, Qt::BottomDockWidgetArea);
    setCorner(Qt::TopRightCorner, Qt::RightDockWidgetArea);
    setCorner(Qt::BottomRightCorner, Qt::BottomDockWidgetArea);
}

void MainWindow::setupConnections()
{
}

void MainWindow::readSettings()
{
    QByteArray geometry = m_settings->value("mainWindow/geometry").toByteArray();
    if (!geometry.isEmpty()) {
        restoreGeometry(geometry);
    } else {
        resize(1200, 800);
    }

    QByteArray state = m_settings->value("mainWindow/state").toByteArray();
    if (!state.isEmpty()) {
        restoreState(state);
    }

    QByteArray splitterState = m_settings->value("mainWindow/splitterState").toByteArray();
    if (!splitterState.isEmpty() && m_centralSplitter) {
        m_centralSplitter->restoreState(splitterState);
    }

    // Restore theme
    QString theme = m_settings->value("mainWindow/theme", "light").toString();
    ThemeManager::instance()->applyTheme(theme == "dark" ? ThemeType::Dark : ThemeType::Light);

    restoreTabState();
}

void MainWindow::writeSettings()
{
    m_settings->setValue("mainWindow/geometry", saveGeometry());
    m_settings->setValue("mainWindow/state", saveState());
    m_settings->setValue("mainWindow/theme",
        ThemeManager::instance()->currentTheme() == ThemeType::Dark ? "dark" : "light");

    if (m_centralSplitter) {
        m_settings->setValue("mainWindow/splitterState", m_centralSplitter->saveState());
    }

    m_settings->sync();
}

void MainWindow::saveTabState()
{
    // Save open tabs
    m_settings->beginWriteArray("tabs");
    for (int i = 0; i < m_tabWidget->count(); ++i) {
        m_settings->setArrayIndex(i);
        m_settings->setValue("title", m_tabWidget->tabText(i));
    }
    m_settings->endArray();
}

void MainWindow::restoreTabState()
{
    // Restore tabs (simplified - just restore tab count for now)
    int tabCount = m_settings->beginReadArray("tabs");
    m_settings->endArray();

    // Currently we just have the welcome tab
    // In future, restore actual query tabs
}

void MainWindow::addQueryTab(const QString& tableName)
{
    QWidget* queryTab = new QWidget();
    auto* layout = new QVBoxLayout(queryTab);

    auto* editor = new QTextEdit();
    if (!tableName.isEmpty()) {
        editor->setPlainText(QString("SELECT * FROM %1 LIMIT 100;").arg(tableName));
    } else {
        editor->setPlaceholderText(tr("Enter SQL query here..."));
    }

    layout->addWidget(editor);

    int index = m_tabWidget->addTab(queryTab, tableName.isEmpty() ? tr("Query") : tableName);
    m_tabWidget->setCurrentIndex(index);
}

void MainWindow::updateWindowTitle()
{
    QString title = "TablePro";
    if (m_currentDriver && m_currentDriver->isConnected()) {
        title += QString(" - %1@%2/%3")
            .arg(m_currentDriver->lastError().isEmpty() ? "" : "")
            .arg("")
            .arg("");
    }
    setWindowTitle(title);
}

void MainWindow::newConnection()
{
    ConnectionDialog dialog(this);

    if (m_currentDriver) {
        dialog.setDriver(m_currentDriver);
    }

    connect(&dialog, &ConnectionDialog::connectionCreated, this, &MainWindow::onConnectionCreated);

    if (dialog.exec() == QDialog::Accepted) {
        ConnectionConfig config = dialog.getConnectionConfig();

        // Create driver based on type
        if (config.type == DatabaseType::PostgreSQL) {
            m_currentDriver = new PostgresDriver(this);
        }

        if (m_currentDriver) {
            bool success = m_currentDriver->connect(config);

            if (success) {
                m_schemaTree->setDriver(m_currentDriver);
                m_schemaTree->loadDatabase(config.database);
                m_statusConnectionLabel->setText(tr("Connected to %1").arg(config.database));
                m_statusMessageLabel->setText(tr("Connection established"));
            } else {
                QMessageBox::critical(this, tr("Connection Failed"),
                    tr("Failed to connect: %1").arg(m_currentDriver->lastError()));
            }
        }
    }
}

void MainWindow::openConnection()
{
    m_statusMessageLabel->setText(tr("Opening connection..."));
    // TODO: Implement open connection from saved list
    newConnection();
}

void MainWindow::saveConnection()
{
    m_statusMessageLabel->setText(tr("Saving connection..."));
    // Connection is saved in ConnectionDialog
}

void MainWindow::showAboutDialog()
{
    QMessageBox::about(this, tr("About TablePro"),
        tr("TablePro Database Client\n\n"
           "Version 1.0.0\n\n"
           "A cross-platform database client supporting PostgreSQL, MySQL, SQLite and other databases.\n\n"
           "Built with Qt %1")
        .arg(QT_VERSION_STR));
}

void MainWindow::toggleFullScreen()
{
    if (isFullScreen()) {
        showNormal();
    } else {
        showFullScreen();
    }
}

void MainWindow::toggleTheme()
{
    ThemeManager* themeManager = ThemeManager::instance();

    if (themeManager->currentTheme() == ThemeType::Light) {
        themeManager->applyTheme(ThemeType::Dark);
        m_statusMessageLabel->setText(tr("Theme: Dark"));
    } else {
        themeManager->applyTheme(ThemeType::Light);
        m_statusMessageLabel->setText(tr("Theme: Light"));
    }
}

void MainWindow::onTableSelected(const QString& schema, const QString& table)
{
    QString fullTableName = schema.isEmpty() ? table : QString("%1.%2").arg(schema, table);
    addQueryTab(fullTableName);
    m_statusMessageLabel->setText(tr("Selected table: %1").arg(fullTableName));
}

void MainWindow::onConnectionCreated(const ConnectionConfig& config)
{
    m_statusMessageLabel->setText(tr("Connection '%1' created").arg(config.toString()));
}

void MainWindow::closeCurrentTab()
{
    int index = m_tabWidget->currentIndex();
    if (index >= 0) {
        QWidget* tab = m_tabWidget->widget(index);
        m_tabWidget->removeTab(index);
        delete tab;
    }
}

void MainWindow::nextTab()
{
    int count = m_tabWidget->count();
    if (count <= 1) return;

    int current = m_tabWidget->currentIndex();
    m_tabWidget->setCurrentIndex((current + 1) % count);
}

void MainWindow::previousTab()
{
    int count = m_tabWidget->count();
    if (count <= 1) return;

    int current = m_tabWidget->currentIndex();
    m_tabWidget->setCurrentIndex((current - 1 + count) % count);
}

} // namespace tablepro