# SSH & SSL Internals (Go)

## SSH Tunneling
Go's `golang.org/x/crypto/ssh` provides a complete SSH client implementation — no external processes or `askpass` scripts needed.

### Tunnel Setup Flow
```go
func (t *SSHTunnel) Start(config SSHTunnelConfig, localPort int) error {
    // 1. Build auth methods
    var authMethods []ssh.AuthMethod
    switch config.AuthMethod {
    case AuthPassword:
        password, _ := keyring.Get("tablepro", "ssh-password:"+config.ConnectionID)
        authMethods = append(authMethods, ssh.Password(password))
    case AuthKeyFile:
        key, _ := os.ReadFile(config.PrivateKeyPath)
        passphrase, _ := keyring.Get("tablepro", "ssh-passphrase:"+config.ConnectionID)
        signer, _ := ssh.ParsePrivateKeyWithPassphrase(key, []byte(passphrase))
        authMethods = append(authMethods, ssh.PublicKeys(signer))
    case AuthAgent:
        conn, _ := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
        agentClient := agent.NewClient(conn)
        authMethods = append(authMethods, ssh.PublicKeysCallback(agentClient.Signers))
    }

    // 2. Connect to SSH server
    sshConfig := &ssh.ClientConfig{
        User:            config.SSHUser,
        Auth:            authMethods,
        HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: known_hosts
    }
    client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", config.SSHHost, config.SSHPort), sshConfig)

    // 3. Start local TCP listener → forward to remote DB via SSH
    listener, _ := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
    go func() {
        for {
            localConn, _ := listener.Accept()
            remoteConn, _ := client.Dial("tcp", fmt.Sprintf("%s:%d", config.DBHost, config.DBPort))
            go io.Copy(localConn, remoteConn)
            go io.Copy(remoteConn, localConn)
        }
    }()
    return nil
}
```

### Tunnel Health & Recovery
- Background goroutine sends SSH keepalive every 30s
- If tunnel dies, exponential backoff retry (1s, 2s, 4s, 8s, max 30s)
- Emit `runtime.EventsEmit(ctx, "connection:tunnel-reconnecting")` to show UI indicator
- On permanent failure, emit `connection:tunnel-failed` and close the session

### Advantages over Swift Implementation
- **No askpass script needed**: Go handles passphrase input programmatically
- **No subprocess spawning**: Pure Go SSH client (no `ssh` binary dependency)
- **Cross-platform**: Works on macOS, Windows, Linux identically
- **SSH Agent**: Direct Unix socket connection to `ssh-agent`

## SSL/TLS Configuration
Go's `crypto/tls` provides native TLS support used by all database drivers.

```go
type SSLConfig struct {
    Enabled    bool   `json:"enabled"`
    Mode       string `json:"mode"` // "disable", "require", "verify-ca", "verify-full"
    CACertPath string `json:"caCertPath"`
    ClientCert string `json:"clientCertPath"`
    ClientKey  string `json:"clientKeyPath"`
}

func (c *SSLConfig) ToTLSConfig() (*tls.Config, error) {
    tlsConfig := &tls.Config{}

    if c.CACertPath != "" {
        caCert, _ := os.ReadFile(c.CACertPath)
        caCertPool := x509.NewCertPool()
        caCertPool.AppendCertsFromPEM(caCert)
        tlsConfig.RootCAs = caCertPool
    }

    if c.ClientCert != "" && c.ClientKey != "" {
        cert, _ := tls.LoadX509KeyPair(c.ClientCert, c.ClientKey)
        tlsConfig.Certificates = []tls.Certificate{cert}
    }

    switch c.Mode {
    case "require":
        tlsConfig.InsecureSkipVerify = true
    case "verify-ca":
        tlsConfig.InsecureSkipVerify = false
    case "verify-full":
        tlsConfig.InsecureSkipVerify = false
    }

    return tlsConfig, nil
}
```

Each database driver accepts the `*tls.Config` and passes it to the underlying Go DB library:
- PostgreSQL (`pgx`): `pgxpool.Config.ConnConfig.TLSConfig`
- MySQL: `mysql.RegisterTLSConfig("custom", tlsConfig)`
- MongoDB: `options.Client().SetTLSConfig(tlsConfig)`
