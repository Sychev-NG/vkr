package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

type TransactionManager struct {
    pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
    return &TransactionManager{pool: pool}
}

func (tm *TransactionManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
    if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok && tx != nil {
        return fn(ctx)
    }
    
    tx, err := tm.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    
    txCtx := context.WithValue(ctx, txKey{}, tx)
    
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback(ctx)
            panic(p)
        }
    }()
    
    if err := fn(txCtx); err != nil {
        tx.Rollback(ctx)
        return err
    }
    
    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }
    
    return nil
}

func GetTx(ctx context.Context) (pgx.Tx, bool) {
    tx, ok := ctx.Value(txKey{}).(pgx.Tx)
    return tx, ok
}