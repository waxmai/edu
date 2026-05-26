package mysql

import (
	"context"

	"gorm.io/gorm"
)

// Transaction executes a transaction within the given function.
// It creates a new context with the transactional DB instance.
func (d *dbProvider) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Use the write DB for transactions
	db := d.DbW.WithContext(ctx)

	return db.Transaction(func(tx *gorm.DB) error {
		// Create a context with the transaction DB
		txCtx := context.WithValue(ctx, ctxTransactionKey{}, tx)
		return fn(txCtx)
	})
}

type ctxTransactionKey struct{}

// GetDBFromContext tries to extract the DB from the context.
// If found, it returns the transactional DB.
// Otherwise, it returns the provided default DB.
// Exported to be used by dao package.
func GetDBFromContext(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(ctxTransactionKey{}).(*gorm.DB); ok {
		return tx
	}
	return defaultDB
}
