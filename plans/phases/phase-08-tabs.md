# Phase 8: Tab Management

**Duration**: 2 weeks | **Priority**: 🟡 Medium | **Tasks**: 20

---

## Overview

Implement tab management for multiple query tabs per connection with state persistence and restoration.

---

## Task Summary

### 8.1 Tab Data Model (3 tasks)
- [ ] 8.1.1 Define QueryTab struct
- [ ] 8.1.2 Define TabState enum (active, background, closed)
- [ ] 8.1.3 Create TypeScript type definitions

### 8.2 Tab CRUD Operations (5 tasks)
- [ ] 8.2.1 Create TabManager service
- [ ] 8.2.2 Implement CreateTab() method
- [ ] 8.2.3 Implement CloseTab() method
- [ ] 8.2.4 Implement SwitchTab() method
- [ ] 8.2.5 Implement DuplicateTab() method

### 8.3 Tab State Persistence (5 tasks)
- [ ] 8.3.1 Save tab state to JSON file per connection
- [ ] 8.3.2 Restore tabs on connection reopen
- [ ] 8.3.3 Persist query text in tabs
- [ ] 8.3.4 Save tab groupings
- [ ] 8.3.5 Handle concurrent tab edits

### 8.4 Tab UI Component (4 tasks)
- [ ] 8.4.1 Create TabBar component
- [ ] 8.4.2 Implement tab drag-and-drop reordering
- [ ] 8.4.3 Add tab close button (x)
- [ ] 8.4.4 Show unsaved changes indicator

### 8.5 Tab Keyboard Shortcuts (3 tasks)
- [ ] 8.5.1 Ctrl/Cmd+T for new tab
- [ ] 8.5.2 Ctrl/Cmd+W for close tab
- [ ] 8.5.3 Ctrl/Cmd+Tab for switch tab

---

## Acceptance Criteria

- [ ] Create/close/switch tabs working
- [ ] Tab state persisted to disk
- [ ] Tabs restored on reconnection
- [ ] Tab bar UI intuitive
- [ ] Keyboard shortcuts functional

---

## Dependencies

← [Phase 7: Data Grid & Mutation](phase-07-datagrid.md)  
→ [Phase 9: Export Service](phase-09-export.md)
