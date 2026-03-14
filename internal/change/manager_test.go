package change

import (
	"fmt"
	"testing"
)

func TestUpdateCell(t *testing.T) {
	m := NewDataChangeManager()
	m.InitTab("tab1", "users", "public", []string{"id"})

	pk := map[string]any{"id": 1}
	err := m.UpdateCell("tab1", 0, "name", "old", "new", pk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !m.HasChanges("tab1") {
		t.Fatal("expected changes")
	}
	if m.GetChangeCount("tab1") != 1 {
		t.Fatalf("expected 1 change, got %d", m.GetChangeCount("tab1"))
	}

	pending, err := m.GetPendingChanges("tab1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pending.CellChanges) != 1 {
		t.Fatalf("expected 1 cell change, got %d", len(pending.CellChanges))
	}
	if pending.CellChanges[0].Column != "name" {
		t.Errorf("expected column 'name', got %q", pending.CellChanges[0].Column)
	}
	if pending.CellChanges[0].OriginalValue != "old" {
		t.Errorf("expected original 'old', got %v", pending.CellChanges[0].OriginalValue)
	}
	if pending.CellChanges[0].NewValue != "new" {
		t.Errorf("expected new 'new', got %v", pending.CellChanges[0].NewValue)
	}
	if pending.Summary.Updates != 1 {
		t.Errorf("expected 1 update in summary, got %d", pending.Summary.Updates)
	}
}

func TestUpdateCell_SameCell(t *testing.T) {
	m := NewDataChangeManager()
	m.InitTab("tab1", "users", "public", []string{"id"})

	pk := map[string]any{"id": 1}
	_ = m.UpdateCell("tab1", 0, "name", "original", "first_edit", pk)
	_ = m.UpdateCell("tab1", 0, "name", "first_edit", "second_edit", pk)

	pending, _ := m.GetPendingChanges("tab1")
	if len(pending.CellChanges) != 1 {
		t.Fatalf("expected 1 cell change (merged), got %d", len(pending.CellChanges))
	}
	if pending.CellChanges[0].OriginalValue != "original" {
		t.Errorf("expected original value preserved as 'original', got %v", pending.CellChanges[0].OriginalValue)
	}
	if pending.CellChanges[0].NewValue != "second_edit" {
		t.Errorf("expected new value 'second_edit', got %v", pending.CellChanges[0].NewValue)
	}
}

func TestUpdateCell_RevertToOriginal(t *testing.T) {
	m := NewDataChangeManager()
	m.InitTab("tab1", "users", "public", []string{"id"})

	pk := map[string]any{"id": 1}
	_ = m.UpdateCell("tab1", 0, "name", "original", "edited", pk)
	_ = m.UpdateCell("tab1", 0, "name", "edited", "original", pk)

	if m.HasChanges("tab1") {
		t.Fatal("expected no changes after reverting to original")
	}
	if m.GetChangeCount("tab1") != 0 {
		t.Fatalf("expected 0 changes, got %d", m.GetChangeCount("tab1"))
	}
}

func TestInsertRow(t *testing.T) {
	m := NewDataChangeManager()
	m.InitTab("tab1", "users", "public", []string{"id"})

	data := map[string]any{"name": "Alice", "email": "alice@example.com"}
	tempID, err := m.InsertRow("tab1", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tempID == "" {
		t.Fatal("expected non-empty tempID")
	}

	pending, _ := m.GetPendingChanges("tab1")
	if len(pending.InsertedRows) != 1 {
		t.Fatalf("expected 1 inserted row, got %d", len(pending.InsertedRows))
	}
	if pending.InsertedRows[0].TempID != tempID {
		t.Errorf("tempID mismatch: %s != %s", pending.InsertedRows[0].TempID, tempID)
	}
	if pending.Summary.Inserts != 1 {
		t.Errorf("expected 1 insert in summary, got %d", pending.Summary.Inserts)
	}
}

func TestDeleteRow(t *testing.T) {
	m := NewDataChangeManager()
	m.InitTab("tab1", "users", "public", []string{"id"})

	pk := map[string]any{"id": 42}
	err := m.DeleteRow("tab1", pk, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pending, _ := m.GetPendingChanges("tab1")
	if len(pending.DeletedRows) != 1 {
		t.Fatalf("expected 1 deleted row, got %d", len(pending.DeletedRows))
	}
	if pending.DeletedRows[0].RowIndex != 5 {
		t.Errorf("expected rowIndex 5, got %d", pending.DeletedRows[0].RowIndex)
	}
	if pending.Summary.Deletes != 1 {
		t.Errorf("expected 1 delete in summary, got %d", pending.Summary.Deletes)
	}
}

func TestDiscardChanges(t *testing.T) {
	m := NewDataChangeManager()
	m.InitTab("tab1", "users", "public", []string{"id"})

	pk := map[string]any{"id": 1}
	_ = m.UpdateCell("tab1", 0, "name", "old", "new", pk)
	_, _ = m.InsertRow("tab1", map[string]any{"name": "Bob"})
	_ = m.DeleteRow("tab1", map[string]any{"id": 2}, 1)

	if m.GetChangeCount("tab1") != 3 {
		t.Fatalf("expected 3 changes before discard, got %d", m.GetChangeCount("tab1"))
	}

	err := m.DiscardChanges("tab1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.HasChanges("tab1") {
		t.Fatal("expected no changes after discard")
	}
	if m.GetChangeCount("tab1") != 0 {
		t.Fatalf("expected 0 changes after discard, got %d", m.GetChangeCount("tab1"))
	}
}

func TestUndoRedo(t *testing.T) {
	m := NewDataChangeManager()
	m.InitTab("tab1", "users", "public", []string{"id"})

	pk := map[string]any{"id": 1}
	_ = m.UpdateCell("tab1", 0, "name", "original", "edited", pk)

	if m.GetChangeCount("tab1") != 1 {
		t.Fatalf("expected 1 change, got %d", m.GetChangeCount("tab1"))
	}

	err := m.Undo("tab1")
	if err != nil {
		t.Fatalf("undo failed: %v", err)
	}
	if m.GetChangeCount("tab1") != 0 {
		t.Fatalf("expected 0 changes after undo, got %d", m.GetChangeCount("tab1"))
	}

	err = m.Redo("tab1")
	if err != nil {
		t.Fatalf("redo failed: %v", err)
	}
	if m.GetChangeCount("tab1") != 1 {
		t.Fatalf("expected 1 change after redo, got %d", m.GetChangeCount("tab1"))
	}

	pending, _ := m.GetPendingChanges("tab1")
	if pending.CellChanges[0].NewValue != "edited" {
		t.Errorf("expected 'edited' after redo, got %v", pending.CellChanges[0].NewValue)
	}

	err = m.Undo("tab1")
	if err != nil {
		t.Fatalf("second undo failed: %v", err)
	}

	err = m.Undo("tab1")
	if err == nil {
		t.Fatal("expected error on empty undo stack")
	}

	err = m.Redo("tab1")
	if err != nil {
		t.Fatalf("redo after undo failed: %v", err)
	}

	_ = m.UpdateCell("tab1", 0, "name", "edited", "newer", pk)

	err = m.Redo("tab1")
	if err == nil {
		t.Fatal("expected error: new action should clear redo stack")
	}
}

func TestUndoRedo_Insert(t *testing.T) {
	m := NewDataChangeManager()
	m.InitTab("tab1", "users", "public", []string{"id"})

	_, _ = m.InsertRow("tab1", map[string]any{"name": "Alice"})
	if m.GetChangeCount("tab1") != 1 {
		t.Fatalf("expected 1 change, got %d", m.GetChangeCount("tab1"))
	}

	_ = m.Undo("tab1")
	if m.GetChangeCount("tab1") != 0 {
		t.Fatalf("expected 0 after undo insert, got %d", m.GetChangeCount("tab1"))
	}

	_ = m.Redo("tab1")
	if m.GetChangeCount("tab1") != 1 {
		t.Fatalf("expected 1 after redo insert, got %d", m.GetChangeCount("tab1"))
	}
}

func TestUndoRedo_Delete(t *testing.T) {
	m := NewDataChangeManager()
	m.InitTab("tab1", "users", "public", []string{"id"})

	_ = m.DeleteRow("tab1", map[string]any{"id": 1}, 0)
	if m.GetChangeCount("tab1") != 1 {
		t.Fatalf("expected 1 change, got %d", m.GetChangeCount("tab1"))
	}

	_ = m.Undo("tab1")
	if m.GetChangeCount("tab1") != 0 {
		t.Fatalf("expected 0 after undo delete, got %d", m.GetChangeCount("tab1"))
	}

	_ = m.Redo("tab1")
	if m.GetChangeCount("tab1") != 1 {
		t.Fatalf("expected 1 after redo delete, got %d", m.GetChangeCount("tab1"))
	}
}

func TestUndoStackLimit(t *testing.T) {
	m := NewDataChangeManager()
	m.InitTab("tab1", "users", "public", []string{"id"})

	for i := 0; i < 120; i++ {
		pk := map[string]any{"id": i}
		_ = m.UpdateCell("tab1", i, "col", fmt.Sprintf("old_%d", i), fmt.Sprintf("new_%d", i), pk)
	}

	m.mu.RLock()
	stackLen := len(m.tabs["tab1"].UndoStack)
	m.mu.RUnlock()

	if stackLen != 100 {
		t.Fatalf("expected undo stack capped at 100, got %d", stackLen)
	}

	undoCount := 0
	for {
		err := m.Undo("tab1")
		if err != nil {
			break
		}
		undoCount++
	}
	if undoCount != 100 {
		t.Fatalf("expected 100 undos, got %d", undoCount)
	}
}

func TestUninitializedTab(t *testing.T) {
	m := NewDataChangeManager()

	err := m.UpdateCell("nope", 0, "col", "a", "b", map[string]any{"id": 1})
	if err == nil {
		t.Fatal("expected error for uninitialized tab")
	}

	_, err = m.InsertRow("nope", map[string]any{})
	if err == nil {
		t.Fatal("expected error for uninitialized tab")
	}

	err = m.DeleteRow("nope", map[string]any{"id": 1}, 0)
	if err == nil {
		t.Fatal("expected error for uninitialized tab")
	}

	_, err = m.GetPendingChanges("nope")
	if err == nil {
		t.Fatal("expected error for uninitialized tab")
	}

	err = m.DiscardChanges("nope")
	if err == nil {
		t.Fatal("expected error for uninitialized tab")
	}

	if m.HasChanges("nope") {
		t.Fatal("expected false for uninitialized tab")
	}
	if m.GetChangeCount("nope") != 0 {
		t.Fatal("expected 0 for uninitialized tab")
	}
}
