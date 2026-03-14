# Connection Management Flow (Go + Wails)

## 1. Connection Storage
```go
type ConnectionManager struct {
    ctx         context.Context
    connections []DatabaseConnection
    configPath  string // ~/.config/tablepro/connections.json
}
```
- **Metadata**: Connections serialized as JSON array to `~/.config/tablepro/connections.json`
- **Passwords**: Stored in OS Keychain via `go-keyring`
  - Key: `tablepro:password:{UUID}`, `tablepro:ssh-password:{UUID}`, `tablepro:ssh-passphrase:{UUID}`
- **Duplicate**: Generate new UUID, suffix name with " (Copy)", copy Keychain entries to new keys

## 2. Connection Form (React)
- React modal with tabs: General, SSH, SSL, Advanced
- All form state managed locally in React (`useState`)
- On Save: calls `ConnectionManager.Save(connection)` + `ConnectionManager.SavePassword(id, password)`
- On Test: calls `ConnectionManager.TestConnection(config)` — returns success/error message
- Database type dropdown dynamically shows relevant fields (e.g., MongoDB shows Auth Source)

## 3. URL Parser (Go)
```go
func ParseConnectionURL(rawURL string) (*ParsedConnectionURL, error) {
    // Handle schemes: postgres://, mysql://, mongodb://, redis://
    // Handle SSH schemes: postgres+ssh://sshuser@bastion:22/dbuser:pass@dbhost:5432/mydb
    // Handle query params: ?sslmode=require&statusColor=red&env=Production

    // Step 1: Check for "+ssh" in scheme
    // Step 2: Split SSH and DB parts manually (url.Parse breaks on dual @)
    // Step 3: Parse query parameters for SSL, color, environment
    // Step 4: Return ParsedConnectionURL struct
}

type ParsedConnectionURL struct {
    DatabaseType DatabaseType
    Host         string
    Port         int
    Database     string
    Username     string
    Password     string
    SSHHost      string
    SSHPort      int
    SSHUser      string
    SSLMode      string
    StatusColor  string
    Schema       string
    TableName    string
    FilterColumn string
    FilterOp     string
    FilterValue  string
}
```

## 4. Deep Linking (Wails)
- Register `tablepro://` URL scheme in Wails app config
- On URL received: parse with `ParseConnectionURL()`
- Search existing connections for match (host + port + database + username)
- If match found → open that connection
- If no match → create transient in-memory connection
- Queue URLs if app not fully loaded (buffer until main window ready)
- Post-connect actions: switch schema, open table, apply filter

## 5. Test Connection Flow
```go
func (cm *ConnectionManager) TestConnection(config DatabaseConnection) error {
    // 1. Create driver for the database type
    driver, err := NewDriver(config.Type)
    if err != nil { return err }

    // 2. If SSH enabled, start temporary tunnel
    var localPort int
    if config.SSH.Enabled {
        tunnel := &SSHTunnel{}
        localPort, err = tunnel.Start(config.SSH)
        if err != nil { return fmt.Errorf("SSH tunnel failed: %w", err) }
        defer tunnel.Close()
        config.Host = "127.0.0.1"
        config.Port = localPort
    }

    // 3. Attempt connection with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    return driver.TestConnection(ctx, config)
}
```

## 6. Pgpass Detection (PostgreSQL only)
```go
func CheckPgpass() *PgpassWarning {
    pgpassPath := filepath.Join(os.Getenv("HOME"), ".pgpass")
    info, err := os.Stat(pgpassPath)
    if err != nil { return nil } // File doesn't exist
    
    mode := info.Mode().Perm()
    if mode & 0077 != 0 { // More permissive than 0600
        return &PgpassWarning{
            Message: "~/.pgpass has insecure permissions. Run: chmod 600 ~/.pgpass",
        }
    }
    return nil
}
```
