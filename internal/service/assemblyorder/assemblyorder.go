package assemblyorder

import (
	"context"
	// "errors"
	// "fmt"
	"log"

	// "time"

	"vkr/internal/entity"
	"vkr/internal/entity/document"
	docAssembly "vkr/internal/entity/document/assembly"
	"vkr/internal/repository/postgres"
)

type TxManager interface {
	RunInTx(ctx context.Context, fn func(txCtx context.Context) error) error
}

type AssemblyProvider interface {
	GetByID(ctx context.Context, id int) (*entity.Assembly, error)
}

type WarehouseProvider interface {
	GetByID(ctx context.Context, id int) (*entity.Warehouse, error)
}

type ProductProvider interface {
	GetByID(ctx context.Context, id int) (*entity.Product, error)
}

type StockService interface {
	Remove(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity float64) error
	Add(ctx context.Context, docVO document.Document, product_id, warehouse_id int, quantity, unit_cost float64) error
	GetAvailableQuantity(ctx context.Context, product_id, warehouse_id int) (float64, error)
}

type AssemblyOrderService struct {
	txManager         TxManager
	f                 *postgres.RepositoryFactory
	assemblyProvider  AssemblyProvider
	warehouseProvider WarehouseProvider
	productProvider   ProductProvider
	stockService      StockService
}

func New(
	txManager TxManager,
	f *postgres.RepositoryFactory,
	ap AssemblyProvider,
	wp WarehouseProvider,
	pp ProductProvider,
	ss StockService,
) *AssemblyOrderService {
	return &AssemblyOrderService{
		txManager:         txManager,
		f:                 f,
		assemblyProvider:  ap,
		warehouseProvider: wp,
		productProvider:   pp,
		stockService:      ss,
	}
}

func (s *AssemblyOrderService) GetByID(ctx context.Context, id int) (*docAssembly.AssemblyOrder, error) {
	return s.f.NewAssemblyOrderRepository(ctx).GetByID(ctx, id)
}

func (s *AssemblyOrderService) GetAll(ctx context.Context) ([]docAssembly.AssemblyOrder, error) {
	return s.f.NewAssemblyOrderRepository(ctx).GetAll(ctx)
}

func (s *AssemblyOrderService) Create(ctx context.Context, vo docAssembly.UpsertAssemblyOrderVO) error {
	// 1. Проверяем существование сборки
	assembly, err := s.assemblyProvider.GetByID(ctx, vo.AssemblyID)
	if err != nil {
		log.Printf("AssemblyOrderService::Create GetByID Error - %v", err)
		return err
	}

	// 2. Проверяем существование склада
	_, err = s.warehouseProvider.GetByID(ctx, vo.WarehouseID)
	if err != nil {
		log.Printf("AssemblyOrderService::Create GetWarehouse Error - %v", err)
		return err
	}

	// 3. Проверяем наличие компонентов и доступное количество
	for _, component := range assembly.Components {
		requiredQty := component.Quantity * vo.QuantityToBuild
		
		available, err := s.stockService.GetAvailableQuantity(ctx, component.ProductID, vo.WarehouseID)
		if err != nil {
			log.Printf("AssemblyOrderService::Create GetAvailableQuantity Error - %v", err)
			return err
		}

		if available < requiredQty {
			return entity.ErrInsufficientComponents
		}
	}

	// 4. Выполняем операции в транзакции
	err = s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		repo := s.f.NewAssemblyOrderRepository(txCtx)
		mRepo := s.f.NewMovementRepository(txCtx)

		// 4.1 Создаем заказ на сборку
		order, err := repo.Create(txCtx, docAssembly.UpsertAssemblyOrderVO{
			AssemblyID:      vo.AssemblyID,
			WarehouseID:     vo.WarehouseID,
			QuantityToBuild: vo.QuantityToBuild,
		})
		if err != nil {
			log.Printf("AssemblyOrderService::Create Create Order Error - %v", err)
			return err
		}

		var totalComponentsCost float64

		// 4.2 Списываем компоненты
		for _, component := range assembly.Components {
			requiredQty := component.Quantity * vo.QuantityToBuild
			
			err = s.stockService.Remove(txCtx, order.ToDocument(), component.ProductID, vo.WarehouseID, requiredQty)
			if err != nil {
				log.Printf("AssemblyOrderService::Create Remove Component Error - %v", err)
				return err
			}

			mCollectiom,err := mRepo.GetByFilter(txCtx, entity.MovementFilter{
				DocumentID: order.ID, 
				DocumentType: string(order.ToDocument().Type),
				ProductID: component.ProductID,
				WarehouseID: vo.WarehouseID,
			})
			if err != nil {
				log.Printf("AssemblyOrderService::Create movements.GetByFilter Error - %v", err)
				return err				
			}

			var fullMovementCost float64
			var movedCount float64
			for _, m := range mCollectiom {
				movedCount += m.StockMovement
				fullMovementCost += m.UnitCost * m.StockMovement
			}

			totalUnitCost := fullMovementCost / movedCount 
			err = repo.AddConsumption(txCtx, order.ID, component.ProductID, requiredQty, totalUnitCost, fullMovementCost)
			if err != nil {
				log.Printf("AssemblyOrderService::Create AddConsumption Error - %v", err)
				return err
			}

			totalComponentsCost += fullMovementCost
		}

		// 4.3 Оприходуем готовую продукцию (как в IncomingDocument)
		outputQty := assembly.OutputQuantity * vo.QuantityToBuild
		
		unitCost := totalComponentsCost / outputQty
		totalCost := totalComponentsCost
		
		err = s.stockService.Add(txCtx, order.ToDocument(), assembly.OutputProductID, vo.WarehouseID, outputQty, unitCost)
		if err != nil {
			log.Printf("AssemblyOrderService::Create Add Output Error - %v", err)
			return err
		}
		
		err = repo.AddOutput(txCtx, order.ID, assembly.OutputProductID, outputQty, unitCost, totalCost)
		if err != nil {
			log.Printf("AssemblyOrderService::Create AddOutput Error - %v", err)
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("AssemblyOrderService::Create Transaction Error - %v", err)
		return err
	}

	return nil
}

// func (s *AssemblyOrderService) CheckAvailability(ctx context.Context, assemblyID, warehouseID int, quantityToBuild float64) ([]ComponentRequirement, error) {
// 	assembly, err := s.assemblyProvider.GetByID(ctx, assemblyID)
// 	if err != nil {
// 		if errors.Is(err, entity.ErrAssemblyNotFound) {
// 			return nil, docAssembly.ErrAssemblyNotFound
// 		}
// 		return nil, err
// 	}

// 	requirements := make([]ComponentRequirement, 0, len(assembly.Components))
// 	for _, component := range assembly.Components {
// 		requiredQty := component.Quantity * quantityToBuild
		
// 		available, err := s.stockService.GetAvailableQuantity(ctx, component.ProductID, warehouseID)
// 		if err != nil {
// 			return nil, err
// 		}

// 		requirements = append(requirements, ComponentRequirement{
// 			ProductID: component.ProductID,
// 			Required:  requiredQty,
// 			Available: available,
// 		})
// 	}

// 	return requirements, nil
// }