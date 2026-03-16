#pragma once

#include <QObject>
#include <QString>
#include <QPalette>
#include <QHash>

class QWidget;

namespace tablepro {

/**
 * Theme types supported by the application.
 */
enum class ThemeType {
    Light,
    Dark,
    System  // Follow system theme
};

/**
 * Manages application theming and styling.
 * Supports light/dark themes with customizable QSS stylesheets.
 */
class ThemeManager : public QObject
{
    Q_OBJECT

public:
    static ThemeManager* instance();

    void applyTheme(ThemeType theme);
    void applyStyleSheet(const QString& styleSheet);
    void setCustomStyle(const QString& name, const QString& styleSheet);

    ThemeType currentTheme() const { return m_currentTheme; }
    QString currentStyleSheet() const { return m_currentStyleSheet; }

    // Color accessors for custom widgets
    QColor backgroundColor() const;
    QColor textColor() const;
    QColor accentColor() const;
    QColor borderColor() const;
    QColor selectionColor() const;

signals:
    void themeChanged(ThemeType newTheme);

private:
    explicit ThemeManager(QObject* parent = nullptr);
    ~ThemeManager() override = default;

    QString loadStyleSheet(ThemeType theme) const;
    QString getLightStyleSheet() const;
    QString getDarkStyleSheet() const;

    static ThemeManager* s_instance;

    ThemeType m_currentTheme = ThemeType::Light;
    QString m_currentStyleSheet;
    QHash<QString, QString> m_customStyles;

    // Theme colors
    QPalette m_lightPalette;
    QPalette m_darkPalette;
};

} // namespace tablepro