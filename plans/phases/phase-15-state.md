# Phase 15: State Management

**Duration**: 1-2 weeks | **Priority**: 🟠 High | **Tasks**: 15

---

## Overview

Implement global state management with Zustand stores, connecting Go backend events to React state.

---

## Task Summary

### 15.1 Store Architecture (4 tasks)
- [ ] 15.1.1 Create connectionStore
- [ ] 15.1.2 Create tabStore
- [ ] 15.1.3 Create queryStore
- [ ] 15.1.4 Create settingsStore

### 15.2 Wails Events Integration (4 tasks)
- [ ] 15.2.1 Set up EventsOn listeners in React
- [ ] 15.2.2 Create useWailsEvent hook
- [ ] 15.2.3 Map Go events to store actions
- [ ] 15.2.4 Clean up listeners on unmount

### 15.3 State Persistence (3 tasks)
- [ ] 15.3.1 Persist UI state to localStorage
- [ ] 15.3.2 Restore state on app load
- [ ] 15.3.3 Handle state migrations

### 15.4 Async State (2 tasks)
- [ ] 15.4.1 Implement React Query for server state
- [ ] 15.4.2 Configure caching and invalidation

### 15.5 DevTools (2 tasks)
- [ ] 15.5.1 Enable Zustand DevTools
- [ ] 15.5.2 Add state debugging view

---

## Acceptance Criteria

- [ ] All stores implemented
- [ ] Go events update React state
- [ ] State persists across restarts
- [ ] No memory leaks from listeners
- [ ] DevTools showing state changes

---

## Dependencies

← [Phase 14: UI Components](phase-14-ui.md)  
→ [Phase 16: Cross-Platform Build](phase-16-build.md)
