package change

import (
	"tablepro/internal/driver"
)

// ChangeAction represents the type of change made to data.
type ChangeAction string

const (
	ChangeActionUpdate ChangeAction = "update"
	ChangeActionInsert ChangeAction = "insert"
	ChangeActionDelete ChangeAction = "delete"
)

// CellChange represents a single cell edit in the data grid.
type CellChange struct {
	RowIndex   int    `json:"rowIndex"`
	ColumnName string `json:"columnName"`
	OldValue   any    `json:"oldValue"`
	NewValue   any    `json:"newValue"`
	DataType   string `json:"dataType"`
}

// InsertedRow represents a new row to be inserted.
type InsertedRow struct {
	TempID  string              `json:"tempId"`
	Values  map[string]any      `json:"values"`
	Columns []driver.ColumnInfo `json:"columns"`
}

// DeletedRow represents a row to be deleted.
type DeletedRow struct {
	RowIndex     int            `json:"rowIndex"`
	PrimaryKeys  map[string]any `json:"primaryKeys"`
	OriginalData map[string]any `json:"originalData"`
}

// TabChanges holds all pending changes for a single tab/query result.
type TabChanges struct {
	TabID        string                  `json:"tabId"`
	TableName    string                  `json:"tableName"`
	SchemaName   string                  `json:"schemaName"`
	DatabaseType driver.DatabaseType     `json:"databaseType"`
	CellChanges  map[string][]CellChange `json:"cellChanges"` // keyed by "rowIndex:columnName"
	InsertedRows []InsertedRow           `json:"insertedRows"`
	DeletedRows  []DeletedRow            `json:"deletedRows"`
	PrimaryKeys  []string                `json:"primaryKeys"`
	Columns      []driver.ColumnInfo     `json:"columns"`
}

// PendingChanges aggregates changes across all tabs.
type PendingChanges struct {
	Tabs map[string]*TabChanges `json:"tabs"` // keyed by tabID
}

// ChangeSummary provides a human-readable summary of pending changes.
type ChangeSummary struct {
	TotalUpdates int `json:"totalUpdates"`
	TotalInserts int `json:"totalInserts"`
	TotalDeletes int `json:"totalDeletes"`
	TotalChanges int `json:"totalChanges"`
	AffectedTabs int `json:"affectedTabs"`
}

// Statement represents a generated SQL statement with its parameters.
type Statement struct {
	SQL    string       `json:"sql"`
	Params []any        `json:"params"`
	Action ChangeAction `json:"action"`
}

// UndoEntry represents an undoable action for the undo/redo stack.
type UndoEntry struct {
	Action      ChangeAction `json:"action"`
	TabID       string       `json:"tabId"`
	RowIndex    int          `json:"rowIndex"`
	ColumnName  string       `json:"columnName"`
	OldValue    any          `json:"oldValue"`
	NewValue    any          `json:"newValue"`
	InsertedRow *InsertedRow `json:"insertedRow,omitempty"`
	DeletedRow  *DeletedRow  `json:"deletedRow,omitempty"`
}

// NewTabChanges creates a new TabChanges for the given table.
func NewTabChanges(tabID, tableName, schemaName string, dbType driver.DatabaseType, primaryKeys []string, columns []driver.ColumnInfo) *TabChanges {
	return &TabChanges{
		TabID:        tabID,
		TableName:    tableName,
		SchemaName:   schemaName,
		DatabaseType: dbType,
		CellChanges:  make(map[string][]CellChange),
		InsertedRows: []InsertedRow{},
		DeletedRows:  []DeletedRow{},
		PrimaryKeys:  primaryKeys,
		Columns:      columns,
	}
}

// NewPendingChanges creates an empty PendingChanges.
func NewPendingChanges() *PendingChanges {
	return &PendingChanges{
		Tabs: make(map[string]*TabChanges),
	}
}
