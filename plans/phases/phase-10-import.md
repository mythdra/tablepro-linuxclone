# Phase 10: Import Service

**Duration**: 2 weeks | **Priority**: 🟡 Medium | **Tasks**: 20

---

## Overview

Implement data import functionality supporting CSV, JSON, and SQL files with validation and error handling.

---

## Task Summary

### 10.1 Import Interface (3 tasks)
- [ ] 10.1.1 Define Importer interface
- [ ] 10.1.2 Define ImportOptions struct
- [ ] 10.1.3 Define ImportResult struct

### 10.2 CSV Importer (4 tasks)
- [ ] 10.2.1 Parse CSV with various delimiters
- [ ] 10.2.2 Handle quoted fields and escapes
- [ ] 10.2.3 Detect column types automatically
- [ ] 10.2.4 Skip/parse header row

### 10.3 JSON Importer (3 tasks)
- [ ] 10.3.1 Parse JSON array format
- [ ] 10.3.2 Parse JSON Lines format
- [ ] 10.3.3 Flatten nested objects

### 10.4 SQL Importer (3 tasks)
- [ ] 10.4.1 Parse INSERT statements
- [ ] 10.4.2 Execute in batches
- [ ] 10.4.3 Handle transaction rollback

### 10.5 Import Validation (4 tasks)
- [ ] 10.5.1 Validate data types match target columns
- [ ] 10.5.2 Check constraint violations
- [ ] 10.5.3 Preview data before import
- [ ] 10.5.4 Report validation errors

### 10.6 Import UI (3 tasks)
- [ ] 10.6.1 Create ImportDialog component
- [ ] 10.6.2 File picker with format detection
- [ ] 10.6.3 Map source columns to target table

---

## Acceptance Criteria

- [ ] CSV/JSON/SQL import working
- [ ] Data validation before import
- [ ] Error reporting clear
- [ ] Import progress shown
- [ ] Rollback on failure

---

## Dependencies

← [Phase 9: Export Service](phase-09-export.md)  
→ [Phase 11: Query History](phase-11-history.md)
