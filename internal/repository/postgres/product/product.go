package product

import (
	"context"
	"errors"
	"log"
	"vkr/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var products = []entity.Product {
	{ID: 1, Name: "Мука", Unit: "кг", TypeName: "raw"},
	{ID: 2, Name: "Дрожжи", Unit: "кг", TypeName: "raw"},
}

type ProductRepository struct {
	pool *pgxpool.Pool
}

func New(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: db}
}

func (pr *ProductRepository) Add(ctx context.Context, name, unit, productType string) (*entity.Product, error) {
	var item entity.Product

	err := pr.pool.QueryRow(
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

func (pr *ProductRepository) Update(ctx context.Context, product entity.Product) (*entity.Product, error) {
	return &products[len(products)-1], nil
}

func (pr *ProductRepository) Delete(ctx context.Context, id int) error {
	_, err :=pr.pool.Exec(ctx, "DELETE FROM products WHERE id=$1", id)
	if err != nil {
		log.Printf("ProductRepository::Delete Error - %v", err)
	}
	return err
}

func (pr *ProductRepository) GetById(ctx context.Context, id int) (*entity.Product, error) {
    var item entity.Product

    err := pr.pool.QueryRow(ctx, "SELECT id, name, unit, type FROM products WHERE id = $1", id).Scan(
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

	rows, err := pr.pool.Query(ctx, "SELECT id, name, unit, type FROM products")
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
