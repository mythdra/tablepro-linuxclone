# Phase 1: Project Setup & Infrastructure

**Duration**: 1-2 weeks | **Priority**: 🔴 Critical | **Tasks**: 25

---

## Overview

This phase establishes the foundation for the entire project. Without proper setup, subsequent phases will face continuous friction.

### Goals
- [ ] Go module initialized with correct structure
- [ ] Wails project running with React template
- [ ] Frontend dependencies installed
- [ ] Development environment documented
- [ ] CI/CD pipeline working

### Non-Goals
- Actual feature implementation
- Complex configurations
- Production build optimization

---

## Task List

### 1.1 Go Module Initialization

#### 1.1.1 Initialize Go module
- **Command**: `go mod init github.com/tablepro/tablepro`
- **Output**: `go.mod` file created
- **Verification**: `go mod tidy` runs without errors
- **Notes**: Use exact module path for consistency

#### 1.1.2 Create directory structure
```
mkdir -p cmd internal frontend
mkdir -p internal/{driver,connection,session,query,change,export,import,history,settings,tab,ssh,license}
```
- **Output**: Directory tree created
- **Verification**: `tree -L 2` shows structure
- **Notes**: Keep structure flat, avoid over-nesting

#### 1.1.3 Add go.mod dependencies
```go
require (
    github.com/wailsapp/wails/v2 v2.8.0
    github.com/jackc/pgx/v5 v5.5.0
    github.com/go-sql-driver/mysql v1.7.1
)
```
- **Output**: go.mod with pinned versions
- **Verification**: `go mod download` succeeds
- **Notes**: Pin all versions for reproducibility

#### 1.1.4 Create .gitignore
```gitignore
# Go
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
/vendor/

# Wails
/build/

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store
Thumbs.db
```
- **Output**: `.gitignore` file
- **Verification**: Git ignores correct files
- **Notes**: Add platform-specific patterns

#### 1.1.5 Set up Go workspace
- Create `go.work` for multi-module development (if needed)
- Configure VS Code Go settings
- Set up gopls language server
- **Verification**: IntelliSense working in VS Code

---

### 1.2 Wails Project Setup

#### 1.2.1 Install Wails CLI
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails version
```
- **Output**: Wails CLI available in GOPATH/bin
- **Verification**: `wails version` returns version number
- **Notes**: Add GOPATH/bin to PATH if not already

#### 1.2.2 Initialize Wails project
```bash
wails init -n tablepro -t react
cd tablepro
```
- **Output**: Wails project structure created
- **Verification**: `wails dev` starts dev server
- **Notes**: May need to merge with existing structure

#### 1.2.3 Configure wails.json
```json
{
  "$scheme": "v2",
  "name": "TablePro",
  "outputfilename": "TablePro",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "TablePro Team",
    "email": "hello@tablepro.app"
  }
}
```
- **Output**: `wails.json` configured
- **Verification**: `wails build` uses correct name
- **Notes**: Update version with releases

#### 1.2.4 Set up app icons
- Create icons in multiple sizes: 1024x1024, 512x512, 256x256, 128x128, 64x64, 32x32, 16x16
- Generate `.icns` for macOS
- Generate `.ico` for Windows
- Generate `.png` for Linux
- Place in `build/appicon/`
- **Verification**: `wails generate icon` succeeds

#### 1.2.5 Configure build options
```json
{
  "platforms": ["darwin", "windows", "linux"],
  "arch": ["amd64", "arm64"],
  "outputType": "platform"
}
```
- **Output**: Build configuration ready
- **Verification**: `wails build -platform darwin` succeeds
- **Notes**: Start with native platform only

---

### 1.3 Frontend Setup

#### 1.3.1 Install npm dependencies
```bash
cd frontend
npm install
```
**Required packages**:
```json
{
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "@ag-grid-community/react": "^31.0.0",
    "@monaco-editor/react": "^4.6.0",
    "zustand": "^4.5.0",
    "tailwindcss": "^3.4.0",
    "lucide-react": "^0.350.0"
  }
}
```
- **Verification**: `npm install` completes without errors
- **Notes**: Use exact versions for reproducibility

#### 1.3.2 Configure TypeScript
```json
{
  "compilerOptions": {
    "target": "ES2020",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "moduleResolution": "bundler"
  }
}
```
- **Output**: `tsconfig.json` configured
- **Verification**: `tsc --noEmit` passes
- **Notes**: Enable strict mode from start

#### 1.3.3 Set up Tailwind CSS
```javascript
// tailwind.config.js
module.exports = {
  content: ['./src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: '#3B82F6',
        success: '#10B981',
        warning: '#F59E0B',
        danger: '#EF4444',
      }
    }
  },
  plugins: []
}
```
- **Verification**: Tailwind classes compile correctly
- **Notes**: Add custom theme colors matching TablePro design

#### 1.3.4 Configure Vitest
```javascript
// vitest.config.ts
import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
  }
})
```
- **Verification**: `npm test` runs test suite
- **Notes**: Configure for React Testing Library

#### 1.3.5 Set up ESLint + Prettier
```javascript
// .eslintrc.js
module.exports = {
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:react/recommended',
    'plugin:react-hooks/recommended'
  ],
  parser: '@typescript-eslint/parser',
  rules: {
    'react/react-in-jsx-scope': 'off'
  }
}
```
- **Verification**: `npm run lint` passes
- **Notes**: Match Go formatting philosophy - opinionated defaults

---

### 1.4 Development Environment

#### 1.4.1 Create Makefile
```makefile
.PHONY: dev build test lint clean

dev:
	wails dev

build:
	wails build

build-all:
	wails build -platform darwin
	wails build -platform windows
	wails build -platform linux

test:
	go test ./...
	cd frontend && npm test

lint:
	go vet ./...
	cd frontend && npm run lint

clean:
	rm -rf build/
	go clean
```
- **Verification**: `make dev` starts development
- **Notes**: Keep commands simple and memorable

#### 1.4.2 VS Code launch configurations
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Wails App",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/main.go"
    },
    {
      "name": "Frontend Tests",
      "type": "node",
      "request": "launch",
      "program": "${workspaceFolder}/frontend/node_modules/vitest/vitest.mjs",
      "cwd": "${workspaceFolder}/frontend"
    }
  ]
}
```
- **Verification**: F5 starts debugger
- **Notes**: Include frontend debugging too

#### 1.4.3 Configure Go language server
```json
{
  "gopls": {
    "formatting.gofumpt": true,
    "analyses": {
      "unusedparams": true,
      "shadow": true,
      "nilness": true
    }
  }
}
```
- **Verification**: Go to definition works
- **Notes**: Enable all useful static analysis

#### 1.4.4 Create .env.example
```bash
# Development
WAILS_DEBUG=true
LOG_LEVEL=debug

# Database (for testing)
TEST_POSTGRES_URL=postgres://localhost:5432/tablepro_test
TEST_MYSQL_URL=root@localhost:3306/tablepro_test
```
- **Verification**: Copy to .env and app starts
- **Notes**: Never commit real credentials

#### 1.4.5 Document setup in README.md
```markdown
# TablePro

Cross-platform database client built with Go + Wails + React.

## Quick Start

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Install dependencies
go mod download
cd frontend && npm install

# Run development
wails dev
```

## Requirements

- Go 1.21+
- Node.js 18+
- Wails v2.8+
```
- **Verification**: New developer can follow and succeed
- **Notes**: Keep it simple, link to detailed docs

---

### 1.5 CI/CD Setup

#### 1.5.1 GitHub Actions for Go tests
```yaml
# .github/workflows/go-tests.yml
name: Go Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: go test -v ./...
```
- **Verification**: Workflow runs on push
- **Notes**: Add test coverage reporting

#### 1.5.2 Frontend test workflow
```yaml
# .github/workflows/frontend-tests.yml
name: Frontend Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '18'
      - run: cd frontend && npm ci
      - run: cd frontend && npm test
```
- **Verification**: Workflow runs on push
- **Notes**: Cache node_modules for speed

#### 1.5.3 Build matrix workflow
```yaml
# .github/workflows/build.yml
name: Build

on: [release]

jobs:
  build:
    strategy:
      matrix:
        platform: [darwin, windows, linux]
        arch: [amd64, arm64]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: wails build -platform ${{ matrix.platform }}
```
- **Verification**: Builds all platforms on release
- **Notes**: Add code signing in future

#### 1.5.4 Code coverage reporting
```yaml
- run: go test -coverprofile=coverage.out ./...
- uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out
```
- **Verification**: Coverage visible in PR checks
- **Notes**: Set 80% threshold

#### 1.5.5 Linting checks
```yaml
- run: go vet ./...
- uses: golangci/golangci-lint-action@v3
- run: cd frontend && npm run lint
```
- **Verification**: PRs blocked on lint failures
- **Notes**: Fix all existing issues first

---

## Deliverables

| Item | Location | Status |
|------|----------|--------|
| Go module | `/go.mod` | ⬜ Not Started |
| Wails project | `/` | ⬜ Not Started |
| Frontend setup | `/frontend/` | ⬜ Not Started |
| Makefile | `/Makefile` | ⬜ Not Started |
| CI/CD workflows | `/.github/workflows/` | ⬜ Not Started |
| README.md | `/README.md` | ⬜ Not Started |

---

## Acceptance Criteria

- [ ] `wails dev` starts without errors
- [ ] `go test ./...` passes
- [ ] `npm test` passes
- [ ] `make lint` passes
- [ ] CI workflows run on push
- [ ] New developer can set up in <30 minutes

---

## Dependencies

**None** - This is the foundation phase.

---

## Next Phase

→ [Phase 2: Backend Infrastructure](phase-02-backend.md)
