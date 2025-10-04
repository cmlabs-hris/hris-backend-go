package postgresql

import (
	"context"
	"fmt"

	"github.com/cmlabs-hris/hris-backend-go/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

// WithTransaction executes fn inside a database transaction
func WithTransaction(ctx context.Context, db *database.DB, fn func(tx pgx.Tx) error) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				fmt.Printf("rollback error during panic recovery: %v\n", rbErr)
			}
			panic(p)
		}
	}()

	// Execute function
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rollback error: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// GetQuerier returns either transaction or pool
// Used in repositories to support both transactional and non-transactional operations
func GetQuerier(ctx context.Context, db *database.DB) database.Querier {
	if tx, ok := ctx.Value("tx").(pgx.Tx); ok {
		return tx
	}
	return db.Pool
}
