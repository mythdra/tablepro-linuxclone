package change

import (
	"context"
	"database/sql"
	"fmt"

	"tablepro/internal/driver"
)

func CommitChanges(ctx context.Context, db *sql.DB, changes *TabChanges, dbType driver.DatabaseType, autoIncrementCols []string) error {
	if changes == nil {
		return fmt.Errorf("changes is nil")
	}
	if !changes.HasChanges() {
		return nil
	}

	gen := NewSQLStatementGenerator(dbType)
	stmts, err := gen.GenerateAll(changes, autoIncrementCols)
	if err != nil {
		return fmt.Errorf("generating statements: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	for i, stmt := range stmts {
		_, err := tx.ExecContext(ctx, stmt.SQL, stmt.Args...)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return fmt.Errorf("statement #%d failed: %w (rollback also failed: %v)", i, err, rbErr)
			}
			return fmt.Errorf("statement #%d failed: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
