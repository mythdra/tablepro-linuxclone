package connection

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"
	"time"

	"tablepro/internal/errors"
	"tablepro/internal/ssh"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

// TestConnectionResult represents the result of a connection test
type TestConnectionResult struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	ResponseTime int64  `json:"responseTimeMs,omitempty"` // Response time in milliseconds
}

// TestConnection tests a database connection and returns detailed results
// Uses 10-second timeout for the entire operation
func (m *ConnectionManager) TestConnection(ctx context.Context, conn *DatabaseConnection) (*TestConnectionResult, error) {
	if conn == nil {
		return &TestConnectionResult{
			Success: false,
			Message: "connection configuration is required",
		}, nil
	}

	// Create context with 10-second timeout
	testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	startTime := time.Now()

	// Start SSH tunnel if needed
	var tunnel *ssh.SSHTunnel
	var tunnelCleanup func() error

	if conn.SSH.Enabled {
		tunnelResult, err := m.startSSHTunnelForTest(testCtx, conn)
		if err != nil {
			return &TestConnectionResult{
				Success: false,
				Message: err.Error(),
			}, nil
		}
		tunnel = tunnelResult.tunnel
		tunnelCleanup = tunnelResult.cleanup
	}

	// Clean up SSH tunnel when done
	if tunnelCleanup != nil {
		defer tunnelCleanup()
	}

	// Determine actual host/port (use tunnel local address if SSH is enabled)
	host := conn.Host
	port := conn.Port

	if tunnel != nil {
		localAddr := tunnel.LocalAddress()
		parts := strings.Split(localAddr, ":")
		if len(parts) == 2 {
			host = "127.0.0.1"
			fmt.Sscanf(parts[1], "%d", &port)
		}
	}

	// Get password from keychain
	password := ""
	if m.keychain != nil {
		passwordKey := fmt.Sprintf("password:%s", conn.ID)
		pwd, err := m.keychain.GetPassword("tablepro", passwordKey)
		if err == nil {
			password = pwd
		}
	}

	// Build connection string and test
	dsn, err := m.buildDSN(conn, host, port, password)
	if err != nil {
		return &TestConnectionResult{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Try to open connection
	db, err := sql.Open(getDriverName(conn.Type), dsn)
	if err != nil {
		return &TestConnectionResult{
			Success: false,
			Message: m.formatConnectionError(err, conn),
		}, nil
	}
	defer db.Close()

	// Set connection timeout
	db.SetConnMaxLifetime(5 * time.Second)

	// Ping the database to verify connection works
	pingCtx, pingCancel := context.WithTimeout(testCtx, 5*time.Second)
	defer pingCancel()

	if err := db.PingContext(pingCtx); err != nil {
		return &TestConnectionResult{
			Success: false,
			Message: m.formatConnectionError(err, conn),
		}, nil
	}

	responseTime := time.Since(startTime).Milliseconds()

	return &TestConnectionResult{
		Success:      true,
		Message:      fmt.Sprintf("Successfully connected to %s", conn.Name),
		ResponseTime: responseTime,
	}, nil
}

// startSSHTunnelForTest starts an SSH tunnel for testing purposes
func (m *ConnectionManager) startSSHTunnelForTest(ctx context.Context, conn *DatabaseConnection) (*tunnelResult, error) {
	// Get SSH password from keychain
	sshPassword := ""
	if m.keychain != nil {
		sshPasswordKey := fmt.Sprintf("ssh-password:%s", conn.ID)
		pwd, err := m.keychain.GetPassword("tablepro", sshPasswordKey)
		if err == nil {
			sshPassword = pwd
		}
	}

	// Get SSH private key path from keychain (stored separately)
	privateKeyPath := ""
	if m.keychain != nil {
		privateKeyKey := fmt.Sprintf("ssh-private-key:%s", conn.ID)
		keyPath, err := m.keychain.GetPassword("tablepro", privateKeyKey)
		if err == nil {
			privateKeyPath = keyPath
		}
	}

	// Get SSH passphrase from keychain
	sshPassphrase := ""
	if m.keychain != nil {
		passphraseKey := fmt.Sprintf("ssh-passphrase:%s", conn.ID)
		phrase, err := m.keychain.GetPassword("tablepro", passphraseKey)
		if err == nil {
			sshPassphrase = phrase
		}
	}

	tunnelConfig := ssh.SSHTunnelConfig{
		Host:              conn.SSH.Host,
		Port:              conn.SSH.Port,
		Username:          conn.SSH.Username,
		Password:          sshPassword,
		PrivateKeyPath:    privateKeyPath,
		UseAgent:          conn.SSH.AuthMethod == "agent",
		KeyPassphrase:     sshPassphrase,
		RemoteHost:        conn.Host,
		RemotePort:        conn.Port,
		LocalPort:         0, // Auto-assign
		ConnectTimeout:    10 * time.Second,
		KeepAliveInterval: 30 * time.Second,
	}

	tunnel := ssh.NewSSHTunnel(tunnelConfig)

	if err := tunnel.Start(); err != nil {
		return nil, errors.New(errors.ErrSSHFailed,
			fmt.Sprintf("SSH tunnel failed: %s. Verify SSH host, port, and authentication credentials.", err)).
			WithContext("host", conn.SSH.Host).
			WithContext("port", conn.SSH.Port)
	}

	return &tunnelResult{
		tunnel: tunnel,
		cleanup: func() error {
			return tunnel.Close()
		},
	}, nil
}

type tunnelResult struct {
	tunnel  *ssh.SSHTunnel
	cleanup func() error
}

// buildDSN builds a database connection string from the configuration
func (m *ConnectionManager) buildDSN(conn *DatabaseConnection, host string, port int, password string) (string, error) {
	switch conn.Type {
	case TypePostgreSQL:
		return buildPostgreSQLDSN(conn, host, port, password), nil
	case TypeMySQL:
		return buildMySQLDSN(conn, host, port, password), nil
	case TypeSQLite:
		return conn.LocalFile, nil
	case TypeMSSQL:
		return buildMSSQLDSN(conn, host, port, password), nil
	case TypeClickHouse:
		return buildClickHouseDSN(conn, host, port, password), nil
	case TypeMongoDB:
		return buildMongoDBDSN(conn, host, port, password), nil
	case TypeRedis:
		return buildRedisDSN(conn, host, port, password), nil
	default:
		return "", errors.New(errors.ErrInvalidConfig,
			fmt.Sprintf("unsupported database type: %s", conn.Type))
	}
}

// buildPostgreSQLDSN builds PostgreSQL connection string
func buildPostgreSQLDSN(conn *DatabaseConnection, host string, port int, password string) string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, conn.Username, password, conn.Database, getSSLMode(conn.SSL))

	if conn.SSL.Enabled && conn.SSL.ServerName != "" {
		dsn += fmt.Sprintf(" sslservername=%s", conn.SSL.ServerName)
	}

	return dsn
}

// buildMySQLDSN builds MySQL connection string
func buildMySQLDSN(conn *DatabaseConnection, host string, port int, password string) string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=5s",
		conn.Username, password, host, port, conn.Database)

	if conn.SSL.Enabled {
		dsn += "&tls=skip-verify"
	}

	return dsn
}

// buildMSSQLDSN builds MSSQL connection string
func buildMSSQLDSN(conn *DatabaseConnection, host string, port int, password string) string {
	return fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s;connection timeout=5",
		host, port, conn.Username, password, conn.Database)
}

// buildClickHouseDSN builds ClickHouse connection string
func buildClickHouseDSN(conn *DatabaseConnection, host string, port int, password string) string {
	return fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s",
		conn.Username, password, host, port, conn.Database)
}

// buildMongoDBDSN builds MongoDB connection string
func buildMongoDBDSN(conn *DatabaseConnection, host string, port int, password string) string {
	auth := ""
	if conn.Username != "" && password != "" {
		auth = fmt.Sprintf("%s:%s@", conn.Username, password)
	}
	return fmt.Sprintf("mongodb://%s%s:%d/%s?connectTimeoutMS=5000",
		auth, host, port, conn.Database)
}

// buildRedisDSN builds Redis connection string
func buildRedisDSN(conn *DatabaseConnection, host string, port int, password string) string {
	auth := ""
	if password != "" {
		auth = password + "@"
	}
	return fmt.Sprintf("%s%s:%d",
		auth, host, port)
}

// getSSLMode returns the appropriate SSL mode string
func getSSLMode(ssl SSLConfig) string {
	if !ssl.Enabled {
		return "disable"
	}
	switch ssl.Mode {
	case "require":
		return "require"
	case "verify-ca":
		return "verify-ca"
	case "verify-full":
		return "verify-full"
	default:
		return "require"
	}
}

// getDriverName returns the database driver name for sql.Open
func getDriverName(dbType DatabaseType) string {
	switch dbType {
	case TypePostgreSQL:
		return "pgx"
	case TypeMySQL:
		return "mysql"
	case TypeSQLite:
		return "sqlite3"
	case TypeMSSQL:
		return "mssql"
	case TypeClickHouse:
		return "clickhouse"
	case TypeMongoDB:
		return "mongo"
	case TypeRedis:
		return "redis"
	default:
		return string(dbType)
	}
}

// formatConnectionError converts raw database errors to user-friendly messages
func (m *ConnectionManager) formatConnectionError(err error, conn *DatabaseConnection) string {
	errStr := err.Error()
	lowerErr := strings.ToLower(errStr)

	// Connection refused
	if strings.Contains(lowerErr, "connection refused") || strings.Contains(lowerErr, "dial tcp") {
		return fmt.Sprintf("Connection refused: check host '%s' and port '%d'. Verify the database server is running and accessible.", conn.Host, conn.Port)
	}

	// Authentication failed
	if strings.Contains(lowerErr, "authentication") || strings.Contains(lowerErr, "password") ||
		strings.Contains(lowerErr, "invalid credentials") || strings.Contains(lowerErr, "access denied") {
		return fmt.Sprintf("Authentication failed: verify username '%s' and password are correct.", conn.Username)
	}

	// Timeout
	if strings.Contains(lowerErr, "timeout") || strings.Contains(lowerErr, "i/o timeout") ||
		strings.Contains(lowerErr, "context deadline") {
		return fmt.Sprintf("Connection timeout: database server at '%s:%d' is not responding. Check network connectivity and firewall settings.", conn.Host, conn.Port)
	}

	// Database not found
	if strings.Contains(lowerErr, "database") && (strings.Contains(lowerErr, "does not exist") || strings.Contains(lowerErr, "unknown database")) {
		return fmt.Sprintf("Database '%s' does not exist. Verify the database name is correct.", conn.Database)
	}

	// SSL/TLS errors
	if strings.Contains(lowerErr, "ssl") || strings.Contains(lowerErr, "tls") || strings.Contains(lowerErr, "certificate") {
		if conn.SSL.Enabled {
			return fmt.Sprintf("SSL/TLS error: check SSL configuration. You may need to verify the CA certificate or disable SSL if not required.")
		}
		return fmt.Sprintf("SSL/TLS error: try disabling SSL or verify the server's certificate configuration.")
	}

	// No route to host
	if strings.Contains(lowerErr, "no route") || strings.Contains(lowerErr, "network is unreachable") {
		return fmt.Sprintf("Network error: cannot reach '%s'. Check network connectivity and firewall rules.", conn.Host)
	}

	// EOF / connection closed
	if strings.Contains(lowerErr, "eof") || strings.Contains(lowerErr, "connection reset") || strings.Contains(lowerErr, "broken pipe") {
		return fmt.Sprintf("Connection closed unexpectedly: database server at '%s:%d' closed the connection. Check if the server is running.", conn.Host, conn.Port)
	}

	// Default: provide generic error with original message
	return fmt.Sprintf("Connection failed: %s", errStr)
}

// TestConnectionTCP performs a simple TCP connection test (fallback when no driver available)
func TestConnectionTCP(ctx context.Context, host string, port int, timeout time.Duration) error {
	address := fmt.Sprintf("%s:%d", host, port)

	dialer := &net.Dialer{
		Timeout: timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}
