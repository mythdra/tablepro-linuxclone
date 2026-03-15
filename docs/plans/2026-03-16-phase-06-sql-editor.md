# Phase 6: SQL Editor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build SQL editor with QScintilla, syntax highlighting, statement execution, autocomplete, and find/replace.

**Architecture:** QScintilla widget with QsciLexerSQL for highlighting. Statement splitter for executing individual statements. Autocomplete via QsciAPIs with schema metadata.

**Tech Stack:** C++20, Qt 6.6, QScintilla 2.14+

---

## Task 1: Add QScintilla Dependency

**Files:**
- Modify: `vcpkg.json`
- Modify: `CMakeLists.txt`

**Step 1: Add to vcpkg.json**

```json
{
  "name": "qscintilla",
  "version>=": "2.14.0"
}
```

**Step 2: Add to CMakeLists.txt**

```cmake
find_package(QScintilla REQUIRED)
target_link_libraries(tablepro PRIVATE QScintilla::QScintilla)
```

**Step 3: Commit dependency**

```bash
git add vcpkg.json CMakeLists.txt
git commit -m "build: Add QScintilla dependency"
```

---

## Task 2: SQL Editor Widget

**Files:**
- Create: `src/ui/editor/sql_editor.hpp`
- Create: `src/ui/editor/sql_editor.cpp`

**Step 1: Create sql_editor.hpp**

```cpp
#pragma once

#include <Qsci/qsciscintilla.h>
#include <Qsci/qscilexersql.h>
#include <Qsci/qsciapis.h>
#include <QStringList>

namespace tablepro {

class SqlEditor : public QsciScintilla {
    Q_OBJECT

public:
    explicit SqlEditor(QWidget* parent = nullptr);

    // Content
    void setSql(const QString& sql);
    QString sql() const;
    QString selectedText() const;
    QString currentStatement() const;
    QString allStatements() const;

    // Execution
    void executeCurrentStatement();
    void executeAllStatements();
    void executeSelected();

    // Formatting
    void formatSql();

    // Autocomplete
    void setTableNames(const QStringList& tables);
    void setColumnNames(const QString& table, const QStringList& columns);
    void refreshAutocomplete();

signals:
    void executeRequested(const QString& sql);
    void executeAllRequested(const QStringList& statements);
    void textChanged();

private:
    void setupLexer();
    void setupEditor();
    void setupMargins();
    void setupShortcuts();

    QsciLexerSQL* m_lexer;
    QsciAPIs* m_apis;
    QStringList m_tableNames;
    QMap<QString, QStringList> m_columnNames;
};

} // namespace tablepro
```

**Step 2: Create sql_editor.cpp**

```cpp
#include "sql_editor.hpp"
#include <QShortcut>
#include <QKeyEvent>
#include <QRegularExpression>

namespace tablepro {

SqlEditor::SqlEditor(QWidget* parent)
    : QsciScintilla(parent)
    , m_lexer(new QsciLexerSQL(this))
    , m_apis(new QsciAPIs(m_lexer))
{
    setupLexer();
    setupEditor();
    setupMargins();
    setupShortcuts();
}

void SqlEditor::setupLexer() {
    // SQL keywords
    m_lexer->setKeywords(QsciLexerSQL::Keywords,
        "SELECT FROM WHERE JOIN ON LEFT RIGHT INNER OUTER AND OR NOT IN IS NULL LIKE "
        "BETWEEN EXISTS CASE WHEN THEN ELSE END AS ORDER BY ASC DESC GROUP BY HAVING "
        "LIMIT OFFSET UNION ALL DISTINCT INSERT INTO VALUES UPDATE SET DELETE CREATE "
        "TABLE DROP ALTER INDEX VIEW PRIMARY KEY FOREIGN REFERENCES UNIQUE DEFAULT "
        "CHECK CONSTRAINT CASCADE RESTRICT GRANT REVOKE COMMIT ROLLBACK BEGIN TRANSACTION "
        "DECLARE CURSOR FETCH OPEN CLOSE EXEC EXECUTE PROCEDURE FUNCTION RETURN RETURNS "
        "TRIGGER BEFORE AFTER EACH ROW FOR IF WHILE DO LOOP EXIT CONTINUE RAISE EXCEPTION "
        "TRY CATCH THROW PRINT");

    // Data types
    m_lexer->setKeywords(QsciLexerSQL::Datatypes,
        "INTEGER INT SMALLINT BIGINT DECIMAL NUMERIC REAL FLOAT DOUBLE PRECISION "
        "CHARACTER CHAR VARCHAR TEXT BOOLEAN BOOL DATE TIME TIMESTAMP DATETIME "
        "INTERVAL BLOB CLOB JSON JSONB UUID ARRAY SERIAL BIGSERIAL");

    // Functions
    m_lexer->setKeywords(QsciLexerSQL::Functions,
        "COUNT SUM AVG MIN MAX COALESCE NULLIF CAST EXTRACT DATE_PART NOW CURRENT_DATE "
        "CURRENT_TIME CURRENT_TIMESTAMP LENGTH LOWER UPPER TRIM SUBSTRING REPLACE "
        "CONCAT SPLIT_PART REGEXP_REPLACE TO_CHAR TO_DATE TO_NUMBER ABS CEIL FLOOR "
        "ROUND RANDOM SETSEED GENERATE_SERIES");

    setLexer(m_lexer);

    // Lexer colors (dark theme)
    m_lexer->setColor(QColor("#CDD6F4"), QsciLexerSQL::Default);
    m_lexer->setColor(QColor("#89B4FA"), QsciLexerSQL::Keyword);
    m_lexer->setColor(QColor("#A6E3A1"), QsciLexerSQL::DoubleQuotedString);
    m_lexer->setColor(QColor("#F9E2AF"), QsciLexerSQL::SingleQuotedString);
    m_lexer->setColor(QColor("#94E2D5"), QsciLexerSQL::Number);
    m_lexer->setColor(QColor("#6C7086"), QsciLexerSQL::Comment);
    m_lexer->setColor(QColor("#F38BA8"), QsciLexerSQL::Operator);
}

void SqlEditor::setupEditor() {
    // Appearance
    setCaretLineVisible(true);
    setCaretLineBackgroundColor(QColor("#313244"));
    setCaretForegroundColor(QColor("#CDD6F4"));

    // Selection
    setSelectionBackgroundColor(QColor("#45475A"));
    setSelectionForegroundColor(QColor("#CDD6F4"));

    // Indentation
    setIndentationsUseTabs(false);
    setTabWidth(4);
    setIndentationGuides(true);
    setAutoIndent(true);

    // Brace matching
    setBraceMatching(QsciScintilla::SloppyBraceMatch);
    setMatchedBraceBackgroundColor(QColor("#A6E3A1"));
    setMatchedBraceForegroundColor(QColor("#1E1E2E"));
    setUnmatchedBraceBackgroundColor(QColor("#F38BA8"));
    setUnmatchedBraceForegroundColor(QColor("#1E1E2E"));

    // Autocomplete
    setAutoCompletionSource(QsciScintilla::AcsAPIs);
    setAutoCompletionThreshold(2);
    setAutoCompletionCaseSensitivity(false);
    setAutoCompletionReplaceWord(true);
    setAutoCompletionShowSingle(true);

    // Edge
    setEdgeMode(QsciScintilla::EdgeLine);
    setEdgeColumn(80);
    setEdgeColor(QColor("#313244"));

    // Whitespace
    setWhitespaceVisibility(QsciScintilla::WsInvisible);

    // EOL
    setEolMode(QsciScintilla::EolUnix);

    // Font
    QFont font("JetBrains Mono", 12);
    font.setStyleHint(QFont::Monospace);
    setFont(font);
}

void SqlEditor::setupMargins() {
    // Line numbers
    setMarginLineNumbers(0, true);
    setMarginWidth(0, "99999");
    setMarginsFont(font());
    setMarginsBackgroundColor(QColor("#181825"));
    setMarginsForegroundColor(QColor("#6C7086"));

    // Folding margin
    setFolding(QsciScintilla::BoxedTreeFoldStyle);
    setFoldMarginColors(QColor("#181825"), QColor("#181825"));
}

void SqlEditor::setupShortcuts() {
    // Ctrl+Enter - Execute current statement
    auto* execShortcut = new QShortcut(QKeySequence(Qt::CTRL | Qt::Key_Return), this);
    connect(execShortcut, &QShortcut::activated, this, &SqlEditor::executeCurrentStatement);

    // Ctrl+Shift+Enter - Execute all
    auto* execAllShortcut = new QShortcut(QKeySequence(Qt::CTRL | Qt::SHIFT | Qt::Key_Return), this);
    connect(execAllShortcut, &QShortcut::activated, this, &SqlEditor::executeAllStatements);

    // F5 - Execute
    auto* f5Shortcut = new QShortcut(QKeySequence(Qt::Key_F5), this);
    connect(f5Shortcut, &QShortcut::activated, this, &SqlEditor::executeCurrentStatement);

    // Ctrl+F - Find
    auto* findShortcut = new QShortcut(QKeySequence::Find, this);
    connect(findShortcut, &QShortcut::activated, this, [this]() {
        // TODO: Show find dialog
    });
}

void SqlEditor::setSql(const QString& sql) {
    setText(sql);
}

QString SqlEditor::sql() const {
    return text();
}

QString SqlEditor::selectedText() const {
    return QsciScintilla::selectedText();
}

QString SqlEditor::currentStatement() const {
    QString currentSql = sql();
    int cursorPos = currentPosition();
    int line, col;
    lineIndexFromPosition(cursorPos, &line, &col);

    // Find statement boundaries
    QStringList statements = currentSql.split(';', Qt::SkipEmptyParts);
    int currentLine = line;

    // Simple logic: find the statement containing current cursor position
    int lineCount = 0;
    for (const auto& stmt : statements) {
        int stmtLines = stmt.count('\n') + 1;
        if (currentLine < lineCount + stmtLines) {
            return stmt.trimmed();
        }
        lineCount += stmtLines;
    }

    return statements.isEmpty() ? currentSql : statements.last().trimmed();
}

QString SqlEditor::allStatements() const {
    return sql();
}

void SqlEditor::executeCurrentStatement() {
    QString stmt = currentStatement();
    if (!stmt.isEmpty()) {
        emit executeRequested(stmt);
    }
}

void SqlEditor::executeAllStatements() {
    QString allSql = sql();
    if (!allSql.isEmpty()) {
        // Split by semicolons, respecting strings
        QStringList statements;
        // TODO: Implement proper statement splitting
        statements = allSql.split(';', Qt::SkipEmptyParts);

        QStringList trimmed;
        for (auto& stmt : statements) {
            QString t = stmt.trimmed();
            if (!t.isEmpty()) {
                trimmed.append(t);
            }
        }

        emit executeAllRequested(trimmed);
    }
}

void SqlEditor::executeSelected() {
    QString selected = selectedText();
    if (!selected.isEmpty()) {
        emit executeRequested(selected);
    }
}

void SqlEditor::formatSql() {
    // TODO: Implement SQL formatting
    // Could use external library or simple regex-based formatting
}

void SqlEditor::setTableNames(const QStringList& tables) {
    m_tableNames = tables;
    refreshAutocomplete();
}

void SqlEditor::setColumnNames(const QString& table, const QStringList& columns) {
    m_columnNames[table] = columns;
    refreshAutocomplete();
}

void SqlEditor::refreshAutocomplete() {
    m_apis->clear();

    // Add tables
    for (const auto& table : m_tableNames) {
        m_apis->add(table);
        m_apis->add(table.toLower());
        m_apis->add(table.toUpper());
    }

    // Add columns with table prefix
    for (auto it = m_columnNames.begin(); it != m_columnNames.end(); ++it) {
        for (const auto& col : it.value()) {
            m_apis->add(col);
            m_apis->add(QString("%1.%2").arg(it.key(), col));
        }
    }

    m_apis->prepare();
}

} // namespace tablepro
```

**Step 3: Commit SQL editor**

```bash
git add src/ui/editor/sql_editor.hpp src/ui/editor/sql_editor.cpp
git commit -m "feat: Add SqlEditor widget with QScintilla"
```

---

## Task 3: Query Panel

**Files:**
- Create: `src/ui/panels/query_panel.hpp`
- Create: `src/ui/panels/query_panel.cpp`

**Step 1: Create query_panel.hpp**

```cpp
#pragma once

#include <QWidget>
#include <QVBoxLayout>
#include <QSplitter>
#include <QLabel>
#include <QPushButton>
#include "../editor/sql_editor.hpp"
#include "../grid/data_grid.hpp"

namespace tablepro {

class QueryPanel : public QWidget {
    Q_OBJECT

public:
    explicit QueryPanel(QWidget* parent = nullptr);

    void setConnectionId(const QString& connectionId);
    QString connectionId() const { return m_connectionId; }

    void setSql(const QString& sql);
    QString sql() const;

    SqlEditor* editor() const { return m_editor; }
    DataGrid* resultsGrid() const { return m_resultsGrid; }

    void showResults(const QueryResult& result);
    void clearResults();

signals:
    void executeRequested(const QString& sql);
    void resultsLoaded(bool success);

private slots:
    void onExecuteClicked();
    void onExecuteAllClicked();
    void onEditorExecuteRequested(const QString& sql);

private:
    void setupUI();
    void setupConnections();
    void updateResultsInfo(const QueryResult& result);

    SqlEditor* m_editor;
    DataGrid* m_resultsGrid;
    QLabel* m_resultsInfo;
    QPushButton* m_executeButton;
    QPushButton* m_executeAllButton;

    QString m_connectionId;
    QueryResult m_lastResult;
};

} // namespace tablepro
```

**Step 2: Create query_panel.cpp**

```cpp
#include "query_panel.hpp"
#include "core/query_executor.hpp"
#include "core/connection_manager.hpp"
#include <QMessageBox>

namespace tablepro {

QueryPanel::QueryPanel(QWidget* parent)
    : QWidget(parent)
    , m_editor(new SqlEditor(this))
    , m_resultsGrid(new DataGrid(this))
    , m_resultsInfo(new QLabel(this))
    , m_executeButton(new QPushButton(tr("Execute (F5)"), this))
    , m_executeAllButton(new QPushButton(tr("Execute All"), this))
{
    setupUI();
    setupConnections();
}

void QueryPanel::setupUI() {
    auto* layout = new QVBoxLayout(this);
    layout->setContentsMargins(0, 0, 0, 0);
    layout->setSpacing(0);

    // Toolbar
    auto* toolbar = new QWidget(this);
    auto* toolbarLayout = new QHBoxLayout(toolbar);
    toolbarLayout->setContentsMargins(8, 4, 8, 4);

    toolbarLayout->addWidget(m_executeButton);
    toolbarLayout->addWidget(m_executeAllButton);
    toolbarLayout->addStretch();
    toolbarLayout->addWidget(m_resultsInfo);

    layout->addWidget(toolbar);

    // Splitter for editor and results
    auto* splitter = new QSplitter(Qt::Vertical, this);
    splitter->addWidget(m_editor);
    splitter->addWidget(m_resultsGrid);
    splitter->setSizes({300, 300});

    layout->addWidget(splitter);

    // Style
    m_executeButton->setStyleSheet(R"(
        QPushButton {
            background-color: #89B4FA;
            color: #1E1E2E;
            border: none;
            padding: 6px 16px;
            border-radius: 4px;
            font-weight: bold;
        }
        QPushButton:hover {
            background-color: #B4BEFE;
        }
        QPushButton:pressed {
            background-color: #74C7EC;
        }
    )");

    m_resultsInfo->setStyleSheet("color: #A6ADC8; font-size: 12px;");
}

void QueryPanel::setupConnections() {
    connect(m_executeButton, &QPushButton::clicked, this, &QueryPanel::onExecuteClicked);
    connect(m_executeAllButton, &QPushButton::clicked, this, &QueryPanel::onExecuteAllClicked);
    connect(m_editor, &SqlEditor::executeRequested, this, &QueryPanel::onEditorExecuteRequested);
}

void QueryPanel::setConnectionId(const QString& connectionId) {
    m_connectionId = connectionId;
}

void QueryPanel::setSql(const QString& sql) {
    m_editor->setSql(sql);
}

QString QueryPanel::sql() const {
    return m_editor->sql();
}

void QueryPanel::showResults(const QueryResult& result) {
    m_lastResult = result;

    if (result.success) {
        m_resultsGrid->setQueryResult(result);
        updateResultsInfo(result);
    } else {
        m_resultsInfo->setText(QString("Error: %1").arg(result.error));
        m_resultsInfo->setStyleSheet("color: #F38BA8; font-size: 12px;");
    }

    emit resultsLoaded(result.success);
}

void QueryPanel::clearResults() {
    m_resultsGrid->setQueryResult(QueryResult());
    m_resultsInfo->clear();
}

void QueryPanel::updateResultsInfo(const QueryResult& result) {
    QString info = QString("%1 row(s) returned in %2 ms")
        .arg(result.rows.size())
        .arg(result.executionTimeMs);

    if (result.rowsAffected > 0 && result.rows.isEmpty()) {
        info = QString("%1 row(s) affected in %2 ms")
            .arg(result.rowsAffected)
            .arg(result.executionTimeMs);
    }

    m_resultsInfo->setText(info);
    m_resultsInfo->setStyleSheet("color: #A6ADC8; font-size: 12px;");
}

void QueryPanel::onExecuteClicked() {
    QString sql = m_editor->currentStatement();
    if (!sql.isEmpty()) {
        emit executeRequested(sql);
    }
}

void QueryPanel::onExecuteAllClicked() {
    QString sql = m_editor->sql();
    if (!sql.isEmpty()) {
        emit executeRequested(sql);
    }
}

void QueryPanel::onEditorExecuteRequested(const QString& sql) {
    emit executeRequested(sql);
}

} // namespace tablepro
```

**Step 3: Commit query panel**

```bash
git add src/ui/panels/query_panel.hpp src/ui/panels/query_panel.cpp
git commit -m "feat: Add QueryPanel combining editor and results"
```

---

## Task 4: Update CMakeLists and Verify

**Files:**
- Modify: `CMakeLists.txt`

**Step 1: Add sources**

```cmake
set(TABLEPRO_SOURCES
    # ... existing ...
    src/ui/editor/sql_editor.cpp
    src/ui/panels/query_panel.cpp
)
```

**Step 2: Build**

```bash
cmake --build build/debug -j$(nproc)
```

**Step 3: Commit**

```bash
git add CMakeLists.txt
git commit -m "build: Add SQL editor sources"
```

---

## Acceptance Criteria

- [ ] SqlEditor with syntax highlighting
- [ ] Line numbers displayed
- [ ] Execute current/all statements
- [ ] Autocomplete for tables/columns
- [ ] QueryPanel combines editor + results
- [ ] Results info shown
- [ ] Keyboard shortcuts work

---

**Phase 6 Complete.** MVP Ready! Next: Phase 7 - Export/Import