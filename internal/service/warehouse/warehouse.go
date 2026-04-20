package warehouse

import (
	"context"
	"strings"
	"vkr/internal/entity"
)

type WarehouseSaver interface {
	Add(ctx context.Context, name, address string) (*entity.Warehouse, error)
	Update(ctx context.Context, id int, name, address string) (*entity.Warehouse, error)
	Delete(ctx context.Context, id int) (error)
}

type WarehouseProvider interface{
	GetById(ctx context.Context, id int) (*entity.Warehouse, error)
	GetAll(ctx context.Context) ([]entity.Warehouse, error)
}

type WarehouseService struct {
	saver 		WarehouseSaver
	provider	WarehouseProvider 
}

func New(ps WarehouseSaver, pp WarehouseProvider) *WarehouseService {
	return &WarehouseService{ps, pp}
}

func (ps *WarehouseService) Add(ctx context.Context, name, address string) (*entity.Warehouse, error) {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return nil, entity.ErrInvalidWarehouseName
	}

	address = strings.TrimSpace(address)
	if len(address) == 0 {
		return nil, entity.ErrInvalidWarehouseAddress
	}

	return ps.saver.Add(ctx, name, address)	
}

func (ps *WarehouseService) GetById(ctx context.Context, id int) (*entity.Warehouse, error) {
	return ps.provider.GetById(ctx, id)
}

func (ps *WarehouseService) GetAll(ctx context.Context) ([]entity.Warehouse, error) {
	return ps.provider.GetAll(ctx)
}

func (ps *WarehouseService) Update(ctx context.Context, id int, name, address string) (*entity.Warehouse, error) {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return nil, entity.ErrInvalidWarehouseName
	}

	address = strings.TrimSpace(address)
	if len(address) == 0 {
		return nil, entity.ErrInvalidWarehouseAddress
	}

	return ps.saver.Update(ctx, id, name, address)	
}

func (ps *WarehouseService) Delete(ctx context.Context, id int) (error) {
	return ps.saver.Delete(ctx, id)
}