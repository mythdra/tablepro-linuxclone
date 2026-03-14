package connection

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// getConfigPath returns the path to the connections JSON file
// Cross-platform: ~/.config/tablepro/connections.json
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	var configDir string

	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Application Support/tablepro
		configDir = filepath.Join(homeDir, "Library", "Application Support", "tablepro")
	case "windows":
		// Windows: %APPDATA%\tablepro
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		configDir = filepath.Join(appData, "tablepro")
	default:
		// Linux/Unix: ~/.config/tablepro
		// Also works for most BSD variants
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig != "" {
			configDir = filepath.Join(xdgConfig, "tablepro")
		} else {
			configDir = filepath.Join(homeDir, ".config", "tablepro")
		}
	}

	return filepath.Join(configDir, "connections.json"), nil
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	var configDir string

	switch runtime.GOOS {
	case "darwin":
		configDir = filepath.Join(homeDir, "Library", "Application Support", "tablepro")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		configDir = filepath.Join(appData, "tablepro")
	default:
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig != "" {
			configDir = filepath.Join(xdgConfig, "tablepro")
		} else {
			configDir = filepath.Join(homeDir, ".config", "tablepro")
		}
	}

	return configDir, nil
}
