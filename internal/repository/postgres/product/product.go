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

func (pr *ProductRepository) Create(ctx context.Context, name, unit string, minStock float64) (*entity.Product, error) {
	var item entity.Product

	err := pr.db.QueryRow(
		ctx,
		"INSERT INTO products (name, unit, min_stock) VALUES ($1, $2, $3) RETURNING id, name, unit, min_stock",
		name,
		unit,
		minStock,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Unit,
		&item.MinStock,
	)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" { // duplicate error code
			return nil, entity.ErrProductDuplicateFound
		}
		log.Printf("ProductRepository::Create Error - %v", err)
		return nil, err
	}

	return &item, nil
}

func (pr *ProductRepository) Update(ctx context.Context, id int, name, unit string, minStock float64) (*entity.Product, error) {
	var item entity.Product

	err := pr.db.QueryRow(
		ctx,
		"UPDATE products SET name=$1, unit=$2, min_stock=$3 WHERE id=$4 RETURNING id, name, unit, min_stock",
		name,
		unit,
		minStock,
		id,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Unit,
		&item.MinStock,
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
	_, err := pr.db.Exec(ctx, "DELETE FROM products WHERE id=$1", id)
	if err != nil {
		log.Printf("ProductRepository::Delete Error - %v", err)
	}
	return err
}

func (pr *ProductRepository) GetByID(ctx context.Context, id int) (*entity.Product, error) {
	var item entity.Product

	err := pr.db.QueryRow(ctx, "SELECT id, name, unit, min_stock FROM products WHERE id = $1", id).Scan(
		&item.ID,
		&item.Name,
		&item.Unit,
		&item.MinStock,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrProductNotFound
		}
		log.Printf("ProductRepository::GetByID Error - %v", err)
		return nil, err
	}

	return &item, nil
}

func (pr *ProductRepository) GetAll(ctx context.Context) ([]entity.Product, error) {
	var items []entity.Product

	rows, err := pr.db.Query(ctx, "SELECT id, name, unit, min_stock FROM products ORDER BY id")
	if err != nil {
		log.Printf("ProductRepository::GetAll Error - %v", err)
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.Product
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Unit,
			&item.MinStock,
		); err != nil {
			log.Printf("ProductRepository::GetAll Scan Error - %v", err)
			continue
		}
		items = append(items, item)
	}

	return items, nil
}