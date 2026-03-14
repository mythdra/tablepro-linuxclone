package connection

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Default ports for each database type
var defaultPorts = map[DatabaseType]int{
	TypePostgreSQL: 5432,
	TypeMySQL:      3306,
	TypeMongoDB:    27017,
	TypeRedis:      6379,
	TypeSQLite:     0,
	TypeDuckDB:     0,
	TypeMSSQL:      1433,
	TypeClickHouse: 8123,
}

// ParsedConnection represents the result of parsing a connection URL
type ParsedConnection struct {
	// Basic connection info
	Scheme   string `json:"scheme"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`

	// Database type (derived from scheme)
	Type DatabaseType `json:"type"`

	// SSH tunnel configuration
	SSH struct {
		Enabled  bool   `json:"enabled"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Query parameters
	Params map[string]string

	// Raw URL for reference
	RawURL string
}

// ParseConnectionURL parses a database connection URL and returns a ParsedConnection struct.
// Supports standard URLs like postgres://user:pass@host:port/db and SSH URLs like postgres+ssh://user:pass@host:port/db
func ParseConnectionURL(rawURL string) (*ParsedConnection, error) {
	parsed := &ParsedConnection{
		RawURL: rawURL,
		Params: make(map[string]string),
	}

	// Trim whitespace
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, fmt.Errorf("empty URL")
	}

	// Check for SSH URL pattern: scheme+ssh://user@host/path
	sshRegex := regexp.MustCompile(`^(\w+)\+ssh://([^@]+)@([^/]+)/(.+)$`)
	matches := sshRegex.FindStringSubmatch(rawURL)

	if matches != nil {
		// SSH URL: postgres+ssh://user:pass@sshhost:port/db or postgres+ssh://user:pass@sshhost/db
		return parseSSHURL(rawURL, matches, parsed)
	}

	// Standard URL: scheme://user:pass@host:port/db
	return parseStandardURL(rawURL, parsed)
}

// parseSSHURL parses an SSH tunnel URL
func parseSSHURL(rawURL string, matches []string, parsed *ParsedConnection) (*ParsedConnection, error) {
	// matches: [full, scheme, credentials, hostWithPort, database]
	scheme := matches[1]
	credentials := matches[2] // user:pass
	hostPart := matches[3]    // host:port
	database := matches[4]

	// Parse scheme (e.g., "postgres+ssh")
	parsed.Scheme = scheme + "+ssh"
	parsed.Type = canonicalType(strings.TrimSuffix(scheme, "+ssh"))

	// Validate database type
	if !isValidScheme(string(parsed.Type)) {
		return nil, fmt.Errorf("unsupported database type: %s", parsed.Type)
	}

	// Parse credentials (user:pass)
	parsed.Username, parsed.Password = parseCredentials(credentials)

	// Parse host and port
	parsed.Host, parsed.Port = parseHostPort(hostPart)

	// Set database (may contain query params)
	parsed.Database = extractDatabaseWithParams(database, parsed.Params)

	// Enable SSH tunnel using the same host as connection
	parsed.SSH.Enabled = true
	parsed.SSH.Host = parsed.Host
	parsed.SSH.Port = 22 // Default SSH port
	parsed.SSH.Username = parsed.Username
	parsed.SSH.Password = parsed.Password

	// Parse SSL mode from params
	parseSSLMode(parsed)

	return parsed, nil
}

// parseStandardURL parses a standard database URL
func parseStandardURL(rawURL string, parsed *ParsedConnection) (*ParsedConnection, error) {
	// Use net/url for standard URL parsing
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	parsed.Scheme = u.Scheme
	parsed.Type = canonicalType(u.Scheme)

	// Validate database type
	if !isValidScheme(parsed.Scheme) {
		return nil, fmt.Errorf("unsupported database type: %s", parsed.Scheme)
	}

	// Parse credentials
	parsed.Username = u.User.Username()
	parsed.Password, _ = u.User.Password()

	// Parse host and port
	parsed.Host = u.Hostname()
	portStr := u.Port()
	if portStr != "" {
		parsed.Port, _ = strconv.Atoi(portStr)
	} else {
		parsed.Port = defaultPorts[parsed.Type]
	}

	// Parse database (path) - remove leading slash
	parsed.Database = strings.TrimPrefix(u.Path, "/")

	// Extract query parameters
	if u.RawQuery != "" {
		queryParams, _ := url.ParseQuery(u.RawQuery)
		for k, v := range queryParams {
			if len(v) > 0 {
				parsed.Params[k] = v[0]
			}
		}
	}

	// Parse SSL mode from params
	parseSSLMode(parsed)

	return parsed, nil
}

// parseCredentials extracts username and password from credential string
func parseCredentials(cred string) (username, password string) {
	parts := strings.SplitN(cred, ":", 2)
	if len(parts) >= 1 {
		username = parts[0]
	}
	if len(parts) >= 2 {
		password = parts[1]
	}
	return username, password
}

// parseHostPort extracts host and port from host:port string
func parseHostPort(hostPart string) (host string, port int) {
	parts := strings.SplitN(hostPart, ":", 2)
	host = parts[0]
	if len(parts) >= 2 {
		port, _ = strconv.Atoi(parts[1])
	}
	return host, port
}

// extractDatabaseWithParams extracts database name and query params from path
func extractDatabaseWithParams(path string, params map[string]string) string {
	// Check if path contains query string
	if idx := strings.Index(path, "?"); idx != -1 {
		dbPath := path[:idx]
		queryStr := path[idx+1:]

		// Parse query params
		queryParts := strings.Split(queryStr, "&")
		for _, qp := range queryParts {
			kv := strings.SplitN(qp, "=", 2)
			if len(kv) == 2 {
				params[kv[0]] = kv[1]
			}
		}
		return dbPath
	}
	return path
}

// parseSSLMode extracts and applies SSL mode from params
func parseSSLMode(parsed *ParsedConnection) {
	// SSL mode can be passed as sslmode or ssl_mode
	if sslmode, ok := parsed.Params["sslmode"]; ok {
		parsed.Params["sslmode"] = sslmode
	} else if sslmode, ok := parsed.Params["ssl_mode"]; ok {
		parsed.Params["sslmode"] = sslmode
	}
}

// isValidScheme checks if the given scheme is supported
func isValidScheme(scheme string) bool {
	validSchemes := map[string]bool{
		"postgres":   true,
		"postgresql": true,
		"mysql":      true,
		"mongodb":    true,
		"mongo":      true,
		"redis":      true,
		"sqlite":     true,
		"duckdb":     true,
		"mssql":      true,
		"sqlserver":  true,
		"clickhouse": true,
	}
	return validSchemes[scheme]
}

func canonicalType(scheme string) DatabaseType {
	switch scheme {
	case "postgres", "postgresql":
		return TypePostgreSQL
	case "mysql":
		return TypeMySQL
	case "mongodb", "mongo":
		return TypeMongoDB
	case "redis":
		return TypeRedis
	case "sqlite":
		return TypeSQLite
	case "duckdb":
		return TypeDuckDB
	case "mssql", "sqlserver":
		return TypeMSSQL
	case "clickhouse":
		return TypeClickHouse
	default:
		return DatabaseType(scheme)
	}
}

// GetConnectionString reconstructs a connection URL from ParsedConnection
func (p *ParsedConnection) GetConnectionString() string {
	var builder strings.Builder

	if p.SSH.Enabled {
		builder.WriteString(string(p.Type))
		builder.WriteString("+ssh://")
	} else {
		builder.WriteString(p.Scheme)
		builder.WriteString("://")
	}

	if p.Username != "" {
		builder.WriteString(p.Username)
		if p.Password != "" {
			builder.WriteString(":")
			builder.WriteString(p.Password)
		}
		builder.WriteString("@")
	}

	builder.WriteString(p.Host)
	if p.Port > 0 {
		builder.WriteString(":")
		builder.WriteString(strconv.Itoa(p.Port))
	}

	if p.Database != "" {
		builder.WriteString("/")
		builder.WriteString(p.Database)
	}

	// Add query params
	if len(p.Params) > 0 {
		builder.WriteString("?")
		first := true
		for k, v := range p.Params {
			if !first {
				builder.WriteString("&")
			}
			builder.WriteString(k)
			builder.WriteString("=")
			builder.WriteString(v)
			first = false
		}
	}

	return builder.String()
}
