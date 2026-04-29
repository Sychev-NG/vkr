package movement

import (
	"context"
	"fmt"
	"log"
	"strings"
	"vkr/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type QueryExecutor interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type MovementRepository struct {
	db QueryExecutor
}

func New(db QueryExecutor) *MovementRepository {
	return &MovementRepository{db: db}
}

func (r *MovementRepository) GetByFilter(ctx context.Context, filter entity.MovementFilter) ([]entity.Movement, error) {
	var items []entity.Movement

	var query strings.Builder
	query.WriteString(`
		SELECT 
			movement_id,
			batch_id,
			product_id,
			warehouse_id,
			quantity,
			unit_cost,
			moved_at
		FROM movements 
		WHERE 1=1
	`)

	args := []interface{}{}

	if filter.DocumentID > 0 {
		query.WriteString(fmt.Sprintf(" AND movement_document_id = $%d", len(args)+1))
		args = append(args, filter.DocumentID)
	}

	if filter.DocumentType != "" {
		query.WriteString(fmt.Sprintf(" AND movement_document_type = $%d", len(args)+1))
		args = append(args, filter.DocumentType)
	}

	if filter.BatchID > 0 {
		query.WriteString(fmt.Sprintf(" AND batch_id = $%d", len(args)+1))
		args = append(args, filter.BatchID)
	}

	if filter.ProductID > 0 {
		query.WriteString(fmt.Sprintf(" AND product_id = $%d", len(args)+1))
		args = append(args, filter.ProductID)
	}

	if filter.WarehouseID > 0 {
		query.WriteString(fmt.Sprintf(" AND warehouse_id = $%d", len(args)+1))
		args = append(args, filter.WarehouseID)
	}

	query.WriteString(" ORDER BY moved_at DESC")

	rows, err := r.db.Query(ctx, query.String(), args...)
	if err != nil {
		log.Printf("MovementRepository::GetByFilter Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Movement
		if err := rows.Scan(
			&item.ID,
			&item.BatchID,
			&item.ProductID,
			&item.WarehouseID,
			&item.StockMovement,
			&item.UnitCost,
			&item.Date,
		); err != nil {
			log.Printf("MovementRepository::GetByFilter Scan Error - %v", err)
			continue
		}
		items = append(items, item)
	}

	return items, rows.Err()
}