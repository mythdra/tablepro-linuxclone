#include <QApplication>
#include <QSettings>
#include <QIcon>
#include "ui/mainwindow.h"

int main(int argc, char *argv[])
{
    QApplication app(argc, argv);

    // Set application properties
    app.setApplicationName("TablePro");
    app.setApplicationVersion("1.0.0");
    app.setOrganizationName("TablePro");
    app.setOrganizationDomain("tablepro.dev");

    // Set application icon
    app.setWindowIcon(QIcon(":/images/tablepro-icon.png"));

    // Create and show main window
    tablepro::MainWindow mainWindow;
    mainWindow.show();

    return app.exec();
}