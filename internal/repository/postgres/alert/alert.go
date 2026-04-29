package alert

import (
	"context"
	"log"
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

type AlertRepository struct {
	db QueryExecutor
}

func New(db QueryExecutor) *AlertRepository {
	return &AlertRepository{db: db}
}

func (r *AlertRepository) Create(ctx context.Context, productID, warehouseID int, message string) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO alerts (product_id, warehouse_id, message, created_at, is_resolved)
		VALUES ($1, $2, $3, $4, $5)`,
		productID, warehouseID, message, time.Now().UTC(), false,
	)
	if err != nil {
		log.Printf("AlertRepository::Create Error - %v", err)
		return err
	}
	return nil
}

func (r *AlertRepository) Resolve(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `
		UPDATE alerts SET is_resolved = true, resolved_at = $1 WHERE id = $2`,
		time.Now().UTC(), id,
	)
	if err != nil {
		log.Printf("AlertRepository::Resolve Error - %v", err)
		return err
	}
	return nil
}

func (r *AlertRepository) GetByFilter(ctx context.Context, filter entity.AlertFilter) ([]entity.Alert, error) {
	var alerts []entity.Alert

	query := `SELECT id, product_id, warehouse_id, message, created_at, is_resolved, resolved_at FROM alerts WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if filter.ProductID != nil {
		query += " AND product_id = $" + string(rune('0'+argPos))
		args = append(args, *filter.ProductID)
		argPos++
	}
	if filter.WarehouseID != nil {
		query += " AND warehouse_id = $" + string(rune('0'+argPos))
		args = append(args, *filter.WarehouseID)
		argPos++
	}
	if filter.IsResolved != nil {
		query += " AND is_resolved = $" + string(rune('0'+argPos))
		args = append(args, *filter.IsResolved)
		argPos++
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		log.Printf("AlertRepository::GetByFilter Error - %v", err)
		return alerts, err
	}
	defer rows.Close()

	for rows.Next() {
		var alert entity.Alert
		err := rows.Scan(
			&alert.ID,
			&alert.ProductID,
			&alert.WarehouseID,
			&alert.Message,
			&alert.CreatedAt,
			&alert.IsResolved,
			&alert.ResolvedAt,
		)
		if err != nil {
			log.Printf("AlertRepository::GetByFilter Scan Error - %v", err)
			continue
		}
		alerts = append(alerts, alert)
	}

	return alerts, rows.Err()
}