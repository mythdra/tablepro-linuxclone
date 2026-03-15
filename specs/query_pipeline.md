# Query Execution Pipeline (C++20 + Qt)

## Overview
Query execution flows from Qt UI → C++ Core → Database Driver → C++ → Qt UI. C++ handles all SQL parsing, execution, and result transformation.

## Execution Flow
```
Qt UI                        C++ Core                       Database
─────                        ────────                       ────────
Cmd+R pressed
  → QueryManager::execute() → Parse & validate SQL
                               Check SafeMode
                               Start timer (QElapsedTimer)
                               driver->execute(sql)  ────────→  Run query
                                                    ←────────  Return rows
                               Transform to QVariantList
                               Stop timer
                               Return QueryResult  ←───────
  ← Model updated with data
  ← Status bar updated via signal
```

## QueryResult Structure

```cpp
// src/core/QueryResult.hpp
#pragma once

#include <QString>
#include <QVariantList>
#include <QElapsedTimer>

namespace tablepro {

struct ColumnInfo {
    QString name;
    QString type;
    bool nullable{true};

    QJsonObject toJson() const {
        return {{"name", name}, {"type", type}, {"nullable", nullable}};
    }
};

class QueryResult {
public:
    QList<ColumnInfo> columns;
    QVariantList rows;           // 2D array: [[cell, cell], [cell, cell]]
    QVariantList::size_type rowCount() const { return rows.size(); }
    qint64 rowsAffected{0};
    double executionTime{0.0};   // seconds
    QString errorMessage;
    bool isSelect{false};
    bool success{false};

    // For QVariant compatibility
    QJsonObject toJson() const;
    static QueryResult fromJson(const QJsonObject& json);
};

} // namespace tablepro

Q_DECLARE_METATYPE(tablepro::QueryResult)
```

## QueryManager Class

```cpp
// src/core/QueryManager.hpp
#pragma once

#include <QObject>
#include <QFuture>
#include <QFutureWatcher>
#include <QtConcurrent>
#include "QueryResult.hpp"
#include "DatabaseDriver.hpp"
#include "SqlGenerator.hpp"

namespace tablepro {

class QueryManager : public QObject {
    Q_OBJECT

public:
    explicit QueryManager(QObject* parent = nullptr);

    // Execute single query
    QFuture<QueryResult> execute(const QUuid& sessionId, const QString& sql);

    // Execute with pagination
    QFuture<QueryResult> executeWithPagination(
        const QUuid& sessionId,
        const QString& baseQuery,
        int offset,
        int limit,
        const QString& orderBy = {},
        const QString& orderDir = "ASC");

    // Execute EXPLAIN
    QFuture<QueryResult> explain(const QUuid& sessionId, const QString& sql);

    // Cancel running query
    void cancelQuery(const QUuid& sessionId);

    // Set driver (called when connection established)
    void setDriver(const QUuid& sessionId, DatabaseDriver* driver);

signals:
    void queryStarted(const QUuid& sessionId);
    void queryFinished(const QUuid& sessionId, const QueryResult& result);
    void queryError(const QUuid& sessionId, const QString& message);
    void queryCancelled(const QUuid& sessionId);

private:
    struct SessionData {
        DatabaseDriver* driver{nullptr};
        QFuture<QueryResult> currentFuture;
        std::atomic<bool> cancelled{false};
    };

    QMap<QUuid, SessionData> m_sessions;
    QHash<DatabaseType, std::unique_ptr<SqlDialect>> m_dialects;
    std::chrono::milliseconds m_timeout{30000};  // 30 second default

    QueryResult executeInternal(
        const QUuid& sessionId,
        const QString& sql,
        std::atomic<bool>& cancelled);
};

} // namespace tablepro
```

## Pagination Implementation

```cpp
// src/core/QueryManager.cpp
#include "QueryManager.hpp"
#include <QtConcurrent>

QFuture<QueryResult> QueryManager::executeWithPagination(
    const QUuid& sessionId,
    const QString& baseQuery,
    int offset,
    int limit,
    const QString& orderBy,
    const QString& orderDir)
{
    auto it = m_sessions.find(sessionId);
    if (it == m_sessions.end() || !it->driver) {
        return QtConcurrent::make_ready_future(QueryResult{
            .errorMessage = "No active connection",
            .success = false
        });
    }

    auto* driver = it->driver;
    auto type = driver->type();
    auto* dialect = m_dialects.value(type).get();

    // Build paginated query
    QString paginatedQuery = dialect->wrapWithPagination(
        baseQuery, offset, limit, orderBy, orderDir);

    // Build count query (runs in parallel)
    QString countQuery = dialect->wrapWithCount(baseQuery);

    // Execute both queries concurrently
    auto* watcher = new QFutureWatcher<std::pair<QueryResult, qint64>>(this);

    auto future = QtConcurrent::run([=]() {
        QueryResult result = driver->execute(paginatedQuery);
        qint64 totalCount = 0;

        if (result.success) {
            // Execute count in same connection if transactional
            auto countResult = driver->execute(countQuery);
            if (countResult.success && !countResult.rows.isEmpty()) {
                totalCount = countResult.rows.first().toList().first().toLongLong();
            }
        }

        return std::make_pair(result, totalCount);
    });

    watcher->setFuture(future);

    // Connect to finished signal
    connect(watcher, &QFutureWatcher<std::pair<QueryResult, qint64>>::finished,
            this, [=]() {
        auto result = watcher->result();
        result.first.rowsAffected = result.second;  // Store total in rowsAffected

        emit queryFinished(sessionId, result.first);
        watcher->deleteLater();
    });

    connect(watcher, &QFutureWatcher<std::pair<QueryResult, qint64>>::canceled,
            this, [=]() {
        emit queryCancelled(sessionId);
        watcher->deleteLater();
    });

    it->currentFuture = future;

    return future;
}
```

## Dialect-Specific Pagination

```cpp
// PostgreSQL / MySQL / SQLite / DuckDB
SELECT * FROM "table" ORDER BY "col" LIMIT 500 OFFSET 0;

// SQL Server
SELECT * FROM [table] ORDER BY [col]
OFFSET 0 ROWS FETCH NEXT 500 ROWS ONLY;

// Oracle
SELECT * FROM "table" ORDER BY "col" FETCH FIRST 500 ROWS ONLY;
```

## Concurrent Query Execution

```cpp
// Each tab executes queries on its own QtConcurrent thread
QFuture<QueryResult> future = QtConcurrent::run([=]() {
    return driver->execute(sql);
});

auto* watcher = new QFutureWatcher<QueryResult>(this);
connect(watcher, &QFutureWatcher<QueryResult>::finished,
        this, [=]() {
    emit queryResultReady(watcher->result());
    watcher->deleteLater();
});

// Cancellation via QFuture
void QueryManager::cancelQuery(const QUuid& sessionId) {
    auto it = m_sessions.find(sessionId);
    if (it != m_sessions.end()) {
        it->cancelled = true;
        it->currentFuture.cancel();
    }
}
```

## EXPLAIN Support

```cpp
QFuture<QueryResult> QueryManager::explain(
    const QUuid& sessionId,
    const QString& sql)
{
    auto it = m_sessions.find(sessionId);
    if (it == m_sessions.end() || !it->driver) {
        return QtConcurrent::make_ready_future(QueryResult{
            .errorMessage = "No active connection",
            .success = false
        });
    }

    auto* dialect = m_dialects.value(it->driver->type()).get();
    QString explainSql = dialect->wrapWithExplain(sql);

    // PostgreSQL: EXPLAIN ANALYZE {sql}
    // MySQL: EXPLAIN {sql}
    // SQLite: EXPLAIN QUERY PLAN {sql}

    return execute(sessionId, explainSql);
}
```

## Statement Splitting (C++)

```cpp
// src/core/StatementSplitter.hpp
#pragma once

#include <QString>
#include <vector>

namespace tablepro {

struct Statement {
    QString sql;
    int startPosition{0};
    int endPosition{0};
    bool isEmpty() const { return sql.trimmed().isEmpty(); }
};

class StatementSplitter {
public:
    static std::vector<Statement> split(const QString& sql);

private:
    enum class State {
        Normal,
        SingleLineComment,
        MultiLineComment,
        SingleQuote,
        DoubleQuote,
        Backtick
    };
};

} // namespace tablepro

// src/core/StatementSplitter.cpp
std::vector<Statement> StatementSplitter::split(const QString& sql) {
    std::vector<Statement> statements;
    State state = State::Normal;
    int start = 0;

    for (int i = 0; i < sql.length(); ++i) {
        QChar c = sql[i];

        switch (state) {
            case State::Normal:
                if (c == '-' && i + 1 < sql.length() && sql[i + 1] == '-') {
                    state = State::SingleLineComment;
                } else if (c == '/' && i + 1 < sql.length() && sql[i + 1] == '*') {
                    state = State::MultiLineComment;
                    ++i;
                } else if (c == '\'') {
                    state = State::SingleQuote;
                } else if (c == '"') {
                    state = State::DoubleQuote;
                } else if (c == '`') {
                    state = State::Backtick;
                } else if (c == ';') {
                    // Statement boundary
                    statements.push_back({
                        .sql = sql.mid(start, i - start),
                        .startPosition = start,
                        .endPosition = i
                    });
                    start = i + 1;
                }
                break;

            case State::SingleLineComment:
                if (c == '\n') {
                    state = State::Normal;
                }
                break;

            case State::MultiLineComment:
                if (c == '*' && i + 1 < sql.length() && sql[i + 1] == '/') {
                    state = State::Normal;
                    ++i;
                }
                break;

            case State::SingleQuote:
            case State::DoubleQuote:
            case State::Backtick:
                // Check for escape or end quote
                if (c == '\\' || (sql[i-1] != '\\' && c == sql[start])) {
                    if (c == sql[start]) {
                        state = State::Normal;
                    }
                }
                break;
        }
    }

    // Add remaining statement
    if (start < sql.length()) {
        statements.push_back({
            .sql = sql.mid(start),
            .startPosition = start,
            .endPosition = sql.length()
        });
    }

    return statements;
}
```

## Safe Mode Enforcement

```cpp
// src/core/SafeModeChecker.hpp
#pragma once

#include "QueryResult.hpp"

namespace tablepro {

class SafeModeChecker {
public:
    enum class Level {
        Off,
        ReadOnly,           // Block UPDATE/DELETE entirely
        RequireWhere        // Require WHERE clause for UPDATE/DELETE
    };

    static QueryResult check(const QString& sql, Level level) {
        if (level == Level::Off) {
            return {};  // No check, return success
        }

        QString upper = sql.trimmed().toUpper();

        if (level == Level::ReadOnly) {
            if (upper.startsWith("UPDATE") || upper.startsWith("DELETE") ||
                upper.startsWith("INSERT") || upper.startsWith("DROP") ||
                upper.startsWith("TRUNCATE")) {
                return {
                    .errorMessage = "Safe Mode: Destructive queries are disabled",
                    .success = false
                };
            }
        }

        if (level == Level::RequireWhere) {
            if ((upper.startsWith("UPDATE") || upper.startsWith("DELETE")) &&
                !upper.contains("WHERE")) {
                return {
                    .errorMessage = "Safe Mode: WHERE clause required for UPDATE/DELETE",
                    .success = false
                };
            }
        }

        return {};  // Passed check
    }
};

} // namespace tablepro
```
