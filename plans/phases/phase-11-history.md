# Phase 11: Query History

**Duration**: 1-2 weeks | **Priority**: 🟡 Medium | **Tasks**: 15

---

## Overview

Implement persistent query history with SQLite FTS5 for full-text search, allowing users to search and reuse past queries.

---

## Task Summary

### 11.1 History Storage (4 tasks)
- [ ] 11.1.1 Initialize embedded SQLite database
- [ ] 11.1.2 Create history table with FTS5
- [ ] 11.1.3 Define HistoryEntry struct
- [ ] 11.1.4 Implement storage path management

### 11.2 History Recording (3 tasks)
- [ ] 11.2.1 Save executed queries automatically
- [ ] 11.2.2 Store execution metadata (duration, rows)
- [ ] 11.2.3 Link queries to connections

### 11.3 Full-Text Search (3 tasks)
- [ ] 11.3.1 Implement FTS5 search queries
- [ ] 11.3.2 Support partial matching
- [ ] 11.3.3 Rank results by relevance

### 11.4 History UI (3 tasks)
- [ ] 11.4.1 Create HistoryPanel component
- [ ] 11.4.2 Search input with debounce
- [ ] 11.4.3 Click to load query in editor

### 11.5 History Management (2 tasks)
- [ ] 11.5.1 Clear history for connection
- [ ] 11.5.2 Set max history size limit

---

## Acceptance Criteria

- [ ] Queries saved to history automatically
- [ ] Full-text search working
- [ ] History panel displays results
- [ ] Click to reuse query working
- [ ] History cleanup functional

---

## Dependencies

← [Phase 10: Import Service](phase-10-import.md)  
→ [Phase 12: Settings Management](phase-12-settings.md)
