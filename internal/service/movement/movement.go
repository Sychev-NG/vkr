package movement

import (
	"context"

	"vkr/internal/entity"
)

type MovementSaver interface {
	// Add(ctx context.Context, product_id, warehouse_id int, quantity float32) (*entity.Movement, error)
	// Update(ctx context.Context, id int, product_id, warehouse_id int, quantity float32) (*entity.Movement, error)
	// Delete(ctx context.Context, id int) (error)
}

type MovementProvider interface{
	GetById(ctx context.Context, id int) (*entity.Movement, error)
	GetAll(ctx context.Context) ([]entity.Movement, error)
	GetByFilter(ctx context.Context, filter entity.MovementFilter) ([]entity.Movement, error)
}

type ProductProvider interface {
	GetById(ctx context.Context, id int) (*entity.Product, error)
}

type MovementService struct {
	saver 			MovementSaver
	provider		MovementProvider
	productProvider	ProductProvider
}

func New(rs MovementSaver, rp MovementProvider, pp ProductProvider) *MovementService {
	return &MovementService{rs, rp, pp}
}

func (ps *MovementService) GetAll(ctx context.Context) ([]entity.Movement, error) {
	return ps.provider.GetAll(ctx)
}

func (ps *MovementService) GetByFilter(ctx context.Context, filter entity.MovementFilter) ([]entity.Movement, error) {
	return ps.provider.GetByFilter(ctx, filter)
}