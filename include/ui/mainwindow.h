#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QMainWindow>
#include <QSplitter>
#include <QTabWidget>
#include <QMenuBar>
#include <QToolBar>
#include <QStatusBar>
#include <QSettings>
#include <QLabel>

namespace tablepro {

class SchemaTreeView;
class ConnectionDialog;
class DatabaseDriver;
class ConnectionConfig;
class ThemeManager;

class MainWindow : public QMainWindow
{
    Q_OBJECT

public:
    explicit MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

protected:
    void closeEvent(QCloseEvent *event) override;
    void keyPressEvent(QKeyEvent *event) override;

private slots:
    void newConnection();
    void openConnection();
    void saveConnection();
    void showAboutDialog();
    void toggleFullScreen();
    void toggleTheme();
    void onTableSelected(const QString& schema, const QString& table);
    void onConnectionCreated(const ConnectionConfig& config);
    void closeCurrentTab();
    void nextTab();
    void previousTab();

private:
    void createMenus();
    void createToolBars();
    void createStatusBar();
    void createCentralWidget();
    void createDockWidgets();
    void setupConnections();
    void readSettings();
    void writeSettings();
    void saveTabState();
    void restoreTabState();
    void addQueryTab(const QString& tableName = QString());
    void updateWindowTitle();

    QSplitter *m_centralSplitter;
    SchemaTreeView *m_schemaTree;
    QTabWidget *m_tabWidget;
    QSettings *m_settings;
    QLabel *m_statusConnectionLabel;
    QLabel *m_statusMessageLabel;

    // Drivers
    DatabaseDriver *m_currentDriver = nullptr;

    // Menu actions
    QAction *m_newConnectionAct;
    QAction *m_openAct;
    QAction *m_saveAct;
    QAction *m_exitAct;
    QAction *m_aboutAct;
    QAction *m_toggleThemeAct;

    // View menu actions
    QAction *m_fullScreenAct;
};

} // namespace tablepro

#endif // MAINWINDOW_H