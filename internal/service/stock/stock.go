package stock

import (
	"context"

	"vkr/internal/entity"
	"vkr/internal/entity/document"
)

type TxManager interface {
	RunInTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

type StockRepository interface {
    GetAll(ctx context.Context) ([]entity.Stock, error)
    GetByFilter(ctx context.Context, filter entity.StockFilter) ([]entity.Stock, error)
    Increase(ctx context.Context, productID, warehouseID int, quantity float32) error
    Decrease(ctx context.Context, productID, warehouseID int, quantity float32) error
}

type MovingRepository interface {
	RegisterIncoming(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity float32) (*entity.Movement, error)
	RegisterOutgoing(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity float32) (*entity.Movement, error)
}

type ProductProvider interface {
	GetById(ctx context.Context, id int) (*entity.Product, error)
}

type StockService struct {
	txManager		TxManager
	stockRepo		StockRepository
	movingkRepo		MovingRepository
	productProvider	ProductProvider
}

func New(tx TxManager, sr StockRepository, mr MovingRepository, pp ProductProvider) *StockService {
	return &StockService{tx, sr, mr, pp}
}

func (ps *StockService) GetAll(ctx context.Context) ([]entity.Stock, error) {
	return ps.stockRepo.GetAll(ctx)
}

func (ps *StockService) GetByFilter(ctx context.Context, filter entity.StockFilter) ([]entity.Stock, error) {
	return ps.stockRepo.GetByFilter(ctx, filter)
}

func (ss *StockService) Add(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity float32) error {
	if quantity <= 0 {
		return entity.ErrInvalidQuantity
	}

	err := ss.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		err := ss.stockRepo.Increase(txCtx, product_id, warehouse_id, quantity)
		if err != nil {
			return err
		}

		ss.movingkRepo.RegisterIncoming(txCtx, docVO, product_id, warehouse_id, quantity)

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (ss *StockService) Remove(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity float32) error {
	if quantity <= 0 {
		return entity.ErrInvalidQuantity
	}

	stocks, err := ss.stockRepo.GetByFilter(ctx, entity.StockFilter{ProductID: product_id, WarehouseID: warehouse_id})
	if err != nil {
		return err
	}

	if len(stocks) == 0 {
		return entity.ErrStockNotFound
	}

	stock := stocks[0]

	if stock.Quantity <= quantity {
		return entity.ErrInsufficientStock
	}

	err = ss.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		err := ss.stockRepo.Decrease(txCtx, product_id, warehouse_id, quantity)
		if err != nil {
			return err
		}

		ss.movingkRepo.RegisterOutgoing(txCtx, docVO, product_id, warehouse_id, quantity)

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}