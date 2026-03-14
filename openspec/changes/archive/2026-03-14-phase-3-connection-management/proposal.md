## Why

Phase 2 đã hoàn thành core backend infrastructure (logging, errors, config). Phase 3 tiếp tục xây dựng Connection Management - tính năng nền tảng cho database client. Người dùng cần tạo, chỉnh sửa, test và quản lý kết nối database với SSH/SSL support và secure credential storage qua OS Keychain.

## What Changes

- **Data Models**: DatabaseConnection, SSHTunnelConfig, SSLConfig, ConnectionSession structs
- **Connection CRUD**: ConnectionManager với Save, Load, Update, Delete, Duplicate operations
- **OS Keychain Integration**: Secure password storage qua go-keyring
- **SSH Tunnel Management**: Port forwarding với golang.org/x/crypto/ssh
- **SSL/TLS Configuration**: Certificate validation, client auth modes
- **Connection URL Parser**: Parse postgres:// và postgres+ssh:// URLs
- **Deep Linking**: tablepro:// URL scheme handler
- **Test Connection**: Connection testing với timeout
- **Connection UI**: React form với tabs (General, SSH, SSL, Advanced)

## Capabilities

### New Capabilities
- `connection-models`: Data structures cho connections, SSH, SSL configs
- `connection-crud`: CRUD operations với JSON persistence
- `keychain-integration`: OS Keychain storage cho passwords
- `ssh-tunnel`: SSH port forwarding với multiple auth methods
- `ssl-config`: SSL/TLS configuration với certificate validation
- `url-parser`: Connection URL parsing với query params support
- `deep-linking`: tablepro:// URL scheme handling
- `connection-testing`: Test connectivity với timeout
- `connection-ui`: React form cho connection management

### Modified Capabilities
- (None - đây là foundational connection management, không modify existing capabilities)

## Impact

- **Code**: Tạo packages: internal/connection/, internal/ssh/, internal/deeplink/
- **Dependencies**: 
  - Go: github.com/zalando/go-keyring, golang.org/x/crypto/ssh, github.com/google/uuid
  - Frontend: React Hook Form cho form handling
- **Systems**: OS Keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- **Platforms**: URL scheme registration trên macOS/Windows/Linux
- **Timeline**: 3-4 weeks cho complete implementation
- **Downstream**: Phase 4 (Database Drivers) phụ thuộc vào connection management này
