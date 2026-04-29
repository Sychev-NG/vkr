package product

import (
	"context"
	"strings"
	"vkr/internal/entity"
)

type ProductSaver interface {
	Create(ctx context.Context, name, unit string, minStock float64) (*entity.Product, error)
	Update(ctx context.Context, id int, name, unit string, minStock float64) (*entity.Product, error)
	Delete(ctx context.Context, id int) (error)
}

type ProductProvider interface{
	GetByID(ctx context.Context, id int) (*entity.Product, error)
	GetAll(ctx context.Context) ([]entity.Product, error)
}

type ProductService struct {
	saver 		ProductSaver
	provider	ProductProvider 
}

func New(ps ProductSaver, pp ProductProvider) *ProductService {
	return &ProductService{ps, pp}
}

func (ps *ProductService) Add(ctx context.Context, name, unit string, minStock float64) (*entity.Product, error) {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return nil, entity.ErrInvalidProductName
	}

	return ps.saver.Create(ctx, name, unit, minStock)	
}

func (ps *ProductService) GetByID(ctx context.Context, id int) (*entity.Product, error) {
	return ps.provider.GetByID(ctx, id)
}

func (ps *ProductService) GetAll(ctx context.Context) ([]entity.Product, error) {
	return ps.provider.GetAll(ctx)
}

func (ps *ProductService) Update(ctx context.Context, id int, name, unit string, minStock float64) (*entity.Product, error) {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return nil, entity.ErrInvalidProductName
	}

	return ps.saver.Update(ctx, id, name, unit, minStock)	
}

func (ps *ProductService) Delete(ctx context.Context, id int) (error) {
	return ps.saver.Delete(ctx, id)
}

