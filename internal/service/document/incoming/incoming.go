package incoming

import (
	"context"
	"errors"
	"log"

	"vkr/internal/entity"
	"vkr/internal/entity/document"
	"vkr/internal/entity/document/incoming"
	repo "vkr/internal/repository/postgres/document/incoming"
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
	GetById(ctx context.Context, id int) (*entity.Product, error)
}

type WarehouseProvider interface {
	GetById(ctx context.Context, id int) (*entity.Warehouse, error)
}

type CounterpartyProvider interface {
	GetById(ctx context.Context, id int) (*entity.Counterparty, error)
}

type StockService interface {
	Add(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity float32) error
}

type RepoFactory interface {
	NewIncomingRepository(ctx context.Context) *repo.IncomingRepository
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
	buyer, err := s.counterpartyProvider.GetById(ctx, vo.CounterPartyID)
	if err != nil {
		if errors.Is(err, entity.ErrCounterpartyNotFound) {
			return incoming.ErrSupplierNotFound
		}
		return err
	}

	if buyer.Role != string(entity.Supplier) {
		return incoming.ErrInvalidSupplier
	}

	_, err = s.warehouseProvider.GetById(ctx, vo.WarehouseID)
	if err != nil {
		return err
	}

	for _, rawMaterial := range vo.Items {
		raw, err := s.productProvider.GetById(ctx, rawMaterial.RawMaterialID)
		if err != nil {
			if errors.Is(err, entity.ErrProductNotFound) {
				return entity.ErrRawProductNotFound
			}

			if entity.ProductType(raw.TypeName) != entity.Raw {
				return entity.ErrInvalidRawMaterial
			}

			log.Printf("IncomingDocumentService::Add productProvider.GetById Error - %v", err)
			return err
		}
	}

	err = s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		iRepo := s.f.NewIncomingRepository(txCtx)
		document, err := iRepo.Add(txCtx, vo)
		if err != nil {
			log.Printf("IncomingDocumentService::Add iRepo.Add Error - %v", err)
			return err
		}

		for _, rawMaterial := range vo.Items {
			err := s.stockService.Add(txCtx, document.ToDocument(), rawMaterial.RawMaterialID, vo.WarehouseID, rawMaterial.Quantity)
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

func (s *IncomingDocumentService) GetAll(ctx context.Context) ([]incoming.IncomingDocument, error) {
	repo := s.f.NewIncomingRepository(ctx)
	return repo.GetAll(ctx)
}
