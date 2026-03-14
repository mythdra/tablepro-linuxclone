# Phase 13: License Validation

**Duration**: 1 week | **Priority**: 🟢 Low | **Tasks**: 15

---

## Overview

Implement license key validation system with online activation and offline grace period.

---

## Task Summary

### 13.1 License Model (3 tasks)
- [ ] 13.1.1 Define License struct
- [ ] 13.1.2 Define LicenseTier enum (free, pro, enterprise)
- [ ] 13.1.3 Define feature flags per tier

### 13.2 License Validation (4 tasks)
- [ ] 13.2.1 Implement license key format
- [ ] 13.2.2 Create signature verification
- [ ] 13.2.3 Implement expiration check
- [ ] 13.2.4 Validate against hardware ID

### 13.3 Online Activation (3 tasks)
- [ ] 13.3.1 Implement activation API call
- [ ] 13.3.2 Handle activation limits
- [ ] 13.3.3 Store activation token

### 13.4 Offline Mode (2 tasks)
- [ ] 13.4.1 Cache license state locally
- [ ] 13.4.2 Grace period for offline users

### 13.5 License UI (3 tasks)
- [ ] 13.5.1 Create LicenseDialog component
- [ ] 13.5.2 Show license status and expiry
- [ ] 13.5.3 Upgrade prompt for free users

---

## Acceptance Criteria

- [ ] License validation working
- [ ] Online activation functional
- [ ] Offline mode with grace period
- [ ] License UI displays status
- [ ] Feature flags enforced per tier

---

## Dependencies

← [Phase 12: Settings Management](phase-12-settings.md)  
→ [Phase 14: UI Components](phase-14-ui.md)
