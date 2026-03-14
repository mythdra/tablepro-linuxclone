package change

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

const undoStackLimit = 100

type DataChangeManager struct {
	mu   sync.RWMutex
	tabs map[string]*TabChanges
}

func NewDataChangeManager() *DataChangeManager {
	return &DataChangeManager{
		tabs: make(map[string]*TabChanges),
	}
}

func (m *DataChangeManager) InitTab(tabID, tableName, schemaName string, primaryKeys []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tabs[tabID] = &TabChanges{
		TabID:        tabID,
		TableName:    tableName,
		SchemaName:   schemaName,
		PrimaryKeys:  primaryKeys,
		CellChanges:  make([]*CellChange, 0),
		InsertedRows: make([]*InsertedRow, 0),
		DeletedRows:  make([]*DeletedRow, 0),
		UndoStack:    make([]ChangeAction, 0),
		RedoStack:    make([]ChangeAction, 0),
	}
}

type updateUndoInfo struct {
	Change    CellChange
	WasNew    bool
	OldNewVal any
}

func (m *DataChangeManager) UpdateCell(tabID string, rowIndex int, column string, originalValue, newValue any, primaryKey map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tab, ok := m.tabs[tabID]
	if !ok {
		return fmt.Errorf("tab not initialized: %s", tabID)
	}

	for i, cc := range tab.CellChanges {
		if cc.RowIndex == rowIndex && cc.Column == column && primaryKeyEquals(cc.PrimaryKey, primaryKey) {
			oldNewVal := cc.NewValue
			if valuesEqual(cc.OriginalValue, newValue) {
				tab.CellChanges = append(tab.CellChanges[:i], tab.CellChanges[i+1:]...)
			} else {
				cc.NewValue = newValue
			}
			changeCopy := *cc
			changeCopy.NewValue = newValue
			m.pushUndo(tab, ChangeAction{Type: "update", Payload: &updateUndoInfo{
				Change:    changeCopy,
				WasNew:    false,
				OldNewVal: oldNewVal,
			}})
			return nil
		}
	}

	cc := &CellChange{
		RowIndex:      rowIndex,
		Column:        column,
		OriginalValue: originalValue,
		NewValue:      newValue,
		PrimaryKey:    primaryKey,
	}
	tab.CellChanges = append(tab.CellChanges, cc)

	m.pushUndo(tab, ChangeAction{Type: "update", Payload: &updateUndoInfo{
		Change: *cc,
		WasNew: true,
	}})
	return nil
}

func (m *DataChangeManager) InsertRow(tabID string, data map[string]any) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	tab, ok := m.tabs[tabID]
	if !ok {
		return "", fmt.Errorf("tab not initialized: %s", tabID)
	}

	tempID := uuid.New().String()
	row := &InsertedRow{
		TempID: tempID,
		Data:   data,
	}
	tab.InsertedRows = append(tab.InsertedRows, row)

	m.pushUndo(tab, ChangeAction{Type: "insert", Payload: row})
	return tempID, nil
}

func (m *DataChangeManager) DeleteRow(tabID string, primaryKey map[string]any, rowIndex int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tab, ok := m.tabs[tabID]
	if !ok {
		return fmt.Errorf("tab not initialized: %s", tabID)
	}

	row := &DeletedRow{
		PrimaryKey: primaryKey,
		RowIndex:   rowIndex,
	}
	tab.DeletedRows = append(tab.DeletedRows, row)

	m.pushUndo(tab, ChangeAction{Type: "delete", Payload: row})
	return nil
}

func (m *DataChangeManager) GetPendingChanges(tabID string) (*PendingChanges, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tab, ok := m.tabs[tabID]
	if !ok {
		return nil, fmt.Errorf("tab not initialized: %s", tabID)
	}

	return &PendingChanges{
		CellChanges:  tab.CellChanges,
		InsertedRows: tab.InsertedRows,
		DeletedRows:  tab.DeletedRows,
		Summary: ChangeSummary{
			Updates: len(tab.CellChanges),
			Inserts: len(tab.InsertedRows),
			Deletes: len(tab.DeletedRows),
		},
	}, nil
}

func (m *DataChangeManager) DiscardChanges(tabID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tab, ok := m.tabs[tabID]
	if !ok {
		return fmt.Errorf("tab not initialized: %s", tabID)
	}

	tab.CellChanges = make([]*CellChange, 0)
	tab.InsertedRows = make([]*InsertedRow, 0)
	tab.DeletedRows = make([]*DeletedRow, 0)
	tab.UndoStack = make([]ChangeAction, 0)
	tab.RedoStack = make([]ChangeAction, 0)
	return nil
}

func (m *DataChangeManager) HasChanges(tabID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tab, ok := m.tabs[tabID]
	if !ok {
		return false
	}

	return len(tab.CellChanges) > 0 || len(tab.InsertedRows) > 0 || len(tab.DeletedRows) > 0
}

func (m *DataChangeManager) GetChangeCount(tabID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tab, ok := m.tabs[tabID]
	if !ok {
		return 0
	}

	return len(tab.CellChanges) + len(tab.InsertedRows) + len(tab.DeletedRows)
}

func (m *DataChangeManager) Undo(tabID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tab, ok := m.tabs[tabID]
	if !ok {
		return fmt.Errorf("tab not initialized: %s", tabID)
	}

	if len(tab.UndoStack) == 0 {
		return fmt.Errorf("nothing to undo")
	}

	action := tab.UndoStack[len(tab.UndoStack)-1]
	tab.UndoStack = tab.UndoStack[:len(tab.UndoStack)-1]

	m.reverseAction(tab, action)
	tab.RedoStack = append(tab.RedoStack, action)
	return nil
}

func (m *DataChangeManager) Redo(tabID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tab, ok := m.tabs[tabID]
	if !ok {
		return fmt.Errorf("tab not initialized: %s", tabID)
	}

	if len(tab.RedoStack) == 0 {
		return fmt.Errorf("nothing to redo")
	}

	action := tab.RedoStack[len(tab.RedoStack)-1]
	tab.RedoStack = tab.RedoStack[:len(tab.RedoStack)-1]

	m.applyAction(tab, action)
	tab.UndoStack = append(tab.UndoStack, action)
	return nil
}

func (m *DataChangeManager) pushUndo(tab *TabChanges, action ChangeAction) {
	tab.UndoStack = append(tab.UndoStack, action)
	if len(tab.UndoStack) > undoStackLimit {
		tab.UndoStack = tab.UndoStack[len(tab.UndoStack)-undoStackLimit:]
	}
	tab.RedoStack = make([]ChangeAction, 0)
}

func (m *DataChangeManager) reverseAction(tab *TabChanges, action ChangeAction) {
	switch action.Type {
	case "update":
		info, ok := action.Payload.(*updateUndoInfo)
		if !ok {
			return
		}
		if info.WasNew {
			for i, existing := range tab.CellChanges {
				if existing.RowIndex == info.Change.RowIndex && existing.Column == info.Change.Column && primaryKeyEquals(existing.PrimaryKey, info.Change.PrimaryKey) {
					tab.CellChanges = append(tab.CellChanges[:i], tab.CellChanges[i+1:]...)
					return
				}
			}
		} else {
			if valuesEqual(info.OldNewVal, info.Change.OriginalValue) {
				for i, existing := range tab.CellChanges {
					if existing.RowIndex == info.Change.RowIndex && existing.Column == info.Change.Column && primaryKeyEquals(existing.PrimaryKey, info.Change.PrimaryKey) {
						tab.CellChanges = append(tab.CellChanges[:i], tab.CellChanges[i+1:]...)
						return
					}
				}
			} else {
				for _, existing := range tab.CellChanges {
					if existing.RowIndex == info.Change.RowIndex && existing.Column == info.Change.Column && primaryKeyEquals(existing.PrimaryKey, info.Change.PrimaryKey) {
						existing.NewValue = info.OldNewVal
						return
					}
				}
				tab.CellChanges = append(tab.CellChanges, &CellChange{
					RowIndex:      info.Change.RowIndex,
					Column:        info.Change.Column,
					OriginalValue: info.Change.OriginalValue,
					NewValue:      info.OldNewVal,
					PrimaryKey:    info.Change.PrimaryKey,
				})
			}
		}
	case "insert":
		row, ok := action.Payload.(*InsertedRow)
		if !ok {
			return
		}
		for i, existing := range tab.InsertedRows {
			if existing.TempID == row.TempID {
				tab.InsertedRows = append(tab.InsertedRows[:i], tab.InsertedRows[i+1:]...)
				return
			}
		}
	case "delete":
		row, ok := action.Payload.(*DeletedRow)
		if !ok {
			return
		}
		for i, existing := range tab.DeletedRows {
			if primaryKeyEquals(existing.PrimaryKey, row.PrimaryKey) {
				tab.DeletedRows = append(tab.DeletedRows[:i], tab.DeletedRows[i+1:]...)
				return
			}
		}
	}
}

func (m *DataChangeManager) applyAction(tab *TabChanges, action ChangeAction) {
	switch action.Type {
	case "update":
		info, ok := action.Payload.(*updateUndoInfo)
		if !ok {
			return
		}
		for _, existing := range tab.CellChanges {
			if existing.RowIndex == info.Change.RowIndex && existing.Column == info.Change.Column && primaryKeyEquals(existing.PrimaryKey, info.Change.PrimaryKey) {
				existing.NewValue = info.Change.NewValue
				return
			}
		}
		tab.CellChanges = append(tab.CellChanges, &CellChange{
			RowIndex:      info.Change.RowIndex,
			Column:        info.Change.Column,
			OriginalValue: info.Change.OriginalValue,
			NewValue:      info.Change.NewValue,
			PrimaryKey:    info.Change.PrimaryKey,
		})
	case "insert":
		row, ok := action.Payload.(*InsertedRow)
		if !ok {
			return
		}
		tab.InsertedRows = append(tab.InsertedRows, row)
	case "delete":
		row, ok := action.Payload.(*DeletedRow)
		if !ok {
			return
		}
		tab.DeletedRows = append(tab.DeletedRows, row)
	}
}

func primaryKeyEquals(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || !valuesEqual(v, bv) {
			return false
		}
	}
	return true
}

func valuesEqual(a, b any) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}
