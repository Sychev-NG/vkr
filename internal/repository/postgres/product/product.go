package product

import (
	"context"
	"vkr/internal/entity"

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
	return &ProductRepository{}
}

func (pr *ProductRepository) Add(ctx context.Context, name, unit, productType string) (*entity.Product, error) {
	return &products[len(products)-1], nil
}

func (pr *ProductRepository) Update(ctx context.Context, product entity.Product) (*entity.Product, error) {
	return &products[len(products)-1], nil
}

func (pr *ProductRepository) Delete(ctx context.Context, id int) error {
	return nil
}

func (pr *ProductRepository) GetById(ctx context.Context, id int) (*entity.Product, error) {
	return &products[len(products)-1], nil
}

func (pr *ProductRepository) GetAll(ctx context.Context) ([]entity.Product, error) {
	return products, nil
}
