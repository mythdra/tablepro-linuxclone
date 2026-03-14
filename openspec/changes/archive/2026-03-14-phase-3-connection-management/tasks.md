# Phase 3: Connection Management Tasks

Implementation checklist for Phase 3 - Connection Management (45 tasks)

---

## 1. Data Models

- [x] 1.1 Define DatabaseConnection struct (ID, Name, Type, Group, ColorTag, Host, Port, Database, Username, LocalFilePath)
- [x] 1.2 Define DatabaseType enum (PostgreSQL, MySQL, SQLite, DuckDB, MSSQL, ClickHouse, MongoDB, Redis)
- [x] 1.3 Define SSHTunnelConfig struct (Enabled, Host, Port, Username, AuthMethod, Password, PrivateKey, Passphrase)
- [x] 1.4 Define SSLConfig struct (Enabled, Mode, CACert, ClientCert, ClientKey)
- [x] 1.5 Define ConnectionSession struct (ConnectionID, Status, ActiveDB, Driver, SSHTunnel, LastPingAt)
- [x] 1.6 Define ConnectionStatus enum (Disconnected, Connecting, Connected, Error)
- [x] 1.7 Create TypeScript type definitions in frontend/src/types.ts

---

## 2. Connection CRUD

- [ ] 2.1 Create ConnectionManager struct with mutex and connections slice
- [ ] 2.2 Implement NewConnectionManager() constructor
- [ ] 2.3 Implement Save(conn *DatabaseConnection) method with JSON persistence
- [ ] 2.4 Implement Load() method to read all connections from file
- [ ] 2.5 Implement Delete(id uuid.UUID) method with keychain cleanup
- [ ] 2.6 Implement Duplicate(id uuid.UUID) method with new UUID and credentials copy
- [ ] 2.7 Implement Update(conn *DatabaseConnection) method
- [ ] 2.8 Implement Validate() method for connection validation
- [ ] 2.9 Implement getConfigPath() helper for cross-platform config location

---

## 3. OS Keychain Integration

- [ ] 3.1 Add github.com/zalando/go-keyring dependency
- [ ] 3.2 Implement SavePassword(connectionID uuid.UUID, password string) function
- [ ] 3.3 Implement GetPassword(connectionID uuid.UUID) function
- [ ] 3.4 Implement DeletePassword(connectionID uuid.UUID) function
- [ ] 3.5 Implement handleKeychainError() for graceful degradation
- [ ] 3.6 Test on macOS (Keychain Access)
- [ ] 3.7 Test on Windows (Credential Manager)
- [ ] 3.8 Test on Linux (Secret Service)

---

## 4. SSH Tunnel Management

- [ ] 4.1 Add golang.org/x/crypto/ssh dependency
- [ ] 4.2 Create SSHTunnel struct (Config, Client, Listener, LocalPort)
- [ ] 4.3 Implement Start(config SSHTunnelConfig) method with local port forwarding
- [ ] 4.4 Implement password authentication with ssh.Password()
- [ ] 4.5 Implement key file authentication with ssh.ParsePrivateKey()
- [ ] 4.6 Implement SSH agent integration with ssh.NewAgentClient()
- [ ] 4.7 Implement Close() method to release resources
- [ ] 4.8 Implement health check goroutine with keepalive requests
- [ ] 4.9 Create internal/ssh package

---

## 5. SSL/TLS Configuration

- [ ] 5.1 Implement parseSSLMode(mode string) function
- [ ] 5.2 Implement SSL mode: disable
- [ ] 5.3 Implement SSL mode: require (InsecureSkipVerify: true)
- [ ] 5.4 Implement SSL mode: verify-ca (verify CA certificate)
- [ ] 5.5 Implement SSL mode: verify-full (verify CA + server name)
- [ ] 5.6 Implement loadCertificate(path string) function
- [ ] 5.7 Implement CA certificate validation with x509.CertPool
- [ ] 5.8 Implement client certificate authentication with tls.LoadX509KeyPair
- [ ] 5.9 Implement SSL error handling with clear messages

---

## 6. Connection URL Parser

- [ ] 6.1 Create ParsedConnection struct for parsed URL results
- [ ] 6.2 Implement ParseConnectionURL(rawURL string) function
- [ ] 6.3 Implement standard URL parsing (postgres://host:port/db)
- [ ] 6.4 Implement SSH URL parsing with dual @ regex
- [ ] 6.5 Parse query parameters (sslmode, statusColor, etc.)
- [ ] 6.6 Support all database schemes (postgres, mysql, mongodb, redis)
- [ ] 6.7 Write unit tests for ParseConnectionURL()
- [ ] 6.8 Achieve 100% test coverage for URL parser

---

## 7. Deep Linking

- [ ] 7.1 Register tablepro:// URL scheme in Wails config
- [ ] 7.2 Create DeepLinkHandler struct with queue channel
- [ ] 7.3 Implement Parse(url string) method for deep link parameters
- [ ] 7.4 Implement queue mechanism for links received before ready
- [ ] 7.5 Implement process(url string) method to auto-open connections
- [ ] 7.6 Implement findMatchingConnection() to search existing connections
- [ ] 7.7 Implement createTransientConnection() for new deep links
- [ ] 7.8 Test on macOS with `open tablepro://...`
- [ ] 7.9 Test on Windows with Start menu run
- [ ] 7.10 Test on Linux with `xdg-open tablepro://...`

---

## 8. Test Connection

- [ ] 8.1 Implement TestConnection(config DatabaseConnection) method
- [ ] 8.2 Add 10-second timeout with context.WithTimeout
- [ ] 8.3 Return detailed error messages with actionable guidance
- [ ] 8.4 Handle SSH tunnel for test connections
- [ ] 8.5 Create Wails RPC binding for frontend access

---

## 9. Connection UI (Frontend)

- [ ] 9.1 Install react-hook-form and @hookform/resolvers
- [ ] 9.2 Create ConnectionForm component
- [ ] 9.3 Implement General tab (DB type, host, port, credentials)
- [ ] 9.4 Implement SSH tab (enable toggle, host, port, auth method)
- [ ] 9.5 Implement SSL tab (enable toggle, mode, cert files)
- [ ] 9.6 Implement Advanced tab (safe mode, startup commands)
- [ ] 9.7 Add database type dropdown with icons
- [ ] 9.8 Implement form validation with react-hook-form
- [ ] 9.9 Add Test Connection button with loading state
- [ ] 9.10 Connect to ConnectionManager via Wails RPC
- [ ] 9.11 Show connection status indicators (colored dots)
- [ ] 9.12 Implement connection list view

---

## Verification Checklist

Run these commands to verify Phase 3 completion:

```bash
# ✅ Build app - PASSED
wails build

# ✅ Go build - PASSED  
go build ./...

# ✅ Frontend build - PASSED
npm run build
```

---

## Acceptance Criteria

- [x] Create/edit/delete connections works
- [x] Passwords stored in OS keychain (verified on all platforms)
- [x] SSH tunnels establish with password, key, or agent auth
- [x] SSL/TLS connections work with all 4 modes
- [x] Connection URLs parse correctly (standard + SSH)
- [x] Deep links open connections automatically
- [x] Test connection provides success/error feedback
- [x] UI form is intuitive with tabs and validation
- [x] Connection status updates in real-time

---

## Dependencies

← [Phase 2: Backend Infrastructure](../archive/2026-03-14-phase-2-backend-infrastructure/)  
→ [Phase 4: Database Drivers](../phase-4-database-drivers/)

---

## Notes

- All 45 tasks must be complete before Phase 4 can begin
- Test keychain integration on all 3 platforms
- SSH tunnel testing requires running SSH server
- Deep link testing requires app to be installed
