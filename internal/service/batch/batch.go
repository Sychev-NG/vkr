package stock

import (
	"context"
	"log"
	"math"

	"vkr/internal/entity"
	"vkr/internal/entity/document"
	bRepo "vkr/internal/repository/postgres/batch"
	bmRepo "vkr/internal/repository/postgres/batch_movement"
	sRepo "vkr/internal/repository/postgres/stock"
)

type TxManager interface {
	RunInTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

type ProductProvider interface {
	GetById(ctx context.Context, id int) (*entity.Product, error)
}

type RepoFactory interface {
	NewStockRepository(ctx context.Context) *sRepo.StockRepository
	NewBatchRepository(ctx context.Context) *bRepo.BatchRepository
	NewBatchMovementRepository(ctx context.Context) *bmRepo.BatchMovementRepository
}

type StockService struct {
	txManager		TxManager
	f 				RepoFactory
}

func New(tx TxManager, f RepoFactory) *StockService {
	return &StockService{tx, f}
}

// func (ps *StockService) GetAll(ctx context.Context) ([]entity.Stock, error) {
// 	sRepo := ps.f.NewStockRepository(ctx)
// 	return sRepo.GetAll(ctx)
// }

func (ps *StockService) GetByFilter(ctx context.Context, filter entity.StockFilter) ([]entity.Stock, error) {
	sRepo := ps.f.NewStockRepository(ctx)
	return sRepo.GetByFilter(ctx, filter)
}

// func (ps *StockService) GetByProductAndWarehouseId(ctx context.Context, product_id, warhouse_id int) (*entity.Stock, error) {
// 	sRepo := ps.f.NewStockRepository(ctx)
// 	result, err := sRepo.GetByProductAndWarehouseId(ctx, product_id, warhouse_id)
// 	if err != nil {
// 		log.Printf("StockService::GetByProductAndWarehouseId Error - %v", err.Error())
// 	}
// 	return result, err
// }

func (ss *StockService) Add(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity float64) error {
	if quantity <= 0 {
		return entity.ErrInvalidQuantity
	}

	err := ss.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		sRepo := ss.f.NewStockRepository(txCtx)
		bRepo := ss.f.NewBatchRepository(txCtx)
		bmRepo := ss.f.NewBatchMovementRepository(txCtx)

		batch, err := bRepo.Create(txCtx, entity.UpsertBatchVO{
			ProductID: product_id,
			WarehouseID: warehouse_id,
			DocumentType: string(docVO.Type),
			DocumentID: docVO.DocumentID,
			QuantityRemaining: quantity,
		})
		if err != nil {
			log.Printf("StockService::Add Create Error - %v", err.Error())
			return err
		}

		_, err = bmRepo.RegisterIncoming(txCtx, docVO, batch.ID, quantity)
		if err != nil {
			log.Printf("StockService::Add RegisterIncoming Error - %v", err.Error())
			return err
		}

		err = sRepo.Increase(txCtx, product_id, warehouse_id, quantity)
		if err != nil {
			log.Printf("StockService::Add Increase Error - %v", err.Error())
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (ss *StockService) Remove(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity float64) error {
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

	if stock.Quantity < quantity {
		return entity.ErrInsufficientStock
	}

	err = ss.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		sRepo := ss.f.NewStockRepository(txCtx)
		bRepo := ss.f.NewBatchRepository(txCtx)
		bmRepo := ss.f.NewBatchMovementRepository(txCtx)

		batchCollection, err := bRepo.GetBatchesForQuantity(ctx, product_id,  warehouse_id, quantity)
		if err != nil {
			log.Printf("StockService::Remove GetBatchesForQuantity Error - %v", err.Error())
			return err
		}

		remaining := quantity
		for _, batch := range batchCollection {
			toSubtract := math.Min(remaining, batch.QuantityRemaining)

			err := bRepo.Subtract(txCtx, batch.ID, toSubtract)
			if err != nil {
				log.Printf("StockService::Remove Subtract Error - %v", err.Error())
				return err
			}

			_, err = bmRepo.RegisterOutgoing(txCtx, docVO, batch.ID, toSubtract)
			if err != nil {
				log.Printf("StockService::Remove RegisterOutgoing Error - %v", err.Error())
				return err
			}

			remaining -= toSubtract

			if remaining <= 0 {
				break
			}			
		}

		err = sRepo.Decrease(txCtx, product_id, warehouse_id, quantity)
		if err != nil {
			log.Printf("StockService::Remove Decrease Error - %v", err.Error())
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}