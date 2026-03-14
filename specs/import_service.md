# Import Service Algorithms (Go)

## 1. Orchestration
```go
type ImportService struct {
    ctx context.Context
}

func (is *ImportService) ImportFile(
    connectionID uuid.UUID,
    filePath string,
    format string,
    encoding string,
) error {
    driver := databaseManager.GetDriver(connectionID)

    // 1. Decompress if .gz
    actualPath, cleanup, err := decompressIfNeeded(filePath)
    if cleanup != nil { defer cleanup() }

    // 2. Open streaming parser
    parser := NewSQLFileParser()
    stmtChan, errChan := parser.ParseFile(actualPath, encoding)

    // 3. Execute in transaction
    driver.BeginTransaction()
    driver.Execute(driver.DialectInfo().DisableFKChecks)

    processed := 0
    for stmt := range stmtChan {
        if err := driver.Execute(stmt.Text); err != nil {
            driver.RollbackTransaction()
            runtime.EventsEmit(is.ctx, "import:error", ImportError{
                Line: stmt.LineNumber, Message: err.Error(), Processed: processed,
            })
            return err
        }
        processed++
        if processed%100 == 0 {
            runtime.EventsEmit(is.ctx, "import:progress", ImportProgress{
                Processed: processed, Status: fmt.Sprintf("Executing statement %d", processed),
            })
        }
    }

    driver.Execute(driver.DialectInfo().EnableFKChecks)
    driver.CommitTransaction()
    runtime.EventsEmit(is.ctx, "import:complete", processed)
    return nil
}
```

## 2. Gzip Decompression
```go
func decompressIfNeeded(path string) (string, func(), error) {
    if !strings.HasSuffix(path, ".gz") {
        return path, nil, nil
    }

    // Use Go's compress/gzip — pure Go, no subprocess needed
    gzFile, _ := os.Open(path)
    defer gzFile.Close()
    gzReader, _ := gzip.NewReader(gzFile)
    defer gzReader.Close()

    tmpFile, _ := os.CreateTemp("", "tablepro-import-*.sql")
    io.Copy(tmpFile, gzReader)
    tmpFile.Close()

    cleanup := func() { os.Remove(tmpFile.Name()) }
    return tmpFile.Name(), cleanup, nil
}
```
> **Advantage over Swift**: Go's `compress/gzip` is pure Go — no subprocess (`gunzip`) needed. Works on all platforms.

## 3. Streaming SQL Parser (Go)
```go
type SQLFileParser struct{}

type Statement struct {
    Text       string
    LineNumber int
}

func (p *SQLFileParser) ParseFile(path string, encoding string) (<-chan Statement, <-chan error) {
    stmtChan := make(chan Statement, 100) // buffered channel
    errChan := make(chan error, 1)

    go func() {
        defer close(stmtChan)
        defer close(errChan)

        file, err := os.Open(path)
        if err != nil { errChan <- err; return }
        defer file.Close()

        reader := bufio.NewReaderSize(file, 64*1024) // 64KB buffer
        var current strings.Builder
        state := stateNormal
        line := 1
        stmtStartLine := 1
        hasContent := false

        for {
            r, _, err := reader.ReadRune()
            if err == io.EOF { break }
            if err != nil { errChan <- err; return }

            if r == '\n' { line++ }

            switch state {
            case stateNormal:
                // Handle --, /* */, ', ", `, ;
                // Same FSM as Swift's SQLFileParser
            case stateSingleLineComment:
                if r == '\n' { state = stateNormal }
            case stateMultiLineComment:
                // Track */ to exit
            case stateSingleQuote, stateDoubleQuote, stateBacktick:
                // Track escape sequences and closing quotes
            }

            if r == ';' && state == stateNormal && hasContent {
                text := strings.TrimSpace(current.String())
                stmtChan <- Statement{Text: text, LineNumber: stmtStartLine}
                current.Reset()
                hasContent = false
            }
        }

        // Emit final statement if any
        if hasContent {
            stmtChan <- Statement{Text: strings.TrimSpace(current.String()), LineNumber: stmtStartLine}
        }
    }()

    return stmtChan, errChan
}
```

## 4. Progress Throttling
- Go emits progress every 100 statements
- React uses `useThrottle(progress, 66)` to cap UI updates at ~15fps
- Import dialog shows: progress bar, processed count, current status

## 5. Cancellation
- React calls `ImportService.CancelImport()`
- Go closes a `context.Done()` channel
- Parser goroutine checks `select { case <-ctx.Done(): return }` in its loop
- Driver rolls back the transaction
