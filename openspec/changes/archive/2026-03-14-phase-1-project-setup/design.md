## Context

TablePro hiện ở giai đoạn pre-implementation với đầy đủ specifications (20 files trong `/specs/`) và implementation plans (2,900 dòng trong `/plans/`). Tuy nhiên, chưa có code thực tế - repository chỉ chứa design documents.

**Ràng buộc**:
- Must use Go + Wails v2 + React (đã quyết định trong architecture spec)
- Must support macOS, Windows, Linux
- Must follow coding conventions trong `/plans/reference/conventions.md`
- Timeline: 1-2 weeks cho Phase 1

## Goals / Non-Goals

**Goals:**
- Go module initialized với `github.com/tablepro/tablepro`
- Wails v2 project running với React template
- Frontend setup với TypeScript strict mode, Tailwind CSS, AG Grid, Monaco Editor
- Development environment documented (Makefile, VS Code configs)
- CI/CD pipeline working với GitHub Actions
- New developer có thể setup trong <30 minutes

**Non-Goals:**
- Feature implementation (thuộc Phase 3+)
- Production optimization (code signing, notarization)
- Performance tuning
- Security hardening beyond basics

## Decisions

### 1. Go Version: 1.21+
**Rationale**: Wails v2 requires Go 1.21+, stability over bleeding edge
**Alternatives**: Go 1.22 (newer, but less tested với Wails)

### 2. Node.js Version: 18 LTS
**Rationale**: Balance between stability and feature support
**Alternatives**: Node 20 (newer, nhưng có thể gặp compatibility issues với بعض packages)

### 3. Wails v2 vs Tauri/Electron
**Rationale**: Already decided in architecture - single binary, Go backend
**Alternatives Considered**: 
- Tauri: Rust learning curve
- Electron: Too heavy (~100MB+)

### 4. TypeScript Strict Mode: ON
**Rationale**: Catch errors early, better IDE support
**Trade-off**: More verbose code, nhưng worth it for large project

### 5. AG Grid Community vs Enterprise
**Rationale**: Community version đủ cho MVP, Enterprise $850/developer/year
**Future**: Upgrade nếu cần advanced features (pivot, charting)

### 6. Vitest vs Jest
**Rationale**: Faster, better TypeScript support, Vite native
**Alternatives**: Jest (slower, nhưng mature hơn)

### 7. Tailwind CSS vs Styled Components
**Rationale**: Utility-first phù hợp cho data-heavy UI, better performance
**Alternatives**: Material-UI, Ant Design (too opinionated cho custom design)

## Risks / Trade-offs

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Wails WebView inconsistencies | High | Medium | Test early on all platforms |
| npm dependency conflicts | Medium | High | Pin exact versions, use npm ci |
| GOPATH vs Go modules confusion | Low | Medium | Document setup clearly |
| CI/CD quota limits (GitHub Actions) | Medium | Low | Use caches, optimize workflows |
| macOS notarization delays | Medium | Medium | Start process early in Phase 16 |
| AG Grid learning curve | Low | Medium | Allocate time for team training |

## Migration Plan

Not applicable - greenfield project, không có migration.

## Open Questions

1. **Code signing**: Self-signed cho development hay skip until Phase 16?
2. **Docker for testing**: Setup local PostgreSQL/MySQL containers ngay hay đợi Phase 4?
3. **Monorepo vs Separate repos**: Giữ Go + React chung repo hay tách? (recommend: giữ chung)
