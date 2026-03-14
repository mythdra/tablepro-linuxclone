package config

import (
	"os"
	"path/filepath"
	"strconv"

	"encoding/json"
)

// Config represents the application configuration
type Config struct {
	App      AppConfig      `json:"app"`
	Database DatabaseConfig `json:"database"`
	Log      LogConfig      `json:"log"`
}

// AppConfig holds application-specific settings
type AppConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Debug   bool   `json:"debug"`
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Timeout int `json:"timeout"` // seconds
}

// LogConfig holds logging settings
type LogConfig struct {
	Level     string `json:"level"`
	MaxSize   int    `json:"maxSize"`   // MB
	MaxBackup int    `json:"maxBackup"` // files
	MaxAge    int    `json:"maxAge"`    // days
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:    "TablePro",
			Version: "0.1.0-dev",
			Debug:   false,
		},
		Database: DatabaseConfig{
			Timeout: 30,
		},
		Log: LogConfig{
			Level:     "info",
			MaxSize:   100,
			MaxBackup: 3,
			MaxAge:    28,
		},
	}
}

// Load loads configuration from a JSON file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), nil
	}

	return &cfg, nil
}

// ApplyEnv applies environment variable overrides to the config
func (c *Config) ApplyEnv() {
	if v := os.Getenv("TABLEPRO_DEBUG"); v != "" {
		c.App.Debug = v == "true" || v == "1"
	}

	if v := os.Getenv("TABLEPRO_LOG_LEVEL"); v != "" {
		c.Log.Level = v
	}

	if v := os.Getenv("TABLEPRO_DB_TIMEOUT"); v != "" {
		if timeout, err := strconv.Atoi(v); err == nil {
			c.Database.Timeout = timeout
		}
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Timeout < 1 {
		return &ConfigError{Field: "database.timeout", Message: "timeout must be >= 1"}
	}

	if c.Log.MaxSize < 1 {
		return &ConfigError{Field: "log.maxSize", Message: "maxSize must be >= 1"}
	}

	return nil
}

// Save saves the configuration to a JSON file
func (c *Config) Save(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return e.Field + ": " + e.Message
}
