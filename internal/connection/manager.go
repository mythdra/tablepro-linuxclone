package connection

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

// ConnectionManager handles database connection CRUD operations
type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]*DatabaseConnection
	configPath  string
	keychain    Keychainer
}

// Keychainer defines the interface for keychain operations
type Keychainer interface {
	DeletePassword(service, account string) error
	SetPassword(service, account, password string) error
	GetPassword(service, account string) (string, error)
}

// NewConnectionManager creates a new ConnectionManager instance
func NewConnectionManager() (*ConnectionManager, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	mgr := &ConnectionManager{
		connections: make(map[string]*DatabaseConnection),
		configPath:  configPath,
		keychain:    &defaultKeychain{},
	}

	// Load existing connections
	if err := mgr.Load(); err != nil {
		// If file doesn't exist, that's ok - start with empty connections
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load connections: %w", err)
		}
	}

	return mgr, nil
}

// Save adds a new connection or updates existing one
func (m *ConnectionManager) Save(conn *DatabaseConnection) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	// Generate UUID if not provided
	if conn.ID == "" {
		conn.ID = uuid.New().String()
	}

	// Validate before saving
	if err := m.Validate(conn); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.connections[conn.ID] = conn

	return m.persist()
}

// Load reads all connections from the JSON file
func (m *ConnectionManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	var connections []*DatabaseConnection
	if err := json.Unmarshal(data, &connections); err != nil {
		return fmt.Errorf("failed to parse connections file: %w", err)
	}

	m.connections = make(map[string]*DatabaseConnection)
	for _, conn := range connections {
		if conn != nil && conn.ID != "" {
			m.connections[conn.ID] = conn
		}
	}

	return nil
}

// Delete removes a connection by ID
func (m *ConnectionManager) Delete(id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid connection ID")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	connID := id.String()
	conn, exists := m.connections[connID]
	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	// Clean up keychain entries
	if m.keychain != nil {
		passwordKey := fmt.Sprintf("password:%s", connID)
		m.keychain.DeletePassword("tablepro", passwordKey)

		if conn.SSH.Enabled {
			sshPasswordKey := fmt.Sprintf("ssh-password:%s", connID)
			m.keychain.DeletePassword("tablepro", sshPasswordKey)
		}
	}

	delete(m.connections, connID)

	return m.persist()
}

// Duplicate creates a copy of an existing connection with a new UUID
func (m *ConnectionManager) Duplicate(id uuid.UUID) (*DatabaseConnection, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid connection ID")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	connID := id.String()
	original, exists := m.connections[connID]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connID)
	}

	// Create a deep copy
	newConn := &DatabaseConnection{
		ID:               uuid.New().String(),
		Name:             original.Name + " (Copy)",
		Type:             original.Type,
		Group:            original.Group,
		ColorTag:         original.ColorTag,
		Host:             original.Host,
		Port:             original.Port,
		Database:         original.Database,
		Username:         original.Username,
		LocalFile:        original.LocalFile,
		SSH:              original.SSH,
		SSL:              original.SSL,
		SafeMode:         original.SafeMode,
		StartupCommand:   original.StartupCommand,
		PreConnectScript: original.PreConnectScript,
	}

	// Copy SSH credentials from keychain if needed
	if m.keychain != nil && original.SSH.Enabled {
		originalPasswordKey := fmt.Sprintf("ssh-password:%s", connID)
		if password, err := m.keychain.GetPassword("tablepro", originalPasswordKey); err == nil && password != "" {
			newPasswordKey := fmt.Sprintf("ssh-password:%s", newConn.ID)
			m.keychain.SetPassword("tablepro", newPasswordKey, password)
		}
	}

	// Copy main database password from keychain if needed
	passwordKey := fmt.Sprintf("password:%s", connID)
	if password, err := m.keychain.GetPassword("tablepro", passwordKey); err == nil && password != "" {
		newPasswordKey := fmt.Sprintf("password:%s", newConn.ID)
		m.keychain.SetPassword("tablepro", newPasswordKey, password)
	}

	return newConn, nil
}

// Update modifies an existing connection
func (m *ConnectionManager) Update(conn *DatabaseConnection) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	if conn.ID == "" {
		return fmt.Errorf("connection ID is required for update")
	}

	// Validate before updating
	if err := m.Validate(conn); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.connections[conn.ID]; !exists {
		return fmt.Errorf("connection not found: %s", conn.ID)
	}

	m.connections[conn.ID] = conn

	return m.persist()
}

// Validate checks if a connection has all required fields
func (m *ConnectionManager) Validate(conn *DatabaseConnection) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}

	if conn.Name == "" {
		return fmt.Errorf("connection name is required")
	}

	if conn.Type == "" {
		return fmt.Errorf("database type is required")
	}

	// Validate based on database type
	switch conn.Type {
	case TypeSQLite, TypeDuckDB:
		// SQLite/DuckDB can use local file
		if conn.LocalFile == "" && conn.Host == "" {
			return fmt.Errorf("either localFile or host is required for %s", conn.Type)
		}
	case TypeMongoDB, TypeRedis:
		// MongoDB and Redis require host
		if conn.Host == "" {
			return fmt.Errorf("host is required for %s", conn.Type)
		}
	default:
		// PostgreSQL, MySQL, MSSQL, ClickHouse require host and port
		if conn.Host == "" {
			return fmt.Errorf("host is required for %s", conn.Type)
		}
		if conn.Port == 0 {
			return fmt.Errorf("port is required for %s", conn.Type)
		}
	}

	// Validate SSH configuration if enabled
	if conn.SSH.Enabled {
		if conn.SSH.Host == "" {
			return fmt.Errorf("SSH host is required when SSH is enabled")
		}
		if conn.SSH.Username == "" {
			return fmt.Errorf("SSH username is required when SSH is enabled")
		}
		if conn.SSH.AuthMethod == "" {
			return fmt.Errorf("SSH auth method is required when SSH is enabled")
		}
	}

	// Validate SSL configuration if enabled
	if conn.SSL.Enabled {
		if conn.SSL.Mode == "" {
			return fmt.Errorf("SSL mode is required when SSL is enabled")
		}
	}

	return nil
}

// Get returns a connection by ID
func (m *ConnectionManager) Get(id uuid.UUID) (*DatabaseConnection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn, exists := m.connections[id.String()]
	return conn, exists
}

// List returns all connections
func (m *ConnectionManager) List() []*DatabaseConnection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	connections := make([]*DatabaseConnection, 0, len(m.connections))
	for _, conn := range m.connections {
		connections = append(connections, conn)
	}

	return connections
}

// persist saves connections to the JSON file
func (m *ConnectionManager) persist() error {
	// Ensure directory exists
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	connections := make([]*DatabaseConnection, 0, len(m.connections))
	for _, conn := range m.connections {
		connections = append(connections, conn)
	}

	data, err := json.MarshalIndent(connections, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal connections: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write connections file: %w", err)
	}

	return nil
}

// defaultKeychain is a no-op implementation for keychain operations
type defaultKeychain struct{}

func (k *defaultKeychain) DeletePassword(service, account string) error {
	// No-op - actual implementation would use go-keyring
	return nil
}

func (k *defaultKeychain) SetPassword(service, account, password string) error {
	// No-op - actual implementation would use go-keyring
	return nil
}

func (k *defaultKeychain) GetPassword(service, account string) (string, error) {
	// No-op - return empty string
	return "", nil
}
