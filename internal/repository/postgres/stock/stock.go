package stock

import (
	"context"
	"errors"
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

type StockRepository struct {
	db QueryExecutor
}

func New(db QueryExecutor) *StockRepository {
	return &StockRepository{db: db}
}

func (r *StockRepository) GetByID(ctx context.Context, id int) (*entity.Stock, error) {
	var item entity.Stock

	err := r.db.QueryRow(ctx, "SELECT id, product_id, warehouse_id, quantity FROM stocks WHERE id = $1", id).Scan(
		&item.ID,
		&item.ProductID,
		&item.WarehouseID,
		&item.Quantity,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrStockNotFound
		}
		log.Printf("StockRepository::GetByID Error - %v", err)
		return nil, err
	}

	return &item, nil
}

func (r *StockRepository) GetByProductID(ctx context.Context, productID int) ([]entity.Stock, error) {
	var items []entity.Stock

	rows, err := r.db.Query(ctx, "SELECT id, product_id, warehouse_id, quantity FROM stocks WHERE product_id=$1", productID)
	if err != nil {
		log.Printf("StockRepository::GetByProductID Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Stock
		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.WarehouseID,
			&item.Quantity,
		); err != nil {
			log.Printf("StockRepository::GetByProductID Scan Error - %v", err)
			continue
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *StockRepository) GetByProductAndWarehouse(ctx context.Context, productID, warehouseID int) (*entity.Stock, error) {
	var result entity.Stock

	err := r.db.QueryRow(ctx,
		"SELECT id, product_id, warehouse_id, quantity FROM stocks WHERE product_id=$1 AND warehouse_id=$2",
		productID, warehouseID,
	).Scan(
		&result.ID,
		&result.ProductID,
		&result.WarehouseID,
		&result.Quantity,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrStockNotFound
		}
		log.Printf("StockRepository::GetByProductAndWarehouse Error - %v", err)
		return nil, err
	}

	return &result, nil
}

func (r *StockRepository) GetAll(ctx context.Context) ([]entity.Stock, error) {
	var items []entity.Stock

	rows, err := r.db.Query(ctx, "SELECT id, product_id, warehouse_id, quantity FROM stocks ORDER BY product_id, warehouse_id")
	if err != nil {
		log.Printf("StockRepository::GetAll Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Stock
		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.WarehouseID,
			&item.Quantity,
		); err != nil {
			log.Printf("StockRepository::GetAll Scan Error - %v", err)
			continue
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *StockRepository) GetByFilter(ctx context.Context, filter entity.StockFilter) ([]entity.Stock, error) {
	var items []entity.Stock

	var query strings.Builder
	query.WriteString("SELECT id, product_id, warehouse_id, quantity FROM stocks WHERE 1=1")

	args := []interface{}{}

	if filter.ProductID > 0 {
		query.WriteString(fmt.Sprintf(" AND product_id = $%d", len(args)+1))
		args = append(args, filter.ProductID)
	}

	if filter.WarehouseID > 0 {
		query.WriteString(fmt.Sprintf(" AND warehouse_id = $%d", len(args)+1))
		args = append(args, filter.WarehouseID)
	}

	query.WriteString(" ORDER BY product_id, warehouse_id")

	rows, err := r.db.Query(ctx, query.String(), args...)
	if err != nil {
		log.Printf("StockRepository::GetByFilter Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Stock
		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.WarehouseID,
			&item.Quantity,
		); err != nil {
			log.Printf("StockRepository::GetByFilter Scan Error - %v", err)
			continue
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *StockRepository) Increase(ctx context.Context, productID, warehouseID int, quantity float64) error {
	stock, err := r.GetByProductAndWarehouse(ctx, productID, warehouseID)
	if err != nil && !errors.Is(err, entity.ErrStockNotFound) {
		log.Printf("StockRepository::Increase GetByProductAndWarehouse Error - %v", err)
		return err
	}

	if stock != nil {
		// Update existing
		_, err = r.updateQuantity(ctx, stock.ID, stock.Quantity+quantity)
		if err != nil {
			return err
		}
	} else {
		// Create new
		_, err = r.Create(ctx, productID, warehouseID, quantity)
		if err != nil {
			return err
		}
	}

	return nil
}

// Decrease уменьшает остаток на указанное количество
func (r *StockRepository) Decrease(ctx context.Context, productID, warehouseID int, quantity float64) error {
	stock, err := r.GetByProductAndWarehouse(ctx, productID, warehouseID)
	if err != nil {
		log.Printf("StockRepository::Decrease GetByProductAndWarehouse Error - %v", err)
		return err
	}

	if stock != nil {
		newQuantity := stock.Quantity - quantity
		if newQuantity < 0 {
			return entity.ErrInsufficientStock
		}
		_, err = r.updateQuantity(ctx, stock.ID, newQuantity)
		if err != nil {
			log.Printf("StockRepository::Decrease updateQuantity Error - %v", err)
			return err
		}
	} else {
		return entity.ErrStockNotFound
	}

	return nil
}

// Create создаёт новую запись остатка
func (r *StockRepository) Create(ctx context.Context, productID, warehouseID int, quantity float64) (*entity.Stock, error) {
	var item entity.Stock

	err := r.db.QueryRow(
		ctx,
		"INSERT INTO stocks (product_id, warehouse_id, quantity) VALUES ($1, $2, $3) RETURNING id, product_id, warehouse_id, quantity",
		productID,
		warehouseID,
		quantity,
	).Scan(
		&item.ID,
		&item.ProductID,
		&item.WarehouseID,
		&item.Quantity,
	)

	if err != nil {
		log.Printf("StockRepository::Create Error - %v", err)
		return nil, err
	}

	return &item, nil
}

// // CreateOrUpdate создаёт или обновляет остаток (upsert)
// func (r *StockRepository) CreateOrUpdate(ctx context.Context, productID, warehouseID int, quantity float64) error {
// 	_, err := r.db.Exec(ctx, `
// 		INSERT INTO stocks (product_id, warehouse_id, quantity) 
// 		VALUES ($1, $2, $3)
// 		ON CONFLICT (product_id, warehouse_id) 
// 		DO UPDATE SET quantity = stocks.quantity + $3
// 	`, productID, warehouseID, quantity)

// 	if err != nil {
// 		log.Printf("StockRepository::CreateOrUpdate Error - %v", err)
// 		return err
// 	}
// 	return nil
// }

// // SetQuantity устанавливает точное значение остатка
// func (r *StockRepository) SetQuantity(ctx context.Context, productID, warehouseID int, quantity float64) error {
// 	_, err := r.db.Exec(ctx, `
// 		INSERT INTO stocks (product_id, warehouse_id, quantity) 
// 		VALUES ($1, $2, $3)
// 		ON CONFLICT (product_id, warehouse_id) 
// 		DO UPDATE SET quantity = $3
// 	`, productID, warehouseID, quantity)

// 	if err != nil {
// 		log.Printf("StockRepository::SetQuantity Error - %v", err)
// 		return err
// 	}
// 	return nil
// }

// updateQuantity обновляет количество по ID
func (r *StockRepository) updateQuantity(ctx context.Context, id int, quantity float64) (*entity.Stock, error) {
	var item entity.Stock

	err := r.db.QueryRow(
		ctx,
		"UPDATE stocks SET quantity=$1 WHERE id=$2 RETURNING id, product_id, warehouse_id, quantity",
		quantity,
		id,
	).Scan(
		&item.ID,
		&item.ProductID,
		&item.WarehouseID,
		&item.Quantity,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrStockNotFound
		}
		log.Printf("StockRepository::updateQuantity Error - %v", err)
		return nil, err
	}

	return &item, nil
}

// Delete удаляет запись остатка
func (r *StockRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, "DELETE FROM stocks WHERE id=$1", id)
	if err != nil {
		log.Printf("StockRepository::Delete Error - %v", err)
	}
	return err
}

// // GetLowStock возвращает остатки ниже порогового значения
// func (r *StockRepository) GetLowStock(ctx context.Context) ([]entity.Stock, error) {
// 	var items []entity.Stock

// 	rows, err := r.db.Query(ctx, `
// 		SELECT s.id, s.product_id, s.warehouse_id, s.quantity, p.min_stock
// 		FROM stocks s
// 		JOIN products p ON s.product_id = p.id
// 		WHERE s.quantity <= p.min_stock AND p.min_stock > 0
// 	`)
// 	if err != nil {
// 		log.Printf("StockRepository::GetLowStock Error - %v", err)
// 		return items, err
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var item entity.Stock
// 		var minStock float64
// 		if err := rows.Scan(
// 			&item.ID,
// 			&item.ProductID,
// 			&item.WarehouseID,
// 			&item.Quantity,
// 			&minStock,
// 		); err != nil {
// 			log.Printf("StockRepository::GetLowStock Scan Error - %v", err)
// 			continue
// 		}
// 		items = append(items, item)
// 	}

// 	return items, rows.Err()
// }