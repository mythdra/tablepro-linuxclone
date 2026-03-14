package connection

import "time"

// DatabaseType represents the type of database
type DatabaseType string

const (
	TypePostgreSQL DatabaseType = "postgres"
	TypeMySQL      DatabaseType = "mysql"
	TypeSQLite     DatabaseType = "sqlite"
	TypeDuckDB     DatabaseType = "duckdb"
	TypeMSSQL      DatabaseType = "mssql"
	TypeClickHouse DatabaseType = "clickhouse"
	TypeMongoDB    DatabaseType = "mongodb"
	TypeRedis      DatabaseType = "redis"
)

// ConnectionStatus represents the status of a connection
type ConnectionStatus string

const (
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusConnected    ConnectionStatus = "connected"
	StatusError        ConnectionStatus = "error"
)

// DatabaseConnection represents a database connection configuration
type DatabaseConnection struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Type      DatabaseType `json:"type"`
	Group     string       `json:"group"`
	ColorTag  string       `json:"colorTag"`
	Host      string       `json:"host"`
	Port      int          `json:"port"`
	Database  string       `json:"database"`
	Username  string       `json:"username"`
	LocalFile string       `json:"localFilePath"`

	// SSH configuration
	SSH SSHTunnelConfig `json:"ssh"`

	// SSL configuration
	SSL SSLConfig `json:"ssl"`

	// Advanced settings
	SafeMode         SafeModeLevel `json:"safeMode"`
	StartupCommand   string        `json:"startupCommand"`
	PreConnectScript string        `json:"preConnectScript"`
}

// SSHTunnelConfig represents SSH tunnel configuration
type SSHTunnelConfig struct {
	Enabled    bool   `json:"enabled"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	AuthMethod string `json:"authMethod"` // password, key, agent
	// Password, PrivateKey, Passphrase are NEVER serialized - stored in OS Keychain
}

// SSLConfig represents SSL/TLS configuration
type SSLConfig struct {
	Enabled    bool   `json:"enabled"`
	Mode       string `json:"mode"` // disable, require, verify-ca, verify-full
	CACert     string `json:"caCert"`
	ClientCert string `json:"clientCert"`
	ServerName string `json:"serverName"` // For verify-full mode, defaults to host
	// ClientKey is NEVER serialized - stored in OS Keychain
}

// SafeModeLevel represents the safe mode level for database operations
type SafeModeLevel string

const (
	SafeModeOff      SafeModeLevel = "off"
	SafeModeSafe     SafeModeLevel = "safe"      // Require WHERE for UPDATE/DELETE
	SafeModeVerySafe SafeModeLevel = "very_safe" // Require WHERE + LIMIT
)

// ConnectionSession represents an active connection session
type ConnectionSession struct {
	ConnectionID string           `json:"connectionId"`
	Status       ConnectionStatus `json:"status"`
	ActiveDB     string           `json:"activeDb"`
	Driver       any              `json:"-"` // Database driver instance (not serialized)
	SSHTunnel    any              `json:"-"` // SSHTunnel instance (not serialized)
	LastPingAt   time.Time        `json:"lastPingAt"`
}
