# Dependencies

## Go Dependencies

### Core Framework
| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/wailsapp/wails/v2` | v2.8.0+ | Desktop framework |
| `github.com/jackc/pgx/v5` | v5.5.0+ | PostgreSQL driver |
| `github.com/go-sql-driver/mysql` | v1.7.1+ | MySQL/MariaDB driver |
| `github.com/mattn/go-sqlite3` | v1.14.18+ | SQLite driver |

### Database Drivers (Additional)
| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/marcboeker/go-duckdb` | latest | DuckDB driver |
| `github.com/microsoft/go-mssqldb` | latest | SQL Server driver |
| `github.com/ClickHouse/clickhouse-go/v2` | v2.19.0+ | ClickHouse driver |
| `go.mongodb.org/mongo-driver` | v1.13.0+ | MongoDB driver |
| `github.com/redis/go-redis/v9` | v9.3.0+ | Redis driver |

### Infrastructure
| Package | Version | Purpose |
|---------|---------|---------|
| `golang.org/x/crypto/ssh` | latest | SSH tunneling |
| `github.com/zalando/go-keyring` | v0.2.3+ | OS Keychain access |
| `github.com/google/uuid` | v1.4.0+ | UUID generation |
| `github.com/mattn/go-sqlite3` | v1.14.18+ | Query history (FTS5) |
| `gopkg.in/natefinch/lumberjack.v2` | v2.2.1+ | Log rotation |

---

## Frontend Dependencies

### Core Libraries
| Package | Version | Purpose |
|---------|---------|---------|
| `react` | ^18.2.0 | UI framework |
| `react-dom` | ^18.2.0 | React DOM rendering |
| `typescript` | ^5.3.0 | Type safety |

### UI Components
| Package | Version | Purpose |
|---------|---------|---------|
| `@ag-grid-community/react` | ^31.0.0 | Data grid |
| `@monaco-editor/react` | ^4.6.0 | SQL editor |
| `@radix-ui/react-*` | latest | UI primitives |
| `lucide-react` | ^0.350.0 | Icons |
| `react-resizable-panels` | ^2.0.0 | Resizable panes |

### State & Utilities
| Package | Version | Purpose |
|---------|---------|---------|
| `zustand` | ^4.5.0 | State management |
| `tailwindcss` | ^3.4.0 | Styling |

### Testing
| Package | Version | Purpose |
|---------|---------|---------|
| `vitest` | ^1.3.0 | Test runner |
| `@testing-library/react` | ^14.2.0 | React testing |
| `@playwright/test` | ^1.42.0 | E2E testing |

### Linting & Formatting
| Package | Version | Purpose |
|---------|---------|---------|
| `eslint` | ^8.57.0 | Linting |
| `prettier` | ^3.2.0 | Formatting |
| `@typescript-eslint/*` | ^7.0.0 | TypeScript ESLint |

---

## Version Compatibility

### Go Versions
- **Minimum**: Go 1.21
- **Recommended**: Go 1.22+
- **CI Tests**: Go 1.21, 1.22

### Node.js Versions
- **Minimum**: Node.js 18
- **Recommended**: Node.js 20 LTS
- **CI Tests**: Node.js 18, 20

### Platform Support
| Platform | Versions | Architectures |
|----------|----------|---------------|
| macOS | 12+ (Monterey) | amd64, arm64 |
| Windows | 10+ | amd64 |
| Linux | Ubuntu 20.04+, Fedora 36+ | amd64 |

---

## Alternative Packages Considered

### Desktop Framework
| Package | Status | Reason Rejected |
|---------|--------|-----------------|
| Electron | ❌ | Too heavy (~100MB+) |
| Tauri | ❌ | Rust learning curve |
| Qt | ❌ | C++ complexity, licensing |

### State Management
| Package | Status | Reason Rejected |
|---------|--------|-----------------|
| Redux | ❌ | Too much boilerplate |
| Recoil | ❌ | Experimental, less stable |
| Jotai | ⚠️ | Considered, Zustand simpler |

### Data Grid
| Package | Status | Reason Rejected |
|---------|--------|-----------------|
| React Table | ❌ | No virtual scrolling built-in |
| TanStack Table | ❌ | Same as above |
| Handsontable | ❌ | Licensing costs |

---

## Dependency Management

### Go
```bash
# Update all dependencies
go get -u ./...

# Tidy go.mod
go mod tidy

# Verify dependencies
go mod verify
```

### npm
```bash
# Update dependencies
npm update

# Check for outdated
npm outdated

# Audit for vulnerabilities
npm audit
```

---

## Security Considerations

1. **Pin all versions** in go.mod and package.json
2. **Regular audits**: `go mod verify`, `npm audit`
3. **Monitor CVEs** for critical dependencies
4. **Update schedule**: Monthly minor, quarterly major
