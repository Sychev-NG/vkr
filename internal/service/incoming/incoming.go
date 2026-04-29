package incoming

import (
	"context"
	"errors"
	"log"

	"vkr/internal/entity"
	"vkr/internal/entity/document"
	"vkr/internal/entity/document/incoming"
	iRepo "vkr/internal/repository/postgres/incoming"
)

type TxManager interface {
	RunInTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

type IncomingDocumentSaver interface {
	Add(ctx context.Context, vo incoming.UpsertIncomingDocumentVO) (*incoming.IncomingDocument, error)
}

type IncomingDocumentProvider interface{
	GetAll(ctx context.Context) ([]incoming.IncomingDocument, error)
}

type ProductProvider interface {
	GetByID(ctx context.Context, id int) (*entity.Product, error)
}

type WarehouseProvider interface {
	GetByID(ctx context.Context, id int) (*entity.Warehouse, error)
}

type CounterpartyProvider interface {
	GetByID(ctx context.Context, id int) (*entity.Counterparty, error)
}

type StockService interface {
	Add(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity, unit_cost float64) error
}

type RepoFactory interface {
	NewIncomingRepository(ctx context.Context) *iRepo.IncomingRepository
}

type IncomingDocumentService struct {
	txManager 				TxManager
	f						RepoFactory
	productProvider			ProductProvider
	warehouseProvider		WarehouseProvider
	counterpartyProvider	CounterpartyProvider
	stockService			StockService
}

func New(
	txManager TxManager, 
	f RepoFactory,
	pp ProductProvider, 
	wp WarehouseProvider, 
	cp CounterpartyProvider,
	ss StockService,
) *IncomingDocumentService {
	return &IncomingDocumentService{txManager, f, pp, wp, cp, ss}
}

func (s *IncomingDocumentService) Add(ctx context.Context, vo incoming.UpsertIncomingDocumentVO) error {
	supplier, err := s.counterpartyProvider.GetByID(ctx, vo.CounterPartyID)
	if err != nil {
		if errors.Is(err, entity.ErrCounterpartyNotFound) {
			return incoming.ErrSupplierNotFound
		}
		return err
	}

	if supplier.Role != string(entity.Supplier) {
		return entity.ErrInvalidCounterpartyRole
	}

	_, err = s.warehouseProvider.GetByID(ctx, vo.WarehouseID)
	if err != nil {
		return err
	}

	for _, items := range vo.Items {
		_, err := s.productProvider.GetByID(ctx, items.ProductID)
		if err != nil {
			if errors.Is(err, entity.ErrProductNotFound) {
				return entity.ErrProductNotFound
			}

			log.Printf("IncomingDocumentService::Add productProvider.GetByID Error - %v", err)
			return err
		}
	}

	err = s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		iRepo := s.f.NewIncomingRepository(txCtx)
		document, err := iRepo.Create(txCtx, vo)
		if err != nil {
			log.Printf("IncomingDocumentService::Add iRepo.Add Error - %v", err)
			return err
		}

		for _, item := range vo.Items {
			err := s.stockService.Add(txCtx, document.ToDocument(), item.ProductID, vo.WarehouseID, item.Quantity, item.Price)
			if err != nil {
				log.Printf("IncomingDocumentService::Add stockService.Add Error - %v", err)
				return err				
			}
		}

        return nil
    })

	if err != nil {
		return err
	}

	return nil
}
