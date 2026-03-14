# Integration Tests

This directory contains integration tests for TablePro's query execution functionality.

## Backend Integration Tests

Location: `internal/query/integration_test.go`

### Requirements

- Docker running (for testcontainers) OR
- PostgreSQL and MySQL databases accessible

### Environment Variables

```bash
# Required to run integration tests
export INTEGRATION_TEST=1

# PostgreSQL configuration
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=postgres
export POSTGRES_DB=postgres

# MySQL configuration
export MYSQL_HOST=localhost
export MYSQL_PORT=3306
export MYSQL_USER=root
export MYSQL_PASSWORD=root
export MYSQL_DB=mysql
```

### Running Tests

```bash
# Run all integration tests
go test -tags=integration ./internal/query/...

# Run with verbose output
go test -tags=integration -v ./internal/query/...

# Run specific test
go test -tags=integration -run TestIntegration_QueryExecution_PostgreSQL ./internal/query/...

# Run with coverage
go test -tags=integration -cover ./internal/query/...
```

### Test Coverage

- **11.1** Query Execution with PostgreSQL
  - Simple SELECT queries
  - ResultSet structure verification
  - Column type mapping
  - Data formatting

- **11.2** Query Execution with MySQL
  - Simple SELECT queries
  - ResultSet structure verification
  - Column type mapping
  - Data formatting

- **11.3** Query Cancellation
  - Cancel long-running queries (pg_sleep)
  - Verify resources cleaned up
  - CancelConnection functionality

- **11.4** Pagination with Large Result Sets
  - Generate 10k+ rows
  - Verify LIMIT/OFFSET applied correctly
  - Page navigation
  - Performance (< 1s per page)

- **11.5** NULL Value Handling
  - NULL values in various column types
  - NULL → JSON null conversion
  - ResultSet JSON serialization

- **11.6** History Tracking and Deduplication
  - Deduplicate identical queries
  - Whitespace normalization
  - LRU eviction at 50 queries
  - Per-connection isolation

## Frontend Integration Tests

Location: `frontend/src/components/QueryEditor.integration.test.tsx`

### Requirements

- Node.js 18+
- npm dependencies installed

### Running Tests

```bash
cd frontend

# Run integration tests
npm run test:integration

# Run with Vitest UI
npx vitest --ui

# Run specific test
npx vitest QueryEditor.integration.test.tsx
```

### Test Coverage

- **11.7** Keyboard Shortcuts
  - Ctrl+Enter: Execute query
  - Shift+Alt+F: Format query
  - Ctrl+T: New tab

- **11.8** Autocomplete with Schema Metadata
  - Table name suggestions
  - Column name suggestions
  - View name suggestions
  - Schema updates

## Playwright E2E Tests

Location: `tests/e2e/query-execution.spec.ts`

### Requirements

- Playwright installed: `npx playwright install`
- TablePro application running in dev mode
- PostgreSQL database available

### Running Tests

```bash
# Install Playwright browsers
npx playwright install

# Run all E2E tests
npx playwright test

# Run specific test file
npx playwright test query-execution

# Run with visible browser (headed)
npx playwright test --headed

# Run in debug mode
npx playwright test --debug

# Run specific browser
npx playwright test --project=chromium
```

### Environment Variables

```bash
# Application URL (default: http://localhost:34115)
export E2E_BASE_URL=http://localhost:34115

# Add delay between actions for debugging (in ms)
export E2E_SLOW_MO=100
```

### Test Coverage

End-to-end tests covering the complete query execution workflow:
- Connection setup
- Basic query execution
- Query cancellation
- Pagination
- NULL value handling
- History tracking
- Keyboard shortcuts
- Autocomplete

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  backend-integration:
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432
      mysql:
        image: mysql:8
        env:
          MYSQL_ROOT_PASSWORD: root
        ports:
          - 3306:3306
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      
      - name: Run backend integration tests
        env:
          INTEGRATION_TEST: 1
          POSTGRES_HOST: localhost
          MYSQL_HOST: localhost
        run: go test -tags=integration ./internal/query/...
  
  frontend-integration:
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
      
      - name: Install dependencies
        working-directory: frontend
        run: npm ci
      
      - name: Run frontend integration tests
        working-directory: frontend
        run: npm run test:integration
  
  e2e-tests:
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Install Playwright
        run: npx playwright install --with-deps
      
      - name: Start application
        run: wails dev &
      
      - name: Run E2E tests
        run: npx playwright test
```

## Troubleshooting

### Connection Refused Errors

If tests fail with connection errors:
1. Verify Docker is running
2. Check database containers are healthy
3. Ensure ports are not in use by other services

### Test Timeout

If tests timeout:
1. Increase timeout in test configuration
2. Check database performance
3. Reduce data set size

### Playwright Tests Fail

If E2E tests fail:
1. Verify application is running on correct port
2. Check `E2E_BASE_URL` environment variable
3. Run with `--headed` to see what's happening
4. Check browser console for errors

## Test Data Cleanup

All integration tests use temporary tables that are automatically dropped:
- PostgreSQL: `CREATE TEMP TABLE`
- MySQL: `CREATE TEMPORARY TABLE`

Tests are idempotent and can be run multiple times safely.
