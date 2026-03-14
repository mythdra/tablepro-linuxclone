package change

// CellChange represents a single cell modification in a table.
type CellChange struct {
	RowIndex      int            `json:"rowIndex"`
	Column        string         `json:"column"`
	OriginalValue any            `json:"originalValue"`
	NewValue      any            `json:"newValue"`
	PrimaryKey    map[string]any `json:"primaryKey"`
}

// InsertedRow represents a new row to be inserted.
type InsertedRow struct {
	TempID string         `json:"tempId"`
	Data   map[string]any `json:"data"`
}

// DeletedRow represents a row marked for deletion.
type DeletedRow struct {
	PrimaryKey map[string]any `json:"primaryKey"`
	RowIndex   int            `json:"rowIndex"`
}

// TabChanges holds all pending changes for a single tab.
type TabChanges struct {
	TabID        string         `json:"tabId"`
	TableName    string         `json:"tableName"`
	SchemaName   string         `json:"schemaName"`
	PrimaryKeys  []string       `json:"primaryKeys"`
	CellChanges  []*CellChange  `json:"cellChanges"`
	InsertedRows []*InsertedRow `json:"insertedRows"`
	DeletedRows  []*DeletedRow  `json:"deletedRows"`
	UndoStack    []ChangeAction `json:"-"`
	RedoStack    []ChangeAction `json:"-"`
}

// HasChanges returns true if there are any pending changes.
func (tc *TabChanges) HasChanges() bool {
	return len(tc.CellChanges) > 0 || len(tc.InsertedRows) > 0 || len(tc.DeletedRows) > 0
}

// Clear removes all pending changes.
func (tc *TabChanges) Clear() {
	tc.CellChanges = nil
	tc.InsertedRows = nil
	tc.DeletedRows = nil
}

// ChangeCount returns the total number of pending changes.
func (tc *TabChanges) ChangeCount() int {
	return len(tc.CellChanges) + len(tc.InsertedRows) + len(tc.DeletedRows)
}

// ChangeAction represents a single undoable/redoable action.
type ChangeAction struct {
	Type    string `json:"type"` // "update", "insert", "delete"
	Payload any    `json:"payload"`
}

// PendingChanges is the summary returned to the frontend.
type PendingChanges struct {
	CellChanges  []*CellChange  `json:"cellChanges"`
	InsertedRows []*InsertedRow `json:"insertedRows"`
	DeletedRows  []*DeletedRow  `json:"deletedRows"`
	Summary      ChangeSummary  `json:"summary"`
}

// ChangeSummary provides counts of each change type.
type ChangeSummary struct {
	Updates int `json:"updates"`
	Inserts int `json:"inserts"`
	Deletes int `json:"deletes"`
}

// Statement represents a generated SQL statement with its arguments.
type Statement struct {
	SQL  string `json:"sql"`
	Args []any  `json:"args"`
	Type string `json:"type"` // "UPDATE", "INSERT", "DELETE"
}
