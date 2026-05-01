package app

import (
	"context"
	"log"
	"vkr/internal/config"

	"vkr/internal/event"
	eventSubscriber "vkr/internal/event/subscriber"
	domainEvent "vkr/internal/entity/event"

	repos "vkr/internal/repository/postgres"
	storage "vkr/internal/storage/postgres"

	"github.com/jackc/pgx/v5/pgxpool"

	// Handlers
	alHandler "vkr/internal/handlers/alert"
	aHandler "vkr/internal/handlers/assembly"
	aoHandler "vkr/internal/handlers/assembly_order"
	bHandler "vkr/internal/handlers/batch"
	cpHandler "vkr/internal/handlers/counterparty"
	iHandler "vkr/internal/handlers/incoming"
	mHandler "vkr/internal/handlers/movement"
	oHandler "vkr/internal/handlers/outgoing"
	pHandler "vkr/internal/handlers/product"
	rHandler "vkr/internal/handlers/report"
	sHandler "vkr/internal/handlers/stock"
	wHandler "vkr/internal/handlers/warehouses"

	// Repositories
	aRepo "vkr/internal/repository/postgres/assembly"
	cpRepo "vkr/internal/repository/postgres/counterparty"
	pRepo "vkr/internal/repository/postgres/product"
	wRepo "vkr/internal/repository/postgres/warehouse"

	// acRepo "vkr/internal/repository/postgres/assembly_component"
	aoRepo "vkr/internal/repository/postgres/assemblyorder"
	// aocRepo "vkr/internal/repository/postgres/assembly_order_consumption"
	// aooRepo "vkr/internal/repository/postgres/assembly_order_output"
	alRepo "vkr/internal/repository/postgres/alert"
	bRepo "vkr/internal/repository/postgres/batch"
	iRepo "vkr/internal/repository/postgres/incoming"
	mRepo "vkr/internal/repository/postgres/movement"
	sRepo "vkr/internal/repository/postgres/stock"

	// iiRepo "vkr/internal/repository/postgres/incoming_item"
	oRepo "vkr/internal/repository/postgres/outgoing"
	// oiRepo "vkr/internal/repository/postgres/outgoing_item"

	// Services
	aService "vkr/internal/service/assembly"
	aoService "vkr/internal/service/assemblyorder"
	cpService "vkr/internal/service/counterparty"
	pService "vkr/internal/service/product"
	sService "vkr/internal/service/stock"
	wService "vkr/internal/service/warehouse"

	// bService "vkr/internal/service/batches"
	alService "vkr/internal/service/alert"
	iService "vkr/internal/service/incoming"
	oService "vkr/internal/service/outgoing"
	cogsService "vkr/internal/service/report/cogs"
)

type App struct {
	Config *config.Config

	DB      *pgxpool.Pool
	TxMan   *storage.TransactionManager
	RepoFactory *repos.RepositoryFactory
	EventDispatcher *event.EventDispatcher

	// Services
	ProductService     		*pService.ProductService
	CounterpartyService 	*cpService.CounterpartyService
	WarehouseService   		*wService.WarehouseService
	AssemblyService    		*aService.AssemblyService
	AssemblyOrderService 	*aoService.AssemblyOrderService
	StockService       		*sService.StockService
	// BatchService       *bService.BatchService
	AlertService       		*alService.AlertService
	IncomingService    		*iService.IncomingDocumentService
	OutgoingService    		*oService.OutgoingDocumentService
	COGSReportService    	*cogsService.COGSReportService

	// Handlers
	ProductHandler      	*pHandler.ProductHandler
	CounterpartyHandler 	*cpHandler.CounterpartyHandler
	WarehouseHandler    	*wHandler.WarehouseHandler
	AssemblyHandler     	*aHandler.AssemblyHandler
	AssemblyOrderHandler 	*aoHandler.AssemblyOrderHandler
	StockHandler        	*sHandler.StockHandler
	BatchHandler        	*bHandler.BatchHandler
	MovementHandler        	*mHandler.MovementHandler
	AlertHandler        	*alHandler.AlertHandler
	IncomingHandler     	*iHandler.IncomingHandler
	OutgoingHandler     	*oHandler.OutgoingHandler
	ReportHandler	       *rHandler.ReportHandler

	ProductRepository 		*pRepo.ProductRepository
	CounterPartyRepository 	*cpRepo.CounterpartyRepository
	WarehouseRepository 	*wRepo.WarehouseRepository

	AlertRepository     	*alRepo.AlertRepository
	IncomingRepository     	*iRepo.IncomingRepository
	OutgoingRepository     	*oRepo.OutgoingRepository
	StockRepository     	*sRepo.StockRepository
	BatchRepository     	*bRepo.BatchRepository
	MovementRepository     	*mRepo.MovementRepository
	AssemblyRepository     	*aRepo.AssemblyRepository
	AssemblyOrderRepository *aoRepo.AssemblyOrderRepository
}

func New(cfg *config.Config) (*App, error) {
	app := &App{Config: cfg}

	if err := app.initDB(); err != nil {
		return nil, err
	}

	app.EventDispatcher = event.New()
	app.RepoFactory = repos.New(app.DB)

	app.initRepos()
	app.initServices()
	app.initHandlers()
	app.initEventBus()

	return app, nil
}

func (app *App) initDB() error {
	ctx := context.Background()
	db, err := storage.NewPool(ctx, *app.Config)
	if err != nil {
		return err
	}

	if err := db.Ping(ctx); err != nil {
		log.Printf("DB Ping error: %v", err)
		return err
	}
	log.Printf("Database connected successfully")

	app.DB = db
	app.TxMan = storage.NewTransactionManager(db)
	return nil
}

func (app *App) initServices() {
	app.StockService = sService.New(app.TxMan, app.RepoFactory, app.EventDispatcher)
	app.ProductService = pService.New(app.ProductRepository, app.ProductRepository)
	app.CounterpartyService = cpService.New(app.CounterPartyRepository, app.CounterPartyRepository)
	app.WarehouseService = wService.New(app.WarehouseRepository, app.WarehouseRepository)
	app.AssemblyService = aService.New(app.TxMan, app.RepoFactory, app.ProductRepository)
	app.AssemblyOrderService = aoService.New(app.TxMan, app.RepoFactory, app.AssemblyRepository, app.WarehouseRepository, app.ProductRepository, app.StockService)
	app.AlertService = alService.New(app.AlertRepository, app.AlertRepository)
	app.IncomingService = iService.New(app.TxMan, app.RepoFactory, app.ProductRepository, app.WarehouseRepository, app.CounterPartyRepository, app.StockService)
	app.OutgoingService = oService.New(app.TxMan, app.RepoFactory, app.ProductRepository, app.WarehouseRepository, app.CounterPartyRepository, app.StockService)
	app.COGSReportService = cogsService.New(app.TxMan, app.RepoFactory, app.ProductRepository, app.CounterPartyRepository)
}

func (app *App) initHandlers() {
	app.ProductHandler = pHandler.New(app.ProductService)
	app.CounterpartyHandler = cpHandler.New(app.CounterpartyService)
	app.WarehouseHandler = wHandler.New(app.WarehouseService)
	app.AssemblyHandler = aHandler.New(app.AssemblyService)
	app.AssemblyOrderHandler = aoHandler.New(app.AssemblyOrderService)
	app.StockHandler = sHandler.New(app.StockService, app.ProductRepository, app.WarehouseRepository)
	app.BatchHandler = bHandler.New(app.BatchRepository, app.ProductRepository, app.WarehouseRepository)
	app.MovementHandler = mHandler.New(app.MovementRepository, app.ProductRepository, app.WarehouseRepository)
	app.AlertHandler = alHandler.New(app.AlertService, app.ProductRepository, app.WarehouseRepository)
	app.IncomingHandler = iHandler.New(app.IncomingService)
	app.OutgoingHandler = oHandler.New(app.OutgoingService)
	app.ReportHandler = rHandler.New(app.COGSReportService)
}

func (app *App) initRepos() {
	app.ProductRepository = pRepo.New(app.DB)
	app.CounterPartyRepository = cpRepo.New(app.DB)
	app.WarehouseRepository = wRepo.New(app.DB)
	app.StockRepository = sRepo.New(app.DB)
	app.BatchRepository = bRepo.New(app.DB)
	app.MovementRepository = mRepo.New(app.DB)
	app.IncomingRepository = iRepo.New(app.DB)
	app.AlertRepository = alRepo.New(app.DB)
	app.OutgoingRepository = oRepo.New(app.DB)
	app.StockRepository = sRepo.New(app.DB)
	app.AssemblyRepository = aRepo.New(app.DB)
}

func (app *App) initEventBus() {
	app.EventDispatcher.Subscribe(
		eventSubscriber.New(app.AlertService, app.ProductRepository), 
		&domainEvent.StockEvent{},
	)
}

func (app *App) Close() {
	if app.DB != nil {
		app.DB.Close()
		log.Println("Database connection closed")
	}
}