# TablePro Tauri Design

Date: 2026-03-17
Status: Approved
Stack: Tauri 2.0 + React 18 + TypeScript + shadcn/ui + AG-Grid + Monaco Editor

## Tech Stack

| Layer | Technology |
|-------|------------|
| Desktop Shell | Tauri 2.0 |
| Backend | Rust with sqlx, tokio |
| Frontend | React 18 + TypeScript |
| UI Library | shadcn/ui + Tailwind CSS |
| DataGrid | AG-Grid Community |
| SQL Editor | Monaco Editor |
| State Management | Zustand + React Query |
| Build | Vite + tauri-cli |

## Architecture

```
┌─────────────────────────────────────────┐
│           React Frontend                │
│  ┌─────────────────────────────────┐   │
│  │  Components: Sidebar, Editor,  │   │
│  │  Grid, Dialogs, Tabs            │   │
│  └─────────────────────────────────┘   │
│  ┌─────────────────────────────────┐   │
│  │  Hooks: useConnection,         │   │
│  │  useQuery, useSchema            │   │
│  └─────────────────────────────────┘   │
│  ┌─────────────────────────────────┐   │
│  │  State: Zustand store          │   │
│  └─────────────────────────────────┘   │
└──────────────────┬──────────────────────┘
                   │ Tauri IPC (invoke)
┌──────────────────▼──────────────────────┐
│           Rust Backend                 │
│  ┌─────────────────────────────────┐   │
│  │  Commands: connection, query,  │   │
│  │  schema, export                │   │
│  └─────────────────────────────────┘   │
│  ┌─────────────────────────────────┐   │
│  │  DB Layer: sqlx + PgPool       │   │
│  └─────────────────────────────────┘   │
│  ┌─────────────────────────────────┐   │
│  │  Error + Secure Storage        │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

## Database Support Priority

1. PostgreSQL (Phase 1)
2. SQLite (Phase 2)
3. MySQL (Phase 2)
4. DuckDB, Redis (Future)

## Key Features (from specs)

- Connection Management (PostgreSQL, SSL, SSH tunneling)
- Schema Browser (sidebar tree)
- SQL Editor with Monaco
- Data Grid with AG-Grid (virtual scrolling, editing)
- Tab Management
- Export (CSV, JSON, SQL, XLSX)
- Import (SQL files)
- Query History
- AI Integration (future)
- Settings

## Design Decisions

1. **Password Storage**: OS Keychain (keyring crate)
2. **State Management**: Zustand for global, React Query for async
3. **Error Handling**: Custom error types with Serialize
4. **Testing**: Vitest + React Testing Library + Playwright E2E
5. **CI/CD**: GitHub Actions for cross-platform builds

## Phased Implementation

### Phase 1: Foundation
- Project setup (Tauri + React + shadcn/ui)
- PostgreSQL connection
- Basic query execution
- Simple data display

### Phase 2: Core Features
- SQL Editor (Monaco)
- Data editing
- Tab management
- Connection management UI

### Phase 3: Advanced Features
- SSH tunneling
- SSL/TLS configuration
- Export/Import
- Query history

### Phase 4: Polish & Extras
- AI integration
- Quick switcher
- Settings
- Licensing
