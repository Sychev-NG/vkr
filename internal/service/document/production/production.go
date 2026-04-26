package production

import (
	"context"
	"log"

	"vkr/internal/entity"
	"vkr/internal/entity/document"
	"vkr/internal/entity/document/production"
	repo "vkr/internal/repository/postgres/document/production"
)

type TxManager interface {
	RunInTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

type ProductionDocumentSaver interface {
	Add(ctx context.Context, vo production.UpsertProductionDocumentVO) (*production.ProductionDocument, error)
}

type ProductionDocumentProvider interface{
	GetAll(ctx context.Context) ([]production.ProductionDocument, error)
}

type ProductProvider interface {
	GetById(ctx context.Context, id int) (*entity.Product, error)
}

type WarehouseProvider interface {
	GetById(ctx context.Context, id int) (*entity.Warehouse, error)
}

type StockService interface {
	Add(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity float32) error
	Remove(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity float32) error
	GetByProductAndWarehouseId(ctx context.Context, product_id, warhouse_id int) (*entity.Stock, error)
}

type RecipeProvider interface {
	GetById(ctx context.Context, id int) (*entity.Recipe, error)
}

type RepoFactory interface {
	NewProductionRepository(ctx context.Context) *repo.ProductionRepository
}

type ProductionDocumentService struct {
	txManager 				TxManager
	f						RepoFactory
	productProvider			ProductProvider
	warehouseProvider		WarehouseProvider
	recipeProvider			RecipeProvider
	stockService			StockService
}

func New(
	txManager TxManager, 
	f RepoFactory,
	pp ProductProvider, 
	wp WarehouseProvider, 
	rp RecipeProvider, 
	ss StockService,
) *ProductionDocumentService {
	return &ProductionDocumentService{txManager, f, pp, wp, rp, ss}
}

func (s *ProductionDocumentService) Add(ctx context.Context, vo production.UpsertProductionDocumentVO) error {
	_, err := s.warehouseProvider.GetById(ctx, vo.WarehouseID)
	if err != nil {
		return err
	}
	
	recpe, err := s.recipeProvider.GetById(ctx, vo.RecipeID)
	if err != nil {
		return err
	}

	err = s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		pRepo := s.f.NewProductionRepository(txCtx)
		document, err := pRepo.Add(txCtx, vo)
		if err != nil {
			log.Printf("ProductionDocumentService::Add pRepo.Add Error - %v", err)
			return err
		}

		log.Printf("ProductionDocumentService::Add txCtx - %v", txCtx)

		for _, ingredient := range recpe.Ingredients {
			err := s.stockService.Remove(txCtx, document.ToDocument(), ingredient.RawMaterialID, vo.WarehouseID, vo.Quantity * ingredient.QuantityPerUnit)
			if err != nil {
				log.Printf("ProductionDocumentService::Add stockService.Remove Error - %v", err)
				return err
			}		
		}
		
		log.Printf("ProductionDocumentService::Add txCtx - %v", txCtx)

		err = s.stockService.Add(txCtx, document.ToDocument(), recpe.ProductID, vo.WarehouseID, vo.Quantity)
		if err != nil {
			log.Printf("ProductionDocumentService::Add stockService.Add Error - %v", err)
			return err
		}

        return nil
    })

	if err != nil {
		log.Printf("ProductionDocumentService::Add RunInTx Error - %v", err)
		return err
	}

	return nil
}

func (s *ProductionDocumentService) GetAll(ctx context.Context) ([]production.ProductionDocument, error) {
	repo := s.f.NewProductionRepository(ctx)
	return repo.GetAll(ctx)
}
