## 1. Dependencies & Setup

- [x] 1.1 Install @monaco-editor/react@^4.6.0
- [x] 1.2 Install @types/monaco dev dependency
- [x] 1.3 Run `npm install` to update package-lock.json
- [x] 1.4 Verify Monaco types compile without errors

## 2. QueryExecutor Service (Backend)

- [x] 2.1 Create `internal/query/executor.go` with QueryExecutor struct
- [x] 2.2 Implement Execute() method with context timeout
- [x] 2.3 Implement Cancel() method for query cancellation
- [x] 2.4 Implement multi-statement query execution
- [x] 2.5 Add query result streaming support
- [x] 2.6 Implement active query tracking per connection
- [x] 2.7 Add Wails bindings for Execute() and Cancel()
- [x] 2.8 Write unit tests for QueryExecutor

## 3. ResultSet Handling (Backend)

- [ ] 3.1 Create `internal/query/resultset.go` with ResultSet struct
- [ ] 3.2 Define ColumnInfo struct with type metadata
- [ ] 3.3 Implement column type mapping from database types
- [ ] 3.4 Add NULL value handling (nil → JSON null)
- [ ] 3.5 Implement data formatting (dates, booleans, numbers)
- [ ] 3.6 Add multiple result sets support
- [ ] 3.7 Write unit tests for ResultSet

## 4. Pagination Service (Backend)

- [x] 4.1 Create `internal/query/pagination.go` with PaginationService
- [x] 4.2 Implement LIMIT/OFFSET pagination for SQL databases
- [x] 4.3 Implement cursor-based pagination for NoSQL (MongoDB, Redis)
- [x] 4.4 Add page size configuration with max limit
- [x] 4.5 Implement total count estimation (exact for small, estimate for large)
- [x] 4.6 Write unit tests for PaginationService

## 5. Query History Tracking (Backend)

- [x] 5.1 Create `internal/query/history.go` with QueryHistory struct
- [x] 5.2 Implement in-memory history storage per connection
- [x] 5.3 Add query deduplication (whitespace-normalized)
- [x] 5.4 Implement LRU eviction (max 50 queries per connection)
- [x] 5.5 Add GetHistory() method for frontend retrieval
- [x] 5.6 Write unit tests for QueryHistory

## 6. QueryEditor Component (Frontend)

- [x] 6.1 Create `frontend/src/components/QueryEditor.tsx`
- [x] 6.2 Integrate Monaco Editor with SQL language mode
- [x] 6.3 Configure SQL syntax highlighting for PostgreSQL/MySQL dialects
- [x] 6.4 Implement autocomplete provider with schema metadata
- [x] 6.5 Add keyboard shortcuts (Ctrl+Enter to execute, Shift+Alt+F to format)
- [x] 6.6 Implement multi-tab support for multiple queries
- [x] 6.7 Add tab close button and focus management
- [x] 6.8 Write component tests for QueryEditor

## 7. ResultView Component (Frontend)

- [ ] 7.1 Create `frontend/src/components/ResultView.tsx`
- [ ] 7.2 Implement result table display with column headers
- [ ] 7.3 Add pagination controls (Previous, Next, page input)
- [ ] 7.4 Show current page info ("Page X of Y (Z rows)")
- [ ] 7.5 Implement Cancel button during query execution
- [ ] 7.6 Add loading indicator for streaming results
- [ ] 7.7 Display error messages with database error details
- [ ] 7.8 Write component tests for ResultView

## 8. Query Store (Frontend State)

- [x] 8.1 Create `frontend/src/stores/queryStore.ts` with Zustand
- [x] 8.2 Add state for active queries per connection
- [x] 8.3 Add state for query history entries
- [x] 8.4 Implement actions: executeQuery, cancelQuery, addToHistory
- [x] 8.5 Connect to Wails Events for query lifecycle events
- [x] 8.6 Write store tests for queryStore

## 9. Wails Event Integration

- [ ] 9.1 Emit `query:executing` event when query starts
- [ ] 9.2 Emit `query:completed` event with duration and row count
- [ ] 9.3 Emit `query:failed` event with error details
- [ ] 9.4 Emit `history:added` event when query added to history
- [x] 9.5 Frontend: Subscribe to events in queryStore
- [x] 9.6 Frontend: Clean up event listeners on unmount

## 10. History UI Panel

- [x] 10.1 Create `frontend/src/components/HistoryPanel.tsx`
- [x] 10.2 Display list of query history entries
- [x] 10.3 Add search input for filtering history
- [x] 10.4 Implement click-to-load query in editor
- [x] 10.5 Show timestamp and duration for each entry
- [x] 10.6 Add "Clear History" button per connection
- [x] 10.7 Write component tests for HistoryPanel

## 11. Integration Testing

- [x] 11.1 Test query execution with PostgreSQL driver
- [x] 11.2 Test query execution with MySQL driver
- [x] 11.3 Test query cancellation during long-running query
- [x] 11.4 Test pagination with large result sets (10k+ rows)
- [x] 11.5 Test NULL value handling in results
- [x] 11.6 Test history tracking and deduplication
- [x] 11.7 Test keyboard shortcuts functionality
- [x] 11.8 Test autocomplete with real schema metadata

## 12. Documentation & Polish

- [x] 12.1 Add JSDoc comments to frontend components
- [x] 12.2 Add Go doc comments to query package
- [x] 12.3 Update README.md with Phase 5 features
- [x] 12.4 Add error handling best practices to AGENTS.md
- [x] 12.5 Run `go test ./internal/query/...` - all tests pass
- [x] 12.6 Run `npm test` - all frontend tests pass
- [x] 12.7 Run `go vet ./...` - no issues
- [x] 12.8 Run `npm run lint` - no issues
