# Phase 16: Cross-Platform Build

**Duration**: 1 week | **Priority**: 🟡 Medium | **Tasks**: 20

---

## Overview

Configure and test production builds for macOS, Windows, and Linux with proper signing and packaging.

---

## Task Summary

### 16.1 Build Configuration (4 tasks)
- [ ] 16.1.1 Configure wails.json for all platforms
- [ ] 16.1.2 Set up build tags per platform
- [ ] 16.1.3 Configure native Windows icons
- [ ] 16.1.4 Configure native macOS icons

### 16.2 macOS Build (4 tasks)
- [ ] 16.2.1 Build for macOS (amd64, arm64)
- [ ] 16.2.2 Configure entitlements
- [ ] 16.2.3 Code signing (if applicable)
- [ ] 16.2.4 Notarization (if applicable)

### 16.3 Windows Build (4 tasks)
- [ ] 16.3.1 Build for Windows (amd64)
- [ ] 16.3.2 Configure Windows manifest
- [ ] 16.3.3 Code signing (if applicable)
- [ ] 16.3.4 Create NSIS installer

### 16.4 Linux Build (4 tasks)
- [ ] 16.4.1 Build for Linux (amd64, arm64)
- [ ] 16.4.2 Create .deb package
- [ ] 16.4.3 Create .rpm package
- [ ] 16.4.4 Create AppImage

### 16.5 Build Automation (4 tasks)
- [ ] 16.5.1 Create Makefile build targets
- [ ] 16.5.2 Set up GitHub Actions build matrix
- [ ] 16.5.3 Configure artifact upload
- [ ] 16.5.4 Add build versioning

---

## Acceptance Criteria

- [ ] macOS build runs correctly
- [ ] Windows build runs correctly
- [ ] Linux build runs correctly
- [ ] All packages install correctly
- [ ] CI/CD builds automatically

---

## Dependencies

← [Phase 15: State Management](phase-15-state.md)  
→ [Phase 17: Testing & Quality](phase-17-testing.md)
