# Helpers & Gotchas (Go + React)

## Go-Specific Helpers

### 1. Connection URL Parser
```go
func ParseConnectionURL(rawURL string) (*ParsedConnection, error) {
    // Handle dual-@ SSH URLs: postgres+ssh://sshuser@bastion:22/dbuser:pass@10.0.0.1:5432/mydb
    // Go's url.Parse() breaks on dual @, so use custom regex:
    sshPattern := regexp.MustCompile(`^(\w+)\+ssh://([^@]+)@([^/]+)/(.+)$`)
    // Extract SSH part, then parse inner URL normally
}
```
> **Gotcha**: Go's `net/url.Parse()` cannot handle `scheme+ssh://user@host/otheruser@otherhost` — must use manual regex splitting (same issue as Swift).

### 2. Goroutine Leak Prevention
```go
// Always use context with timeout for database operations
ctx, cancel := context.WithTimeout(parentCtx, time.Duration(settings.QueryTimeout)*time.Second)
defer cancel()
result, err := driver.ExecuteWithContext(ctx, query)
```
> **Gotcha**: Forgetting `defer cancel()` causes goroutine leaks. Every DB call must have a bounded context.

### 3. JSON Serialization of `any` Types
```go
// Database drivers return []any for cell values
// Go's json.Marshal handles most types, but:
// - time.Time → format as ISO 8601 string
// - []byte → base64 encode (or hex for binary display)
// - nil → JSON null (AG Grid renders as "NULL" styled cell)
```

### 4. Large Query Truncation
```go
const MaxPersistableQuerySize = 500 * 1024 // 500KB
func truncateForPersistence(query string) string {
    if len(query) > MaxPersistableQuerySize {
        return "" // Don't persist oversized queries
    }
    return query
}
```

## React-Specific Helpers

### 1. Debounced Search
```typescript
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedValue(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);
  return debouncedValue;
}
```

### 2. Progress Throttling
```typescript
// Go emits progress events very rapidly during imports
// Throttle React re-renders to ~15fps
const throttledProgress = useThrottle(progress, 66); // 66ms = ~15fps
```

### 3. Clipboard Handling
```typescript
// Copy cell or row data to clipboard
async function copyToClipboard(text: string) {
  await navigator.clipboard.writeText(text);
  // Wails also provides runtime.ClipboardSetText() for cross-platform
}
```

### 4. Date Formatting
```typescript
// Display database timestamps in user's locale
function formatDate(value: string, locale: string): string {
  const date = new Date(value);
  return new Intl.DateTimeFormat(locale, {
    dateStyle: 'medium',
    timeStyle: 'medium',
  }).format(date);
}
```

## Architecture Gotchas

### WebView Memory
- AG Grid with 100K rows in DOM = ~200MB WebView memory
- **Must use** AG Grid's Server-Side Row Model for large datasets
- Keep result set transfers from Go → React under 50MB

### JSON Transfer Overhead
- Wails serializes Go structs to JSON for frontend consumption
- For 100K rows × 20 columns, JSON payload can be 50-100MB
- **Mitigation**: Paginate results (default 500 rows per page)
- **Mitigation**: Use compact column types (don't stringify everything)

### Concurrent Access in Go
- All database driver calls happen on goroutines
- `sync.RWMutex` on shared state (SchemaCache, ConnectionSessions)
- Never hold a lock while calling `runtime.EventsEmit()` (deadlock risk)
