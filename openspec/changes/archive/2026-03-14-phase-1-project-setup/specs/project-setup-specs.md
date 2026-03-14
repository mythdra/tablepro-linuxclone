# Phase 1: Project Setup Specifications

## ADDED Requirements

### Capability: project-structure

The system SHALL provide a standardized Go module structure for TablePro.

#### Scenario: Go module initialized
- **WHEN** developer runs `go mod init github.com/tablepro/tablepro`
- **THEN** `go.mod` file is created with correct module path

#### Scenario: Directory structure created
- **WHEN** setup completes
- **THEN** directories exist: `cmd/`, `internal/`, `frontend/`

#### Scenario: Internal packages scaffolded
- **WHEN** setup completes
- **THEN** subdirectories exist: `internal/{driver,connection,session,query,change,export,import,history,settings,tab,ssh,license}`

### Capability: wails-app

The system SHALL run a Wails v2 application with React frontend.

#### Scenario: Wails CLI installed
- **WHEN** developer runs `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **THEN** `wails version` returns version number

#### Scenario: Wails project initialized
- **WHEN** developer runs `wails init -n tablepro -t react`
- **THEN** Wails project structure is created with React template

#### Scenario: Development server starts
- **WHEN** developer runs `wails dev`
- **THEN** Development server starts and opens browser to app

#### Scenario: wails.json configured
- **WHEN** setup completes
- **THEN** `wails.json` contains app name, version, and build settings

### Capability: frontend-toolchain

The system SHALL provide a complete frontend development toolchain.

#### Scenario: npm dependencies installed
- **WHEN** developer runs `npm install` in frontend directory
- **THEN** All required packages are installed: react, @ag-grid-community/react, @monaco-editor/react, zustand, tailwindcss, lucide-react

#### Scenario: TypeScript configured
- **WHEN** setup completes
- **THEN** `tsconfig.json` enables strict mode with target ES2020

#### Scenario: Tailwind CSS configured
- **WHEN** setup completes
- **THEN** `tailwind.config.js` includes custom theme colors (primary, success, warning, danger)

#### Scenario: Vitest configured
- **WHEN** developer runs `npm test`
- **THEN** Test suite runs with jsdom environment

#### Scenario: ESLint and Prettier configured
- **WHEN** developer runs `npm run lint`
- **THEN** Linting passes without errors

### Capability: dev-environment

The system SHALL provide a complete development environment setup.

#### Scenario: Makefile created
- **WHEN** setup completes
- **THEN** `Makefile` contains targets: dev, build, test, lint, clean

#### Scenario: VS Code launch configurations
- **WHEN** developer presses F5
- **THEN** Debugger starts for Go backend

#### Scenario: Go language server configured
- **WHEN** developer opens Go file
- **THEN** IntelliSense and go-to-definition work

#### Scenario: .env.example created
- **WHEN** setup completes
- **THEN** `.env.example` contains development settings and test database URLs

#### Scenario: README.md setup documented
- **WHEN** new developer follows README.md
- **THEN** Development environment is ready in <30 minutes

### Capability: ci-cd-pipeline

The system SHALL provide automated CI/CD pipelines via GitHub Actions.

#### Scenario: Go tests workflow
- **WHEN** code is pushed to repository
- **THEN** GitHub Actions runs `go test -v ./...`

#### Scenario: Frontend tests workflow
- **WHEN** code is pushed to repository
- **THEN** GitHub Actions runs `npm test` in frontend directory

#### Scenario: Build matrix workflow
- **WHEN** release is created
- **THEN** GitHub Actions builds for darwin, windows, linux platforms

#### Scenario: Code coverage reporting
- **WHEN** tests run
- **THEN** Coverage report is generated and uploaded to Codecov

#### Scenario: Linting checks
- **WHEN** pull request is created
- **THEN** GitHub Actions runs `go vet` and `npm run lint`
