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

func (pr *StockRepository) GetById(ctx context.Context, id int) (*entity.Stock, error) {
	var item entity.Stock

    err := pr.db.QueryRow(ctx, "SELECT id, product_id, warehouse_id, quantity FROM stocks WHERE id = $1", id).Scan(
		&item.ID, 
		&item.ProductID, 
		&item.WarehouseID, 
		&item.Quantity, 
	)
    
    if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrStockNotFound
		}

		log.Printf("StockRepository::GetById Error - %v", err)

        return nil, err
    }
        
    return &item, nil
}

func (pr *StockRepository) GetByProductId(ctx context.Context, id int) ([]entity.Stock, error) {
	var items []entity.Stock

	rows, err := pr.db.Query(ctx, "SELECT id, product_id, warehouse_id, quantity FROM stocks WHERE product_id=$1", id)
	if err != nil {
		log.Printf("StockRepository::GetByProductId Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Stock
		rows.Scan(
			&item.ID, 
			&item.ProductID, 
			&item.WarehouseID, 
			&item.Quantity, 
		)
		items = append(items, item)
	}

	return items, err
}

func (pr *StockRepository) GetByProductAndWarehouseId(ctx context.Context, product_id, warehouse_id int) (*entity.Stock, error) {
	var result entity.Stock

	err := pr.db.QueryRow(ctx, 
		"SELECT id, product_id, warehouse_id, quantity FROM stocks WHERE product_id=$1 AND warehouse_id=$2", 
		product_id, warehouse_id,
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
		log.Printf("StockRepository::GetByProductAndWarehouseId Error - %v", err)
		return nil, err
	}

	return &result, nil
}

func (pr *StockRepository) GetAll(ctx context.Context) ([]entity.Stock, error) {
	var items []entity.Stock

	rows, err := pr.db.Query(ctx, "SELECT id, product_id, warehouse_id, quantity FROM stocks")
	if err != nil {
		log.Printf("StockRepository::GetAll Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Stock
		rows.Scan(
			&item.ID, 
			&item.ProductID, 
			&item.WarehouseID, 
			&item.Quantity, 
		)
		items = append(items, item)
	}

	return items, err
}

func (pr *StockRepository) GetByFilter(ctx context.Context, filter entity.StockFilter) ([]entity.Stock, error) {
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
	
	rows, err := pr.db.Query(ctx, query.String(), args...)
	if err != nil {
		log.Printf("StockRepository::GetAll Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Stock
		rows.Scan(
			&item.ID, 
			&item.ProductID, 
			&item.WarehouseID, 
			&item.Quantity, 
		)
		items = append(items, item)
	}

	return items, err
}

func (pr *StockRepository) Increase(ctx context.Context, product_id, warehouse_id int, quantity float32) error {    
	stock, err := pr.GetByProductAndWarehouseId(ctx, product_id, warehouse_id)   
    if err != nil && !errors.Is(err, entity.ErrStockNotFound) {
        return err
    }

	if stock != nil {
		// Update
		_, err = pr.updateQuantity(ctx, stock.ID, stock.Quantity + quantity)
		if err != nil {
        	return err
    	}
	} else {
		// Add
		_, err = pr.Add(ctx, product_id, warehouse_id, quantity)
		if err != nil {
        	return err
    	}
	}
    
    return nil
}

func (pr *StockRepository) Decrease(ctx context.Context, product_id, warehouse_id int, quantity float32) error {
	return nil
}

func (pr *StockRepository) Add(ctx context.Context, product_id, warehouse_id int, quantity float32) (*entity.Stock, error) {
    var item entity.Stock
    
    err := pr.db.QueryRow(
        ctx, 
        "INSERT INTO stocks (product_id, warehouse_id, quantity) VALUES ($1, $2, $3) RETURNING id, product_id, warehouse_id, quantity", 
        product_id, 
        warehouse_id, 
        quantity,
    ).Scan(
        &item.ID, 
        &item.ProductID, 
        &item.WarehouseID, 
        &item.Quantity, 
    )
    
    if err != nil {
        return nil, err
    }
    
    return &item, nil
}

func (pr *StockRepository) updateQuantity(ctx context.Context, id int, quantity float32) (*entity.Stock, error) {
	var item entity.Stock

	err := pr.db.QueryRow(
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

		log.Printf("ProductRepository::Update Error - %v", err)
		return nil, err
	}

	return &item, nil
}

func (pr *StockRepository) Delete(ctx context.Context, id int) error {
	_, err :=pr.db.Exec(ctx, "DELETE FROM stocks WHERE id=$1", id)
	if err != nil {
		log.Printf("StockRepository::Delete Error - %v", err)
	}
	return err
}