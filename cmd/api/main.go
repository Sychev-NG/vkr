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

	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			ctx := context.Background()
			err := Application.DB.Ping(ctx)
			if err != nil {
				c.Status(501)
				return				
			}

			c.Status(200)
		})

		// Группа товаров
		productGroup := api.Group("/product")
		{
			productGroup.GET("", Application.ProductHandler.List)
			productGroup.GET("/:id", Application.ProductHandler.Get)
			productGroup.POST("", Application.ProductHandler.Create)
			productGroup.PATCH("/:id", Application.ProductHandler.Update)
			productGroup.DELETE("/:id", Application.ProductHandler.Delete)
		}

        // Группа контрагентов
        counterpartyGroup := api.Group("/counterparty")
        {
            counterpartyGroup.GET("", Application.CounterPartyHandler.List)
            counterpartyGroup.GET("/:id", Application.CounterPartyHandler.Get)
            counterpartyGroup.POST("", Application.CounterPartyHandler.Create)
            counterpartyGroup.PATCH("/:id", Application.CounterPartyHandler.Update)
            counterpartyGroup.DELETE("/:id", Application.CounterPartyHandler.Delete)
        }

        // Группа складов
        warehouseGroup := api.Group("/warehouse")
        {
            warehouseGroup.GET("", Application.WarehouseHandler.List)
            warehouseGroup.GET("/:id", Application.WarehouseHandler.Get)
            warehouseGroup.POST("", Application.WarehouseHandler.Create)
            warehouseGroup.PATCH("/:id", Application.WarehouseHandler.Update)
            warehouseGroup.DELETE("/:id", Application.WarehouseHandler.Delete)
        }

        // Группа рецептов
        recipeGroup := api.Group("/recipe")
        {
            recipeGroup.GET("", Application.RecipeHandler.List)
            recipeGroup.GET("/:id", Application.RecipeHandler.Get)
            recipeGroup.POST("", Application.RecipeHandler.Create)
            recipeGroup.PATCH("/:id", Application.RecipeHandler.Update)
            recipeGroup.DELETE("/:id", Application.RecipeHandler.Delete)
        }

		stocksGroup := api.Group("/stocks")
		{
			stocksGroup.GET("", Application.StockHandler.List)
		}

		movementsGroup := api.Group("/movements")
		{
			movementsGroup.GET("", Application.MovementHandler.List)
		}

		incomingGroup := api.Group("/incoming")
		{
			incomingGroup.POST("", Application.IncomingHandler.Create)
		}

		productionGroup := api.Group("/production")
		{
			productionGroup.POST("", Application.ProductionHandler.Create)
		}

		outgoingGroup := api.Group("/outgoing")
		{
			outgoingGroup.POST("", Application.OutgoingHandler.Create)
		}
	}

	return router
}