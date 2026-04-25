package outgoing

import (
	"context"
	"errors"
	"log"

	"vkr/internal/entity"
	"vkr/internal/entity/document"
	"vkr/internal/entity/document/outgoing"
	repo "vkr/internal/repository/postgres/document/outgoing"
)

type TxManager interface {
	RunInTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

type OutgoingDocumentSaver interface {
	Add(ctx context.Context, vo outgoing.UpsertOutgoingDocumentVO) (*outgoing.OutgoingDocument, error)
}

type OutgoingDocumentProvider interface{
	GetAll(ctx context.Context) ([]outgoing.OutgoingDocument, error)
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
	Remove(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity float32) error
}

type RepoFactory interface {
	NewOutgoingRepository(ctx context.Context) *repo.OutgoingRepository
}

type OutgoingDocumentService struct {
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
) *OutgoingDocumentService {
	return &OutgoingDocumentService{txManager, f, pp, wp, cp, ss}
}

func (s *OutgoingDocumentService) Add(ctx context.Context, vo outgoing.UpsertOutgoingDocumentVO) error {
	buyer, err := s.counterpartyProvider.GetById(ctx, vo.CounterPartyID)
	if err != nil {
		if errors.Is(err, entity.ErrCounterpartyNotFound) {
			return outgoing.ErrBuyerNotFound
		}
		return err
	}

	if buyer.Role != string(entity.Buyer) {
		return outgoing.ErrInvalidBuyer
	}

	_, err = s.warehouseProvider.GetById(ctx, vo.WarehouseID)
	if err != nil {
		return err
	}

	for _, rawMaterial := range vo.Items {
		finished, err := s.productProvider.GetById(ctx, rawMaterial.FinishedMaterialID)
		if err != nil {
			if errors.Is(err, entity.ErrProductNotFound) {
				return entity.ErrFinishedProductNotFound
			}

			if entity.ProductType(finished.TypeName) != entity.Finished {
				return entity.ErrInvalidFinishedMaterial
			}

			log.Printf("OutgoingDocumentService::Add productProvider.GetById Error - %v", err)
			return err
		}
	}

	err = s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		iRepo := s.f.NewOutgoingRepository(txCtx)
		document, err := iRepo.Add(txCtx, vo)
		if err != nil {
			log.Printf("OutgoingDocumentService::Add iRepo.Add Error - %v", err)
			return err
		}

		for _, rawMaterial := range vo.Items {
			err := s.stockService.Remove(txCtx, document.ToDocument(), rawMaterial.FinishedMaterialID, vo.WarehouseID, rawMaterial.Quantity)
			if err != nil {
				log.Printf("OutgoingDocumentService::Add stockService.Remove Error - %v", err)
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

func (s *OutgoingDocumentService) GetAll(ctx context.Context) ([]outgoing.OutgoingDocument, error) {
	repo := s.f.NewOutgoingRepository(ctx)
	return repo.GetAll(ctx)
}
