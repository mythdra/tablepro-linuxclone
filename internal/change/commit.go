package change

import (
	"context"
	"database/sql"
	"fmt"
)

type CommitResult struct {
	StatementsExecuted int   `json:"statementsExecuted"`
	RowsAffected       int64 `json:"rowsAffected"`
}

func CommitChanges(ctx context.Context, db *sql.DB, statements []Statement) (*CommitResult, error) {
	if len(statements) == 0 {
		return &CommitResult{}, nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction failed: %w", err)
	}

	var totalRows int64

	for i, stmt := range statements {
		result, err := tx.ExecContext(ctx, stmt.SQL, stmt.Params...)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return nil, fmt.Errorf("statement %d failed: %w (rollback also failed: %v)", i, err, rollbackErr)
			}
			return nil, fmt.Errorf("statement %d failed (rolled back): %w", i, err)
		}

		rows, _ := result.RowsAffected()
		totalRows += rows
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit failed: %w", err)
	}

	return &CommitResult{
		StatementsExecuted: len(statements),
		RowsAffected:       totalRows,
	}, nil
}
