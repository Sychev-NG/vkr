package product

import (
	"context"
	"vkr/internal/entity"
)

type ProductSaver interface {
	Add(ctx context.Context, name, unit, productType string) (*entity.Product, error)
	Update(ctx context.Context, product entity.Product) (*entity.Product, error)
	Delete(ctx context.Context, id int) (error)
}

type ProductProvider interface{
	GetById(ctx context.Context, id int) (*entity.Product, error)
	GetAll(ctx context.Context) ([]entity.Product, error)
}

type ProductService struct {
	saver 		ProductSaver
	provider	ProductProvider 
}

func New(ps ProductSaver, pp ProductProvider) *ProductService {
	return &ProductService{ps, pp}
}

func (ps *ProductService) Add(ctx context.Context, name, unit, productType string) (*entity.Product, error) {
	pt := entity.ProductType(productType)
	if pt != entity.Raw && pt != entity.Finished {
		return nil, entity.ErrInvalidProductType
	}

	pu := entity.ProductUnit(unit)
	if pu != entity.KG {
		return nil, entity.ErrInvalidProductUnit
	}

	return ps.saver.Add(ctx, name, unit, productType)	
}

func (ps *ProductService) GetById(ctx context.Context, id int) (*entity.Product, error) {
	return ps.provider.GetById(ctx, id)
}

func (ps *ProductService) GetAll(ctx context.Context) ([]entity.Product, error) {
	return ps.provider.GetAll(ctx)
}

func (ps *ProductService) Update(ctx context.Context, id int, name, unit, productType string) (*entity.Product, error) {
	return nil, nil
	// return ps.saver.Update()
}

func (ps *ProductService) Delete(ctx context.Context, id int) (error) {
	return ps.saver.Delete(ctx, id)
}