// Package query provides query history tracking with in-memory storage.
package query

import (
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// QueryHistoryEntry represents a query in the history.
// Contains execution metadata including timing, success status, and result info.
type QueryHistoryEntry struct {
	ID         uuid.UUID
	Query      string
	Timestamp  time.Time
	Duration   time.Duration
	Success    bool
	Error      string
	RowCount   int64
	Connection uuid.UUID
}

// QueryHistory manages in-memory query history tracking per connection.
// It provides deduplication, LRU eviction, and thread-safe operations.
// Maximum entries per connection is configurable (default: 50).
type QueryHistory struct {
	mu               sync.RWMutex
	entries          map[uuid.UUID][]QueryHistoryEntry // connection ID → entries
	maxPerConnection int
}

// DefaultMaxHistoryEntries is the default maximum number of history entries per connection (50).
const DefaultMaxHistoryEntries = 50

// NewQueryHistory creates a new QueryHistory instance.
// If maxEntries is not provided or <= 0, the default (50) is used.
func NewQueryHistory(maxEntries ...int) *QueryHistory {
	max := DefaultMaxHistoryEntries
	if len(maxEntries) > 0 && maxEntries[0] > 0 {
		max = maxEntries[0]
	}

	return &QueryHistory{
		entries:          make(map[uuid.UUID][]QueryHistoryEntry),
		maxPerConnection: max,
	}
}

// normalizeQuery normalizes a query string for deduplication.
// It trims whitespace and converts to lowercase for comparison.
func normalizeQuery(query string) string {
	return strings.ToLower(strings.TrimSpace(query))
}

// AddQuery adds a query to the history for a connection.
// It deduplicates queries by normalizing whitespace and case.
// If the query already exists, it updates the timestamp instead of creating a duplicate.
// When the history limit is exceeded, the oldest entry is evicted (FIFO).
func (h *QueryHistory) AddQuery(connID uuid.UUID, query string, duration time.Duration, success bool, errMsg string, rowCount int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Get or create entries slice for this connection
	entries := h.entries[connID]
	if entries == nil {
		entries = make([]QueryHistoryEntry, 0)
	}

	// Normalize the query for deduplication
	normalizedQuery := normalizeQuery(query)

	// Check for duplicate
	foundIndex := -1
	for i, entry := range entries {
		if normalizeQuery(entry.Query) == normalizedQuery {
			foundIndex = i
			break
		}
	}

	now := time.Now()
	if foundIndex >= 0 {
		// Update existing entry with new timestamp and duration
		entries[foundIndex].Timestamp = now
		entries[foundIndex].Duration = duration
		entries[foundIndex].Success = success
		entries[foundIndex].Error = errMsg
		entries[foundIndex].RowCount = rowCount

		// Move to end (most recent)
		entry := entries[foundIndex]
		entries = append(entries[:foundIndex], entries[foundIndex+1:]...)
		entries = append(entries, entry)
	} else {
		// Create new entry
		entry := QueryHistoryEntry{
			ID:         uuid.New(),
			Query:      query,
			Timestamp:  now,
			Duration:   duration,
			Success:    success,
			Error:      errMsg,
			RowCount:   rowCount,
			Connection: connID,
		}
		entries = append(entries, entry)
	}

	// Evict oldest if limit exceeded
	if len(entries) > h.maxPerConnection {
		entries = entries[1:] // Remove oldest (first element)
	}

	h.entries[connID] = entries
}

// GetHistory returns the last N queries for a connection.
// Returns entries sorted by timestamp descending (newest first).
// If limit is 0 or negative, all entries are returned.
// Returns a copy to prevent external modification.
func (h *QueryHistory) GetHistory(connID uuid.UUID, limit int) []QueryHistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	entries, exists := h.entries[connID]
	if !exists {
		return []QueryHistoryEntry{}
	}

	// Return all if limit is 0 or negative
	if limit <= 0 {
		// Return a copy to prevent external modification
		result := make([]QueryHistoryEntry, len(entries))
		copy(result, entries)
		return result
	}

	// Return last N entries in reverse order (newest first)
	start := len(entries) - limit
	if start < 0 {
		start = 0
	}

	result := make([]QueryHistoryEntry, 0, limit)
	for i := len(entries) - 1; i >= start; i-- {
		result = append(result, entries[i])
	}

	return result
}

// GetHistoryCount returns the number of history entries for a connection.
// Returns 0 if connection has no history.
func (h *QueryHistory) GetHistoryCount(connID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	entries, exists := h.entries[connID]
	if !exists {
		return 0
	}

	return len(entries)
}

// ClearHistory clears all history entries for a connection.
// Does not affect other connections' history.
func (h *QueryHistory) ClearHistory(connID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.entries[connID]; exists {
		h.entries[connID] = make([]QueryHistoryEntry, 0)
	}
}
