package product

import (
	"context"
	"errors"
	"log"
	"vkr/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type QueryExecutor interface {
    QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
    Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}

type ProductRepository struct {
	db QueryExecutor
}

func New(db QueryExecutor) *ProductRepository {
	return &ProductRepository{db: db}
}

func (pr *ProductRepository) Add(ctx context.Context, name, unit, productType string) (*entity.Product, error) {
	var item entity.Product

	err := pr.db.QueryRow(
		ctx, 
		"INSERT INTO products (name, unit, type) VALUES ($1, $2, $3) RETURNING id, name, unit, type", 
		name, 
		unit, 
		productType,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Unit,
		&item.TypeName,
	)

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (pr *ProductRepository) Update(ctx context.Context, id int, name, unit, productType string) (*entity.Product, error) {
	var item entity.Product

	err := pr.db.QueryRow(
		ctx, 
		"UPDATE products SET name=$1, unit=$2, type=$3 WHERE id=$4 RETURNING id, name, unit, type", 
		name, 
		unit, 
		productType,
		id,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Unit,
		&item.TypeName,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrProductNotFound
		}

		log.Printf("ProductRepository::Update Error - %v", err)
		return nil, err
	}

	return &item, nil
}

func (pr *ProductRepository) Delete(ctx context.Context, id int) error {
	_, err :=pr.db.Exec(ctx, "DELETE FROM products WHERE id=$1", id)
	if err != nil {
		log.Printf("ProductRepository::Delete Error - %v", err)
	}
	return err
}

func (pr *ProductRepository) GetById(ctx context.Context, id int) (*entity.Product, error) {
    var item entity.Product

    err := pr.db.QueryRow(ctx, "SELECT id, name, unit, type FROM products WHERE id = $1", id).Scan(
		&item.ID, 
		&item.Name, 
		&item.Unit, 
		&item.TypeName,
	)
    
    if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrProductNotFound
		}

        return nil, err
    }
        
    return &item, nil
}

func (pr *ProductRepository) GetAll(ctx context.Context) ([]entity.Product, error) {
	var items []entity.Product

	rows, err := pr.db.Query(ctx, "SELECT id, name, unit, type FROM products")
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Product
		rows.Scan(
			&item.ID, 
			&item.Name, 
			&item.Unit, 
			&item.TypeName,
		)
		items = append(items, item)
	}

	return items, err
}
