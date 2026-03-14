package deeplink

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"tablepro/internal/connection"
	"tablepro/internal/log"
)

// DeepLinkAction represents a parsed deep link action
type DeepLinkAction struct {
	Action   string            `json:"action"`   // "open", "new"
	Params   map[string]string `json:"params"`   // Query parameters
	ConnInfo *ConnectionInfo   `json:"connInfo"` // Extracted connection info
}

// ConnectionInfo holds connection parameters from deep link
type ConnectionInfo struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"db"`
	Type     string `json:"type"`
	Username string `json:"user"`
	Password string `json:"pass"` // Only for transient connections
	Name     string `json:"name"`
	Group    string `json:"group"`
}

// DeepLinkHandler handles deep links (tablepro:// URLs)
type DeepLinkHandler struct {
	mu           sync.RWMutex
	queue        []string                                   // Queue for links received before app is ready
	isReady      bool                                       // Whether the app has finished startup
	connectionCB func(*connection.DatabaseConnection) error // Callback to open connection
}

// NewDeepLinkHandler creates a new DeepLinkHandler
func NewDeepLinkHandler() *DeepLinkHandler {
	return &DeepLinkHandler{
		queue:   make([]string, 0),
		isReady: false,
	}
}

// SetConnectionCallback sets the callback function to open a connection
func (h *DeepLinkHandler) SetConnectionCallback(cb func(*connection.DatabaseConnection) error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connectionCB = cb
}

// MarkReady marks the handler as ready and processes any queued links
func (h *DeepLinkHandler) MarkReady() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.isReady = true
	log.Info("DeepLinkHandler marked as ready, processing queue", "queueLength", len(h.queue))

	// Process all queued links
	for _, link := range h.queue {
		go h.process(link)
	}
	h.queue = make([]string, 0) // Clear the queue
}

// Handle processes an incoming deep link URL
func (h *DeepLinkHandler) Handle(rawURL string) {
	log.Info("Received deep link", "url", rawURL)

	h.mu.Lock()
	if !h.isReady {
		// Queue the link for later processing
		h.queue = append(h.queue, rawURL)
		log.Info("Deep link queued (app not ready)", "queueLength", len(h.queue))
		h.mu.Unlock()
		return
	}
	h.mu.Unlock()

	// Process immediately if ready
	go h.process(rawURL)
}

// Parse parses a deep link URL into a DeepLinkAction
func (h *DeepLinkHandler) Parse(rawURL string) (*DeepLinkAction, error) {
	// Parse the URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Verify scheme
	if parsed.Scheme != "tablepro" {
		return nil, fmt.Errorf("invalid scheme: %s (expected tablepro)", parsed.Scheme)
	}

	// Extract action from host (e.g., tablepro://open -> action = "open")
	action := parsed.Host
	if action == "" {
		// Try path-based format: tablepro:///open
		action = strings.TrimPrefix(parsed.Path, "/")
		if action == "" {
			action = "open" // Default action
		}
	}

	// Parse query parameters
	params := make(map[string]string)
	queryParams := parsed.Query()
	for key, values := range queryParams {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	// Extract connection info
	connInfo := &ConnectionInfo{
		Host:     params["host"],
		Port:     parsePort(params["port"]),
		Database: params["db"],
		Type:     params["type"],
		Username: params["user"],
		Password: params["pass"],
		Name:     params["name"],
		Group:    params["group"],
	}

	return &DeepLinkAction{
		Action:   action,
		Params:   params,
		ConnInfo: connInfo,
	}, nil
}

// process handles the deep link - finds or creates connection and opens it
func (h *DeepLinkHandler) process(rawURL string) {
	action, err := h.Parse(rawURL)
	if err != nil {
		log.Error("Failed to parse deep link", "error", err, "url", rawURL)
		return
	}

	log.Info("Processing deep link action", "action", action.Action, "params", action.Params)

	switch action.Action {
	case "open", "connect":
		h.handleOpen(action)
	case "new":
		h.handleNew(action)
	default:
		log.Warn("Unknown deep link action", "action", action.Action)
	}
}

// handleOpen processes an "open" deep link - tries to find existing connection or creates transient
func (h *DeepLinkHandler) handleOpen(action *DeepLinkAction) {
	connInfo := action.ConnInfo
	if connInfo == nil {
		log.Error("No connection info in deep link")
		return
	}

	// Try to find matching existing connection
	conn := h.findMatchingConnection(connInfo)
	if conn != nil {
		log.Info("Found matching connection, opening", "id", conn.ID, "name", conn.Name)
		h.openConnection(conn)
		return
	}

	// No matching connection found - create transient connection
	log.Info("No matching connection found, creating transient connection",
		"host", connInfo.Host, "port", connInfo.Port, "db", connInfo.Database)

	transientConn := h.createTransientConnection(connInfo)
	h.openConnection(transientConn)
}

// handleNew processes a "new" deep link - always creates new transient connection
func (h *DeepLinkHandler) handleNew(action *DeepLinkAction) {
	connInfo := action.ConnInfo
	if connInfo == nil {
		log.Error("No connection info in deep link")
		return
	}

	log.Info("Creating new transient connection",
		"host", connInfo.Host, "port", connInfo.Port, "db", connInfo.Database)

	transientConn := h.createTransientConnection(connInfo)
	h.openConnection(transientConn)
}

// findMatchingConnection searches for an existing connection matching the provided info
func (h *DeepLinkHandler) findMatchingConnection(info *ConnectionInfo) *connection.DatabaseConnection {
	// This would need to access the connection manager to find existing connections
	// For now, we'll implement a basic matching logic that checks:
	// 1. Host matches
	// 2. Port matches
	// 3. Database matches
	// 4. Type matches

	// Note: In a full implementation, this would query the connection manager
	// For now, return nil to always create transient connections
	// TODO: Integrate with ConnectionManager to find matching connections

	log.Debug("findMatchingConnection called", "host", info.Host, "port", info.Port, "db", info.Database, "type", info.Type)

	// Placeholder - would use connection manager to find matching connection
	return nil
}

// createTransientConnection creates a new transient connection from deep link info
func (h *DeepLinkHandler) createTransientConnection(info *ConnectionInfo) *connection.DatabaseConnection {
	connType := connection.DatabaseType(info.Type)
	if connType == "" {
		connType = connection.TypePostgreSQL // Default to PostgreSQL
	}

	// Generate a name if not provided
	name := info.Name
	if name == "" {
		name = generateConnectionName(info)
	}

	conn := &connection.DatabaseConnection{
		ID:       generateUUID(),
		Name:     name,
		Type:     connType,
		Group:    info.Group,
		ColorTag: "#3B82F6", // Default blue color
		Host:     info.Host,
		Port:     info.Port,
		Database: info.Database,
		Username: info.Username,
	}

	log.Info("Created transient connection",
		"name", conn.Name, "type", conn.Type, "host", conn.Host, "port", conn.Port)

	return conn
}

// openConnection opens a connection using the callback
func (h *DeepLinkHandler) openConnection(conn *connection.DatabaseConnection) {
	h.mu.RLock()
	cb := h.connectionCB
	h.mu.RUnlock()

	if cb == nil {
		log.Error("No connection callback set - cannot open connection")
		return
	}

	if err := cb(conn); err != nil {
		log.Error("Failed to open connection", "error", err, "name", conn.Name)
		return
	}

	log.Info("Connection opened successfully", "name", conn.Name)
}

// parsePort parses a port string to int
func parsePort(portStr string) int {
	if portStr == "" {
		return 0
	}
	var port int
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil {
		return 0
	}
	return port
}

// generateConnectionName generates a descriptive name for a connection
func generateConnectionName(info *ConnectionInfo) string {
	dbType := info.Type
	if dbType == "" {
		dbType = "db"
	}
	return fmt.Sprintf("%s://%s:%d/%s", dbType, info.Host, info.Port, info.Database)
}

// generateUUID generates a simple UUID-like string (placeholder)
func generateUUID() string {
	// In production, use github.com/google/uuid
	return fmt.Sprintf("deeplink-%d", time.Now().UnixNano())
}
