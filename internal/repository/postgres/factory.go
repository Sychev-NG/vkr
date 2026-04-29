package postgres

import (
	"context"
	storage "vkr/internal/storage/postgres"

	pRepo "vkr/internal/repository/postgres/product"
	cpRepo "vkr/internal/repository/postgres/counterparty"
	wRepo "vkr/internal/repository/postgres/warehouse"
	aRepo "vkr/internal/repository/postgres/assembly"
	// // acRepo "vkr/internal/repository/postgres/assembly_component"
	aoRepo "vkr/internal/repository/postgres/assemblyorder"
	// aocRepo "vkr/internal/repository/postgres/assembly_order_consumption"
	// aooRepo "vkr/internal/repository/postgres/assembly_order_output"
	sRepo "vkr/internal/repository/postgres/stock"
	bRepo "vkr/internal/repository/postgres/batch"
	bmRepo "vkr/internal/repository/postgres/batch_movement"
	mRepo "vkr/internal/repository/postgres/movement"
	// alRepo "vkr/internal/repository/postgres/alert"
	iRepo "vkr/internal/repository/postgres/incoming"
	// iiRepo "vkr/internal/repository/postgres/incoming_item"
	oRepo "vkr/internal/repository/postgres/outgoing"
	// oiRepo "vkr/internal/repository/postgres/outgoing_item"
	// mvRepo "vkr/internal/repository/postgres/movement" // VIEW movements

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryFactory struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *RepositoryFactory {
	return &RepositoryFactory{pool: pool}
}

// ========== СПРАВОЧНИКИ ==========

func (f *RepositoryFactory) NewProductRepository(ctx context.Context) *pRepo.ProductRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return pRepo.New(tx)
	}
	return pRepo.New(f.pool)
}

func (f *RepositoryFactory) NewCounterpartyRepository(ctx context.Context) *cpRepo.CounterpartyRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return cpRepo.New(tx)
	}
	return cpRepo.New(f.pool)
}

func (f *RepositoryFactory) NewWarehouseRepository(ctx context.Context) *wRepo.WarehouseRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return wRepo.New(tx)
	}
	return wRepo.New(f.pool)
}

func (f *RepositoryFactory) NewAssemblyRepository(ctx context.Context) *aRepo.AssemblyRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return aRepo.New(tx)
	}
	return aRepo.New(f.pool)
}

// func (f *RepositoryFactory) NewAssemblyComponentRepository(ctx context.Context) *acRepo.AssemblyComponentRepository {
// 	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
// 		return acRepo.New(tx)
// 	}
// 	return acRepo.New(f.pool)
// }

// // ========== ДОКУМЕНТЫ ==========

func (f *RepositoryFactory) NewIncomingRepository(ctx context.Context) *iRepo.IncomingRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return iRepo.New(tx)
	}
	return iRepo.New(f.pool)
}

// func (f *RepositoryFactory) NewIncomingItemRepository(ctx context.Context) *iiRepo.IncomingItemRepository {
// 	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
// 		return iiRepo.New(tx)
// 	}
// 	return iiRepo.New(f.pool)
// }

func (f *RepositoryFactory) NewOutgoingRepository(ctx context.Context) *oRepo.OutgoingRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return oRepo.New(tx)
	}
	return oRepo.New(f.pool)
}

// func (f *RepositoryFactory) NewOutgoingItemRepository(ctx context.Context) *oiRepo.OutgoingItemRepository {
// 	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
// 		return oiRepo.New(tx)
// 	}
// 	return oiRepo.New(f.pool)
// }

func (f *RepositoryFactory) NewAssemblyOrderRepository(ctx context.Context) *aoRepo.AssemblyOrderRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return aoRepo.New(tx)
	}
	return aoRepo.New(f.pool)
}

// func (f *RepositoryFactory) NewAssemblyOrderConsumptionRepository(ctx context.Context) *aocRepo.AssemblyOrderConsumptionRepository {
// 	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
// 		return aocRepo.New(tx)
// 	}
// 	return aocRepo.New(f.pool)
// }

// func (f *RepositoryFactory) NewAssemblyOrderOutputRepository(ctx context.Context) *aooRepo.AssemblyOrderOutputRepository {
// 	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
// 		return aooRepo.New(tx)
// 	}
// 	return aooRepo.New(f.pool)
// }

// // ========== ПАРТИИ И ДВИЖЕНИЯ ==========

func (f *RepositoryFactory) NewBatchRepository(ctx context.Context) *bRepo.BatchRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return bRepo.New(tx)
	}
	return bRepo.New(f.pool)
}

func (f *RepositoryFactory) NewBatchMovementRepository(ctx context.Context) *bmRepo.BatchMovementRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return bmRepo.New(tx)
	}
	return bmRepo.New(f.pool)
}

func (f *RepositoryFactory) NewMovementRepository(ctx context.Context) *mRepo.MovementRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return mRepo.New(tx)
	}
	return mRepo.New(f.pool)
}

// // ========== РЕГИСТРЫ ==========

func (f *RepositoryFactory) NewStockRepository(ctx context.Context) *sRepo.StockRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return sRepo.New(tx)
	}
	return sRepo.New(f.pool)
}

// func (f *RepositoryFactory) NewMovementRepository(ctx context.Context) *mvRepo.MovementRepository {
// 	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
// 		return mvRepo.New(tx)
// 	}
// 	return mvRepo.New(f.pool)
// }

// func (f *RepositoryFactory) NewAlertRepository(ctx context.Context) *alRepo.AlertRepository {
// 	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
// 		return alRepo.New(tx)
// 	}
// 	return alRepo.New(f.pool)
// }