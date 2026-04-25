package app

import (
	"context"
	"log"
	"vkr/internal/config"
	storage "vkr/internal/storage/postgres"
	repos "vkr/internal/repository/postgres"

	"github.com/jackc/pgx/v5/pgxpool"

	pHandler "vkr/internal/handlers/product"
	pRepo "vkr/internal/repository/postgres/product"
	pService "vkr/internal/service/product"

	cpHandler "vkr/internal/handlers/counterparty"
	cpRepo "vkr/internal/repository/postgres/counterparty"
	cpService "vkr/internal/service/counterparty"

	wHandler "vkr/internal/handlers/warehouses"
	wRepo "vkr/internal/repository/postgres/warehouse"
	wService "vkr/internal/service/warehouse"

	rHandler "vkr/internal/handlers/recipe"
	rRepo "vkr/internal/repository/postgres/recipe"
	rService "vkr/internal/service/recipe"

	sHandler "vkr/internal/handlers/stock"
	sRepo "vkr/internal/repository/postgres/stock"
	sService "vkr/internal/service/stock"

	mHandler "vkr/internal/handlers/movement"
	mRepo "vkr/internal/repository/postgres/movement"
	mService "vkr/internal/service/movement"

	iHandler "vkr/internal/handlers/incoming"
	iService "vkr/internal/service/document/incoming"
	iRepo "vkr/internal/repository/postgres/document/incoming"
)

type App struct {
	Config *config.Config

	DB 		*pgxpool.Pool
	TxMan 	*storage.TransactionManager
	RepoFactory 	*repos.RepositoryFactory

	ProductService *pService.ProductService
	CounterPartyService *cpService.CounterpartyService
	WarehouseService *wService.WarehouseService
	RecipeService *rService.RecipeService
	StockService *sService.StockService
	MovementService *mService.MovementService
	IncomingService *iService.IncomingDocumentService

	ProductHandler *pHandler.ProductHandler
	CounterPartyHandler *cpHandler.CounterpartyHandler
	WarehouseHandler *wHandler.WarehouseHandler
	RecipeHandler *rHandler.RecipeHandler
	StockHandler *sHandler.StockHandler
	MovementHandler *mHandler.MovementHandler
	IncomingHandler *iHandler.IncomingHandler

	ProductRepository *pRepo.ProductRepository
	CounterPartyRepository *cpRepo.CounterpartyRepository
	WarehouseRepository *wRepo.WarehouseRepository
	RecipeRepository *rRepo.RecipeRepository
	StockRepository *sRepo.StockRepository
	MovementRepository *mRepo.MovementRepository
	IncomingRepository *iRepo.IncomingRepository
}

func New(cfg *config.Config) (*App, error) {
	app := &App{Config: cfg}

	if err := app.initDB(); err != nil {
		return nil, err
	}

	app.RepoFactory = repos.New(app.DB)
	
	app.initRepos()
	app.initService()
	app.initHandlers()

	return app, nil
}

func (app *App) initDB() error {
    ctx := context.Background()
    db, err := storage.NewPool(ctx, *app.Config)
    if err != nil {
        return err
    }

	db.Ping(ctx)
	log.Printf("Pinging DB %v", err)

    app.DB = db
    app.TxMan = storage.NewTransactionManager(db)
    return nil
}

func (app *App) initService() {
	app.ProductService = pService.New(app.ProductRepository, app.ProductRepository)
	app.CounterPartyService = cpService.New(app.CounterPartyRepository, app.CounterPartyRepository)
	app.WarehouseService = wService.New(app.WarehouseRepository, app.WarehouseRepository)
	app.RecipeService = rService.New(app.TxMan, app.RecipeRepository, app.RecipeRepository, app.ProductRepository)
	app.StockService = sService.New(
		app.TxMan, 
		app.RepoFactory,
		app.ProductRepository,
	)
	app.MovementService = mService.New(app.MovementRepository, app.MovementRepository, app.ProductRepository)
	app.IncomingService = iService.New(
		app.TxMan, 
		app.RepoFactory,
		app.ProductRepository,
		app.WarehouseRepository, 
		app.CounterPartyRepository, 
		app.StockService, 
	)
}

func (app *App) initHandlers() {
	app.ProductHandler = pHandler.New(app.ProductService)
	app.CounterPartyHandler = cpHandler.New(app.CounterPartyService)
	app.WarehouseHandler = wHandler.New(app.WarehouseService)
	app.RecipeHandler = rHandler.New(app.RecipeService, app.ProductRepository)
	app.StockHandler = sHandler.New(app.StockService, app.ProductRepository, app.WarehouseRepository)
	app.MovementHandler = mHandler.New(app.MovementService, app.ProductRepository, app.WarehouseRepository)
	app.IncomingHandler = iHandler.New(app.IncomingService, app.ProductRepository)
}

func (app *App) initRepos() {
	app.ProductRepository = pRepo.New(app.DB)
	app.CounterPartyRepository = cpRepo.New(app.DB)
	app.WarehouseRepository = wRepo.New(app.DB)
	app.RecipeRepository = rRepo.New(app.DB)
	app.StockRepository = sRepo.New(app.DB)
	app.MovementRepository = mRepo.New(app.DB)
	app.IncomingRepository = iRepo.New(app.DB)
}

func (app *App) Close() {
	app.DB.Close()
	log.Println("Database is closed")
}