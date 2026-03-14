## Context

**Current State:**
- Phase 4 complete: 8 database drivers implement `DatabaseDriver` interface với `Execute()` method
- App struct đã có `startup()`/`shutdown()` với logging và events system
- Frontend có ConnectionForm và ConnectionList, chưa có QueryEditor
- Monaco Editor chưa được install
- Query history chưa có implementation

**Constraints:**
- Phải hoạt động với tất cả 8 drivers (PostgreSQL, MySQL, SQLite, DuckDB, MSSQL, ClickHouse, MongoDB, Redis)
- Support large result sets (millions of rows) → cần pagination và streaming
- Context timeout cho tất cả DB operations (theo AGENTS.md conventions)
- Không suppress type errors (`as any`, `@ts-ignore` forbidden)

**Stakeholders:**
- Users: Cần viết và execute SQL queries với feedback nhanh
- Developers: Cần code pattern nhất quán, dễ test, dễ maintain

## Goals / Non-Goals

**Goals:**
- Monaco Editor với SQL syntax highlighting và autocomplete
- Query execution với timeout (configurable, default 30s) và cancellation
- Result pagination (server-side) để handle large datasets
- Result formatting đúng types (NULL, dates, numbers, booleans)
- In-memory query history tracking (last 50 queries per connection)
- Events system integration (`query:executing`, `query:completed`, `query:failed`)

**Non-Goals:**
- Persistent query history (SQLite FTS5) → Phase 11
- Query result export → Phase 9
- Inline editing và change tracking → Phase 7
- Tab persistence → Phase 8
- Query builder visual → Không thuộc scope Phase 5

## Decisions

### Decision 1: Monaco Editor vs CodeMirror 6

**Choice:** Monaco Editor (`@monaco-editor/react`)

**Rationale:**
- Monaco = editor của VS Code → UX quen thuộc cho developers
- Built-in SQL syntax highlighting với multiple dialects
- Better autocomplete/intellisense support
- Performance tốt hơn với large files
- Consistent với spec-driven requirements

**Alternatives considered:**
- CodeMirror 6: Nhẹ hơn nhưng autocomplete phức tạp hơn
- Ace Editor: Ít maintain hơn, không có TypeScript types tốt

### Decision 2: QueryExecutor Architecture

**Choice:** Service pattern với struct-based design

```go
type QueryExecutor struct {
    mu         sync.RWMutex
    sessions   map[uuid.UUID]*QuerySession  // connection ID → active queries
    defaultTimeout time.Duration
}

type QuerySession struct {
    ConnectionID uuid.UUID
    ActiveQuery  *ActiveQuery
    History      []QueryHistoryEntry
}

type ActiveQuery struct {
    ID        uuid.UUID
    Query     string
    Context   context.Context
    Cancel    context.CancelFunc
    StartedAt time.Time
}
```

**Rationale:**
- Match với ConnectionManager pattern (đã có trong Phase 3)
- Thread-safe với sync.RWMutex
- Per-connection query tracking → dễ implement cancellation
- Context-based timeout → consistent với Go conventions

### Decision 3: Pagination Strategy

**Choice:** Server-side LIMIT/OFFSET cho SQL, cursor-based cho NoSQL

**SQL (PostgreSQL, MySQL, etc.):**
```go
func (e *QueryExecutor) ExecutePaginated(ctx context.Context, connID uuid.UUID, query string, page int, pageSize int) (*ResultSet, error) {
    // Append LIMIT/OFFSET
    paginatedQuery := fmt.Sprintf("%s LIMIT %d OFFSET %d", query, pageSize, (page-1)*pageSize)
    // Execute and return
}
```

**NoSQL (MongoDB, Redis):**
- MongoDB: Use `.skip().limit()` với cursor
- Redis: Use SCAN command với COUNT parameter

**Rationale:**
- Server-side pagination → không load toàn bộ result vào memory
- LIMIT/OFFSET simple và supported bởi tất cả SQL databases
- Cursor-based better cho NoSQL và infinite scrolling

### Decision 4: ResultSet Data Structure

**Choice:** Column-oriented storage với metadata

```go
type ResultSet struct {
    Columns      []ColumnInfo    // name, type, nullable
    Rows         [][]interface{} // column-oriented data
    RowCount     int64           // total rows (for pagination)
    QueryTime    time.Duration
    Statement    string          // executed statement
}

type ColumnInfo struct {
    Name     string
    Type     string      // database type name
    DataType DataType    // normalized type enum
    Nullable bool
}
```

**Rationale:**
- Column-oriented → efficient for AG Grid server-side row model
- DataType enum → easy frontend formatting
- Include QueryTime → debugging và performance monitoring

### Decision 5: Query Cancellation

**Choice:** Context-based cancellation với user-triggered Cancel

```go
// Go: Execute returns cancellation token
func (e *QueryExecutor) Execute(ctx context.Context, connID uuid.UUID, query string) (uuid.UUID, error)

// Go: Cancel method
func (e *QueryExecutor) Cancel(queryID uuid.UUID) error

// Frontend: Track query ID, show Cancel button
const [queryId, setQueryId] = useState<string | null>(null)
{queryId && <Button onClick={() => QueryExecutor.Cancel(queryId)}>Cancel</Button>}
```

**Rationale:**
- Context cancellation → clean resource cleanup
- Query ID tracking → users can cancel long-running queries
- Consistent với Go best practices

### Decision 6: In-Memory History Storage

**Choice:** LRU cache per connection với configurable limit

```go
type QueryHistory struct {
    mu       sync.RWMutex
    entries  map[uuid.UUID][]HistoryEntry  // connection ID → entries
    maxPerConnection int
}

func (h *QueryHistory) Add(connID uuid.UUID, entry HistoryEntry) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    entries := h.entries[connID]
    entries = append(entries, entry)
    
    // LRU: Remove oldest if exceeds limit
    if len(entries) > h.maxPerConnection {
        entries = entries[1:]
    }
    h.entries[connID] = entries
}
```

**Rationale:**
- Per-connection → dễ filter và search
- LRU eviction → không lo về memory leak
- In-memory → fast access, Phase 11 sẽ persist sang SQLite

## Risks / Trade-offs

| Risk | Impact | Mitigation |
|------|--------|------------|
| Monaco Editor bundle size (~2MB) | Medium | Load editor lazily, code splitting |
| Server-side pagination với complex queries | Medium | Document limitation, recommend WHERE clauses |
| Context cancellation không stop executing query ở driver level | High | Drivers phải respect context timeout (Phase 4 đã implement) |
| Memory usage với large result sets | Medium | Stream results, AG Grid virtual scrolling |
| In-memory history lost on restart | Low | Acceptable for Phase 5, Phase 11 adds persistence |
| MongoDB/Redis pagination khác biệt | Medium | Abstract behind PaginationService interface |

## Migration Plan

**Phase 5 Implementation Order:**
1. Install dependencies (`@monaco-editor/react`, `@types/monaco`)
2. Create `internal/query/` package (QueryExecutor, ResultSet, PaginationService)
3. Create `QueryEditor.tsx` component với Monaco integration
4. Create `ResultView.tsx` component với pagination controls
5. Wire up Wails bindings
6. Add keyboard shortcuts và events listeners
7. Test với tất cả 8 database drivers

**Rollback Strategy:**
- Phase 5 không modify existing code (Phase 1-4)
- Chỉ thêm mới → rollback = delete new files, remove dependencies

## Open Questions

1. **Autocomplete implementation:** Should we use Monaco's built-in SQL completion hoặc custom provider với schema metadata từ drivers?
   - leaning toward: Custom provider với GetSchema() từ Phase 4

2. **Query result streaming:** Should we implement chunked streaming cho very large results (>10k rows)?
   - leaning toward: Defer to Phase 7 (Data Grid), start với simple pagination

3. **History deduplication:** What defines "duplicate" query? Exact match hoặc normalized (whitespace-insensitive)?
   - leaning toward: Normalized match (trim whitespace, uppercase keywords)
