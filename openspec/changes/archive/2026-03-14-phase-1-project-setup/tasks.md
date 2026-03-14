# Phase 1: Project Setup Tasks

Implementation checklist for Phase 1 - Project Setup & Infrastructure (25 tasks)

---

## 1. Go Module Initialization

- [x] 1.1 Initialize Go module with `go mod init github.com/tablepro/tablepro`
- [x] 1.2 Create directory structure: cmd/, internal/, frontend/
- [x] 1.3 Create internal package directories: driver/, connection/, session/, query/, change/, export/, import/, history/, settings/, tab/, ssh/, license/
- [x] 1.4 Add go.mod dependencies: wails/v2, pgx/v5, go-sql-driver/mysql
- [x] 1.5 Create .gitignore for Go projects (exe, dll, so, dylib, test, out, vendor/, build/)
- [x] 1.6 Set up Go workspace configuration (VS Code settings, gopls)

---

## 2. Wails Project Setup

- [x] 2.1 Install Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- [x] 2.2 Initialize Wails project with React template: `wails init -n tablepro -t react`
- [x] 2.3 Configure wails.json with app metadata (name: TablePro, version, author)
- [x] 2.4 Set up app icon resources (1024x1024, 512x512, 256x256, 128x128, 64x64, 32x32, 16x16)
- [x] 2.5 Generate platform-specific icons: .icns (macOS), .ico (Windows), .png (Linux)
- [x] 2.6 Configure build options for darwin, windows, linux platforms

---

## 3. Frontend Setup

- [x] 3.1 Install npm dependencies in frontend/ directory
- [x] 3.2 Verify required packages: react@^18.2.0, react-dom@^18.2.0, @ag-grid-community/react@^31.0.0, @monaco-editor/react@^4.6.0, zustand@^4.5.0, tailwindcss@^3.4.0, lucide-react@^0.350.0
- [x] 3.3 Configure TypeScript (tsconfig.json) with strict mode, target ES2020, module ESNext
- [x] 3.4 Set up Tailwind CSS configuration with custom colors (primary: #3B82F6, success: #10B981, warning: #F59E0B, danger: #EF4444)
- [x] 3.5 Configure Vitest test runner with jsdom environment
- [x] 3.6 Set up ESLint with TypeScript and React plugins
- [x] 3.7 Configure Prettier for code formatting

---

## 4. Development Environment

- [x] 4.1 Create Makefile with targets: dev, build, build-all, test, lint, clean
- [x] 4.2 Create VS Code launch configuration for debugging Wails app (Go debugger)
- [x] 4.3 Create VS Code launch configuration for frontend tests (Node.js debugger)
- [x] 4.4 Configure Go language server (gopls) with gofumpt and static analysis
- [x] 4.5 Create .env.example template with WAILS_DEBUG, LOG_LEVEL, TEST_POSTGRES_URL, TEST_MYSQL_URL
- [x] 4.6 Document setup process in README.md (Quick Start, Requirements, Commands)

---

## 5. CI/CD Setup

- [x] 5.1 Create .github/workflows/go-tests.yml (runs on push/PR, tests with Go 1.21)
- [x] 5.2 Create .github/workflows/frontend-tests.yml (runs on push/PR, tests with Node 18)
- [x] 5.3 Create .github/workflows/build.yml (runs on release, builds all platforms)
- [x] 5.4 Configure code coverage reporting with Codecov integration
- [x] 5.5 Add linting checks to CI (go vet, golangci-lint, npm lint)
- [x] 5.6 Configure npm caching in GitHub Actions for faster builds

---

## Verification Checklist

Run these commands to verify Phase 1 completion:

```bash
# Go module
go mod tidy
go list ./...

# Wails
wails dev

# Frontend
cd frontend && npm install
cd frontend && npm run dev

# Tests
make test

# Linting
make lint
```

---

## Acceptance Criteria

- [x] `wails dev` starts without errors
- [x] `go test ./...` passes (no tests yet, but command works)
- [x] `npm test` passes
- [x] `make lint` passes
- [x] CI workflows run on push to GitHub
- [x] New developer can set up in <30 minutes following README.md

---

## Dependencies

**None** - This is the foundation phase.

---

## Next Steps

After completing Phase 1:
→ Proceed to [Phase 2: Backend Infrastructure](../phase-2-backend-infrastructure/)
