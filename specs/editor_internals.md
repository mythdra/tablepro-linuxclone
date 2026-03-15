# Editor Internals (QScintilla + Qt)

## Overview
The SQL Editor uses **QScintilla** — a Qt port of Scintilla, the same editing component used in Notepad++ and many IDEs. It provides enterprise-grade editing capabilities with native Qt integration.

## Key Capabilities

| Feature | Implementation |
|---------|----------------|
| Syntax Highlighting | QScintilla lexer (`QsciLexerSQL`) |
| Autocomplete | `QsciAPIs` with table/column names from C++ backend |
| Multi-cursor | Limited (QScintilla doesn't support true multi-cursor) |
| Vim Mode | QScintilla Vim extension or custom key bindings |
| Minimap | Not built-in (can implement with secondary QScintilla) |
| Find & Replace | Built-in (`findFirst()`, `replace()`) |
| Line Numbers | Built-in margin |
| Word Wrap | Configurable (`SendScintilla(SCI_SETWRAPMODE)`) |
| Code Folding | Built-in (`SendScintilla(SCI_TOGGLEFOLD)`) |

## QScintilla Setup

```cpp
// src/ui/Editor/SqlEditor.hpp
#pragma once

#include <Qsci/qsciscintilla.h>
#include <Qsci/qscilexersql.h>
#include <Qsci/qsciapis.h>
#include <Qsci/qsicommands.h>

namespace tablepro {

class SqlEditor : public QsciScintilla {
    Q_OBJECT

public:
    explicit SqlEditor(QWidget* parent = nullptr);

    // Configuration
    void setupLexer();
    void setupAPIs();
    void setupMargins();
    void setupIndentation();
    void setupCaret();
    void setupBraceMatching();

    // SQL-specific features
    void executeCurrentStatement();
    void executeAllStatements();
    void executeSelection();
    QString currentStatement() const;
    QStringList splitStatements() const;

    // Formatting
    void formatSQL();

    // Vim mode
    void setVimMode(bool enabled);

signals:
    void executeRequested(const QString& sql, bool all);
    void cursorPositionChanged(int line, int column);

private:
    QsciLexerSQL* m_lexer;
    QsciAPIs* m_apis;
    bool m_vimMode{false};

    // Find current statement boundaries
    std::pair<int, int> findStatementBoundaries(int position) const;
};

} // namespace tablepro
```

```cpp
// src/ui/Editor/SqlEditor.cpp
#include "SqlEditor.hpp"

SqlEditor::SqlEditor(QWidget* parent)
    : QsciScintilla(parent)
    , m_lexer(new QsciLexerSQL(this))
    , m_apis(new QsciAPIs(m_lexer))
{
    setupLexer();
    setupAPIs();
    setupMargins();
    setupIndentation();
    setupCaret();
    setupBraceMatching();

    // Connect cursor position signal
    connect(this, &QsciScintilla::cursorPositionChanged,
            this, [this](int line, int col) {
        emit cursorPositionChanged(line, col);
    });
}

void SqlEditor::setupLexer() {
    setLexer(m_lexer);

    // Custom colors (dark theme example)
    m_lexer->setColor(QColor("#569CD6"), QsciLexerSQL::Keyword);
    m_lexer->setColor(QColor("#CE9178"), QsciLexerSQL::DoubleQuotedString);
    m_lexer->setColor(QColor("#CE9178"), QsciLexerSQL::SingleQuotedString);
    m_lexer->setColor(QColor("#6A9955"), QsciLexerSQL::Comment);
    m_lexer->setColor(QColor("#6A9955"), QsciLexerSQL::CommentLine);
    m_lexer->setColor(QColor("#B5CEA8"), QsciLexerSQL::Number);
    m_lexer->setColor(QColor("#9CDCFE"), QsciLexerSQL::Identifier);

    // Font
    QFont font("JetBrains Mono", 14);
    m_lexer->setFont(font);

    // Paper (background)
    m_lexer->setPaper(QColor("#1E1E2E"));
    setPaper(m_lexer->paper());

    // Auto-indent
    m_lexer->setAutoIndentStyle(QsciLexerSQL::AutoIndent);
}

void SqlEditor::setupAPIs() {
    // SQL keywords are built into the lexer
    // Custom APIs (table names, column names) loaded dynamically

    connect(this, &SqlEditor::textChanged,
            this, [this]() {
        // Trigger autocomplete refresh if needed
    });
}

void SqlEditor::setupMargins() {
    // Line numbers margin
    setMarginLineNumbers(0, true);
    setMarginWidth(0, QFontMetrics(font()).horizontalAdvance("9999") + 6);

    // Folding margin (optional)
    setMarginLineNumbers(1, false);
    setMarginWidth(1, 12);
    setFolding(QsciScintilla::BoxedTreeFoldStyle, 1);

    // Marker margin for breakpoints (future feature)
    setMarginWidth(2, 0);
}

void SqlEditor::setupIndentation() {
    setIndentationsUseTabs(false);
    setIndentationWidth(4);
    setAutoIndent(true);
    setBackspaceUnindents(true);
}

void SqlEditor::setupCaret() {
    setCaretWidth(2);
    setCaretLineVisible(true);
    setCaretLineColor(QColor("#333344"));
    setCursorWidth(2);
}

void SqlEditor::setupBraceMatching() {
    setBraceMatching(QsciScintilla::SloppyBraceMatch);
}

void SqlEditor::setupAPIs() {
    // Load table and column names for autocomplete
    // Called when connection changes
}

void SqlEditor::loadSchemaAPIs(
    const QString& connectionId,
    const QStringList& tableNames,
    const QMap<QString, QStringList>& tableColumns)
{
    m_apis->clear();

    // Add table names
    for (const auto& table : tableNames) {
        m_apis->add(table);
    }

    // Add column names with table prefix
    for (auto it = tableColumns.begin(); it != tableColumns.end(); ++it) {
        for (const auto& column : it.value()) {
            m_apis->add(it.key() + "." + column);
        }
    }

    // Add SQL keywords (if not using built-in)
    static const QStringList keywords = {
        "SELECT", "FROM", "WHERE", "INSERT", "UPDATE", "DELETE",
        "JOIN", "LEFT", "RIGHT", "INNER", "OUTER", "ON",
        "GROUP BY", "ORDER BY", "HAVING", "LIMIT", "OFFSET",
        "CREATE", "ALTER", "DROP", "TABLE", "INDEX", "VIEW",
        "AS", "AND", "OR", "NOT", "IN", "BETWEEN", "LIKE",
        "NULL", "IS NULL", "IS NOT NULL", "DISTINCT", "ALL"
    };
    for (const auto& kw : keywords) {
        m_apis->add(kw);
    }

    m_apis->prepare();
}
```

## Statement Execution Logic

```cpp
void SqlEditor::executeCurrentStatement() {
    // Check if there's a selection
    int start, end;
    bool hasSelection = getSelection(&start, &end);

    QString sql;
    if (hasSelection && end > start) {
        // Execute selected text
        sql = text().mid(start, end - start);
    } else {
        // Find statement under cursor
        auto [stmtStart, stmtEnd] = findStatementBoundaries(position());
        sql = text().mid(stmtStart, stmtEnd - stmtStart);
    }

    emit executeRequested(sql.trimmed(), false);
}

void SqlEditor::executeAllStatements() {
    emit executeRequested(text(), true);
}

std::pair<int, int> SqlEditor::findStatementBoundaries(int position) const {
    // Find the statement containing the cursor position
    // Splits on semicolons outside of strings and comments

    const QString fullText = text();
    int length = fullText.length();

    // Find previous semicolon (or start)
    int start = 0;
    for (int i = position - 1; i >= 0; --i) {
        QChar c = fullText[i];
        if (c == ';' && !isInsideStringOrComment(fullText, i)) {
            start = i + 1;
            break;
        }
    }

    // Find next semicolon (or end)
    int end = length;
    for (int i = position; i < length; ++i) {
        QChar c = fullText[i];
        if (c == ';' && !isInsideStringOrComment(fullText, i)) {
            end = i;
            break;
        }
    }

    return {start, end};
}

bool SqlEditor::isInsideStringOrComment(const QString& text, int position) const {
    // Check if position is inside a string literal or comment
    // This is a simplified version - full implementation needs state machine

    int singleQuotes = 0;
    int doubleQuotes = 0;

    for (int i = 0; i < position && i < text.length(); ++i) {
        QChar c = text[i];
        QChar next = (i + 1 < text.length()) ? text[i + 1] : QChar();

        // Skip escaped quotes
        if (c == '\\' || (i > 0 && text[i-1] == '\\'))
            continue;

        if (c == '\'' && doubleQuotes % 2 == 0)
            singleQuotes++;
        else if (c == '"' && singleQuotes % 2 == 0)
            doubleQuotes++;
    }

    // Inside string if odd number of quotes
    if (singleQuotes % 2 != 0 || doubleQuotes % 2 != 0)
        return true;

    // Check for comment
    if (position >= 1) {
        QString before = text.mid(position - 2, 2);
        if (before == "--")
            return true;
        // Multi-line comment check would need more logic
    }

    return false;
}
```

## Statement Splitting

```cpp
QStringList SqlEditor::splitStatements() const {
    QStringList statements;
    const QString fullText = text();

    int start = 0;
    for (int i = 0; i < fullText.length(); ++i) {
        if (fullText[i] == ';' && !isInsideStringOrComment(fullText, i)) {
            QString stmt = fullText.mid(start, i - start).trimmed();
            if (!stmt.isEmpty())
                statements.append(stmt);
            start = i + 1;
        }
    }

    // Add remaining
    QString remaining = fullText.mid(start).trimmed();
    if (!remaining.isEmpty())
        statements.append(remaining);

    return statements;
}
```

## Theme Customization

```cpp
void SqlEditor::setDarkTheme() {
    // Editor background
    setPaper(QColor("#1E1E2E"));

    // Lexer colors
    m_lexer->setColor(QColor("#CDD6F4"), QsciLexerSQL::Default);
    m_lexer->setColor(QColor("#569CD6"), QsciLexerSQL::Keyword);
    m_lexer->setColor(QColor("#CE9178"), QsciLexerSQL::DoubleQuotedString);
    m_lexer->setColor(QColor("#CE9178"), QsciLexerSQL::SingleQuotedString);
    m_lexer->setColor(QColor("#6A9955"), QsciLexerSQL::Comment);
    m_lexer->setColor(QColor("#6A9955"), QsciLexerSQL::CommentLine);
    m_lexer->setColor(QColor("#B5CEA8"), QsciLexerSQL::Number);
    m_lexer->setColor(QColor("#9CDCFE"), QsciLexerSQL::Identifier);
    m_lexer->setColor(QColor("#DCDCAA"), QsciLexerSQL::Operator);

    // Margin colors
    setMarginBackgroundColor(0, QColor("#1E1E2E"));
    m_lexer->setPaper(QColor("#1E1E2E"));

    // Caret
    setCaretColor(QColor("#FFFFFF"));
    setCaretLineBackgroundColor(QColor("#333344"));
}

void SqlEditor::setLightTheme() {
    setPaper(QColor("#FFFFFF"));

    m_lexer->setColor(QColor("#000000"), QsciLexerSQL::Default);
    m_lexer->setColor(QColor("#0000FF"), QsciLexerSQL::Keyword);
    m_lexer->setColor(QColor("#A31515"), QsciLexerSQL::DoubleQuotedString);
    m_lexer->setColor(QColor("#A31515"), QsciLexerSQL::SingleQuotedString);
    m_lexer->setColor(QColor("#008000"), QsciLexerSQL::Comment);
    m_lexer->setColor(QColor("#008000"), QsciLexerSQL::CommentLine);
    m_lexer->setColor(QColor("#098658"), QsciLexerSQL::Number);

    setMarginBackgroundColor(0, QColor("#F3F3F3"));
    setCaretColor(QColor("#000000"));
    setCaretLineBackgroundColor(QColor("#E8F2FE"));
}
```

## Vim Mode

```cpp
void SqlEditor::setVimMode(bool enabled) {
    m_vimMode = enabled;

    if (enabled) {
        // Enable QScintilla Vim extension
        // Note: QScintilla doesn't have built-in Vim mode
        // Would need to use QScintilla's command interface to emulate
        // Or integrate with a Vim emulation library

        // For now, set up basic Vim-like key bindings
        installEventFilter(this);
    } else {
        removeEventFilter(this);
    }
}

bool SqlEditor::eventFilter(QObject* watched, QEvent* event) override {
    if (m_vimMode && event->type() == QEvent::KeyPress) {
        auto* keyEvent = static_cast<QKeyEvent*>(event);

        // Basic Vim-like navigation (normal mode simulation)
        // Full implementation would need proper state machine
        switch (keyEvent->key()) {
            case Qt::Key_H:
                if (keyEvent->modifiers() == Qt::NoModifier) {
                    sendScintilla(SCI_CHARLEFT);
                    return true;
                }
                break;
            case Qt::Key_J:
                if (keyEvent->modifiers() == Qt::NoModifier) {
                    sendScintilla(SCI_LINEDOWN);
                    return true;
                }
                break;
            case Qt::Key_K:
                if (keyEvent->modifiers() == Qt::NoModifier) {
                    sendScintilla(SCI_LINEUP);
                    return true;
                }
                break;
            case Qt::Key_L:
                if (keyEvent->modifiers() == Qt::NoModifier) {
                    sendScintilla(SCI_CHARRIGHT);
                    return true;
                }
                break;
        }
    }

    return QsciScintilla::eventFilter(watched, event);
}
```

## SQL Formatting

```cpp
void SqlEditor::formatSQL() {
    // Option 1: Simple regex-based formatter
    QString formatted = SQLFormatterService::format(text());

    // Option 2: External formatter (if available)
    // QProcess process;
    // process.start("sqlfmt", {"-", "--dialect=postgres"});
    // process.write(text().toUtf8());
    // process.closeWriteChannel();
    // process.waitForFinished();
    // QString formatted = QString::fromUtf8(process.readAllStandardOutput());

    int pos = position();
    setText(formatted);
    setPosition(pos);  // Try to preserve cursor position
}
```

## Keyboard Shortcuts

```cpp
// In MainWindow or parent widget
void setupEditorShortcuts(SqlEditor* editor) {
    // Execute current statement: Ctrl+R / F5
    auto* execShortcut = new QShortcut(QKeySequence("Ctrl+R"), editor);
    connect(execShortcut, &QShortcut::activated,
            editor, &SqlEditor::executeCurrentStatement);

    auto* execAllShortcut = new QShortcut(QKeySequence("Ctrl+Shift+R"), editor);
    connect(execAllShortcut, &QShortcut::activated,
            editor, &SqlEditor::executeAllStatements);

    // Format SQL: Ctrl+Shift+F
    auto* formatShortcut = new QShortcut(QKeySequence("Ctrl+Shift+F"), editor);
    connect(formatShortcut, &QShortcut::activated,
            editor, &SqlEditor::formatSQL);

    // Find: Ctrl+F
    auto* findShortcut = new QShortcut(QKeySequence::Find, editor);
    connect(findShortcut, &QShortcut::activated,
            editor, [editor]() { editor->findFirst("", false, false, false, true); });
}
```
