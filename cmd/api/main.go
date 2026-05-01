package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vkr/internal/app"
	"vkr/internal/config"

	"github.com/gin-gonic/gin"
)

var Application *app.App

func main() {
	cfg := config.MustLoad()
	log.Println("Config is loaded")

	var err error

	Application, err = app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	r := setupRouter()
	
	srv := http.Server{
		Addr: 		":"+cfg.API.Port,
		Handler:	r, 
	}
	
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failder to start API server: %v", err)
		}
	}()

	log.Println("Preparing gracefull shutdown")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down the app gracefully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	defer Application.Close()
}

func setupRouter() *gin.Engine {
	router := gin.Default()

	router.Use(func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Header("Access-Control-Expose-Headers", "Content-Length")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    })

	api := router.Group("/api/v1")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			ctx := context.Background()
			err := Application.DB.Ping(ctx)
			if err != nil {
				c.Status(501)
				return
			}
			c.Status(200)
		})

		// ========== СПРАВОЧНИКИ ==========
		
		// Товары (products)
		productGroup := api.Group("/products")
		{
			productGroup.GET("", Application.ProductHandler.List)
			productGroup.GET("/:id", Application.ProductHandler.Get)
			productGroup.POST("", Application.ProductHandler.Create)
			productGroup.PATCH("/:id", Application.ProductHandler.Update)
			productGroup.DELETE("/:id", Application.ProductHandler.Delete)
		}

		// Контрагенты (counterparties)
		counterpartyGroup := api.Group("/counterparties")
		{
			counterpartyGroup.GET("", Application.CounterpartyHandler.List)
			counterpartyGroup.GET("/:id", Application.CounterpartyHandler.Get)
			counterpartyGroup.POST("", Application.CounterpartyHandler.Create)
			counterpartyGroup.PATCH("/:id", Application.CounterpartyHandler.Update)
			counterpartyGroup.DELETE("/:id", Application.CounterpartyHandler.Delete)
		}

		// Склады (warehouses)
		warehouseGroup := api.Group("/warehouses")
		{
			warehouseGroup.GET("", Application.WarehouseHandler.List)
			warehouseGroup.GET("/:id", Application.WarehouseHandler.Get)
			warehouseGroup.POST("", Application.WarehouseHandler.Create)
			warehouseGroup.PATCH("/:id", Application.WarehouseHandler.Update)
			warehouseGroup.DELETE("/:id", Application.WarehouseHandler.Delete)
		}

		// // Спецификации сборки (assemblies)
		assemblyGroup := api.Group("/assemblies")
		{
			assemblyGroup.GET("", Application.AssemblyHandler.List)
			assemblyGroup.GET("/:id", Application.AssemblyHandler.Get)
			assemblyGroup.POST("", Application.AssemblyHandler.Create)
			assemblyGroup.PUT("/:id", Application.AssemblyHandler.Update)
			assemblyGroup.DELETE("/:id", Application.AssemblyHandler.Delete)
			// assemblyGroup.GET("/:id/requirements", Application.AssemblyHandler.GetRequirements)
		}

		// ========== СКЛАДСКИЕ ОПЕРАЦИИ ==========

		// Приход (incoming)
		incomingGroup := api.Group("/incoming")
		{
			incomingGroup.POST("", Application.IncomingHandler.Create)
		}

		// Отгрузка (outgoing)
		outgoingGroup := api.Group("/outgoing")
		{
			outgoingGroup.POST("", Application.OutgoingHandler.Create)
		}

		// Сборка (assembly order)
		assemblyOrderGroup := api.Group("/assembly")
		{
			assemblyOrderGroup.POST("", Application.AssemblyOrderHandler.Create)
		// 	assemblyOrderGroup.GET("/:id", Application.AssemblyOrderHandler.Get)
		// 	assemblyOrderGroup.GET("", Application.AssemblyOrderHandler.List)
		}

		// ========== ОТЧЁТЫ И ПРОСМОТРЫ ==========

		// Остатки (stocks)
		stocksGroup := api.Group("/stocks")
		{
			stocksGroup.GET("", Application.StockHandler.List)
		}

		// Движения (movements) — VIEW
		movementsGroup := api.Group("/movements")
		{
			movementsGroup.GET("", Application.MovementHandler.List)
		}

		// Партии (batches)
		batchesGroup := api.Group("/batches")
		{
			batchesGroup.GET("", Application.BatchHandler.List)
			batchesGroup.GET("/:id", Application.BatchHandler.Get)
		}

		// // Уведомления (alerts)
		alertsGroup := api.Group("/alerts")
		{
			alertsGroup.GET("", Application.AlertHandler.List)
			alertsGroup.PATCH("/:id/resolve", Application.AlertHandler.Resolve)
		}

		// // Отчёты (reports)
		reportsGroup := api.Group("/reports")
		{
			reportsGroup.GET("/cogs", Application.ReportHandler.COGS)
		}
	}

	return router
}