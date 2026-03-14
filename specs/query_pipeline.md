# Query Execution Pipeline (Go Backend)

## Overview
Query execution flows from React → Go → Database Driver → Go → React. Go handles all SQL parsing, execution, and result transformation.

## Execution Flow
```
React                          Go Backend                     Database
─────                          ──────────                     ────────
Cmd+R pressed
  → QueryManager.Execute()  → Parse & validate SQL
                               Check SafeMode
                               Start timer
                               driver.Execute(sql)  ────────→  Run query
                                                    ←────────  Return rows
                               Transform rows to [][]any
                               Stop timer
                               Return QueryResult   ←───────
  ← Update AG Grid with data
  ← Update status bar
```

## QueryResult Structure
```go
type QueryResult struct {
    Columns       []ColumnInfo `json:"columns"`
    Rows          [][]any      `json:"rows"`
    RowsAffected  int64        `json:"rowsAffected"`
    ExecutionTime float64      `json:"executionTime"` // seconds
    ErrorMessage  string       `json:"errorMessage"`
    IsSelect      bool         `json:"isSelect"`
}

type ColumnInfo struct {
    Name     string `json:"name"`
    Type     string `json:"type"`
    Nullable bool   `json:"nullable"`
}
```

## Pagination
```go
func (qm *QueryManager) ExecuteWithPagination(
    connectionID uuid.UUID,
    tabID uuid.UUID,
    baseQuery string,
    offset int,
    limit int,
    orderBy string,
    orderDir string,
) (*QueryResult, error) {
    dialect := qm.getDialect(connectionID)
    paginatedQuery := dialect.WrapWithPagination(baseQuery, offset, limit, orderBy, orderDir)
    countQuery := dialect.WrapWithCount(baseQuery)

    // Execute both in parallel using goroutines
    var wg sync.WaitGroup
    var result *QueryResult
    var totalCount int64

    wg.Add(2)
    go func() { defer wg.Done(); result, _ = qm.execute(connectionID, paginatedQuery) }()
    go func() { defer wg.Done(); totalCount, _ = qm.executeCount(connectionID, countQuery) }()
    wg.Wait()

    result.TotalRowCount = totalCount
    return result, nil
}
```

## Dialect-Specific Pagination
```go
// PostgreSQL / MySQL / SQLite / DuckDB
SELECT * FROM "table" ORDER BY "col" LIMIT 500 OFFSET 0;

// SQL Server
SELECT * FROM [table] ORDER BY [col] OFFSET 0 ROWS FETCH NEXT 500 ROWS ONLY;

// Oracle
SELECT * FROM "table" ORDER BY "col" FETCH FIRST 500 ROWS ONLY;
```

## Concurrent Query Execution
- Each tab executes queries on its own goroutine
- Go `context.Context` with timeout for cancellation
- User clicks Cancel → `context.CancelFunc()` called → driver returns immediately
- Emit `runtime.EventsEmit(ctx, "query:cancelled", tabID)`

## EXPLAIN Support
```go
func (qm *QueryManager) Explain(connectionID uuid.UUID, sql string) (*QueryResult, error) {
    dialect := qm.getDialect(connectionID)
    explainSQL := dialect.WrapWithExplain(sql)
    // PostgreSQL: EXPLAIN ANALYZE {sql}
    // MySQL: EXPLAIN {sql}
    // SQLite: EXPLAIN QUERY PLAN {sql}
    return qm.execute(connectionID, explainSQL)
}
```

## Statement Splitting (Go)
```go
func SplitStatements(sql string) []Statement {
    // Finite state machine identical to Swift's SQLFileParser
    // States: normal, singleLineComment, multiLineComment, singleQuote, doubleQuote, backtick
    // Yields on unquoted ';'
}
```
