package postgres

import (
	"context"
	storage "vkr/internal/storage/postgres"

	pRepo "vkr/internal/repository/postgres/product"
	cpRepo "vkr/internal/repository/postgres/counterparty"
	wRepo "vkr/internal/repository/postgres/warehouse"
	rRepo "vkr/internal/repository/postgres/recipe"
	sRepo "vkr/internal/repository/postgres/stock"
	mRepo "vkr/internal/repository/postgres/movement"
	iRepo "vkr/internal/repository/postgres/document/incoming"
	prRepo "vkr/internal/repository/postgres/document/production"
	oRepo "vkr/internal/repository/postgres/document/outgoing"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryFactory struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *RepositoryFactory {
	return &RepositoryFactory{pool: pool}
}

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

func (f *RepositoryFactory) NewRecipeRepository(ctx context.Context) *rRepo.RecipeRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return rRepo.New(tx)
	}
	return rRepo.New(f.pool)
}

func (f *RepositoryFactory) NewStockRepository(ctx context.Context) *sRepo.StockRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return sRepo.New(tx)
	}
	return sRepo.New(f.pool)
}

func (f *RepositoryFactory) NewMovementRepository(ctx context.Context) *mRepo.MovementRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return mRepo.New(tx)
	}
	return mRepo.New(f.pool)
}

func (f *RepositoryFactory) NewIncomingRepository(ctx context.Context) *iRepo.IncomingRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return iRepo.New(tx)
	}
	return iRepo.New(f.pool)
}

func (f *RepositoryFactory) NewProductionRepository(ctx context.Context) *prRepo.ProductionRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return prRepo.New(tx)
	}
	return prRepo.New(f.pool)
}

func (f *RepositoryFactory) NewOutgoingRepository(ctx context.Context) *oRepo.OutgoingRepository {
	if tx, ok := storage.GetTx(ctx); ok && tx != nil {
		return oRepo.New(tx)
	}
	return oRepo.New(f.pool)
}