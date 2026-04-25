package stock

import (
	"context"
	"log"

	"vkr/internal/entity"
	"vkr/internal/entity/document"
	mRepo "vkr/internal/repository/postgres/movement"
	sRepo "vkr/internal/repository/postgres/stock"
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

type RepoFactory interface {
	NewStockRepository(ctx context.Context) *sRepo.StockRepository
	NewMovementRepository(ctx context.Context) *mRepo.MovementRepository
}

type StockService struct {
	txManager		TxManager
	f 				RepoFactory
	productProvider	ProductProvider
}

func New(tx TxManager, f RepoFactory, pp ProductProvider) *StockService {
	return &StockService{tx, f, pp}
}

func (ps *StockService) GetAll(ctx context.Context) ([]entity.Stock, error) {
	sRepo := ps.f.NewStockRepository(ctx)
	return sRepo.GetAll(ctx)
}

func (ps *StockService) GetByFilter(ctx context.Context, filter entity.StockFilter) ([]entity.Stock, error) {
	sRepo := ps.f.NewStockRepository(ctx)
	return sRepo.GetByFilter(ctx, filter)
}

func (ps *StockService) GetByProductAndWarehouseId(ctx context.Context, product_id, warhouse_id int) (*entity.Stock, error) {
	sRepo := ps.f.NewStockRepository(ctx)
	return sRepo.GetByProductAndWarehouseId(ctx, product_id, warhouse_id)
}

func (ss *StockService) Add(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity float32) error {
	if quantity <= 0 {
		return entity.ErrInvalidQuantity
	}

	err := ss.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		sRepo := ss.f.NewStockRepository(txCtx)
		mRepo := ss.f.NewMovementRepository(txCtx)

		err := sRepo.Increase(txCtx, product_id, warehouse_id, quantity)
		if err != nil {
			log.Printf("StockService::Add Increase Error - %v", err.Error())
			return err
		}

		_, err = mRepo.RegisterIncoming(txCtx, docVO, product_id, warehouse_id, quantity)
		if err != nil {
			log.Printf("StockService::Add RegisterIncoming Error - %v", err.Error())
			return err
		}

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

	sRepo := ss.f.NewStockRepository(ctx)
	stocks, err := sRepo.GetByFilter(ctx, entity.StockFilter{ProductID: product_id, WarehouseID: warehouse_id})
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
		sRepo := ss.f.NewStockRepository(txCtx)
		mRepo := ss.f.NewMovementRepository(txCtx)

		err := sRepo.Decrease(txCtx, product_id, warehouse_id, quantity)
		if err != nil {
			return err
		}

		mRepo.RegisterOutgoing(txCtx, docVO, product_id, warehouse_id, quantity)

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}