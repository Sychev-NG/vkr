package movement

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
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

type MovementRepository struct {
	db QueryExecutor
}

func New(db QueryExecutor) *MovementRepository {
	return &MovementRepository{db: db}
}

func (pr *MovementRepository) GetById(ctx context.Context, id int) (*entity.Movement, error) {
	var item entity.Movement

    err := pr.db.QueryRow(ctx, "SELECT id, product_id, warehouse_id, document_id, quantity, date FROM movements WHERE id = $1", id).Scan(
		&item.ID, 
		&item.ProductID, 
		&item.WarehouseID, 
		&item.DocumentID, 
		&item.Quantity,
		&item.Date,
	)
    
    if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrMovementNotFound
		}

		log.Printf("MovementRepository::GetById Error - %v", err)

        return nil, err
    }
        
    return &item, nil
}

func (pr *MovementRepository) GetAll(ctx context.Context) ([]entity.Movement, error) {
	var items []entity.Movement

	rows, err := pr.db.Query(ctx, "SELECT id, product_id, warehouse_id, document_id, document_type, quantity, date FROM movements")
	if err != nil {
		log.Printf("MovementRepository::GetAll Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Movement
		var typeName string
		rows.Scan(
			&item.ID, 
			&item.ProductID, 
			&item.WarehouseID, 
			&item.DocumentID, 
			&typeName, 
			&item.Quantity,
			&item.Date,
		)

		item.DocumentType = entity.DocumentType(typeName)

		items = append(items, item)
	}

	return items, err
}

func (pr *MovementRepository) GetByFilter(ctx context.Context, filter entity.MovementFilter) ([]entity.Movement, error) {
	var items []entity.Movement

	var query strings.Builder
	query.WriteString("SELECT id, product_id, warehouse_id, document_id, document_type, quantity, date FROM movements WHERE 1=1")
	
	args := []interface{}{}
	
	if filter.ProductID > 0 {
		query.WriteString(fmt.Sprintf(" AND product_id = $%d", len(args)+1))
		args = append(args, filter.ProductID)
	}
	
	if filter.WarehouseID > 0 {
		query.WriteString(fmt.Sprintf(" AND warehouse_id = $%d", len(args)+1))
		args = append(args, filter.WarehouseID)
	}
	
	rows, err := pr.db.Query(ctx, query.String(), args...)
	if err != nil {
		log.Printf("MovementRepository::GetByFilter Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Movement
		var typeName string
		rows.Scan(
			&item.ID, 
			&item.ProductID, 
			&item.WarehouseID, 
			&item.DocumentID, 
			&typeName,
			&item.Quantity, 
			&item.Date, 
		)

		item.DocumentType = entity.DocumentType(typeName)

		items = append(items, item)
	}

	return items, err
}
func (pr *MovementRepository) RegisterIncoming(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity float32) (*entity.Movement, error) {
	var result entity.Movement

	err := pr.db.QueryRow(
		ctx, 
		"INSERT INTO movements (product_id, warehouse_id, document_id, document_type, quantity, date) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, product_id, warehouse_id, document_id, document_type, quantity, date", 
		product_id, 
		warehouse_id, 
		docVO.DocumentID,
		docVO.Type,
		quantity,
		time.Now().UTC(),
	).Scan(
		&result.ID, 
		&result.ProductID, 
		&result.WarehouseID, 
		&result.DocumentID, 
		&result.DocumentType, 
		&result.Quantity,
		&result.Date,
	)

	if err != nil {
		log.Printf("MovementRepository::RegisterIncoming QueryRow Error - %v", err)
		return nil, err
	}

	return &result, nil
}

func (pr *MovementRepository) RegisterOutgoing(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity float32) (*entity.Movement, error) {
	var result entity.Movement

	outgoingQuantity := -quantity

	err := pr.db.QueryRow(
		ctx, 
		"INSERT INTO movements (product_id, warehouse_id, document_id, document_type, quantity, date) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, product_id, warehouse_id, document_id, document_type, quantity, date", 
		product_id, 
		warehouse_id, 
		docVO.DocumentID,
		docVO.Type,
		outgoingQuantity,
		time.Now().UTC(),
	).Scan(
		&result.ID, 
		&result.ProductID, 
		&result.WarehouseID, 
		&result.DocumentID, 
		&result.DocumentType, 
		&result.Quantity,
		&result.Date,
	)

	if err != nil {
		log.Printf("MovementRepository::RegisterOutgoing QueryRow Error - %v", err)
		return nil, err
	}

	return &result, nil
}