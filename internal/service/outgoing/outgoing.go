package outgoing

import (
	"context"
	"errors"
	"log"

	"vkr/internal/entity"
	"vkr/internal/entity/document"
	"vkr/internal/entity/document/outgoing"
	
	repos "vkr/internal/repository/postgres"
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
	Remove(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity float64) error
}

type OutgoingDocumentService struct {
	txManager 				TxManager
	f						*repos.RepositoryFactory
	productProvider			ProductProvider
	warehouseProvider		WarehouseProvider
	counterpartyProvider	CounterpartyProvider
	stockService			StockService
}

func New(
	txManager TxManager, 
	f *repos.RepositoryFactory,
	pp ProductProvider, 
	wp WarehouseProvider, 
	cp CounterpartyProvider,
	ss StockService,
) *OutgoingDocumentService {
	return &OutgoingDocumentService{txManager, f, pp, wp, cp, ss}
}

func (s *OutgoingDocumentService) Add(ctx context.Context, vo outgoing.UpsertOutgoingDocumentVO) error {
	buyer, err := s.counterpartyProvider.GetByID(ctx, vo.CounterPartyID)
	if err != nil {
		if errors.Is(err, entity.ErrCounterpartyNotFound) {
			return outgoing.ErrSupplierNotFound
		}
		return err
	}

	if buyer.Role != string(entity.Buyer) {
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

			log.Printf("OutgoingDocumentService::Add productProvider.GetByID Error - %v", err)
			return err
		}
	}

	err = s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		oRepo := s.f.NewOutgoingRepository(txCtx)
		mRepo := s.f.NewMovementRepository(txCtx)
		document, err := oRepo.Create(txCtx, vo)
		if err != nil {
			log.Printf("OutgoingDocumentService::Add oRepo.Add Error - %v", err)
			return err
		}

		for _, item := range vo.Items {
			err := s.stockService.Remove(txCtx, document.ToDocument(), item.ProductID, vo.WarehouseID, item.Quantity)
			if err != nil {
				log.Printf("OutgoingDocumentService::Add stockService.Remove Error - %v", err)
				return err				
			}

			mCollectiom,err := mRepo.GetByFilter(txCtx, entity.MovementFilter{
				DocumentID: document.ID, 
				DocumentType: string(document.ToDocument().Type),
				ProductID: item.ProductID,
				WarehouseID: vo.WarehouseID,
			})
			if err != nil {
				log.Printf("OutgoingDocumentService::Add movements.GetByFilter Error - %v", err)
				return err				
			}

			var fullMovementCost float64
			var movedCount float64
			for _, m := range mCollectiom {
				movedCount += m.StockMovement
				fullMovementCost += m.UnitCost * m.StockMovement
			}

			totalUnitCost := fullMovementCost / movedCount 
			err = oRepo.SetUnitCost(txCtx, document.ID, item.ProductID, totalUnitCost)
			if err != nil {
				log.Printf("OutgoingDocumentService::Add oRepo.SetUnitCost Error - %v", err)
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
