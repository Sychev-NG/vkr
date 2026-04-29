package batchmovement

import (
	"context"
	"log"
	"time"
	"vkr/internal/entity"
	"vkr/internal/entity/document"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type QueryExecutor interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type BatchMovementRepository struct {
	db QueryExecutor
}

func New(db QueryExecutor) *BatchMovementRepository {
	return &BatchMovementRepository{db: db}
}

func (pr *BatchMovementRepository) RegisterIncoming(ctx context.Context, docVO document.Document, batch_id int, quantity float64) (*entity.BatchMovement, error) {
	var result entity.BatchMovement

	err := pr.db.QueryRow(
		ctx, 
		`INSERT INTO batch_movements (batch_id, document_type, document_id, quantity, moved_at) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, batch_id, document_type, document_id, quantity, moved_at`, 
		batch_id,  
		docVO.Type,
		docVO.DocumentID,
		quantity,
		time.Now().UTC(),
	).Scan(
		&result.ID,  
		&result.BatchID,  
		&result.DocumentType,
		&result.DocumentID,  
		&result.Quantity,
		&result.CreatedAt,
	)

	if err != nil {
		log.Printf("BatchMovementRepository::RegisterIncoming QueryRow Error - %v", err)
		return nil, err
	}

	return &result, nil
}

func (pr *BatchMovementRepository) RegisterOutgoing(ctx context.Context, docVO document.Document, batch_id int, quantity float64) (*entity.BatchMovement, error) {
	var result entity.BatchMovement

	err := pr.db.QueryRow(
		ctx, 
		`INSERT INTO batch_movements (batch_id, document_type, document_id, quantity, moved_at) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, batch_id, document_type, document_id, quantity, moved_at`, 
		batch_id,  
		docVO.Type,
		docVO.DocumentID,
		-quantity,
		time.Now().UTC(),
	).Scan(
		&result.ID,  
		&result.BatchID,  
		&result.DocumentType,
		&result.DocumentID,  
		&result.Quantity,
		&result.CreatedAt,
	)

	if err != nil {
		log.Printf("BatchMovementRepository::RegisterOutgoing QueryRow Error - %v", err)
		return nil, err
	}

	return &result, nil
}

func (r *BatchMovementRepository) GetByID(ctx context.Context, batchID int) (*entity.BatchMovement, error) {
	var result entity.BatchMovement

	err := r.db.QueryRow(ctx, `
		SELECT id, batch_id, document_type, document_id, quantity, moved_at
		FROM batch_movements WHERE batch_id = $1 ORDER BY created_at ASC`, batchID,
	).Scan(
		&result.ID,  
		&result.BatchID,  
		&result.DocumentType,
		&result.DocumentID,  
		&result.Quantity,
		&result.CreatedAt,
	)

	if err != nil {
		log.Printf("BatchMovementRepository::GetByID Scan Error - %v", err)
		return nil, err
	}

	return &result, nil
}

func (r *BatchMovementRepository) GetByBatchID(ctx context.Context, batchID int) ([]entity.BatchMovement, error) {
	var result []entity.BatchMovement

	rows, err := r.db.Query(ctx, `
		SELECT id, batch_id, target_batch_id, document_type, document_id, quantity, quantity_remaining_before, quantity_remaining_after, created_at
		FROM batch_movements WHERE batch_id = $1 ORDER BY created_at ASC`, batchID,
	)
	if err != nil {
		log.Printf("BatchMovementRepository::GetByBatchID Error - %v", err)
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var mov entity.BatchMovement
		err := rows.Scan(
			&mov.ID,  
			&mov.BatchID,  
			&mov.DocumentType,
			&mov.DocumentID,  
			&mov.Quantity,
			&mov.CreatedAt,
		)
		if err != nil {
			log.Printf("BatchMovementRepository::GetByBatchID Scan Error - %v", err)
			continue
		}

		mov.Quantity = -mov.Quantity
		result = append(result, mov)
	}

	return result, rows.Err()
}