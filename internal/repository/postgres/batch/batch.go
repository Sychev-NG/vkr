package batch

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"vkr/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type QueryExecutor interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type BatchRepository struct {
	db QueryExecutor
}

func New(db QueryExecutor) *BatchRepository {
	return &BatchRepository{db: db}
}

func (r *BatchRepository) Create(ctx context.Context, vo entity.UpsertBatchVO) (*entity.Batch, error) {
	var batch entity.Batch
	var err error

	err = r.db.QueryRow(ctx, `
		INSERT INTO batches (
			product_id, 
			warehouse_id, 
			document_id, 
			document_type, 
			quantity_remaining, 
			unit_cost, 
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, product_id, warehouse_id, document_id, document_type, quantity_remaining, unit_cost, created_at`,
		vo.ProductID,
		vo.WarehouseID,
		vo.DocumentID,
		vo.DocumentType,
		vo.QuantityRemaining,
		vo.UnitCost,
		time.Now().UTC(),
	).Scan(
		&batch.ID,
		&batch.ProductID,
		&batch.WarehouseID,
		&batch.DocumentID,
		&batch.DocumentType,
		&batch.QuantityRemaining,
		&batch.UnitCost,
		&batch.CreatedAt,
	)

	if err != nil {
		log.Printf("BatchRepository::createIncomingBatch Error - %v", err)
		return nil, err
	}

	return &batch, nil
}

func (r *BatchRepository) GetByID(ctx context.Context, id int) (*entity.Batch, error) {
	var batch entity.Batch

	err := r.db.QueryRow(ctx, `
		SELECT 
			id, 
			product_id, 
			warehouse_id, 
			document_id, 
			document_type, 
			quantity_remaining,
			unit_cost,
			created_at
		FROM batches
		WHERE id = $1`,
		id,
	).Scan(
		&batch.ID,
		&batch.ProductID,
		&batch.WarehouseID,
		&batch.DocumentID,
		&batch.DocumentType,
		&batch.QuantityRemaining,
		&batch.UnitCost,
		&batch.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("batch not found")
		}
		log.Printf("BatchRepository::GetByID Error - %v", err)
		return nil, err
	}

	return &batch, nil
}

// в batch/repository.go
func (r *BatchRepository) GetByFilter(ctx context.Context, filter entity.BatchFilter) ([]entity.Batch, error) {
	var items []entity.Batch

	var query strings.Builder
	query.WriteString(`
		SELECT 
			id, 
			product_id, 
			warehouse_id, 
			document_id, 
			document_type, 
			quantity_remaining,
			unit_cost, 
			created_at 
		FROM batches 
		WHERE 1=1
	`)

	args := []interface{}{}

	if filter.ProductID > 0 {
		query.WriteString(fmt.Sprintf(" AND product_id = $%d", len(args)+1))
		args = append(args, filter.ProductID)
	}

	if filter.WarehouseID > 0 {
		query.WriteString(fmt.Sprintf(" AND warehouse_id = $%d", len(args)+1))
		args = append(args, filter.WarehouseID)
	}

	query.WriteString(" ORDER BY created_at ASC")

	rows, err := r.db.Query(ctx, query.String(), args...)
	if err != nil {
		log.Printf("BatchRepository::GetByFilter Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Batch
		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.WarehouseID,
			&item.DocumentID,
			&item.DocumentType,
			&item.QuantityRemaining,
			&item.UnitCost,
			&item.CreatedAt,
		); err != nil {
			log.Printf("BatchRepository::GetByFilter Scan Error - %v", err)
			continue
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *BatchRepository) GetBatchesForQuantity(ctx context.Context, productID, warehouseID int, neededQuantity float64) ([]*entity.Batch, error) {
	query := `
		WITH running_total AS (
			SELECT 
				id, 
				product_id, 
				warehouse_id, 
				document_id, 
				document_type, 
				quantity_remaining,
				unit_cost,
				created_at,
				SUM(quantity_remaining) OVER (ORDER BY created_at ASC) as running_sum
			FROM batches
			WHERE product_id = $1 
				AND warehouse_id = $2 
				AND quantity_remaining > 0
		)
		SELECT 
			id, 
			product_id, 
			warehouse_id, 
			document_id, 
			document_type, 
			quantity_remaining,
			unit_cost,
			created_at
		FROM running_total
		WHERE running_sum - quantity_remaining < $3
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, productID, warehouseID, neededQuantity)
	if err != nil {
		log.Printf("BatchRepository::GetBatchesForQuantity Query Error - %v", err)
		return nil, err
	}
	defer rows.Close()

	var batches []*entity.Batch
	for rows.Next() {
		var batch entity.Batch
		err := rows.Scan(
			&batch.ID,
			&batch.ProductID,
			&batch.WarehouseID,
			&batch.DocumentID,
			&batch.DocumentType,
			&batch.QuantityRemaining,
			&batch.UnitCost,
			&batch.CreatedAt,
		)
		if err != nil {
			log.Printf("BatchRepository::GetBatchesForQuantity Scan Error - %v", err)
			return nil, err
		}
		batches = append(batches, &batch)
	}

	return batches, rows.Err()
}

func (r *BatchRepository) Subtract(ctx context.Context, batchID int, quantity float64) error {
	result, err := r.db.Exec(ctx, `
		UPDATE batches
		SET quantity_remaining = quantity_remaining - $1
		WHERE id = $2 AND quantity_remaining >= $1
	`, quantity, batchID)
	
	if err != nil {
		log.Printf("BatchRepository::SubtractQuantity Error - %v", err)
		return err
	}
	
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return entity.ErrInsufficientBatch
	}
	
	return nil
}