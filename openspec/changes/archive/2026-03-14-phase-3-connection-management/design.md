## Context

Phase 2 đã hoàn thành với logging, error handling, và configuration system. Hiện tại đã có `internal/log`, `internal/errors`, `internal/config` packages, và App struct với startup/shutdown lifecycle.

Phase 3 xây dựng Connection Management - tính năng cốt lõi cho database client. Đây là complex phase với nhiều dependencies external (OS Keychain, SSH library, SSL/TLS).

**Ràng buộc**:
- Must use OS Keychain cho password security (cross-platform)
- SSH tunnel với multiple auth methods (password, key, agent)
- SSL modes phải match PostgreSQL spec
- URL parser phải handle dual @ symbols cho SSH URLs
- Timeline: 3-4 weeks

## Goals / Non-Goals

**Goals:**
- Data models cho DatabaseConnection, SSHTunnelConfig, SSLConfig
- Connection CRUD với JSON persistence
- OS Keychain integration cho secure password storage
- SSH tunnel với port forwarding
- SSL/TLS configuration với certificate validation
- Connection URL parser (postgres:// và postgres+ssh://)
- Deep linking với tablepro:// scheme
- Connection testing với 10s timeout
- React form với tabs cho connection management

**Non-Goals:**
- Actual database driver implementation (Phase 4)
- Query execution (Phase 5)
- Connection pooling (Phase 6 - Session Management)
- UI polish beyond basic functionality

## Decisions

### 1. Password Storage: OS Keychain Only
**Rationale**: Security requirement, native integration, user expectation
**Alternatives Considered**: 
- Encrypted file: Less secure, complex key management
- Environment variables: Not persistent, insecure
- Plain JSON: Absolutely not for passwords

### 2. Connection Storage: JSON File
**Rationale**: Human-readable, easy backup, cross-platform
**Alternatives**:
- SQLite: Overkill cho simple connection list
- YAML: No advantage over JSON cho this use case

### 3. SSH Library: golang.org/x/crypto/ssh
**Rationale**: Official Go package, well-maintained, comprehensive
**Alternatives**:
- gliderlabs/ssh: Server-side focused
- ssh package wrappers: Unnecessary abstraction

### 4. UUID Generation: github.com/google/uuid
**Rationale**: Standard, well-tested, RFC 4122 compliant
**Alternatives**:
- github.com/satori/go.uuid: Satori has known issues
- crypto/rand: More complex, overkill

### 5. SSL Modes: PostgreSQL Compatibility
**Rationale**: PostgreSQL has most comprehensive SSL mode support
**Modes**: disable, require, verify-ca, verify-full
**Alternatives**: Simplified modes (on/off) - không đủ cho enterprise users

### 6. URL Parser: Custom Regex for SSH URLs
**Rationale**: Go's url.Parse() cannot handle dual @ symbols
**Pattern**: `^(\w+)\+ssh://([^@]+)@([^/]+)/(.+)$`
**Alternatives**: 
- url.Parse() với manual preprocessing - complex
- Third-party parser - dependency không cần thiết

### 7. Deep Link Handler: Queue Pattern
**Rationale**: App may not be ready when link is received
**Pattern**: Queue links, process after startup complete
**Alternatives**:
- Reject early links - poor UX
- Block startup - delays app launch

### 8. Connection Testing: Separate from Connect
**Rationale**: User feedback before saving connection
**Pattern**: TestConnection() method với 10s timeout
**Alternatives**:
- Test during save - prevents saving untestable connections
- No testing - poor UX, risky

## Risks / Trade-offs

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| OS Keychain unavailable on some Linux | Medium | Medium | Graceful fallback with warning |
| SSH tunnel memory leaks | High | Low | Proper Close() with defer, health checks |
| SSL certificate validation complexity | Medium | Medium | Clear error messages, documentation |
| URL parser edge cases | Low | Low | Comprehensive unit tests |
| Deep link race conditions | Medium | Low | Queue pattern with mutex |
| Connection test hangs | High | Low | 10s timeout với context cancellation |
| Password in logs accidentally | High | Medium | Never log passwords, use struct tags |
| Threading issues in ConnectionManager | High | Medium | sync.RWMutex với proper locking |

## Migration Plan

Not applicable - greenfield development.

## Open Questions

1. **Connection file location**: ~/.config/tablepro/connections.json hay trong app bundle? (recommend: ~/.config)
2. **Keychain fallback strategy**: In-memory encrypted storage hay prompt user? (recommend: warning + in-memory)
3. **SSH agent forwarding**: Support cho multi-hop SSH? (recommend: Phase 3 chỉ single-hop)
