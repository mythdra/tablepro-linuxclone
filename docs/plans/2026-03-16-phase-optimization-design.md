# TablePro Phase Optimization Design

**Date:** 2026-03-16
**Author:** Claude
**Status:** Approved

## Overview

Optimized TablePro's 18 development phases into 12 phases for solo developer workflow.

## Context

- **Developer:** Solo
- **Timeline:** No hard deadline
- **Database Priority:** PostgreSQL first, others later
- **Goal:** Reduce mental overhead while maintaining clear deliverables

## Decision: Consolidate to 12 Phases

### Comparison

| Metric | Original (18 phases) | Optimized (12 phases) |
|--------|---------------------|----------------------|
| **Total Phases** | 18 | 12 |
| **Total Tasks** | ~400 | ~320 |
| **Estimated Duration** | 9-12 months | 8-10 months |
| **MVP Duration** | ~18 weeks | ~17 weeks |
| **Context Switches** | 18 | 12 |

### Key Optimizations

1. **Merged related phases**
   - Export + Import → Phase 7 (Data Transfer Services)
   - History + Settings → Phase 8 (Persistence Services)
   - Setup + Build basics → Phase 1

2. **Split PostgreSQL driver**
   - Original Phase 4 had 8 drivers (60 tasks)
   - PostgreSQL now Phase 3 (35 tasks, dedicated)
   - Other 7 drivers → Phase 9

3. **Integrated testing**
   - No separate testing phase
   - Tests written with each phase

4. **Logical grouping for solo dev**
   - Fewer, larger phases = less context switching
   - Clear deliverables per phase

## Optimized Phase Structure

### Phase Diagram

```
Phase 1: Setup & Infrastructure
    ↓
Phase 2: Backend Core
    ↓
Phase 3: PostgreSQL Driver ──────────────┐
    ↓                                      │
Phase 4: UI Foundation                    │
    ↓                                      │
Phase 5: Data Grid & Mutation             │
    ↓                                      │
Phase 6: SQL Editor                       │
    ↓                                      │
Phase 7: Export/Import Services           │
    ↓                                      │
Phase 8: History & Settings               │
    ↓                                      │
Phase 9: Additional Drivers ←─────────────┘
    ↓
Phase 10: SSH/SSL & Security
    ↓
Phase 11: Licensing & Polish
    ↓
Phase 12: Release & Docs
```

### MVP Critical Path

Phases 1 → 2 → 3 → 4 → 5 → 6 = **~17-21 weeks**

---

## Phase Details

### Phase 1: Setup & Infrastructure (2-3 weeks, 25 tasks)

**Merged from:** Original Phase 1 + parts of Phase 16

**Goals:**
- CMake project với Qt 6.6 + vcpkg
- Basic app window chạy được
- CI/CD pipeline
- Dev environment docs

**Key Tasks:**
- 1.1 CMake + vcpkg setup (CMakeLists.txt, vcpkg.json)
- 1.2 Qt application skeleton (main.cpp, MainWindow)
- 1.3 Dark theme stylesheet
- 1.4 VS Code / Qt Creator config
- 1.5 GitHub Actions CI (build matrix)
- 1.6 README với setup instructions

**Deliverable:** App launches với empty window, CI runs on all platforms

---

### Phase 2: Backend Core (3-4 weeks, 30 tasks)

**Merged from:** Original Phase 2 + parts of Phase 3

**Goals:**
- DatabaseDriver interface
- ConnectionManager với Qt signals/slots
- QueryExecutor basic
- Error handling framework

**Key Tasks:**
- 2.1 DatabaseDriver abstract interface
- 2.2 QueryResult, TableInfo, ColumnInfo structs
- 2.3 ConnectionManager class
- 2.4 QueryExecutor
- 2.5 Error types
- 2.6 Logging system

**Deliverable:** Can define connections, execute mock queries

---

### Phase 3: PostgreSQL Driver (3-4 weeks, 35 tasks)

**Split from:** Original Phase 4 (was 60 tasks for 8 drivers)

**Goals:**
- Complete PostgreSQL driver với libpq
- Schema introspection
- Prepared statements
- Transaction support

**Key Tasks:**
- 3.1 libpq dependency + basic connection
- 3.2 PostgresDriver::connect()
- 3.3 PostgresDriver::execute()
- 3.4 fetchSchema()
- 3.5 fetchTables(), fetchColumns()
- 3.6 Type mapping
- 3.7 Transaction support
- 3.8 Unit tests

**Deliverable:** Full PostgreSQL support

---

### Phase 4: UI Foundation (3-4 weeks, 35 tasks)

**Merged from:** Original Phase 8 + Phase 14 parts

**Goals:**
- MainWindow layout
- Sidebar với schema tree
- Tab management
- Toolbar + status bar

**Key Tasks:**
- 4.1 MainWindow layout
- 4.2 SchemaTree widget
- 4.3 Database selector
- 4.4 TabBar
- 4.5 Tab persistence
- 4.6 Toolbar
- 4.7 StatusBar

**Deliverable:** UI skeleton với working tabs

---

### Phase 5: Data Grid & Mutation (3-4 weeks, 35 tasks)

**From:** Original Phase 7

**Goals:**
- QTableView với virtual scrolling
- Inline cell editing
- Change tracking
- SQL generation for commits

**Key Tasks:**
- 5.1 DataGrid class
- 5.2 ResultSetModel
- 5.3 Column delegates
- 5.4 Inline editing
- 5.5 DataChangeManager
- 5.6 Visual indicators
- 5.7 SqlGenerator
- 5.8 CommitService
- 5.9 Undo/Redo

**Deliverable:** View and edit query results

---

### Phase 6: SQL Editor (2-3 weeks, 25 tasks)

**From:** Original Phase 14 parts

**Goals:**
- QScintilla editor
- Statement execution
- Autocomplete
- Find & Replace

**Key Tasks:**
- 6.1 QScintilla setup
- 6.2 Statement splitter
- 6.3 Execute statements
- 6.4 Autocomplete
- 6.5 Find & Replace

**Deliverable:** Full SQL editor

---

### Phase 7: Export/Import Services (2-3 weeks, 25 tasks)

**Merged from:** Original Phase 9 + Phase 10

**Goals:**
- Export to CSV, JSON, SQL, XLSX
- Import SQL dump files
- Streaming for large files

**Key Tasks:**
- 7.1 ExportService
- 7.2 Streaming export
- 7.3 Export dialog
- 7.4 ImportService
- 7.5 Streaming import
- 7.6 Gzip support
- 7.7 Progress bar

**Deliverable:** Export/Import working

---

### Phase 8: History & Settings (1-2 weeks, 20 tasks)

**Merged from:** Original Phase 11 + Phase 12

**Goals:**
- Query history với FTS
- Application settings
- Theme switch

**Key Tasks:**
- 8.1 HistoryService
- 8.2 Query history panel
- 8.3 SettingsManager
- 8.4 Settings dialog
- 8.5 Theme switcher

**Deliverable:** Searchable history, persisted settings

---

### Phase 9: Additional Drivers (6-8 weeks, 50 tasks)

**From:** Original Phase 4 (remaining drivers)

**Drivers:**
1. MySQL (libmysql)
2. SQLite (Qt SQL)
3. DuckDB (duckdb C API)
4. SQL Server (ODBC)
5. ClickHouse (HTTP/TCP)
6. MongoDB (libmongocxx)
7. Redis (hiredis)

**Deliverable:** All 8 database types supported

---

### Phase 10: SSH/SSL & Security (2 weeks, 20 tasks)

**From:** Original Phase 3 (SSH parts)

**Goals:**
- SSH tunnel support
- SSL/TLS configuration
- Password storage (QKeychain)

**Key Tasks:**
- 10.1 SshTunnel class
- 10.2 SSH key support
- 10.3 SSL configuration
- 10.4 QKeychain integration

**Deliverable:** Secure connections

---

### Phase 11: Licensing & Polish (2 weeks, 20 tasks)

**From:** Original Phase 13 + UI polish

**Goals:**
- License key validation
- Feature gating
- UI polish

**Key Tasks:**
- 11.1 LicenseManager
- 11.2 Feature gating
- 11.3 UI polish
- 11.4 Accessibility

**Deliverable:** Licensing working, UI polished

---

### Phase 12: Release & Docs (1-2 weeks, 15 tasks)

**From:** Original Phase 16 + Phase 18

**Goals:**
- Cross-platform builds
- Code signing
- Documentation
- v1.0.0 release

**Key Tasks:**
- 12.1 macOS app bundle
- 12.2 Windows installer
- 12.3 Linux packages
- 12.4 Documentation
- 12.5 v1.0.0 release

**Deliverable:** v1.0.0 released

---

## MVP Timeline

```
Week 1-3:   Phase 1 (Setup)
Week 4-7:   Phase 2 (Backend Core)
Week 8-11:  Phase 3 (PostgreSQL)
Week 12-15: Phase 4 (UI Foundation)
Week 16-19: Phase 5 (Data Grid)
Week 20-22: Phase 6 (SQL Editor)
────────────────────────────────
Week 22:    MVP READY ✅
Week 23-25: Phase 7 (Export/Import)
Week 26-27: Phase 8 (History/Settings)
Week 28-35: Phase 9 (Additional Drivers)
Week 36-37: Phase 10 (SSH/SSL)
Week 38-39: Phase 11 (Licensing)
Week 40-41: Phase 12 (Release)
────────────────────────────────
Week 41:    v1.0.0 RELEASE ✅
```

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Scope creep | Each phase has clear deliverable |
| Driver complexity | PostgreSQL first as reference |
| Burnout (solo) | No hard deadline, flexible phases |
| Testing debt | Tests written with each phase |

## Next Steps

1. Create implementation plan for each phase
2. Update `/plans/phases/` directory with new structure
3. Archive old phase files