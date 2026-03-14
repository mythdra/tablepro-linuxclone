package query

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestQueryHistory_NewQueryHistory(t *testing.T) {
	t.Run("default max entries", func(t *testing.T) {
		h := NewQueryHistory()
		assert.NotNil(t, h)
		assert.Equal(t, 50, h.maxPerConnection)
	})

	t.Run("custom max entries", func(t *testing.T) {
		h := NewQueryHistory(100)
		assert.NotNil(t, h)
		assert.Equal(t, 100, h.maxPerConnection)
	})
}

func TestQueryHistory_AddQuery(t *testing.T) {
	t.Run("adds query to history", func(t *testing.T) {
		h := NewQueryHistory()
		connID := uuid.New()
		query := "SELECT * FROM users"

		h.AddQuery(connID, query, 100*time.Millisecond, true, "", 10)

		entries := h.GetHistory(connID, 10)
		assert.Len(t, entries, 1)
		assert.Equal(t, query, entries[0].Query)
		assert.True(t, entries[0].Success)
		assert.Equal(t, int64(10), entries[0].RowCount)
		assert.Greater(t, entries[0].Duration, time.Duration(0))
	})

	t.Run("per connection isolation", func(t *testing.T) {
		h := NewQueryHistory()
		connID1 := uuid.New()
		connID2 := uuid.New()

		h.AddQuery(connID1, "SELECT * FROM users", 50*time.Millisecond, true, "", 5)
		h.AddQuery(connID2, "SELECT * FROM products", 75*time.Millisecond, true, "", 3)

		entries1 := h.GetHistory(connID1, 10)
		entries2 := h.GetHistory(connID2, 10)

		assert.Len(t, entries1, 1)
		assert.Len(t, entries2, 1)
		assert.Equal(t, "SELECT * FROM users", entries1[0].Query)
		assert.Equal(t, "SELECT * FROM products", entries2[0].Query)
	})

	t.Run("records failed queries", func(t *testing.T) {
		h := NewQueryHistory()
		connID := uuid.New()

		h.AddQuery(connID, "SELECT * FROM invalid", 10*time.Millisecond, false, "table not found", 0)

		entries := h.GetHistory(connID, 10)
		assert.Len(t, entries, 1)
		assert.False(t, entries[0].Success)
		assert.Equal(t, "table not found", entries[0].Error)
	})
}

func TestQueryHistory_Deduplication(t *testing.T) {
	t.Run("deduplicates identical queries", func(t *testing.T) {
		h := NewQueryHistory()
		connID := uuid.New()
		query := "SELECT * FROM users"

		h.AddQuery(connID, query, 50*time.Millisecond, true, "", 5)
		time.Sleep(10 * time.Millisecond)
		h.AddQuery(connID, query, 75*time.Millisecond, true, "", 5)

		entries := h.GetHistory(connID, 10)
		assert.Len(t, entries, 1)
		assert.Greater(t, entries[0].Timestamp.UnixNano(), int64(0))
	})

	t.Run("deduplicates queries with different whitespace", func(t *testing.T) {
		h := NewQueryHistory()
		connID := uuid.New()
		query1 := "SELECT * FROM users"
		query2 := "  SELECT * FROM users  "
		query3 := "\tSELECT * FROM users\n"

		h.AddQuery(connID, query1, 50*time.Millisecond, true, "", 5)
		h.AddQuery(connID, query2, 60*time.Millisecond, true, "", 5)
		h.AddQuery(connID, query3, 70*time.Millisecond, true, "", 5)

		entries := h.GetHistory(connID, 10)
		assert.Len(t, entries, 1)
		assert.Equal(t, query1, entries[0].Query)
	})

	t.Run("deduplicates queries with different case", func(t *testing.T) {
		h := NewQueryHistory()
		connID := uuid.New()
		query1 := "SELECT * FROM users"
		query2 := "select * from users"
		query3 := "SeLeCt * FrOm users"

		h.AddQuery(connID, query1, 50*time.Millisecond, true, "", 5)
		h.AddQuery(connID, query2, 60*time.Millisecond, true, "", 5)
		h.AddQuery(connID, query3, 70*time.Millisecond, true, "", 5)

		entries := h.GetHistory(connID, 10)
		assert.Len(t, entries, 1)
	})

	t.Run("normalizes whitespace and case together", func(t *testing.T) {
		h := NewQueryHistory()
		connID := uuid.New()
		query1 := "SELECT * FROM users WHERE id = 1"
		query2 := "  select * from users where id = 1  "

		h.AddQuery(connID, query1, 50*time.Millisecond, true, "", 5)
		h.AddQuery(connID, query2, 60*time.Millisecond, true, "", 5)

		entries := h.GetHistory(connID, 10)
		assert.Len(t, entries, 1)
	})

	t.Run("different queries are not deduplicated", func(t *testing.T) {
		h := NewQueryHistory()
		connID := uuid.New()

		h.AddQuery(connID, "SELECT * FROM users", 50*time.Millisecond, true, "", 5)
		h.AddQuery(connID, "SELECT * FROM products", 60*time.Millisecond, true, "", 3)

		entries := h.GetHistory(connID, 10)
		assert.Len(t, entries, 2)
	})
}

func TestQueryHistory_LRUEviction(t *testing.T) {
	t.Run("evicts oldest entry when limit exceeded", func(t *testing.T) {
		maxEntries := 5
		h := NewQueryHistory(maxEntries)
		connID := uuid.New()

		for i := 0; i < maxEntries+1; i++ {
			h.AddQuery(connID, fmt.Sprintf("Query %c", 'A'+i), time.Duration(i)*10*time.Millisecond, true, "", int64(i))
			time.Sleep(time.Millisecond)
		}

		entries := h.GetHistory(connID, 10)
		assert.Len(t, entries, maxEntries)
		// GetHistory returns newest first, so Query B (oldest remaining) should be last
		assert.Equal(t, "Query B", entries[4].Query)
		assert.Equal(t, "Query F", entries[0].Query)
	})

	t.Run("evicts in FIFO order", func(t *testing.T) {
		maxEntries := 3
		h := NewQueryHistory(maxEntries)
		connID := uuid.New()

		h.AddQuery(connID, "Query 1", 10*time.Millisecond, true, "", 1)
		time.Sleep(time.Millisecond)
		h.AddQuery(connID, "Query 2", 20*time.Millisecond, true, "", 2)
		time.Sleep(time.Millisecond)
		h.AddQuery(connID, "Query 3", 30*time.Millisecond, true, "", 3)
		time.Sleep(time.Millisecond)
		h.AddQuery(connID, "Query 4", 40*time.Millisecond, true, "", 4)

		entries := h.GetHistory(connID, 10)
		assert.Len(t, entries, 3)
		// Newest first: Query 4, Query 3, Query 2
		assert.Equal(t, "Query 2", entries[2].Query)
		assert.Equal(t, "Query 3", entries[1].Query)
		assert.Equal(t, "Query 4", entries[0].Query)
	})

	t.Run("eviction respects per connection isolation", func(t *testing.T) {
		maxEntries := 2
		h := NewQueryHistory(maxEntries)
		connID1 := uuid.New()
		connID2 := uuid.New()

		h.AddQuery(connID1, "Conn1 Query 1", 10*time.Millisecond, true, "", 1)
		h.AddQuery(connID1, "Conn1 Query 2", 20*time.Millisecond, true, "", 2)
		h.AddQuery(connID1, "Conn1 Query 3", 30*time.Millisecond, true, "", 3)

		h.AddQuery(connID2, "Conn2 Query 1", 10*time.Millisecond, true, "", 1)
		h.AddQuery(connID2, "Conn2 Query 2", 20*time.Millisecond, true, "", 2)

		entries1 := h.GetHistory(connID1, 10)
		entries2 := h.GetHistory(connID2, 10)

		assert.Len(t, entries1, 2)
		assert.Len(t, entries2, 2)
		// Newest first
		assert.Equal(t, "Conn1 Query 2", entries1[1].Query)
		assert.Equal(t, "Conn1 Query 3", entries1[0].Query)
		assert.Equal(t, "Conn2 Query 1", entries2[1].Query)
		assert.Equal(t, "Conn2 Query 2", entries2[0].Query)
	})

	t.Run("deduplicated query does not cause eviction", func(t *testing.T) {
		maxEntries := 2
		h := NewQueryHistory(maxEntries)
		connID := uuid.New()

		h.AddQuery(connID, "SELECT 1", 10*time.Millisecond, true, "", 1)
		h.AddQuery(connID, "SELECT 2", 20*time.Millisecond, true, "", 2)
		time.Sleep(time.Millisecond)
		h.AddQuery(connID, "SELECT 1", 30*time.Millisecond, true, "", 1)

		entries := h.GetHistory(connID, 10)
		assert.Len(t, entries, 2)
		// SELECT 1 was deduplicated and moved to end (newest)
		assert.Equal(t, "SELECT 1", entries[0].Query)
		assert.Equal(t, "SELECT 2", entries[1].Query)
	})
}

func TestQueryHistory_GetHistory(t *testing.T) {
	t.Run("returns entries sorted by timestamp descending", func(t *testing.T) {
		h := NewQueryHistory()
		connID := uuid.New()

		h.AddQuery(connID, "Query 1", 10*time.Millisecond, true, "", 1)
		time.Sleep(time.Millisecond)
		h.AddQuery(connID, "Query 2", 20*time.Millisecond, true, "", 2)
		time.Sleep(time.Millisecond)
		h.AddQuery(connID, "Query 3", 30*time.Millisecond, true, "", 3)

		entries := h.GetHistory(connID, 10)
		assert.Len(t, entries, 3)
		assert.Equal(t, "Query 1", entries[2].Query)
		assert.Equal(t, "Query 2", entries[1].Query)
		assert.Equal(t, "Query 3", entries[0].Query)
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		h := NewQueryHistory()
		connID := uuid.New()

		for i := 1; i <= 10; i++ {
			h.AddQuery(connID, fmt.Sprintf("Query %d", i), time.Duration(i)*10*time.Millisecond, true, "", int64(i))
			time.Sleep(time.Millisecond)
		}

		entries := h.GetHistory(connID, 5)
		assert.Len(t, entries, 5)
		assert.Equal(t, "Query 6", entries[4].Query)
		assert.Equal(t, "Query 10", entries[0].Query)
	})

	t.Run("returns all entries when limit is 0", func(t *testing.T) {
		h := NewQueryHistory(100)
		connID := uuid.New()

		for i := 1; i <= 10; i++ {
			h.AddQuery(connID, fmt.Sprintf("Query %d", i), time.Duration(i)*10*time.Millisecond, true, "", int64(i))
			time.Sleep(time.Millisecond)
		}

		entries := h.GetHistory(connID, 0)
		assert.Len(t, entries, 10)
	})

	t.Run("returns all entries when limit exceeds history size", func(t *testing.T) {
		h := NewQueryHistory()
		connID := uuid.New()

		h.AddQuery(connID, "Query 1", 10*time.Millisecond, true, "", 1)
		h.AddQuery(connID, "Query 2", 20*time.Millisecond, true, "", 2)

		entries := h.GetHistory(connID, 100)
		assert.Len(t, entries, 2)
	})

	t.Run("returns empty slice for unknown connection", func(t *testing.T) {
		h := NewQueryHistory()
		connID := uuid.New()

		entries := h.GetHistory(connID, 10)
		assert.Empty(t, entries)
	})
}

func TestQueryHistory_ClearHistory(t *testing.T) {
	t.Run("clears history for connection", func(t *testing.T) {
		h := NewQueryHistory()
		connID := uuid.New()

		h.AddQuery(connID, "Query 1", 10*time.Millisecond, true, "", 1)
		h.AddQuery(connID, "Query 2", 20*time.Millisecond, true, "", 2)

		h.ClearHistory(connID)

		entries := h.GetHistory(connID, 10)
		assert.Empty(t, entries)
	})

	t.Run("clear only affects specified connection", func(t *testing.T) {
		h := NewQueryHistory()
		connID1 := uuid.New()
		connID2 := uuid.New()

		h.AddQuery(connID1, "Conn1 Query", 10*time.Millisecond, true, "", 1)
		h.AddQuery(connID2, "Conn2 Query", 20*time.Millisecond, true, "", 2)

		h.ClearHistory(connID1)

		entries1 := h.GetHistory(connID1, 10)
		entries2 := h.GetHistory(connID2, 10)

		assert.Empty(t, entries1)
		assert.Len(t, entries2, 1)
	})
}

func TestQueryHistory_GetHistoryCount(t *testing.T) {
	t.Run("returns correct count", func(t *testing.T) {
		h := NewQueryHistory()
		connID := uuid.New()

		h.AddQuery(connID, "Query 1", 10*time.Millisecond, true, "", 1)
		h.AddQuery(connID, "Query 2", 20*time.Millisecond, true, "", 2)
		h.AddQuery(connID, "Query 3", 30*time.Millisecond, true, "", 3)

		count := h.GetHistoryCount(connID)
		assert.Equal(t, 3, count)
	})

	t.Run("returns 0 for unknown connection", func(t *testing.T) {
		h := NewQueryHistory()
		connID := uuid.New()

		count := h.GetHistoryCount(connID)
		assert.Equal(t, 0, count)
	})
}

func TestQueryHistory_ThreadSafety(t *testing.T) {
	t.Run("concurrent adds are thread-safe", func(t *testing.T) {
		h := NewQueryHistory(100)
		connID := uuid.New()

		done := make(chan bool)
		for i := 0; i < 50; i++ {
			go func(id int) {
				h.AddQuery(connID, "Query "+string(rune('A'+id)), time.Duration(id)*10*time.Millisecond, true, "", int64(id))
				done <- true
			}(i)
		}

		for i := 0; i < 50; i++ {
			<-done
		}

		entries := h.GetHistory(connID, 100)
		assert.NotEmpty(t, entries)
	})

	t.Run("concurrent reads and writes are thread-safe", func(t *testing.T) {
		h := NewQueryHistory(100)
		connID := uuid.New()

		done := make(chan bool)

		for i := 0; i < 25; i++ {
			go func(id int) {
				h.AddQuery(connID, "Query "+string(rune('A'+id)), time.Duration(id)*10*time.Millisecond, true, "", int64(id))
				done <- true
			}(i)
		}

		for i := 0; i < 25; i++ {
			go func() {
				h.GetHistory(connID, 10)
				done <- true
			}()
		}

		for i := 0; i < 50; i++ {
			<-done
		}

		entries := h.GetHistory(connID, 100)
		assert.NotEmpty(t, entries)
	})
}
