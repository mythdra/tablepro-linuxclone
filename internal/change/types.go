package change

type CellChange struct {
	RowIndex    int            `json:"rowIndex"`
	ColumnName  string         `json:"columnName"`
	OldValue    any            `json:"oldValue"`
	NewValue    any            `json:"newValue"`
	PrimaryKeys map[string]any `json:"primaryKeys"`
}

type InsertedRow struct {
	TempID string         `json:"tempId"`
	Values map[string]any `json:"values"`
}

type DeletedRow struct {
	RowIndex    int            `json:"rowIndex"`
	PrimaryKeys map[string]any `json:"primaryKeys"`
}

type TabChanges struct {
	SchemaName   string        `json:"schemaName"`
	TableName    string        `json:"tableName"`
	CellChanges  []CellChange  `json:"cellChanges"`
	InsertedRows []InsertedRow `json:"insertedRows"`
	DeletedRows  []DeletedRow  `json:"deletedRows"`
}

func (tc *TabChanges) HasChanges() bool {
	return len(tc.CellChanges) > 0 || len(tc.InsertedRows) > 0 || len(tc.DeletedRows) > 0
}

func (tc *TabChanges) Clear() {
	tc.CellChanges = nil
	tc.InsertedRows = nil
	tc.DeletedRows = nil
}

func (tc *TabChanges) ChangeCount() int {
	return len(tc.CellChanges) + len(tc.InsertedRows) + len(tc.DeletedRows)
}
