package change

import (
	"fmt"
	"sync"

	"tablepro/internal/driver"
)

const MaxUndoStackSize = 100

type DataChangeManager struct {
	mu        sync.RWMutex
	pending   *PendingChanges
	undoStack []UndoEntry
	redoStack []UndoEntry
}

func NewDataChangeManager() *DataChangeManager {
	return &DataChangeManager{
		pending:   NewPendingChanges(),
		undoStack: make([]UndoEntry, 0),
		redoStack: make([]UndoEntry, 0),
	}
}

func (m *DataChangeManager) EnsureTab(tabID, tableName, schemaName string, dbType driver.DatabaseType, primaryKeys []string, columns []driver.ColumnInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.pending.Tabs[tabID]; !exists {
		m.pending.Tabs[tabID] = NewTabChanges(tabID, tableName, schemaName, dbType, primaryKeys, columns)
	}
}

func (m *DataChangeManager) TrackCellChange(tabID string, rowIndex int, columnName string, oldValue, newValue any, dataType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tab, exists := m.pending.Tabs[tabID]
	if !exists {
		return fmt.Errorf("tab %s not found", tabID)
	}

	key := fmt.Sprintf("%d:%s", rowIndex, columnName)

	change := CellChange{
		RowIndex:   rowIndex,
		ColumnName: columnName,
		OldValue:   oldValue,
		NewValue:   newValue,
		DataType:   dataType,
	}

	tab.CellChanges[key] = []CellChange{change}

	m.pushUndo(UndoEntry{
		Action:     ChangeActionUpdate,
		TabID:      tabID,
		RowIndex:   rowIndex,
		ColumnName: columnName,
		OldValue:   oldValue,
		NewValue:   newValue,
	})

	m.redoStack = m.redoStack[:0]

	return nil
}

func (m *DataChangeManager) TrackInsert(tabID string, row InsertedRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tab, exists := m.pending.Tabs[tabID]
	if !exists {
		return fmt.Errorf("tab %s not found", tabID)
	}

	tab.InsertedRows = append(tab.InsertedRows, row)

	m.pushUndo(UndoEntry{
		Action:      ChangeActionInsert,
		TabID:       tabID,
		InsertedRow: &row,
	})

	m.redoStack = m.redoStack[:0]

	return nil
}

func (m *DataChangeManager) TrackDelete(tabID string, row DeletedRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tab, exists := m.pending.Tabs[tabID]
	if !exists {
		return fmt.Errorf("tab %s not found", tabID)
	}

	for i, inserted := range tab.InsertedRows {
		if inserted.TempID != "" && row.RowIndex == -1 {
			continue
		}
		_ = i
	}

	tab.DeletedRows = append(tab.DeletedRows, row)

	m.pushUndo(UndoEntry{
		Action:     ChangeActionDelete,
		TabID:      tabID,
		DeletedRow: &row,
	})

	m.redoStack = m.redoStack[:0]

	return nil
}

func (m *DataChangeManager) Undo() (*UndoEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.undoStack) == 0 {
		return nil, nil
	}

	entry := m.undoStack[len(m.undoStack)-1]
	m.undoStack = m.undoStack[:len(m.undoStack)-1]

	tab, exists := m.pending.Tabs[entry.TabID]
	if !exists {
		return nil, fmt.Errorf("tab %s not found for undo", entry.TabID)
	}

	switch entry.Action {
	case ChangeActionUpdate:
		key := fmt.Sprintf("%d:%s", entry.RowIndex, entry.ColumnName)
		delete(tab.CellChanges, key)

	case ChangeActionInsert:
		if entry.InsertedRow != nil && len(tab.InsertedRows) > 0 {
			tab.InsertedRows = tab.InsertedRows[:len(tab.InsertedRows)-1]
		}

	case ChangeActionDelete:
		if len(tab.DeletedRows) > 0 {
			tab.DeletedRows = tab.DeletedRows[:len(tab.DeletedRows)-1]
		}
	}

	m.redoStack = append(m.redoStack, entry)

	return &entry, nil
}

func (m *DataChangeManager) Redo() (*UndoEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.redoStack) == 0 {
		return nil, nil
	}

	entry := m.redoStack[len(m.redoStack)-1]
	m.redoStack = m.redoStack[:len(m.redoStack)-1]

	tab, exists := m.pending.Tabs[entry.TabID]
	if !exists {
		return nil, fmt.Errorf("tab %s not found for redo", entry.TabID)
	}

	switch entry.Action {
	case ChangeActionUpdate:
		key := fmt.Sprintf("%d:%s", entry.RowIndex, entry.ColumnName)
		tab.CellChanges[key] = []CellChange{{
			RowIndex:   entry.RowIndex,
			ColumnName: entry.ColumnName,
			OldValue:   entry.OldValue,
			NewValue:   entry.NewValue,
		}}

	case ChangeActionInsert:
		if entry.InsertedRow != nil {
			tab.InsertedRows = append(tab.InsertedRows, *entry.InsertedRow)
		}

	case ChangeActionDelete:
		if entry.DeletedRow != nil {
			tab.DeletedRows = append(tab.DeletedRows, *entry.DeletedRow)
		}
	}

	m.undoStack = append(m.undoStack, entry)

	return &entry, nil
}

func (m *DataChangeManager) DiscardAll(tabID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tabID == "" {
		m.pending = NewPendingChanges()
		m.undoStack = m.undoStack[:0]
		m.redoStack = m.redoStack[:0]
		return
	}

	delete(m.pending.Tabs, tabID)

	var newUndo []UndoEntry
	for _, e := range m.undoStack {
		if e.TabID != tabID {
			newUndo = append(newUndo, e)
		}
	}
	m.undoStack = newUndo

	var newRedo []UndoEntry
	for _, e := range m.redoStack {
		if e.TabID != tabID {
			newRedo = append(newRedo, e)
		}
	}
	m.redoStack = newRedo
}

func (m *DataChangeManager) GetChanges(tabID string) *TabChanges {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pending.Tabs[tabID]
}

func (m *DataChangeManager) GetAllChanges() *PendingChanges {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pending
}

func (m *DataChangeManager) GetSummary() ChangeSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := ChangeSummary{}
	for _, tab := range m.pending.Tabs {
		updates := len(tab.CellChanges)
		inserts := len(tab.InsertedRows)
		deletes := len(tab.DeletedRows)

		if updates+inserts+deletes > 0 {
			summary.AffectedTabs++
		}

		summary.TotalUpdates += updates
		summary.TotalInserts += inserts
		summary.TotalDeletes += deletes
	}
	summary.TotalChanges = summary.TotalUpdates + summary.TotalInserts + summary.TotalDeletes
	return summary
}

func (m *DataChangeManager) GetChangeCount(tabID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tab, exists := m.pending.Tabs[tabID]
	if !exists {
		return 0
	}
	return len(tab.CellChanges) + len(tab.InsertedRows) + len(tab.DeletedRows)
}

func (m *DataChangeManager) HasChanges(tabID string) bool {
	return m.GetChangeCount(tabID) > 0
}

func (m *DataChangeManager) GetUndoStackSize() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.undoStack)
}

func (m *DataChangeManager) GetRedoStackSize() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.redoStack)
}

func (m *DataChangeManager) pushUndo(entry UndoEntry) {
	m.undoStack = append(m.undoStack, entry)
	if len(m.undoStack) > MaxUndoStackSize {
		m.undoStack = m.undoStack[len(m.undoStack)-MaxUndoStackSize:]
	}
}
