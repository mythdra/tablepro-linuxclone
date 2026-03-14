package change

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tablepro/internal/driver"
)

func setupManager(t *testing.T) *DataChangeManager {
	t.Helper()
	m := NewDataChangeManager()
	columns := []driver.ColumnInfo{
		{Name: "id", IsPrimaryKey: true, IsAutoIncrement: true},
		{Name: "name"},
		{Name: "email"},
		{Name: "age"},
	}
	m.EnsureTab("tab-a", "users", "", driver.DatabaseTypeSQLite, []string{"id"}, columns)
	return m
}

func TestManager_ConcurrentAccess(t *testing.T) {
	m := setupManager(t)

	var wg sync.WaitGroup
	errChan := make(chan error, 200)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			err := m.TrackCellChange("tab-a", idx, "name", fmt.Sprintf("old-%d", idx), fmt.Sprintf("new-%d", idx), "text")
			if err != nil {
				errChan <- err
			}
		}(i)
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.GetSummary()
		}()
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.GetChangeCount("tab-a")
		}()
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Fatalf("concurrent access error: %v", err)
	}

	assert.Equal(t, 100, m.GetChangeCount("tab-a"))
}

func TestManager_MultipleTabsIsolation(t *testing.T) {
	m := NewDataChangeManager()
	columns := []driver.ColumnInfo{
		{Name: "id", IsPrimaryKey: true},
		{Name: "name"},
	}

	m.EnsureTab("tab-a", "users", "", driver.DatabaseTypeSQLite, []string{"id"}, columns)
	m.EnsureTab("tab-b", "products", "", driver.DatabaseTypeSQLite, []string{"id"}, columns)

	err := m.TrackCellChange("tab-a", 0, "name", "old-a", "new-a", "text")
	require.NoError(t, err)

	err = m.TrackCellChange("tab-a", 1, "name", "old-a2", "new-a2", "text")
	require.NoError(t, err)

	err = m.TrackCellChange("tab-b", 0, "name", "old-b", "new-b", "text")
	require.NoError(t, err)

	assert.Equal(t, 2, m.GetChangeCount("tab-a"))
	assert.Equal(t, 1, m.GetChangeCount("tab-b"))

	tabA := m.GetChanges("tab-a")
	tabB := m.GetChanges("tab-b")
	assert.Equal(t, "users", tabA.TableName)
	assert.Equal(t, "products", tabB.TableName)

	m.DiscardAll("tab-a")
	assert.Equal(t, 0, m.GetChangeCount("tab-a"))
	assert.Equal(t, 1, m.GetChangeCount("tab-b"))

	summary := m.GetSummary()
	assert.Equal(t, 1, summary.TotalChanges)
	assert.Equal(t, 1, summary.AffectedTabs)
}

func TestManager_UndoStackLimit(t *testing.T) {
	m := setupManager(t)

	for i := 0; i < MaxUndoStackSize+50; i++ {
		err := m.TrackCellChange("tab-a", i, "name", fmt.Sprintf("old-%d", i), fmt.Sprintf("new-%d", i), "text")
		require.NoError(t, err)
	}

	assert.Equal(t, MaxUndoStackSize, m.GetUndoStackSize())
}

func TestManager_UndoAfterDiscard(t *testing.T) {
	m := setupManager(t)

	err := m.TrackCellChange("tab-a", 0, "name", "old", "new", "text")
	require.NoError(t, err)

	m.DiscardAll("tab-a")

	assert.Equal(t, 0, m.GetUndoStackSize())

	entry, err := m.Undo()
	require.NoError(t, err)
	assert.Nil(t, entry)
}

func TestManager_InsertThenDelete(t *testing.T) {
	m := setupManager(t)

	row := InsertedRow{
		TempID: "tmp-1",
		Values: map[string]any{"name": "New User", "email": "new@test.com"},
	}
	err := m.TrackInsert("tab-a", row)
	require.NoError(t, err)
	assert.Equal(t, 1, m.GetChangeCount("tab-a"))

	entry, err := m.Undo()
	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, ChangeActionInsert, entry.Action)

	assert.Equal(t, 0, m.GetChangeCount("tab-a"))
	assert.False(t, m.HasChanges("tab-a"))
}

func TestManager_GetChangeCount(t *testing.T) {
	m := setupManager(t)

	assert.Equal(t, 0, m.GetChangeCount("tab-a"))
	assert.Equal(t, 0, m.GetChangeCount("nonexistent"))

	err := m.TrackCellChange("tab-a", 0, "name", "old", "new", "text")
	require.NoError(t, err)
	assert.Equal(t, 1, m.GetChangeCount("tab-a"))

	err = m.TrackCellChange("tab-a", 1, "email", "old@t.com", "new@t.com", "text")
	require.NoError(t, err)
	assert.Equal(t, 2, m.GetChangeCount("tab-a"))

	err = m.TrackInsert("tab-a", InsertedRow{TempID: "tmp-1", Values: map[string]any{"name": "X"}})
	require.NoError(t, err)
	assert.Equal(t, 3, m.GetChangeCount("tab-a"))

	err = m.TrackDelete("tab-a", DeletedRow{RowIndex: 5, PrimaryKeys: map[string]any{"id": 5}})
	require.NoError(t, err)
	assert.Equal(t, 4, m.GetChangeCount("tab-a"))

	summary := m.GetSummary()
	assert.Equal(t, 2, summary.TotalUpdates)
	assert.Equal(t, 1, summary.TotalInserts)
	assert.Equal(t, 1, summary.TotalDeletes)
	assert.Equal(t, 4, summary.TotalChanges)
	assert.Equal(t, 1, summary.AffectedTabs)
}

func TestManager_UndoRedoFullCycle(t *testing.T) {
	m := setupManager(t)

	err := m.TrackCellChange("tab-a", 0, "name", "Alice", "Bob", "text")
	require.NoError(t, err)

	err = m.TrackCellChange("tab-a", 1, "name", "Charlie", "Dave", "text")
	require.NoError(t, err)

	assert.Equal(t, 2, m.GetUndoStackSize())
	assert.Equal(t, 0, m.GetRedoStackSize())

	entry, err := m.Undo()
	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, "Charlie", entry.OldValue)
	assert.Equal(t, 1, m.GetUndoStackSize())
	assert.Equal(t, 1, m.GetRedoStackSize())

	redoEntry, err := m.Redo()
	require.NoError(t, err)
	require.NotNil(t, redoEntry)
	assert.Equal(t, "Dave", redoEntry.NewValue)
	assert.Equal(t, 2, m.GetUndoStackSize())
	assert.Equal(t, 0, m.GetRedoStackSize())

	assert.Equal(t, 2, m.GetChangeCount("tab-a"))
}

func TestManager_TrackChangeToNonexistentTab(t *testing.T) {
	m := NewDataChangeManager()

	err := m.TrackCellChange("ghost-tab", 0, "name", "a", "b", "text")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ghost-tab")
}

func TestManager_EnsureTabIdempotent(t *testing.T) {
	m := NewDataChangeManager()
	columns := []driver.ColumnInfo{{Name: "id", IsPrimaryKey: true}}

	m.EnsureTab("tab-x", "users", "", driver.DatabaseTypeSQLite, []string{"id"}, columns)

	err := m.TrackCellChange("tab-x", 0, "id", 1, 2, "integer")
	require.NoError(t, err)

	m.EnsureTab("tab-x", "users", "", driver.DatabaseTypeSQLite, []string{"id"}, columns)

	assert.Equal(t, 1, m.GetChangeCount("tab-x"))
}

func TestManager_DiscardAllGlobal(t *testing.T) {
	m := NewDataChangeManager()
	columns := []driver.ColumnInfo{
		{Name: "id", IsPrimaryKey: true},
		{Name: "name"},
	}

	m.EnsureTab("tab-1", "t1", "", driver.DatabaseTypeSQLite, []string{"id"}, columns)
	m.EnsureTab("tab-2", "t2", "", driver.DatabaseTypeSQLite, []string{"id"}, columns)

	_ = m.TrackCellChange("tab-1", 0, "name", "a", "b", "text")
	_ = m.TrackCellChange("tab-2", 0, "name", "c", "d", "text")

	assert.Equal(t, 2, m.GetSummary().TotalChanges)

	m.DiscardAll("")

	assert.Equal(t, 0, m.GetSummary().TotalChanges)
	assert.Equal(t, 0, m.GetUndoStackSize())
	assert.Equal(t, 0, m.GetRedoStackSize())
}

func TestManager_NewChangeClearsRedoStack(t *testing.T) {
	m := setupManager(t)

	_ = m.TrackCellChange("tab-a", 0, "name", "a", "b", "text")
	_ = m.TrackCellChange("tab-a", 1, "name", "c", "d", "text")

	_, _ = m.Undo()
	assert.Equal(t, 1, m.GetRedoStackSize())

	_ = m.TrackCellChange("tab-a", 2, "name", "e", "f", "text")
	assert.Equal(t, 0, m.GetRedoStackSize())
}

func TestManager_HasChanges(t *testing.T) {
	m := setupManager(t)

	assert.False(t, m.HasChanges("tab-a"))
	assert.False(t, m.HasChanges("nonexistent"))

	_ = m.TrackCellChange("tab-a", 0, "name", "a", "b", "text")
	assert.True(t, m.HasChanges("tab-a"))
}

func TestManager_GetAllChanges(t *testing.T) {
	m := NewDataChangeManager()
	columns := []driver.ColumnInfo{{Name: "id", IsPrimaryKey: true}, {Name: "name"}}

	m.EnsureTab("t1", "users", "", driver.DatabaseTypeSQLite, []string{"id"}, columns)
	m.EnsureTab("t2", "products", "", driver.DatabaseTypeMySQL, []string{"id"}, columns)

	_ = m.TrackCellChange("t1", 0, "name", "a", "b", "text")
	_ = m.TrackCellChange("t2", 0, "name", "c", "d", "text")

	all := m.GetAllChanges()
	require.NotNil(t, all)
	assert.Len(t, all.Tabs, 2)
	assert.Contains(t, all.Tabs, "t1")
	assert.Contains(t, all.Tabs, "t2")
}
