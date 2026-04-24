package postgres

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

type TransactionManager struct {
    pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) *TransactionManager {
    return &TransactionManager{pool: pool}
}

func (tm *TransactionManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
    tx, err := tm.pool.Begin(ctx)
    if err != nil {
        return err
    }
    
    txCtx := context.WithValue(ctx, "tx", tx)
    
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
    
    return tx.Commit(ctx)
}