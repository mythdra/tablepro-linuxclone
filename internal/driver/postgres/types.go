package postgres

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holds PostgreSQL connection configuration
type Config struct {
	Host      string
	Port      int
	Database  string
	Username  string
	Password  string
	SSLMode   string
	SSLCert   string
	SSLKey    string
	SSLCACert string

	// Connection pool settings
	MinConnections  int32
	MaxConnections  int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// PoolConfig returns a pgxpool.Config from Config
func (c *Config) PoolConfig() (*pgxpool.Config, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.Host, c.Port, c.Database, c.Username, c.Password, c.SSLMode,
	)

	if c.SSLCert != "" && c.SSLKey != "" {
		connStr += fmt.Sprintf(" sslcert=%s sslkey=%s", c.SSLCert, c.SSLKey)
	}
	if c.SSLCACert != "" {
		connStr += fmt.Sprintf(" sslrootcert=%s", c.SSLCACert)
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply pool settings
	if c.MinConnections > 0 {
		config.MinConns = c.MinConnections
	}
	if c.MaxConnections > 0 {
		config.MaxConns = c.MaxConnections
	} else {
		config.MaxConns = 10 // default
	}
	if c.MaxConnLifetime > 0 {
		config.MaxConnLifetime = c.MaxConnLifetime
	}
	if c.MaxConnIdleTime > 0 {
		config.MaxConnIdleTime = c.MaxConnIdleTime
	}

	return config, nil
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            5432,
		SSLMode:         "prefer",
		MinConnections:  2,
		MaxConnections:  10,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
	}
}
