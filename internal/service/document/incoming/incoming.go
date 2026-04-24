package incoming

import (
	"context"
	"errors"

	"vkr/internal/entity"
	"vkr/internal/entity/document"
	"vkr/internal/entity/document/incoming"
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

type IncomingDocumentService struct {
	txManager 				TxManager
	saver 					IncomingDocumentSaver
	provider				IncomingDocumentProvider
	productProvider			ProductProvider
	warehouseProvider		WarehouseProvider
	counterpartyProvider	CounterpartyProvider
	stockService			StockService
}

func New(
	txManager TxManager, 
	rs IncomingDocumentSaver, 
	rp IncomingDocumentProvider, 
	pp ProductProvider, 
	wp WarehouseProvider, 
	cp CounterpartyProvider,
	ss StockService,
) *IncomingDocumentService {
	return &IncomingDocumentService{txManager, rs, rp, pp, wp, cp, ss}
}

func (ps *IncomingDocumentService) Add(ctx context.Context, vo incoming.UpsertIncomingDocumentVO) error {
	_, err := ps.counterpartyProvider.GetById(ctx, vo.CounterPartyID)
	if err != nil {
		if errors.Is(err, entity.ErrCounterpartyNotFound) {
			return incoming.ErrSupplierNotFound
		}
		return err
	}

	_, err = ps.warehouseProvider.GetById(ctx, vo.CounterPartyID)
	if err != nil {
		return err
	}

	for _, rawMaterial := range vo.Items {
		_, err = ps.productProvider.GetById(ctx, rawMaterial.RawMaterialID)
		if err != nil {
			if errors.Is(err, entity.ErrProductNotFound) {
				return entity.ErrRawProductNotFound
			}
			return err
		}
	}

	err = ps.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		document, err := ps.saver.Add(txCtx, vo)
		if err != nil {
			return err
		}

		for _, rawMaterial := range vo.Items {
			err := ps.stockService.Add(txCtx, document.ToDocument(), rawMaterial.RawMaterialID, vo.WarehouseID, rawMaterial.Quantity)
			if err != nil {
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

func (ps *IncomingDocumentService) GetAll(ctx context.Context) ([]incoming.IncomingDocument, error) {
	return ps.provider.GetAll(ctx)
}
