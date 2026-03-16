#include "theme_manager.h"

#include <QApplication>
#include <QFile>
#include <QColor>
#include <QStyle>
#include <QStyleFactory>
#include <QDebug>

namespace tablepro {

ThemeManager* ThemeManager::s_instance = nullptr;

ThemeManager* ThemeManager::instance()
{
    if (!s_instance) {
        s_instance = new ThemeManager(qApp);
    }
    return s_instance;
}

ThemeManager::ThemeManager(QObject* parent)
    : QObject(parent)
{
    // Initialize light palette
    m_lightPalette.setColor(QPalette::Window, QColor(240, 240, 240));
    m_lightPalette.setColor(QPalette::WindowText, QColor(0, 0, 0));
    m_lightPalette.setColor(QPalette::Base, QColor(255, 255, 255));
    m_lightPalette.setColor(QPalette::AlternateBase, QColor(245, 245, 245));
    m_lightPalette.setColor(QPalette::ToolTipBase, QColor(255, 255, 220));
    m_lightPalette.setColor(QPalette::ToolTipText, QColor(0, 0, 0));
    m_lightPalette.setColor(QPalette::Text, QColor(0, 0, 0));
    m_lightPalette.setColor(QPalette::Button, QColor(240, 240, 240));
    m_lightPalette.setColor(QPalette::ButtonText, QColor(0, 0, 0));
    m_lightPalette.setColor(QPalette::Highlight, QColor(76, 163, 224));
    m_lightPalette.setColor(QPalette::HighlightedText, QColor(255, 255, 255));
    m_lightPalette.setColor(QPalette::Link, QColor(42, 130, 218));

    // Initialize dark palette
    m_darkPalette.setColor(QPalette::Window, QColor(53, 53, 53));
    m_darkPalette.setColor(QPalette::WindowText, QColor(255, 255, 255));
    m_darkPalette.setColor(QPalette::Base, QColor(25, 25, 25));
    m_darkPalette.setColor(QPalette::AlternateBase, QColor(53, 53, 53));
    m_darkPalette.setColor(QPalette::ToolTipBase, QColor(255, 255, 220));
    m_darkPalette.setColor(QPalette::ToolTipText, QColor(0, 0, 0));
    m_darkPalette.setColor(QPalette::Text, QColor(255, 255, 255));
    m_darkPalette.setColor(QPalette::Button, QColor(53, 53, 53));
    m_darkPalette.setColor(QPalette::ButtonText, QColor(255, 255, 255));
    m_darkPalette.setColor(QPalette::Highlight, QColor(42, 130, 218));
    m_darkPalette.setColor(QPalette::HighlightedText, QColor(0, 0, 0));
    m_darkPalette.setColor(QPalette::Link, QColor(42, 130, 218));
}

void ThemeManager::applyTheme(ThemeType theme)
{
    m_currentTheme = theme;

    QString styleSheet = loadStyleSheet(theme);
    applyStyleSheet(styleSheet);

    // Apply palette based on theme
    if (theme == ThemeType::Dark) {
        qApp->setPalette(m_darkPalette);
    } else {
        qApp->setPalette(m_lightPalette);
    }

    emit themeChanged(theme);
}

void ThemeManager::applyStyleSheet(const QString& styleSheet)
{
    m_currentStyleSheet = styleSheet;
    qApp->setStyleSheet(styleSheet);
}

void ThemeManager::setCustomStyle(const QString& name, const QString& styleSheet)
{
    m_customStyles[name] = styleSheet;
}

QString ThemeManager::loadStyleSheet(ThemeType theme) const
{
    if (theme == ThemeType::Dark) {
        return getDarkStyleSheet();
    }
    return getLightStyleSheet();
}

QString ThemeManager::getLightStyleSheet() const
{
    return R"(
        /* Main Window */
        QMainWindow {
            background-color: #f0f0f0;
        }

        /* Menu Bar */
        QMenuBar {
            background-color: #f0f0f0;
            color: #000000;
            border-bottom: 1px solid #c0c0c0;
        }

        QMenuBar::item {
            padding: 4px 8px;
        }

        QMenuBar::item:selected {
            background-color: #4ca3e0;
            color: white;
        }

        /* Menu */
        QMenu {
            background-color: #ffffff;
            border: 1px solid #c0c0c0;
        }

        QMenu::item {
            padding: 4px 32px 4px 20px;
        }

        QMenu::item:selected {
            background-color: #4ca3e0;
            color: white;
        }

        /* Tool Bar */
        QToolBar {
            background-color: #f5f5f5;
            border-bottom: 1px solid #c0c0c0;
            spacing: 4px;
            padding: 4px;
        }

        QToolBar::separator {
            background-color: #c0c0c0;
            width: 1px;
            margin: 4px;
        }

        /* Status Bar */
        QStatusBar {
            background-color: #f0f0f0;
            border-top: 1px solid #c0c0c0;
        }

        /* Tree View */
        QTreeView {
            background-color: #ffffff;
            border: 1px solid #c0c0c0;
            selection-background-color: #4ca3e0;
            selection-color: white;
        }

        QTreeView::item {
            padding: 2px;
            border: none;
        }

        QTreeView::item:selected {
            background-color: #4ca3e0;
            color: white;
        }

        QTreeView::item:hover {
            background-color: #e0e0e0;
        }

        /* Tab Widget */
        QTabWidget::pane {
            border: 1px solid #c0c0c0;
            background-color: #ffffff;
        }

        QTabBar::tab {
            background-color: #e0e0e0;
            border: 1px solid #c0c0c0;
            padding: 6px 12px;
            margin-right: 2px;
        }

        QTabBar::tab:selected {
            background-color: #ffffff;
            border-bottom-color: #ffffff;
        }

        QTabBar::tab:hover {
            background-color: #f0f0f0;
        }

        /* Line Edit */
        QLineEdit {
            background-color: #ffffff;
            border: 1px solid #c0c0c0;
            padding: 4px;
            border-radius: 2px;
        }

        QLineEdit:focus {
            border: 1px solid #4ca3e0;
        }

        /* Combo Box */
        QComboBox {
            background-color: #ffffff;
            border: 1px solid #c0c0c0;
            padding: 4px;
            border-radius: 2px;
        }

        QComboBox:hover {
            border: 1px solid #4ca3e0;
        }

        QComboBox::drop-down {
            border: none;
            width: 20px;
        }

        /* Push Button */
        QPushButton {
            background-color: #f0f0f0;
            border: 1px solid #c0c0c0;
            padding: 6px 12px;
            border-radius: 2px;
        }

        QPushButton:hover {
            background-color: #e0e0e0;
        }

        QPushButton:pressed {
            background-color: #d0d0d0;
        }

        QPushButton:default {
            background-color: #4ca3e0;
            color: white;
            border: 1px solid #3a8bc9;
        }

        QPushButton:default:hover {
            background-color: #3a8bc9;
        }

        /* Spin Box */
        QSpinBox {
            background-color: #ffffff;
            border: 1px solid #c0c0c0;
            padding: 4px;
            border-radius: 2px;
        }

        QSpinBox:focus {
            border: 1px solid #4ca3e0;
        }

        /* Group Box */
        QGroupBox {
            font-weight: bold;
            border: 1px solid #c0c0c0;
            border-radius: 4px;
            margin-top: 8px;
            padding-top: 8px;
        }

        QGroupBox::title {
            subcontrol-origin: margin;
            left: 8px;
            padding: 0 4px;
        }

        /* Scroll Bar */
        QScrollBar:vertical {
            background-color: #f0f0f0;
            width: 12px;
        }

        QScrollBar::handle:vertical {
            background-color: #c0c0c0;
            border-radius: 6px;
            min-height: 20px;
        }

        QScrollBar::handle:vertical:hover {
            background-color: #a0a0a0;
        }

        /* Splitter */
        QSplitter::handle {
            background-color: #c0c0c0;
        }

        QSplitter::handle:horizontal {
            width: 2px;
        }

        QSplitter::handle:vertical {
            height: 2px;
        }
    )";
}

QString ThemeManager::getDarkStyleSheet() const
{
    return R"(
        /* Main Window */
        QMainWindow {
            background-color: #353535;
        }

        /* Menu Bar */
        QMenuBar {
            background-color: #353535;
            color: #ffffff;
            border-bottom: 1px solid #555555;
        }

        QMenuBar::item {
            padding: 4px 8px;
        }

        QMenuBar::item:selected {
            background-color: #2a82da;
            color: white;
        }

        /* Menu */
        QMenu {
            background-color: #353535;
            color: #ffffff;
            border: 1px solid #555555;
        }

        QMenu::item {
            padding: 4px 32px 4px 20px;
        }

        QMenu::item:selected {
            background-color: #2a82da;
            color: white;
        }

        /* Tool Bar */
        QToolBar {
            background-color: #353535;
            border-bottom: 1px solid #555555;
            spacing: 4px;
            padding: 4px;
        }

        QToolBar::separator {
            background-color: #555555;
            width: 1px;
            margin: 4px;
        }

        /* Status Bar */
        QStatusBar {
            background-color: #353535;
            color: #ffffff;
            border-top: 1px solid #555555;
        }

        /* Tree View */
        QTreeView {
            background-color: #191919;
            color: #ffffff;
            border: 1px solid #555555;
            selection-background-color: #2a82da;
        }

        QTreeView::item {
            padding: 2px;
            border: none;
        }

        QTreeView::item:selected {
            background-color: #2a82da;
            color: white;
        }

        QTreeView::item:hover {
            background-color: #454545;
        }

        /* Tab Widget */
        QTabWidget::pane {
            border: 1px solid #555555;
            background-color: #191919;
        }

        QTabBar::tab {
            background-color: #353535;
            color: #ffffff;
            border: 1px solid #555555;
            padding: 6px 12px;
            margin-right: 2px;
        }

        QTabBar::tab:selected {
            background-color: #191919;
            border-bottom-color: #191919;
        }

        QTabBar::tab:hover {
            background-color: #454545;
        }

        /* Line Edit */
        QLineEdit {
            background-color: #191919;
            color: #ffffff;
            border: 1px solid #555555;
            padding: 4px;
            border-radius: 2px;
        }

        QLineEdit:focus {
            border: 1px solid #2a82da;
        }

        /* Combo Box */
        QComboBox {
            background-color: #353535;
            color: #ffffff;
            border: 1px solid #555555;
            padding: 4px;
            border-radius: 2px;
        }

        QComboBox:hover {
            border: 1px solid #2a82da;
        }

        QComboBox::drop-down {
            border: none;
            width: 20px;
        }

        QComboBox QAbstractItemView {
            background-color: #353535;
            color: #ffffff;
            selection-background-color: #2a82da;
        }

        /* Push Button */
        QPushButton {
            background-color: #353535;
            color: #ffffff;
            border: 1px solid #555555;
            padding: 6px 12px;
            border-radius: 2px;
        }

        QPushButton:hover {
            background-color: #454545;
        }

        QPushButton:pressed {
            background-color: #555555;
        }

        QPushButton:default {
            background-color: #2a82da;
            color: white;
            border: 1px solid #2a82da;
        }

        QPushButton:default:hover {
            background-color: #2370b5;
        }

        /* Spin Box */
        QSpinBox {
            background-color: #191919;
            color: #ffffff;
            border: 1px solid #555555;
            padding: 4px;
            border-radius: 2px;
        }

        QSpinBox:focus {
            border: 1px solid #2a82da;
        }

        /* Group Box */
        QGroupBox {
            color: #ffffff;
            font-weight: bold;
            border: 1px solid #555555;
            border-radius: 4px;
            margin-top: 8px;
            padding-top: 8px;
        }

        QGroupBox::title {
            subcontrol-origin: margin;
            left: 8px;
            padding: 0 4px;
        }

        /* Scroll Bar */
        QScrollBar:vertical {
            background-color: #353535;
            width: 12px;
        }

        QScrollBar::handle:vertical {
            background-color: #555555;
            border-radius: 6px;
            min-height: 20px;
        }

        QScrollBar::handle:vertical:hover {
            background-color: #666666;
        }

        /* Splitter */
        QSplitter::handle {
            background-color: #555555;
        }

        QSplitter::handle:horizontal {
            width: 2px;
        }

        QSplitter::handle:vertical {
            height: 2px;
        }

        /* Labels */
        QLabel {
            color: #ffffff;
        }

        /* Check Box */
        QCheckBox {
            color: #ffffff;
        }

        QCheckBox::indicator {
            width: 16px;
            height: 16px;
            border: 1px solid #555555;
            border-radius: 2px;
        }

        QCheckBox::indicator:checked {
            background-color: #2a82da;
            border: 1px solid #2a82da;
        }
    )";
}

QColor ThemeManager::backgroundColor() const
{
    if (m_currentTheme == ThemeType::Dark) {
        return QColor(53, 53, 53);
    }
    return QColor(240, 240, 240);
}

QColor ThemeManager::textColor() const
{
    if (m_currentTheme == ThemeType::Dark) {
        return QColor(255, 255, 255);
    }
    return QColor(0, 0, 0);
}

QColor ThemeManager::accentColor() const
{
    return QColor(42, 130, 218);
}

QColor ThemeManager::borderColor() const
{
    if (m_currentTheme == ThemeType::Dark) {
        return QColor(85, 85, 85);
    }
    return QColor(192, 192, 192);
}

QColor ThemeManager::selectionColor() const
{
    return QColor(42, 130, 218);
}

} // namespace tablepro